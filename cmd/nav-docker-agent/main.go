package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	protocol "docker-go/agent"
	"docker-go/dockerops"
	"docker-go/domains"

	"github.com/gorilla/websocket"
)

var agentVersion = "m6-dev"

type apiResponse struct {
	Code int             `json:"code"`
	Data json.RawMessage `json:"data"`
	Msg  string          `json:"msg"`
}

type agentApp struct {
	serverURL string
	agentGuid string
	token     string
	name      string
	executor  *dockerops.Executor
	writeMu   sync.Mutex
}

func main() {
	serverURL := flag.String("server", "http://127.0.0.1:8888/api", "server API base URL")
	agentGuid := flag.String("agent-guid", "", "registered agent guid")
	token := flag.String("token", "", "agent token")
	name := flag.String("name", hostname(), "agent display name")
	dockerHost := flag.String("docker-host", "", "optional Docker host, for example unix:///var/run/docker.sock")
	workspace := flag.String("workspace", "", "compose workspace directory, default is ~/.nav-docker/workspaces")
	upgradePublicKey := flag.String("upgrade-public-key", "", "optional Ed25519 public key for verifying agent upgrade binaries")
	requireUpgradeSignature := flag.Bool("require-upgrade-signature", false, "require agent upgrade payloads to include a valid Ed25519 signature")
	maxUpgradeMB := flag.Int64("max-upgrade-mb", 200, "maximum agent upgrade binary size in MiB")
	flag.Parse()

	if strings.TrimSpace(*token) == "" {
		log.Fatal("missing -token")
	}
	executor, err := dockerops.NewExecutor(dockerops.Options{
		DockerHost:              *dockerHost,
		WorkspaceDir:            *workspace,
		AgentVersion:            agentVersion,
		UpgradePublicKey:        *upgradePublicKey,
		RequireUpgradeSignature: *requireUpgradeSignature,
		MaxUpgradeBytes:         *maxUpgradeMB * 1024 * 1024,
	})
	if err != nil {
		log.Fatalf("create docker executor: %v", err)
	}
	defer executor.Close()

	app := &agentApp{
		serverURL: strings.TrimRight(strings.TrimSpace(*serverURL), "/"),
		agentGuid: strings.TrimSpace(*agentGuid),
		token:     strings.TrimSpace(*token),
		name:      strings.TrimSpace(*name),
		executor:  executor,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if app.agentGuid == "" {
		if err := app.register(ctx); err != nil {
			log.Fatalf("register agent: %v", err)
		}
	}
	app.run(ctx)
}

func (a *agentApp) register(ctx context.Context) error {
	payload := map[string]any{
		"token":         a.token,
		"name":          a.name,
		"version":       agentVersion,
		"dockerVersion": a.executor.DockerVersion(ctx),
		"os":            runtime.GOOS,
		"arch":          runtime.GOARCH,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.serverURL+"/agents/register", strings.NewReader(string(raw)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}
	var apiResp apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return err
	}
	if apiResp.Code != http.StatusOK {
		return errors.New(apiResp.Msg)
	}
	var agent domains.Agent
	if err := json.Unmarshal(apiResp.Data, &agent); err != nil {
		return err
	}
	if agent.Guid == "" {
		return errors.New("register response missing agent guid")
	}
	a.agentGuid = agent.Guid
	log.Printf("registered agent: %s", a.agentGuid)
	return nil
}

func (a *agentApp) run(ctx context.Context) {
	for ctx.Err() == nil {
		if err := a.connectAndServe(ctx); err != nil && ctx.Err() == nil {
			log.Printf("agent connection closed: %v", err)
			time.Sleep(3 * time.Second)
		}
	}
}

func (a *agentApp) connectAndServe(ctx context.Context) error {
	wsURL, err := a.wsURL()
	if err != nil {
		return err
	}
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, wsURL, nil)
	if err != nil {
		return err
	}
	defer conn.Close()
	log.Printf("connected to %s", wsURL)

	done := make(chan struct{})
	go a.heartbeatLoop(ctx, conn, done)
	defer close(done)

	for {
		var msg protocol.Message
		if err := conn.ReadJSON(&msg); err != nil {
			return err
		}
		if msg.Type != protocol.MessageTypeTaskDispatch {
			continue
		}
		go a.handleTask(ctx, conn, msg)
	}
}

func (a *agentApp) handleTask(parent context.Context, conn *websocket.Conn, msg protocol.Message) {
	var payload protocol.TaskDispatchPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		a.sendTaskEvent(conn, msg.TaskGuid, domains.TaskStatusFailed, "invalid dispatch payload", nil, err)
		return
	}
	timeout := payload.TimeoutSeconds
	if timeout <= 0 {
		timeout = 600
	}
	ctx, cancel := context.WithTimeout(parent, time.Duration(timeout)*time.Second)
	defer cancel()

	a.sendTaskEvent(conn, msg.TaskGuid, domains.TaskStatusRunning, "task running", map[string]any{"type": payload.Type}, nil)
	if payload.Type == domains.TaskTypeDockerContainerStream {
		a.handleLogStream(ctx, conn, msg.TaskGuid, payload.Payload)
		return
	}
	result, err := a.executor.Execute(ctx, payload.Type, payload.Payload)
	if err != nil {
		a.sendTaskEvent(conn, msg.TaskGuid, domains.TaskStatusFailed, "task failed", nil, err)
		return
	}
	a.sendTaskEvent(conn, msg.TaskGuid, domains.TaskStatusSuccess, "task success", result, nil)
}

func (a *agentApp) handleLogStream(ctx context.Context, conn *websocket.Conn, taskGuid string, rawPayload json.RawMessage) {
	err := a.executor.StreamContainerLogs(ctx, rawPayload, func(line domains.DockerLogLine) {
		a.sendTaskEvent(conn, taskGuid, domains.TaskStatusRunning, "log", line, nil)
	})
	if err != nil {
		a.sendTaskEvent(conn, taskGuid, domains.TaskStatusFailed, "log stream failed", nil, err)
		return
	}
	a.sendTaskEvent(conn, taskGuid, domains.TaskStatusSuccess, "log stream finished", map[string]any{"finished": true}, nil)
}

func (a *agentApp) heartbeatLoop(ctx context.Context, conn *websocket.Conn, done <-chan struct{}) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-done:
			return
		case <-ticker.C:
			a.writeJSON(conn, protocol.Message{
				Type:      protocol.MessageTypeAgentHeartbeat,
				AgentGuid: a.agentGuid,
				Data:      json.RawMessage(fmt.Sprintf(`{"version":%q}`, agentVersion)),
				Timestamp: time.Now().UnixMilli(),
			})
		}
	}
}

func (a *agentApp) sendTaskEvent(conn *websocket.Conn, taskGuid string, status string, message string, result any, taskErr error) {
	msg := protocol.Message{
		Type:      protocol.MessageTypeTaskEvent,
		TaskGuid:  taskGuid,
		Status:    status,
		Message:   message,
		Data:      json.RawMessage(`{}`),
		Timestamp: time.Now().UnixMilli(),
	}
	if result != nil {
		raw, err := json.Marshal(result)
		if err != nil {
			msg.Status = domains.TaskStatusFailed
			msg.Message = "marshal task result failed"
			msg.ErrorMessage = err.Error()
		} else {
			msg.Data = raw
			if status == domains.TaskStatusSuccess {
				msg.Result = raw
			}
		}
	}
	if taskErr != nil {
		msg.ErrorMessage = taskErr.Error()
	}
	a.writeJSON(conn, msg)
}

func (a *agentApp) writeJSON(conn *websocket.Conn, msg protocol.Message) {
	a.writeMu.Lock()
	defer a.writeMu.Unlock()
	if msg.Timestamp == 0 {
		msg.Timestamp = time.Now().UnixMilli()
	}
	if err := conn.WriteJSON(msg); err != nil {
		log.Printf("write websocket message failed: %v", err)
	}
}

func (a *agentApp) wsURL() (string, error) {
	u, err := url.Parse(a.serverURL)
	if err != nil {
		return "", err
	}
	switch u.Scheme {
	case "http":
		u.Scheme = "ws"
	case "https":
		u.Scheme = "wss"
	case "ws", "wss":
	default:
		return "", fmt.Errorf("unsupported server scheme: %s", u.Scheme)
	}
	u.Path = "/" + strings.TrimPrefix(path.Join(u.Path, "agent/ws"), "/")
	query := u.Query()
	query.Set("agentGuid", a.agentGuid)
	query.Set("token", a.token)
	u.RawQuery = query.Encode()
	return u.String(), nil
}

func hostname() string {
	name, err := os.Hostname()
	if err != nil || strings.TrimSpace(name) == "" {
		return "nav-docker-agent"
	}
	return name
}

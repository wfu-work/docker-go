package dockerops

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"time"

	"docker-go/domains"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
	gopsutilNet "github.com/shirou/gopsutil/v4/net"
)

type Executor struct {
	cli                     *client.Client
	workspaceDir            string
	agentVersion            string
	upgradePublicKey        string
	requireUpgradeSignature bool
	maxUpgradeBytes         int64
}

type Options struct {
	DockerHost              string
	WorkspaceDir            string
	AgentVersion            string
	UpgradePublicKey        string
	RequireUpgradeSignature bool
	MaxUpgradeBytes         int64
}

type containerActionPayload struct {
	ContainerID   string `json:"containerId"`
	Force         bool   `json:"force"`
	RemoveVolumes bool   `json:"removeVolumes"`
	Timeout       int    `json:"timeout"`
}

type imagePullPayload struct {
	Image        string `json:"image"`
	Platform     string `json:"platform"`
	RegistryAuth string `json:"registryAuth"`
}

type logsPayload struct {
	ContainerID string `json:"containerId"`
	Tail        string `json:"tail"`
	Since       string `json:"since"`
	Until       string `json:"until"`
	Timestamps  bool   `json:"timestamps"`
	Follow      bool   `json:"follow"`
}

func NewExecutor(options Options) (*Executor, error) {
	opts := []client.Opt{
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	}
	if strings.TrimSpace(options.DockerHost) != "" {
		opts = append(opts, client.WithHost(strings.TrimSpace(options.DockerHost)))
	}
	cli, err := client.NewClientWithOpts(opts...)
	if err != nil {
		return nil, err
	}
	return &Executor{
		cli:                     cli,
		workspaceDir:            normalizeWorkspaceDir(options.WorkspaceDir),
		agentVersion:            strings.TrimSpace(options.AgentVersion),
		upgradePublicKey:        strings.TrimSpace(options.UpgradePublicKey),
		requireUpgradeSignature: options.RequireUpgradeSignature,
		maxUpgradeBytes:         normalizeMaxUpgradeBytes(options.MaxUpgradeBytes),
	}, nil
}

func (e *Executor) Close() error {
	if e == nil || e.cli == nil {
		return nil
	}
	return e.cli.Close()
}

func (e *Executor) DockerVersion(ctx context.Context) string {
	if e == nil || e.cli == nil {
		return ""
	}
	version, err := e.cli.ServerVersion(ctx)
	if err != nil {
		return ""
	}
	return version.Version
}

func (e *Executor) Execute(ctx context.Context, taskType string, rawPayload json.RawMessage) (any, error) {
	if e == nil || e.cli == nil {
		return nil, errors.New("docker executor is not initialized")
	}
	switch taskType {
	case domains.TaskTypeAgentPing:
		return map[string]any{"pong": true}, nil
	case domains.TaskTypeAgentUpgrade:
		return e.UpgradeAgent(ctx, rawPayload)
	case domains.TaskTypeDockerContainerList:
		return e.ListContainers(ctx)
	case domains.TaskTypeDockerContainerStart:
		payload, err := parseContainerPayload(rawPayload)
		if err != nil {
			return nil, err
		}
		return e.StartContainer(ctx, payload.ContainerID)
	case domains.TaskTypeDockerContainerStop:
		payload, err := parseContainerPayload(rawPayload)
		if err != nil {
			return nil, err
		}
		return e.StopContainer(ctx, payload.ContainerID, payload.Timeout)
	case domains.TaskTypeDockerContainerRestart:
		payload, err := parseContainerPayload(rawPayload)
		if err != nil {
			return nil, err
		}
		return e.RestartContainer(ctx, payload.ContainerID, payload.Timeout)
	case domains.TaskTypeDockerContainerRemove:
		payload, err := parseContainerPayload(rawPayload)
		if err != nil {
			return nil, err
		}
		return e.RemoveContainer(ctx, payload)
	case domains.TaskTypeDockerContainerLogs:
		payload, err := parseLogsPayload(rawPayload)
		if err != nil {
			return nil, err
		}
		return e.ContainerLogs(ctx, payload)
	case domains.TaskTypeDockerImageList:
		return e.ListImages(ctx)
	case domains.TaskTypeDockerImagePull:
		payload, err := parseImagePullPayload(rawPayload)
		if err != nil {
			return nil, err
		}
		return e.PullImage(ctx, payload)
	case domains.TaskTypeDockerMetricsSnapshot:
		return e.MetricsSnapshot(ctx)
	case domains.TaskTypeDockerConfigValidate:
		return e.ComposeValidate(ctx, rawPayload)
	case domains.TaskTypeDockerConfigDeploy,
		domains.TaskTypeDockerComposeUp,
		domains.TaskTypeDockerComposeDown,
		domains.TaskTypeDockerComposeRestart,
		domains.TaskTypeDockerComposePull:
		return e.ComposeAction(ctx, taskType, rawPayload)
	default:
		return nil, errors.New("unsupported task type: " + taskType)
	}
}

func (e *Executor) ListContainers(ctx context.Context) (domains.DockerResourceSnapshot, error) {
	items, err := e.cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return domains.DockerResourceSnapshot{}, err
	}
	containers := make([]domains.DockerContainerSnapshot, 0, len(items))
	for _, item := range items {
		containers = append(containers, domains.DockerContainerSnapshot{
			ContainerID: item.ID,
			Names:       item.Names,
			Image:       item.Image,
			ImageID:     item.ImageID,
			Command:     item.Command,
			State:       item.State,
			Status:      item.Status,
			Ports:       item.Ports,
			Labels:      item.Labels,
			CreatedAt:   item.Created,
		})
	}
	return domains.DockerResourceSnapshot{Containers: containers}, nil
}

func (e *Executor) StartContainer(ctx context.Context, containerID string) (map[string]any, error) {
	containerID, err := requireContainerID(containerID)
	if err != nil {
		return nil, err
	}
	if err := e.cli.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		return nil, err
	}
	return map[string]any{"containerId": containerID, "action": domains.TaskTypeDockerContainerStart}, nil
}

func (e *Executor) StopContainer(ctx context.Context, containerID string, timeout int) (map[string]any, error) {
	containerID, err := requireContainerID(containerID)
	if err != nil {
		return nil, err
	}
	options := container.StopOptions{}
	if timeout != 0 {
		options.Timeout = &timeout
	}
	if err := e.cli.ContainerStop(ctx, containerID, options); err != nil {
		return nil, err
	}
	return map[string]any{"containerId": containerID, "action": domains.TaskTypeDockerContainerStop}, nil
}

func (e *Executor) RestartContainer(ctx context.Context, containerID string, timeout int) (map[string]any, error) {
	containerID, err := requireContainerID(containerID)
	if err != nil {
		return nil, err
	}
	options := container.StopOptions{}
	if timeout != 0 {
		options.Timeout = &timeout
	}
	if err := e.cli.ContainerRestart(ctx, containerID, options); err != nil {
		return nil, err
	}
	return map[string]any{"containerId": containerID, "action": domains.TaskTypeDockerContainerRestart}, nil
}

func (e *Executor) RemoveContainer(ctx context.Context, payload containerActionPayload) (map[string]any, error) {
	containerID, err := requireContainerID(payload.ContainerID)
	if err != nil {
		return nil, err
	}
	if err := e.cli.ContainerRemove(ctx, containerID, container.RemoveOptions{
		Force:         payload.Force,
		RemoveVolumes: payload.RemoveVolumes,
	}); err != nil {
		return nil, err
	}
	return map[string]any{"containerId": containerID, "action": domains.TaskTypeDockerContainerRemove}, nil
}

func (e *Executor) ListImages(ctx context.Context) (domains.DockerResourceSnapshot, error) {
	items, err := e.cli.ImageList(ctx, image.ListOptions{All: true})
	if err != nil {
		return domains.DockerResourceSnapshot{}, err
	}
	images := make([]domains.DockerImageSnapshot, 0, len(items))
	for _, item := range items {
		images = append(images, domains.DockerImageSnapshot{
			ImageID:     item.ID,
			RepoTags:    item.RepoTags,
			RepoDigests: item.RepoDigests,
			Labels:      item.Labels,
			Size:        item.Size,
			Containers:  item.Containers,
			CreatedAt:   item.Created,
		})
	}
	return domains.DockerResourceSnapshot{Images: images}, nil
}

func (e *Executor) PullImage(ctx context.Context, payload imagePullPayload) (map[string]any, error) {
	imageName := strings.TrimSpace(payload.Image)
	if imageName == "" {
		return nil, errors.New("missing image")
	}
	reader, err := e.cli.ImagePull(ctx, imageName, image.PullOptions{
		Platform:     strings.TrimSpace(payload.Platform),
		RegistryAuth: strings.TrimSpace(payload.RegistryAuth),
	})
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	if _, err := io.Copy(io.Discard, reader); err != nil {
		return nil, err
	}
	return map[string]any{
		"image":    imageName,
		"platform": strings.TrimSpace(payload.Platform),
		"action":   domains.TaskTypeDockerImagePull,
	}, nil
}

func (e *Executor) ContainerLogs(ctx context.Context, payload logsPayload) (domains.DockerLogResult, error) {
	containerID, err := requireContainerID(payload.ContainerID)
	if err != nil {
		return domains.DockerLogResult{}, err
	}
	reader, err := e.cli.ContainerLogs(ctx, containerID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Since:      strings.TrimSpace(payload.Since),
		Until:      strings.TrimSpace(payload.Until),
		Timestamps: payload.Timestamps,
		Tail:       normalizeTail(payload.Tail),
	})
	if err != nil {
		return domains.DockerLogResult{}, err
	}
	defer reader.Close()
	lines, err := collectLogLines(reader, containerID, e.containerTTY(ctx, containerID))
	if err != nil {
		return domains.DockerLogResult{}, err
	}
	return domains.DockerLogResult{ContainerID: containerID, Lines: lines}, nil
}

func (e *Executor) StreamContainerLogs(ctx context.Context, rawPayload json.RawMessage, emit func(domains.DockerLogLine)) error {
	payload, err := parseLogsPayload(rawPayload)
	if err != nil {
		return err
	}
	containerID, err := requireContainerID(payload.ContainerID)
	if err != nil {
		return err
	}
	reader, err := e.cli.ContainerLogs(ctx, containerID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Since:      strings.TrimSpace(payload.Since),
		Until:      strings.TrimSpace(payload.Until),
		Timestamps: payload.Timestamps,
		Follow:     true,
		Tail:       normalizeTail(payload.Tail),
	})
	if err != nil {
		return err
	}
	defer reader.Close()
	if e.containerTTY(ctx, containerID) {
		return streamRawLogLines(reader, containerID, "stdout", emit)
	}
	stdout := newLogLineEmitter(containerID, "stdout", emit)
	stderr := newLogLineEmitter(containerID, "stderr", emit)
	_, err = stdcopy.StdCopy(stdout, stderr, reader)
	stdout.Flush()
	stderr.Flush()
	return err
}

func (e *Executor) MetricsSnapshot(ctx context.Context) (domains.MetricsSnapshot, error) {
	now := time.Now().UnixMilli()
	snapshot := domains.MetricsSnapshot{
		Host: domains.HostMetricSnapshot{CollectedAt: now},
	}
	cpuPercents, _ := cpu.PercentWithContext(ctx, 200*time.Millisecond, false)
	if len(cpuPercents) > 0 {
		snapshot.Host.CPUPercent = cpuPercents[0]
	}
	snapshot.Host.CPUCores, _ = cpu.CountsWithContext(ctx, true)
	if vm, err := mem.VirtualMemoryWithContext(ctx); err == nil {
		snapshot.Host.MemoryTotal = vm.Total
		snapshot.Host.MemoryUsed = vm.Used
		snapshot.Host.MemoryPercent = vm.UsedPercent
	}
	if usage, err := disk.UsageWithContext(ctx, "/"); err == nil {
		snapshot.Host.DiskTotal = usage.Total
		snapshot.Host.DiskUsed = usage.Used
		snapshot.Host.DiskPercent = usage.UsedPercent
	}
	if counters, err := gopsutilNet.IOCountersWithContext(ctx, false); err == nil && len(counters) > 0 {
		snapshot.Host.NetworkRxBytes = counters[0].BytesRecv
		snapshot.Host.NetworkTxBytes = counters[0].BytesSent
	}

	containers, err := e.cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return snapshot, err
	}
	for _, item := range containers {
		if item.State != "running" {
			continue
		}
		metric, err := e.containerMetric(ctx, item, now)
		if err == nil {
			snapshot.Containers = append(snapshot.Containers, metric)
		}
	}
	return snapshot, nil
}

func parseContainerPayload(raw json.RawMessage) (containerActionPayload, error) {
	var payload containerActionPayload
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &payload); err != nil {
			return payload, err
		}
	}
	_, err := requireContainerID(payload.ContainerID)
	return payload, err
}

func parseLogsPayload(raw json.RawMessage) (logsPayload, error) {
	var payload logsPayload
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &payload); err != nil {
			return payload, err
		}
	}
	payload.ContainerID = strings.TrimSpace(payload.ContainerID)
	if payload.ContainerID == "" {
		return payload, errors.New("missing container id")
	}
	payload.Tail = normalizeTail(payload.Tail)
	return payload, nil
}

func parseImagePullPayload(raw json.RawMessage) (imagePullPayload, error) {
	var payload imagePullPayload
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &payload); err != nil {
			return payload, err
		}
	}
	payload.Image = strings.TrimSpace(payload.Image)
	if payload.Image == "" {
		return payload, errors.New("missing image")
	}
	return payload, nil
}

func requireContainerID(containerID string) (string, error) {
	containerID = strings.TrimSpace(containerID)
	if containerID == "" {
		return "", errors.New("missing container id")
	}
	return containerID, nil
}

func normalizeTail(tail string) string {
	tail = strings.TrimSpace(tail)
	if tail == "" {
		return "200"
	}
	return tail
}

func (e *Executor) containerTTY(ctx context.Context, containerID string) bool {
	inspect, err := e.cli.ContainerInspect(ctx, containerID)
	return err == nil && inspect.Config != nil && inspect.Config.Tty
}

func collectLogLines(reader io.Reader, containerID string, tty bool) ([]domains.DockerLogLine, error) {
	raw, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	lines := make([]domains.DockerLogLine, 0)
	emit := func(line domains.DockerLogLine) {
		lines = append(lines, line)
	}
	if tty {
		return splitRawLogBytes(raw, containerID, "stdout"), nil
	}
	stdout := newLogLineEmitter(containerID, "stdout", emit)
	stderr := newLogLineEmitter(containerID, "stderr", emit)
	if _, err := stdcopy.StdCopy(stdout, stderr, bytes.NewReader(raw)); err != nil {
		return splitRawLogBytes(raw, containerID, "combined"), nil
	}
	stdout.Flush()
	stderr.Flush()
	if len(lines) == 0 && len(raw) > 0 {
		return splitRawLogBytes(raw, containerID, "combined"), nil
	}
	return lines, nil
}

func splitRawLogBytes(raw []byte, containerID string, stream string) []domains.DockerLogLine {
	scanner := bufio.NewScanner(bytes.NewReader(raw))
	scanner.Buffer(make([]byte, 1024), 1024*1024)
	lines := make([]domains.DockerLogLine, 0)
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\r\n")
		if line == "" {
			continue
		}
		lines = append(lines, domains.DockerLogLine{
			ContainerID: containerID,
			Stream:      stream,
			Line:        line,
			Timestamp:   time.Now().UnixMilli(),
		})
	}
	return lines
}

func streamRawLogLines(reader io.Reader, containerID string, stream string, emit func(domains.DockerLogLine)) error {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\r\n")
		if line == "" {
			continue
		}
		emit(domains.DockerLogLine{
			ContainerID: containerID,
			Stream:      stream,
			Line:        line,
			Timestamp:   time.Now().UnixMilli(),
		})
	}
	return scanner.Err()
}

type logLineEmitter struct {
	containerID string
	stream      string
	emit        func(domains.DockerLogLine)
	buffer      []byte
}

func newLogLineEmitter(containerID string, stream string, emit func(domains.DockerLogLine)) *logLineEmitter {
	return &logLineEmitter{containerID: containerID, stream: stream, emit: emit}
}

func (w *logLineEmitter) Write(p []byte) (int, error) {
	w.buffer = append(w.buffer, p...)
	for {
		index := bytes.IndexByte(w.buffer, '\n')
		if index < 0 {
			break
		}
		w.emitLine(w.buffer[:index])
		w.buffer = w.buffer[index+1:]
	}
	return len(p), nil
}

func (w *logLineEmitter) Flush() {
	if len(w.buffer) == 0 {
		return
	}
	w.emitLine(w.buffer)
	w.buffer = nil
}

func (w *logLineEmitter) emitLine(raw []byte) {
	line := strings.TrimRight(string(raw), "\r\n")
	if line == "" {
		return
	}
	w.emit(domains.DockerLogLine{
		ContainerID: w.containerID,
		Stream:      w.stream,
		Line:        line,
		Timestamp:   time.Now().UnixMilli(),
	})
}

func (e *Executor) containerMetric(ctx context.Context, item container.Summary, collectedAt int64) (domains.ContainerMetricSnapshot, error) {
	reader, err := e.cli.ContainerStatsOneShot(ctx, item.ID)
	if err != nil {
		return domains.ContainerMetricSnapshot{}, err
	}
	defer reader.Body.Close()
	var stats container.StatsResponse
	if err := json.NewDecoder(reader.Body).Decode(&stats); err != nil {
		return domains.ContainerMetricSnapshot{}, err
	}
	metric := domains.ContainerMetricSnapshot{
		ContainerID: item.ID,
		Name:        firstContainerName(item.Names),
		CollectedAt: collectedAt,
	}
	metric.CPUPercent = calculateCPUPercent(stats)
	metric.MemoryUsage = calculateMemoryUsage(stats)
	metric.MemoryLimit = stats.MemoryStats.Limit
	if metric.MemoryLimit > 0 {
		metric.MemoryPercent = float64(metric.MemoryUsage) / float64(metric.MemoryLimit) * 100
	}
	for _, item := range stats.Networks {
		metric.NetworkRxBytes += item.RxBytes
		metric.NetworkTxBytes += item.TxBytes
	}
	for _, entry := range stats.BlkioStats.IoServiceBytesRecursive {
		switch strings.ToLower(entry.Op) {
		case "read":
			metric.BlockReadBytes += entry.Value
		case "write":
			metric.BlockWriteBytes += entry.Value
		}
	}
	metric.PidsCurrent = stats.PidsStats.Current
	return metric, nil
}

func firstContainerName(names []string) string {
	if len(names) == 0 {
		return ""
	}
	return strings.TrimPrefix(names[0], "/")
}

func calculateCPUPercent(stats container.StatsResponse) float64 {
	cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(stats.CPUStats.SystemUsage - stats.PreCPUStats.SystemUsage)
	onlineCPUs := float64(stats.CPUStats.OnlineCPUs)
	if onlineCPUs == 0 {
		onlineCPUs = float64(len(stats.CPUStats.CPUUsage.PercpuUsage))
	}
	if systemDelta > 0 && cpuDelta > 0 && onlineCPUs > 0 {
		return cpuDelta / systemDelta * onlineCPUs * 100
	}
	return 0
}

func calculateMemoryUsage(stats container.StatsResponse) uint64 {
	usage := stats.MemoryStats.Usage
	if cache := stats.MemoryStats.Stats["cache"]; cache > 0 && usage > cache {
		return usage - cache
	}
	return usage
}

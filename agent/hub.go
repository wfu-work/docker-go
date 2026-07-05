package agent

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"docker-go/domains"
	"docker-go/services"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var (
	DefaultHub      = NewHub()
	ErrAgentOffline = errors.New("agent is offline")
)

type Hub struct {
	mu               sync.RWMutex
	sessions         map[string]*Session
	hostAgents       map[string]string
	taskTimers       map[string]*time.Timer
	eventSubscribers map[string]map[chan Message]struct{}
	wsUpgrader       websocket.Upgrader
}

func NewHub() *Hub {
	return &Hub{
		sessions:         map[string]*Session{},
		hostAgents:       map[string]string{},
		taskTimers:       map[string]*time.Timer{},
		eventSubscribers: map[string]map[chan Message]struct{}{},
		wsUpgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (h *Hub) SubscribeTaskEvents(taskGuid string) (<-chan Message, func()) {
	ch := make(chan Message, 128)
	h.mu.Lock()
	if h.eventSubscribers[taskGuid] == nil {
		h.eventSubscribers[taskGuid] = map[chan Message]struct{}{}
	}
	h.eventSubscribers[taskGuid][ch] = struct{}{}
	h.mu.Unlock()
	cancel := func() {
		h.mu.Lock()
		if subscribers := h.eventSubscribers[taskGuid]; subscribers != nil {
			delete(subscribers, ch)
			if len(subscribers) == 0 {
				delete(h.eventSubscribers, taskGuid)
			}
		}
		h.mu.Unlock()
		close(ch)
	}
	return ch, cancel
}

func (h *Hub) ForwardTaskEvent(msg Message) {
	h.mu.RLock()
	subscribers := h.eventSubscribers[msg.TaskGuid]
	copied := make([]chan Message, 0, len(subscribers))
	for ch := range subscribers {
		copied = append(copied, ch)
	}
	h.mu.RUnlock()
	for _, ch := range copied {
		select {
		case ch <- msg:
		default:
		}
	}
}

func (h *Hub) Handle(c *gin.Context, agent domains.Agent, token string) error {
	conn, err := h.wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return err
	}
	session := NewSession(h, conn, agent, token)
	session.Run()
	return nil
}

func (h *Hub) Register(session *Session) {
	var old *Session
	h.mu.Lock()
	old = h.sessions[session.agent.Guid]
	h.sessions[session.agent.Guid] = session
	h.hostAgents[session.agent.HostGuid] = session.agent.Guid
	h.mu.Unlock()

	if old != nil && old != session {
		old.Close()
	}
	_, _ = services.AgentServiceApp.MarkConnected(session.agent.Guid)
	_ = session.Send(NewConnectedMessage(session.agent.Guid, session.agent.HostGuid))
	go h.DispatchPendingForAgent(session.agent)
}

func (h *Hub) Unregister(session *Session) {
	shouldMarkOffline := false
	h.mu.Lock()
	if current := h.sessions[session.agent.Guid]; current == session {
		delete(h.sessions, session.agent.Guid)
		if h.hostAgents[session.agent.HostGuid] == session.agent.Guid {
			delete(h.hostAgents, session.agent.HostGuid)
		}
		shouldMarkOffline = true
	}
	h.mu.Unlock()

	session.Close()
	if shouldMarkOffline {
		_, _ = services.AgentServiceApp.MarkDisconnected(session.agent.Guid)
	}
}

func (h *Hub) IsOnline(agentGuid string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.sessions[agentGuid]
	return ok
}

func (h *Hub) DispatchPendingForAgent(agent domains.Agent) {
	tasks, err := services.TaskServiceApp.ListPendingForAgent(agent)
	if err != nil {
		return
	}
	for i := range tasks {
		_ = h.DispatchTask(&tasks[i])
	}
}

func (h *Hub) DispatchTask(task *domains.Task) error {
	if task == nil {
		return errors.New("missing task")
	}
	if task.Status != domains.TaskStatusPending {
		return nil
	}
	session, err := h.sessionForTask(task)
	if err != nil {
		return err
	}
	msg, err := NewTaskDispatchMessage(*task, session.agent.Guid)
	if err != nil {
		return err
	}
	if err = session.Send(msg); err != nil {
		h.Unregister(session)
		return err
	}
	dispatched, err := services.TaskServiceApp.MarkDispatched(task.Guid, session.agent.Guid)
	if err != nil {
		return err
	}
	h.StartTaskTimer(*dispatched)
	return nil
}

func (h *Hub) StartTaskTimer(task domains.Task) {
	timeoutSeconds := task.TimeoutSeconds
	if timeoutSeconds <= 0 {
		timeoutSeconds = 600
	}
	h.StopTaskTimer(task.Guid)
	timer := time.AfterFunc(time.Duration(timeoutSeconds)*time.Second, func() {
		_, _ = services.TaskServiceApp.MarkTimeout(task.Guid)
		h.StopTaskTimer(task.Guid)
	})
	h.mu.Lock()
	h.taskTimers[task.Guid] = timer
	h.mu.Unlock()
}

func (h *Hub) StopTaskTimer(taskGuid string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if timer := h.taskTimers[taskGuid]; timer != nil {
		timer.Stop()
		delete(h.taskTimers, taskGuid)
	}
}

func (h *Hub) sessionForTask(task *domains.Task) (*Session, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if task.AgentGuid != "" {
		session := h.sessions[task.AgentGuid]
		if session == nil {
			return nil, ErrAgentOffline
		}
		if session.agent.HostGuid != task.HostGuid {
			return nil, errors.New("agent host does not match task")
		}
		return session, nil
	}
	agentGuid := h.hostAgents[task.HostGuid]
	if agentGuid == "" {
		return nil, ErrAgentOffline
	}
	session := h.sessions[agentGuid]
	if session == nil {
		return nil, ErrAgentOffline
	}
	return session, nil
}

package agent

import (
	"encoding/json"
	"errors"
	"time"

	"docker-go/domains"
	"docker-go/services"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 70 * time.Second
	pingPeriod     = 30 * time.Second
	maxMessageSize = 1 << 20
	sendBufferSize = 64
)

var ErrSessionClosed = errors.New("agent session closed")

type Session struct {
	hub   *Hub
	conn  *websocket.Conn
	send  chan Message
	done  chan struct{}
	agent domains.Agent
	token string
}

func NewSession(hub *Hub, conn *websocket.Conn, agent domains.Agent, token string) *Session {
	return &Session{
		hub:   hub,
		conn:  conn,
		send:  make(chan Message, sendBufferSize),
		done:  make(chan struct{}),
		agent: agent,
		token: token,
	}
}

func (s *Session) Run() {
	s.hub.Register(s)
	go s.writeLoop()
	s.readLoop()
}

func (s *Session) Send(msg Message) error {
	select {
	case <-s.done:
		return ErrSessionClosed
	case s.send <- msg:
		return nil
	default:
		return errors.New("agent session send buffer is full")
	}
}

func (s *Session) Close() {
	select {
	case <-s.done:
		return
	default:
		close(s.done)
		_ = s.conn.Close()
	}
}

func (s *Session) readLoop() {
	defer s.hub.Unregister(s)
	s.conn.SetReadLimit(maxMessageSize)
	_ = s.conn.SetReadDeadline(time.Now().Add(pongWait))
	s.conn.SetPongHandler(func(string) error {
		return s.conn.SetReadDeadline(time.Now().Add(pongWait))
	})
	for {
		var msg Message
		if err := s.conn.ReadJSON(&msg); err != nil {
			return
		}
		s.handleMessage(msg)
	}
}

func (s *Session) writeLoop() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		s.Close()
	}()
	for {
		select {
		case <-s.done:
			return
		case msg := <-s.send:
			if err := s.writeJSON(msg); err != nil {
				return
			}
		case <-ticker.C:
			if err := s.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				return
			}
			if err := s.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (s *Session) writeJSON(msg Message) error {
	if msg.Timestamp == 0 {
		msg.Timestamp = nowMilli()
	}
	if err := s.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
		return err
	}
	return s.conn.WriteJSON(msg)
}

func (s *Session) handleMessage(msg Message) {
	switch msg.Type {
	case MessageTypeAgentHeartbeat:
		var data struct {
			Version       string `json:"version"`
			DockerVersion string `json:"dockerVersion"`
		}
		if len(msg.Data) > 0 && json.Valid(msg.Data) {
			_ = json.Unmarshal(msg.Data, &data)
		}
		_, _ = services.AgentServiceApp.Heartbeat(services.AgentHeartbeatRequest{
			AgentGuid:     s.agent.Guid,
			Token:         s.token,
			Version:       data.Version,
			DockerVersion: data.DockerVersion,
			Status:        domains.AgentStatusOnline,
		})
	case MessageTypeTaskEvent:
		s.handleTaskEvent(msg)
	default:
		_ = s.Send(NewErrorMessage("unsupported message type: " + msg.Type))
	}
}

func (s *Session) handleTaskEvent(msg Message) {
	if msg.TaskGuid == "" {
		_ = s.Send(NewErrorMessage("missing task guid"))
		return
	}
	data := msg.Data
	if len(data) == 0 || !json.Valid(data) {
		data = json.RawMessage(`{}`)
	}
	req := services.TaskEventCreateRequest{
		AgentGuid:    s.agent.Guid,
		Token:        s.token,
		Status:       msg.Status,
		Message:      msg.Message,
		Data:         data,
		Result:       msg.Result,
		ErrorMessage: msg.ErrorMessage,
	}
	event, err := services.TaskServiceApp.AppendEvent(msg.TaskGuid, req)
	if err != nil {
		_ = s.Send(NewErrorMessage(err.Error()))
		return
	}
	s.hub.ForwardTaskEvent(msg)
	if isFinalStatus(event.Status) {
		s.hub.StopTaskTimer(event.TaskGuid)
	}
}

func isFinalStatus(status string) bool {
	switch status {
	case domains.TaskStatusSuccess, domains.TaskStatusFailed, domains.TaskStatusTimeout, domains.TaskStatusCancelled:
		return true
	default:
		return false
	}
}

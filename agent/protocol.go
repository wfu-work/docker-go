package agent

import (
	"encoding/json"
	"time"

	"docker-go/domains"
)

const (
	MessageTypeAgentConnected = "agent.connected"
	MessageTypeAgentHeartbeat = "agent.heartbeat"
	MessageTypeTaskDispatch   = "task.dispatch"
	MessageTypeTaskEvent      = "task.event"
	MessageTypeError          = "error"
)

type Message struct {
	Type         string          `json:"type"`
	TaskGuid     string          `json:"taskGuid,omitempty"`
	HostGuid     string          `json:"hostGuid,omitempty"`
	AgentGuid    string          `json:"agentGuid,omitempty"`
	Payload      json.RawMessage `json:"payload,omitempty"`
	Status       string          `json:"status,omitempty"`
	Message      string          `json:"message,omitempty"`
	Data         json.RawMessage `json:"data,omitempty"`
	Result       json.RawMessage `json:"result,omitempty"`
	ErrorMessage string          `json:"errorMessage,omitempty"`
	Timestamp    int64           `json:"timestamp"`
}

type TaskDispatchPayload struct {
	Type           string          `json:"type"`
	Payload        json.RawMessage `json:"payload"`
	TimeoutSeconds int             `json:"timeoutSeconds"`
}

func NewConnectedMessage(agentGuid string, hostGuid string) Message {
	return Message{
		Type:      MessageTypeAgentConnected,
		AgentGuid: agentGuid,
		HostGuid:  hostGuid,
		Message:   "agent connected",
		Timestamp: nowMilli(),
	}
}

func NewTaskDispatchMessage(task domains.Task, agentGuid string) (Message, error) {
	rawPayload := json.RawMessage(task.Payload)
	if len(rawPayload) == 0 || !json.Valid(rawPayload) {
		rawPayload = json.RawMessage(`{}`)
	}
	payload, err := json.Marshal(TaskDispatchPayload{
		Type:           task.Type,
		Payload:        rawPayload,
		TimeoutSeconds: task.TimeoutSeconds,
	})
	if err != nil {
		return Message{}, err
	}
	return Message{
		Type:      MessageTypeTaskDispatch,
		TaskGuid:  task.Guid,
		HostGuid:  task.HostGuid,
		AgentGuid: agentGuid,
		Payload:   payload,
		Timestamp: nowMilli(),
	}, nil
}

func NewErrorMessage(message string) Message {
	return Message{
		Type:         MessageTypeError,
		ErrorMessage: message,
		Timestamp:    nowMilli(),
	}
}

func nowMilli() int64 {
	return time.Now().UnixMilli()
}

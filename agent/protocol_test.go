package agent

import (
	"encoding/json"
	"testing"

	"docker-go/domains"
)

func TestNewTaskDispatchMessage(t *testing.T) {
	task := domains.Task{
		HostGuid:       "host-1",
		Type:           "agent.ping",
		Payload:        `{"hello":"world"}`,
		TimeoutSeconds: 30,
	}
	task.Guid = "task-1"

	msg, err := NewTaskDispatchMessage(task, "agent-1")
	if err != nil {
		t.Fatalf("NewTaskDispatchMessage returned error: %v", err)
	}
	if msg.Type != MessageTypeTaskDispatch {
		t.Fatalf("message type = %q, want %q", msg.Type, MessageTypeTaskDispatch)
	}
	if msg.TaskGuid != task.Guid {
		t.Fatalf("task guid = %q, want %q", msg.TaskGuid, task.Guid)
	}

	var payload TaskDispatchPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		t.Fatalf("unmarshal dispatch payload: %v", err)
	}
	if payload.Type != task.Type {
		t.Fatalf("payload type = %q, want %q", payload.Type, task.Type)
	}
	if payload.TimeoutSeconds != task.TimeoutSeconds {
		t.Fatalf("timeout = %d, want %d", payload.TimeoutSeconds, task.TimeoutSeconds)
	}
	if string(payload.Payload) != task.Payload {
		t.Fatalf("payload = %s, want %s", payload.Payload, task.Payload)
	}
}

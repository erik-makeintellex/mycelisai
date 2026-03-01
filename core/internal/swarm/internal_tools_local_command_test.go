package swarm

import (
	"context"
	"encoding/json"
	"testing"
)

func TestInternalToolRegistry_LocalCommand_Registered(t *testing.T) {
	r := NewInternalToolRegistry(InternalToolDeps{})
	if r.Get("local_command") == nil {
		t.Fatal("expected local_command tool to be registered")
	}
}

func TestInternalToolRegistry_LocalCommand_Success(t *testing.T) {
	t.Setenv("MYCELIS_LOCAL_COMMAND_ALLOWLIST", "hostname")
	r := NewInternalToolRegistry(InternalToolDeps{})

	out, err := r.handleLocalCommand(context.Background(), map[string]any{
		"command": "hostname",
	})
	if err != nil {
		t.Fatalf("handleLocalCommand: %v", err)
	}

	var payload struct {
		Status  string `json:"status"`
		Command string `json:"command"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if payload.Status != "success" {
		t.Fatalf("status=%q, want success", payload.Status)
	}
	if payload.Command != "hostname" {
		t.Fatalf("command=%q, want hostname", payload.Command)
	}
}

func TestInternalToolRegistry_LocalCommand_Disallowed(t *testing.T) {
	t.Setenv("MYCELIS_LOCAL_COMMAND_ALLOWLIST", "hostname")
	r := NewInternalToolRegistry(InternalToolDeps{})
	if _, err := r.handleLocalCommand(context.Background(), map[string]any{
		"command": "whoami",
	}); err == nil {
		t.Fatal("expected error for disallowed command")
	}
}

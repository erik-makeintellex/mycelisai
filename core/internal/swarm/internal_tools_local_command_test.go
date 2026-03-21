package swarm

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestInternalToolRegistry_LocalCommand_Registered(t *testing.T) {
	r := NewInternalToolRegistry(InternalToolDeps{})
	if r.Get("local_command") == nil {
		t.Fatal("expected local_command tool to be registered")
	}
	if r.Get("save_cached_image") == nil {
		t.Fatal("expected save_cached_image tool to be registered")
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

func TestInternalToolRegistry_LocalCommand_RejectsShellSnippet(t *testing.T) {
	t.Setenv("MYCELIS_LOCAL_COMMAND_ALLOWLIST", "hostname")
	r := NewInternalToolRegistry(InternalToolDeps{})

	if _, err := r.handleLocalCommand(context.Background(), map[string]any{
		"command": "echo 'Hello, this is a simple greeting letter.'",
	}); err == nil || !strings.Contains(err.Error(), "shell snippets are not allowed") {
		t.Fatalf("expected shell snippet guidance, got %v", err)
	}
}

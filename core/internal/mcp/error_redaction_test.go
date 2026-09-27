package mcp

import (
	"strings"
	"testing"

	mcplib "github.com/mark3labs/mcp-go/mcp"
)

func TestCallToolResultErrorRedactsSecrets(t *testing.T) {
	text := "denied: Authorization: Bearer abc.def-ghi; OPENAI_API_KEY=sk-live-123 password: hunter2 " +
		"GITHUB_TOKEN='ghp_x' fetch https://user:pa55@db.internal/x failed; path=/workspace/a.md"
	err := CallToolResultError("fetch", mcplib.NewToolResultError(text))
	if err == nil {
		t.Fatal("isError result must be a failure")
	}
	msg := err.Error()
	for _, secret := range []string{"abc.def-ghi", "sk-live-123", "hunter2", "ghp_x", "user:pa55"} {
		if strings.Contains(msg, secret) {
			t.Fatalf("secret %q leaked: %s", secret, msg)
		}
	}
	for _, kept := range []string{"denied", "Bearer [REDACTED]", "OPENAI_API_KEY=[REDACTED]", "https://[REDACTED]@db.internal/x", "path=/workspace/a.md"} {
		if !strings.Contains(msg, kept) {
			t.Fatalf("message lost %q: %s", kept, msg)
		}
	}
}

func TestCallToolResultErrorCapsLength(t *testing.T) {
	err := CallToolResultError("probe", mcplib.NewToolResultError(strings.Repeat("é", 5000)))
	var toolErr *ToolResultError
	if e, ok := err.(*ToolResultError); ok {
		toolErr = e
	}
	if toolErr == nil {
		t.Fatalf("err = %v", err)
	}
	if len(toolErr.Message) > maxToolErrorBytes+len(" ... [truncated]") || !strings.HasSuffix(toolErr.Message, "[truncated]") {
		t.Fatalf("message length = %d, want capped with truncation marker", len(toolErr.Message))
	}
	if !strings.HasPrefix(toolErr.Message, "é") || strings.ContainsRune(toolErr.Message, '�') {
		t.Fatal("cap split a multi-byte rune")
	}
	short := CallToolResultError("probe", mcplib.NewToolResultError("plain failure"))
	if !strings.HasSuffix(short.Error(), "plain failure") {
		t.Fatalf("short message altered: %v", short)
	}
}

package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	mcplib "github.com/mark3labs/mcp-go/mcp"
)

func TestFinishMCPToolCall_IsErrorResultIsFailure(t *testing.T) {
	s := newTestServer()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/mcp/servers/x/tools/write_file/call", nil)
	result := mcplib.NewToolResultError("access denied: /etc/passwd is outside allowed directories (Authorization: Bearer s3cr3t-tok)")

	s.finishMCPToolCall(rr, req, uuid.New(), "filesystem", "write_file", map[string]any{"path": "/etc/passwd"}, result)

	assertStatus(t, rr, http.StatusBadGateway)
	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("failure body is not JSON: %v (%s)", err, rr.Body.String())
	}
	if body["is_error"] != true {
		t.Fatalf("is_error = %v, want true", body["is_error"])
	}
	errText, _ := body["error"].(string)
	if !strings.Contains(errText, "tool call failed") || !strings.Contains(errText, "access denied") {
		t.Fatalf("error = %q, want tool failure with server text", errText)
	}
	if strings.Contains(rr.Body.String(), "s3cr3t-tok") {
		t.Fatalf("failure body leaked a bearer token: %s", rr.Body.String())
	}
	if _, ok := body["result"]; ok {
		t.Fatalf("failure body echoes the raw MCP result: %v", body)
	}
	if _, ok := body["execution_summary"]; ok {
		t.Fatalf("failed MCP call must not carry a completed execution summary: %v", body)
	}
}

func TestFinishMCPToolCall_SuccessResultIsCompleted(t *testing.T) {
	s := newTestServer()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/mcp/servers/x/tools/read_file/call", nil)

	s.finishMCPToolCall(rr, req, uuid.New(), "filesystem", "read_file", map[string]any{"path": "README.md"}, mcplib.NewToolResultText("hello"))

	assertStatus(t, rr, http.StatusOK)
	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if _, ok := body["execution_summary"]; !ok {
		t.Fatalf("success body missing execution_summary: %v", body)
	}
	if body["isError"] == true || body["is_error"] == true {
		t.Fatalf("success body flagged as error: %v", body)
	}
}

func TestMCPToolCallExchangeInput_StatusFollowsFailure(t *testing.T) {
	id := uuid.New()
	if got := mcpToolCallExchangeInput(id, "filesystem", "write_file", "denied", true, nil, nil).Status; got != "failed" {
		t.Fatalf("failed exchange status = %q, want failed", got)
	}
	if got := mcpToolCallExchangeInput(id, "filesystem", "read_file", "ok", false, nil, nil).Status; got != "completed" {
		t.Fatalf("completed exchange status = %q, want completed", got)
	}
}

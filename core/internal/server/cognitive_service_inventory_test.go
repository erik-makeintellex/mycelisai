package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/mycelis/core/pkg/protocol"
)

func TestIsServiceInventoryQuestionDistinguishesRawToolDebug(t *testing.T) {
	if !isServiceInventoryQuestion("list of services?") {
		t.Fatal("expected list of services to use user-facing inventory path")
	}
	if !isServiceInventoryQuestion("services?") {
		t.Fatal("expected short services ask to use user-facing inventory path")
	}
	if !isServiceInventoryQuestion("what can Soma use right now?") {
		t.Fatal("expected capability phrasing to use user-facing inventory path")
	}
	if isServiceInventoryQuestion("show internal tool names") {
		t.Fatal("raw tool-name request should remain available for explicit technical inventory")
	}
	if isServiceInventoryQuestion("debug MCP status") {
		t.Fatal("debug MCP request should remain available for technical path")
	}
}

func TestHandleChat_ServiceInventoryUsesUserLanguage(t *testing.T) {
	opt, mock := withMCPDB(t)
	mock.ExpectQuery("SELECT id, name, transport, command, args, env, url, headers, status, error_message, created_at, updated_at FROM mcp_servers").
		WillReturnRows(sqlmock.NewRows(mcpServerColumns()).
			AddRow(uuid.New(), "filesystem", "stdio", "filesystem", []byte(`[]`), []byte(`{}`), "", []byte(`{}`), "connected", nil, time.Now(), time.Now()).
			AddRow(uuid.New(), "fetch", "stdio", "fetch", []byte(`[]`), []byte(`{}`), "", []byte(`{}`), "error", "failed initialization", time.Now(), time.Now()))
	s := newTestServer(opt)

	reqBody := bytes.NewBufferString(`{"messages":[{"role":"user","content":"list of services?"}]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat", reqBody)
	rr := httptest.NewRecorder()

	http.HandlerFunc(s.HandleChat).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d body=%s", rr.Code, http.StatusOK, rr.Body.String())
	}
	payload := decodeChatPayloadFromResponse(t, rr.Body.Bytes())
	text := payload.Text
	for _, want := range []string{
		"Here is what Soma can use right now:",
		"Available",
		"Soma workspace",
		"System status checks",
		"Needs attention",
		"Web and URL access is not available right now.",
		"Open Settings -> Connected Tools and repair Fetch.",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("answer missing %q:\n%s", want, text)
		}
	}
	for _, raw := range []string{"list_teams", "delegate_task", "publish_signal", "read_text_file", "filesystem (connected)", "fetch (error)"} {
		if strings.Contains(text, raw) {
			t.Fatalf("answer leaked raw inventory %q:\n%s", raw, text)
		}
	}
	if payload.Provenance == nil || payload.Provenance.ResolvedIntent != "service_inventory" {
		t.Fatalf("provenance = %#v, want service_inventory", payload.Provenance)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func decodeChatPayloadFromResponse(t *testing.T, body []byte) protocol.ChatResponsePayload {
	t.Helper()
	var resp protocol.APIResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	data, _ := json.Marshal(resp.Data)
	var envelope protocol.CTSEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	var payload protocol.ChatResponsePayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	return payload
}

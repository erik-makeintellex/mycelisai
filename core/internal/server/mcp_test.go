package server

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/mycelis/core/pkg/protocol"
)

func TestHandleMCPList_NilSubsystem(t *testing.T) {
	s := newTestServer()
	rr := doRequest(t, http.HandlerFunc(s.handleMCPList), "GET", "/api/v1/mcp/servers", "")
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

func TestHandleMCPDelete_NilSubsystem(t *testing.T) {
	s := newTestServer()
	mux := setupMux(t, "DELETE /api/v1/mcp/servers/{id}", s.handleMCPDelete)
	rr := doRequest(t, mux, "DELETE", "/api/v1/mcp/servers/some-id", "")
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

func TestHandleMCPDelete_InvalidUUID(t *testing.T) {
	s := newTestServer(withMCPStubs())
	mux := setupMux(t, "DELETE /api/v1/mcp/servers/{id}", s.handleMCPDelete)
	rr := doRequest(t, mux, "DELETE", "/api/v1/mcp/servers/not-a-uuid", "")
	assertStatus(t, rr, http.StatusBadRequest)
}

func TestHandleMCPToolCall_NilSubsystem(t *testing.T) {
	s := newTestServer()
	mux := setupMux(t, "POST /api/v1/mcp/servers/{id}/tools/{tool}/call", s.handleMCPToolCall)
	rr := doRequest(t, mux, "POST", "/api/v1/mcp/servers/some-id/tools/some-tool/call", `{"arguments":{}}`)
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

func TestHandleMCPToolCall_InvalidUUID(t *testing.T) {
	s := newTestServer(withMCPStubs())
	mux := setupMux(t, "POST /api/v1/mcp/servers/{id}/tools/{tool}/call", s.handleMCPToolCall)
	rr := doRequest(t, mux, "POST", "/api/v1/mcp/servers/not-a-uuid/tools/test/call", `{"arguments":{}}`)
	assertStatus(t, rr, http.StatusBadRequest)
}

func TestHandleMCPToolCall_UnavailableServerReturnsHelpfulError(t *testing.T) {
	s := newTestServer(withMCPStubs())
	mux := setupMux(t, "POST /api/v1/mcp/servers/{id}/tools/{tool}/call", s.handleMCPToolCall)
	rr := doRequest(t, mux, "POST", "/api/v1/mcp/servers/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa/tools/read_text_file/call", `{"arguments":{"path":"README.md"}}`)
	assertStatus(t, rr, http.StatusBadGateway)
	if body := rr.Body.String(); !strings.Contains(body, "tool call failed") || !strings.Contains(body, "not found in pool") {
		t.Fatalf("error body = %q, want helpful tool call pool failure", body)
	}
}

func TestDecodeMCPToolCallArguments_AcceptsWrapperAndDirectPayload(t *testing.T) {
	wrapped, err := decodeMCPToolCallArguments(strings.NewReader(`{"arguments":{"path":"README.md"}}`))
	if err != nil {
		t.Fatalf("decode wrapped args: %v", err)
	}
	if wrapped["path"] != "README.md" {
		t.Fatalf("wrapped args = %#v, want path", wrapped)
	}

	direct, err := decodeMCPToolCallArguments(strings.NewReader(`{"path":"README.md"}`))
	if err != nil {
		t.Fatalf("decode direct args: %v", err)
	}
	if direct["path"] != "README.md" {
		t.Fatalf("direct args = %#v, want path", direct)
	}
}

func TestDecodeMCPToolCallArguments_RejectsNonObjectArguments(t *testing.T) {
	_, err := decodeMCPToolCallArguments(strings.NewReader(`{"arguments":"README.md"}`))
	if err == nil || !strings.Contains(err.Error(), "arguments must be an object") {
		t.Fatalf("err = %v, want object validation", err)
	}
}

func TestHandleMCPToolsList_NilSubsystem(t *testing.T) {
	s := newTestServer()
	rr := doRequest(t, http.HandlerFunc(s.handleMCPToolsList), "GET", "/api/v1/mcp/tools", "")
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

func TestMCPToolCallResponse_AddsExecutionSummaryToObjectResult(t *testing.T) {
	summary := buildMCPToolCallExecutionSummary("filesystem", "read_file", "Read workspace brief.", "exchange-1")
	response := mcpToolCallResponse(map[string]any{
		"content": []any{map[string]any{"type": "text", "text": "hello"}},
	}, summary, "exchange-1")

	object, ok := response.(map[string]any)
	if !ok {
		t.Fatalf("response = %T, want map", response)
	}
	if object["exchange_item_id"] != "exchange-1" {
		t.Fatalf("exchange_item_id = %v", object["exchange_item_id"])
	}
	gotSummary, ok := object["execution_summary"].(*protocol.ExecutionSummary)
	if !ok {
		t.Fatalf("execution_summary = %T", object["execution_summary"])
	}
	if gotSummary.Proof.RunClass != protocol.ExecutionRunClassNoRun || gotSummary.Proof.NoRunReason == "" {
		t.Fatalf("proof no-run classification = %+v", gotSummary.Proof)
	}
	if gotSummary.Proof.ExchangeItemID != "exchange-1" {
		t.Fatalf("exchange proof item = %q", gotSummary.Proof.ExchangeItemID)
	}
}

func TestHandleMCPList_HappyPath(t *testing.T) {
	opt, mock := withMCPDB(t)
	s := newTestServer(opt)
	now := time.Now()

	serverID := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	serverUUID := uuid.MustParse(serverID)
	mock.ExpectQuery("SELECT .+ FROM mcp_servers").
		WillReturnRows(sqlmock.NewRows(mcpServerColumns()).
			AddRow(serverID, "filesystem", "stdio", "npx", `[]`, `{}`, "", `{}`, "connected", nil, now, now))
	mock.ExpectQuery("SELECT .+ FROM mcp_tools").
		WithArgs(serverUUID).
		WillReturnRows(sqlmock.NewRows(mcpToolColumns()).
			AddRow("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb", serverID, "read_file", "Read file", []byte(`{}`)))

	rr := doRequest(t, http.HandlerFunc(s.handleMCPList), "GET", "/api/v1/mcp/servers", "")
	assertStatus(t, rr, http.StatusOK)
}

func TestHandleMCPList_EmptyDB(t *testing.T) {
	opt, mock := withMCPDB(t)
	s := newTestServer(opt)

	mock.ExpectQuery("SELECT .+ FROM mcp_servers").
		WillReturnRows(sqlmock.NewRows(mcpServerColumns()))

	rr := doRequest(t, http.HandlerFunc(s.handleMCPList), "GET", "/api/v1/mcp/servers", "")
	assertStatus(t, rr, http.StatusOK)
}

func TestHandleMCPDelete_HappyPath(t *testing.T) {
	opt, mock := withMCPDB(t)
	s := newTestServer(opt)
	serverUUID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")

	mock.ExpectExec("DELETE FROM mcp_servers WHERE id = \\$1").
		WithArgs(serverUUID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mux := setupMux(t, "DELETE /api/v1/mcp/servers/{id}", s.handleMCPDelete)
	rr := doRequest(t, mux, "DELETE", "/api/v1/mcp/servers/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", "")
	assertStatus(t, rr, http.StatusOK)
}

func TestHandleMCPToolsList_HappyPath(t *testing.T) {
	opt, mock := withMCPDB(t)
	s := newTestServer(opt)

	mock.ExpectQuery("SELECT .+ FROM mcp_tools .+ JOIN mcp_servers").
		WillReturnRows(sqlmock.NewRows(mcpToolWithServerColumns()).
			AddRow("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb", "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", "filesystem", "read_file", "Read file", []byte(`{}`)))

	rr := doRequest(t, http.HandlerFunc(s.handleMCPToolsList), "GET", "/api/v1/mcp/tools", "")
	assertStatus(t, rr, http.StatusOK)
}

func TestHandleMCPInstall_Forbidden(t *testing.T) {
	s := newTestServer()
	mux := http.NewServeMux()
	s.RegisterRoutes(mux)

	rr := doRequest(t, mux, "POST", "/api/v1/mcp/install", `{"name":"test"}`)
	assertStatus(t, rr, http.StatusForbidden)
}

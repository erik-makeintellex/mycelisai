package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/mycelis/core/internal/exchange"
	"github.com/mycelis/core/internal/mcp"
)

// RED: MCP redaction hardening (MCPS-QA and MCPA-QA evidence).

var redQASecrets = []string{"zzqa-json-secret-1", "zzqa-hdr-secret-2", "zzqa-camel-secret-3", "zzqa-bare-ghp-4"}

const redQALeakyText = `denied {"token": "zzqa-json-secret-1"} X-Api-Key: zzqa-hdr-secret-2 apiKey=zzqa-camel-secret-3 invalid credential ghp_zzqa-bare-ghp-4`

func redAssertNoLeak(t *testing.T, where, text string, secrets ...string) {
	t.Helper()
	for _, s := range secrets {
		if strings.Contains(text, s) {
			t.Fatalf("%s leaked %q: %s", where, s, text)
		}
	}
}

func TestRedSensitiveArgumentKeyStyles(t *testing.T) {
	for _, key := range []string{
		"apiKey", "APIKey", "accessToken", "privateKey", "clientSecret", "refreshToken",
		"api-key", "X-Api-Key", "Client_Secret", "AUTH_TOKEN", "secret-key", "passwd",
		"credentials", "Authorization", "password", "token",
	} {
		if !isSensitiveMCPArgumentKey(key) {
			t.Errorf("key %q not treated as sensitive", key)
		}
	}
	for _, key := range []string{"path", "monkey", "tokenizer", "max_tokens", "keyboard", "sort_order", "content", "keys"} {
		if isSensitiveMCPArgumentKey(key) {
			t.Errorf("key %q wrongly treated as sensitive", key)
		}
	}
}

func TestRedArgumentsRedactNestedCamelAndText(t *testing.T) {
	args := map[string]any{
		"path": "workspace/a.md",
		"config": map[string]any{
			"apiKey": "redarg-secret-1",
			"items":  []any{map[string]any{"clientSecret": "redarg-secret-2"}, "Authorization: Basic cmVkYXJnOnNlY3JldDM="},
		},
		"note": "use ghp_redarg-secret-4 please",
	}
	raw, _ := json.Marshal(redactMCPToolArguments(args))
	redAssertNoLeak(t, "retained args", string(raw), "redarg-secret-1", "redarg-secret-2", "cmVkYXJnOnNlY3JldDM", "redarg-secret-4")
	if !strings.Contains(string(raw), "workspace/a.md") || !strings.Contains(string(raw), "use ghp_[REDACTED] please") {
		t.Fatalf("non-secret content lost: %s", raw)
	}
	if args["config"].(map[string]any)["apiKey"] != "redarg-secret-1" {
		t.Fatal("the tool's own arguments were mutated")
	}
}

// redExchangeServer wires an Exchange whose item insert is captured.
func redExchangeServer(t *testing.T) (*AdminServer, *mcpsCapture, sqlmock.Sqlmock) {
	t.Helper()
	exDB, exMock := mcpsNewMock(t)
	h := &mcpsH{t: t, exMock: exMock}
	payload := h.expectExchange()
	s := newTestServer(func(s *AdminServer) { s.Exchange = exchange.NewService(exDB, nil, nil) })
	return s, payload, exMock
}

func TestRedToolResultRedactedInExchangeOnly(t *testing.T) {
	s, payload, exMock := redExchangeServer(t)
	result := mcplib.NewToolResultText("config: " + redQALeakyText)
	result.StructuredContent = map[string]any{"accessToken": "redres-secret-5", "rows": []any{"sk-proj-redresSecret6abcdefgh"}}
	req := httptest.NewRequest(http.MethodPost, "/x", nil)
	req = req.WithContext(context.WithValue(req.Context(), ctxKeyIdentity, standardUserIdentity()))
	rr := httptest.NewRecorder()
	s.finishMCPToolCall(rr, req, uuid.New(), "filesystem", "read_text_file", map[string]any{"path": "a"}, result)
	assertStatus(t, rr, http.StatusOK)
	if err := exMock.ExpectationsWereMet(); err != nil {
		t.Fatalf("exchange not retained: %v", err)
	}
	retained := payload.value()
	redAssertNoLeak(t, "exchange payload", retained, append(redQASecrets, "redres-secret-5", "redresSecret6")...)
	if !strings.Contains(retained, "[REDACTED]") || !strings.Contains(retained, "config: denied") {
		t.Fatalf("retained result lost its redacted shape: %s", retained)
	}
	// The caller asked for the output: the HTTP response stays raw.
	if !strings.Contains(rr.Body.String(), "zzqa-json-secret-1") || !strings.Contains(rr.Body.String(), "redres-secret-5") {
		t.Fatalf("HTTP response was redacted: %s", rr.Body.String())
	}
}

func TestRedToolResultIsErrorPathUnchanged(t *testing.T) {
	s, payload, _ := redExchangeServer(t)
	req := httptest.NewRequest(http.MethodPost, "/x", nil)
	rr := httptest.NewRecorder()
	s.finishMCPToolCall(rr, req, uuid.New(), "filesystem", "write_file", nil, mcplib.NewToolResultError(redQALeakyText))
	assertStatus(t, rr, http.StatusBadGateway)
	redAssertNoLeak(t, "isError response", rr.Body.String(), redQASecrets...)
	redAssertNoLeak(t, "isError exchange", payload.value(), redQASecrets...)
	if !strings.Contains(payload.value(), `"is_error":true`) {
		t.Fatalf("isError retention shape changed: %s", payload.value())
	}
}

// Port of TestZzQaMCPA_ToolCall502RedactionForms.
func TestRedToolCall502RedactionForms(t *testing.T) {
	fixture := mcpserver.NewMCPServer("red-failing", "0")
	fixture.AddTool(mcplib.NewTool("write_file"), func(context.Context, mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
		return nil, errors.New(redQALeakyText)
	})
	url := newLocalHTTPTestServer(t, mcpserver.NewStreamableHTTPServer(fixture)).URL
	opt, mock := mcpaSharedDB(t)
	s := newTestServer(opt)
	id := uuid.MustParse(mcpaServerID)
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM mcp_tools").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO mcp_tools").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	mock.ExpectExec("UPDATE mcp_servers").WillReturnResult(sqlmock.NewResult(0, 1))
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := s.MCPPool.Connect(ctx, mcp.ServerConfig{ID: id, Name: "filesystem", Transport: "sse", URL: url}); err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(s.MCPPool.ShutdownAll)
	now := time.Now()
	mock.ExpectQuery("SELECT .+ FROM mcp_servers").WithArgs(id).WillReturnRows(sqlmock.NewRows(mcpServerColumns()).
		AddRow(mcpaServerID, "filesystem", "sse", "", `[]`, `{}`, url, `{}`, "connected", nil, now, now))
	mock.ExpectQuery("SELECT .+ FROM mcp_tools").WithArgs(id).WillReturnRows(sqlmock.NewRows(mcpToolColumns()).
		AddRow(uuid.NewString(), mcpaServerID, "write_file", "", []byte(`{}`)))
	mock.ExpectExec("INSERT INTO log_entries").WillReturnResult(sqlmock.NewResult(0, 1))
	mux := setupMux(t, mcpsCallRoute, s.handleMCPToolCall)
	rr := doAuthenticatedRequestAs(t, mux, "POST", "/api/v1/mcp/servers/"+mcpaServerID+"/tools/write_file/call", `{"path":"a"}`, mcpsWebAdmin())
	assertStatus(t, rr, http.StatusBadGateway)
	if !json.Valid(rr.Body.Bytes()) {
		t.Fatalf("502 body not valid JSON: %s", rr.Body.String())
	}
	redAssertNoLeak(t, "tool_call_502", rr.Body.String(), redQASecrets...)
	if !strings.Contains(rr.Body.String(), "invalid credential") {
		t.Fatalf("502 lost the non-secret error text: %s", rr.Body.String())
	}
}

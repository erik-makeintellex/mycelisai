package server

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func assertNoMCPSecretLeak(t *testing.T, body string) {
	t.Helper()
	if strings.Contains(body, "live-secret") || strings.Contains(body, "Bearer live-secret") {
		t.Fatalf("MCP response leaked secret config: %s", body)
	}
	if !strings.Contains(body, redactedMCPSecretValue) {
		t.Fatalf("MCP response did not include redaction marker: %s", body)
	}
}

func TestHandleMCPList_RedactsEnvAndHeaders(t *testing.T) {
	opt, mock := withMCPDB(t)
	s := newTestServer(opt)
	now := time.Now()
	serverID := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	serverUUID := uuid.MustParse(serverID)

	mock.ExpectQuery("SELECT .+ FROM mcp_servers").
		WillReturnRows(sqlmock.NewRows(mcpServerColumns()).
			AddRow(serverID, "brave-search", "stdio", "npx", `[]`, `{"BRAVE_API_KEY":"live-secret"}`, "", `{"Authorization":"Bearer live-secret"}`, "connected", nil, now, now))
	mock.ExpectQuery("SELECT .+ FROM mcp_tools").
		WithArgs(serverUUID).
		WillReturnRows(sqlmock.NewRows(mcpToolColumns()))

	rr := doRequest(t, http.HandlerFunc(s.handleMCPList), "GET", "/api/v1/mcp/servers", "")
	assertStatus(t, rr, http.StatusOK)
	assertNoMCPSecretLeak(t, rr.Body.String())
}

func TestHandleMCPLibraryInstall_RedactsEnvAndHeaders(t *testing.T) {
	s := newSecretFetchInstallServer(t)

	rr := doAuthenticatedRequest(t, http.HandlerFunc(s.handleMCPLibraryInstall), "POST", "/api/v1/mcp/library/install", `{"name":"fetch"}`)
	assertStatus(t, rr, http.StatusOK)
	assertNoMCPSecretLeak(t, rr.Body.String())
}

func TestHandleMCPLibraryApply_RedactsEnvAndHeaders(t *testing.T) {
	s := newSecretFetchInstallServer(t)

	rr := doAuthenticatedRequest(t, http.HandlerFunc(s.handleMCPLibraryApply), "POST", "/api/v1/mcp/library/apply", `{"name":"fetch"}`)
	assertStatus(t, rr, http.StatusOK)
	assertNoMCPSecretLeak(t, rr.Body.String())
}

// newSecretFetchInstallServer wires a live fetch install whose stored row
// carries secret env and headers (MCPA: audit first, real connect).
func newSecretFetchInstallServer(t *testing.T) *AdminServer {
	t.Helper()
	url := mcpaFixtureURL(t)
	opt, mock := mcpaSharedDB(t)
	s := newTestServer(opt, mcpaLiveLibrary(url))
	mcpaExpectLiveInstall(t, s, mock, "fetch", sqlmock.AnyArg(), url, `{"FETCH_TOKEN":"live-secret"}`, `{"Authorization":"Bearer live-secret"}`, false)
	return s
}

// MCPS D7: nested keys, arrays and header-style keys are redacted in a copy.
func TestMCPSRedactToolArgumentsDeepCopy(t *testing.T) {
	args := map[string]any{
		"path":     "notes.md",
		"headers":  map[string]any{"Authorization": "Bearer abc", "X-Api-Key": "k1", "accept": "json"},
		"items":    []any{"https://u:p@h/x", map[string]any{"client_secret": "s1"}, 7},
		"Password": "pw", "credential": "c1", "refresh_token": "r1",
	}
	got := redactMCPToolArguments(args)
	headers := got["headers"].(map[string]any)
	items := got["items"].([]any)
	if got["path"] != "notes.md" || headers["accept"] != "json" || items[2] != 7 {
		t.Fatalf("non-secret values changed: %v", got)
	}
	for _, v := range []any{headers["Authorization"], headers["X-Api-Key"], items[1].(map[string]any)["client_secret"], got["Password"], got["credential"], got["refresh_token"]} {
		if v != redactedMCPArgumentValue {
			t.Fatalf("value %v not redacted: %v", v, got)
		}
	}
	if strings.Contains(items[0].(string), "u:p@") {
		t.Fatalf("URL userinfo kept: %v", items[0])
	}
	if args["Password"] != "pw" || args["headers"].(map[string]any)["X-Api-Key"] != "k1" {
		t.Fatalf("redaction mutated the arguments the tool receives: %v", args)
	}
	if redactMCPToolArguments(nil) != nil {
		t.Fatal("nil arguments must stay nil")
	}
}

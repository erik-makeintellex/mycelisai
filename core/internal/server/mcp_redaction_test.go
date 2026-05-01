package server

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/mycelis/core/internal/mcp"
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
	opt, mock := withMCPDB(t)
	s := newTestServer(opt, withSecretFetchLibrary())
	expectSecretFetchInstall(mock)

	rr := doRequest(t, http.HandlerFunc(s.handleMCPLibraryInstall), "POST", "/api/v1/mcp/library/install", `{"name":"fetch"}`)
	assertStatus(t, rr, http.StatusOK)
	assertNoMCPSecretLeak(t, rr.Body.String())
}

func TestHandleMCPLibraryApply_RedactsEnvAndHeaders(t *testing.T) {
	opt, mock := withMCPDB(t)
	s := newTestServer(opt, withSecretFetchLibrary())
	expectSecretFetchInstall(mock)

	rr := doRequest(t, http.HandlerFunc(s.handleMCPLibraryApply), "POST", "/api/v1/mcp/library/apply", `{"name":"fetch"}`)
	assertStatus(t, rr, http.StatusOK)
	assertNoMCPSecretLeak(t, rr.Body.String())
}

func withSecretFetchLibrary() func(*AdminServer) {
	return func(s *AdminServer) {
		s.MCPLibrary = &mcp.Library{Categories: []mcp.LibraryCategory{{
			Name:    "Default",
			Servers: []mcp.LibraryEntry{{Name: "fetch", Transport: "unsupported"}},
		}}}
	}
}

func expectSecretFetchInstall(mock sqlmock.Sqlmock) {
	now := time.Now()
	mock.ExpectQuery("INSERT INTO mcp_servers").
		WithArgs("fetch", "unsupported", "", sqlmock.AnyArg(), sqlmock.AnyArg(), "", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows(mcpServerColumns()).
			AddRow("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", "fetch", "unsupported", "", `[]`, `{"FETCH_TOKEN":"live-secret"}`, "", `{"Authorization":"Bearer live-secret"}`, "installed", nil, now, now))
	mock.ExpectExec("UPDATE mcp_servers").
		WithArgs("error", sqlmock.AnyArg(), "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("SELECT .+ FROM mcp_tools").
		WithArgs("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa").
		WillReturnRows(sqlmock.NewRows(mcpToolColumns()))
}

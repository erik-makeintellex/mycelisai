package server

import (
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/mcp"
)

func TestHandleMCPLibraryInstall_NilSubsystem(t *testing.T) {
	s := newTestServer()
	rr := doRequest(t, http.HandlerFunc(s.handleMCPLibraryInstall), "POST", "/api/v1/mcp/library/install", `{"name":"test"}`)
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

func TestHandleMCPLibraryInstall_MissingName(t *testing.T) {
	s := newTestServer(withMCPStubs(), func(s *AdminServer) {
		s.MCPLibrary = &mcp.Library{Categories: []mcp.LibraryCategory{}}
	})
	rr := doRequest(t, http.HandlerFunc(s.handleMCPLibraryInstall), "POST", "/api/v1/mcp/library/install", `{"env":{}}`)
	assertStatus(t, rr, http.StatusBadRequest)
}

func TestHandleMCPLibraryInstall_NotFoundInLibrary(t *testing.T) {
	s := newTestServer(withMCPStubs(), func(s *AdminServer) {
		s.MCPLibrary = &mcp.Library{Categories: []mcp.LibraryCategory{}}
	})
	rr := doRequest(t, http.HandlerFunc(s.handleMCPLibraryInstall), "POST", "/api/v1/mcp/library/install", `{"name":"nonexistent"}`)
	assertStatus(t, rr, http.StatusNotFound)
}

func TestHandleMCPLibraryInstall_RemoteConfigReturnsApprovalBoundary(t *testing.T) {
	s := newTestServer(withMCPStubs(), func(s *AdminServer) {
		s.MCPLibrary = &mcp.Library{
			Categories: []mcp.LibraryCategory{
				{
					Name: "Default",
					Servers: []mcp.LibraryEntry{
						{Name: "remote-knowledge", Transport: "sse", URL: "https://mcp.example.com/sse", Tags: []string{"remote"}},
					},
				},
			},
		}
	})

	rr := doAuthenticatedRequest(t, http.HandlerFunc(s.handleMCPLibraryInstall), "POST", "/api/v1/mcp/library/install", `{"name":"remote-knowledge"}`)
	assertStatus(t, rr, http.StatusAccepted)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	if resp["requires_approval"] != true {
		t.Fatalf("requires_approval = %v, want true", resp["requires_approval"])
	}
}

func TestHandleMCPLibraryInstall_StandardLibraryGitHubReturnsApprovalBoundary(t *testing.T) {
	s := newTestServer(withMCPStubs(), func(s *AdminServer) {
		s.MCPLibrary = loadStandardMCPLibrary(t)
	})

	rr := doAuthenticatedRequest(t, http.HandlerFunc(s.handleMCPLibraryInstall), "POST", "/api/v1/mcp/library/install", `{"name":"github"}`)
	assertStatus(t, rr, http.StatusAccepted)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	if resp["requires_approval"] != true {
		t.Fatalf("requires_approval = %v, want true", resp["requires_approval"])
	}
	inspection, ok := resp["inspection"].(map[string]any)
	if !ok {
		t.Fatalf("expected inspection object, got %T", resp["inspection"])
	}
	if inspection["deployment_boundary"] != "external_saas" {
		t.Fatalf("deployment_boundary = %v, want external_saas", inspection["deployment_boundary"])
	}
}

func TestHandleMCPLibraryApply_RemoteConfigReturnsStructuredApprovalBoundary(t *testing.T) {
	s := newTestServer(withMCPStubs(), func(s *AdminServer) {
		s.MCPLibrary = &mcp.Library{
			Categories: []mcp.LibraryCategory{
				{
					Name: "Default",
					Servers: []mcp.LibraryEntry{
						{Name: "remote-knowledge", Transport: "sse", URL: "https://mcp.example.com/sse", Tags: []string{"remote"}},
					},
				},
			},
		}
	})
	mux := http.NewServeMux()
	s.RegisterRoutes(mux)

	rr := doAuthenticatedRequest(t, mux, "POST", "/api/v1/mcp/library/apply", `{"name":"remote-knowledge"}`)
	assertStatus(t, rr, http.StatusAccepted)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	if resp["status"] != "requires_approval" {
		t.Fatalf("status = %v, want requires_approval", resp["status"])
	}
	if resp["requires_approval"] != true {
		t.Fatalf("requires_approval = %v, want true", resp["requires_approval"])
	}
	if _, ok := resp["inspection"].(map[string]any); !ok {
		t.Fatalf("expected inspection object, got %T", resp["inspection"])
	}
}

func TestHandleMCPInstall_ForbiddenForStandardLibraryEntryToo(t *testing.T) {
	s := newTestServer(func(s *AdminServer) {
		s.MCPLibrary = loadStandardMCPLibrary(t)
	})
	mux := http.NewServeMux()
	s.RegisterRoutes(mux)

	rr := doRequest(t, mux, "POST", "/api/v1/mcp/install", `{"name":"filesystem","transport":"stdio","command":"npx"}`)
	assertStatus(t, rr, http.StatusForbidden)
}

func TestHandleMCPLibraryInstall_HappyPath(t *testing.T) {
	opt, mock := withMCPDB(t)
	s := newTestServer(opt, func(s *AdminServer) {
		s.MCPLibrary = &mcp.Library{
			Categories: []mcp.LibraryCategory{
				{
					Name: "Default",
					Servers: []mcp.LibraryEntry{
						{Name: "fetch", Transport: "unsupported"},
					},
				},
			},
		}
	})
	now := time.Now()

	mock.ExpectQuery("INSERT INTO mcp_servers").
		WithArgs("fetch", "unsupported", "", sqlmock.AnyArg(), sqlmock.AnyArg(), "", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows(mcpServerColumns()).
			AddRow("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", "fetch", "unsupported", "", `[]`, `{}`, "", `{}`, "installed", nil, now, now))
	mock.ExpectExec("UPDATE mcp_servers").
		WithArgs("error", sqlmock.AnyArg(), "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("SELECT .+ FROM mcp_tools").
		WithArgs("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa").
		WillReturnRows(sqlmock.NewRows(mcpToolColumns()))

	rr := doRequest(t, http.HandlerFunc(s.handleMCPLibraryInstall), "POST", "/api/v1/mcp/library/install", `{"name":"fetch"}`)
	assertStatus(t, rr, http.StatusOK)
}

func TestHandleMCPLibraryApply_HappyPath(t *testing.T) {
	opt, mock := withMCPDB(t)
	s := newTestServer(opt, func(s *AdminServer) {
		s.MCPLibrary = &mcp.Library{
			Categories: []mcp.LibraryCategory{
				{
					Name: "Default",
					Servers: []mcp.LibraryEntry{
						{Name: "fetch", Transport: "unsupported"},
					},
				},
			},
		}
	})
	now := time.Now()

	mock.ExpectQuery("INSERT INTO mcp_servers").
		WithArgs("fetch", "unsupported", "", sqlmock.AnyArg(), sqlmock.AnyArg(), "", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows(mcpServerColumns()).
			AddRow("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", "fetch", "unsupported", "", `[]`, `{}`, "", `{}`, "installed", nil, now, now))
	mock.ExpectExec("UPDATE mcp_servers").
		WithArgs("error", sqlmock.AnyArg(), "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("SELECT .+ FROM mcp_tools").
		WithArgs("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa").
		WillReturnRows(sqlmock.NewRows(mcpToolColumns()))

	rr := doRequest(t, http.HandlerFunc(s.handleMCPLibraryApply), "POST", "/api/v1/mcp/library/apply", `{"name":"fetch"}`)
	assertStatus(t, rr, http.StatusOK)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	if resp["status"] != "installed" {
		t.Fatalf("status = %v, want installed", resp["status"])
	}
	if resp["requires_approval"] != false {
		t.Fatalf("requires_approval = %v, want false", resp["requires_approval"])
	}
	if _, ok := resp["inspection"].(map[string]any); !ok {
		t.Fatalf("expected inspection object, got %T", resp["inspection"])
	}
}

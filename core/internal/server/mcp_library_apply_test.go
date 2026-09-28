package server

import (
	"net/http"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/mcp"
)

func TestHandleMCPLibraryInstall_NilSubsystem(t *testing.T) {
	s := newTestServer()
	rr := doAuthenticatedRequest(t, http.HandlerFunc(s.handleMCPLibraryInstall), "POST", "/api/v1/mcp/library/install", `{"name":"test"}`)
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

func TestHandleMCPLibraryInstall_MissingName(t *testing.T) {
	s := newTestServer(withMCPStubs(), func(s *AdminServer) {
		s.MCPLibrary = &mcp.Library{Categories: []mcp.LibraryCategory{}}
	})
	rr := doAuthenticatedRequest(t, http.HandlerFunc(s.handleMCPLibraryInstall), "POST", "/api/v1/mcp/library/install", `{"env":{}}`)
	assertStatus(t, rr, http.StatusBadRequest)
}

func TestHandleMCPLibraryInstall_NotFoundInLibrary(t *testing.T) {
	s := newTestServer(withMCPStubs(), func(s *AdminServer) {
		s.MCPLibrary = &mcp.Library{Categories: []mcp.LibraryCategory{}}
	})
	rr := doAuthenticatedRequest(t, http.HandlerFunc(s.handleMCPLibraryInstall), "POST", "/api/v1/mcp/library/install", `{"name":"nonexistent"}`)
	assertStatus(t, rr, http.StatusNotFound)
}

func withRemoteKnowledgeLibrary() func(*AdminServer) {
	return func(s *AdminServer) {
		s.MCPLibrary = &mcp.Library{Categories: []mcp.LibraryCategory{{
			Name: "Default",
			Servers: []mcp.LibraryEntry{
				{Name: "remote-knowledge", Transport: "sse", URL: "https://mcp.example.com/sse", Tags: []string{"remote"}},
			},
		}}}
	}
}

// MCPA D5: a require_approval entry is never answered 202; a caller without
// approvals:decide gets 403 admin_required naming the missing scope.
func TestHandleMCPLibraryInstall_RemoteConfigNeedsApprover(t *testing.T) {
	s := newTestServer(withMCPStubs(), withRemoteKnowledgeLibrary())
	rr := doAuthenticatedRequestAs(t, http.HandlerFunc(s.handleMCPLibraryInstall), "POST", "/api/v1/mcp/library/install", `{"name":"remote-knowledge"}`, adminWithScopes(scopeMCPConfigWrite))
	env := mcpsAssertBlocker(t, rr, http.StatusForbidden, codeAdminRequired)
	if env.Data["required_scope"] != scopeApprovalsDecide {
		t.Fatalf("required_scope = %v, want %s", env.Data["required_scope"], scopeApprovalsDecide)
	}
}

func TestHandleMCPLibraryInstall_StandardLibraryGitHubNeedsApprover(t *testing.T) {
	s := newTestServer(withMCPStubs(), func(s *AdminServer) {
		s.MCPLibrary = loadStandardMCPLibrary(t)
	})
	rr := doAuthenticatedRequestAs(t, http.HandlerFunc(s.handleMCPLibraryInstall), "POST", "/api/v1/mcp/library/install", `{"name":"github"}`, adminWithScopes(scopeMCPConfigWrite))
	env := mcpsAssertBlocker(t, rr, http.StatusForbidden, codeAdminRequired)
	if env.Data["required_scope"] != scopeApprovalsDecide {
		t.Fatalf("required_scope = %v, want %s", env.Data["required_scope"], scopeApprovalsDecide)
	}
}

func TestHandleMCPLibraryApply_RemoteConfigNeedsApproverThroughRoutes(t *testing.T) {
	s := newTestServer(withMCPStubs(), withRemoteKnowledgeLibrary())
	mux := http.NewServeMux()
	s.RegisterRoutes(mux)

	rr := doAuthenticatedRequestAs(t, mux, "POST", "/api/v1/mcp/library/apply", `{"name":"remote-knowledge"}`, adminWithScopes(scopeMCPConfigWrite))
	env := mcpsAssertBlocker(t, rr, http.StatusForbidden, codeAdminRequired)
	if env.Data["required_scope"] != scopeApprovalsDecide {
		t.Fatalf("required_scope = %v, want %s", env.Data["required_scope"], scopeApprovalsDecide)
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
	url := mcpaFixtureURL(t)
	opt, mock := mcpaSharedDB(t)
	s := newTestServer(opt, mcpaLiveLibrary(url))
	mcpaExpectLiveInstall(t, s, mock, "fetch", sqlmock.AnyArg(), url, `{}`, `{}`, false)

	rr := doAuthenticatedRequest(t, http.HandlerFunc(s.handleMCPLibraryInstall), "POST", "/api/v1/mcp/library/install", `{"name":"fetch"}`)
	assertStatus(t, rr, http.StatusOK)
	mcpaAssertMet(t, mock)
}

func TestHandleMCPLibraryApply_HappyPath(t *testing.T) {
	url := mcpaFixtureURL(t)
	opt, mock := mcpaSharedDB(t)
	s := newTestServer(opt, mcpaLiveLibrary(url))
	mcpaExpectLiveInstall(t, s, mock, "fetch", sqlmock.AnyArg(), url, `{}`, `{}`, false)

	rr := doAuthenticatedRequest(t, http.HandlerFunc(s.handleMCPLibraryApply), "POST", "/api/v1/mcp/library/apply", `{"name":"fetch"}`)
	assertStatus(t, rr, http.StatusOK)
	mcpaAssertMet(t, mock)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	if resp["status"] != "installed" {
		t.Fatalf("status = %v, want installed", resp["status"])
	}
	if resp["requires_approval"] != false || resp["self_approved"] != false {
		t.Fatalf("requires_approval/self_approved = %v/%v, want false/false", resp["requires_approval"], resp["self_approved"])
	}
	if _, ok := resp["inspection"].(map[string]any); !ok {
		t.Fatalf("expected inspection object, got %T", resp["inspection"])
	}
}

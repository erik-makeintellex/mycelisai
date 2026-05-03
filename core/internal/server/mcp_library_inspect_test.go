package server

import (
	"net/http"
	"testing"

	"github.com/mycelis/core/internal/mcp"
)

func TestHandleMCPLibrary_NilLibrary(t *testing.T) {
	s := newTestServer()
	rr := doRequest(t, http.HandlerFunc(s.handleMCPLibrary), "GET", "/api/v1/mcp/library", "")
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

func TestHandleMCPLibrary_WithLibrary(t *testing.T) {
	s := newTestServer(func(s *AdminServer) {
		s.MCPLibrary = &mcp.Library{Categories: []mcp.LibraryCategory{}}
	})
	rr := doRequest(t, http.HandlerFunc(s.handleMCPLibrary), "GET", "/api/v1/mcp/library", "")
	assertStatus(t, rr, http.StatusOK)
}

func TestHandleMCPLibraryInspect_LocalOwnedConfigIsAutoAllowed(t *testing.T) {
	s := newTestServer(func(s *AdminServer) {
		s.MCPLibrary = &mcp.Library{
			Categories: []mcp.LibraryCategory{
				{
					Name: "Default",
					Servers: []mcp.LibraryEntry{
						{Name: "filesystem", Transport: "stdio", Command: "npx", Tags: []string{"local"}},
					},
				},
			},
		}
	})

	rr := doAuthenticatedRequest(t, http.HandlerFunc(s.handleMCPLibraryInspect), "POST", "/api/v1/mcp/library/inspect", `{"name":"filesystem"}`)
	assertStatus(t, rr, http.StatusOK)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	if resp["decision"] != "allow" {
		t.Fatalf("decision = %v, want allow", resp["decision"])
	}
	governance, ok := resp["governance"].(map[string]any)
	if !ok {
		t.Fatalf("expected governance object, got %T", resp["governance"])
	}
	if governance["approval_required"] != false {
		t.Fatalf("approval_required = %v, want false", governance["approval_required"])
	}
	if governance["approval_reason"] != "user_owned_mcp_config" {
		t.Fatalf("approval_reason = %v, want user_owned_mcp_config", governance["approval_reason"])
	}
}

func TestHandleMCPLibraryInspect_RemoteConfigRequiresApproval(t *testing.T) {
	s := newTestServer(func(s *AdminServer) {
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

	rr := doAuthenticatedRequest(t, http.HandlerFunc(s.handleMCPLibraryInspect), "POST", "/api/v1/mcp/library/inspect", `{"name":"remote-knowledge"}`)
	assertStatus(t, rr, http.StatusOK)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	if resp["decision"] != "require_approval" {
		t.Fatalf("decision = %v, want require_approval", resp["decision"])
	}
}

func TestHandleMCPLibraryInspect_StandardLibraryFilesystemIsAutoAllowed(t *testing.T) {
	s := newTestServer(func(s *AdminServer) {
		s.MCPLibrary = loadStandardMCPLibrary(t)
	})

	rr := doAuthenticatedRequest(t, http.HandlerFunc(s.handleMCPLibraryInspect), "POST", "/api/v1/mcp/library/inspect", `{"name":"filesystem"}`)
	assertStatus(t, rr, http.StatusOK)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	if resp["decision"] != "allow" {
		t.Fatalf("decision = %v, want allow", resp["decision"])
	}

	governance, ok := resp["governance"].(map[string]any)
	if !ok {
		t.Fatalf("expected governance object, got %T", resp["governance"])
	}
	if governance["approval_reason"] != "user_owned_mcp_config" {
		t.Fatalf("approval_reason = %v, want user_owned_mcp_config", governance["approval_reason"])
	}
}

func TestHandleMCPLibraryInspect_StandardLibraryGitHubRequiresApproval(t *testing.T) {
	s := newTestServer(func(s *AdminServer) {
		s.MCPLibrary = loadStandardMCPLibrary(t)
	})

	rr := doAuthenticatedRequest(t, http.HandlerFunc(s.handleMCPLibraryInspect), "POST", "/api/v1/mcp/library/inspect", `{"name":"github"}`)
	assertStatus(t, rr, http.StatusOK)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	if resp["decision"] != "require_approval" {
		t.Fatalf("decision = %v, want require_approval", resp["decision"])
	}
	if resp["network_locality"] != "remote" {
		t.Fatalf("network_locality = %v, want remote", resp["network_locality"])
	}
	if resp["deployment_boundary"] != "external_saas" {
		t.Fatalf("deployment_boundary = %v, want external_saas", resp["deployment_boundary"])
	}
	if resp["credential_boundary"] != "secret_required" {
		t.Fatalf("credential_boundary = %v, want secret_required", resp["credential_boundary"])
	}
	if resp["bundle_install_path"] != "curated_library_only" {
		t.Fatalf("bundle_install_path = %v, want curated_library_only", resp["bundle_install_path"])
	}
	if resp["bundle_version_posture"] != "floating" {
		t.Fatalf("bundle_version_posture = %v, want floating", resp["bundle_version_posture"])
	}
	secrets, ok := resp["secrets_declared"].([]interface{})
	if !ok || len(secrets) == 0 || secrets[0] != "GITHUB_PERSONAL_ACCESS_TOKEN" {
		t.Fatalf("secrets_declared = %v, want github token metadata", resp["secrets_declared"])
	}
	governance, ok := resp["governance"].(map[string]any)
	if !ok {
		t.Fatalf("expected governance object, got %T", resp["governance"])
	}
	if governance["approval_reason"] != "credentialed_external_mcp_config" {
		t.Fatalf("approval_reason = %v, want credentialed_external_mcp_config", governance["approval_reason"])
	}
}

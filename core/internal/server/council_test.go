package server

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/mycelis/core/internal/swarm"
	"github.com/mycelis/core/pkg/protocol"
)

// ── Test fixtures ─────────────────────────────────────────────

func testCouncilSoma() *swarm.Soma {
	return swarm.NewTestSoma([]*swarm.TeamManifest{
		{
			ID:   "admin-core",
			Name: "Admin",
			Type: swarm.TeamTypeAction,
			Members: []protocol.AgentManifest{
				{ID: "admin", Role: "admin"},
			},
		},
		{
			ID:   "council-core",
			Name: "Council",
			Type: swarm.TeamTypeAction,
			Members: []protocol.AgentManifest{
				{ID: "council-architect", Role: "architect"},
				{ID: "council-coder", Role: "coder"},
				{ID: "council-creative", Role: "creative"},
				{ID: "council-sentry", Role: "sentry"},
			},
		},
		// Mission team — should NOT appear in council endpoints
		{
			ID:   "mission-abc-team-1",
			Name: "Scraper Team",
			Type: swarm.TeamTypeAction,
			Members: []protocol.AgentManifest{
				{ID: "scraper-agent", Role: "worker"},
			},
		},
	})
}

func withCouncilSoma() func(*AdminServer) {
	return func(s *AdminServer) {
		s.Soma = testCouncilSoma()
	}
}

// ── GET /api/v1/council/members ───────────────────────────────

func TestHandleListCouncilMembers(t *testing.T) {
	s := newTestServer(withCouncilSoma())
	rr := doRequest(t, http.HandlerFunc(s.HandleListCouncilMembers), "GET", "/api/v1/council/members", "")
	assertStatus(t, rr, http.StatusOK)

	var resp protocol.APIResponse
	assertJSON(t, rr, &resp)

	if !resp.OK {
		t.Fatalf("Expected ok=true, got ok=false: %s", resp.Error)
	}

	// Decode data as []CouncilMemberInfo
	raw, _ := json.Marshal(resp.Data)
	var members []CouncilMemberInfo
	if err := json.Unmarshal(raw, &members); err != nil {
		t.Fatalf("Failed to decode members: %v", err)
	}

	if len(members) != 5 {
		t.Fatalf("Expected 5 council members, got %d", len(members))
	}

	// Build lookup
	byID := map[string]CouncilMemberInfo{}
	for _, m := range members {
		byID[m.ID] = m
	}

	// Verify admin
	if m, ok := byID["admin"]; !ok {
		t.Error("Missing member: admin")
	} else {
		if m.Role != "admin" {
			t.Errorf("admin role: got %q, want %q", m.Role, "admin")
		}
		if m.Team != "admin-core" {
			t.Errorf("admin team: got %q, want %q", m.Team, "admin-core")
		}
	}

	// Verify architect
	if m, ok := byID["council-architect"]; !ok {
		t.Error("Missing member: council-architect")
	} else if m.Role != "architect" {
		t.Errorf("architect role: got %q, want %q", m.Role, "architect")
	}

	// Verify mission agent is excluded
	if _, ok := byID["scraper-agent"]; ok {
		t.Error("Mission agent 'scraper-agent' should NOT appear in council members")
	}
}

func TestHandleListCouncilMembers_NilSoma(t *testing.T) {
	s := newTestServer() // No Soma
	rr := doRequest(t, http.HandlerFunc(s.HandleListCouncilMembers), "GET", "/api/v1/council/members", "")
	assertStatus(t, rr, http.StatusServiceUnavailable)

	var resp protocol.APIResponse
	assertJSON(t, rr, &resp)
	if resp.OK {
		t.Error("Expected ok=false when Soma is nil")
	}
	if resp.Error == "" {
		t.Error("Expected error message when Soma is nil")
	}
}

// ── POST /api/v1/council/{member}/chat — Validation ──────────

func TestHandleCouncilChat_UnknownMember(t *testing.T) {
	s := newTestServer(withCouncilSoma())
	mux := setupMux(t, "POST /api/v1/council/{member}/chat", s.HandleCouncilChat)

	body := `{"messages":[{"role":"user","content":"hello"}]}`
	rr := doRequest(t, mux, "POST", "/api/v1/council/nonexistent/chat", body)
	assertStatus(t, rr, http.StatusNotFound)

	var resp protocol.APIResponse
	assertJSON(t, rr, &resp)
	if resp.OK {
		t.Error("Expected ok=false for unknown member")
	}
	if resp.Error == "" {
		t.Error("Expected error message for unknown member")
	}
}

func TestHandleCouncilChat_MissionAgentBlocked(t *testing.T) {
	s := newTestServer(withCouncilSoma())
	mux := setupMux(t, "POST /api/v1/council/{member}/chat", s.HandleCouncilChat)

	body := `{"messages":[{"role":"user","content":"hello"}]}`
	rr := doRequest(t, mux, "POST", "/api/v1/council/scraper-agent/chat", body)
	assertStatus(t, rr, http.StatusNotFound)
}

func TestHandleCouncilChat_EmptyMessages(t *testing.T) {
	s := newTestServer(withCouncilSoma())
	mux := setupMux(t, "POST /api/v1/council/{member}/chat", s.HandleCouncilChat)

	rr := doRequest(t, mux, "POST", "/api/v1/council/admin/chat", `{"messages":[]}`)
	assertStatus(t, rr, http.StatusBadRequest)

	var resp protocol.APIResponse
	assertJSON(t, rr, &resp)
	if resp.OK {
		t.Error("Expected ok=false for empty messages")
	}
}

func TestHandleCouncilChat_MissingMessages(t *testing.T) {
	s := newTestServer(withCouncilSoma())
	mux := setupMux(t, "POST /api/v1/council/{member}/chat", s.HandleCouncilChat)

	rr := doRequest(t, mux, "POST", "/api/v1/council/admin/chat", `{}`)
	assertStatus(t, rr, http.StatusBadRequest)
}

func TestHandleCouncilChat_BadJSON(t *testing.T) {
	s := newTestServer(withCouncilSoma())
	mux := setupMux(t, "POST /api/v1/council/{member}/chat", s.HandleCouncilChat)

	rr := doRequest(t, mux, "POST", "/api/v1/council/admin/chat", `not-json`)
	assertStatus(t, rr, http.StatusBadRequest)

	var resp protocol.APIResponse
	assertJSON(t, rr, &resp)
	if resp.OK {
		t.Error("Expected ok=false for bad JSON")
	}
}

func TestHandleCouncilChat_NilSoma(t *testing.T) {
	s := newTestServer() // No Soma
	mux := setupMux(t, "POST /api/v1/council/{member}/chat", s.HandleCouncilChat)

	body := `{"messages":[{"role":"user","content":"hello"}]}`
	rr := doRequest(t, mux, "POST", "/api/v1/council/admin/chat", body)
	// isCouncilMember returns false when Soma is nil → 404
	assertStatus(t, rr, http.StatusNotFound)
}

func TestHandleCouncilChat_NilNATS(t *testing.T) {
	s := newTestServer(withCouncilSoma()) // Soma present but NC is nil
	mux := setupMux(t, "POST /api/v1/council/{member}/chat", s.HandleCouncilChat)

	body := `{"messages":[{"role":"user","content":"hello"}]}`
	rr := doRequest(t, mux, "POST", "/api/v1/council/admin/chat", body)
	assertStatus(t, rr, http.StatusServiceUnavailable)

	var resp protocol.APIResponse
	assertJSON(t, rr, &resp)
	if resp.OK {
		t.Error("Expected ok=false when NATS offline")
	}
	if resp.Error == "" {
		t.Error("Expected error message about swarm offline")
	}
}

// ── isCouncilMember unit tests ────────────────────────────────

func TestIsCouncilMember(t *testing.T) {
	s := newTestServer(withCouncilSoma())

	tests := []struct {
		memberID string
		wantOK   bool
		wantTeam string
		wantRole string
	}{
		{"admin", true, "admin-core", "admin"},
		{"council-architect", true, "council-core", "architect"},
		{"council-coder", true, "council-core", "coder"},
		{"council-creative", true, "council-core", "creative"},
		{"council-sentry", true, "council-core", "sentry"},
		{"nonexistent", false, "", ""},
		{"scraper-agent", false, "", ""},  // mission agent — not council
		{"", false, "", ""},               // empty ID
	}

	for _, tt := range tests {
		t.Run(tt.memberID, func(t *testing.T) {
			teamID, role, ok := s.isCouncilMember(tt.memberID)
			if ok != tt.wantOK {
				t.Errorf("isCouncilMember(%q) ok = %v, want %v", tt.memberID, ok, tt.wantOK)
			}
			if teamID != tt.wantTeam {
				t.Errorf("isCouncilMember(%q) team = %q, want %q", tt.memberID, teamID, tt.wantTeam)
			}
			if role != tt.wantRole {
				t.Errorf("isCouncilMember(%q) role = %q, want %q", tt.memberID, role, tt.wantRole)
			}
		})
	}
}

func TestIsCouncilMember_NilSoma(t *testing.T) {
	s := newTestServer() // No Soma
	_, _, ok := s.isCouncilMember("admin")
	if ok {
		t.Error("Expected false when Soma is nil")
	}
}

// ── APIResponse helpers ───────────────────────────────────────

func TestRespondAPIJSON_Success(t *testing.T) {
	s := newTestServer(withCouncilSoma())
	rr := doRequest(t, http.HandlerFunc(s.HandleListCouncilMembers), "GET", "/api/v1/council/members", "")

	// Verify Content-Type
	ct := rr.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type: got %q, want %q", ct, "application/json")
	}

	// Verify it's valid JSON with ok=true
	var resp protocol.APIResponse
	assertJSON(t, rr, &resp)
	if !resp.OK {
		t.Error("Expected ok=true in success response")
	}
}

func TestRespondAPIJSON_Error(t *testing.T) {
	s := newTestServer() // nil Soma → triggers respondAPIError
	rr := doRequest(t, http.HandlerFunc(s.HandleListCouncilMembers), "GET", "/api/v1/council/members", "")

	ct := rr.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type: got %q, want %q", ct, "application/json")
	}

	var resp protocol.APIResponse
	assertJSON(t, rr, &resp)
	if resp.OK {
		t.Error("Expected ok=false in error response")
	}
	if resp.Error == "" {
		t.Error("Expected non-empty error in error response")
	}
}

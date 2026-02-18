package server

import (
	"net/http"
	"testing"
)

// ── GET /api/v1/proposals ──────────────────────────────────────────

func TestHandleProposals_List(t *testing.T) {
	s := newTestServer(func(s *AdminServer) {
		s.Proposals = NewProposalStore() // Seeds 2 default proposals
	})

	rr := doRequest(t, http.HandlerFunc(s.HandleProposals), "GET", "/api/v1/proposals", "")
	assertStatus(t, rr, http.StatusOK)

	var result map[string]any
	assertJSON(t, rr, &result)
	count := result["count"].(float64)
	if count != 2 {
		t.Errorf("Expected 2 seeded proposals, got %v", count)
	}
}

func TestHandleProposals_NilStore(t *testing.T) {
	s := newTestServer()
	rr := doRequest(t, http.HandlerFunc(s.HandleProposals), "GET", "/api/v1/proposals", "")
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

// ── POST /api/v1/proposals ─────────────────────────────────────────

func TestHandleProposals_Create(t *testing.T) {
	s := newTestServer(func(s *AdminServer) {
		s.Proposals = NewProposalStore()
	})

	body := `{"name":"Test Team","role":"scout","reason":"Testing","agents":[{"id":"scout-1","role":"scout"}]}`
	rr := doRequest(t, http.HandlerFunc(s.HandleProposals), "POST", "/api/v1/proposals", body)
	assertStatus(t, rr, http.StatusCreated)

	var result TeamProposal
	assertJSON(t, rr, &result)
	if result.Name != "Test Team" {
		t.Errorf("Expected name 'Test Team', got %q", result.Name)
	}
	if result.ID == "" {
		t.Error("Expected non-empty proposal ID")
	}

	// Verify it's in the store (2 seeds + 1 new = 3)
	all := s.Proposals.List()
	if len(all) != 3 {
		t.Errorf("Expected 3 proposals in store, got %d", len(all))
	}
}

func TestHandleProposals_CreateMissingName(t *testing.T) {
	s := newTestServer(func(s *AdminServer) {
		s.Proposals = NewProposalStore()
	})

	rr := doRequest(t, http.HandlerFunc(s.HandleProposals), "POST", "/api/v1/proposals", `{"role":"scout"}`)
	assertStatus(t, rr, http.StatusBadRequest)
}

func TestHandleProposals_CreateInvalidJSON(t *testing.T) {
	s := newTestServer(func(s *AdminServer) {
		s.Proposals = NewProposalStore()
	})

	rr := doRequest(t, http.HandlerFunc(s.HandleProposals), "POST", "/api/v1/proposals", "not-json")
	assertStatus(t, rr, http.StatusBadRequest)
}

// ── POST /api/v1/proposals/{id}/approve ────────────────────────────

func TestHandleProposalApprove(t *testing.T) {
	store := NewProposalStore()
	s := newTestServer(func(s *AdminServer) { s.Proposals = store })

	// Get a seeded proposal ID
	proposals := store.List()
	id := proposals[0].ID

	mux := setupMux(t, "POST /api/v1/proposals/{id}/approve", s.HandleProposalApprove)
	rr := doRequest(t, mux, "POST", "/api/v1/proposals/"+id+"/approve", "")
	assertStatus(t, rr, http.StatusOK)

	var result map[string]any
	assertJSON(t, rr, &result)
	if result["status"] != "approved" {
		t.Errorf("Expected status 'approved', got %v", result["status"])
	}

	// Verify the proposal is now approved
	p, _ := store.Get(id)
	if p.Status != ProposalApproved {
		t.Errorf("Expected proposal status %q, got %q", ProposalApproved, p.Status)
	}
}

func TestHandleProposalApprove_NotFound(t *testing.T) {
	s := newTestServer(func(s *AdminServer) { s.Proposals = NewProposalStore() })
	mux := setupMux(t, "POST /api/v1/proposals/{id}/approve", s.HandleProposalApprove)
	rr := doRequest(t, mux, "POST", "/api/v1/proposals/nonexistent/approve", "")
	assertStatus(t, rr, http.StatusNotFound)
}

func TestHandleProposalApprove_AlreadyResolved(t *testing.T) {
	store := NewProposalStore()
	s := newTestServer(func(s *AdminServer) { s.Proposals = store })

	proposals := store.List()
	id := proposals[0].ID
	store.UpdateStatus(id, ProposalApproved) // Pre-approve

	mux := setupMux(t, "POST /api/v1/proposals/{id}/approve", s.HandleProposalApprove)
	rr := doRequest(t, mux, "POST", "/api/v1/proposals/"+id+"/approve", "")
	assertStatus(t, rr, http.StatusConflict)
}

// ── POST /api/v1/proposals/{id}/reject ─────────────────────────────

func TestHandleProposalReject(t *testing.T) {
	store := NewProposalStore()
	s := newTestServer(func(s *AdminServer) { s.Proposals = store })

	proposals := store.List()
	id := proposals[0].ID

	mux := setupMux(t, "POST /api/v1/proposals/{id}/reject", s.HandleProposalReject)
	rr := doRequest(t, mux, "POST", "/api/v1/proposals/"+id+"/reject", "")
	assertStatus(t, rr, http.StatusOK)

	var result map[string]any
	assertJSON(t, rr, &result)
	if result["status"] != "rejected" {
		t.Errorf("Expected status 'rejected', got %v", result["status"])
	}
}

func TestHandleProposalReject_NotFound(t *testing.T) {
	s := newTestServer(func(s *AdminServer) { s.Proposals = NewProposalStore() })
	mux := setupMux(t, "POST /api/v1/proposals/{id}/reject", s.HandleProposalReject)
	rr := doRequest(t, mux, "POST", "/api/v1/proposals/nonexistent/reject", "")
	assertStatus(t, rr, http.StatusNotFound)
}

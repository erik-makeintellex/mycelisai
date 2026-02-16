package server

import (
	"net/http"
	"os"
	"testing"

	"github.com/mycelis/core/internal/governance"
	pb "github.com/mycelis/core/pkg/pb/swarm"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ── GET /api/v1/governance/policy ──────────────────────────────────

func TestHandleGetPolicy(t *testing.T) {
	s := newTestServer(withGuard(defaultTestPolicyConfig()))

	rr := doRequest(t, http.HandlerFunc(s.handleGetPolicy), "GET", "/api/v1/governance/policy", "")
	assertStatus(t, rr, http.StatusOK)

	var cfg governance.PolicyConfig
	assertJSON(t, rr, &cfg)
	if cfg.Defaults.DefaultAction != governance.ActionAllow {
		t.Errorf("Expected default action %q, got %q", governance.ActionAllow, cfg.Defaults.DefaultAction)
	}
	if len(cfg.Groups) != 1 {
		t.Errorf("Expected 1 group, got %d", len(cfg.Groups))
	}
}

func TestHandleGetPolicy_NilGuard(t *testing.T) {
	s := newTestServer()
	rr := doRequest(t, http.HandlerFunc(s.handleGetPolicy), "GET", "/api/v1/governance/policy", "")
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

// ── PUT /api/v1/governance/policy ──────────────────────────────────

func TestHandleUpdatePolicy(t *testing.T) {
	// SavePolicyToFile writes to "config/policy.yaml" (relative).
	// Use t.Chdir so the file write succeeds without polluting the repo.
	t.Chdir(t.TempDir())
	os.MkdirAll("config", 0755)

	s := newTestServer(withGuard(defaultTestPolicyConfig()))

	body := `{"groups":[],"defaults":{"default_action":"DENY"}}`
	rr := doRequest(t, http.HandlerFunc(s.handleUpdatePolicy), "PUT", "/api/v1/governance/policy", body)
	assertStatus(t, rr, http.StatusOK)

	// Verify in-memory update persisted
	cfg := s.Guard.GetPolicyConfig()
	if cfg.Defaults.DefaultAction != governance.ActionDeny {
		t.Errorf("Expected default action updated to %q, got %q", governance.ActionDeny, cfg.Defaults.DefaultAction)
	}
}

func TestHandleUpdatePolicy_NilGuard(t *testing.T) {
	s := newTestServer()
	body := `{"groups":[],"defaults":{"default_action":"ALLOW"}}`
	rr := doRequest(t, http.HandlerFunc(s.handleUpdatePolicy), "PUT", "/api/v1/governance/policy", body)
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

func TestHandleUpdatePolicy_MissingDefault(t *testing.T) {
	s := newTestServer(withGuard(defaultTestPolicyConfig()))
	body := `{"groups":[],"defaults":{}}`
	rr := doRequest(t, http.HandlerFunc(s.handleUpdatePolicy), "PUT", "/api/v1/governance/policy", body)
	assertStatus(t, rr, http.StatusBadRequest)
}

func TestHandleUpdatePolicy_InvalidJSON(t *testing.T) {
	s := newTestServer(withGuard(defaultTestPolicyConfig()))
	rr := doRequest(t, http.HandlerFunc(s.handleUpdatePolicy), "PUT", "/api/v1/governance/policy", "not-json")
	assertStatus(t, rr, http.StatusBadRequest)
}

// ── GET /api/v1/governance/pending ─────────────────────────────────

func TestHandleGetPendingApprovals(t *testing.T) {
	s := newTestServer(withGuard(defaultTestPolicyConfig()))

	// Seed a pending approval request
	s.Guard.PendingBuffer["req-1"] = &pb.ApprovalRequest{
		RequestId: "req-1",
		Reason:    "test reason",
		OriginalMessage: &pb.MsgEnvelope{
			SourceAgentId: "agent-1",
			TeamId:        "team-1",
		},
		ExpiresAt: timestamppb.Now(),
	}

	rr := doRequest(t, http.HandlerFunc(s.handleGetPendingApprovals), "GET", "/api/v1/governance/pending", "")
	assertStatus(t, rr, http.StatusOK)

	var result []pendingApprovalJSON
	assertJSON(t, rr, &result)
	if len(result) != 1 {
		t.Fatalf("Expected 1 pending approval, got %d", len(result))
	}
	if result[0].ID != "req-1" {
		t.Errorf("Expected ID 'req-1', got %q", result[0].ID)
	}
	if result[0].SourceAgent != "agent-1" {
		t.Errorf("Expected source agent 'agent-1', got %q", result[0].SourceAgent)
	}
}

func TestHandleGetPendingApprovals_Empty(t *testing.T) {
	s := newTestServer(withGuard(defaultTestPolicyConfig()))
	rr := doRequest(t, http.HandlerFunc(s.handleGetPendingApprovals), "GET", "/api/v1/governance/pending", "")
	assertStatus(t, rr, http.StatusOK)

	var result []pendingApprovalJSON
	assertJSON(t, rr, &result)
	if len(result) != 0 {
		t.Errorf("Expected empty array, got %d items", len(result))
	}
}

func TestHandleGetPendingApprovals_NilGuard(t *testing.T) {
	s := newTestServer()
	rr := doRequest(t, http.HandlerFunc(s.handleGetPendingApprovals), "GET", "/api/v1/governance/pending", "")
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

// ── POST /api/v1/governance/resolve/{id} ───────────────────────────

func TestHandleResolveApproval_Approve(t *testing.T) {
	s := newTestServer(withGuard(defaultTestPolicyConfig()))
	s.Guard.PendingBuffer["req-1"] = &pb.ApprovalRequest{
		RequestId:       "req-1",
		Reason:          "test",
		OriginalMessage: &pb.MsgEnvelope{SourceAgentId: "agent-1"},
	}

	mux := setupMux(t, "POST /api/v1/governance/resolve/{id}", s.handleResolveApproval)
	rr := doRequest(t, mux, "POST", "/api/v1/governance/resolve/req-1", `{"action":"APPROVE"}`)
	assertStatus(t, rr, http.StatusOK)

	var result map[string]string
	assertJSON(t, rr, &result)
	if result["status"] != "resolved" {
		t.Errorf("Expected status 'resolved', got %q", result["status"])
	}
	if result["action"] != "APPROVE" {
		t.Errorf("Expected action 'APPROVE', got %q", result["action"])
	}

	// Verify pending buffer is now empty
	if len(s.Guard.PendingBuffer) != 0 {
		t.Errorf("Expected pending buffer empty, got %d items", len(s.Guard.PendingBuffer))
	}
}

func TestHandleResolveApproval_Reject(t *testing.T) {
	s := newTestServer(withGuard(defaultTestPolicyConfig()))
	s.Guard.PendingBuffer["req-2"] = &pb.ApprovalRequest{
		RequestId:       "req-2",
		Reason:          "test",
		OriginalMessage: &pb.MsgEnvelope{SourceAgentId: "agent-2"},
	}

	mux := setupMux(t, "POST /api/v1/governance/resolve/{id}", s.handleResolveApproval)
	rr := doRequest(t, mux, "POST", "/api/v1/governance/resolve/req-2", `{"action":"REJECT"}`)
	assertStatus(t, rr, http.StatusOK)

	var result map[string]string
	assertJSON(t, rr, &result)
	if result["action"] != "REJECT" {
		t.Errorf("Expected action 'REJECT', got %q", result["action"])
	}
}

func TestHandleResolveApproval_InvalidAction(t *testing.T) {
	s := newTestServer(withGuard(defaultTestPolicyConfig()))
	mux := setupMux(t, "POST /api/v1/governance/resolve/{id}", s.handleResolveApproval)
	rr := doRequest(t, mux, "POST", "/api/v1/governance/resolve/req-1", `{"action":"MAYBE"}`)
	assertStatus(t, rr, http.StatusBadRequest)
}

func TestHandleResolveApproval_NotFound(t *testing.T) {
	s := newTestServer(withGuard(defaultTestPolicyConfig()))
	mux := setupMux(t, "POST /api/v1/governance/resolve/{id}", s.handleResolveApproval)
	rr := doRequest(t, mux, "POST", "/api/v1/governance/resolve/nonexistent", `{"action":"APPROVE"}`)
	assertStatus(t, rr, http.StatusNotFound)
}

func TestHandleResolveApproval_MissingID(t *testing.T) {
	// Call handler directly (no mux) so r.PathValue("id") returns ""
	s := newTestServer(withGuard(defaultTestPolicyConfig()))
	rr := doRequest(t, http.HandlerFunc(s.handleResolveApproval), "POST", "/api/v1/governance/resolve/", `{"action":"APPROVE"}`)
	assertStatus(t, rr, http.StatusBadRequest)
}

func TestHandleResolveApproval_InvalidJSON(t *testing.T) {
	s := newTestServer(withGuard(defaultTestPolicyConfig()))
	mux := setupMux(t, "POST /api/v1/governance/resolve/{id}", s.handleResolveApproval)
	rr := doRequest(t, mux, "POST", "/api/v1/governance/resolve/req-1", "not-json")
	assertStatus(t, rr, http.StatusBadRequest)
}

func TestHandleResolveApproval_NilGuard(t *testing.T) {
	s := newTestServer()
	mux := setupMux(t, "POST /api/v1/governance/resolve/{id}", s.handleResolveApproval)
	rr := doRequest(t, mux, "POST", "/api/v1/governance/resolve/req-1", `{"action":"APPROVE"}`)
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

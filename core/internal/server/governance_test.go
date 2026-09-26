package server

import (
	"net/http"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/governance"
	"github.com/mycelis/core/internal/router"
	pb "github.com/mycelis/core/pkg/pb/swarm"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const resolvePattern = "POST /api/v1/governance/resolve/{id}"

func expectAudits(mock sqlmock.Sqlmock, n int) {
	for i := 0; i < n; i++ {
		mock.ExpectExec("INSERT INTO log_entries").WillReturnResult(sqlmock.NewResult(1, 1))
	}
}

// approvalServer wires a guard, audit DB, embedded NATS and a real Router.
func approvalServer(t *testing.T) (*AdminServer, sqlmock.Sqlmock) {
	t.Helper()
	dbOpt, mock := withDB(t)
	s := newTestServer(withGuard(defaultTestPolicyConfig()), dbOpt, withNATS(t))
	s.Router = router.NewRouter(s.NC, s.Guard)
	return s, mock
}

func seedPending(s *AdminServer, id string) {
	s.Guard.PendingBuffer[id] = &pb.ApprovalRequest{
		RequestId: id,
		Reason:    "test reason",
		OriginalMessage: &pb.MsgEnvelope{
			SourceAgentId: "agent-1",
			TeamId:        "team-1",
			Payload:       &pb.MsgEnvelope_Event{Event: &pb.EventPayload{EventType: "k8s.delete.cluster"}},
		},
		ExpiresAt: timestamppb.Now(),
	}
}

// ── GET /api/v1/governance/policy ──────────────────────────────────

func TestHandleGetPolicy(t *testing.T) {
	s := newTestServer(withGuard(defaultTestPolicyConfig()))
	rr := doAuthenticatedRequest(t, http.HandlerFunc(s.handleGetPolicy), "GET", "/api/v1/governance/policy", "")
	assertStatus(t, rr, http.StatusOK)
	var cfg governance.PolicyConfig
	assertJSON(t, rr, &cfg)
	if cfg.Defaults.DefaultAction != governance.ActionAllow || len(cfg.Groups) != 1 {
		t.Errorf("unexpected policy: %+v", cfg)
	}
}

func TestHandleGetPolicy_NilGuard(t *testing.T) {
	s := newTestServer()
	rr := doAuthenticatedRequest(t, http.HandlerFunc(s.handleGetPolicy), "GET", "/api/v1/governance/policy", "")
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

// ── PUT /api/v1/governance/policy ──────────────────────────────────

func TestHandleUpdatePolicy(t *testing.T) {
	t.Chdir(t.TempDir())
	os.MkdirAll("config", 0755)
	dbOpt, mock := withDB(t)
	expectAudits(mock, 2)
	s := newTestServer(withGuard(defaultTestPolicyConfig()), dbOpt)

	body := `{"groups":[],"defaults":{"default_action":"DENY"}}`
	rr := doAuthenticatedRequest(t, http.HandlerFunc(s.handleUpdatePolicy), "PUT", "/api/v1/governance/policy", body)
	assertStatus(t, rr, http.StatusOK)
	if cfg := s.Guard.GetPolicyConfig(); cfg.Defaults.DefaultAction != governance.ActionDeny {
		t.Errorf("Expected default action %q, got %q", governance.ActionDeny, cfg.Defaults.DefaultAction)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestHandleUpdatePolicy_NilGuard(t *testing.T) {
	s := newTestServer()
	body := `{"groups":[],"defaults":{"default_action":"ALLOW"}}`
	rr := doAuthenticatedRequest(t, http.HandlerFunc(s.handleUpdatePolicy), "PUT", "/api/v1/governance/policy", body)
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

func TestHandleUpdatePolicy_RejectsBadDefaults(t *testing.T) {
	s := newTestServer(withGuard(defaultTestPolicyConfig()))
	for _, body := range []string{
		`{"groups":[],"defaults":{}}`,
		`{"groups":[],"defaults":{"default_action":"DNY"}}`,
		"not-json",
	} {
		rr := doAuthenticatedRequest(t, http.HandlerFunc(s.handleUpdatePolicy), "PUT", "/api/v1/governance/policy", body)
		assertStatus(t, rr, http.StatusBadRequest)
	}
}

// ── GET /api/v1/governance/pending ─────────────────────────────────

func TestHandleGetPendingApprovals(t *testing.T) {
	s := newTestServer(withGuard(defaultTestPolicyConfig()))
	seedPending(s, "req-1")
	rr := doAuthenticatedRequest(t, http.HandlerFunc(s.handleGetPendingApprovals), "GET", "/api/v1/governance/pending", "")
	assertStatus(t, rr, http.StatusOK)
	var result []pendingApprovalJSON
	assertJSON(t, rr, &result)
	if len(result) != 1 || result[0].ID != "req-1" || result[0].SourceAgent != "agent-1" {
		t.Fatalf("unexpected pending list: %+v", result)
	}
}

func TestHandleGetPendingApprovals_Empty(t *testing.T) {
	s := newTestServer(withGuard(defaultTestPolicyConfig()))
	rr := doAuthenticatedRequest(t, http.HandlerFunc(s.handleGetPendingApprovals), "GET", "/api/v1/governance/pending", "")
	assertStatus(t, rr, http.StatusOK)
	var result []pendingApprovalJSON
	assertJSON(t, rr, &result)
	if len(result) != 0 {
		t.Errorf("Expected empty array, got %d items", len(result))
	}
}

func TestHandleGetPendingApprovals_NilGuard(t *testing.T) {
	s := newTestServer()
	rr := doAuthenticatedRequest(t, http.HandlerFunc(s.handleGetPendingApprovals), "GET", "/api/v1/governance/pending", "")
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

// ── POST /api/v1/governance/resolve/{id} ───────────────────────────

func TestHandleResolveApproval_Approve(t *testing.T) {
	s, mock := approvalServer(t)
	expectAudits(mock, 2)
	seedPending(s, "req-1")
	rr := doAuthenticatedRequest(t, setupMux(t, resolvePattern, s.handleResolveApproval), "POST", "/api/v1/governance/resolve/req-1", `{"action":"APPROVE"}`)
	assertStatus(t, rr, http.StatusOK)
	var result map[string]string
	assertJSON(t, rr, &result)
	if result["status"] != "resolved" || result["action"] != "APPROVE" || result["audit_id"] == "" {
		t.Errorf("unexpected result: %+v", result)
	}
	if len(s.Guard.ListPending()) != 0 {
		t.Error("Expected pending buffer empty")
	}
}

func TestHandleResolveApproval_Reject(t *testing.T) {
	s, mock := approvalServer(t)
	expectAudits(mock, 2)
	seedPending(s, "req-2")
	rr := doAuthenticatedRequest(t, setupMux(t, resolvePattern, s.handleResolveApproval), "POST", "/api/v1/governance/resolve/req-2", `{"action":"REJECT"}`)
	assertStatus(t, rr, http.StatusOK)
	var result map[string]string
	assertJSON(t, rr, &result)
	if result["action"] != "REJECT" {
		t.Errorf("Expected action 'REJECT', got %q", result["action"])
	}
}

func TestHandleResolveApproval_BadRequests(t *testing.T) {
	s := newTestServer(withGuard(defaultTestPolicyConfig()))
	seedPending(s, "req-1")
	mux := setupMux(t, resolvePattern, s.handleResolveApproval)
	for _, body := range []string{`{"action":"MAYBE"}`, `{"action":"DENY"}`, `{"action":"approve"}`, "not-json"} {
		rr := doAuthenticatedRequest(t, mux, "POST", "/api/v1/governance/resolve/req-1", body)
		assertStatus(t, rr, http.StatusBadRequest)
	}
	// Direct call (no mux) so r.PathValue("id") returns "".
	rr := doAuthenticatedRequest(t, http.HandlerFunc(s.handleResolveApproval), "POST", "/api/v1/governance/resolve/", `{"action":"APPROVE"}`)
	assertStatus(t, rr, http.StatusBadRequest)
	if _, ok := s.Guard.PendingRequest("req-1"); !ok {
		t.Fatal("bad requests must leave the approval pending")
	}
}

func TestHandleResolveApproval_NotFound(t *testing.T) {
	s, _ := approvalServer(t)
	rr := doAuthenticatedRequest(t, setupMux(t, resolvePattern, s.handleResolveApproval), "POST", "/api/v1/governance/resolve/nonexistent", `{"action":"APPROVE"}`)
	assertStatus(t, rr, http.StatusNotFound)
}

func TestHandleUpdatePolicy_RejectsFailOpenPolicies(t *testing.T) {
	t.Chdir(t.TempDir())
	os.MkdirAll("config", 0o755)
	os.WriteFile(defaultPolicyPath, []byte("original\n"), 0o644)
	dbOpt, mock := withDB(t) // no expectations: any audit insert fails the test
	s := newTestServer(withGuard(defaultTestPolicyConfig()), dbOpt)
	before := s.Guard.GetPolicyConfig()
	for name, body := range map[string]string{
		"empty":        `{}`,
		"typo-default": `{"groups":[{"name":"g","targets":["*"],"rules":[{"intent":"x","action":"DENY"}]}],"defaults":{"default_action":"DENNY"}}`,
		"typo-rule":    `{"groups":[{"name":"g","targets":["*"],"rules":[{"intent":"x","action":"deny"}]}],"defaults":{"default_action":"ALLOW"}}`,
		"allow-only":   `{"groups":[],"defaults":{"default_action":"ALLOW"}}`,
	} {
		rr := doAuthenticatedRequest(t, http.HandlerFunc(s.handleUpdatePolicy), "PUT", "/api/v1/governance/policy", body)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("%s: expected 400, got %d %s", name, rr.Code, rr.Body.String())
		}
	}
	if s.Guard.GetPolicyConfig() != before {
		t.Fatal("rejected policy changed memory")
	}
	if got, _ := os.ReadFile(defaultPolicyPath); string(got) != "original\n" {
		t.Fatal("rejected policy touched the file")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

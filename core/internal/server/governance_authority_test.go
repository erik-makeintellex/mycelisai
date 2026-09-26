package server

import (
	"bytes"
	"crypto/sha256"
	"database/sql/driver"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/governance"
	"github.com/nats-io/nats.go"
)

func standardUserIdentity() *RequestIdentity {
	return &RequestIdentity{UserID: "u-std", Username: "std@example.com", Role: "operator", EffectiveRole: "operator",
		PrincipalType: "google_workspace_user", AuthSource: "web_google", Scopes: []string{"soma:work", "runs:read", "outputs:read"}}
}

func adminWithScopes(scopes ...string) *RequestIdentity {
	id := localAdminIdentityForTest()
	id.Scopes = scopes
	return id
}

type governanceRoute struct {
	name, method, path, body, scope string
	handler                         func(s *AdminServer) http.Handler
}

func governanceRoutes() []governanceRoute {
	mux := func(pattern string, h func(s *AdminServer) http.HandlerFunc) func(s *AdminServer) http.Handler {
		return func(s *AdminServer) http.Handler { m := http.NewServeMux(); m.HandleFunc(pattern, h(s)); return m }
	}
	direct := func(h func(s *AdminServer) http.HandlerFunc) func(s *AdminServer) http.Handler {
		return func(s *AdminServer) http.Handler { return h(s) }
	}
	return []governanceRoute{
		{"get-policy", "GET", "/api/v1/governance/policy", "", scopeGovernanceRead, mux("GET /api/v1/governance/policy", func(s *AdminServer) http.HandlerFunc { return s.handleGetPolicy })},
		{"get-pending", "GET", "/api/v1/governance/pending", "", scopeGovernanceRead, mux("GET /api/v1/governance/pending", func(s *AdminServer) http.HandlerFunc { return s.handleGetPendingApprovals })},
		{"admin-approvals", "GET", "/admin/approvals", "", scopeGovernanceRead, direct(func(s *AdminServer) http.HandlerFunc { return s.handleApprovals })},
		{"put-policy", "PUT", "/api/v1/governance/policy", `{"groups":[],"defaults":{"default_action":"DENY"}}`, scopeGovernanceWrite, mux("PUT /api/v1/governance/policy", func(s *AdminServer) http.HandlerFunc { return s.handleUpdatePolicy })},
		{"resolve", "POST", "/api/v1/governance/resolve/req-1", `{"action":"APPROVE"}`, scopeApprovalsDecide, mux(resolvePattern, func(s *AdminServer) http.HandlerFunc { return s.handleResolveApproval })},
		{"admin-approval-action", "POST", "/admin/approvals/req-1", `{"action":"APPROVE"}`, scopeApprovalsDecide, direct(func(s *AdminServer) http.HandlerFunc { return s.handleApprovalAction })},
	}
}

// Every denial: no data in the body, no file write, no audit, no Guard change.
func TestGovernanceAuthorityMatrixDenials(t *testing.T) {
	t.Chdir(t.TempDir())
	os.MkdirAll("config", 0o755)
	original := []byte("original-policy\n")
	os.WriteFile(defaultPolicyPath, original, 0o644)

	for _, route := range governanceRoutes() {
		cases := map[string]*RequestIdentity{
			"anonymous":         nil,
			"standard-user":     standardUserIdentity(),
			"operator-with-all": {UserID: "u-op", Role: "operator", Scopes: []string{"*"}},
		}
		// An admin holding a different governance scope must not pass.
		switch route.scope {
		case scopeGovernanceWrite:
			cases["admin-read-only"] = adminWithScopes(scopeGovernanceRead)
		case scopeApprovalsDecide:
			cases["admin-read-write"] = adminWithScopes(scopeGovernanceRead, scopeGovernanceWrite)
		case scopeGovernanceRead:
			cases["admin-approvals-only"] = adminWithScopes(scopeApprovalsDecide)
		}
		for name, identity := range cases {
			t.Run(route.name+"/"+name, func(t *testing.T) {
				dbOpt, mock := withDB(t) // no expectations: any audit insert fails the test
				s := newTestServer(withGuard(defaultTestPolicyConfig()), dbOpt)
				seedPending(s, "req-1")
				before := s.Guard.GetPolicyConfig()
				var rr *httptest.ResponseRecorder
				if identity == nil {
					rr = doRequest(t, route.handler(s), route.method, route.path, route.body)
				} else {
					rr = doAuthenticatedRequestAs(t, route.handler(s), route.method, route.path, route.body, identity)
				}
				want := http.StatusForbidden
				if identity == nil {
					want = http.StatusUnauthorized
				}
				assertStatus(t, rr, want)
				body := rr.Body.String()
				if strings.Contains(body, `"data"`) || strings.Contains(body, "req-1") || strings.Contains(body, "test-group") {
					t.Fatalf("denial leaked data: %s", body)
				}
				if s.Guard.GetPolicyConfig() != before {
					t.Fatal("denial changed the live policy")
				}
				if _, ok := s.Guard.PendingRequest("req-1"); !ok {
					t.Fatal("denial resolved the pending approval")
				}
				if got, _ := os.ReadFile(defaultPolicyPath); !bytes.Equal(got, original) {
					t.Fatal("denial wrote the policy file")
				}
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}

func TestGovernanceAuthorityScopedAdminAllowed(t *testing.T) {
	s := newTestServer(withGuard(defaultTestPolicyConfig()))
	rr := doAuthenticatedRequestAs(t, http.HandlerFunc(s.handleGetPolicy), "GET", "/api/v1/governance/policy", "", adminWithScopes(scopeGovernanceRead))
	assertStatus(t, rr, http.StatusOK)

	s2, mock := approvalServer(t)
	expectAudits(mock, 2)
	seedPending(s2, "req-9")
	rr = doAuthenticatedRequestAs(t, setupMux(t, resolvePattern, s2.handleResolveApproval), "POST", "/api/v1/governance/resolve/req-9", `{"action":"APPROVE"}`, adminWithScopes("approvals:*"))
	assertStatus(t, rr, http.StatusOK)
}

func TestGovernanceAuthorityRejectsForgedWebIdentity(t *testing.T) {
	t.Setenv("MYCELIS_WEB_SESSION_SECRET", "")
	t.Setenv("MYCELIS_WEB_IDENTITY_FORWARD_SECRET", testForwardSecret)
	payload := encodeForwardedWebIdentityForTest(t, forwardedWebIdentityPayload{Sub: "attacker", Email: "attacker@example.com", Role: "admin", Provider: "google", IAT: time.Now().Unix()})
	for _, route := range governanceRoutes() {
		s := newTestServer(withGuard(defaultTestPolicyConfig()))
		seedPending(s, "req-1")
		req, _ := http.NewRequest(route.method, route.path, strings.NewReader(route.body))
		req.Header.Set("Authorization", "Bearer test-key")
		req.Header.Set(forwardedWebIdentityHeader, payload) // unsigned
		rr := httptest.NewRecorder()
		AuthMiddleware("test-key", route.handler(s)).ServeHTTP(rr, req)
		assertStatus(t, rr, http.StatusUnauthorized)
		if _, ok := s.Guard.PendingRequest("req-1"); !ok {
			t.Fatalf("%s: forged identity resolved an approval", route.name)
		}
	}
}

// ── PUT fail-closed ordering ───────────────────────────────────────

func putPolicy(t *testing.T, s *AdminServer, body string) *httptest.ResponseRecorder {
	return doAuthenticatedRequest(t, http.HandlerFunc(s.handleUpdatePolicy), "PUT", "/api/v1/governance/policy", body)
}

func TestUpdatePolicyInvalidPostureLeavesFileUntouched(t *testing.T) {
	t.Chdir(t.TempDir())
	os.MkdirAll("config", 0o755)
	os.WriteFile(defaultPolicyPath, []byte("original\n"), 0o644)
	s := newTestServer(withGuard(defaultTestPolicyConfig()))
	rr := putPolicy(t, s, `{"groups":[{"name":"p","targets":["posture:x"],"rules":[{"intent":"^.*$","action":"ALLOW"}]}],"defaults":{"default_action":"ALLOW"}}`)
	assertStatus(t, rr, http.StatusBadRequest)
	if got, _ := os.ReadFile(defaultPolicyPath); string(got) != "original\n" {
		t.Fatal("invalid policy touched the file")
	}
}

func TestUpdatePolicyRejectsOversizedBody(t *testing.T) {
	s := newTestServer(withGuard(defaultTestPolicyConfig()))
	body := `{"groups":[],"defaults":{"default_action":"DENY"},"pad":"` + strings.Repeat("x", maxPolicyBodyBytes) + `"}`
	rr := putPolicy(t, s, body)
	assertStatus(t, rr, http.StatusRequestEntityTooLarge)
	if s.Guard.GetPolicyConfig().Defaults.DefaultAction != governance.ActionAllow {
		t.Fatal("oversized body changed the policy")
	}
}

func TestUpdatePolicyAuditFailureChangesNothing(t *testing.T) {
	t.Chdir(t.TempDir())
	os.MkdirAll("config", 0o755)
	os.WriteFile(defaultPolicyPath, []byte("original\n"), 0o644)
	body := `{"groups":[],"defaults":{"default_action":"DENY"}}`

	dbOpt, mock := withDB(t)
	mock.ExpectExec("INSERT INTO log_entries").WillReturnError(errors.New("audit down"))
	for name, s := range map[string]*AdminServer{
		"insert-error": newTestServer(withGuard(defaultTestPolicyConfig()), dbOpt),
		"no-audit-db":  newTestServer(withGuard(defaultTestPolicyConfig())),
	} {
		before := s.Guard.GetPolicyConfig()
		rr := putPolicy(t, s, body)
		assertStatus(t, rr, http.StatusServiceUnavailable)
		if !strings.Contains(rr.Body.String(), governanceAuditUnavailableCode) {
			t.Fatalf("%s: expected audit-unavailable code, got %s", name, rr.Body.String())
		}
		if s.Guard.GetPolicyConfig() != before {
			t.Fatalf("%s: memory changed without audit", name)
		}
		if got, _ := os.ReadFile(defaultPolicyPath); string(got) != "original\n" {
			t.Fatalf("%s: file changed without audit", name)
		}
	}
}

func TestUpdatePolicyWriteFailureLeavesMemoryUnchanged(t *testing.T) {
	t.Chdir(t.TempDir()) // no config/ directory: the atomic write cannot create its temp file
	dbOpt, mock := withDB(t)
	expectAudits(mock, 2) // requested + failed
	s := newTestServer(withGuard(defaultTestPolicyConfig()), dbOpt)
	before := s.Guard.GetPolicyConfig()
	rr := putPolicy(t, s, `{"groups":[],"defaults":{"default_action":"DENY"}}`)
	assertStatus(t, rr, http.StatusInternalServerError)
	if s.Guard.GetPolicyConfig() != before {
		t.Fatal("write failure swapped memory")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

type auditContextCapture struct{ values *[]string }

func (c auditContextCapture) Match(v driver.Value) bool {
	if b, ok := v.([]byte); ok {
		*c.values = append(*c.values, string(b))
	}
	return true
}

func TestUpdatePolicySuccessAuditsDigestsNotBody(t *testing.T) {
	t.Chdir(t.TempDir())
	os.MkdirAll("config", 0o755)
	dbOpt, mock := withDB(t)
	var contexts []string
	for i := 0; i < 2; i++ {
		mock.ExpectExec("INSERT INTO log_entries").
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), auditContextCapture{&contexts}).
			WillReturnResult(sqlmock.NewResult(1, 1))
	}
	s := newTestServer(withGuard(defaultTestPolicyConfig()), dbOpt)
	rr := putPolicy(t, s, `{"groups":[{"name":"secret-group-name","targets":["team:zeta"],"rules":[{"intent":"pay","action":"REQUIRE_APPROVAL"}]}],"defaults":{"default_action":"DENY"}}`)
	assertStatus(t, rr, http.StatusOK)
	written, err := os.ReadFile(defaultPolicyPath)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(written)
	digest := hex.EncodeToString(sum[:])
	if !strings.Contains(rr.Body.String(), digest) || !strings.Contains(rr.Body.String(), `"audit_id"`) {
		t.Fatalf("response must carry the file digest and audit id: %s", rr.Body.String())
	}
	if len(contexts) != 2 {
		t.Fatalf("expected requested + applied audit rows, got %d", len(contexts))
	}
	for i, status := range []string{"requested", "applied"} {
		ctx := contexts[i]
		for _, want := range []string{`"governance_policy_update"`, `"new_digest":"` + digest, `"result_status":"` + status, `"group_count":1`, `"actor_identity"`} {
			if !strings.Contains(ctx, want) {
				t.Fatalf("audit %d missing %s: %s", i, want, ctx)
			}
		}
		if strings.Contains(ctx, "secret-group-name") || strings.Contains(ctx, "team:zeta") {
			t.Fatalf("audit leaked the policy body: %s", ctx)
		}
	}
}

// ── Degraded policy: 503, status row, recovery by PUT ──────────────

func TestDegradedPolicyRecoversAfterValidPut(t *testing.T) {
	t.Chdir(t.TempDir())
	os.MkdirAll("config", 0o755)
	dbOpt, mock := withDB(t)
	expectAudits(mock, 2)
	s := newTestServer(dbOpt)
	s.Guard = governance.NewDegradedGuard(errors.New("open /secret/raw-path.yaml: permission denied"))

	rr := doAuthenticatedRequest(t, http.HandlerFunc(s.handleGetPolicy), "GET", "/api/v1/governance/policy", "")
	assertStatus(t, rr, http.StatusServiceUnavailable)
	body := rr.Body.String()
	if !strings.Contains(body, governancePolicyUnavailableCode) || !strings.Contains(body, "recommended_action") || strings.Contains(body, "raw-path") {
		t.Fatalf("unexpected degraded body: %s", body)
	}
	row := s.governanceServiceStatus()
	if row.Status != "degraded" || !strings.HasPrefix(row.Detail, governance.PolicyUnavailableCode) || strings.Contains(row.Detail, "raw-path") {
		t.Fatalf("unexpected degraded status row: %+v", row)
	}
	if required, _ := s.Guard.PostureRequiresApproval("delivery-posture-lean", nil, nil); !required {
		t.Fatal("degraded guard must require posture approval")
	}

	assertStatus(t, putPolicy(t, s, `{"groups":[],"defaults":{"default_action":"DENY"}}`), http.StatusOK)
	if s.Guard.Degraded() || s.governanceServiceStatus().Status != "online" {
		t.Fatal("valid PUT must restore online governance")
	}
	assertStatus(t, doAuthenticatedRequest(t, http.HandlerFunc(s.handleGetPolicy), "GET", "/api/v1/governance/policy", ""), http.StatusOK)
}

// ── Approval decisions ─────────────────────────────────────────────

func TestResolveApprovalConcurrentDoubleApprove(t *testing.T) {
	s, mock := approvalServer(t)
	mock.MatchExpectationsInOrder(false)
	expectAudits(mock, 4)
	seedPending(s, "req-dup")
	var published int64
	sub, err := s.NC.Subscribe("swarm.team.team-1.agent.agent-1.output", func(*nats.Msg) { atomic.AddInt64(&published, 1) })
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Unsubscribe()
	s.NC.Flush()

	mux := setupMux(t, resolvePattern, s.handleResolveApproval)
	codes := make([]int, 2)
	var wg sync.WaitGroup
	for i := range codes {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			codes[i] = doAuthenticatedRequest(t, mux, "POST", "/api/v1/governance/resolve/req-dup", `{"action":"APPROVE"}`).Code
		}(i)
	}
	wg.Wait()
	s.NC.Flush()
	time.Sleep(100 * time.Millisecond)
	if !(codes[0] == 200 && codes[1] == 404 || codes[0] == 404 && codes[1] == 200) {
		t.Fatalf("expected exactly one 200 and one 404, got %v", codes)
	}
	if got := atomic.LoadInt64(&published); got != 1 {
		t.Fatalf("expected exactly one re-publish, got %d", got)
	}
}

func TestAdminApprovalActionRejectsUnknownDecision(t *testing.T) {
	s := newTestServer(withGuard(defaultTestPolicyConfig()))
	seedPending(s, "req-1")
	for _, body := range []string{`{"action":"DENY"}`, `{"action":""}`, `{"action":"approve"}`, `{}`} {
		rr := doAuthenticatedRequest(t, http.HandlerFunc(s.handleApprovalAction), "POST", "/admin/approvals/req-1", body)
		assertStatus(t, rr, http.StatusBadRequest)
	}
	if _, ok := s.Guard.PendingRequest("req-1"); !ok {
		t.Fatal("unknown decisions must leave the request pending")
	}
}

func TestApprovalRoutesNilGuardOrRouterReturn503(t *testing.T) {
	for name, s := range map[string]*AdminServer{
		"nil-guard":  newTestServer(),
		"nil-router": newTestServer(withGuard(defaultTestPolicyConfig())),
	} {
		if s.Guard != nil {
			seedPending(s, "req-1")
		}
		rr := doAuthenticatedRequest(t, setupMux(t, resolvePattern, s.handleResolveApproval), "POST", "/api/v1/governance/resolve/req-1", `{"action":"APPROVE"}`)
		assertStatus(t, rr, http.StatusServiceUnavailable)
		// admin.go used to dereference a nil Guard here and panic.
		rr = doAuthenticatedRequest(t, http.HandlerFunc(s.handleApprovalAction), "POST", "/admin/approvals/req-1", `{"action":"APPROVE"}`)
		assertStatus(t, rr, http.StatusServiceUnavailable)
		if s.Guard != nil {
			if _, ok := s.Guard.PendingRequest("req-1"); !ok {
				t.Fatalf("%s: 503 must leave the request pending", name)
			}
		}
	}
}

func TestResolveApprovalAuditFailureKeepsPending(t *testing.T) {
	s, mock := approvalServer(t)
	mock.ExpectExec("INSERT INTO log_entries").WillReturnError(errors.New("audit down"))
	seedPending(s, "req-1")
	rr := doAuthenticatedRequest(t, http.HandlerFunc(s.handleApprovalAction), "POST", "/admin/approvals/req-1", `{"action":"APPROVE"}`)
	assertStatus(t, rr, http.StatusServiceUnavailable)
	if _, ok := s.Guard.PendingRequest("req-1"); !ok {
		t.Fatal("audit failure must leave the request pending")
	}
}

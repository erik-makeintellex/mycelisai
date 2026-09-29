package server

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/governance"
)

func expectAudits(mock sqlmock.Sqlmock, n int) {
	for i := 0; i < n; i++ {
		mock.ExpectExec("INSERT INTO log_entries").WillReturnResult(sqlmock.NewResult(1, 1))
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

// auditRecorder captures the context JSON of every log_entries insert.
type auditRecorder struct {
	mu   sync.Mutex
	ctxs []map[string]any
}

func (a *auditRecorder) Match(v driver.Value) bool {
	b, _ := v.([]byte)
	var m map[string]any
	if json.Unmarshal(b, &m) != nil {
		return false
	}
	a.mu.Lock()
	a.ctxs = append(a.ctxs, m)
	a.mu.Unlock()
	return true
}

// A2b item 4: concurrent PUTs read previous_digest under applyMu, so each
// requested audit's previous_digest is the preceding PUT's new_digest.
func TestUpdatePolicyConcurrentPreviousDigestChain(t *testing.T) {
	t.Chdir(t.TempDir())
	os.MkdirAll("config", 0o755)
	const n = 12
	dbOpt, mock := withDB(t)
	mock.MatchExpectationsInOrder(false)
	rec := &auditRecorder{}
	a := sqlmock.AnyArg()
	for i := 0; i < 2*n; i++ {
		mock.ExpectExec("INSERT INTO log_entries").WithArgs(a, a, a, a, a, a, a, rec).
			WillReturnResult(sqlmock.NewResult(1, 1))
	}
	s := newTestServer(withGuard(defaultTestPolicyConfig()), dbOpt)
	_, initial, _ := canonicalPolicy(s.Guard.GetPolicyConfig())
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			body := fmt.Sprintf(`{"groups":[{"name":"g%d","targets":["*"],"rules":[{"intent":"x%d","action":"DENY"}]}],"defaults":{"default_action":"ALLOW"}}`, i, i)
			if rr := putPolicy(t, s, body); rr.Code != http.StatusOK {
				t.Errorf("PUT %d: %d %s", i, rr.Code, rr.Body.String())
			}
		}(i)
	}
	wg.Wait()
	// sqlmock may evaluate an argument matcher more than once per insert; each
	// PUT has a unique new_digest, so keep the first sighting in insert order.
	var requested []map[string]any
	once := map[any]bool{}
	for _, c := range rec.ctxs {
		if c["result_status"] == "requested" && !once[c["new_digest"]] {
			once[c["new_digest"]] = true
			requested = append(requested, c)
		}
	}
	if len(requested) != n {
		t.Fatalf("expected %d requested audits, got %d", n, len(requested))
	}
	prev, seen := initial, map[any]bool{}
	for i, c := range requested {
		if c["previous_digest"] != prev {
			t.Fatalf("requested audit %d: previous_digest %v, want %v", i, c["previous_digest"], prev)
		}
		if seen[c["previous_digest"]] {
			t.Fatalf("duplicate previous_digest at %d", i)
		}
		seen[c["previous_digest"]] = true
		prev, _ = c["new_digest"].(string)
	}
	if _, live, _ := canonicalPolicy(s.Guard.GetPolicyConfig()); live != prev {
		t.Fatal("live policy must be the last requested new_digest")
	}
}

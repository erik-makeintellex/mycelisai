package server

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/cognitive"
)

// B1R-B F6: usage and policy reads are scoped. A root admin with
// cognitive:read or cognitive:write reads any ref; anyone else reads a team's
// usage only when the persisted team ownership binding plus an active group
// membership proves they belong to it. Agent and run usage have no ownership
// record yet, so non-admins are denied there.

const b1rbMemberID = "7a4f2c9e-3b1d-4e8a-9c6f-2d5b8e1a0f34"

var b1rbMemberQuery = `SELECT EXISTS \(SELECT 1 FROM runtime_team_manifests`

// b1rbUsageServer charges 1124 tokens to team secret-team, agent x, run r-1.
func b1rbUsageServer(t *testing.T, opts ...func(*AdminServer)) (*AdminServer, *http.ServeMux) {
	t.Helper()
	router := tokenBudgetRouter()
	router.Adapters["local"] = &b1rbCountingProvider{}
	correlation := cognitive.InferenceCorrelation{RunID: "r-1", TeamID: "secret-team", AgentID: "x"}
	ctx := cognitive.WithExecutionMeter(context.Background(), cognitive.NewExecutionMeter(cognitive.ExecutionKindAgentTurn, correlation))
	if _, err := router.InferWithContract(ctx, cognitive.InferRequest{Profile: "chat", Prompt: "p", Correlation: correlation}); err != nil {
		t.Fatal(err)
	}
	s := newTestServer(append([]func(*AdminServer){func(s *AdminServer) { s.Cognitive = router }}, opts...)...)
	return s, tokenBudgetMux(s)
}

func b1rbAssertUsageDenied(t *testing.T, status int, body string) {
	t.Helper()
	if status != http.StatusForbidden || !strings.Contains(body, codeTokenBudgetUsageForbidden) || strings.Contains(body, `"used"`) {
		t.Fatalf("want 403 %s without usage, got %d %s", codeTokenBudgetUsageForbidden, status, body)
	}
}

func TestTokenBudgetUsageScope_NonAdminsAreDeniedWithoutProof(t *testing.T) {
	callers := map[string]*RequestIdentity{
		"standard user":                 {UserID: b1rbMemberID, Role: "user", Scopes: []string{"cognitive:read", "cognitive:write"}},
		"admin without cognitive scope": {UserID: b1rbMemberID, Role: "admin", Scopes: []string{"config_documents:write"}},
	}
	queries := []string{"team_id=secret-team", "agent_id=x", "run_id=r-1", "team_id=no-such-team"}
	for name, identity := range callers {
		for _, query := range queries {
			t.Run(name+" "+query+" without DB", func(t *testing.T) {
				_, mux := b1rbUsageServer(t)
				rr := doAuthenticatedRequestAs(t, mux, http.MethodGet, "/api/v1/cognitive/budgets/usage?"+query, "", identity)
				b1rbAssertUsageDenied(t, rr.Code, rr.Body.String())
			})
		}
		// Agent and run scopes never consult membership: nothing proves ownership.
		for _, query := range []string{"agent_id=x", "run_id=r-1"} {
			t.Run(name+" "+query+" with DB", func(t *testing.T) {
				dbOpt, mock := withDirectDB(t)
				_, mux := b1rbUsageServer(t, dbOpt)
				rr := doAuthenticatedRequestAs(t, mux, http.MethodGet, "/api/v1/cognitive/budgets/usage?"+query, "", identity)
				b1rbAssertUsageDenied(t, rr.Code, rr.Body.String())
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}

func TestTokenBudgetUsageScope_TeamMembershipIsProvenFromPersistedState(t *testing.T) {
	user := &RequestIdentity{UserID: b1rbMemberID, Role: "user"}
	t.Run("member reads own team", func(t *testing.T) {
		dbOpt, mock := withDirectDB(t)
		mock.ExpectQuery(b1rbMemberQuery).WithArgs("secret-team", b1rbMemberID).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
		_, mux := b1rbUsageServer(t, dbOpt)
		rr := doAuthenticatedRequestAs(t, mux, http.MethodGet, "/api/v1/cognitive/budgets/usage?team_id=secret-team", "", user)
		assertStatus(t, rr, http.StatusOK)
		if data := tokenBudgetData(t, rr); data["ref"] != "secret-team" || data["used"] != float64(1124) {
			t.Fatalf("member usage = %v", data)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("non-member and unknown team are indistinguishable", func(t *testing.T) {
		for _, team := range []string{"secret-team", "no-such-team"} {
			dbOpt, mock := withDirectDB(t)
			mock.ExpectQuery(b1rbMemberQuery).WithArgs(team, b1rbMemberID).
				WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
			_, mux := b1rbUsageServer(t, dbOpt)
			rr := doAuthenticatedRequestAs(t, mux, http.MethodGet, "/api/v1/cognitive/budgets/usage?team_id="+team, "", user)
			b1rbAssertUsageDenied(t, rr.Code, rr.Body.String())
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		}
	})
	t.Run("non-UUID principal is denied without a query", func(t *testing.T) {
		dbOpt, mock := withDirectDB(t)
		_, mux := b1rbUsageServer(t, dbOpt)
		rr := doAuthenticatedRequestAs(t, mux, http.MethodGet, "/api/v1/cognitive/budgets/usage?team_id=secret-team", "",
			&RequestIdentity{UserID: "secret-team", Role: "user"})
		b1rbAssertUsageDenied(t, rr.Code, rr.Body.String())
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("membership store error fails closed", func(t *testing.T) {
		dbOpt, mock := withDirectDB(t)
		mock.ExpectQuery(b1rbMemberQuery).WillReturnError(errors.New("connection reset"))
		_, mux := b1rbUsageServer(t, dbOpt)
		rr := doAuthenticatedRequestAs(t, mux, http.MethodGet, "/api/v1/cognitive/budgets/usage?team_id=secret-team", "", user)
		if rr.Code != http.StatusServiceUnavailable || strings.Contains(rr.Body.String(), `"used"`) || strings.Contains(rr.Body.String(), "connection reset") {
			t.Fatalf("store error = %d %s", rr.Code, rr.Body.String())
		}
	})
}

func TestTokenBudgetUsageScope_ScopedRootAdminReadsAnyRef(t *testing.T) {
	for _, scope := range []string{"cognitive:read", "cognitive:write"} {
		admin := &RequestIdentity{UserID: "a-1", Role: "admin", Scopes: []string{scope}}
		dbOpt, mock := withDirectDB(t)
		_, mux := b1rbUsageServer(t, dbOpt)
		for query, want := range map[string]float64{"team_id=secret-team": 1124, "agent_id=x": 1124, "run_id=r-1": 1124, "team_id=no-such-team": 0} {
			rr := doAuthenticatedRequestAs(t, mux, http.MethodGet, "/api/v1/cognitive/budgets/usage?"+query, "", admin)
			assertStatus(t, rr, http.StatusOK)
			if data := tokenBudgetData(t, rr); data["used"] != want {
				t.Fatalf("%s %s usage = %v", scope, query, data)
			}
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("admin read consulted membership: %v", err)
		}
	}
}

func TestGetTokenBudgetsScope_ProviderIdentityIsAdminOnly(t *testing.T) {
	s := newTestServer(func(s *AdminServer) { s.Cognitive = tokenBudgetRouter() })
	for _, identity := range []*RequestIdentity{
		{UserID: "u-1", Role: "user", Scopes: []string{"cognitive:read"}},
		{UserID: "u-2", Role: "admin", Scopes: []string{"config_documents:write"}},
	} {
		rr := doAuthenticatedRequestAs(t, tokenBudgetMux(s), http.MethodGet, "/api/v1/cognitive/budgets", "", identity)
		assertStatus(t, rr, http.StatusOK)
		body := rr.Body.String()
		if strings.Contains(body, "provider_id") || strings.Contains(body, "model_id") || strings.Contains(body, "qwen3:14b") || strings.Contains(body, `"claude"`) {
			t.Fatalf("%s/%v sees provider identity: %s", identity.Role, identity.Scopes, body)
		}
		if data := tokenBudgetData(t, rr); data["classes"] == nil || data["global"] == nil || data["default_class"] != "local_large" {
			t.Fatalf("effective table missing: %v", data)
		}
	}
	for _, scope := range []string{"cognitive:read", "cognitive:write"} {
		rr := doAuthenticatedRequestAs(t, tokenBudgetMux(s), http.MethodGet, "/api/v1/cognitive/budgets", "", &RequestIdentity{UserID: "a-1", Role: "admin", Scopes: []string{scope}})
		if providers, _ := tokenBudgetData(t, rr)["providers"].([]any); len(providers) != 2 {
			t.Fatalf("%s admin providers = %v", scope, providers)
		}
	}
}

// Real PostgreSQL: the membership query runs against the current schema and an
// unbound team proves nothing (needs MYCELIS_TOKEN_BUDGET_TEST_DSN).
func TestTokenBudgetUsageScope_RealDBMembershipQueryFailsClosed(t *testing.T) {
	db := openTokenBudgetRealDB(t)
	var member bool
	if err := db.QueryRow(tokenBudgetTeamMemberSQL, "b1rb-unbound-team", b1rbMemberID).Scan(&member); err != nil || member {
		t.Fatalf("membership query = %v, %v", member, err)
	}
}

// Kept from the review probe: bodies that smuggle a float, a traversal ref,
// or a write without durable storage all fail closed.
func TestTokenBudgetOverridesScope_ProbeCasesFailClosed(t *testing.T) {
	s := newTestServer(func(s *AdminServer) { s.Cognitive = tokenBudgetRouter() })
	for path, want := range map[string]int{
		"/api/v1/cognitive/budgets/overrides/team/t":        http.StatusBadRequest,
		"/api/v1/cognitive/budgets/overrides/class/../../x": http.StatusTemporaryRedirect,
	} {
		if rr := doAuthenticatedRequest(t, tokenBudgetMux(s), http.MethodPut, path, `{"per_execution":1e3}`); rr.Code != want {
			t.Fatalf("PUT %s = %d, want %d", path, rr.Code, want)
		}
	}
	rr := doAuthenticatedRequest(t, tokenBudgetMux(s), http.MethodPut, "/api/v1/cognitive/budgets/overrides/team/t", `{"per_execution":2048}`)
	assertStatus(t, rr, http.StatusServiceUnavailable)
	if _, applied := s.Cognitive.Budgets.Policy().Overrides.Team["t"]; applied {
		t.Fatal("override applied without durable storage")
	}
}

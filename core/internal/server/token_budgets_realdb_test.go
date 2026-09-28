package server

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"strings"
	"testing"

	_ "github.com/lib/pq"
	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/pkg/protocol"
)

// Real PostgreSQL proof. MYCELIS_TOKEN_BUDGET_TEST_DSN must point at a
// disposable database with core/migrations/001_current_schema.sql applied.
func openTokenBudgetRealDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := strings.TrimSpace(os.Getenv("MYCELIS_TOKEN_BUDGET_TEST_DSN"))
	if dsn == "" {
		t.Skip("requires disposable MYCELIS_TOKEN_BUDGET_TEST_DSN")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	for _, statement := range []string{
		`DELETE FROM config_document_activation_history WHERE document_id = 'token-budget-overrides'`,
		`DELETE FROM config_document_activations WHERE document_id = 'token-budget-overrides'`,
		`DELETE FROM config_documents WHERE document_id = 'token-budget-overrides'`,
		`DELETE FROM token_usage_ledger`,
	} {
		if _, err := db.Exec(statement); err != nil {
			t.Fatalf("clean: %v", err)
		}
	}
	return db
}

func TestTokenBudgetOverrides_RealDBLifecycleIsDurableAndAudited(t *testing.T) {
	db := openTokenBudgetRealDB(t)
	router := tokenBudgetRouter()
	router.Budgets = cognitive.NewBudgetGovernor(protocol.DefaultTokenBudgetPolicySpec(), cognitive.NewPostgresBudgetLedger(db))
	s := newTestServer(func(s *AdminServer) { s.DB = db; s.Cognitive = router })
	mux := tokenBudgetMux(s)
	subject := cognitive.BudgetSubject{AgentID: "coder", TeamID: "team-a", Class: protocol.TokenBudgetClassLocalLarge}

	// Team beats class; a spoofed actor in the body is ignored.
	rr := doAuthenticatedRequest(t, mux, http.MethodPut, "/api/v1/cognitive/budgets/overrides/team/team-a", `{"per_execution":8192,"actor":"root-spoof","created_by":"system:bootstrap"}`)
	assertStatus(t, rr, http.StatusOK)
	data := tokenBudgetData(t, rr)
	if data["changed"] != true || data["after"].(map[string]any)["per_execution"] != float64(8192) || data["audit_event_id"] == "" || data["record_id"] == "" {
		t.Fatalf("put team = %v", data)
	}
	if got, _ := cognitive.ResolveBudgetLimits(router.Budgets.Policy(), subject); got.PerExecution != 8192 {
		t.Fatalf("team override not applied: %+v", got)
	}
	// Agent beats team.
	assertStatus(t, doAuthenticatedRequest(t, mux, http.MethodPut, "/api/v1/cognitive/budgets/overrides/agent/coder", `{"per_execution":4096}`), http.StatusOK)
	if got, _ := cognitive.ResolveBudgetLimits(router.Budgets.Policy(), subject); got.PerExecution != 4096 {
		t.Fatalf("agent override not applied: %+v", got)
	}

	// A restart reloads the same effective policy from the active revision.
	reloaded, err := loadTokenBudgetPolicy(context.Background(), db)
	if err != nil || reloaded.Overrides.Agent["coder"].PerExecution != 4096 || reloaded.Overrides.Team["team-a"].PerExecution != 8192 {
		t.Fatalf("reload = %+v err=%v", reloaded.Overrides, err)
	}

	// DELETE restores the default at both levels.
	assertStatus(t, doAuthenticatedRequest(t, mux, http.MethodDelete, "/api/v1/cognitive/budgets/overrides/agent/coder", ""), http.StatusOK)
	assertStatus(t, doAuthenticatedRequest(t, mux, http.MethodDelete, "/api/v1/cognitive/budgets/overrides/team/team-a", ""), http.StatusOK)
	if got, _ := cognitive.ResolveBudgetLimits(router.Budgets.Policy(), subject); got.PerExecution != 64000 {
		t.Fatalf("delete did not restore the class default: %+v", got)
	}
	again := tokenBudgetData(t, doAuthenticatedRequest(t, mux, http.MethodDelete, "/api/v1/cognitive/budgets/overrides/team/team-a", ""))
	if again["changed"] != false {
		t.Fatalf("idempotent delete = %v", again)
	}

	var revisions, history int
	var actors string
	if err := db.QueryRow(`SELECT count(*), string_agg(DISTINCT created_by, ',') FROM config_documents WHERE document_id='token-budget-overrides' AND kind='TokenBudgetPolicy' AND scope_kind='operator' AND source_kind='api'`).Scan(&revisions, &actors); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT count(*) FROM config_document_activation_history WHERE document_id='token-budget-overrides' AND audit_event_id IS NOT NULL`).Scan(&history); err != nil {
		t.Fatal(err)
	}
	if revisions != 4 || history != 4 || actors != localAdminIdentityForTest().UserID {
		t.Fatalf("revisions=%d history=%d actors=%q, want 4 audited api revisions by the caller", revisions, history, actors)
	}
	var audits int
	if err := db.QueryRow(`SELECT count(*) FROM log_entries WHERE level='audit' AND context->>'action'='token_budget_changed' AND context->>'user' IS NOT NULL AND context::text NOT LIKE '%root-spoof%'`).Scan(&audits); err != nil {
		t.Fatal(err)
	}
	if audits != 5 {
		t.Fatalf("token_budget_changed audit events = %d, want 5 (4 changes + 1 noop)", audits)
	}
}

func TestPostgresBudgetLedger_TotalsAndReservationFlag(t *testing.T) {
	db := openTokenBudgetRealDB(t)
	ledger := cognitive.NewPostgresBudgetLedger(db)
	spec := protocol.DefaultTokenBudgetPolicySpec()
	governor := cognitive.NewBudgetGovernor(spec, ledger)
	provider := &draftUsageProvider{used: 1234}
	router := tokenBudgetRouter()
	router.Adapters["local"] = provider
	router.Budgets = governor
	correlation := cognitive.InferenceCorrelation{RunID: "run-pg", TeamID: "team-pg", AgentID: "agent-pg"}
	for i := 0; i < 3; i++ {
		ctx := cognitive.WithExecutionMeter(context.Background(), cognitive.NewExecutionMeter(cognitive.ExecutionKindAgentTurn, correlation))
		if _, err := router.InferWithContract(ctx, cognitive.InferRequest{Profile: "chat", Prompt: "x", Correlation: correlation}); err != nil {
			t.Fatal(err)
		}
	}
	// A fresh governor (a restart) reads the durable period totals.
	restarted := cognitive.NewBudgetGovernor(spec, ledger)
	usage := restarted.Usage(context.Background(), "", protocol.TokenBudgetScopeTeamDay, "team-pg", protocol.TokenBudgetLimits{PerTeamDay: 100000, WarnPct: 80})
	if usage.Used != 3702 || usage.Period != protocol.TokenBudgetPeriodUTCDay || !usage.UsageReported {
		t.Fatalf("durable usage = %+v", usage)
	}
	var rows int
	if err := db.QueryRow(`SELECT count(*) FROM token_usage_ledger WHERE run_id='run-pg' AND outcome='charged' AND usage_reported AND budget_class='local_large' AND execution_kind='agent_turn'`).Scan(&rows); err != nil || rows != 3 {
		t.Fatalf("ledger rows = %d err=%v", rows, err)
	}
	if _, err := db.Exec(`INSERT INTO token_usage_ledger (execution_id, execution_kind, provider_id, model_id, budget_class, total_tokens, usage_reported, outcome) VALUES ('x','unbounded','p','m','c',1,true,'charged')`); err == nil {
		t.Fatal("ledger accepted an unknown execution_kind")
	}
}

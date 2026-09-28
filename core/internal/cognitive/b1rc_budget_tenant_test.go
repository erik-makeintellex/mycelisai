package cognitive

import (
	"context"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

// Condition 2 (MED): counters, ledger rows and usage reads are keyed by
// tenant + scope + ref. QA showed two tenants sharing one team_id counter.

func b1rcTeamSpec(limit int) protocol.TokenBudgetPolicySpec {
	spec := protocol.DefaultTokenBudgetPolicySpec()
	spec.Overrides.Team = map[string]protocol.TokenBudgetLimits{"T": {PerTeamDay: limit}}
	return spec
}

func TestB1RC_TenantsNeverShareCountersOrUsage(t *testing.T) {
	ledger := &b1raLedger{}
	r := b1raRouter(&b1rcAdapter{prompt: 10}, b1rcTeamSpec(4200), ledger)
	limits := r.Budgets.Limits(BudgetSubject{TeamID: "T", Class: b1raLocal.Class})
	def := InferenceCorrelation{TeamID: "T", AgentID: "A", RunID: "R"}
	t2 := InferenceCorrelation{TenantID: "t2", TeamID: "T", AgentID: "A", RunID: "R"}

	// Tenant default exhausts team T (2058 + 2058 + refused).
	for i := 0; i < 2; i++ {
		if err := b1raInfer(r, def); err != nil {
			t.Fatal(err)
		}
	}
	if stop := AsTokenBudgetExhausted(b1raInfer(r, def)); stop == nil {
		t.Fatal("tenant default was not stopped at its team limit")
	}
	// Tenant t2's team T is untouched and still admitted.
	if err := b1raInfer(r, t2); err != nil {
		t.Fatalf("tenant t2 shares tenant default's counter: %v", err)
	}
	ctx := context.Background()
	for _, tc := range []struct {
		tenant string
		want   int
	}{{"", 4116}, {DefaultBudgetTenant, 4116}, {"t2", 2058}, {"t3", 0}} {
		for _, scope := range []struct{ scope, ref string }{{protocol.TokenBudgetScopeTeamDay, "T"}, {protocol.TokenBudgetScopeAgentDay, "A"}, {protocol.TokenBudgetScopeRun, "R"}} {
			if u := r.Budgets.Usage(ctx, tc.tenant, scope.scope, scope.ref, limits); u.Used != tc.want {
				t.Fatalf("tenant %q %s used = %d, want %d", tc.tenant, scope.scope, u.Used, tc.want)
			}
		}
	}
	// Ledger rows carry the tenant, and a restart reads each tenant apart.
	tenants := map[string]int{}
	for _, e := range ledger.mem.entries {
		if e.Outcome == TokenLedgerOutcomeCharged {
			tenants[e.TenantID] += e.TotalTokens
		}
	}
	if tenants[DefaultBudgetTenant] != 4116 || tenants["t2"] != 2058 || len(tenants) != 2 {
		t.Fatalf("ledger tenants = %v", tenants)
	}
	restarted := NewBudgetGovernor(b1rcTeamSpec(4200), ledger)
	if u := restarted.Usage(ctx, "t2", protocol.TokenBudgetScopeTeamDay, "T", limits); u.Used != 2058 || u.Period != protocol.TokenBudgetPeriodUTCDay {
		t.Fatalf("restarted t2 usage = %+v", u)
	}
	if u := restarted.Usage(ctx, "", protocol.TokenBudgetScopeTeamDay, "T", limits); u.Used != 4116 {
		t.Fatalf("restarted default usage = %+v", u)
	}
}

// A tenant-only correlation is not a scope: the call falls back to the meter's
// correlation, and an empty tenant is "default".
func TestB1RC_BudgetTenantNormalization(t *testing.T) {
	for in, want := range map[string]string{"": DefaultBudgetTenant, "  ": DefaultBudgetTenant, "default": DefaultBudgetTenant, "t2": "t2", " t2 ": "t2"} {
		if got := BudgetTenant(in); got != want {
			t.Fatalf("BudgetTenant(%q) = %q, want %q", in, got, want)
		}
	}
	r := b1raRouter(&b1rcAdapter{prompt: 10}, protocol.DefaultTokenBudgetPolicySpec(), &b1raLedger{})
	meter := NewExecutionMeter(ExecutionKindAgentTurn, InferenceCorrelation{TenantID: "t2", TeamID: "T"})
	if _, err := r.InferWithContract(WithExecutionMeter(context.Background(), meter), InferRequest{Profile: "chat", Prompt: "p", Correlation: InferenceCorrelation{TenantID: "t2"}}); err != nil {
		t.Fatal(err)
	}
	limits := r.Budgets.Limits(b1raLocal)
	if u := r.Budgets.Usage(context.Background(), "t2", protocol.TokenBudgetScopeTeamDay, "T", limits); u.Used != 2058 {
		t.Fatalf("tenant-only correlation did not fall back to the meter: used %d", u.Used)
	}
}

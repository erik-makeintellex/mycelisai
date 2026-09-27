package cognitive

import (
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

func TestResolveBudgetClass_DefaultsByDataBoundaryAndExplicitClassWins(t *testing.T) {
	cases := []struct {
		name string
		cfg  ProviderConfig
		want string
	}{
		{"local_only", ProviderConfig{DataBoundary: DataBoundaryLocalOnly}, protocol.TokenBudgetClassLocalLarge},
		{"empty boundary fails closed to local", ProviderConfig{}, protocol.TokenBudgetClassLocalLarge},
		{"leaves_org fails toward tighter", ProviderConfig{DataBoundary: DataBoundaryLeavesOrg}, protocol.TokenBudgetClassHostedPremium},
		{"explicit class wins", ProviderConfig{DataBoundary: DataBoundaryLeavesOrg, BudgetClass: protocol.TokenBudgetClassHostedStandard}, protocol.TokenBudgetClassHostedStandard},
		{"explicit local_small", ProviderConfig{BudgetClass: protocol.TokenBudgetClassLocalSmall}, protocol.TokenBudgetClassLocalSmall},
		{"unknown class ignored", ProviderConfig{DataBoundary: DataBoundaryLeavesOrg, BudgetClass: "unlimited"}, protocol.TokenBudgetClassHostedPremium},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ResolveBudgetClass(tc.cfg); got != tc.want {
				t.Fatalf("class = %q, want %q", got, tc.want)
			}
		})
	}
	spec := protocol.DefaultTokenBudgetPolicySpec()
	local, _ := ResolveBudgetLimits(spec, BudgetSubject{Class: ResolveBudgetClass(ProviderConfig{DataBoundary: DataBoundaryLocalOnly})})
	if local.PerExecution != 64000 || local.PerRun != 256000 || local.PerTeamDay != 2000000 {
		t.Fatalf("local_large limits = %+v", local)
	}
	hosted, _ := ResolveBudgetLimits(spec, BudgetSubject{Class: ResolveBudgetClass(ProviderConfig{DataBoundary: DataBoundaryLeavesOrg})})
	if hosted.PerExecution != 24000 {
		t.Fatalf("hosted_premium per_execution = %d, want 24000", hosted.PerExecution)
	}
	unknown, _ := ResolveBudgetLimits(spec, BudgetSubject{Class: "not-a-class"})
	if unknown != spec.Global {
		t.Fatalf("unknown class = %+v, want global %+v", unknown, spec.Global)
	}
}

func TestResolveBudgetLimits_MostSpecificWinsPerField(t *testing.T) {
	spec := protocol.DefaultTokenBudgetPolicySpec()
	subject := BudgetSubject{AgentID: "coder", TeamID: "team-a", Profile: "chat", Class: protocol.TokenBudgetClassLocalLarge}

	base, sources := ResolveBudgetLimits(spec, subject)
	if base.PerExecution != 64000 || sources["per_execution"] != "class_default" {
		t.Fatalf("baseline = %+v sources=%v", base, sources)
	}

	spec.Overrides.Class = map[string]protocol.TokenBudgetLimits{protocol.TokenBudgetClassLocalLarge: {PerRun: 200000}}
	spec.Overrides.Profile = map[string]protocol.TokenBudgetLimits{"chat": {PerExecution: 50000}}
	spec.Overrides.Team = map[string]protocol.TokenBudgetLimits{"team-a": {PerExecution: 40000, PerTeamDay: 900000}}
	got, sources := ResolveBudgetLimits(spec, subject)
	if got.PerExecution != 40000 || sources["per_execution"] != "team" {
		t.Fatalf("team override must beat profile/class: %+v %v", got, sources)
	}
	if got.PerRun != 200000 || sources["per_run"] != "class" {
		t.Fatalf("class override must beat class default: %+v %v", got, sources)
	}
	if got.PerTeamDay != 900000 || got.PerAgentDay != 1000000 || got.WarnPct != 80 {
		t.Fatalf("unset fields must fall through: %+v", got)
	}

	spec.Overrides.Agent = map[string]protocol.TokenBudgetLimits{"coder": {PerExecution: 2048}}
	got, sources = ResolveBudgetLimits(spec, subject)
	if got.PerExecution != 2048 || sources["per_execution"] != "agent" || got.PerTeamDay != 900000 {
		t.Fatalf("agent override must beat team per field only: %+v %v", got, sources)
	}
	other, _ := ResolveBudgetLimits(spec, BudgetSubject{AgentID: "writer", TeamID: "team-b", Class: protocol.TokenBudgetClassLocalLarge})
	if other.PerExecution != 64000 || other.PerRun != 200000 {
		t.Fatalf("overrides leaked to another subject: %+v", other)
	}

	delete(spec.Overrides.Agent, "coder")
	delete(spec.Overrides.Team, "team-a")
	restored, _ := ResolveBudgetLimits(spec, subject)
	if restored.PerExecution != 50000 {
		t.Fatalf("removing overrides must restore the next level: %+v", restored)
	}
}

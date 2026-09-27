package protocol

import (
	"encoding/json"
	"testing"
)

func tokenBudgetDocument(t *testing.T, spec TokenBudgetPolicySpec) ConfigDocument {
	t.Helper()
	raw, err := json.Marshal(spec)
	if err != nil {
		t.Fatal(err)
	}
	return ConfigDocument{
		APIVersion: ConfigDocumentAPIVersionV1,
		Kind:       ConfigDocumentKindTokenBudgetPolicy,
		Metadata: ConfigDocumentMetadata{
			ID: TokenBudgetOverridesDocumentID, Name: "Token budget overrides", Version: "1", OwnerID: "operator",
			Scope:      ConfigDocumentScope{Kind: ConfigDocumentScopeOperator, Ref: "default"},
			Enabled:    true,
			Source:     ConfigDocumentSource{Kind: ConfigDocumentSourceAPI, Ref: "api:/api/v1/cognitive/budgets"},
			Governance: ConfigDocumentGovernance{RiskLevel: ConfigDocumentRiskMedium, ApprovalPosture: ApprovalPostureRequired},
		},
		Spec: raw,
	}
}

func TestDefaultTokenBudgetPolicy_MatchesOwnerTable(t *testing.T) {
	spec := DefaultTokenBudgetPolicySpec()
	want := map[string]TokenBudgetLimits{
		TokenBudgetClassLocalLarge:     {PerExecution: 64000, PerRun: 256000, PerTeamDay: 2000000, PerAgentDay: 1000000, WarnPct: 80},
		TokenBudgetClassLocalSmall:     {PerExecution: 24000, PerRun: 96000, PerTeamDay: 1000000, PerAgentDay: 500000, WarnPct: 80},
		TokenBudgetClassHostedStandard: {PerExecution: 32000, PerRun: 128000, PerTeamDay: 300000, PerAgentDay: 150000, WarnPct: 80},
		TokenBudgetClassHostedPremium:  {PerExecution: 24000, PerRun: 96000, PerTeamDay: 150000, PerAgentDay: 75000, WarnPct: 80},
	}
	for class, limits := range want {
		if spec.Classes[class] != limits {
			t.Fatalf("class %s = %+v, want %+v", class, spec.Classes[class], limits)
		}
	}
	global := TokenBudgetLimits{PerExecution: 32000, PerRun: 128000, PerTeamDay: 500000, PerAgentDay: 250000, WarnPct: 80}
	if spec.Global != global {
		t.Fatalf("global = %+v, want %+v", spec.Global, global)
	}
	if issues := ValidateNewConfigDocument(tokenBudgetDocument(t, spec)); len(issues) != 0 {
		t.Fatalf("default policy invalid: %+v", issues)
	}
}

func TestTokenBudgetPolicySpec_RejectsAdversarialValues(t *testing.T) {
	cases := map[string]func(*TokenBudgetPolicySpec){
		"zero global":       func(s *TokenBudgetPolicySpec) { s.Global.PerExecution = 0 },
		"negative override": func(s *TokenBudgetPolicySpec) { s.Overrides.Team = map[string]TokenBudgetLimits{"t": {PerRun: -5}} },
		"below minimum": func(s *TokenBudgetPolicySpec) {
			s.Overrides.Agent = map[string]TokenBudgetLimits{"a": {PerExecution: 1023}}
		},
		"above maximum": func(s *TokenBudgetPolicySpec) {
			s.Overrides.Profile = map[string]TokenBudgetLimits{"chat": {PerTeamDay: 5000001}}
		},
		"unknown class": func(s *TokenBudgetPolicySpec) { s.Classes["unlimited"] = s.Global },
		"unknown class ovr": func(s *TokenBudgetPolicySpec) {
			s.Overrides.Class = map[string]TokenBudgetLimits{"gpu": {PerRun: 4096}}
		},
		"empty override": func(s *TokenBudgetPolicySpec) { s.Overrides.Team = map[string]TokenBudgetLimits{"t": {}} },
		"warn pct 100":   func(s *TokenBudgetPolicySpec) { s.Global.WarnPct = 100 },
		"exec above run": func(s *TokenBudgetPolicySpec) { s.Global.PerExecution = s.Global.PerRun + 1 },
		"run above team day": func(s *TokenBudgetPolicySpec) {
			s.Classes[TokenBudgetClassLocalLarge] = TokenBudgetLimits{PerExecution: 64000, PerRun: 3000000, PerTeamDay: 2000000, PerAgentDay: 1000000, WarnPct: 80}
		},
		"bad ref": func(s *TokenBudgetPolicySpec) {
			s.Overrides.Agent = map[string]TokenBudgetLimits{"../x": {PerRun: 4096}}
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			spec := DefaultTokenBudgetPolicySpec()
			mutate(&spec)
			if issues := ValidateNewConfigDocument(tokenBudgetDocument(t, spec)); len(issues) == 0 {
				t.Fatalf("%s accepted", name)
			}
		})
	}
}

func TestTokenBudgetPolicySpec_RejectsUnknownFields(t *testing.T) {
	document := tokenBudgetDocument(t, DefaultTokenBudgetPolicySpec())
	document.Spec = json.RawMessage(`{"global":{"per_execution":2048,"per_run":4096,"per_team_day":8192,"per_agent_day":8192,"warn_pct":80},"unlimited":true}`)
	if issues := ValidateNewConfigDocument(document); len(issues) == 0 {
		t.Fatal("unknown spec field accepted")
	}
}

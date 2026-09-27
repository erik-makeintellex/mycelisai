package agentry

import (
	"context"
	"errors"
	"testing"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/pkg/protocol"
)

type budgetProvider struct {
	calls int
	corr  cognitive.InferenceCorrelation
}

func (p *budgetProvider) Infer(_ context.Context, _ string, opts cognitive.InferOptions) (*cognitive.InferResponse, error) {
	p.calls++
	p.corr = opts.Correlation
	return &cognitive.InferResponse{Text: "draft that fails verification", PromptTokens: 1500, CompletionTokens: 400, TokensUsed: 1900}, nil
}

func (p *budgetProvider) Probe(context.Context) (bool, error) { return true, nil }

func TestRunner_RunIsOneMeteredExecutionAndStopsHonestly(t *testing.T) {
	spec := protocol.DefaultTokenBudgetPolicySpec()
	spec.Overrides.Agent = map[string]protocol.TokenBudgetLimits{"verifier": {PerExecution: 2048}}
	provider := &budgetProvider{}
	brain := &cognitive.Router{
		Config: &cognitive.BrainConfig{
			Providers: map[string]cognitive.ProviderConfig{"local": {Type: "openai_compatible", ModelID: "m", Enabled: true, MaxOutputTokens: 2048}},
			Profiles:  map[string]string{"chat": "local"},
		},
		Adapters: map[string]cognitive.LLMProvider{"local": provider},
		Budgets:  cognitive.NewBudgetGovernor(spec, nil),
	}
	manifest := protocol.AgentManifest{ID: "verifier", Role: "tester", Verification: &protocol.Verification{
		Strategy: protocol.VerifyEmpirical, ValidationCommand: "exit 1",
	}}
	envelope, err := NewRunner(brain).Run(context.Background(), manifest, "write it")
	if !errors.Is(err, cognitive.ErrTokenBudgetExhausted) {
		t.Fatalf("err = %v, want token budget exhausted", err)
	}
	if envelope != nil {
		t.Fatalf("budget stop returned an envelope: %+v", envelope)
	}
	if provider.calls != 1 {
		t.Fatalf("provider calls = %d, want 1 (retry refused below the floor)", provider.calls)
	}
	if provider.corr.AgentID != "verifier" {
		t.Fatalf("correlation = %+v, want agent verifier", provider.corr)
	}
}

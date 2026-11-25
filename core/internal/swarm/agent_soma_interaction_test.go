package swarm

import (
	"context"
	"strings"
	"testing"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/pkg/protocol"
)

type sequencedSomaProvider struct {
	replies []string
	calls   int
}

func (p *sequencedSomaProvider) Infer(context.Context, string, cognitive.InferOptions) (*cognitive.InferResponse, error) {
	idx := p.calls
	p.calls++
	text := ""
	if idx < len(p.replies) {
		text = p.replies[idx]
	} else if len(p.replies) > 0 {
		text = p.replies[len(p.replies)-1]
	}
	return &cognitive.InferResponse{
		Text:      text,
		ModelUsed: "test-model",
		Provider:  "mock",
	}, nil
}

func (p *sequencedSomaProvider) Probe(context.Context) (bool, error) {
	return true, nil
}

func newSequencedRouter(p *sequencedSomaProvider) *cognitive.Router {
	return &cognitive.Router{
		Config: &cognitive.BrainConfig{
			Profiles: map[string]string{"chat": "mock"},
			Providers: map[string]cognitive.ProviderConfig{
				"mock": {Type: "mock", Enabled: true, ModelID: "test-model"},
			},
		},
		Adapters: map[string]cognitive.LLMProvider{
			"mock": p,
		},
	}
}

func TestExtractDirectAnswerPayload_UnwrapsHeadersMovesReferentialReviewAndDropsRefusalPairs(t *testing.T) {
	prior := []cognitive.ChatMessage{
		{Role: "user", Content: "[WORKSPACE CONTEXT]\nOrganization: Acme."},
		{Role: "user", Content: "[REFERENTIAL REVIEW]\nPrior deliverable: snake-game/index.html."},
		{Role: "user", Content: "[USER GOVERNANCE PROFILE]\nRole Title: Admin\n\nOriginal request:\nHow do teams communicate?"},
		{Role: "assistant", Content: "Teams communicate over governed NATS subjects."},
		{Role: "user", Content: "[DIRECT ANSWER ROUTE]\n\nOriginal request:\nBuild something broken"},
		{Role: "assistant", Content: "I'm sorry, I can't deliver that request right now."},
	}
	input := "[DIRECT ANSWER ROUTE]\nDo not call tools.\n\nOriginal request:\nWhat is the current platform status?"

	cleanInput, workspaceCtx, cleanHistory, isDirect := extractDirectAnswerPayload(input, prior)
	if !isDirect {
		t.Fatal("expected isDirect = true")
	}
	if cleanInput != "What is the current platform status?" {
		t.Fatalf("cleanInput = %q", cleanInput)
	}
	if !strings.Contains(workspaceCtx, "[WORKSPACE CONTEXT]") || !strings.Contains(workspaceCtx, "[REFERENTIAL REVIEW]") {
		t.Fatalf("workspaceCtx missing blocks: %q", workspaceCtx)
	}
	if len(cleanHistory) != 2 {
		t.Fatalf("len(cleanHistory) = %d, want 2: %+v", len(cleanHistory), cleanHistory)
	}
	if cleanHistory[0].Role != "user" || cleanHistory[0].Content != "How do teams communicate?" {
		t.Fatalf("cleanHistory[0] = %+v", cleanHistory[0])
	}
	if cleanHistory[1].Role != "assistant" || cleanHistory[1].Content != "Teams communicate over governed NATS subjects." {
		t.Fatalf("cleanHistory[1] = %+v", cleanHistory[1])
	}
}

func TestRunToolLoop_DirectAnswerRouteSkipsUnexecutedActionPolicyCorrection(t *testing.T) {
	prov := &sequencedSomaProvider{
		replies: []string{"Should never be called on direct route"},
	}
	agent := NewAgent(context.Background(), protocol.AgentManifest{
		ID:    "admin",
		Role:  "admin",
		Tools: []string{"delegate_task", "write_file"},
	}, "admin-core", nil, newSequencedRouter(prov), &mcpExchangeExecutor{})

	input := "[DIRECT ANSWER ROUTE]\n\nOriginal request:\nHow can I organize my teams?"
	req := &cognitive.InferRequest{
		Profile: "chat",
		Messages: []cognitive.ChatMessage{
			{Role: "user", Content: "How can I organize my teams?"},
		},
	}
	initialResp := &cognitive.InferResponse{
		Text:      "Step 1: you can use delegate_task to assign work across teams. Would you like me to do that?",
		ModelUsed: "test-model",
		Provider:  "mock",
	}

	res := agent.runToolLoop(input, nil, req, initialResp, "chat")
	if prov.calls != 0 {
		t.Fatalf("provider calls = %d, want 0 (direct answer route must not trigger policy correction)", prov.calls)
	}
	if res.responseText != initialResp.Text {
		t.Fatalf("responseText = %q, want %q", res.responseText, initialResp.Text)
	}
}

func TestProcessMessageStructured_RecoversFromInitialCantDeliverRefusal(t *testing.T) {
	prov := &sequencedSomaProvider{
		replies: []string{
			"I'm sorry, I can't deliver that right now.",
			"Soma is ready. Active teams and governed workspaces are healthy.",
		},
	}
	agent := NewAgent(context.Background(), protocol.AgentManifest{
		ID:    "admin",
		Role:  "admin",
		Tools: []string{"read_file"},
	}, "admin-core", nil, newSequencedRouter(prov), nil)

	res := agent.processMessageStructured(
		"[DIRECT ANSWER ROUTE]\n\nOriginal request:\nGive me a quick overview of the workspace.",
		[]cognitive.ChatMessage{
			{Role: "user", Content: "[WORKSPACE CONTEXT]\nOrganization: Mycelis."},
		},
	)
	if prov.calls != 2 {
		t.Fatalf("provider calls = %d, want 2 (initial refusal + clean recovery)", prov.calls)
	}
	if strings.Contains(strings.ToLower(res.Text), "can't deliver") {
		t.Fatalf("expected recovery text, got refusal: %q", res.Text)
	}
	if res.Text != "Soma is ready. Active teams and governed workspaces are healthy." {
		t.Fatalf("res.Text = %q", res.Text)
	}
}

type failingSomaProvider struct{}

func (failingSomaProvider) Infer(context.Context, string, cognitive.InferOptions) (*cognitive.InferResponse, error) {
	return nil, context.DeadlineExceeded
}

func (failingSomaProvider) Probe(context.Context) (bool, error) {
	return false, context.DeadlineExceeded
}

func TestProcessMessageStructured_BrainFreezeSetsProviderOfflineAvailability(t *testing.T) {
	router := &cognitive.Router{
		Config: &cognitive.BrainConfig{
			Profiles: map[string]string{"chat": "ollama"},
			Providers: map[string]cognitive.ProviderConfig{
				"ollama": {Type: "openai_compatible", Enabled: true, ModelID: "qwen2.5-coder:7b-instruct"},
			},
		},
		Adapters: map[string]cognitive.LLMProvider{
			"ollama": failingSomaProvider{},
		},
	}
	agent := NewAgent(context.Background(), protocol.AgentManifest{
		ID:   "admin",
		Role: "admin",
	}, "admin-core", nil, router, nil)

	res := agent.processMessageStructured("[DIRECT ANSWER ROUTE]\n\nOriginal request:\nHello", nil)
	if res.Availability == nil {
		t.Fatal("expected availability struct on brain freeze")
	}
	if res.Availability.Available {
		t.Fatal("expected res.Availability.Available = false on brain freeze")
	}
	if res.Availability.Code != cognitive.ExecutionProviderOffline {
		t.Fatalf("code = %q, want %q", res.Availability.Code, cognitive.ExecutionProviderOffline)
	}
	if !res.Availability.SetupRequired {
		t.Fatal("expected SetupRequired = true on brain freeze")
	}
}

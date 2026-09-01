package swarm

import (
	"bytes"
	"context"
	"errors"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/mycelis/core/internal/cognitive"
)

type boundedInferenceProvider struct {
	messages []cognitive.ChatMessage
	wait     bool
	response string
	calls    int
}

func (provider *boundedInferenceProvider) Infer(ctx context.Context, _ string, options cognitive.InferOptions) (*cognitive.InferResponse, error) {
	provider.calls++
	provider.messages = append([]cognitive.ChatMessage(nil), options.Messages...)
	if provider.wait {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	return &cognitive.InferResponse{Text: provider.response, Provider: "mock", ModelUsed: "bounded"}, nil
}

func (provider *boundedInferenceProvider) Probe(context.Context) (bool, error) { return true, nil }

func TestInferWithExecutionBoundsOrdersSystemMessagesFirst(t *testing.T) {
	provider := &boundedInferenceProvider{response: "done"}
	agent := resultContractTestAgent(provider, &resultContractToolExecutor{})
	req := cognitive.InferRequest{Profile: "chat", Messages: []cognitive.ChatMessage{
		{Role: "system", Content: "base"}, {Role: "user", Content: "request"}, {Role: "system", Content: "contract"},
	}}
	if _, err := agent.inferWithExecutionBounds(req, "test", 1); err != nil {
		t.Fatal(err)
	}
	seenNonSystem := false
	for _, message := range provider.messages {
		if message.Role != "system" {
			seenNonSystem = true
		} else if seenNonSystem {
			t.Fatalf("system message followed user content: %#v", provider.messages)
		}
	}
}

func TestBuildInferRequestCarriesAuthoritativeAgentCorrelation(t *testing.T) {
	agent := resultContractTestAgent(&boundedInferenceProvider{response: "done"}, &resultContractToolExecutor{})
	agent.runID = "run-123"

	req, _ := agent.buildInferRequest("complete the work", nil)
	if req.Correlation.RunID != "run-123" || req.Correlation.TeamID != "delivery-team" || req.Correlation.AgentID != "worker" {
		t.Fatalf("correlation = %#v, want authoritative agent scope", req.Correlation)
	}
}

func TestInferWithExecutionBoundsAppliesDeadline(t *testing.T) {
	previous := agentInferenceTimeout
	agentInferenceTimeout = 10 * time.Millisecond
	t.Cleanup(func() { agentInferenceTimeout = previous })
	provider := &boundedInferenceProvider{wait: true}
	agent := resultContractTestAgent(provider, &resultContractToolExecutor{})
	_, err := agent.inferWithExecutionBounds(cognitive.InferRequest{Profile: "chat"}, "test_timeout", 1)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("error = %v, want deadline exceeded", err)
	}
}

func TestInferenceLogDoesNotExposeMessageContent(t *testing.T) {
	provider := &boundedInferenceProvider{response: "done"}
	agent := resultContractTestAgent(provider, &resultContractToolExecutor{})
	var output bytes.Buffer
	previous := log.Writer()
	log.SetOutput(&output)
	t.Cleanup(func() { log.SetOutput(previous) })
	secret := "never-log-this-secret"
	_, err := agent.inferWithExecutionBounds(cognitive.InferRequest{Profile: "chat", Messages: []cognitive.ChatMessage{{Role: "user", Content: secret}}}, "sanitized", 1)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(output.String(), secret) || !strings.Contains(output.String(), "messages=1") || !strings.Contains(output.String(), "chars=21") {
		t.Fatalf("unsafe or incomplete inference log: %q", output.String())
	}
}

func TestToolLoopAllowsOneUnsafeCorrectionPerToolTarget(t *testing.T) {
	malformed := `{"tool_call":{"name":"write_file","arguments":{"path":"groups/delivery-team/generated/note.txt","content":"broken`
	provider := &resultContractProvider{responses: []string{malformed, malformed, malformed}}
	executor := &resultContractToolExecutor{}
	agent := resultContractTestAgent(provider, executor)
	result := agent.processMessageStructuredWithPosture("Write the note file.", nil, false)
	if provider.calls != 2 {
		t.Fatalf("provider calls = %d, want initial plus one correction", provider.calls)
	}
	if len(executor.calls) != 0 {
		t.Fatalf("unsafe executor calls = %v", executor.calls)
	}
	if result.Availability == nil || result.Availability.Code != "empty_provider_output" {
		t.Fatalf("availability = %#v", result.Availability)
	}
}

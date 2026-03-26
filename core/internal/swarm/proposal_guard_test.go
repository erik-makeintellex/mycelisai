package swarm

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
)

type proposalPlanningProvider struct {
	resp *cognitive.InferResponse
}

func (p *proposalPlanningProvider) Infer(context.Context, string, cognitive.InferOptions) (*cognitive.InferResponse, error) {
	return p.resp, nil
}

func (p *proposalPlanningProvider) Probe(context.Context) (bool, error) {
	return true, nil
}

type countingToolExecutor struct {
	findCalls int
	callCalls int
	serverID  uuid.UUID
}

func (c *countingToolExecutor) FindToolByName(_ context.Context, name string) (uuid.UUID, string, error) {
	c.findCalls++
	if c.serverID == uuid.Nil {
		c.serverID = InternalServerID
	}
	return c.serverID, name, nil
}

func (c *countingToolExecutor) CallTool(_ context.Context, _ uuid.UUID, toolName string, _ map[string]any) (string, error) {
	c.callCalls++
	return "executed:" + toolName, nil
}

func TestBlocksProposalPlanningToolClassification(t *testing.T) {
	tests := map[string]bool{
		"write_file":            true,
		"publish_signal":        true,
		"send_external_message": true,
		"read_file":             false,
		"consult_council":       false,
		"generate_blueprint":    false,
	}

	for toolName, want := range tests {
		if got := blocksProposalPlanningTool(toolName); got != want {
			t.Fatalf("blocksProposalPlanningTool(%q) = %v, want %v", toolName, got, want)
		}
	}
}

func TestProcessMessageStructured_SkipsMutationExecutionDuringProposalPlanning(t *testing.T) {
	workspaceRoot := t.TempDir()
	t.Setenv("MYCELIS_WORKSPACE", workspaceRoot)

	router := &cognitive.Router{
		Config: &cognitive.BrainConfig{
			Profiles: map[string]string{"chat": "mock"},
			Providers: map[string]cognitive.ProviderConfig{
				"mock": {Type: "mock", Enabled: true, ModelID: "test-model"},
			},
		},
		Adapters: map[string]cognitive.LLMProvider{
			"mock": &proposalPlanningProvider{
				resp: &cognitive.InferResponse{
					Text:      `{"tool_call":{"name":"write_file","arguments":{"path":"workspace/logs/qa_browser_preconfirm_probe.py","content":"print('hello world')"}}}`,
					Provider:  "mock",
					ModelUsed: "test-model",
				},
			},
		},
	}

	exec := &countingToolExecutor{serverID: InternalServerID}
	agent := NewAgent(context.Background(), protocol.AgentManifest{
		ID:       "admin",
		Role:     "admin",
		Provider: "mock",
		Tools:    []string{"write_file"},
	}, "admin-core", nil, router, exec)
	agent.SetToolDescriptions(map[string]string{
		"write_file": "Write content to a file within the workspace sandbox.",
	})

	result := agent.processMessageStructured("Create a simple python file named workspace/logs/qa_browser_preconfirm_probe.py that prints hello world.", nil)

	if exec.findCalls != 0 {
		t.Fatalf("FindToolByName called %d times, want 0", exec.findCalls)
	}
	if exec.callCalls != 0 {
		t.Fatalf("CallTool called %d times, want 0", exec.callCalls)
	}
	if len(result.ToolsUsed) != 1 || result.ToolsUsed[0] != "write_file" {
		t.Fatalf("tools_used = %#v, want [write_file]", result.ToolsUsed)
	}

	targetPath := filepath.Join(workspaceRoot, "workspace", "logs", "qa_browser_preconfirm_probe.py")
	if _, err := os.Stat(targetPath); !os.IsNotExist(err) {
		t.Fatalf("expected %s to remain absent before confirmation, got err=%v", targetPath, err)
	}
}

func TestCompositeToolExecutor_BlocksProposalPlanningWriteFile(t *testing.T) {
	workspaceRoot := t.TempDir()
	t.Setenv("MYCELIS_WORKSPACE", workspaceRoot)

	reg := NewInternalToolRegistry(InternalToolDeps{})
	composite := NewCompositeToolExecutor(reg, nil)

	ctx := WithToolInvocationContext(context.Background(), ToolInvocationContext{
		PlanningOnly: true,
	})

	_, err := composite.CallTool(ctx, InternalServerID, "write_file", map[string]any{
		"path":    "workspace/logs/planning-only.txt",
		"content": "hello",
	})
	if err == nil || !strings.Contains(err.Error(), "proposal planning") {
		t.Fatalf("expected proposal-planning block, got %v", err)
	}

	targetPath := filepath.Join(workspaceRoot, "workspace", "logs", "planning-only.txt")
	if _, statErr := os.Stat(targetPath); !os.IsNotExist(statErr) {
		t.Fatalf("expected %s to remain absent, got err=%v", targetPath, statErr)
	}
}

func TestCompositeToolExecutor_BlocksProposalPlanningSignalPublish(t *testing.T) {
	srv, nc := startTestNATS(t)
	defer srv.Shutdown()
	defer nc.Close()

	reg := NewInternalToolRegistry(InternalToolDeps{NC: nc})
	composite := NewCompositeToolExecutor(reg, nil)

	subject := "swarm.team.alpha.signal.status"
	published := make(chan struct{}, 1)
	if _, err := nc.Subscribe(subject, func(msg *nats.Msg) {
		published <- struct{}{}
	}); err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	if err := nc.Flush(); err != nil {
		t.Fatalf("flush: %v", err)
	}

	ctx := WithToolInvocationContext(context.Background(), ToolInvocationContext{
		PlanningOnly: true,
	})

	_, err := composite.CallTool(ctx, InternalServerID, "publish_signal", map[string]any{
		"subject": subject,
		"message": `{"phase":"sync","state":"running"}`,
	})
	if err == nil || !strings.Contains(err.Error(), "proposal planning") {
		t.Fatalf("expected proposal-planning block, got %v", err)
	}

	select {
	case <-published:
		t.Fatal("unexpected publish_signal side effect before confirmation")
	case <-time.After(250 * time.Millisecond):
	}
}

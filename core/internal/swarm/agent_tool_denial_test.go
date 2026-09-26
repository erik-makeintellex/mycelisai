package swarm

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/internal/mcp"
	"github.com/mycelis/core/pkg/protocol"
	"gopkg.in/yaml.v3"
)

func agentManifestWithTools(id string, tools []string) protocol.AgentManifest {
	return protocol.AgentManifest{ID: id, Role: "worker", Provider: "mock", Tools: tools}
}

type scriptedDenialProvider struct {
	mu      sync.Mutex
	replies []string
	prompts []string
}

func (p *scriptedDenialProvider) Infer(_ context.Context, prompt string, opts cognitive.InferOptions) (*cognitive.InferResponse, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, message := range opts.Messages {
		prompt += "\n" + message.Content
	}
	p.prompts = append(p.prompts, prompt)
	text := "Done without that tool."
	if len(p.replies) > 0 {
		text, p.replies = p.replies[0], p.replies[1:]
	}
	return &cognitive.InferResponse{Text: text, Provider: "mock", ModelUsed: "test-model"}, nil
}

func (p *scriptedDenialProvider) Probe(context.Context) (bool, error) { return true, nil }

type recordingEmitter struct {
	mu     sync.Mutex
	events []protocol.EventType
	loads  []map[string]interface{}
}

func (r *recordingEmitter) Emit(_ context.Context, _ string, eventType protocol.EventType, _ protocol.EventSeverity, _, _ string, payload map[string]interface{}) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, eventType)
	r.loads = append(r.loads, payload)
	return "evt", nil
}

func (r *recordingEmitter) denied(t *testing.T) map[string]interface{} {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		r.mu.Lock()
		for i, eventType := range r.events {
			if eventType == protocol.EventToolDenied {
				payload := r.loads[i]
				r.mu.Unlock()
				return payload
			}
		}
		r.mu.Unlock()
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("no tool.denied event emitted")
	return nil
}

func denialTestAgent(replies ...string) (*Agent, *countingToolExecutor, *scriptedDenialProvider, *recordingEmitter) {
	provider := &scriptedDenialProvider{replies: replies}
	router := &cognitive.Router{
		Config: &cognitive.BrainConfig{
			Profiles:  map[string]string{"chat": "mock"},
			Providers: map[string]cognitive.ProviderConfig{"mock": {Type: "mock", Enabled: true, ModelID: "test-model"}},
		},
		Adapters: map[string]cognitive.LLMProvider{"mock": provider},
	}
	exec := &countingToolExecutor{serverID: InternalServerID}
	agent := NewAgent(context.Background(), agentManifestWithTools("worker", []string{"write_file"}), "team-a", nil, router, exec)
	agent.SetToolDescriptions(map[string]string{"write_file": "Write a workspace file."})
	emitter := &recordingEmitter{}
	agent.SetEventEmitter(emitter, "run-1")
	return agent, exec, provider, emitter
}

func TestAgentUndeclaredToolDeniedWithEventAndFeedback(t *testing.T) {
	agent, exec, provider, emitter := denialTestAgent(`{"tool_call":{"name":"broadcast","arguments":{"message":"all teams stop"}}}`)

	result := agent.processMessageStructuredWithPosture("Tell every team to stop.", nil, false)

	if exec.findCalls != 0 || exec.callCalls != 0 {
		t.Fatalf("undeclared tool reached the executor: find=%d call=%d", exec.findCalls, exec.callCalls)
	}
	for _, used := range result.ToolsUsed {
		if used == "broadcast" {
			t.Fatalf("denied tool reported as used: %v", result.ToolsUsed)
		}
	}
	payload := emitter.denied(t)
	if payload["tool"] != "broadcast" || payload["phase"] != "lookup" || payload["reason"] != "not_declared" {
		t.Fatalf("tool.denied payload = %#v", payload)
	}
	provider.mu.Lock()
	defer provider.mu.Unlock()
	if len(provider.prompts) < 2 || !strings.Contains(provider.prompts[len(provider.prompts)-1], "is not permitted for this agent and was not executed") {
		t.Fatal("model did not receive the normalized denial feedback")
	}
}

func TestAgentPlanningCaptureRejectsUndeclaredTool(t *testing.T) {
	agent, exec, _, emitter := denialTestAgent(`{"tool_call":{"name":"local_command","arguments":{"command":"rm -rf /tmp/x"}}}`)

	result := agent.processMessageStructuredWithPosture("Clean the temp folder.", nil, true)

	if len(result.PlannedToolCalls) != 0 {
		t.Fatalf("undeclared tool captured as planned call: %#v", result.PlannedToolCalls)
	}
	if exec.findCalls != 0 || exec.callCalls != 0 {
		t.Fatalf("planning denial reached the executor: find=%d call=%d", exec.findCalls, exec.callCalls)
	}
	if payload := emitter.denied(t); payload["phase"] != "planning" {
		t.Fatalf("tool.denied payload = %#v", payload)
	}
}

// shippedAgentTools walks a team or template YAML and returns agent id -> tools.
func shippedAgentTools(t *testing.T, path string) map[string][]string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var doc any
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	agents := map[string][]string{}
	var walk func(node any)
	walk = func(node any) {
		switch typed := node.(type) {
		case map[string]any:
			if id, ok := typed["id"].(string); ok {
				if tools, ok := typed["tools"].([]any); ok {
					for _, tool := range tools {
						agents[id] = append(agents[id], tool.(string))
					}
				}
			}
			for _, child := range typed {
				walk(child)
			}
		case []any:
			for _, child := range typed {
				walk(child)
			}
		}
	}
	walk(doc)
	return agents
}

func TestShippedAgentDeclaredToolsResolve(t *testing.T) {
	files, _ := filepath.Glob(filepath.Join("..", "..", "config", "teams", "*.yaml"))
	templates, _ := filepath.Glob(filepath.Join("..", "..", "config", "templates", "*.yaml"))
	files = append(files, templates...)
	if len(files) < 8 {
		t.Fatalf("found %d shipped team/template files, want >= 8", len(files))
	}
	library, err := mcp.LoadLibrary(filepath.Join("..", "..", "config", "mcp-library.yaml"))
	if err != nil {
		t.Fatalf("load MCP library: %v", err)
	}
	registry := NewInternalToolRegistry(InternalToolDeps{})
	composite := NewCompositeToolExecutor(registry, nil)
	seen := map[string]bool{}
	for _, file := range files {
		for id, tools := range shippedAgentTools(t, file) {
			seen[id] = true
			scoped := NewScopedToolExecutor(composite, tools, nil)
			for _, tool := range tools {
				switch entry := strings.TrimSpace(tool); {
				case entry != tool:
					t.Errorf("%s %s: tool %q is not canonical (whitespace)", file, id, tool)
				case mcp.IsToolSetRef(entry):
					t.Errorf("%s %s: shipped agents must not depend on DB toolset %q", file, id, entry)
				case mcp.IsMCPRef(entry):
					if ref := mcp.ParseToolRef(entry); ref == nil || library.FindByName(ref.ServerName) == nil {
						t.Errorf("%s %s: MCP ref %q names no library server", file, id, entry)
					}
				default:
					if serverID, _, err := scoped.FindToolByName(context.Background(), entry); err != nil || serverID != InternalServerID {
						t.Errorf("%s %s: declared tool %q does not resolve to a registered internal tool: %v", file, id, entry, err)
					}
				}
			}
		}
	}
	for _, id := range []string{"admin", "council-architect", "council-coder", "council-creative", "council-sentry"} {
		if !seen[id] {
			t.Errorf("system agent %s has no explicit declared tool list", id)
		}
	}
	names := make([]string, 0, len(seen))
	for id := range seen {
		names = append(names, id)
	}
	sort.Strings(names)
	t.Logf("resolved declared tools for %d shipped agents: %v", len(names), names)
}

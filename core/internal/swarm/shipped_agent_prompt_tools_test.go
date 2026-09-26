package swarm

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

type shippedAgent struct {
	file, teamID, id, role, prompt string
	tools                          []string
}

// shippedAgents walks team and template YAML, recording each agent with its
// owning team id (the nearest ancestor that has members).
func shippedAgents(t *testing.T) []shippedAgent {
	t.Helper()
	files, _ := filepath.Glob(filepath.Join("..", "..", "config", "teams", "*.yaml"))
	templates, _ := filepath.Glob(filepath.Join("..", "..", "config", "templates", "*.yaml"))
	var agents []shippedAgent
	for _, file := range append(files, templates...) {
		raw, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		var doc any
		if err := yaml.Unmarshal(raw, &doc); err != nil {
			t.Fatalf("parse %s: %v", file, err)
		}
		var walk func(node any, teamID string)
		walk = func(node any, teamID string) {
			switch typed := node.(type) {
			case map[string]any:
				id, _ := typed["id"].(string)
				if members, ok := typed["members"].([]any); ok {
					for _, member := range members {
						if agent, ok := member.(map[string]any); ok {
							a := shippedAgent{file: filepath.Base(file), teamID: id}
							a.id, _ = agent["id"].(string)
							a.role, _ = agent["role"].(string)
							a.prompt, _ = agent["system_prompt"].(string)
							if tools, ok := agent["tools"].([]any); ok {
								for _, tool := range tools {
									a.tools = append(a.tools, tool.(string))
								}
							}
							agents = append(agents, a)
						}
					}
					return
				}
				for _, child := range typed {
					walk(child, teamID)
				}
			case []any:
				for _, child := range typed {
					walk(child, teamID)
				}
			}
		}
		walk(doc, "")
	}
	if len(agents) < 10 {
		t.Fatalf("found %d shipped agents, want >= 10", len(agents))
	}
	return agents
}

// TestShippedAgentPromptToolsAreDeclared prevents silent denials. Every
// registered tool an agent's YAML prompt cites must be declared on it (the
// runtime-owned base toolset is not model-callable, so it does not count).
// The shared runtime lead protocol is checked as the agent sees it (after
// withoutUndeclaredToolLines), and Soma's admin must see all of it.
func TestShippedAgentPromptToolsAreDeclared(t *testing.T) {
	registry := NewInternalToolRegistry(InternalToolDeps{})
	checked := 0
	for _, agent := range shippedAgents(t) {
		if len(agent.tools) == 0 {
			continue // no tool loop runs, so prompt mentions cannot become denied calls
		}
		checked++
		declared := map[string]bool{}
		for _, tool := range agent.tools {
			declared[tool] = true
		}
		runtimeContext := registry.BuildContext(agent.id, agent.teamID, agent.role, nil, nil, "")
		sources := map[string]string{
			"yaml prompt":     agent.prompt,
			"runtime context": registry.withoutUndeclaredToolLines(runtimeContext, agent.tools),
		}
		if agent.id == "admin" {
			sources["unfiltered runtime context"] = runtimeContext
		}
		for source, text := range sources {
			missing := map[string]bool{}
			for _, name := range registry.promptToolNames(text) {
				if !declared[name] {
					missing[name] = true
				}
			}
			names := make([]string, 0, len(missing))
			for name := range missing {
				names = append(names, name)
			}
			sort.Strings(names)
			if len(names) > 0 {
				t.Errorf("%s %s/%s %s cites undeclared tools: %s", agent.file, agent.teamID, agent.id, source, strings.Join(names, ", "))
			}
		}
	}
	if checked < 8 {
		t.Fatalf("checked %d tool-declaring agents, want >= 8", checked)
	}
}

func TestWithoutUndeclaredToolLinesKeepsDeclaredGuidance(t *testing.T) {
	registry := NewInternalToolRegistry(InternalToolDeps{})
	text := "intro\n1. use `recall`\n2. use `local_command` or `recall`\n3. mention `customer_context`\n"
	got := registry.withoutUndeclaredToolLines(text, []string{"recall"})
	if got != "intro\n1. use `recall`\n3. mention `customer_context`\n" {
		t.Fatalf("filtered context = %q", got)
	}
}

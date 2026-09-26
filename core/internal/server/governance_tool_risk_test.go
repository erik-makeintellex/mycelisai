package server

import (
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/mycelis/core/internal/swarm"
	"github.com/mycelis/core/pkg/protocol"
	"gopkg.in/yaml.v3"
)

// The explicit risk table must cover exactly the runtime internal tool registry.
func TestRegisteredToolRiskMatchesRuntimeRegistry(t *testing.T) {
	names := swarm.NewInternalToolRegistry(swarm.InternalToolDeps{}).ListNames()
	slices.Sort(names)
	table := make([]string, 0, len(registeredInternalToolRisk))
	for name, risk := range registeredInternalToolRisk {
		table = append(table, name)
		if risk != "low" && risk != "medium" && risk != "high" {
			t.Errorf("%s: risk %q must be explicit low/medium/high", name, risk)
		}
	}
	slices.Sort(table)
	if !slices.Equal(names, table) {
		t.Fatalf("risk table drifted from the internal tool registry:\nregistry %v\ntable    %v", names, table)
	}
}

// A2b C1b: anything that is not an exact registered internal name is high.
func TestToolRiskFailsClosed(t *testing.T) {
	for _, tool := range []string{"unknown_tool", "Broadcast", "READ_FILE", "toolset:ops", "local_command",
		"send_external_message", "slack", "mcp:filesystem/write_file", "delegate", "", "read_file\x00"} {
		if got := capabilityRiskForTool(tool, nil); got != "high" {
			t.Errorf("%q: risk %q, want high", tool, got)
		}
		if tier, _ := approverTier(buildScopeFromBlueprint(tierBlueprint(1, tool))); tier != approverTierApprover {
			t.Errorf("%q: blueprint tier %d, want 2", tool, tier)
		}
	}
	for tool, want := range map[string]string{"read_file": "low", " read_file ": "low", "search_memory": "low", "write_file": "medium"} {
		if got := capabilityRiskForTool(tool, nil); got != want {
			t.Errorf("%q: risk %q, want %q", tool, got, want)
		}
	}
	if tier, _ := approverTier(buildScopeFromBlueprint(tierBlueprint(2, "read_file", "search_memory"))); tier != approverTierAuto {
		t.Fatalf("registered low tools must stay tier 0, got %d", tier)
	}
}

// Fail-open examples are refused at commit for a standard user.
func TestIntentCommitUnregisteredToolNeedsApprover(t *testing.T) {
	for _, tool := range []string{"Broadcast", "toolset:ops", "local_command", "send_external_message", "slack", "unknown_tool"} {
		t.Run(tool, func(t *testing.T) {
			dbOpt, mock := withDB(t)
			s := newTestServer(dbOpt)
			bp := tierBlueprint(1, tool)
			expectCommitToken(t, mock, buildScopeFromBlueprint(bp), bp, "u-std")
			rr := commitAs(t, s, bp, standardUserIdentity())
			assertStatus(t, rr, http.StatusForbidden)
			if !strings.Contains(rr.Body.String(), "approver_required") {
				t.Fatalf("expected approver_required, got %s", rr.Body.String())
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// Chat/council (buildApprovalPolicy) and blueprint classification agree.
func TestChatAndBlueprintToolTiersAgree(t *testing.T) {
	profile := defaultUserGovernanceProfile("")
	for _, tools := range [][]string{{"read_file"}, {"write_file"}, {"broadcast"}, {"mcp:github/create_issue"},
		{"Broadcast"}, {"toolset:ops"}, {"read_file", "local_command"}, {"search_memory", "recall"}} {
		planned := []protocol.PlannedToolCall{}
		for _, tool := range tools {
			planned = append(planned, protocol.PlannedToolCall{Name: tool})
		}
		chat, _ := approverTier(&protocol.ScopeValidation{Approval: applyApproverTier(buildApprovalPolicy(profile, planned, nil))})
		blueprint, _ := approverTier(buildScopeFromBlueprintFor(tierBlueprint(1, tools...), profile))
		if chat != blueprint {
			t.Errorf("%v: chat tier %d, blueprint tier %d", tools, chat, blueprint)
		}
	}
}

// Every tool named in shipped team and template config is a registered
// internal name or an `mcp:<server>/<tool>` ref, and classifies as expected.
func TestShippedConfigToolTiers(t *testing.T) {
	files, _ := filepath.Glob("../../config/teams/*.yaml")
	templates, _ := filepath.Glob("../../config/templates/*.yaml")
	files = append(files, templates...)
	if len(files) == 0 {
		t.Fatal("no shipped team/template config found")
	}
	seen := 0
	for _, path := range files {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var doc any
		if err := yaml.Unmarshal(raw, &doc); err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		for _, tool := range collectToolNames(doc) {
			seen++
			_, registered := registeredInternalToolRisk[tool]
			isMCP := strings.HasPrefix(tool, "mcp:") && strings.Contains(tool, "/")
			switch {
			case registered:
				if got := capabilityRiskForTool(tool, nil); got != registeredInternalToolRisk[tool] {
					t.Errorf("%s: %s risk %q, want %q", path, tool, got, registeredInternalToolRisk[tool])
				}
			case isMCP:
				if tier, _ := approverTier(buildScopeFromBlueprint(tierBlueprint(1, tool))); tier != approverTierApprover {
					t.Errorf("%s: %s must be tier 2", path, tool)
				}
			default:
				t.Errorf("%s: tool %q is neither a registered internal tool nor mcp:<server>/<tool>", path, tool)
			}
		}
	}
	if seen == 0 {
		t.Fatal("expected shipped tool lists")
	}
}

func collectToolNames(node any) []string {
	var out []string
	switch v := node.(type) {
	case map[string]any:
		for key, child := range v {
			if key == "tools" {
				if list, ok := child.([]any); ok {
					for _, item := range list {
						if name, ok := item.(string); ok {
							out = append(out, name)
						}
					}
					continue
				}
			}
			out = append(out, collectToolNames(child)...)
		}
	case []any:
		for _, child := range v {
			out = append(out, collectToolNames(child)...)
		}
	}
	return out
}

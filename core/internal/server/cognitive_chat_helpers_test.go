package server

import (
	"strings"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

func TestParsePlannedToolCall_NormalizesWriteFileAliases(t *testing.T) {
	call, ok := parsePlannedToolCall(`{"tool_call":{"name":"write_file","arguments":{"file_path":"workspace/logs/alias_test.py","body":"print('hello world')"}}}`)
	if !ok {
		t.Fatal("expected planned tool call to parse")
	}
	if call.Arguments["path"] != "workspace/logs/alias_test.py" {
		t.Fatalf("path = %#v", call.Arguments["path"])
	}
	if call.Arguments["content"] != "print('hello world')" {
		t.Fatalf("content = %#v", call.Arguments["content"])
	}
}

func TestParsePlannedToolCall_PreservesMCPToolRef(t *testing.T) {
	call, ok := parsePlannedToolCall(`{"tool_call":{"tool_ref":"mcp:filesystem/read_text_file","arguments":{"path":"workspace/logs/proof.md"}}}`)
	if !ok {
		t.Fatal("expected planned MCP tool call to parse")
	}
	if call.Name != "read_text_file" {
		t.Fatalf("name = %q, want read_text_file", call.Name)
	}
	if call.ToolRef != "mcp:filesystem/read_text_file" {
		t.Fatalf("tool_ref = %q", call.ToolRef)
	}
	if call.Arguments["path"] != "workspace/logs/proof.md" {
		t.Fatalf("path = %#v", call.Arguments["path"])
	}
	tools := toolsForPlannedCalls([]protocol.PlannedToolCall{call}, nil)
	if len(tools) != 1 || tools[0] != "mcp:filesystem/read_text_file" {
		t.Fatalf("effective tools = %#v, want MCP ref", tools)
	}
}

func TestBuildPlannedToolCalls_PrefersExplicitCreateTeamRequest(t *testing.T) {
	calls := buildPlannedToolCalls(chatAgentResult{
		Text: `{"tool_call":{"name":"generate_blueprint","arguments":{"topic":"wrong"}}}`,
	}, "Create a team with team_id qa-visibility-team named QA Visibility Team. Role researcher.", []string{"generate_blueprint", "delegate"})

	if len(calls) != 1 {
		t.Fatalf("planned calls = %#v, want one create_team call", calls)
	}
	if calls[0].Name != "create_team" {
		t.Fatalf("planned call = %q, want create_team", calls[0].Name)
	}
	if calls[0].Arguments["team_id"] != "qa-visibility-team" {
		t.Fatalf("team_id = %#v", calls[0].Arguments["team_id"])
	}
	if calls[0].Arguments["name"] != "QA Visibility Team" {
		t.Fatalf("name = %#v", calls[0].Arguments["name"])
	}
	if calls[0].Arguments["role"] != "researcher" {
		t.Fatalf("role = %#v", calls[0].Arguments["role"])
	}
	if calls[0].Arguments["staffing_mode"] != "lead_only_start" || calls[0].Arguments["initial_member_count"] != 1 {
		t.Fatalf("staffing args = %#v, want lead-only start", calls[0].Arguments)
	}
	tools := toolsForPlannedCalls(calls, []string{"generate_blueprint", "delegate"})
	if len(tools) != 1 || tools[0] != "create_team" {
		t.Fatalf("effective tools = %#v, want create_team only", tools)
	}
}

func TestBuildPlannedToolCalls_KeepsTeamAndConcreteCodeOutput(t *testing.T) {
	request := "Create a compact team with team_id qa-browser-game-team. Then have that team create a simple browser click game at path workspace/logs/qa_team_game.html containing '<!doctype html><title>QA Click Game</title><button id=coin>Coin</button><p id=score>0</p><script>let s=0;coin.onclick=()=>score.textContent=++s</script>'"
	calls := buildPlannedToolCalls(chatAgentResult{
		Text: `{"tool_call":{"name":"generate_blueprint","arguments":{"topic":"wrong"}}}`,
	}, request, []string{"write_file", "generate_blueprint", "delegate"})

	if len(calls) != 2 {
		t.Fatalf("planned calls = %#v, want create_team + write_file", calls)
	}
	if calls[0].Name != "create_team" {
		t.Fatalf("first call = %q, want create_team", calls[0].Name)
	}
	if calls[0].Arguments["team_id"] != "qa-browser-game-team" {
		t.Fatalf("team_id = %#v", calls[0].Arguments["team_id"])
	}
	if calls[0].Arguments["name"] != "Qa Browser Game Team" {
		t.Fatalf("name = %#v", calls[0].Arguments["name"])
	}
	if calls[0].Arguments["initial_member_count"] != 1 {
		t.Fatalf("initial_member_count = %#v, want 1", calls[0].Arguments["initial_member_count"])
	}
	if calls[1].Name != "write_file" {
		t.Fatalf("second call = %q, want write_file", calls[1].Name)
	}
	if calls[1].Arguments["path"] != "workspace/logs/qa_team_game.html" {
		t.Fatalf("path = %#v", calls[1].Arguments["path"])
	}
	if !strings.Contains(calls[1].Arguments["content"].(string), "coin.onclick") {
		t.Fatalf("content = %#v, want simple game code", calls[1].Arguments["content"])
	}
	tools := toolsForPlannedCalls(calls, []string{"write_file", "generate_blueprint", "delegate"})
	if len(tools) != 2 || tools[0] != "create_team" || tools[1] != "write_file" {
		t.Fatalf("effective tools = %#v, want create_team + write_file", tools)
	}
}

func TestBuildPlannedToolCalls_PreservesExplicitCreateTeamArguments(t *testing.T) {
	request := "Create a compact team with team_id rich-team. Then have that team create a small note at path workspace/logs/rich_team_note.md containing 'ready'"
	calls := buildPlannedToolCalls(chatAgentResult{
		Text: `{"tool_call":{"name":"create_team","arguments":{"team_id":"rich-team","name":"Rich Team","role":"validator","manifest":{"ask_routing":{"validation":"validator"}},"required_capabilities":["write_file"]}}}`,
	}, request, []string{"create_team", "write_file"})

	if len(calls) != 2 {
		t.Fatalf("planned calls = %#v, want create_team + write_file", calls)
	}
	if calls[0].Name != "create_team" {
		t.Fatalf("first call = %q, want create_team", calls[0].Name)
	}
	if calls[0].Arguments["name"] != "Rich Team" || calls[0].Arguments["role"] != "validator" {
		t.Fatalf("create_team arguments = %#v, want rich model-provided arguments", calls[0].Arguments)
	}
	if _, ok := calls[0].Arguments["manifest"].(map[string]any); !ok {
		t.Fatalf("manifest = %#v, want preserved model-provided manifest", calls[0].Arguments["manifest"])
	}
	if calls[1].Name != "write_file" || calls[1].Arguments["path"] != "workspace/logs/rich_team_note.md" {
		t.Fatalf("write_file call = %#v", calls[1])
	}
}

func TestInferCreateTeamPlanFromRequest_DefaultResearchTeam(t *testing.T) {
	call, ok := inferCreateTeamPlanFromRequest("i need an indepth ai research team that can take on current research")
	if !ok {
		t.Fatal("expected create_team plan")
	}
	if call.Arguments["team_id"] != "ai-research-team" {
		t.Fatalf("team_id = %#v", call.Arguments["team_id"])
	}
	if call.Arguments["role"] != "researcher" {
		t.Fatalf("role = %#v", call.Arguments["role"])
	}
	if call.Arguments["initial_member_count"] != 1 || call.Arguments["recommended_member_limit"] != 3 {
		t.Fatalf("minimal staffing args = %#v, want lead-only start with bounded expansion", call.Arguments)
	}
}

func TestResolveChatAskContract(t *testing.T) {
	direct := resolveChatAskContract("soma", false, chatAgentResult{})
	if direct.AskClass != protocol.AskClassDirectAnswer {
		t.Fatalf("direct ask class = %q, want %q", direct.AskClass, protocol.AskClassDirectAnswer)
	}
	if direct.TemplateID != protocol.TemplateChatToAnswer {
		t.Fatalf("direct template = %q, want %q", direct.TemplateID, protocol.TemplateChatToAnswer)
	}
	if direct.DefaultExecutionMode != protocol.ModeAnswer {
		t.Fatalf("direct mode = %q, want %q", direct.DefaultExecutionMode, protocol.ModeAnswer)
	}

	artifact := resolveChatAskContract("soma", false, chatAgentResult{
		Artifacts: []protocol.ChatArtifactRef{{Type: "document", Title: "Brief"}},
	})
	if artifact.AskClass != protocol.AskClassGovernedArtifact {
		t.Fatalf("artifact ask class = %q, want %q", artifact.AskClass, protocol.AskClassGovernedArtifact)
	}
	if artifact.DefaultExecutionMode != protocol.ModeAnswer {
		t.Fatalf("artifact mode = %q, want %q", artifact.DefaultExecutionMode, protocol.ModeAnswer)
	}

	specialist := resolveChatAskContract("specialist", false, chatAgentResult{})
	if specialist.AskClass != protocol.AskClassSpecialist {
		t.Fatalf("specialist ask class = %q, want %q", specialist.AskClass, protocol.AskClassSpecialist)
	}
	if specialist.DefaultAgentTarget != "specialist" {
		t.Fatalf("specialist default target = %q, want specialist", specialist.DefaultAgentTarget)
	}

	consulted := resolveChatAskContract("soma", false, chatAgentResult{
		Consultations: []protocol.ConsultationEntry{{Member: "council-architect", Summary: "Reviewed the plan."}},
	})
	if consulted.AskClass != protocol.AskClassSpecialist {
		t.Fatalf("consulted ask class = %q, want %q", consulted.AskClass, protocol.AskClassSpecialist)
	}

	mutation := resolveChatAskContract("soma", true, chatAgentResult{})
	if mutation.AskClass != protocol.AskClassGovernedMutation {
		t.Fatalf("mutation ask class = %q, want %q", mutation.AskClass, protocol.AskClassGovernedMutation)
	}
	if mutation.TemplateID != protocol.TemplateChatToProposal {
		t.Fatalf("mutation template = %q, want %q", mutation.TemplateID, protocol.TemplateChatToProposal)
	}
	if mutation.DefaultExecutionMode != protocol.ModeProposal {
		t.Fatalf("mutation mode = %q, want %q", mutation.DefaultExecutionMode, protocol.ModeProposal)
	}
	if !mutation.RequiresConfirmation {
		t.Fatal("mutation contract should require confirmation")
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

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

func TestBuildPlannedToolCalls_StartsTeamGameDeliverable(t *testing.T) {
	request := "Create a team named SNES-Style Browser Game Team and get them to work on developing a detailed game"
	calls := buildPlannedToolCalls(chatAgentResult{
		Text: `{"tool_call":{"name":"generate_blueprint","arguments":{"topic":"wrong"}}}`,
	}, request, []string{"generate_blueprint", "delegate"})

	if len(calls) != 2 {
		t.Fatalf("planned calls = %#v, want create_team + write_file", calls)
	}
	if calls[0].Name != "create_team" {
		t.Fatalf("first call = %q, want create_team", calls[0].Name)
	}
	if calls[0].Arguments["name"] != "SNES-Style Browser Game Team" {
		t.Fatalf("name = %#v", calls[0].Arguments["name"])
	}
	if calls[1].Name != "write_file" {
		t.Fatalf("second call = %q, want write_file", calls[1].Name)
	}
	path, _ := calls[1].Arguments["path"].(string)
	if path != "groups/snes-style-browser-game-team/generated/first-game/index.html" {
		t.Fatalf("path = %q, want group-scoped browser-game entrypoint", path)
	}
	content, _ := calls[1].Arguments["content"].(string)
	if !strings.Contains(content, "<canvas id=\"game\"") || !strings.Contains(content, "requestAnimationFrame(loop)") {
		t.Fatalf("content does not look like a playable browser game: %.120q", content)
	}
	if calls[1].Arguments["package_kind"] != "project_package" {
		t.Fatalf("package_kind = %#v, want project_package", calls[1].Arguments["package_kind"])
	}
	tools := toolsForPlannedCalls(calls, []string{"generate_blueprint", "delegate"})
	if len(tools) != 2 || tools[0] != "create_team" || tools[1] != "write_file" {
		t.Fatalf("effective tools = %#v, want create_team + write_file", tools)
	}
}

func TestBuildPlannedToolCalls_GameDeliverableIncludesReadmeWhenRequested(t *testing.T) {
	request := strings.Join([]string{
		"Create a team named First Demo Game Team and get them to build a playable browser game.",
		"The package metadata must include files index.html and README.md plus validation notes from opening the browser game.",
		"After approval, return a retained project_package output with entrypoint, folder, files, and validation.",
	}, " ")
	calls := buildPlannedToolCalls(chatAgentResult{
		Text: `{"tool_call":{"name":"generate_blueprint","arguments":{"topic":"wrong"}}}`,
	}, request, []string{"generate_blueprint", "delegate"})

	if len(calls) != 2 || calls[1].Name != "write_file" {
		t.Fatalf("planned calls = %#v, want create_team + write_file", calls)
	}
	files := confirmedActionStringSlice(calls[1].Arguments["package_files"])
	if !containsString(files, "index.html") || !containsString(files, "README.md") {
		t.Fatalf("package_files = %#v, want index.html and README.md", files)
	}
	if calls[1].Arguments["package_kind"] != "project_package" {
		t.Fatalf("package_kind = %#v, want project_package", calls[1].Arguments["package_kind"])
	}
}

func TestDeterministicGovernedMutationResult_BuildsWriteFileProposalWithoutAgent(t *testing.T) {
	request := "Create a simple python file named workspace/logs/hello.py that prints hello world."
	result, ok := deterministicGovernedMutationResult(request, []string{"write_file"})
	if !ok {
		t.Fatal("expected deterministic governed mutation result")
	}
	if !containsToolName(result.ToolsUsed, "write_file") {
		t.Fatalf("tools_used = %#v, want write_file", result.ToolsUsed)
	}
	if !strings.Contains(result.Text, "workspace/logs/hello.py") {
		t.Fatalf("text = %q, want target path", result.Text)
	}

	calls := buildPlannedToolCalls(result, request, result.ToolsUsed)
	if len(calls) != 1 || calls[0].Name != "write_file" {
		t.Fatalf("planned calls = %#v, want one write_file call", calls)
	}
	if calls[0].Arguments["path"] != "workspace/logs/hello.py" {
		t.Fatalf("path = %#v, want workspace/logs/hello.py", calls[0].Arguments["path"])
	}
}

func TestDeterministicGovernedMutationResult_BuildsFirstDemoTeamGameProposal(t *testing.T) {
	request := strings.Join([]string{
		"Create a team with team_id first-demo-game-team named First Demo Game Team.",
		"Ask Soma for the exact first demo deliverable: a playable browser game project package.",
		"Retain it at workspace/generated/first-demo-game-team-first-game with entrypoint workspace/generated/first-demo-game-team-first-game/index.html.",
		"The package metadata must include files index.html and README.md plus validation notes from opening the browser game.",
	}, " ")
	result, ok := deterministicGovernedMutationResult(request, []string{"write_file", "generate_blueprint", "delegate"})
	if !ok {
		t.Fatal("expected deterministic governed mutation result")
	}
	if !containsToolName(result.ToolsUsed, "create_team") || !containsToolName(result.ToolsUsed, "write_file") {
		t.Fatalf("tools_used = %#v, want create_team and write_file", result.ToolsUsed)
	}

	calls := buildPlannedToolCalls(result, request, result.ToolsUsed)
	if len(calls) != 2 || calls[0].Name != "create_team" || calls[1].Name != "write_file" {
		t.Fatalf("planned calls = %#v, want create_team + write_file", calls)
	}
	if calls[0].Arguments["name"] != "First Demo Game Team" {
		t.Fatalf("team name = %#v, want First Demo Game Team", calls[0].Arguments["name"])
	}
	if calls[1].Arguments["package_title"] != "First Demo Game Team First Playable" {
		t.Fatalf("package_title = %#v, want human team name title", calls[1].Arguments["package_title"])
	}
	files := confirmedActionStringSlice(calls[1].Arguments["package_files"])
	if !containsString(files, "README.md") {
		t.Fatalf("package_files = %#v, want README.md", files)
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

func TestBuildPlannedToolCalls_ComicTeamIncludesSpecialistsAndMediaDeliverable(t *testing.T) {
	request := "Create a team to write a comic book page. We need an artist, someone who comes up with characters, and someone who writes the lines. Generate a comic image using local ComfyUI."
	result, ok := deterministicGovernedMutationResult(request, []string{"create_team", "generate_image", "save_cached_image"})
	if !ok || !strings.Contains(result.Text, "groups/soma-requested-team/media") {
		t.Fatalf("deterministic proposal = %#v, ok=%v, want group media target", result, ok)
	}
	calls := buildPlannedToolCalls(chatAgentResult{}, request, []string{"create_team", "generate_image", "save_cached_image"})
	if len(calls) != 3 {
		t.Fatalf("planned calls = %#v, want create_team + generate_image + save_cached_image", calls)
	}
	if calls[0].Name != "create_team" || calls[1].Name != "generate_image" || calls[2].Name != "save_cached_image" {
		t.Fatalf("planned call names = %#v", calls)
	}
	agents, ok := calls[0].Arguments["agents"].([]map[string]any)
	if !ok || len(agents) < 5 || calls[0].Arguments["staffing_mode"] != "specialist_delivery" {
		t.Fatalf("team args = %#v, agents = %#v", calls[0].Arguments, agents)
	}
	if calls[1].Arguments["size"] != "768x1024" || calls[2].Arguments["folder"] != "groups/soma-requested-team/media" {
		t.Fatalf("media calls = %#v %#v", calls[1], calls[2])
	}
	tools := toolsForPlannedCalls(calls, nil)
	if !containsString(tools, "generate_image") || !containsString(tools, "save_cached_image") {
		t.Fatalf("tools = %#v, want media tools", tools)
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

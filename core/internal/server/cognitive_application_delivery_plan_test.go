package server

import (
	"strings"
	"testing"
)

func TestBuildPlannedToolCalls_ComplexAppAskDelegatesWithPackageContract(t *testing.T) {
	request := strings.Join([]string{
		"Create a team named Operations App Team and get them to build a deployable browser app for commercial data review.",
		"The app should include readable tables, usage notes, validation proof, and a direct launch path.",
	}, " ")

	calls := plannedCallsFromWrongBlueprint(request, []string{"generate_blueprint", "delegate"})
	requirePlannedCallNames(t, calls, "create_team", "write_file", "write_file", "delegate_task")

	teamID, _ := calls[0].Arguments["team_id"].(string)
	if teamID == "" {
		t.Fatalf("team_id = %#v", calls[0].Arguments["team_id"])
	}
	if calls[1].Arguments["path"] != "groups/"+teamID+"/planning/TEAM_EVOCATION.md" {
		t.Fatalf("evocation path = %#v", calls[1].Arguments["path"])
	}
	tools := confirmedActionStringSlice(calls[0].Arguments["tools"])
	for _, want := range []string{"write_file", "store_artifact", "research_for_blueprint", "consult_council", "read_file", "local_command"} {
		if !containsToolName(tools, want) {
			t.Fatalf("team tools = %#v, missing %q", tools, want)
		}
	}
	if calls[2].Arguments["path"] != "groups/"+teamID+"/planning/RESEARCH_COUNCIL_HANDOFF.md" {
		t.Fatalf("handoff path = %#v", calls[2].Arguments["path"])
	}

	ask := mapArgument(calls[3].Arguments["ask"])
	context := mapArgument(ask["context"])
	resultContract := mapArgument(context["result_contract"])
	if resultContract["kind"] != "project_package" || resultContract["repair_channel"] != "soma" {
		t.Fatalf("result contract = %#v", resultContract)
	}
	if resultContract["team_id"] != teamID {
		t.Fatalf("result contract team_id = %#v, want %q", resultContract["team_id"], teamID)
	}
	if resultContract["package_folder"] != "groups/"+teamID+"/generated/package" {
		t.Fatalf("package_folder = %#v", resultContract["package_folder"])
	}
	if resultContract["package_entrypoint"] != "groups/"+teamID+"/generated/package/index.html" {
		t.Fatalf("package_entrypoint = %#v", resultContract["package_entrypoint"])
	}
	if target := firstPlannedOutputTarget(calls); target != "groups/"+teamID+"/generated/package/index.html" {
		t.Fatalf("proposal target = %q, want user-facing package entrypoint", target)
	}
	for _, want := range []string{"index.html", "README.md", "PROOF.md", "project-package.json"} {
		if !containsToolName(confirmedActionStringSlice(resultContract["files_required"]), want) {
			t.Fatalf("files_required = %#v, missing %q", resultContract["files_required"], want)
		}
	}
	if !containsToolName(confirmedActionStringSlice(resultContract["expected_outputs"]), "openable application package") {
		t.Fatalf("expected_outputs = %#v", resultContract["expected_outputs"])
	}
}

func TestBuildPlannedToolCalls_OutcomeLanguageCreatesDeliveryTeam(t *testing.T) {
	request := "Develop a playable browser game with a clear objective, controls, restart, and a direct launch link."

	mutationTools := inferMutationToolsFromText(request)
	for _, want := range []string{"write_file", "generate_blueprint", "delegate"} {
		if !containsToolName(mutationTools, want) {
			t.Fatalf("mutation tools = %#v, missing %q", mutationTools, want)
		}
	}
	result, ok := deterministicGovernedMutationResult(request, mutationTools)
	if !ok {
		t.Fatal("natural complex deliverable must enter the governed proposal path")
	}
	if strings.Contains(result.Text, "TEAM_EVOCATION.md") || !strings.Contains(result.Text, "/generated/package/index.html") {
		t.Fatalf("proposal text = %q, want generated package target instead of planning brief", result.Text)
	}
	calls := buildPlannedToolCalls(result, request, result.ToolsUsed)
	requirePlannedCallNames(t, calls, "create_team", "write_file", "write_file", "delegate_task")

	teamID := firstNonEmptyString(calls[0].Arguments["team_id"])
	if teamID == "" {
		t.Fatal("complex deliverable must receive a Soma-generated team ID")
	}
	for _, call := range calls {
		if path := firstNonEmptyString(call.Arguments["path"]); strings.HasPrefix(path, "workspace/generated/") {
			t.Fatalf("complex deliverable leaked into general output bucket: %q", path)
		}
	}
	if calls[2].Arguments["path"] != "groups/"+teamID+"/planning/RESEARCH_COUNCIL_HANDOFF.md" {
		t.Fatalf("handoff path = %#v, want isolated team planning path", calls[2].Arguments["path"])
	}
	ask := mapArgument(calls[3].Arguments["ask"])
	ownedScope := confirmedActionStringSlice(ask["owned_scope"])
	if !containsString(ownedScope, "groups/"+teamID) {
		t.Fatalf("owned_scope = %#v, want isolated team workspace", ownedScope)
	}
}

func TestBuildPlannedToolCalls_SVGWebPageStaysOnCodePackagePath(t *testing.T) {
	request := "Create a web page and using SVG code imagine and create an image of the Mycelis infrastructure and your place in it."
	mutationTools := inferMutationToolsFromText(request)
	for _, want := range []string{"write_file", "generate_blueprint", "delegate"} {
		if !containsToolName(mutationTools, want) {
			t.Fatalf("mutation tools = %#v, missing %q", mutationTools, want)
		}
	}
	for _, forbidden := range []string{"generate_image", "save_cached_image"} {
		if containsToolName(mutationTools, forbidden) {
			t.Fatalf("mutation tools = %#v, SVG web page must not require %q", mutationTools, forbidden)
		}
	}
	result, ok := deterministicGovernedMutationResult(request, mutationTools)
	if !ok {
		t.Fatal("SVG web page must enter the governed application delivery path")
	}
	calls := buildPlannedToolCalls(result, request, result.ToolsUsed)
	requirePlannedCallNames(t, calls, "create_team", "write_file", "write_file", "delegate_task")
	display := buildProposalDisplayContract(calls, request, result.ToolsUsed)
	if display.WorkIntent == nil || display.WorkIntent.OutputContract == nil || display.WorkIntent.OutputContract.Shape != "app_package" {
		t.Fatalf("work intent = %#v, want app_package output", display.WorkIntent)
	}
}

func TestBuildPlannedToolCalls_UsesRequestedPackageTargetAndTitle(t *testing.T) {
	request := strings.Join([]string{
		"Create a team with team_id qa-game-team named QA Game Team.",
		"Have that team build a playable browser game project package.",
		"Retain it at groups/qa-game-team/generated/first-game with entrypoint groups/qa-game-team/generated/first-game/index.html.",
		"Use the package title QA Game Team First Playable.",
		"After approval, return a retained project_package output with entrypoint, folder, files, validation, and proof.",
	}, " ")

	calls := plannedCallsFromWrongBlueprint(request, []string{"generate_blueprint", "delegate"})
	requirePlannedCallNames(t, calls, "create_team", "write_file", "write_file", "delegate_task")

	ask := mapArgument(calls[3].Arguments["ask"])
	context := mapArgument(ask["context"])
	resultContract := mapArgument(context["result_contract"])
	if resultContract["package_folder"] != "groups/qa-game-team/generated/first-game" {
		t.Fatalf("package_folder = %#v", resultContract["package_folder"])
	}
	if resultContract["package_entrypoint"] != "groups/qa-game-team/generated/first-game/index.html" {
		t.Fatalf("package_entrypoint = %#v", resultContract["package_entrypoint"])
	}
	if resultContract["package_title"] != "QA Game Team First Playable" {
		t.Fatalf("package_title = %#v", resultContract["package_title"])
	}
	if target := firstPlannedOutputTarget(calls); target != "groups/qa-game-team/generated/first-game/index.html" {
		t.Fatalf("proposal target = %q, want requested entrypoint", target)
	}
	display := buildProposalDisplayContract(calls, request, []string{"create_team", "write_file", "delegate_task"})
	if !strings.Contains(display.ExpectedResult, "QA Game Team First Playable") || strings.Contains(display.ExpectedResult, "groups/") {
		t.Fatalf("expected_result = %q, want named result without storage path", display.ExpectedResult)
	}
	if strings.Contains(display.ExpectedResult, "TEAM_EVOCATION.md") || strings.Contains(display.ExpectedResult, "RESEARCH_COUNCIL_HANDOFF.md") {
		t.Fatalf("expected_result = %q, planning handoff must not replace the user deliverable", display.ExpectedResult)
	}
	if display.WorkIntent == nil || display.WorkIntent.OutputContract == nil {
		t.Fatalf("work intent output contract = %#v", display.WorkIntent)
	}
	if got := display.WorkIntent.OutputContract.PrimaryDeliverable; got != "groups/qa-game-team/generated/first-game/index.html" {
		t.Fatalf("primary deliverable = %q, want requested package entrypoint", got)
	}
}

func TestDeliveryTeamInferenceLeavesExplicitSingleFileAskDirect(t *testing.T) {
	request := "Create a Python script at workspace/generated/check.py that prints the current status."
	if requestRequiresDeliveryTeam(strings.ToLower(request)) {
		t.Fatal("explicit single-file work must not instantiate a delivery team")
	}
	if got := extensionForWriteFileRequest("Research JavaScript frameworks for a browser app"); got != ".html" {
		t.Fatalf("JavaScript extension = %q, want .html and never accidental .py", got)
	}
}

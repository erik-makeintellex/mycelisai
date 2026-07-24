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
	for _, want := range []string{"README.md", "PROOF.md", "project-package.json"} {
		if !containsToolName(confirmedActionStringSlice(resultContract["files_required"]), want) {
			t.Fatalf("files_required = %#v, missing %q", resultContract["files_required"], want)
		}
	}
	if !containsToolName(confirmedActionStringSlice(resultContract["expected_outputs"]), "openable application package") {
		t.Fatalf("expected_outputs = %#v", resultContract["expected_outputs"])
	}
}

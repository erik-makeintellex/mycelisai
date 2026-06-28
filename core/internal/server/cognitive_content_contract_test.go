package server

import (
	"regexp"
	"strings"
	"testing"
)

func TestInferCreateTeamPlanFromRequest_GeneratedTeamIDUsesReadableUUIDSuffix(t *testing.T) {
	call, ok := inferCreateTeamPlanFromRequest("Create a temporary game team to build a playable browser game.")
	if !ok {
		t.Fatal("expected create_team plan")
	}
	name, _ := call.Arguments["name"].(string)
	if name != "Temporary Game Delivery Team" {
		t.Fatalf("name = %q, want Soma-inferred readable game team name", name)
	}
	teamID, _ := call.Arguments["team_id"].(string)
	if !regexp.MustCompile(`^temp-game-delivery-team-[0-9a-f]{5}$`).MatchString(teamID) {
		t.Fatalf("team_id = %q, want temp prefix plus five-char uuid suffix", teamID)
	}
}

func TestInferCreateTeamPlanFromRequest_ExplicitTeamNameStillWins(t *testing.T) {
	call, ok := inferCreateTeamPlanFromRequest("Create a team named Arcade Reliability Crew with goal test the generated game.")
	if !ok {
		t.Fatal("expected create_team plan")
	}
	name, _ := call.Arguments["name"].(string)
	if name != "Arcade Reliability Crew" {
		t.Fatalf("name = %q, want explicit operator name", name)
	}
}

func TestInferCreateTeamPlanFromRequest_BackendChoosesStewardName(t *testing.T) {
	call, ok := inferCreateTeamPlanFromRequest("Create a standing team to watch generated content and react to changes.")
	if !ok {
		t.Fatal("expected create_team plan")
	}
	name, _ := call.Arguments["name"].(string)
	if name != "Standing Content Steward Team" {
		t.Fatalf("name = %q, want Soma-inferred steward name", name)
	}
	teamID, _ := call.Arguments["team_id"].(string)
	if !regexp.MustCompile(`^standing-content-steward-team-[0-9a-f]{5}$`).MatchString(teamID) {
		t.Fatalf("team_id = %q, want standing steward prefix plus five-char uuid suffix", teamID)
	}
}

func TestInferCreateTeamPlanFromRequest_ContentContractCoversMixedGameMediaText(t *testing.T) {
	request := "Create a standing team to build a browser game, generate cover art media, and write a README report."
	call, ok := inferCreateTeamPlanFromRequest(request)
	if !ok {
		t.Fatal("expected create_team plan")
	}
	teamID, _ := call.Arguments["team_id"].(string)
	if !regexp.MustCompile(`^standing-mixed-output-team-[0-9a-f]{5}$`).MatchString(teamID) {
		t.Fatalf("team_id = %q, want standing prefix plus five-char uuid suffix", teamID)
	}
	contract, ok := call.Arguments["content_contract"].(map[string]any)
	if !ok {
		t.Fatalf("content_contract = %#v", call.Arguments["content_contract"])
	}
	contentTypes := confirmedActionStringSlice(contract["content_types"])
	for _, want := range []string{"game", "media", "text"} {
		if !containsString(contentTypes, want) {
			t.Fatalf("content_types = %#v, missing %q", contentTypes, want)
		}
	}
	criteria := strings.Join(confirmedActionStringSlice(contract["acceptance_criteria"]), "\n")
	for _, want := range []string{"playable controls", "winning route", "play-tests", "Soma as a repair request", "direct launch", "music or action audio", "visual review", "structure matches"} {
		if !strings.Contains(criteria, want) {
			t.Fatalf("criteria = %q, missing %q", criteria, want)
		}
	}
	proof := strings.Join(confirmedActionStringSlice(contract["proof_required"]), "\n")
	for _, want := range []string{"gameplay proof", "screenshots", "Soma repair-turn transcript", "Resources launch", "audio unlock"} {
		if !strings.Contains(proof, want) {
			t.Fatalf("proof = %q, missing %q", proof, want)
		}
	}
	capabilities := confirmedActionStringSlice(call.Arguments["required_capabilities"])
	for _, want := range []string{"write_file", "generate_image", "save_cached_image", "store_artifact", "research_for_blueprint", "consult_council"} {
		if !containsString(capabilities, want) {
			t.Fatalf("capabilities = %#v, missing %q", capabilities, want)
		}
	}
	prep := strings.Join(confirmedActionStringSlice(contract["team_preparation"]), "\n")
	for _, want := range []string{"research", "consult council", "implementation stack", "specialist additions"} {
		if !strings.Contains(prep, want) {
			t.Fatalf("team_preparation = %q, missing %q", prep, want)
		}
	}
	evocation, ok := call.Arguments["team_evocation"].(map[string]any)
	if !ok {
		t.Fatalf("team_evocation = %#v", call.Arguments["team_evocation"])
	}
	if evocation["mode"] != "research_council_then_staff" || evocation["research_required"] != true || evocation["council_review_required"] != true {
		t.Fatalf("team_evocation = %#v, want research/council staffing mode", evocation)
	}
}

func TestInferWriteFilePlanFromRequest_TextValidationMetadata(t *testing.T) {
	call, ok := inferWriteFilePlanFromRequest("Write a markdown report at workspace/logs/review.md about the game proof.")
	if !ok {
		t.Fatal("expected write_file plan")
	}
	validation, _ := call.Arguments["validation"].(string)
	if !strings.Contains(validation, "Retained text output") || !strings.Contains(validation, "requested structure") {
		t.Fatalf("validation = %q", validation)
	}
}

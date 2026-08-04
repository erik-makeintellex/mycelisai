package server

import (
	"regexp"
	"strings"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
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
	if call.Arguments["profile_ref"] != "default.builder" {
		t.Fatalf("profile_ref = %#v, want default.builder", call.Arguments["profile_ref"])
	}
}

func TestInferCreateTeamPlanFromRequest_ContentContractCoversTableAndAppOutputs(t *testing.T) {
	call, ok := inferCreateTeamPlanFromRequest("Create a team to build a commercial data app with a customer risk table and CSV export.")
	if !ok {
		t.Fatal("expected create_team plan")
	}
	name, _ := call.Arguments["name"].(string)
	if name != "Application Delivery Team" {
		t.Fatalf("name = %q, want application delivery team", name)
	}
	teamID, _ := call.Arguments["team_id"].(string)
	if !regexp.MustCompile(`^application-delivery-team-[0-9a-f]{5}$`).MatchString(teamID) {
		t.Fatalf("team_id = %q, want application delivery prefix plus five-char uuid suffix", teamID)
	}
	contract, ok := call.Arguments["content_contract"].(map[string]any)
	if !ok {
		t.Fatalf("content_contract = %#v", call.Arguments["content_contract"])
	}
	contentTypes := confirmedActionStringSlice(contract["content_types"])
	for _, want := range []string{"table_data", "application_package"} {
		if !containsString(contentTypes, want) {
			t.Fatalf("content_types = %#v, missing %q", contentTypes, want)
		}
	}
	criteria := strings.Join(confirmedActionStringSlice(contract["acceptance_criteria"]), "\n")
	for _, want := range []string{"columns and rows", "direct open or launch path", "data-mycelis-primary-action", "primary user workflows", "browser or runtime validation"} {
		if !strings.Contains(criteria, want) {
			t.Fatalf("criteria = %q, missing %q", criteria, want)
		}
	}
	proof := strings.Join(confirmedActionStringSlice(contract["proof_required"]), "\n")
	for _, want := range []string{"application entrypoint", "headed browser", "validation report", "Resources reference"} {
		if !strings.Contains(proof, want) {
			t.Fatalf("proof = %q, missing %q", proof, want)
		}
	}
	capabilities := confirmedActionStringSlice(call.Arguments["required_capabilities"])
	for _, want := range []string{"write_file", "store_artifact", "research_for_blueprint", "consult_council"} {
		if !containsString(capabilities, want) {
			t.Fatalf("capabilities = %#v, missing %q", capabilities, want)
		}
	}
	evocation, ok := call.Arguments["team_evocation"].(map[string]any)
	if !ok {
		t.Fatalf("team_evocation = %#v", call.Arguments["team_evocation"])
	}
	workstreams := strings.Join(confirmedActionStringSlice(evocation["suggested_workstreams"]), "\n")
	for _, want := range []string{"product/application lead", "browser/runtime QA lead", "data model lead"} {
		if !strings.Contains(workstreams, want) {
			t.Fatalf("workstreams = %q, missing %q", workstreams, want)
		}
	}
}

func TestContentContract_PackageMetadataDoesNotImplyTableData(t *testing.T) {
	contract := contentContractForTeamRequest("Build a browser game package. The package metadata must include index.html, README.md, and validation notes.")
	contentTypes := confirmedActionStringSlice(contract["content_types"])
	if containsString(contentTypes, "table_data") {
		t.Fatalf("content_types = %#v, package metadata should not imply table_data", contentTypes)
	}
	for _, want := range []string{"game", "text"} {
		if !containsString(contentTypes, want) {
			t.Fatalf("content_types = %#v, missing %q", contentTypes, want)
		}
	}
}

func TestContentContract_InteractiveBrowserOutputCarriesRunnableProbe(t *testing.T) {
	contract := contentContractForTeamRequest("Build a browser application for managing customer records.")
	plan, ok := contract["output_validation"].(*protocol.OutputValidationPlan)
	if !ok {
		t.Fatalf("output_validation = %#v, want typed plan", contract["output_validation"])
	}
	if err := plan.Validate(); err != nil {
		t.Fatalf("output validation plan is not runnable: %v", err)
	}
	if plan.Probe.Action.Kind != protocol.OutputValidationActionClick ||
		plan.Probe.Observe.Kind != protocol.OutputValidationObserveVisualChange {
		t.Fatalf("probe = %#v, want generic click and visual-change probe", plan.Probe)
	}
}

func TestContentContract_ExplicitHeldKeyBecomesPrimaryRuntimeProbe(t *testing.T) {
	contract := contentContractForTeamRequest("Build a browser application that visibly moves while ArrowRight is held.")
	plan, ok := contract["output_validation"].(*protocol.OutputValidationPlan)
	if !ok || plan.Probe == nil {
		t.Fatalf("output_validation = %#v, want typed probe", contract["output_validation"])
	}
	if plan.Probe.Action.Kind != protocol.OutputValidationActionKeyHold ||
		plan.Probe.Action.Key != "ArrowRight" || plan.Probe.Action.DurationMS != 600 {
		t.Fatalf("action = %#v, want held ArrowRight probe", plan.Probe.Action)
	}
	if plan.Probe.Observe.Kind != protocol.OutputValidationObserveVisualChange ||
		plan.Probe.Observe.Target != "[data-mycelis-validation-surface]" {
		t.Fatalf("observation = %#v, want visual surface change", plan.Probe.Observe)
	}
}

func TestContentContract_NonInteractiveOutputOmitsBrowserProbe(t *testing.T) {
	for _, request := range []string{
		"Write a markdown research brief.",
		"Build a native executable package for offline use.",
	} {
		contract := contentContractForTeamRequest(request)
		if plan := contract["output_validation"]; plan != nil {
			t.Fatalf("request %q output_validation = %#v, want no browser probe", request, plan)
		}
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

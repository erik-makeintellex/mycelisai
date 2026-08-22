package server

import (
	"slices"
	"strings"
	"testing"
)

func TestWorkerProfileIntentUsesSharedConfigurationTools(t *testing.T) {
	if got := inferReadOnlyConfigToolsFromText("Preview this worker profile"); !slices.Equal(got, []string{"preview_config_document"}) {
		t.Fatalf("read-only tools = %v", got)
	}
	tools, recognized := outcomeTemplateMutationTools("Save this worker profile")
	if !recognized || !slices.Equal(tools, []string{"store_config_document"}) {
		t.Fatalf("mutation tools = %v, recognized = %v", tools, recognized)
	}
	tools, recognized = outcomeTemplateMutationTools("Activate this worker profile")
	if !recognized || !slices.Equal(tools, []string{"activate_config_document"}) {
		t.Fatalf("activation tools = %v, recognized = %v", tools, recognized)
	}
	tools, recognized = outcomeTemplateMutationTools("Should I activate this worker profile?")
	if !recognized || len(tools) != 0 {
		t.Fatalf("decision question tools = %v, recognized = %v", tools, recognized)
	}
}

func TestInlineWorkerProfileBypassesConversationTemplateReview(t *testing.T) {
	s := newTestServer()
	request := "Save this Worker Profile for reuse. Use exactly this YAML:\n" + retainedWorkerProfileYAML
	review := s.buildSomaReferentialReviewUnlessConfigDocument(t.Context(), []chatRequestMessage{{
		Role: "user", Content: request,
	}})
	if review.active() || review.EffectiveRequest != strings.TrimSpace(request) {
		t.Fatalf("inline Worker Profile was intercepted: %#v", review)
	}
}

func TestWorkerProfileTeamRequestRoutesToGovernedCreateTeamPlan(t *testing.T) {
	request := "Create a temporary team using this Worker Profile with team_id qa-profile-a-12345."
	tools := inferMutationToolsFromText(request)
	if len(tools) == 0 {
		t.Fatal("Worker Profile team request was classified as read-only")
	}
	call, ok := inferCreateTeamPlanFromRequest(request)
	if !ok || call.Name != "create_team" {
		t.Fatalf("create-team plan = (%#v, %v)", call, ok)
	}
}

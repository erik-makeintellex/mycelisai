package server

import (
	"net/http"
	"strings"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

func testReviewLoopHome() OrganizationHomePayload {
	return normalizeOrganizationHome(OrganizationHomePayload{
		OrganizationSummary: OrganizationSummary{
			ID:                        "org-review",
			Name:                      "Northstar Labs",
			Purpose:                   "Keep delivery reviews safe and structured",
			StartMode:                 OrganizationStartModeTemplate,
			Status:                    "ready",
			TeamLeadLabel:             "Team Lead",
			AdvisorCount:              1,
			DepartmentCount:           1,
			SpecialistCount:           2,
			AIEngineProfileID:         string(OrganizationAIEngineProfileBalanced),
			AIEngineSettingsSummary:   "Balanced",
			ResponseContractProfileID: string(ResponseContractProfileClearBalanced),
			ResponseContractSummary:   "Clear & Balanced",
			MemoryPersonalitySummary:  "Prepared for guided work",
		},
		Departments: []OrganizationDepartmentSummary{
			{
				ID:              "platform",
				Name:            "Platform Department",
				SpecialistCount: 2,
				AgentTypeProfiles: []OrganizationAgentTypeProfileSummary{
					{
						ID:                               "planner",
						Name:                             "Planner",
						HelpsWith:                        "Turns organization goals into practical next steps.",
						AIEngineBindingProfileID:         string(OrganizationAIEngineProfileHighReasoning),
						ResponseContractBindingProfileID: string(ResponseContractProfileStructuredAnalytical),
					},
					{
						ID:                               "reviewer",
						Name:                             "Reviewer",
						HelpsWith:                        "Checks work for quality and readiness.",
						AIEngineBindingProfileID:         string(OrganizationAIEngineProfileHighReasoning),
						ResponseContractBindingProfileID: string(ResponseContractProfileStructuredAnalytical),
					},
				},
			},
		},
	})
}

func TestHandleTriggerLoop_ExecutesReviewLoopAndStoresResult(t *testing.T) {
	s := newTestServer()
	home := s.organizationStore().Save(testReviewLoopHome())
	s.loopProfileStore().EnsureDefaults(home)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/internal/organizations/{id}/loops/{loopId}/trigger", s.handleTriggerLoop)
	mux.HandleFunc("GET /api/v1/internal/organizations/{id}/loops/results", s.handleListLoopResults)

	triggerRR := doRequest(t, mux, "POST", "/api/v1/internal/organizations/org-review/loops/"+DefaultDepartmentReviewLoopID+"/trigger", "")
	assertStatus(t, triggerRR, http.StatusOK)

	var triggerResp protocol.APIResponse
	assertJSON(t, triggerRR, &triggerResp)
	if !triggerResp.OK {
		t.Fatalf("expected ok trigger response, got %+v", triggerResp)
	}

	result, ok := triggerResp.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected object loop result, got %T", triggerResp.Data)
	}
	if result["loop_id"] != DefaultDepartmentReviewLoopID {
		t.Fatalf("unexpected loop result payload: %+v", result)
	}
	owner, ok := result["owner"].(map[string]any)
	if !ok || owner["type"] != string(LoopOwnerTypeTeam) || owner["id"] != "platform" {
		t.Fatalf("expected team owner resolution, got %+v", result)
	}
	review, ok := result["review"].(map[string]any)
	if !ok {
		t.Fatalf("expected structured review output, got %+v", result)
	}
	if review["status"] == "" {
		t.Fatalf("expected review status, got %+v", review)
	}
	findings, ok := review["findings"].([]any)
	if !ok || len(findings) < 2 {
		t.Fatalf("expected structured findings, got %+v", review)
	}
	suggestions, ok := review["suggestions"].([]any)
	if !ok || len(suggestions) < 2 {
		t.Fatalf("expected structured suggestions, got %+v", review)
	}

	listRR := doRequest(t, mux, "GET", "/api/v1/internal/organizations/org-review/loops/results", "")
	assertStatus(t, listRR, http.StatusOK)

	var listResp protocol.APIResponse
	assertJSON(t, listRR, &listResp)
	results, ok := listResp.Data.([]any)
	if !ok || len(results) != 1 {
		t.Fatalf("expected one stored loop result, got %+v", listResp.Data)
	}

	after, ok := s.organizationStore().Get(home.ID)
	if !ok {
		t.Fatalf("expected organization to remain available after loop execution")
	}
	if after.DepartmentCount != home.DepartmentCount || after.SpecialistCount != home.SpecialistCount || after.AIEngineProfileID != home.AIEngineProfileID {
		t.Fatalf("expected read-only review loop to preserve organization state, before=%+v after=%+v", home, after)
	}
}

func TestHandleTriggerLoop_ResolvesAgentTypeOwner(t *testing.T) {
	s := newTestServer()
	home := s.organizationStore().Save(testReviewLoopHome())
	s.loopProfileStore().EnsureDefaults(home)

	mux := setupMux(t, "POST /api/v1/internal/organizations/{id}/loops/{loopId}/trigger", s.handleTriggerLoop)
	rr := doRequest(t, mux, "POST", "/api/v1/internal/organizations/org-review/loops/"+DefaultAgentTypeReviewLoopID+"/trigger", "")
	assertStatus(t, rr, http.StatusOK)

	var resp protocol.APIResponse
	assertJSON(t, rr, &resp)
	result, ok := resp.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected object loop result, got %T", resp.Data)
	}
	owner, ok := result["owner"].(map[string]any)
	if !ok {
		t.Fatalf("expected owner object, got %+v", result)
	}
	if owner["type"] != string(LoopOwnerTypeAgentType) || owner["id"] != "planner" || owner["name"] != "Planner" {
		t.Fatalf("expected agent type owner resolution, got %+v", owner)
	}
}

func TestHandleTriggerLoop_RejectsInvalidLoopProfile(t *testing.T) {
	s := newTestServer()
	home := s.organizationStore().Save(testReviewLoopHome())
	s.loopProfileStore().Save(home.ID, LoopProfile{
		ID:   "scheduled-review",
		Name: "Scheduled review",
		Type: LoopProfileType("scheduled"),
		Owner: LoopOwnerRef{
			Type: LoopOwnerTypeTeam,
			ID:   "platform",
		},
	})

	mux := setupMux(t, "POST /api/v1/internal/organizations/{id}/loops/{loopId}/trigger", s.handleTriggerLoop)
	rr := doRequest(t, mux, "POST", "/api/v1/internal/organizations/org-review/loops/scheduled-review/trigger", "")
	assertStatus(t, rr, http.StatusBadRequest)

	var resp protocol.APIResponse
	assertJSON(t, rr, &resp)
	if resp.OK {
		t.Fatalf("expected invalid loop rejection, got %+v", resp)
	}
	if !strings.Contains(resp.Error, "review loop") {
		t.Fatalf("expected review-loop validation message, got %+v", resp)
	}
}

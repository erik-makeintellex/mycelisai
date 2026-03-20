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

func TestHandleListLoopActivity_ReturnsSafeRecentActivity(t *testing.T) {
	s := newTestServer()
	home := s.organizationStore().Save(testReviewLoopHome())
	s.loopResultStore().Add(home.ID, ReviewLoopResult{
		ID:              "result-1",
		LoopID:          DefaultDepartmentReviewLoopID,
		LoopName:        "Department readiness review",
		OrganizationID:  home.ID,
		ActivityStatus:  LoopActivityStatusSuccess,
		ActivitySummary: "No issues detected",
		ReviewedAt:      "2026-03-19T18:00:00Z",
	})
	s.loopResultStore().Add(home.ID, ReviewLoopResult{
		ID:              "result-2",
		LoopID:          DefaultAgentTypeReviewLoopID,
		LoopName:        "Agent type readiness review",
		OrganizationID:  home.ID,
		ActivityStatus:  LoopActivityStatusWarning,
		ActivitySummary: "2 items flagged",
		ReviewedAt:      "2026-03-19T18:01:00Z",
	})

	mux := setupMux(t, "GET /api/v1/organizations/{id}/loop-activity", s.handleListLoopActivity)
	rr := doRequest(t, mux, "GET", "/api/v1/organizations/org-review/loop-activity", "")
	assertStatus(t, rr, http.StatusOK)

	var resp protocol.APIResponse
	assertJSON(t, rr, &resp)
	items, ok := resp.Data.([]any)
	if !ok || len(items) != 2 {
		t.Fatalf("expected two activity entries, got %+v", resp.Data)
	}

	first, ok := items[0].(map[string]any)
	if !ok {
		t.Fatalf("expected object activity item, got %T", items[0])
	}
	if first["name"] != "Specialist review" || first["status"] != string(LoopActivityStatusWarning) || first["summary"] != "2 items flagged" {
		t.Fatalf("unexpected recent activity payload: %+v", first)
	}
	if strings.Contains(strings.ToLower(first["name"].(string)), "loop") || strings.Contains(strings.ToLower(first["name"].(string)), "scheduler") {
		t.Fatalf("expected safe activity wording, got %+v", first)
	}
}

func TestTriggerReviewLoopsForEvent_ExecutesMatchingProfilesAndStoresActivity(t *testing.T) {
	s := newTestServer()
	home := s.organizationStore().Save(testReviewLoopHome())
	s.loopProfileStore().EnsureDefaults(home)

	stats, err := s.triggerReviewLoopsForEvent(home.ID, ReviewLoopEventOrganizationCreated)
	if err != nil {
		t.Fatalf("triggerReviewLoopsForEvent returned error: %v", err)
	}
	if stats.executed != 2 || stats.failed != 0 || stats.skippedBusy != 0 {
		t.Fatalf("expected both default review loops to execute, got %+v", stats)
	}

	results := s.loopResultStore().List(home.ID)
	if len(results) != 2 {
		t.Fatalf("expected two stored event-driven results, got %+v", results)
	}
	for _, result := range results {
		if result.Trigger != "event:organization_created" {
			t.Fatalf("expected event-driven trigger label, got %+v", result)
		}
		if result.ActivitySummary == "" {
			t.Fatalf("expected safe activity summary, got %+v", result)
		}
	}
}

func TestTriggerReviewLoopsForEvent_SkipsBusyLoop(t *testing.T) {
	s := newTestServer()
	home := s.organizationStore().Save(testReviewLoopHome())
	s.loopProfileStore().EnsureDefaults(home)

	key := scheduledLoopExecutionKey(home.ID, DefaultDepartmentReviewLoopID)
	if !s.loopExecutionTracker().TryStart(key) {
		t.Fatal("expected to reserve execution tracker for overlap test")
	}
	defer s.loopExecutionTracker().Finish(key)

	stats, err := s.triggerReviewLoopsForEvent(home.ID, ReviewLoopEventTeamLeadActionCompleted)
	if err != nil {
		t.Fatalf("triggerReviewLoopsForEvent returned error: %v", err)
	}
	if stats.executed != 0 || stats.failed != 0 || stats.skippedBusy != 1 {
		t.Fatalf("expected one busy skip and no executions, got %+v", stats)
	}
	if results := s.loopResultStore().List(home.ID); len(results) != 0 {
		t.Fatalf("expected no stored results when the matching loop is already running, got %+v", results)
	}
}

func TestTriggerReviewLoopsForEvent_RejectsUnknownEvent(t *testing.T) {
	s := newTestServer()
	home := s.organizationStore().Save(testReviewLoopHome())

	_, err := s.triggerReviewLoopsForEvent(home.ID, ReviewLoopEventKind("external_api_called"))
	if err == nil || !strings.Contains(err.Error(), "allowed internal review events") {
		t.Fatalf("expected unknown event rejection, got %v", err)
	}
}

func TestTriggerReviewLoopsForEvent_RecordsFailureActivity(t *testing.T) {
	s := newTestServer()
	home := s.organizationStore().Save(testReviewLoopHome())
	s.loopProfileStore().Save(home.ID, LoopProfile{
		ID:          "missing-owner-review",
		Name:        "Department readiness review",
		Type:        LoopProfileTypeReview,
		Description: "Uses a missing owner to force a safe failure record.",
		Owner: LoopOwnerRef{
			Type: LoopOwnerTypeTeam,
			ID:   "missing-department",
		},
		EventTriggers: []ReviewLoopEventKind{ReviewLoopEventOrganizationCreated},
	})

	stats, err := s.triggerReviewLoopsForEvent(home.ID, ReviewLoopEventOrganizationCreated)
	if err != nil {
		t.Fatalf("triggerReviewLoopsForEvent returned error: %v", err)
	}
	if stats.failed != 1 {
		t.Fatalf("expected one failed activity record, got %+v", stats)
	}

	results := s.loopResultStore().List(home.ID)
	if len(results) != 1 {
		t.Fatalf("expected one stored failure result, got %+v", results)
	}
	if results[0].ActivityStatus != LoopActivityStatusFailed || results[0].ActivitySummary != "Review owner unavailable" {
		t.Fatalf("expected safe failed activity wording, got %+v", results[0])
	}
}

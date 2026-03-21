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

func TestHandleListLearningInsights_ReturnsSafeReadableInsights(t *testing.T) {
	s := newTestServer()
	home := s.organizationStore().Save(testReviewLoopHome())
	s.loopResultStore().Add(home.ID, ReviewLoopResult{
		ID:             "result-1",
		LoopID:         DefaultDepartmentReviewLoopID,
		LoopName:       "Department readiness review",
		OrganizationID: home.ID,
		Owner: LoopOwnerResolution{
			Type:      LoopOwnerTypeTeam,
			ID:        "platform",
			Name:      "Platform Department",
			HelpsWith: "Platform Department organizes 2 Specialists for this AI Organization.",
		},
		ActivityStatus:  LoopActivityStatusSuccess,
		ActivitySummary: "No issues detected",
		ReviewedAt:      "2026-03-19T18:02:00Z",
	})
	s.loopResultStore().Add(home.ID, ReviewLoopResult{
		ID:             "result-2",
		LoopID:         DefaultDepartmentReviewLoopID,
		LoopName:       "Department readiness review",
		OrganizationID: home.ID,
		Owner: LoopOwnerResolution{
			Type:      LoopOwnerTypeTeam,
			ID:        "platform",
			Name:      "Platform Department",
			HelpsWith: "Platform Department organizes 2 Specialists for this AI Organization.",
		},
		ActivityStatus:  LoopActivityStatusSuccess,
		ActivitySummary: "No issues detected",
		ReviewedAt:      "2026-03-19T18:01:00Z",
	})
	s.loopResultStore().Add(home.ID, ReviewLoopResult{
		ID:             "result-3",
		LoopID:         DefaultDepartmentReviewLoopID,
		LoopName:       "Department readiness review",
		OrganizationID: home.ID,
		Owner: LoopOwnerResolution{
			Type:      LoopOwnerTypeTeam,
			ID:        "platform",
			Name:      "Platform Department",
			HelpsWith: "Platform Department organizes 2 Specialists for this AI Organization.",
		},
		ActivityStatus:  LoopActivityStatusSuccess,
		ActivitySummary: "No issues detected",
		ReviewedAt:      "2026-03-19T18:00:00Z",
	})
	s.loopResultStore().Add(home.ID, ReviewLoopResult{
		ID:             "result-4",
		LoopID:         DefaultAgentTypeReviewLoopID,
		LoopName:       "Agent type readiness review",
		OrganizationID: home.ID,
		Owner: LoopOwnerResolution{
			Type:      LoopOwnerTypeAgentType,
			ID:        "planner",
			Name:      "Planner",
			HelpsWith: "Turns organization goals into practical next steps, delivery sequencing, and clear priorities.",
		},
		ActivityStatus:  LoopActivityStatusWarning,
		ActivitySummary: "2 items flagged",
		ReviewedAt:      "2026-03-19T17:59:00Z",
	})

	mux := setupMux(t, "GET /api/v1/organizations/{id}/learning-insights", s.handleListLearningInsights)
	rr := doRequest(t, mux, "GET", "/api/v1/organizations/org-review/learning-insights", "")
	assertStatus(t, rr, http.StatusOK)

	var resp protocol.APIResponse
	assertJSON(t, rr, &resp)
	items, ok := resp.Data.([]any)
	if !ok || len(items) != 2 {
		t.Fatalf("expected two insight entries, got %+v", resp.Data)
	}

	first, ok := items[0].(map[string]any)
	if !ok {
		t.Fatalf("expected object insight item, got %T", items[0])
	}
	if first["source"] != "Team: Platform Department" {
		t.Fatalf("expected safe team source label, got %+v", first)
	}
	if first["strength"] != string(LearningInsightStrengthStrong) {
		t.Fatalf("expected strong strength label, got %+v", first)
	}
	summary, _ := first["summary"].(string)
	if !strings.Contains(summary, "Platform Department is building a steadier execution lane") {
		t.Fatalf("expected readable learning summary, got %+v", first)
	}
	if strings.Contains(strings.ToLower(summary), "vector") || strings.Contains(strings.ToLower(summary), "loop") || strings.Contains(strings.ToLower(summary), "memory promotion") {
		t.Fatalf("expected no internal wording in learning summary, got %+v", first)
	}

	second, ok := items[1].(map[string]any)
	if !ok {
		t.Fatalf("expected object insight item, got %T", items[1])
	}
	if second["source"] != "Specialist role: Planner" {
		t.Fatalf("expected safe role source label, got %+v", second)
	}
	if second["strength"] != string(LearningInsightStrengthEmerging) {
		t.Fatalf("expected emerging strength label, got %+v", second)
	}
	secondSummary, _ := second["summary"].(string)
	if !strings.Contains(secondSummary, "Planner specialists are identifying recurring gaps") {
		t.Fatalf("expected readable role learning summary, got %+v", second)
	}
}

func TestHandleListAutomations_ReturnsSafeReadableDefinitions(t *testing.T) {
	s := newTestServer()
	home := s.organizationStore().Save(testReviewLoopHome())
	s.loopProfileStore().EnsureDefaults(home)
	s.loopResultStore().Add(home.ID, ReviewLoopResult{
		ID:              "result-1",
		LoopID:          DefaultDepartmentReviewLoopID,
		LoopName:        "Department readiness review",
		OrganizationID:  home.ID,
		ActivityStatus:  LoopActivityStatusSuccess,
		ActivitySummary: "No issues detected",
		ReviewedAt:      "2026-03-20T18:00:00Z",
	})
	s.loopResultStore().Add(home.ID, ReviewLoopResult{
		ID:              "result-2",
		LoopID:          DefaultAgentTypeReviewLoopID,
		LoopName:        "Agent type readiness review",
		OrganizationID:  home.ID,
		ActivityStatus:  LoopActivityStatusWarning,
		ActivitySummary: "2 items flagged",
		ReviewedAt:      "2026-03-20T18:01:00Z",
	})

	mux := setupMux(t, "GET /api/v1/organizations/{id}/automations", s.handleListAutomations)
	rr := doRequest(t, mux, "GET", "/api/v1/organizations/org-review/automations", "")
	assertStatus(t, rr, http.StatusOK)

	var resp protocol.APIResponse
	assertJSON(t, rr, &resp)
	items, ok := resp.Data.([]any)
	if !ok || len(items) != 2 {
		t.Fatalf("expected two automation entries, got %+v", resp.Data)
	}

	first, ok := items[0].(map[string]any)
	if !ok {
		t.Fatalf("expected object automation item, got %T", items[0])
	}
	if first["name"] != "Agent type readiness review" {
		t.Fatalf("expected readable automation name, got %+v", first)
	}
	if first["trigger_type"] != string(AutomationTriggerTypeEventDriven) {
		t.Fatalf("expected event-driven trigger type, got %+v", first)
	}
	if first["owner_label"] != "Specialist role: Planner" {
		t.Fatalf("expected safe owner label, got %+v", first)
	}
	if !strings.Contains(first["trigger_summary"].(string), "organization setup") {
		t.Fatalf("expected safe trigger summary, got %+v", first)
	}
	if strings.Contains(strings.ToLower(first["trigger_summary"].(string)), "scheduler") || strings.Contains(strings.ToLower(first["trigger_summary"].(string)), "loop profile") {
		t.Fatalf("expected no internal wording in trigger summary, got %+v", first)
	}
	outcomes, ok := first["recent_outcomes"].([]any)
	if !ok || len(outcomes) != 1 {
		t.Fatalf("expected recent outcomes, got %+v", first)
	}
}

func TestHandleListAutomations_FallsBackGracefullyWhenOwnerIsUnavailable(t *testing.T) {
	s := newTestServer()
	home := s.organizationStore().Save(testReviewLoopHome())
	s.loopProfileStore().Save(home.ID, LoopProfile{
		ID:          "missing-owner-review",
		Name:        "Missing owner review",
		Type:        LoopProfileTypeReview,
		Description: "Keeps a safe record even if the original owner is unavailable.",
		Owner: LoopOwnerRef{
			Type: LoopOwnerTypeTeam,
			ID:   "missing-team",
		},
		EventTriggers: []ReviewLoopEventKind{ReviewLoopEventOrganizationCreated},
	})

	mux := setupMux(t, "GET /api/v1/organizations/{id}/automations", s.handleListAutomations)
	rr := doRequest(t, mux, "GET", "/api/v1/organizations/org-review/automations", "")
	assertStatus(t, rr, http.StatusOK)

	var resp protocol.APIResponse
	assertJSON(t, rr, &resp)
	items, ok := resp.Data.([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("expected one automation entry, got %+v", resp.Data)
	}

	first, ok := items[0].(map[string]any)
	if !ok {
		t.Fatalf("expected object automation item, got %T", items[0])
	}
	if first["owner_label"] != "Team: Team owner unavailable" {
		t.Fatalf("expected safe unavailable owner label, got %+v", first)
	}
	if !strings.Contains(first["watches"].(string), "AI Organization") || strings.Contains(strings.ToLower(first["watches"].(string)), "loop") {
		t.Fatalf("expected safe watches summary, got %+v", first)
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

package server

import (
	"net/http"
	"strings"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

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

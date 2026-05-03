package server

import (
	"net/http"
	"strings"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

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

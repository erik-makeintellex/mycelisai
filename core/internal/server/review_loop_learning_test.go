package server

import (
	"net/http"
	"strings"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

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

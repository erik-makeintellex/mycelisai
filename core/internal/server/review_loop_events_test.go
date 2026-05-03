package server

import (
	"strings"
	"testing"
)

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

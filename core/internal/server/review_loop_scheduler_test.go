package server

import (
	"sync/atomic"
	"testing"
	"time"
)

func testScheduledDepartmentLoop(intervalSeconds int) LoopProfile {
	return LoopProfile{
		ID:          "scheduled-department-review",
		Name:        "Scheduled Department Review",
		Type:        LoopProfileTypeReview,
		Description: "Reviews a Department on a fixed interval.",
		Owner: LoopOwnerRef{
			Type: LoopOwnerTypeTeam,
			ID:   "platform",
		},
		IntervalSeconds: intervalSeconds,
	}
}

func testScheduledAgentTypeLoop(intervalSeconds int) LoopProfile {
	return LoopProfile{
		ID:          "scheduled-agent-type-review",
		Name:        "Scheduled Agent Type Review",
		Type:        LoopProfileTypeReview,
		Description: "Reviews an Agent Type on a fixed interval.",
		Owner: LoopOwnerRef{
			Type: LoopOwnerTypeAgentType,
			ID:   "planner",
		},
		IntervalSeconds: intervalSeconds,
	}
}

func TestLoopScheduler_ExecutesLoopOnInterval(t *testing.T) {
	s := newTestServer()
	home := s.organizationStore().Save(testReviewLoopHome())
	s.loopProfileStore().Save(home.ID, testScheduledDepartmentLoop(5))

	scheduler := NewLoopScheduler(s)
	startedAt := time.Date(2026, 3, 19, 17, 30, 0, 0, time.UTC)

	firstTick := scheduler.runDueLoopsAt(startedAt)
	if firstTick.executed != 1 || firstTick.failed != 0 {
		t.Fatalf("expected first due loop execution, got %+v", firstTick)
	}

	secondTick := scheduler.runDueLoopsAt(startedAt.Add(4 * time.Second))
	if secondTick.executed != 0 || secondTick.failed != 0 {
		t.Fatalf("expected no execution before interval elapses, got %+v", secondTick)
	}

	thirdTick := scheduler.runDueLoopsAt(startedAt.Add(5 * time.Second))
	if thirdTick.executed != 1 || thirdTick.failed != 0 {
		t.Fatalf("expected second execution at interval boundary, got %+v", thirdTick)
	}

	results := s.loopResultStore().List(home.ID)
	if len(results) != 2 {
		t.Fatalf("expected two stored scheduled results, got %+v", results)
	}
	if results[0].Trigger != "scheduled" || results[1].Trigger != "scheduled" {
		t.Fatalf("expected scheduled trigger labels, got %+v", results)
	}
}

func TestLoopScheduler_HandlesMultipleLoops(t *testing.T) {
	s := newTestServer()
	home := s.organizationStore().Save(testReviewLoopHome())
	s.loopProfileStore().Save(home.ID, testScheduledDepartmentLoop(10))
	s.loopProfileStore().Save(home.ID, testScheduledAgentTypeLoop(10))

	scheduler := NewLoopScheduler(s)
	stats := scheduler.runDueLoopsAt(time.Date(2026, 3, 19, 17, 40, 0, 0, time.UTC))
	if stats.executed != 2 || stats.failed != 0 {
		t.Fatalf("expected two scheduled loop executions, got %+v", stats)
	}

	results := s.loopResultStore().List(home.ID)
	if len(results) != 2 {
		t.Fatalf("expected two stored results, got %+v", results)
	}

	loopIDs := map[string]bool{}
	for _, result := range results {
		loopIDs[result.LoopID] = true
	}
	if !loopIDs["scheduled-department-review"] || !loopIDs["scheduled-agent-type-review"] {
		t.Fatalf("expected both scheduled loops to run, got %+v", results)
	}
}

func TestLoopScheduler_PreventsOverlap(t *testing.T) {
	s := newTestServer()
	home := s.organizationStore().Save(testReviewLoopHome())
	profile := testScheduledDepartmentLoop(1)

	scheduler := NewLoopScheduler(s)
	started := make(chan struct{})
	release := make(chan struct{})
	done := make(chan struct{})
	var executions atomic.Int32

	scheduler.execute = func(home OrganizationHomePayload, profile LoopProfile, trigger string) (ReviewLoopResult, error) {
		executions.Add(1)
		close(started)
		<-release
		return ReviewLoopResult{
			ID:               "scheduled-result",
			LoopID:           profile.ID,
			LoopName:         profile.Name,
			LoopType:         profile.Type,
			OrganizationID:   home.ID,
			OrganizationName: home.Name,
			Trigger:          trigger,
			Owner:            LoopOwnerResolution{Type: profile.Owner.Type, ID: profile.Owner.ID, Name: "Platform Department"},
			Review:           ReviewLoopStructuredOutput{Status: "healthy"},
			ReviewedAt:       time.Now().UTC().Format(time.RFC3339),
		}, nil
	}

	go func() {
		scheduler.tryExecuteScheduledLoop(home, profile, time.Date(2026, 3, 19, 17, 50, 0, 0, time.UTC))
		close(done)
	}()

	<-started
	executed, failed, skippedBusy := scheduler.tryExecuteScheduledLoop(home, profile, time.Date(2026, 3, 19, 17, 50, 1, 0, time.UTC))
	if executed || failed || !skippedBusy {
		t.Fatalf("expected overlapping execution to be skipped, got executed=%v failed=%v skippedBusy=%v", executed, failed, skippedBusy)
	}

	close(release)
	<-done

	if executions.Load() != 1 {
		t.Fatalf("expected only one scheduled execution, got %d", executions.Load())
	}
}

func TestLoopScheduler_RejectsInvalidConfig(t *testing.T) {
	s := newTestServer()
	home := s.organizationStore().Save(testReviewLoopHome())
	s.loopProfileStore().Save(home.ID, LoopProfile{
		ID:          "invalid-interval-review",
		Name:        "Invalid Interval Review",
		Type:        LoopProfileTypeReview,
		Description: "Invalid profile for scheduler rejection.",
		Owner: LoopOwnerRef{
			Type: LoopOwnerTypeTeam,
			ID:   "platform",
		},
		IntervalSeconds: -5,
	})

	scheduler := NewLoopScheduler(s)
	stats := scheduler.runDueLoopsAt(time.Date(2026, 3, 19, 18, 0, 0, 0, time.UTC))
	if stats.executed != 0 || stats.failed != 1 {
		t.Fatalf("expected one invalid-config failure and no executions, got %+v", stats)
	}

	if results := s.loopResultStore().List(home.ID); len(results) != 0 {
		t.Fatalf("expected no stored loop results for invalid config, got %+v", results)
	}
}

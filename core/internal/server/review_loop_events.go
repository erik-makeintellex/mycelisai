package server

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

func isAllowedReviewLoopEventKind(kind ReviewLoopEventKind) bool {
	switch kind {
	case ReviewLoopEventOrganizationCreated,
		ReviewLoopEventTeamLeadActionCompleted,
		ReviewLoopEventOrganizationAIEngineChanged,
		ReviewLoopEventResponseContractChanged:
		return true
	default:
		return false
	}
}

func matchesReviewLoopEvent(profile LoopProfile, eventKind ReviewLoopEventKind) bool {
	for _, trigger := range profile.EventTriggers {
		if trigger == eventKind {
			return true
		}
	}
	return false
}

func loopEventTriggerLabel(eventKind ReviewLoopEventKind) string {
	return "event:" + string(eventKind)
}

func (s *AdminServer) triggerReviewLoopsForEvent(orgID string, eventKind ReviewLoopEventKind) (loopEventDispatchStats, error) {
	if !isAllowedReviewLoopEventKind(eventKind) {
		return loopEventDispatchStats{}, fmt.Errorf("review event must be one of the allowed internal review events")
	}

	home, ok := s.organizationStore().Get(orgID)
	if !ok {
		return loopEventDispatchStats{}, fmt.Errorf("organization not found")
	}

	home = normalizeOrganizationHome(home)
	s.loopProfileStore().EnsureDefaults(home)

	stats := loopEventDispatchStats{}
	triggerLabel := loopEventTriggerLabel(eventKind)
	for _, profile := range s.loopProfileStore().ListByOrganization(home.ID) {
		if !matchesReviewLoopEvent(profile, eventKind) {
			continue
		}

		key := scheduledLoopExecutionKey(home.ID, profile.ID)
		if !s.loopExecutionTracker().TryStart(key) {
			stats.skippedBusy++
			log.Printf("[review-loop-event] organization=%s loop=%s event=%s skipped_previous_execution_still_running=true", home.ID, profile.ID, eventKind)
			continue
		}

		func() {
			defer s.loopExecutionTracker().Finish(key)

			result, err := s.executeReviewLoop(home, profile, triggerLabel)
			if err != nil {
				s.recordFailedLoopResult(home, profile, triggerLabel, err)
				stats.failed++
				log.Printf("[review-loop-event] organization=%s loop=%s event=%s failed error=%v", home.ID, profile.ID, eventKind, err)
				return
			}

			stats.executed++
			log.Printf("[review-loop-event] organization=%s loop=%s event=%s completed result_id=%s", home.ID, profile.ID, eventKind, result.ID)
		}()
	}

	return stats, nil
}

func (s *AdminServer) recordFailedLoopResult(home OrganizationHomePayload, profile LoopProfile, trigger string, err error) ReviewLoopResult {
	status, summary := summarizeActivityFromError(err)
	result := ReviewLoopResult{
		ID:               uuid.NewString(),
		LoopID:           profile.ID,
		LoopName:         profile.Name,
		LoopType:         profile.Type,
		OrganizationID:   home.ID,
		OrganizationName: safeOrganizationName(home.Name),
		Trigger:          trigger,
		ActivityStatus:   status,
		ActivitySummary:  summary,
		ReviewedAt:       time.Now().UTC().Format(time.RFC3339),
	}
	s.loopResultStore().Add(home.ID, result)
	return result
}

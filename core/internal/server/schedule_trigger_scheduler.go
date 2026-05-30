package server

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mycelis/core/internal/triggers"
)

func (ls *LoopScheduler) runDueScheduleTriggersAt(now time.Time) int {
	if ls.server == nil || ls.server.Triggers == nil {
		return 0
	}
	rules, err := ls.server.Triggers.ListDueScheduleRules(context.Background(), now, 25)
	if err != nil {
		log.Printf("[schedule-rules] due query failed: %v", err)
		return 0
	}
	proposed := 0
	for _, rule := range rules {
		if proposeScheduleRule(context.Background(), ls.server.Triggers, rule, now) {
			proposed++
		}
	}
	if proposed > 0 {
		log.Printf("[schedule-rules] proposed=%d", proposed)
	}
	return proposed
}

func proposeScheduleRule(ctx context.Context, store *triggers.Store, rule triggers.TriggerRule, now time.Time) bool {
	if rule.ScheduleIntervalSeconds <= 0 {
		_ = store.LogExecution(ctx, &triggers.TriggerExecution{
			RuleID:     rule.ID,
			EventID:    scheduleRuleEventID(rule.ID, now),
			Status:     "skipped",
			SkipReason: "schedule_interval_seconds required",
		})
		return false
	}

	nextRun := now.Add(time.Duration(rule.ScheduleIntervalSeconds) * time.Second)
	if err := store.MarkScheduleProposed(ctx, rule.ID, now, nextRun); err != nil {
		log.Printf("[schedule-rules] mark proposed failed rule=%s: %v", rule.ID, err)
		return false
	}
	if err := store.LogExecution(ctx, &triggers.TriggerExecution{
		RuleID:  rule.ID,
		EventID: scheduleRuleEventID(rule.ID, now),
		Status:  "proposed",
	}); err != nil {
		log.Printf("[schedule-rules] execution log failed rule=%s: %v", rule.ID, err)
		return false
	}
	return true
}

func scheduleRuleEventID(ruleID string, at time.Time) string {
	return fmt.Sprintf("schedule:%s:%s", ruleID, at.UTC().Format(time.RFC3339Nano))
}

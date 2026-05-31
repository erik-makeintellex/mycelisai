package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/mycelis/core/internal/triggers"
	"github.com/mycelis/core/pkg/protocol"
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
		if ls.server.proposeScheduleRuleHandoff(context.Background(), rule, now) {
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
	return proposeScheduleRuleWithHandoffRefs(ctx, store, rule, now, "", "", nil)
}

func (s *AdminServer) proposeScheduleRuleHandoff(ctx context.Context, rule triggers.TriggerRule, now time.Time) bool {
	if s == nil || s.Triggers == nil {
		return false
	}
	if rule.ScheduleIntervalSeconds <= 0 {
		_ = s.Triggers.LogExecution(ctx, &triggers.TriggerExecution{
			RuleID:     rule.ID,
			EventID:    scheduleRuleEventID(rule.ID, now),
			Status:     "skipped",
			SkipReason: "schedule_interval_seconds required",
		})
		return false
	}

	dueAt := scheduleRuleDueAt(rule, now)
	nextRun := dueAt.Add(time.Duration(rule.ScheduleIntervalSeconds) * time.Second)
	handoffKey := scheduleRuleHandoffKey(rule.ID, dueAt)
	payload := scheduleRuleHandoffPayload(rule, dueAt, nextRun)

	existing, err := s.Triggers.GetExecutionByHandoffKey(ctx, rule.ID, handoffKey)
	if err != nil {
		log.Printf("[schedule-rules] handoff lookup failed rule=%s: %v", rule.ID, err)
		return false
	}
	if existing != nil && existing.IntentProofID != "" && existing.ContractID != "" {
		if err := s.Triggers.MarkScheduleProposed(ctx, rule.ID, dueAt, nextRun); err != nil {
			log.Printf("[schedule-rules] mark existing handoff proposed failed rule=%s: %v", rule.ID, err)
			return false
		}
		return true
	}

	proof, err := s.createIntentProof(protocol.TemplateChatToProposal, scheduleRuleResolvedIntent(rule), scheduleRuleScope(rule), "")
	if err != nil {
		_ = s.Triggers.LogExecution(ctx, &triggers.TriggerExecution{
			RuleID:     rule.ID,
			EventID:    handoffKey,
			Status:     "skipped",
			SkipReason: "schedule_handoff_failed: " + err.Error(),
		})
		log.Printf("[schedule-rules] handoff proof failed rule=%s: %v", rule.ID, err)
		return false
	}
	if proof == nil {
		_ = s.Triggers.LogExecution(ctx, &triggers.TriggerExecution{
			RuleID:     rule.ID,
			EventID:    handoffKey,
			Status:     "skipped",
			SkipReason: "schedule_handoff_failed: database not available",
		})
		return false
	}
	if existing != nil {
		if err := s.Triggers.UpdateExecutionHandoffRefs(ctx, existing.ID, proof.ID, proof.ContractID, "awaiting_approval", payload); err != nil {
			log.Printf("[schedule-rules] update existing handoff refs failed rule=%s: %v", rule.ID, err)
			return false
		}
		if err := s.Triggers.MarkScheduleProposed(ctx, rule.ID, dueAt, nextRun); err != nil {
			log.Printf("[schedule-rules] mark updated handoff proposed failed rule=%s: %v", rule.ID, err)
			return false
		}
		return true
	}
	return proposeScheduleRuleWithHandoffRefs(ctx, s.Triggers, rule, dueAt, proof.ID, proof.ContractID, payload)
}

func proposeScheduleRuleWithHandoffRefs(ctx context.Context, store *triggers.Store, rule triggers.TriggerRule, proposedAt time.Time, intentProofID, contractID string, payload json.RawMessage) bool {
	nextRun := proposedAt.Add(time.Duration(rule.ScheduleIntervalSeconds) * time.Second)
	if err := store.MarkScheduleProposed(ctx, rule.ID, proposedAt, nextRun); err != nil {
		log.Printf("[schedule-rules] mark proposed failed rule=%s: %v", rule.ID, err)
		return false
	}
	handoffKey := ""
	proposalStatus := ""
	eventID := scheduleRuleEventID(rule.ID, proposedAt)
	if intentProofID != "" || contractID != "" {
		handoffKey = scheduleRuleHandoffKey(rule.ID, proposedAt)
		proposalStatus = "awaiting_approval"
		eventID = handoffKey
	}
	if err := store.LogExecution(ctx, &triggers.TriggerExecution{
		RuleID:         rule.ID,
		EventID:        eventID,
		Status:         "proposed",
		HandoffKey:     handoffKey,
		IntentProofID:  intentProofID,
		ContractID:     contractID,
		ProposalStatus: proposalStatus,
		HandoffPayload: payload,
	}); err != nil {
		log.Printf("[schedule-rules] execution log failed rule=%s: %v", rule.ID, err)
		return false
	}
	return true
}

func scheduleRuleEventID(ruleID string, at time.Time) string {
	return fmt.Sprintf("schedule:%s:%s", ruleID, at.UTC().Format(time.RFC3339Nano))
}

func scheduleRuleHandoffKey(ruleID string, dueAt time.Time) string {
	return scheduleRuleEventID(ruleID, dueAt)
}

func scheduleRuleDueAt(rule triggers.TriggerRule, fallback time.Time) time.Time {
	if rule.NextRunAt != nil {
		return *rule.NextRunAt
	}
	return fallback
}

func scheduleRuleResolvedIntent(rule triggers.TriggerRule) string {
	if rule.Name != "" {
		return fmt.Sprintf("schedule cadence proposal: %s", rule.Name)
	}
	return fmt.Sprintf("schedule cadence proposal for %s", rule.TargetMissionID)
}

func scheduleRuleScope(rule triggers.TriggerRule) *protocol.ScopeValidation {
	return &protocol.ScopeValidation{
		Tools:             []string{"schedule_handoff"},
		AffectedResources: []string{"trigger_rules/" + rule.ID, "missions/" + rule.TargetMissionID},
		RiskLevel:         "medium",
		Approval: &protocol.ApprovalPolicy{
			ApprovalRequired:     true,
			ApprovalReason:       "schedule cadence requires operator approval before execution",
			ApprovalMode:         "required",
			CapabilityRisk:       "medium",
			CapabilityIDs:        []string{"scheduler_cadence"},
			RequiredApproverRole: "owner",
			ApprovalSteps:        []string{"review schedule handoff", "approve execution explicitly"},
		},
		CapabilityIDs: []string{"scheduler_cadence"},
	}
}

func scheduleRuleHandoffPayload(rule triggers.TriggerRule, dueAt, nextRunAt time.Time) json.RawMessage {
	payload, _ := json.Marshal(map[string]any{
		"trigger_kind":         "schedule",
		"target_mission_id":    rule.TargetMissionID,
		"proof_expectations":   rule.ProofExpectations,
		"recovery_behavior":    rule.RecoveryBehavior,
		"due_at":               dueAt.UTC().Format(time.RFC3339Nano),
		"next_run_at":          nextRunAt.UTC().Format(time.RFC3339Nano),
		"autonomous_execution": false,
	})
	return payload
}

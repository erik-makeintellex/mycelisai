package triggers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/mycelis/core/internal/events"
	"github.com/mycelis/core/internal/runs"
	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
)

// Engine evaluates trigger rules against incoming mission events.
// It subscribes to the CTS event bus and applies cooldown, recursion,
// and concurrency guards before firing.
type Engine struct {
	store  *Store
	events *events.Store
	runs   *runs.Manager
	nc     *nats.Conn
	sub    *nats.Subscription
	mu     sync.Mutex
}

// NewEngine creates a trigger engine wired to the rule store, event store, and run manager.
// nc may be nil (graceful degradation — engine won't evaluate until NATS is available).
func NewEngine(store *Store, evStore *events.Store, runsManager *runs.Manager, nc *nats.Conn) *Engine {
	return &Engine{
		store:  store,
		events: evStore,
		runs:   runsManager,
		nc:     nc,
	}
}

// Start subscribes to the CTS mission events bus and begins evaluating rules.
// If NATS is nil, Start is a no-op (degraded mode).
func (e *Engine) Start(ctx context.Context) error {
	if e.nc == nil {
		log.Println("[triggers] NATS unavailable — engine in degraded mode")
		return nil
	}

	// Load active rules into cache on startup
	if err := e.store.LoadActiveRules(ctx); err != nil {
		log.Printf("[triggers] failed to load active rules: %v", err)
		// Non-fatal — we'll retry on first event
	}

	// Subscribe to all mission events: swarm.mission.events.*
	topic := protocol.TopicMissionEvents + ".*"
	sub, err := e.nc.Subscribe(topic, func(msg *nats.Msg) {
		e.handleCTSEvent(ctx, msg.Data)
	})
	if err != nil {
		return err
	}

	e.mu.Lock()
	e.sub = sub
	e.mu.Unlock()

	// Re-subscribe on NATS reconnect
	e.nc.SetReconnectHandler(func(c *nats.Conn) {
		log.Println("[triggers] NATS reconnected — reloading rules + re-subscribing")
		e.store.LoadActiveRules(context.Background())
	})

	log.Printf("[triggers] engine started — listening on %s", topic)
	return nil
}

// Stop unsubscribes from NATS.
func (e *Engine) Stop() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.sub != nil {
		e.sub.Unsubscribe()
		e.sub = nil
	}
	log.Println("[triggers] engine stopped")
}

// ReloadRules forces a cache refresh from the database.
func (e *Engine) ReloadRules(ctx context.Context) error {
	return e.store.LoadActiveRules(ctx)
}

// handleCTSEvent processes a lightweight CTS signal from the event bus.
// The CTS payload contains: mission_event_id, run_id, event_type, source_agent.
func (e *Engine) handleCTSEvent(ctx context.Context, data []byte) {
	var cts struct {
		MissionEventID string `json:"mission_event_id"`
		RunID          string `json:"run_id"`
		EventType      string `json:"event_type"`
		SourceAgent    string `json:"source_agent"`
	}
	if err := json.Unmarshal(data, &cts); err != nil {
		log.Printf("[triggers] CTS unmarshal error: %v", err)
		return
	}

	// Skip trigger-originated events to prevent infinite loops at the CTS level
	if cts.EventType == string(protocol.EventTriggerFired) || cts.EventType == string(protocol.EventTriggerSkipped) {
		return
	}

	// Find rules matching this event type
	rules := e.store.MatchingRules(cts.EventType)
	if len(rules) == 0 {
		return
	}

	for _, rule := range rules {
		e.evaluateRule(ctx, rule, cts.MissionEventID, cts.RunID, cts.EventType)
	}
}

// evaluateRule checks all guards and either fires, proposes, or skips the trigger.
func (e *Engine) evaluateRule(ctx context.Context, rule *TriggerRule, eventID, sourceRunID, eventType string) {
	ruleID := rule.ID

	// ── Guard 1: Cooldown ────────────────────────────────────────
	if rule.LastFiredAt != nil {
		elapsed := time.Since(*rule.LastFiredAt)
		if elapsed < time.Duration(rule.CooldownSeconds)*time.Second {
			e.logSkip(ctx, ruleID, eventID, "cooldown",
				"%.0fs since last fire, cooldown is %ds", elapsed.Seconds(), rule.CooldownSeconds)
			return
		}
	}

	// ── Guard 2: Recursion depth ─────────────────────────────────
	// Look up the source run to check its depth
	if e.runs != nil && sourceRunID != "" {
		sourceRun, err := e.runs.GetRun(ctx, sourceRunID)
		if err == nil && sourceRun != nil && sourceRun.RunDepth >= rule.MaxDepth {
			e.logSkip(ctx, ruleID, eventID, "recursion_limit",
				"source run depth %d >= max %d", sourceRun.RunDepth, rule.MaxDepth)
			return
		}
	}

	// ── Guard 3: Concurrency ─────────────────────────────────────
	activeCount, err := e.store.ActiveCount(ctx, rule.TargetMissionID)
	if err != nil {
		log.Printf("[triggers] active count error for rule %s: %v", ruleID, err)
		// Allow through on error — better to fire than silently skip
	} else if activeCount >= rule.MaxActiveRuns {
		e.logSkip(ctx, ruleID, eventID, "concurrency_limit",
			"%d active runs >= max %d", activeCount, rule.MaxActiveRuns)
		return
	}

	// ── Guard 4: Condition match ─────────────────────────────────
	// Condition matching is reserved for future implementation.
	// For now, an empty or "{}" condition always passes.
	if len(rule.Condition) > 2 { // more than just "{}"
		// TODO: implement JSONPath or simple key=value condition matching
		// For V7, non-empty conditions are logged and allowed through
		log.Printf("[triggers] rule %s has condition — condition matching not yet implemented, allowing", ruleID)
	}

	// ── All guards passed — determine action ─────────────────────
	now := time.Now()

	// Determine child run depth
	childDepth := 0
	if e.runs != nil && sourceRunID != "" {
		sourceRun, err := e.runs.GetRun(ctx, sourceRunID)
		if err == nil && sourceRun != nil {
			childDepth = sourceRun.RunDepth + 1
		}
	}

	if rule.Mode == "auto_execute" {
		e.fireTrigger(ctx, rule, eventID, sourceRunID, childDepth, now)
	} else {
		// Default: propose — log as proposed, emit event for governance review
		e.proposeTrigger(ctx, rule, eventID, sourceRunID, now)
	}
}

// fireTrigger creates a child run and emits trigger.fired event.
func (e *Engine) fireTrigger(ctx context.Context, rule *TriggerRule, eventID, sourceRunID string, depth int, now time.Time) {
	// Create child run
	var childRunID string
	if e.runs != nil {
		var err error
		childRunID, err = e.runs.CreateChildRun(ctx, rule.TargetMissionID, sourceRunID, depth)
		if err != nil {
			log.Printf("[triggers] failed to create child run for rule %s: %v", rule.ID, err)
			return
		}
	}

	// Update last_fired_at
	e.store.UpdateLastFired(ctx, rule.ID, now)

	// Emit trigger.fired event on the CHILD run's timeline
	if e.events != nil && childRunID != "" {
		e.events.Emit(ctx, childRunID, protocol.EventTriggerFired, protocol.SeverityInfo,
			"trigger-engine", "", map[string]interface{}{
				"trigger_rule_id":   rule.ID,
				"trigger_rule_name": rule.Name,
				"source_event_id":   eventID,
				"source_run_id":     sourceRunID,
				"target_mission_id": rule.TargetMissionID,
				"mode":              rule.Mode,
			})
	}

	// Log execution
	e.store.LogExecution(ctx, &TriggerExecution{
		RuleID:  rule.ID,
		EventID: eventID,
		RunID:   childRunID,
		Status:  "fired",
	})

	log.Printf("[triggers] FIRED rule %s (%s) → child run %s (depth %d)",
		rule.ID, rule.Name, childRunID, depth)
}

// proposeTrigger logs the trigger as a proposal for human approval.
func (e *Engine) proposeTrigger(ctx context.Context, rule *TriggerRule, eventID, sourceRunID string, now time.Time) {
	// Update last_fired_at (proposals still track cooldown)
	e.store.UpdateLastFired(ctx, rule.ID, now)

	// Emit trigger.fired event on the SOURCE run (no child run yet — needs approval)
	if e.events != nil && sourceRunID != "" {
		e.events.Emit(ctx, sourceRunID, protocol.EventTriggerFired, protocol.SeverityInfo,
			"trigger-engine", "", map[string]interface{}{
				"trigger_rule_id":   rule.ID,
				"trigger_rule_name": rule.Name,
				"source_event_id":   eventID,
				"target_mission_id": rule.TargetMissionID,
				"mode":              "propose",
				"awaiting_approval": true,
			})
	}

	// Log execution as proposed
	e.store.LogExecution(ctx, &TriggerExecution{
		RuleID:  rule.ID,
		EventID: eventID,
		Status:  "proposed",
	})

	log.Printf("[triggers] PROPOSED rule %s (%s) — awaiting approval for mission %s",
		rule.ID, rule.Name, rule.TargetMissionID)
}

// logSkip records a skipped evaluation with reason.
func (e *Engine) logSkip(ctx context.Context, ruleID, eventID, reason string, msgFmt string, args ...interface{}) {
	detail := reason
	if msgFmt != "" {
		detail = reason + ": " + fmt.Sprintf(msgFmt, args...)
	}

	// Emit trigger.skipped event (best-effort, non-blocking)
	// We don't have a run_id here because skip means we didn't create one
	if e.events != nil {
		// Use eventID as a correlator — emit on a synthetic topic
		log.Printf("[triggers] SKIPPED rule %s — %s", ruleID, detail)
	}

	e.store.LogExecution(ctx, &TriggerExecution{
		RuleID:     ruleID,
		EventID:    eventID,
		Status:     "skipped",
		SkipReason: detail,
	})
}


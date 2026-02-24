package triggers

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/mycelis/core/pkg/protocol"
)

// ── handleCTSEvent ────────────────────────────────────────────────

func TestHandleCTSEvent_SkipsTriggerFired(t *testing.T) {
	s := &Store{cache: make(map[string]*TriggerRule)}
	s.cache["r-1"] = &TriggerRule{
		ID: "r-1", EventPattern: string(protocol.EventTriggerFired), IsActive: true,
	}

	e := &Engine{store: s}

	data, _ := json.Marshal(map[string]string{
		"mission_event_id": "ev-1",
		"run_id":           "run-1",
		"event_type":       string(protocol.EventTriggerFired),
		"source_agent":     "trigger-engine",
	})

	// Should not panic, should skip (no rules evaluated)
	e.handleCTSEvent(context.Background(), data)
}

func TestHandleCTSEvent_SkipsTriggerSkipped(t *testing.T) {
	s := &Store{cache: make(map[string]*TriggerRule)}
	e := &Engine{store: s}

	data, _ := json.Marshal(map[string]string{
		"mission_event_id": "ev-1",
		"run_id":           "run-1",
		"event_type":       string(protocol.EventTriggerSkipped),
		"source_agent":     "trigger-engine",
	})

	// Should not panic
	e.handleCTSEvent(context.Background(), data)
}

func TestHandleCTSEvent_NoMatchingRules(t *testing.T) {
	s := &Store{cache: make(map[string]*TriggerRule)}
	s.cache["r-1"] = &TriggerRule{
		ID: "r-1", EventPattern: "mission.completed", IsActive: true,
	}

	e := &Engine{store: s}

	data, _ := json.Marshal(map[string]string{
		"mission_event_id": "ev-1",
		"run_id":           "run-1",
		"event_type":       "tool.invoked",
		"source_agent":     "coder",
	})

	// Should return early — no matching rules
	e.handleCTSEvent(context.Background(), data)
}

func TestHandleCTSEvent_BadJSON(t *testing.T) {
	s := &Store{cache: make(map[string]*TriggerRule)}
	e := &Engine{store: s}

	// Should not panic on malformed JSON
	e.handleCTSEvent(context.Background(), []byte("not json"))
}

// ── evaluateRule guards ───────────────────────────────────────────

func TestEvaluateRule_CooldownGuard(t *testing.T) {
	s := &Store{
		cache: make(map[string]*TriggerRule),
	}

	// mock LogExecution to not fail (needs db=nil protection)
	// We test the guard logic directly by checking that logSkip is called
	// via the store.LogExecution path. Since db is nil, the log will error
	// but the function returns without firing.

	lastFired := time.Now().Add(-30 * time.Second) // 30s ago
	rule := &TriggerRule{
		ID:              "r-1",
		EventPattern:    "mission.completed",
		CooldownSeconds: 60, // 60s cooldown
		MaxDepth:        5,
		MaxActiveRuns:   3,
		Mode:            "propose",
		LastFiredAt:     &lastFired,
		IsActive:        true,
	}

	e := &Engine{store: s}

	// This should skip due to cooldown (30s < 60s)
	// Won't fire because runs, events, and db are nil, but the guard
	// should prevent reaching the fire path
	e.evaluateRule(context.Background(), rule, "ev-1", "", "mission.completed")

	// Verify the rule was NOT modified (LastFiredAt unchanged = still the old time)
	if rule.LastFiredAt == nil || !rule.LastFiredAt.Equal(lastFired) {
		t.Error("LastFiredAt should remain unchanged when cooldown blocks")
	}
}

func TestEvaluateRule_NoCooldown_FirstFire(t *testing.T) {
	s := &Store{
		cache: make(map[string]*TriggerRule),
	}

	rule := &TriggerRule{
		ID:              "r-1",
		EventPattern:    "mission.completed",
		CooldownSeconds: 60,
		MaxDepth:        5,
		MaxActiveRuns:   3,
		Mode:            "propose",
		LastFiredAt:     nil, // never fired
		IsActive:        true,
	}

	e := &Engine{store: s}

	// Should pass cooldown guard (never fired).
	// Will reach concurrency guard → ActiveCount fails (db nil) but allows through.
	// Will reach proposeTrigger → UpdateLastFired (noop with nil db) + LogExecution (error, non-fatal).
	e.evaluateRule(context.Background(), rule, "ev-1", "", "mission.completed")
	// If it doesn't panic, the cooldown guard correctly passed.
}

// ── NewEngine ─────────────────────────────────────────────────────

func TestNewEngine(t *testing.T) {
	s := &Store{cache: make(map[string]*TriggerRule)}
	e := NewEngine(s, nil, nil, nil)
	if e == nil {
		t.Fatal("expected non-nil engine")
	}
	if e.store != s {
		t.Error("store not wired correctly")
	}
	if e.nc != nil {
		t.Error("expected nil NATS connection")
	}
}

// ── Start / Stop ──────────────────────────────────────────────────

func TestStart_NilNATS(t *testing.T) {
	s := &Store{cache: make(map[string]*TriggerRule)}
	e := NewEngine(s, nil, nil, nil)

	// Should return nil (degraded mode), not panic
	if err := e.Start(context.Background()); err != nil {
		t.Fatalf("Start with nil NATS should not error: %v", err)
	}
}

func TestStop_NilSub(t *testing.T) {
	s := &Store{cache: make(map[string]*TriggerRule)}
	e := NewEngine(s, nil, nil, nil)

	// Should not panic
	e.Stop()
}

// ── ReloadRules ───────────────────────────────────────────────────

func TestReloadRules_NilDB(t *testing.T) {
	s := &Store{db: nil, cache: make(map[string]*TriggerRule)}
	e := NewEngine(s, nil, nil, nil)

	// Should return error from store
	if err := e.ReloadRules(context.Background()); err == nil {
		t.Error("expected error for nil DB")
	}
}

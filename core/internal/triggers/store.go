// Package triggers implements the V7 Trigger Engine — declarative IF/THEN rules
// evaluated on event ingest with cooldown, recursion, and concurrency guards.
package triggers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

// TriggerRule is the DB + API representation of a trigger rule.
type TriggerRule struct {
	ID                      string          `json:"id"`
	TenantID                string          `json:"tenant_id"`
	Name                    string          `json:"name"`
	Description             string          `json:"description,omitempty"`
	TriggerKind             string          `json:"trigger_kind"`      // "event" | "schedule"
	EventPattern            string          `json:"event_pattern"`     // e.g. "mission.completed"
	Condition               json.RawMessage `json:"condition"`         // optional payload filter
	TargetMissionID         string          `json:"target_mission_id"` // mission to launch
	Mode                    string          `json:"mode"`              // "propose" | "auto_execute"
	CooldownSeconds         int             `json:"cooldown_seconds"`
	ScheduleIntervalSeconds int             `json:"schedule_interval_seconds,omitempty"`
	NextRunAt               *time.Time      `json:"next_run_at,omitempty"`
	ProofExpectations       string          `json:"proof_expectations,omitempty"`
	RecoveryBehavior        string          `json:"recovery_behavior,omitempty"`
	MaxDepth                int             `json:"max_depth"`       // recursion guard
	MaxActiveRuns           int             `json:"max_active_runs"` // concurrency guard
	IsActive                bool            `json:"is_active"`
	LastFiredAt             *time.Time      `json:"last_fired_at,omitempty"`
	CreatedAt               time.Time       `json:"created_at"`
	UpdatedAt               time.Time       `json:"updated_at"`
}

// TriggerExecution is the audit record for a single evaluation.
type TriggerExecution struct {
	ID             string          `json:"id"`
	RuleID         string          `json:"rule_id"`
	EventID        string          `json:"event_id"`
	RunID          string          `json:"run_id,omitempty"`
	Status         string          `json:"status"` // "fired" | "skipped" | "proposed"
	SkipReason     string          `json:"skip_reason,omitempty"`
	HandoffKey     string          `json:"handoff_key,omitempty"`
	IntentProofID  string          `json:"intent_proof_id,omitempty"`
	ContractID     string          `json:"contract_id,omitempty"`
	ProposalStatus string          `json:"proposal_status,omitempty"`
	HandoffPayload json.RawMessage `json:"handoff_payload,omitempty"`
	ExecutedAt     time.Time       `json:"executed_at"`
}

// Store manages trigger rules in the database with an in-memory cache for
// fast pattern matching during event evaluation.
type Store struct {
	db    *sql.DB
	cache map[string]*TriggerRule // id → rule
	mu    sync.RWMutex
}

// NewStore creates a trigger rule store backed by the shared DB.
func NewStore(db *sql.DB) *Store {
	return &Store{
		db:    db,
		cache: make(map[string]*TriggerRule),
	}
}

// MatchingRules returns all cached rules whose event_pattern matches the given event type.
func (s *Store) MatchingRules(eventType string) []*TriggerRule {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var matches []*TriggerRule
	for _, r := range s.cache {
		if r.EventPattern == eventType && r.IsActive {
			matches = append(matches, r)
		}
	}
	return matches
}

// Create inserts a new trigger rule and adds it to the cache if active.
func (s *Store) Create(ctx context.Context, r *TriggerRule) error {
	if s.db == nil {
		return fmt.Errorf("triggers: database not available")
	}
	if r.Name == "" {
		return fmt.Errorf("triggers: name is required")
	}
	if r.TargetMissionID == "" {
		return fmt.Errorf("triggers: target_mission_id is required")
	}
	normalizeRuleDefaults(r)
	if !isValidTriggerKind(r.TriggerKind) {
		return fmt.Errorf("triggers: trigger_kind must be event or schedule")
	}
	if r.TriggerKind != "schedule" && r.EventPattern == "" {
		return fmt.Errorf("triggers: event_pattern is required")
	}
	if r.TriggerKind == "schedule" && r.ScheduleIntervalSeconds <= 0 {
		return fmt.Errorf("triggers: schedule_interval_seconds is required")
	}

	r.ID = uuid.New().String()

	err := s.db.QueryRowContext(ctx, `
		INSERT INTO trigger_rules
		    (id, tenant_id, name, description, event_pattern, condition,
		     target_mission_id, mode, cooldown_seconds, max_depth, max_active_runs, is_active,
		     trigger_kind, schedule_interval_seconds, next_run_at, proof_expectations, recovery_behavior)
		VALUES ($1, 'default', $2, NULLIF($3,''), $4, $5, $6, $7, $8, $9, $10, $11,
		        $12, NULLIF($13,0), $14, NULLIF($15,''), NULLIF($16,''))
		RETURNING created_at, updated_at`,
		r.ID, r.Name, r.Description, r.EventPattern, []byte(r.Condition),
		r.TargetMissionID, r.Mode, r.CooldownSeconds, r.MaxDepth, r.MaxActiveRuns, r.IsActive,
		r.TriggerKind, r.ScheduleIntervalSeconds, r.NextRunAt, r.ProofExpectations, r.RecoveryBehavior,
	).Scan(&r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		return fmt.Errorf("triggers: insert failed: %w", err)
	}

	r.TenantID = "default"

	// Update cache
	if r.IsActive {
		s.mu.Lock()
		s.cache[r.ID] = r
		s.mu.Unlock()
	}

	log.Printf("[triggers] created rule %s (%s → %s)", r.ID, r.EventPattern, r.TargetMissionID)
	return nil
}

// Update modifies an existing rule and refreshes the cache.
func (s *Store) Update(ctx context.Context, r *TriggerRule) error {
	if s.db == nil {
		return fmt.Errorf("triggers: database not available")
	}
	if len(r.Condition) == 0 {
		r.Condition = json.RawMessage("{}")
	}
	normalizeRuleDefaults(r)
	if !isValidTriggerKind(r.TriggerKind) {
		return fmt.Errorf("triggers: trigger_kind must be event or schedule")
	}
	if r.TriggerKind != "schedule" && r.EventPattern == "" {
		return fmt.Errorf("triggers: event_pattern is required")
	}
	if r.TriggerKind == "schedule" && r.ScheduleIntervalSeconds <= 0 {
		return fmt.Errorf("triggers: schedule_interval_seconds is required")
	}

	res, err := s.db.ExecContext(ctx, `
		UPDATE trigger_rules
		SET name=$1, description=NULLIF($2,''), event_pattern=$3, condition=$4,
		    target_mission_id=$5, mode=$6, cooldown_seconds=$7, max_depth=$8,
		    max_active_runs=$9, is_active=$10, trigger_kind=$11,
		    schedule_interval_seconds=NULLIF($12,0), next_run_at=$13,
		    proof_expectations=NULLIF($14,''), recovery_behavior=NULLIF($15,''), updated_at=NOW()
		WHERE id=$16 AND tenant_id='default'`,
		r.Name, r.Description, r.EventPattern, []byte(r.Condition),
		r.TargetMissionID, r.Mode, r.CooldownSeconds, r.MaxDepth,
		r.MaxActiveRuns, r.IsActive, r.TriggerKind, r.ScheduleIntervalSeconds, r.NextRunAt,
		r.ProofExpectations, r.RecoveryBehavior, r.ID)
	if err != nil {
		return fmt.Errorf("triggers: update failed: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("triggers: rule not found: %s", r.ID)
	}

	// Refresh cache
	s.mu.Lock()
	if r.IsActive {
		s.cache[r.ID] = r
	} else {
		delete(s.cache, r.ID)
	}
	s.mu.Unlock()

	return nil
}

func isValidTriggerKind(kind string) bool {
	return kind == "event" || kind == "schedule"
}

func normalizeRuleDefaults(r *TriggerRule) {
	if r.TriggerKind == "" {
		r.TriggerKind = "event"
	}
	if r.TriggerKind == "schedule" {
		r.Mode = "propose"
		if r.EventPattern == "" {
			r.EventPattern = "scheduler.due"
		}
		if r.NextRunAt == nil && r.ScheduleIntervalSeconds > 0 {
			next := time.Now().Add(time.Duration(r.ScheduleIntervalSeconds) * time.Second)
			r.NextRunAt = &next
		}
	}
	if r.Mode == "" {
		r.Mode = "propose"
	}
	if r.CooldownSeconds <= 0 {
		r.CooldownSeconds = 60
	}
	if r.MaxDepth <= 0 {
		r.MaxDepth = 5
	}
	if r.MaxActiveRuns <= 0 {
		r.MaxActiveRuns = 3
	}
	if len(r.Condition) == 0 {
		r.Condition = json.RawMessage("{}")
	}
}

// Delete removes a rule from the database and cache.
func (s *Store) Delete(ctx context.Context, id string) error {
	if s.db == nil {
		return fmt.Errorf("triggers: database not available")
	}

	res, err := s.db.ExecContext(ctx,
		"DELETE FROM trigger_rules WHERE id=$1 AND tenant_id='default'", id)
	if err != nil {
		return fmt.Errorf("triggers: delete failed: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("triggers: rule not found: %s", id)
	}

	s.mu.Lock()
	delete(s.cache, id)
	s.mu.Unlock()

	log.Printf("[triggers] deleted rule %s", id)
	return nil
}

// SetActive toggles a rule's is_active state and updates the cache.
func (s *Store) SetActive(ctx context.Context, id string, active bool) error {
	if s.db == nil {
		return fmt.Errorf("triggers: database not available")
	}

	res, err := s.db.ExecContext(ctx,
		"UPDATE trigger_rules SET is_active=$1, updated_at=NOW() WHERE id=$2 AND tenant_id='default'",
		active, id)
	if err != nil {
		return fmt.Errorf("triggers: toggle failed: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("triggers: rule not found: %s", id)
	}

	// Refresh cache entry
	s.mu.Lock()
	if !active {
		delete(s.cache, id)
	} else {
		// Re-read from DB to populate cache
		s.mu.Unlock()
		rule, err := s.Get(ctx, id)
		if err == nil && rule != nil {
			s.mu.Lock()
			s.cache[id] = rule
			s.mu.Unlock()
		}
		return nil
	}
	s.mu.Unlock()
	return nil
}

// UpdateLastFired sets last_fired_at on a rule (both DB and cache).
func (s *Store) UpdateLastFired(ctx context.Context, id string, t time.Time) {
	if s.db == nil {
		return
	}
	s.db.ExecContext(ctx,
		"UPDATE trigger_rules SET last_fired_at=$1 WHERE id=$2", t, id)

	s.mu.Lock()
	if r, ok := s.cache[id]; ok {
		r.LastFiredAt = &t
	}
	s.mu.Unlock()
}

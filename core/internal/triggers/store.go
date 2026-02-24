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
	ID               string          `json:"id"`
	TenantID         string          `json:"tenant_id"`
	Name             string          `json:"name"`
	Description      string          `json:"description,omitempty"`
	EventPattern     string          `json:"event_pattern"`              // e.g. "mission.completed"
	Condition        json.RawMessage `json:"condition"`                  // optional payload filter
	TargetMissionID  string          `json:"target_mission_id"`          // mission to launch
	Mode             string          `json:"mode"`                       // "propose" | "auto_execute"
	CooldownSeconds  int             `json:"cooldown_seconds"`
	MaxDepth         int             `json:"max_depth"`                  // recursion guard
	MaxActiveRuns    int             `json:"max_active_runs"`            // concurrency guard
	IsActive         bool            `json:"is_active"`
	LastFiredAt      *time.Time      `json:"last_fired_at,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

// TriggerExecution is the audit record for a single evaluation.
type TriggerExecution struct {
	ID         string    `json:"id"`
	RuleID     string    `json:"rule_id"`
	EventID    string    `json:"event_id"`
	RunID      string    `json:"run_id,omitempty"`
	Status     string    `json:"status"`       // "fired" | "skipped" | "proposed"
	SkipReason string    `json:"skip_reason,omitempty"`
	ExecutedAt time.Time `json:"executed_at"`
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

// LoadActiveRules populates the in-memory cache from the database.
// Safe to call multiple times (full refresh).
func (s *Store) LoadActiveRules(ctx context.Context) error {
	if s.db == nil {
		return fmt.Errorf("triggers: database not available")
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, tenant_id, name, COALESCE(description,''), event_pattern,
		       condition, target_mission_id, mode, cooldown_seconds, max_depth,
		       max_active_runs, is_active, last_fired_at, created_at, updated_at
		FROM trigger_rules
		WHERE tenant_id = 'default' AND is_active = true`)
	if err != nil {
		return fmt.Errorf("triggers: load query failed: %w", err)
	}
	defer rows.Close()

	fresh := make(map[string]*TriggerRule)
	for rows.Next() {
		r := &TriggerRule{}
		var lastFired sql.NullTime
		if err := rows.Scan(
			&r.ID, &r.TenantID, &r.Name, &r.Description, &r.EventPattern,
			&r.Condition, &r.TargetMissionID, &r.Mode, &r.CooldownSeconds,
			&r.MaxDepth, &r.MaxActiveRuns, &r.IsActive, &lastFired,
			&r.CreatedAt, &r.UpdatedAt,
		); err != nil {
			log.Printf("[triggers] scan error: %v", err)
			continue
		}
		if lastFired.Valid {
			t := lastFired.Time
			r.LastFiredAt = &t
		}
		fresh[r.ID] = r
	}

	s.mu.Lock()
	s.cache = fresh
	s.mu.Unlock()

	log.Printf("[triggers] loaded %d active rules into cache", len(fresh))
	return rows.Err()
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

// ListAll returns all trigger rules (active and inactive) for the tenant.
func (s *Store) ListAll(ctx context.Context) ([]TriggerRule, error) {
	if s.db == nil {
		return nil, fmt.Errorf("triggers: database not available")
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, tenant_id, name, COALESCE(description,''), event_pattern,
		       condition, target_mission_id, mode, cooldown_seconds, max_depth,
		       max_active_runs, is_active, last_fired_at, created_at, updated_at
		FROM trigger_rules
		WHERE tenant_id = 'default'
		ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("triggers: list query failed: %w", err)
	}
	defer rows.Close()

	rules := make([]TriggerRule, 0)
	for rows.Next() {
		var r TriggerRule
		var lastFired sql.NullTime
		if err := rows.Scan(
			&r.ID, &r.TenantID, &r.Name, &r.Description, &r.EventPattern,
			&r.Condition, &r.TargetMissionID, &r.Mode, &r.CooldownSeconds,
			&r.MaxDepth, &r.MaxActiveRuns, &r.IsActive, &lastFired,
			&r.CreatedAt, &r.UpdatedAt,
		); err != nil {
			log.Printf("[triggers] scan error: %v", err)
			continue
		}
		if lastFired.Valid {
			t := lastFired.Time
			r.LastFiredAt = &t
		}
		rules = append(rules, r)
	}
	return rules, rows.Err()
}

// Get retrieves a single rule by ID.
func (s *Store) Get(ctx context.Context, id string) (*TriggerRule, error) {
	if s.db == nil {
		return nil, fmt.Errorf("triggers: database not available")
	}

	var r TriggerRule
	var lastFired sql.NullTime
	err := s.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, name, COALESCE(description,''), event_pattern,
		       condition, target_mission_id, mode, cooldown_seconds, max_depth,
		       max_active_runs, is_active, last_fired_at, created_at, updated_at
		FROM trigger_rules
		WHERE id = $1 AND tenant_id = 'default'`, id).
		Scan(&r.ID, &r.TenantID, &r.Name, &r.Description, &r.EventPattern,
			&r.Condition, &r.TargetMissionID, &r.Mode, &r.CooldownSeconds,
			&r.MaxDepth, &r.MaxActiveRuns, &r.IsActive, &lastFired,
			&r.CreatedAt, &r.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("triggers: get failed: %w", err)
	}
	if lastFired.Valid {
		t := lastFired.Time
		r.LastFiredAt = &t
	}
	return &r, nil
}

// Create inserts a new trigger rule and adds it to the cache if active.
func (s *Store) Create(ctx context.Context, r *TriggerRule) error {
	if s.db == nil {
		return fmt.Errorf("triggers: database not available")
	}
	if r.Name == "" {
		return fmt.Errorf("triggers: name is required")
	}
	if r.EventPattern == "" {
		return fmt.Errorf("triggers: event_pattern is required")
	}
	if r.TargetMissionID == "" {
		return fmt.Errorf("triggers: target_mission_id is required")
	}

	r.ID = uuid.New().String()
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

	err := s.db.QueryRowContext(ctx, `
		INSERT INTO trigger_rules
		    (id, tenant_id, name, description, event_pattern, condition,
		     target_mission_id, mode, cooldown_seconds, max_depth, max_active_runs, is_active)
		VALUES ($1, 'default', $2, NULLIF($3,''), $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING created_at, updated_at`,
		r.ID, r.Name, r.Description, r.EventPattern, []byte(r.Condition),
		r.TargetMissionID, r.Mode, r.CooldownSeconds, r.MaxDepth, r.MaxActiveRuns, r.IsActive,
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

	res, err := s.db.ExecContext(ctx, `
		UPDATE trigger_rules
		SET name=$1, description=NULLIF($2,''), event_pattern=$3, condition=$4,
		    target_mission_id=$5, mode=$6, cooldown_seconds=$7, max_depth=$8,
		    max_active_runs=$9, is_active=$10, updated_at=NOW()
		WHERE id=$11 AND tenant_id='default'`,
		r.Name, r.Description, r.EventPattern, []byte(r.Condition),
		r.TargetMissionID, r.Mode, r.CooldownSeconds, r.MaxDepth,
		r.MaxActiveRuns, r.IsActive, r.ID)
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

// LogExecution records a trigger evaluation in the audit table.
func (s *Store) LogExecution(ctx context.Context, exec *TriggerExecution) error {
	if s.db == nil {
		return fmt.Errorf("triggers: database not available")
	}

	exec.ID = uuid.New().String()
	exec.ExecutedAt = time.Now()

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO trigger_executions (id, rule_id, event_id, run_id, status, skip_reason, executed_at)
		VALUES ($1, $2, $3, NULLIF($4,''), $5, NULLIF($6,''), $7)`,
		exec.ID, exec.RuleID, exec.EventID, exec.RunID, exec.Status, exec.SkipReason, exec.ExecutedAt)
	if err != nil {
		return fmt.Errorf("triggers: log execution failed: %w", err)
	}
	return nil
}

// ListExecutions returns recent executions for a rule, newest first.
func (s *Store) ListExecutions(ctx context.Context, ruleID string, limit int) ([]TriggerExecution, error) {
	if s.db == nil {
		return nil, fmt.Errorf("triggers: database not available")
	}
	if limit <= 0 {
		limit = 20
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, rule_id, event_id, COALESCE(run_id,''), status, COALESCE(skip_reason,''), executed_at
		FROM trigger_executions
		WHERE rule_id = $1
		ORDER BY executed_at DESC
		LIMIT $2`, ruleID, limit)
	if err != nil {
		return nil, fmt.Errorf("triggers: list executions failed: %w", err)
	}
	defer rows.Close()

	execs := make([]TriggerExecution, 0)
	for rows.Next() {
		var e TriggerExecution
		if err := rows.Scan(&e.ID, &e.RuleID, &e.EventID, &e.RunID, &e.Status, &e.SkipReason, &e.ExecutedAt); err != nil {
			continue
		}
		execs = append(execs, e)
	}
	return execs, rows.Err()
}

// ActiveCount returns the count of active (running) runs for a given mission.
func (s *Store) ActiveCount(ctx context.Context, missionID string) (int, error) {
	if s.db == nil {
		return 0, fmt.Errorf("triggers: database not available")
	}

	var count int
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM mission_runs
		WHERE mission_id = $1 AND status IN ('pending','running')
		AND tenant_id = 'default'`, missionID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("triggers: active count query failed: %w", err)
	}
	return count, nil
}

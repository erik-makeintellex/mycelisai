package triggers

import (
	"context"
	"database/sql"
	"fmt"
	"log"
)

// LoadActiveRules populates the in-memory cache from the database.
func (s *Store) LoadActiveRules(ctx context.Context) error {
	if s.db == nil {
		return fmt.Errorf("triggers: database not available")
	}

	rows, err := s.db.QueryContext(ctx, triggerRuleSelectSQL+`
		WHERE tenant_id = 'default' AND is_active = true`)
	if err != nil {
		return fmt.Errorf("triggers: load query failed: %w", err)
	}
	defer rows.Close()

	fresh := make(map[string]*TriggerRule)
	for rows.Next() {
		r, err := scanTriggerRule(rows)
		if err != nil {
			log.Printf("[triggers] scan error: %v", err)
			continue
		}
		fresh[r.ID] = r
	}

	s.mu.Lock()
	s.cache = fresh
	s.mu.Unlock()

	log.Printf("[triggers] loaded %d active rules into cache", len(fresh))
	return rows.Err()
}

func (s *Store) ListAll(ctx context.Context) ([]TriggerRule, error) {
	if s.db == nil {
		return nil, fmt.Errorf("triggers: database not available")
	}

	rows, err := s.db.QueryContext(ctx, triggerRuleSelectSQL+`
		WHERE tenant_id = 'default'
		ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("triggers: list query failed: %w", err)
	}
	defer rows.Close()

	rules := make([]TriggerRule, 0)
	for rows.Next() {
		r, err := scanTriggerRule(rows)
		if err != nil {
			log.Printf("[triggers] scan error: %v", err)
			continue
		}
		rules = append(rules, *r)
	}
	return rules, rows.Err()
}

func (s *Store) Get(ctx context.Context, id string) (*TriggerRule, error) {
	if s.db == nil {
		return nil, fmt.Errorf("triggers: database not available")
	}

	var r TriggerRule
	var lastFired sql.NullTime
	err := s.db.QueryRowContext(ctx, triggerRuleSelectSQL+`
		WHERE id = $1 AND tenant_id = 'default'`, id).
		Scan(triggerRuleScanTargets(&r, &lastFired)...)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("triggers: get failed: %w", err)
	}
	applyLastFired(&r, lastFired)
	return &r, nil
}

const triggerRuleSelectSQL = `
	SELECT id, tenant_id, name, COALESCE(description,''), event_pattern,
	       condition, target_mission_id, mode, cooldown_seconds, max_depth,
	       max_active_runs, is_active, last_fired_at, created_at, updated_at
	FROM trigger_rules`

type triggerRuleScanner interface {
	Scan(dest ...any) error
}

func scanTriggerRule(scanner triggerRuleScanner) (*TriggerRule, error) {
	var r TriggerRule
	var lastFired sql.NullTime
	if err := scanner.Scan(triggerRuleScanTargets(&r, &lastFired)...); err != nil {
		return nil, err
	}
	applyLastFired(&r, lastFired)
	return &r, nil
}

func triggerRuleScanTargets(r *TriggerRule, lastFired *sql.NullTime) []any {
	return []any{
		&r.ID, &r.TenantID, &r.Name, &r.Description, &r.EventPattern,
		&r.Condition, &r.TargetMissionID, &r.Mode, &r.CooldownSeconds,
		&r.MaxDepth, &r.MaxActiveRuns, &r.IsActive, lastFired,
		&r.CreatedAt, &r.UpdatedAt,
	}
}

func applyLastFired(r *TriggerRule, lastFired sql.NullTime) {
	if lastFired.Valid {
		t := lastFired.Time
		r.LastFiredAt = &t
	}
}

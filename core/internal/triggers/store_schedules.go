package triggers

import (
	"context"
	"fmt"
	"time"
)

func (s *Store) ListDueScheduleRules(ctx context.Context, now time.Time, limit int) ([]TriggerRule, error) {
	if s.db == nil {
		return nil, fmt.Errorf("triggers: database not available")
	}
	if limit <= 0 {
		limit = 20
	}

	rows, err := s.db.QueryContext(ctx, triggerRuleSelectSQL+`
		WHERE tenant_id = 'default'
		  AND is_active = true
		  AND trigger_kind = 'schedule'
		  AND next_run_at IS NOT NULL
		  AND next_run_at <= $1
		ORDER BY next_run_at ASC
		LIMIT $2`, now, limit)
	if err != nil {
		return nil, fmt.Errorf("triggers: due schedule query failed: %w", err)
	}
	defer rows.Close()

	rules := make([]TriggerRule, 0)
	for rows.Next() {
		r, err := scanTriggerRule(rows)
		if err != nil {
			return nil, err
		}
		rules = append(rules, *r)
	}
	return rules, rows.Err()
}

func (s *Store) MarkScheduleProposed(ctx context.Context, id string, proposedAt time.Time, nextRunAt time.Time) error {
	if s.db == nil {
		return fmt.Errorf("triggers: database not available")
	}
	_, err := s.db.ExecContext(ctx, `
		UPDATE trigger_rules
		SET last_fired_at=$1, next_run_at=$2, updated_at=NOW()
		WHERE id=$3 AND tenant_id='default'`, proposedAt, nextRunAt, id)
	if err != nil {
		return fmt.Errorf("triggers: mark schedule proposed failed: %w", err)
	}

	s.mu.Lock()
	if r, ok := s.cache[id]; ok {
		r.LastFiredAt = &proposedAt
		r.NextRunAt = &nextRunAt
	}
	s.mu.Unlock()
	return nil
}

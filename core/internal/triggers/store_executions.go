package triggers

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

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

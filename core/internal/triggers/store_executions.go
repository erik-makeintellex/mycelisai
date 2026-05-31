package triggers

import (
	"context"
	"database/sql"
	"encoding/json"
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

	if exec.ProposalStatus == "" {
		exec.ProposalStatus = "recorded"
		if exec.HandoffKey != "" {
			exec.ProposalStatus = "awaiting_approval"
		}
	}
	if len(exec.HandoffPayload) == 0 {
		exec.HandoffPayload = json.RawMessage(`{}`)
	}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO trigger_executions
			(id, rule_id, event_id, run_id, status, skip_reason, handoff_key,
			 intent_proof_id, contract_id, proposal_status, handoff_payload, executed_at)
		VALUES ($1, $2, $3, NULLIF($4,''), $5, NULLIF($6,''), NULLIF($7,''),
		        NULLIF($8,'')::uuid, NULLIF($9,'')::uuid, $10, $11, $12)
		ON CONFLICT DO NOTHING`,
		exec.ID, exec.RuleID, exec.EventID, exec.RunID, exec.Status, exec.SkipReason, exec.HandoffKey,
		exec.IntentProofID, exec.ContractID, exec.ProposalStatus, []byte(exec.HandoffPayload), exec.ExecutedAt)
	if err != nil {
		return fmt.Errorf("triggers: log execution failed: %w", err)
	}
	return nil
}

func (s *Store) GetExecutionByHandoffKey(ctx context.Context, ruleID, handoffKey string) (*TriggerExecution, error) {
	if s.db == nil {
		return nil, fmt.Errorf("triggers: database not available")
	}
	if handoffKey == "" {
		return nil, nil
	}
	row := s.db.QueryRowContext(ctx, triggerExecutionSelectSQL+`
		WHERE rule_id = $1 AND handoff_key = $2
		LIMIT 1`, ruleID, handoffKey)
	exec, err := scanTriggerExecution(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("triggers: get handoff execution failed: %w", err)
	}
	return exec, nil
}

func (s *Store) UpdateExecutionHandoffRefs(ctx context.Context, execID, intentProofID, contractID, proposalStatus string, payload json.RawMessage) error {
	if s.db == nil {
		return fmt.Errorf("triggers: database not available")
	}
	if proposalStatus == "" {
		proposalStatus = "awaiting_approval"
	}
	if len(payload) == 0 {
		payload = json.RawMessage(`{}`)
	}
	_, err := s.db.ExecContext(ctx, `
		UPDATE trigger_executions
		SET intent_proof_id=NULLIF($1,'')::uuid,
		    contract_id=NULLIF($2,'')::uuid,
		    proposal_status=$3,
		    handoff_payload=$4
		WHERE id=$5`,
		intentProofID, contractID, proposalStatus, []byte(payload), execID)
	if err != nil {
		return fmt.Errorf("triggers: update handoff refs failed: %w", err)
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

	rows, err := s.db.QueryContext(ctx, triggerExecutionSelectSQL+`
		WHERE rule_id = $1
		ORDER BY executed_at DESC
		LIMIT $2`, ruleID, limit)
	if err != nil {
		return nil, fmt.Errorf("triggers: list executions failed: %w", err)
	}
	defer rows.Close()

	execs := make([]TriggerExecution, 0)
	for rows.Next() {
		e, err := scanTriggerExecution(rows)
		if err != nil {
			continue
		}
		execs = append(execs, *e)
	}
	return execs, rows.Err()
}

const triggerExecutionSelectSQL = `
		SELECT id, rule_id, event_id, COALESCE(run_id,''), status, COALESCE(skip_reason,''),
		       COALESCE(handoff_key,''), COALESCE(intent_proof_id::text,''), COALESCE(contract_id::text,''),
		       COALESCE(proposal_status,''), COALESCE(handoff_payload, '{}'::jsonb), executed_at
		FROM trigger_executions`

type triggerExecutionScanner interface {
	Scan(dest ...any) error
}

func scanTriggerExecution(scanner triggerExecutionScanner) (*TriggerExecution, error) {
	var e TriggerExecution
	var payload []byte
	if err := scanner.Scan(
		&e.ID, &e.RuleID, &e.EventID, &e.RunID, &e.Status, &e.SkipReason,
		&e.HandoffKey, &e.IntentProofID, &e.ContractID, &e.ProposalStatus, &payload, &e.ExecutedAt,
	); err != nil {
		return nil, err
	}
	if len(payload) > 0 && string(payload) != "{}" {
		e.HandoffPayload = json.RawMessage(payload)
	}
	return &e, nil
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

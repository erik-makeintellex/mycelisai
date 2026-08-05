package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/runs"
	"github.com/mycelis/core/pkg/protocol"
)

const teamWorkRecoveryPollInterval = 2 * time.Second

// StartTeamWorkRecoveryReconciler makes accepted asynchronous work recoverable
// even when NATS is disconnected or Core restarts after the handoff.
func StartTeamWorkRecoveryReconciler(ctx context.Context, s *AdminServer) error {
	if s == nil || s.getDB() == nil {
		return errors.New("team work recovery reconciler requires database")
	}
	go s.runTeamWorkRecoveryReconciler(ctx)
	return nil
}

func (s *AdminServer) runTeamWorkRecoveryReconciler(ctx context.Context) {
	ticker := time.NewTicker(teamWorkRecoveryPollInterval)
	defer ticker.Stop()
	for {
		if _, err := s.reconcileOneOverdueTeamWork(ctx); err != nil && ctx.Err() == nil {
			log.Printf("component=team_work_recovery action=reconcile status=failed error=%q", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (s *AdminServer) reconcileOneOverdueTeamWork(ctx context.Context) (bool, error) {
	tx, err := s.getDB().BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	var workItemID, teamID string
	err = tx.QueryRowContext(ctx, `
		SELECT id::text, team_id
		FROM team_work_items
		WHERE tenant_id='default'
		  AND recovery_deadline_at <= NOW()
		  AND state IN ('queued','running','reviewing')
		ORDER BY recovery_deadline_at ASC
		FOR UPDATE SKIP LOCKED
		LIMIT 1`).Scan(&workItemID, &teamID)
	if errors.Is(err, sql.ErrNoRows) {
		return false, tx.Commit()
	}
	if err != nil {
		return false, err
	}
	item, err := lockTeamWorkItemTx(ctx, tx, teamID, workItemID)
	if err != nil {
		return false, err
	}
	item.State = protocol.TeamWorkStateDegraded
	item.NeedsOperator = true
	item.DegradationState = "team_work_recovery_deadline_exceeded"
	item.RecoveryOptions = []string{
		"Ask Soma to retry this work with the same team after checking team and capability availability.",
		"Steer the work with updated guidance before retrying.",
		"Archive the work if it is no longer needed.",
	}
	event := overdueTeamWorkStatusEvent(item)
	if err := s.insertTeamStatusEventExec(ctx, tx, &event); err != nil {
		return false, err
	}
	if err := s.updateTeamWorkItemLastEventExec(ctx, tx, &item, event); err != nil {
		return false, err
	}
	interaction := overdueTeamWorkInteraction(item)
	if err := s.insertTeamInteractionExec(ctx, tx, &interaction); err != nil {
		return false, err
	}
	if err := s.markRunDegradedWhenSettledTx(ctx, tx, item); err != nil {
		return false, err
	}
	if err := tx.Commit(); err != nil {
		return false, err
	}
	s.broadcastTeamWorkResultThreadEvent(item, event)
	return true, nil
}

func overdueTeamWorkStatusEvent(item protocol.TeamWorkItem) protocol.TeamStatusEvent {
	return protocol.NormalizeTeamStatusEvent(protocol.TeamStatusEvent{
		EventID: uuid.NewString(), TeamID: item.TeamID, WorkItemID: item.WorkItemID,
		RunID: item.RunID, IntentProofID: item.IntentProofID, ContractID: item.ContractID,
		State: item.State, Headline: "Team work needs recovery",
		Details:           "No terminal team result arrived before the durable recovery deadline.",
		ConfidencePosture: "operator_attention", BlockedBy: []string{"recovery_deadline_exceeded"},
		NextAction: item.RecoveryOptions[0], ExpectedOutputs: item.ExpectedOutputs,
		ExpectedProof: item.ExpectedProof, ExecutionMode: item.ExecutionMode, WorkIntent: item.WorkIntent,
		SourceKind: string(protocol.SourceKindSystem), SourceChannel: "team-work.recovery-reconciler",
		PayloadKind: string(protocol.PayloadKindError), Version: "v1",
	})
}

func overdueTeamWorkInteraction(item protocol.TeamWorkItem) protocol.TeamInteraction {
	return protocol.NormalizeTeamInteraction(protocol.TeamInteraction{
		InteractionID: uuid.NewString(), TeamID: item.TeamID, WorkItemID: item.WorkItemID,
		RunID: item.RunID, IntentProofID: item.IntentProofID, ContractID: item.ContractID,
		SourceKind: string(protocol.SourceKindSystem), SourceChannel: "team-work.recovery-reconciler",
		ActorRef: "Soma", Verb: "degraded", Summary: "The durable recovery deadline elapsed without a terminal result.",
		PayloadKind: string(protocol.PayloadKindError),
		Payload:     map[string]any{"degradation_state": item.DegradationState, "recovery_options": item.RecoveryOptions}, Version: "v1",
	})
}

func (s *AdminServer) markRunDegradedWhenSettledTx(ctx context.Context, tx *sql.Tx, item protocol.TeamWorkItem) error {
	if strings.TrimSpace(item.RunID) == "" {
		return nil
	}
	var unresolved, degraded int
	if err := tx.QueryRowContext(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE execution_shape <> 'create_team' AND state NOT IN ('output_ready','degraded','needs_operator','archived')),
			COUNT(*) FILTER (WHERE execution_shape <> 'create_team' AND state IN ('degraded','needs_operator'))
		FROM team_work_items
		WHERE run_id=$1`, item.RunID).Scan(&unresolved, &degraded); err != nil {
		return fmt.Errorf("derive linked run recovery state: %w", err)
	}
	if unresolved > 0 || degraded == 0 {
		return nil
	}
	result, err := tx.ExecContext(ctx, `
		UPDATE mission_runs
		SET status=$1, completed_at=GREATEST(NOW(), started_at)
		WHERE id=$2 AND status NOT IN ($1,$3,$4)`, runs.StatusDegraded, item.RunID, runs.StatusCompleted, runs.StatusFailed)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return nil
	}
	payload := fmt.Sprintf(`{"proof_id":%q,"execution_state":"degraded","run_status":%q,"reason":"linked team work requires recovery"}`, item.IntentProofID, runs.StatusDegraded)
	_, err = tx.ExecContext(ctx, `
		INSERT INTO mission_events
		(id, run_id, tenant_id, event_type, severity, source_agent, source_team, payload, emitted_at)
		VALUES ($1,$2,'default',$3,$4,'soma',$5,$6::jsonb,NOW())`,
		uuid.NewString(), item.RunID, string(protocol.EventMissionDegraded), string(protocol.SeverityWarn), item.TeamID, payload)
	return err
}

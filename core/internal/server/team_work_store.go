package server

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/mycelis/core/pkg/protocol"
)

func (s *AdminServer) listTeamWorkItemsDB(ctx context.Context, teamID string, limit int, includeArchived bool) ([]protocol.TeamWorkItem, error) {
	db := s.getDB()
	if db == nil {
		return nil, errors.New("database not available")
	}
	archivedFilter := ""
	if !includeArchived {
		archivedFilter = "AND state <> 'archived'"
	}
	rows, err := db.QueryContext(ctx, `
		SELECT id::text, team_id, COALESCE(run_id::text,''), COALESCE(intent_proof_id::text,''),
		       COALESCE(contract_id,''), COALESCE(proof_id,''), objective, scope, owner,
		       execution_shape, expected_outputs, expected_proof, capability_requirements,
		       governance_posture, state, COALESCE(last_event, 'null'::jsonb), needs_operator,
		       degradation_state, recovery_options, output_refs, proof_refs, audit_refs,
		       created_at, updated_at, version
		FROM team_work_items
		WHERE tenant_id='default' AND team_id=$1
		`+archivedFilter+`
		ORDER BY updated_at DESC
		LIMIT $2`, strings.TrimSpace(teamID), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]protocol.TeamWorkItem, 0)
	for rows.Next() {
		item, scanErr := scanTeamWorkItem(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (s *AdminServer) getTeamWorkItemDB(ctx context.Context, teamID, workItemID string) (protocol.TeamWorkItem, error) {
	db := s.getDB()
	if db == nil {
		return protocol.TeamWorkItem{}, errors.New("database not available")
	}
	return scanTeamWorkItem(db.QueryRowContext(ctx, `
		SELECT id::text, team_id, COALESCE(run_id::text,''), COALESCE(intent_proof_id::text,''),
		       COALESCE(contract_id,''), COALESCE(proof_id,''), objective, scope, owner,
		       execution_shape, expected_outputs, expected_proof, capability_requirements,
		       governance_posture, state, COALESCE(last_event, 'null'::jsonb), needs_operator,
		       degradation_state, recovery_options, output_refs, proof_refs, audit_refs,
		       created_at, updated_at, version
		FROM team_work_items
		WHERE tenant_id='default' AND team_id=$1 AND id=$2`,
		strings.TrimSpace(teamID), workItemID,
	))
}

func (s *AdminServer) insertTeamWorkItemDB(ctx context.Context, item *protocol.TeamWorkItem) error {
	db := s.getDB()
	if db == nil {
		return errors.New("database not available")
	}
	return s.insertTeamWorkItemExec(ctx, db, item)
}

func (s *AdminServer) insertTeamWorkItemExec(ctx context.Context, exec teamWorkSQLExecutor, item *protocol.TeamWorkItem) error {
	if exec == nil {
		return errors.New("database not available")
	}
	if strings.TrimSpace(item.WorkItemID) == "" {
		item.WorkItemID = uuid.NewString()
	}
	item.TargetRef = protocol.NormalizeTargetRef(item.TargetRef)
	if item.TargetRef == nil {
		item.TargetRef = protocol.TargetRefForTeamWork(*item)
	}
	if err := validateTeamWorkUUIDLinks(*item); err != nil {
		return err
	}
	return exec.QueryRowContext(ctx, `
		INSERT INTO team_work_items (
			id, tenant_id, team_id, run_id, intent_proof_id, contract_id, proof_id,
			objective, scope, owner, execution_shape, expected_outputs, expected_proof,
			capability_requirements, governance_posture, state, needs_operator,
			degradation_state, recovery_options, output_refs, proof_refs, audit_refs, version
		) VALUES (
			$1, 'default', $2, $3, $4, NULLIF($5,''), NULLIF($6,''),
			$7, $8, $9, $10, $11, $12,
			$13, $14, $15, $16,
			$17, $18, $19, $20, $21, $22
		)
		RETURNING created_at, updated_at`,
		item.WorkItemID, item.TeamID, nullableUUID(item.RunID), nullableUUID(item.IntentProofID),
		item.ContractID, item.ProofID, item.Objective, jsonArray(item.Scope), item.Owner,
		string(item.ExecutionShape), jsonArray(item.ExpectedOutputs), jsonArray(item.ExpectedProof),
		jsonArray(item.CapabilityRequirements), string(item.GovernancePosture), string(item.State),
		item.NeedsOperator, item.DegradationState, jsonArray(item.RecoveryOptions),
		jsonArray(item.OutputRefs), jsonArray(item.ProofRefs), jsonArray(item.AuditRefs), item.Version,
	).Scan(&item.CreatedAt, &item.UpdatedAt)
}

func (s *AdminServer) insertTeamStatusEventDB(ctx context.Context, event *protocol.TeamStatusEvent) error {
	db := s.getDB()
	if db == nil {
		return errors.New("database not available")
	}
	return s.insertTeamStatusEventExec(ctx, db, event)
}

func (s *AdminServer) insertTeamStatusEventExec(ctx context.Context, exec teamWorkSQLExecutor, event *protocol.TeamStatusEvent) error {
	if exec == nil {
		return errors.New("database not available")
	}
	if strings.TrimSpace(event.EventID) == "" {
		event.EventID = uuid.NewString()
	}
	if err := validateTeamStatusEventUUIDLinks(*event); err != nil {
		return err
	}
	if strings.TrimSpace(event.Version) == "" {
		event.Version = "v1"
	}
	event.TargetRef = protocol.NormalizeTargetRef(event.TargetRef)
	if event.TargetRef == nil {
		event.TargetRef = protocol.TargetRefForTeamStatusEvent(*event)
	}
	if err := exec.QueryRowContext(ctx, `
		INSERT INTO team_status_events (
			id, tenant_id, team_id, work_item_id, run_id, intent_proof_id, contract_id, proof_id,
			state, headline, details, confidence_posture, blocked_by, next_action,
			source_kind, source_channel, payload_kind, audit_refs, version
		) VALUES (
			$1, 'default', $2, $3, $4, $5, NULLIF($6,''), NULLIF($7,''),
			$8, $9, $10, $11, $12, $13,
			$14, $15, $16, $17, $18
		)
		RETURNING timestamp`,
		event.EventID, event.TeamID, event.WorkItemID, nullableUUID(event.RunID),
		nullableUUID(event.IntentProofID), event.ContractID, event.ProofID,
		string(event.State), event.Headline, event.Details, event.ConfidencePosture,
		jsonArray(event.BlockedBy), event.NextAction, event.SourceKind, event.SourceChannel,
		event.PayloadKind, jsonArray(event.AuditRefs), event.Version,
	).Scan(&event.Timestamp); err != nil {
		return err
	}
	return s.insertTeamWorkMissionEventExec(ctx, exec, event)
}

func (s *AdminServer) updateTeamWorkItemLastEventDB(ctx context.Context, item *protocol.TeamWorkItem, event protocol.TeamStatusEvent) error {
	db := s.getDB()
	if db == nil {
		return errors.New("database not available")
	}
	if err := s.updateTeamWorkItemLastEventExec(ctx, db, item, event); err != nil {
		return err
	}
	item.LastEvent = &event
	return nil
}

func (s *AdminServer) updateTeamWorkItemLastEventExec(ctx context.Context, exec teamWorkSQLExecutor, item *protocol.TeamWorkItem, event protocol.TeamStatusEvent) error {
	if exec == nil {
		return errors.New("database not available")
	}
	if item == nil {
		return errors.New("team work item is required")
	}
	item.TargetRef = protocol.NormalizeTargetRef(item.TargetRef)
	if item.TargetRef == nil {
		item.TargetRef = protocol.TargetRefForTeamWork(*item)
	}
	event.TargetRef = protocol.NormalizeTargetRef(event.TargetRef)
	if event.TargetRef == nil {
		event.TargetRef = protocol.TargetRefForTeamStatusEvent(event)
	}
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return err
	}
	_, err = exec.ExecContext(ctx, `
		UPDATE team_work_items
		SET state=$2,
		    last_event=$3,
		    needs_operator=$4,
		    degradation_state=$5,
		    recovery_options=$6,
		    output_refs=$7,
		    proof_refs=$8,
		    audit_refs=$9,
		    updated_at=NOW()
		WHERE id=$1 AND tenant_id='default'`,
		item.WorkItemID, string(item.State), eventJSON, item.NeedsOperator,
		item.DegradationState, jsonArray(item.RecoveryOptions), jsonArray(item.OutputRefs),
		jsonArray(item.ProofRefs), jsonArray(item.AuditRefs),
	)
	return err
}

func (s *AdminServer) listTeamStatusEventsDB(ctx context.Context, teamID, workItemID string, limit int) ([]protocol.TeamStatusEvent, error) {
	db := s.getDB()
	if db == nil {
		return nil, errors.New("database not available")
	}
	rows, err := db.QueryContext(ctx, `
		SELECT id::text, team_id, work_item_id::text, COALESCE(run_id::text,''),
		       COALESCE(intent_proof_id::text,''), COALESCE(contract_id,''), COALESCE(proof_id,''),
		       state, headline, details, confidence_posture, blocked_by, next_action,
		       source_kind, source_channel, payload_kind, audit_refs, timestamp, version
		FROM team_status_events
		WHERE tenant_id='default' AND team_id=$1 AND work_item_id=$2
		ORDER BY timestamp ASC
		LIMIT $3`, strings.TrimSpace(teamID), workItemID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]protocol.TeamStatusEvent, 0)
	for rows.Next() {
		item, scanErr := scanTeamStatusEvent(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

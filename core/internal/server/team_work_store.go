package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/mycelis/core/pkg/protocol"
)

func (s *AdminServer) listTeamWorkItemsDB(ctx context.Context, teamID string, limit int) ([]protocol.TeamWorkItem, error) {
	db := s.getDB()
	if db == nil {
		return nil, errors.New("database not available")
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

func (s *AdminServer) insertTeamWorkItemDB(ctx context.Context, item *protocol.TeamWorkItem) error {
	db := s.getDB()
	if db == nil {
		return errors.New("database not available")
	}
	if strings.TrimSpace(item.WorkItemID) == "" {
		item.WorkItemID = uuid.NewString()
	}
	if err := validateTeamWorkUUIDLinks(*item); err != nil {
		return err
	}
	return db.QueryRowContext(ctx, `
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

func (s *AdminServer) listTeamInteractionsDB(ctx context.Context, teamID, workItemID string, limit int) ([]protocol.TeamInteraction, error) {
	db := s.getDB()
	if db == nil {
		return nil, errors.New("database not available")
	}
	rows, err := db.QueryContext(ctx, `
		SELECT id::text, team_id, work_item_id::text, COALESCE(run_id::text,''),
		       COALESCE(intent_proof_id::text,''), COALESCE(contract_id,''), COALESCE(proof_id,''),
		       source_kind, source_channel, actor_ref, verb, summary, payload_kind,
		       payload_ref, COALESCE(payload, 'null'::jsonb), approval_ref, audit_refs, timestamp, version
		FROM team_interactions
		WHERE tenant_id='default' AND team_id=$1 AND work_item_id=$2
		ORDER BY timestamp ASC
		LIMIT $3`, strings.TrimSpace(teamID), workItemID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]protocol.TeamInteraction, 0)
	for rows.Next() {
		item, scanErr := scanTeamInteraction(rows)
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

func (s *AdminServer) insertTeamInteractionDB(ctx context.Context, item *protocol.TeamInteraction) error {
	db := s.getDB()
	if db == nil {
		return errors.New("database not available")
	}
	if strings.TrimSpace(item.InteractionID) == "" {
		item.InteractionID = uuid.NewString()
	}
	if err := validateTeamInteractionUUIDLinks(*item); err != nil {
		return err
	}
	return db.QueryRowContext(ctx, `
		INSERT INTO team_interactions (
			id, tenant_id, team_id, work_item_id, run_id, intent_proof_id, contract_id, proof_id,
			source_kind, source_channel, actor_ref, verb, summary, payload_kind,
			payload_ref, payload, approval_ref, audit_refs, version
		) VALUES (
			$1, 'default', $2, $3, $4, $5, NULLIF($6,''), NULLIF($7,''),
			$8, $9, $10, $11, $12, $13,
			$14, $15, $16, $17, $18
		)
		RETURNING timestamp`,
		item.InteractionID, item.TeamID, item.WorkItemID, nullableUUID(item.RunID),
		nullableUUID(item.IntentProofID), item.ContractID, item.ProofID, item.SourceKind,
		item.SourceChannel, item.ActorRef, item.Verb, item.Summary, item.PayloadKind,
		item.PayloadRef, jsonObjectOrNil(item.Payload), item.ApprovalRef, jsonArray(item.AuditRefs), item.Version,
	).Scan(&item.Timestamp)
}

func scanTeamWorkItem(scanner interface{ Scan(dest ...any) error }) (protocol.TeamWorkItem, error) {
	var item protocol.TeamWorkItem
	var executionShape, governancePosture, state string
	var scope, outputs, proof, caps, lastEvent, recovery, outputRefs, proofRefs, auditRefs []byte
	if err := scanner.Scan(
		&item.WorkItemID, &item.TeamID, &item.RunID, &item.IntentProofID, &item.ContractID, &item.ProofID,
		&item.Objective, &scope, &item.Owner, &executionShape, &outputs, &proof, &caps,
		&governancePosture, &state, &lastEvent, &item.NeedsOperator, &item.DegradationState,
		&recovery, &outputRefs, &proofRefs, &auditRefs, &item.CreatedAt, &item.UpdatedAt, &item.Version,
	); err != nil {
		return item, err
	}
	item.ExecutionShape = protocol.TeamExecutionShape(executionShape)
	item.GovernancePosture = protocol.ApprovalPosture(governancePosture)
	item.State = protocol.TeamWorkState(state)
	item.Scope = decodeStringList(scope)
	item.ExpectedOutputs = decodeStringList(outputs)
	item.ExpectedProof = decodeStringList(proof)
	item.CapabilityRequirements = decodeStringList(caps)
	item.RecoveryOptions = decodeStringList(recovery)
	item.ProofRefs = decodeStringList(proofRefs)
	item.AuditRefs = decodeStringList(auditRefs)
	_ = json.Unmarshal(outputRefs, &item.OutputRefs)
	if len(lastEvent) > 0 && string(lastEvent) != "null" {
		var event protocol.TeamStatusEvent
		if err := json.Unmarshal(lastEvent, &event); err == nil {
			item.LastEvent = &event
		}
	}
	return item, nil
}

func scanTeamInteraction(scanner interface{ Scan(dest ...any) error }) (protocol.TeamInteraction, error) {
	var item protocol.TeamInteraction
	var payload, auditRefs []byte
	if err := scanner.Scan(
		&item.InteractionID, &item.TeamID, &item.WorkItemID, &item.RunID, &item.IntentProofID,
		&item.ContractID, &item.ProofID, &item.SourceKind, &item.SourceChannel, &item.ActorRef,
		&item.Verb, &item.Summary, &item.PayloadKind, &item.PayloadRef, &payload, &item.ApprovalRef,
		&auditRefs, &item.Timestamp, &item.Version,
	); err != nil {
		return item, err
	}
	item.AuditRefs = decodeStringList(auditRefs)
	if len(payload) > 0 && string(payload) != "null" {
		_ = json.Unmarshal(payload, &item.Payload)
	}
	return item, nil
}

func validateTeamWorkUUIDLinks(item protocol.TeamWorkItem) error {
	if err := validateOptionalUUID("work_item_id", item.WorkItemID); err != nil {
		return err
	}
	if err := validateOptionalUUID("run_id", item.RunID); err != nil {
		return err
	}
	return validateOptionalUUID("intent_proof_id", item.IntentProofID)
}

func validateTeamInteractionUUIDLinks(item protocol.TeamInteraction) error {
	if err := validateOptionalUUID("interaction_id", item.InteractionID); err != nil {
		return err
	}
	if err := validateOptionalUUID("work_item_id", item.WorkItemID); err != nil {
		return err
	}
	if err := validateOptionalUUID("run_id", item.RunID); err != nil {
		return err
	}
	return validateOptionalUUID("intent_proof_id", item.IntentProofID)
}

func validateOptionalUUID(label, value string) error {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	if _, err := uuid.Parse(value); err != nil {
		return fmt.Errorf("%s must be a UUID", label)
	}
	return nil
}

func nullableUUID(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return strings.TrimSpace(value)
}

func jsonArray(value any) []byte {
	raw, _ := json.Marshal(value)
	if len(raw) == 0 || string(raw) == "null" {
		return []byte("[]")
	}
	return raw
}

func jsonObjectOrNil(value map[string]any) any {
	if len(value) == 0 {
		return nil
	}
	raw, _ := json.Marshal(value)
	return raw
}

func decodeStringList(raw []byte) []string {
	if len(raw) == 0 {
		return []string{}
	}
	items, err := unmarshalStringList(raw)
	if err != nil {
		return []string{}
	}
	return items
}

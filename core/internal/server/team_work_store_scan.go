package server

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/mycelis/core/pkg/protocol"
)

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
			event.TargetRef = protocol.TargetRefForTeamStatusEvent(event)
			item.LastEvent = &event
		}
	}
	return protocol.NormalizeTeamWorkItem(item), nil
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

func scanTeamStatusEvent(scanner interface{ Scan(dest ...any) error }) (protocol.TeamStatusEvent, error) {
	var item protocol.TeamStatusEvent
	var state string
	var blockedBy, auditRefs []byte
	if err := scanner.Scan(
		&item.EventID, &item.TeamID, &item.WorkItemID, &item.RunID, &item.IntentProofID,
		&item.ContractID, &item.ProofID, &state, &item.Headline, &item.Details,
		&item.ConfidencePosture, &blockedBy, &item.NextAction, &item.SourceKind,
		&item.SourceChannel, &item.PayloadKind, &auditRefs, &item.Timestamp, &item.Version,
	); err != nil {
		return item, err
	}
	item.State = protocol.TeamWorkState(state)
	item.BlockedBy = decodeStringList(blockedBy)
	item.AuditRefs = decodeStringList(auditRefs)
	item.TargetRef = protocol.TargetRefForTeamStatusEvent(item)
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

func validateTeamStatusEventUUIDLinks(item protocol.TeamStatusEvent) error {
	if err := validateOptionalUUID("event_id", item.EventID); err != nil {
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

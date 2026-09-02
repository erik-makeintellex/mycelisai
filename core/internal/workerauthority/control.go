package workerauthority

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

func (s *Store) StageApprovalTx(ctx context.Context, tx *sql.Tx, request ApprovalRequest) (bool, error) {
	if s == nil || s.db == nil || tx == nil {
		return false, ErrUnavailable
	}
	if strings.TrimSpace(request.ID) == "" || strings.TrimSpace(request.RunID) == "" ||
		strings.TrimSpace(request.ApprovalID) == "" || strings.TrimSpace(request.Kind) == "" ||
		strings.TrimSpace(request.Summary) == "" || strings.TrimSpace(request.RiskLevel) == "" ||
		strings.TrimSpace(request.Action) == "" || !digestPattern.MatchString(request.RequestDigest) {
		return false, fmt.Errorf("complete approval identity and digest are required")
	}
	var inserted string
	err := tx.QueryRowContext(ctx, `
INSERT INTO worker_approval_requests
    (id,run_id,approval_id,request_digest,kind,summary,risk_level,requested_action)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
ON CONFLICT DO NOTHING RETURNING id::text`, request.ID, request.RunID,
		request.ApprovalID, request.RequestDigest, request.Kind, request.Summary,
		request.RiskLevel, request.Action).Scan(&inserted)
	if err == nil {
		return true, nil
	}
	if err != sql.ErrNoRows {
		return false, err
	}
	var existing ApprovalRequest
	err = tx.QueryRowContext(ctx, `
SELECT id::text,run_id::text,approval_id,request_digest,kind,summary,risk_level,requested_action
FROM worker_approval_requests WHERE id=$1 OR (run_id=$2 AND approval_id=$3)`,
		request.ID, request.RunID, request.ApprovalID).Scan(&existing.ID, &existing.RunID,
		&existing.ApprovalID, &existing.RequestDigest, &existing.Kind, &existing.Summary,
		&existing.RiskLevel, &existing.Action)
	if err != nil {
		return false, err
	}
	if existing == (ApprovalRequest{ID: request.ID, RunID: request.RunID, ApprovalID: request.ApprovalID,
		RequestDigest: request.RequestDigest, Kind: request.Kind, Summary: request.Summary,
		RiskLevel: request.RiskLevel, Action: request.Action}) {
		return false, nil
	}
	return false, ErrConflict
}

func (s *Store) DecideApproval(ctx context.Context, id, decision, actor, reason string, expectedVersion int64) error {
	if s == nil || s.db == nil {
		return ErrUnavailable
	}
	if decision != "approve" && decision != "deny" {
		return fmt.Errorf("approval decision must be approve or deny")
	}
	if strings.TrimSpace(id) == "" || strings.TrimSpace(actor) == "" || actor != strings.TrimSpace(actor) || expectedVersion < 0 {
		return fmt.Errorf("complete canonical approval authority and version are required")
	}
	result, err := s.db.ExecContext(ctx, `
UPDATE worker_approval_requests
SET state='decided',decision=$2,decided_by=NULLIF($3,''),decision_reason=NULLIF($4,''),
    version=version+1,updated_at=NOW()
WHERE id=$1 AND state='pending' AND version=$5`, id, decision, actor, reason, expectedVersion)
	return requireCAS(result, err)
}

func (s *Store) StageCommandTx(ctx context.Context, tx *sql.Tx, command ControlCommand) (bool, error) {
	if s == nil || s.db == nil || tx == nil {
		return false, ErrUnavailable
	}
	if strings.TrimSpace(command.CommandID) == "" || strings.TrimSpace(command.RunID) == "" ||
		strings.TrimSpace(command.IdempotencyKey) == "" || !digestPattern.MatchString(command.PayloadDigest) ||
		!json.Valid(command.Payload) || command.ExpectedServiceVersion < 1 {
		return false, fmt.Errorf("complete command identity, version, digest, and payload are required")
	}
	if command.Kind != "approve" && command.Kind != "deny" && command.Kind != "stop" {
		return false, fmt.Errorf("unsupported worker command kind %q", command.Kind)
	}
	if (command.Kind == "stop") != (strings.TrimSpace(command.ApprovalRequestID) == "") {
		return false, fmt.Errorf("approval commands require approval_request_id; stop must omit it")
	}
	var inserted string
	err := tx.QueryRowContext(ctx, `
INSERT INTO worker_control_commands
    (command_id,run_id,approval_request_id,idempotency_key,kind,expected_service_version,payload_digest,payload)
VALUES ($1,$2,NULLIF($3,'')::uuid,$4,$5,$6,$7,$8::jsonb)
ON CONFLICT DO NOTHING RETURNING command_id::text`, command.CommandID, command.RunID,
		command.ApprovalRequestID, command.IdempotencyKey, command.Kind,
		command.ExpectedServiceVersion, command.PayloadDigest, string(command.Payload)).Scan(&inserted)
	if err == nil {
		return true, nil
	}
	if err != sql.ErrNoRows {
		return false, err
	}
	var id, runID, approvalID, key, kind, digest string
	var expected int64
	var payload json.RawMessage
	err = tx.QueryRowContext(ctx, `
SELECT command_id::text,run_id::text,COALESCE(approval_request_id::text,''),idempotency_key,
       kind,expected_service_version,payload_digest,payload
FROM worker_control_commands WHERE command_id=$1 OR idempotency_key=$2`,
		command.CommandID, command.IdempotencyKey).Scan(&id, &runID, &approvalID, &key,
		&kind, &expected, &digest, &payload)
	if err != nil {
		return false, err
	}
	if id == command.CommandID && runID == command.RunID && approvalID == command.ApprovalRequestID &&
		key == command.IdempotencyKey && kind == command.Kind && expected == command.ExpectedServiceVersion &&
		digest == command.PayloadDigest && semanticJSONEqual(payload, command.Payload) {
		return false, nil
	}
	return false, ErrConflict
}

func semanticJSONEqual(left, right json.RawMessage) bool {
	var leftValue, rightValue any
	if json.Unmarshal(left, &leftValue) != nil || json.Unmarshal(right, &rightValue) != nil {
		return false
	}
	return reflect.DeepEqual(leftValue, rightValue)
}

func (s *Store) TransitionCommand(ctx context.Context, id, from, to, lastError string, expectedVersion int64) error {
	if s == nil || s.db == nil {
		return ErrUnavailable
	}
	allowed := map[string]map[string]bool{
		CommandStaged:    {CommandPending: true, CommandFailed: true},
		CommandPending:   {CommandAcknowledged: true, CommandFailed: true, CommandUncertain: true},
		CommandUncertain: {CommandPending: true, CommandAcknowledged: true, CommandFailed: true},
	}
	if !allowed[from][to] || expectedVersion < 0 || strings.TrimSpace(id) == "" {
		return fmt.Errorf("invalid worker command transition")
	}
	result, err := s.db.ExecContext(ctx, `
UPDATE worker_control_commands
SET state=$3,last_error=NULLIF($4,''),version=version+1,updated_at=NOW(),
    acknowledged_at=CASE WHEN $3='acknowledged' THEN NOW() ELSE acknowledged_at END
WHERE command_id=$1 AND state=$2 AND version=$5`, id, from, to, lastError, expectedVersion)
	return requireCAS(result, err)
}

func requireCAS(result sql.Result, err error) error {
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return ErrStaleVersion
	}
	return nil
}

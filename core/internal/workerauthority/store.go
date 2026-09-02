package workerauthority

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

var digestPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

type Store struct{ db *sql.DB }

func NewStore(db *sql.DB) *Store { return &Store{db: db} }

func (s *Store) StageBindingTx(ctx context.Context, tx *sql.Tx, binding Binding) (bool, error) {
	if s == nil || s.db == nil || tx == nil {
		return false, ErrUnavailable
	}
	if err := validateBinding(binding); err != nil {
		return false, err
	}
	var inserted string
	err := tx.QueryRowContext(ctx, `
INSERT INTO worker_run_bindings (
    run_id, backend, protocol, intent_proof_id, execution_contract_id,
    team_id, work_item_id, outcome_id, idempotency_key, graph_revision,
    source_kind, source_channel, payload_kind, request_digest
) VALUES ($1,'framework_runs','runs_api',$2,$3,NULLIF($4,''),$5,NULLIF($6,''),$7,$8,$9,$10,$11,$12)
ON CONFLICT DO NOTHING RETURNING run_id::text`,
		binding.RunID, binding.IntentProofID, binding.ExecutionContractID,
		binding.TeamID, binding.WorkItemID, binding.OutcomeID, binding.IdempotencyKey,
		binding.GraphRevision, binding.SourceKind, binding.SourceChannel,
		binding.PayloadKind, binding.RequestDigest).Scan(&inserted)
	if err == nil {
		return true, nil
	}
	if err != sql.ErrNoRows {
		return false, err
	}
	existing, err := loadBindingTx(ctx, tx, binding.RunID, binding.IdempotencyKey)
	if err != nil {
		return false, err
	}
	if sameBinding(existing, binding) {
		return false, nil
	}
	return false, ErrConflict
}

func (s *Store) BindRun(ctx context.Context, runID string, serviceVersion, expectedVersion int64) error {
	if s == nil || s.db == nil {
		return ErrUnavailable
	}
	if strings.TrimSpace(runID) == "" || serviceVersion < 1 || expectedVersion < 0 {
		return fmt.Errorf("valid run identity and versions are required")
	}
	result, err := s.db.ExecContext(ctx, `
UPDATE worker_run_bindings
SET service_version=$2, state='bound', run_status='accepted',
    binding_version=binding_version+1, bound_at=COALESCE(bound_at,NOW()), updated_at=NOW()
WHERE run_id=$1 AND binding_version=$3 AND state IN ('create_pending','reconcile_required')`,
		runID, serviceVersion, expectedVersion)
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

// ProjectEvent claims a receipt, applies the existing Core projection, and
// advances the contiguous cursor in the same transaction.
func (s *Store) ProjectEvent(ctx context.Context, receipt EventReceipt, project func(*sql.Tx) (string, error)) (bool, error) {
	if s == nil || s.db == nil {
		return false, ErrUnavailable
	}
	if err := validateReceipt(receipt); err != nil {
		return false, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	var authority Binding
	err = tx.QueryRowContext(ctx, `
SELECT state,intent_proof_id::text,execution_contract_id::text,COALESCE(team_id,''),
       work_item_id::text,COALESCE(outcome_id,''),idempotency_key,graph_revision,
       source_kind,source_channel,payload_kind,request_digest,COALESCE(service_version,0),
       last_event_sequence,cursor_version
FROM worker_run_bindings WHERE run_id=$1 FOR UPDATE`, receipt.RunID).
		Scan(&authority.State, &authority.IntentProofID, &authority.ExecutionContractID,
			&authority.TeamID, &authority.WorkItemID, &authority.OutcomeID,
			&authority.IdempotencyKey, &authority.GraphRevision, &authority.SourceKind,
			&authority.SourceChannel, &authority.PayloadKind, &authority.RequestDigest,
			&authority.ServiceVersion, &authority.LastEventSequence, &authority.CursorVersion)
	if err != nil {
		return false, err
	}
	authority.RunID = receipt.RunID
	if authority.State == BindingCreatePending || !sameCorrelation(authority, receipt.Correlation) {
		return false, ErrConflict
	}
	if receipt.Sequence <= authority.LastEventSequence {
		duplicate, duplicateErr := exactReceiptTx(ctx, tx, receipt)
		if duplicateErr != nil {
			return false, duplicateErr
		}
		if !duplicate {
			return false, ErrConflict
		}
		return false, tx.Commit()
	}
	if authority.State == BindingTerminal {
		return false, ErrConflict
	}
	if receipt.Sequence != authority.LastEventSequence+1 {
		return false, ErrEventGap
	}
	if receipt.ExpectedCursorVersion != authority.CursorVersion {
		return false, ErrStaleVersion
	}
	_, err = tx.ExecContext(ctx, `
INSERT INTO worker_event_receipts
    (id,run_id,event_id,sequence,event_kind,service_version,payload_digest,normalized_payload)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8::jsonb)`, receipt.ID, receipt.RunID, receipt.EventID,
		receipt.Sequence, receipt.EventKind, receipt.ServiceVersion, receipt.PayloadDigest, string(receipt.NormalizedPayload))
	if err != nil {
		return false, fmt.Errorf("claim worker event receipt: %w", err)
	}
	projectionID, err := project(tx)
	if err != nil {
		return false, err
	}
	if _, err := uuid.Parse(projectionID); err != nil {
		return false, fmt.Errorf("projected mission event id is invalid: %w", err)
	}
	if _, err = tx.ExecContext(ctx, `
UPDATE worker_event_receipts
SET projected_mission_event_id=NULLIF($3,'')::uuid,projected_at=NOW()
WHERE run_id=$1 AND sequence=$2`, receipt.RunID, receipt.Sequence, projectionID); err != nil {
		return false, err
	}
	result, err := tx.ExecContext(ctx, `
UPDATE worker_run_bindings
SET last_event_sequence=$2,cursor_version=cursor_version+1,
    service_version=GREATEST(COALESCE(service_version,0),$3),
    run_status=NULLIF($4,''),
    state=CASE WHEN $4 IN ('completed','failed','cancelled') THEN 'terminal' ELSE state END,
    updated_at=NOW()
WHERE run_id=$1 AND cursor_version=$5 AND last_event_sequence=$6`, receipt.RunID,
		receipt.Sequence, receipt.ServiceVersion, receipt.RunStatus,
		receipt.ExpectedCursorVersion, authority.LastEventSequence)
	if err != nil {
		return false, err
	}
	count, err := result.RowsAffected()
	if err != nil || count != 1 {
		if err != nil {
			return false, err
		}
		return false, ErrStaleVersion
	}
	if err := tx.Commit(); err != nil {
		return false, err
	}
	return true, nil
}

func exactReceiptTx(ctx context.Context, tx *sql.Tx, receipt EventReceipt) (bool, error) {
	var eventID, digest string
	var sequence int64
	err := tx.QueryRowContext(ctx, `
SELECT event_id,sequence,payload_digest FROM worker_event_receipts
WHERE run_id=$1 AND (event_id=$2 OR sequence=$3)`, receipt.RunID, receipt.EventID, receipt.Sequence).
		Scan(&eventID, &sequence, &digest)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return eventID == receipt.EventID && sequence == receipt.Sequence && digest == receipt.PayloadDigest, err
}

func loadBindingTx(ctx context.Context, tx *sql.Tx, runID, key string) (Binding, error) {
	var b Binding
	err := tx.QueryRowContext(ctx, `
SELECT run_id::text,intent_proof_id::text,execution_contract_id::text,COALESCE(team_id,''),
       work_item_id::text,COALESCE(outcome_id,''),idempotency_key,graph_revision,
       source_kind,source_channel,payload_kind,request_digest
FROM worker_run_bindings WHERE run_id=$1 OR idempotency_key=$2`, runID, key).
		Scan(&b.RunID, &b.IntentProofID, &b.ExecutionContractID, &b.TeamID,
			&b.WorkItemID, &b.OutcomeID, &b.IdempotencyKey, &b.GraphRevision,
			&b.SourceKind, &b.SourceChannel, &b.PayloadKind, &b.RequestDigest)
	return b, err
}

func sameBinding(a, b Binding) bool {
	return a.RunID == b.RunID && a.IntentProofID == b.IntentProofID &&
		a.ExecutionContractID == b.ExecutionContractID && a.TeamID == b.TeamID &&
		a.WorkItemID == b.WorkItemID && a.OutcomeID == b.OutcomeID &&
		a.IdempotencyKey == b.IdempotencyKey && a.GraphRevision == b.GraphRevision &&
		a.SourceKind == b.SourceKind && a.SourceChannel == b.SourceChannel &&
		a.PayloadKind == b.PayloadKind && a.RequestDigest == b.RequestDigest
}

func sameCorrelation(binding Binding, correlation Correlation) bool {
	return binding.RunID == correlation.RunID && binding.IntentProofID == correlation.IntentProofID &&
		binding.ExecutionContractID == correlation.ExecutionContractID &&
		binding.TeamID == correlation.TeamID && binding.WorkItemID == correlation.WorkItemID &&
		binding.OutcomeID == correlation.OutcomeID && binding.IdempotencyKey == correlation.IdempotencyKey &&
		binding.GraphRevision == correlation.GraphRevision && binding.SourceKind == correlation.SourceKind &&
		binding.SourceChannel == correlation.SourceChannel && binding.PayloadKind == correlation.PayloadKind
}

func validateBinding(b Binding) error {
	for label, value := range map[string]string{"run_id": b.RunID, "intent_proof_id": b.IntentProofID,
		"execution_contract_id": b.ExecutionContractID, "work_item_id": b.WorkItemID,
		"idempotency_key": b.IdempotencyKey, "graph_revision": b.GraphRevision,
		"source_kind": b.SourceKind, "source_channel": b.SourceChannel, "payload_kind": b.PayloadKind} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", label)
		}
	}
	if !digestPattern.MatchString(b.RequestDigest) {
		return fmt.Errorf("request_digest must be lowercase SHA-256")
	}
	canonicalSources := map[string]bool{
		"workspace_ui": true, "web_api": true, "automation_trigger": true,
		"scheduler": true, "sensor": true, "iot": true, "internal_tool": true,
		"mcp": true, "system": true,
	}
	if b.PayloadKind != "command" || !canonicalSources[b.SourceKind] {
		return fmt.Errorf("worker binding vocabulary or identity is not canonical")
	}
	for _, value := range []string{b.RunID, b.IntentProofID, b.ExecutionContractID,
		b.TeamID, b.WorkItemID, b.OutcomeID, b.IdempotencyKey, b.GraphRevision,
		b.SourceKind, b.SourceChannel, b.PayloadKind} {
		if value != strings.TrimSpace(value) {
			return fmt.Errorf("worker binding identity must not contain surrounding whitespace")
		}
	}
	return nil
}

func validateReceipt(r EventReceipt) error {
	if strings.TrimSpace(r.ID) == "" || strings.TrimSpace(r.RunID) == "" ||
		strings.TrimSpace(r.EventID) == "" || strings.TrimSpace(r.EventKind) == "" || r.Sequence <= 0 ||
		r.ServiceVersion < 1 || r.ExpectedCursorVersion < 0 {
		return fmt.Errorf("complete worker event identity and positive sequence are required")
	}
	for _, value := range []string{r.Correlation.RunID, r.Correlation.IntentProofID,
		r.Correlation.ExecutionContractID, r.Correlation.WorkItemID,
		r.Correlation.IdempotencyKey, r.Correlation.GraphRevision,
		r.Correlation.SourceKind, r.Correlation.SourceChannel, r.Correlation.PayloadKind} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("worker event correlation is incomplete")
		}
	}
	computedDigest := fmt.Sprintf("%x", sha256.Sum256(r.NormalizedPayload))
	if r.Correlation.RunID != r.RunID || r.Correlation.PayloadKind != "command" ||
		!digestPattern.MatchString(r.PayloadDigest) || r.PayloadDigest != computedDigest ||
		!json.Valid(r.NormalizedPayload) {
		return fmt.Errorf("worker event digest and normalized payload are invalid")
	}
	pairs := map[string]string{
		"accepted": "accepted", "progress": "running", "approval_needed": "approval_needed",
		"completed": "completed", "failed": "failed", "cancelled": "cancelled",
	}
	if pairs[r.EventKind] != r.RunStatus {
		return fmt.Errorf("worker event kind and status are inconsistent")
	}
	return nil
}

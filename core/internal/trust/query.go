package trust

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mycelis/core/pkg/protocol"
)

type ListOptions struct {
	Limit         int
	RunID         string
	IntentProofID string
	ContractID    string
	Status        string
}

func (s *Store) ListExecutionContracts(ctx context.Context, opts ListOptions) ([]protocol.ExecutionContractRecord, error) {
	if s == nil || s.db == nil {
		return nil, fmt.Errorf("trust: database not available")
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id::text, COALESCE(intent_proof_id::text,''), COALESCE(run_id::text,''), template_id,
		       resolved_intent, execution_shape, status, execution_status, validation_source,
		       evidence_strength, proof_quality, COALESCE(latest_proof_artifact_id::text,''),
		       COALESCE(audit_event_id::text,''), output_refs, audit_refs, review_lineage,
		       degradation, recovery, created_at, confirmed_at, completed_at, updated_at
		FROM execution_contracts
		WHERE tenant_id='default'
		  AND ($1 = '' OR run_id::text = $1)
		  AND ($2 = '' OR intent_proof_id::text = $2)
		  AND ($3 = '' OR status = $3)
		ORDER BY updated_at DESC
		LIMIT $4`,
		strings.TrimSpace(opts.RunID), strings.TrimSpace(opts.IntentProofID), strings.TrimSpace(opts.Status), boundedTrustLimit(opts.Limit),
	)
	if err != nil {
		return nil, fmt.Errorf("trust: list execution contracts: %w", err)
	}
	defer rows.Close()
	items := make([]protocol.ExecutionContractRecord, 0)
	for rows.Next() {
		item, scanErr := scanExecutionContract(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) GetExecutionContract(ctx context.Context, id string) (protocol.ExecutionContractRecord, error) {
	if s == nil || s.db == nil {
		return protocol.ExecutionContractRecord{}, fmt.Errorf("trust: database not available")
	}
	row := s.db.QueryRowContext(ctx, `
		SELECT id::text, COALESCE(intent_proof_id::text,''), COALESCE(run_id::text,''), template_id,
		       resolved_intent, execution_shape, status, execution_status, validation_source,
		       evidence_strength, proof_quality, COALESCE(latest_proof_artifact_id::text,''),
		       COALESCE(audit_event_id::text,''), output_refs, audit_refs, review_lineage,
		       degradation, recovery, created_at, confirmed_at, completed_at, updated_at
		FROM execution_contracts
		WHERE tenant_id='default' AND id=$1`, strings.TrimSpace(id))
	item, err := scanExecutionContract(row)
	if err != nil {
		return protocol.ExecutionContractRecord{}, err
	}
	return item, nil
}

func (s *Store) ListProofArtifacts(ctx context.Context, opts ListOptions) ([]protocol.ProofArtifactRecord, error) {
	if s == nil || s.db == nil {
		return nil, fmt.Errorf("trust: database not available")
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id::text, COALESCE(contract_id::text,''), COALESCE(intent_proof_id::text,''),
		       COALESCE(run_id::text,''), artifact_kind, status, proof_class, validation_source,
		       evidence_strength, proof_quality, output_refs, audit_refs, review_lineage,
		       degradation, recovery, payload, created_at
		FROM proof_artifacts
		WHERE tenant_id='default'
		  AND ($1 = '' OR contract_id::text = $1)
		  AND ($2 = '' OR run_id::text = $2)
		  AND ($3 = '' OR intent_proof_id::text = $3)
		  AND ($4 = '' OR status = $4)
		ORDER BY created_at DESC
		LIMIT $5`,
		strings.TrimSpace(opts.ContractID), strings.TrimSpace(opts.RunID), strings.TrimSpace(opts.IntentProofID),
		strings.TrimSpace(opts.Status), boundedTrustLimit(opts.Limit),
	)
	if err != nil {
		return nil, fmt.Errorf("trust: list proof artifacts: %w", err)
	}
	defer rows.Close()
	items := make([]protocol.ProofArtifactRecord, 0)
	for rows.Next() {
		item, scanErr := scanProofArtifact(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) GetProofArtifact(ctx context.Context, id string) (protocol.ProofArtifactRecord, error) {
	if s == nil || s.db == nil {
		return protocol.ProofArtifactRecord{}, fmt.Errorf("trust: database not available")
	}
	row := s.db.QueryRowContext(ctx, `
		SELECT id::text, COALESCE(contract_id::text,''), COALESCE(intent_proof_id::text,''),
		       COALESCE(run_id::text,''), artifact_kind, status, proof_class, validation_source,
		       evidence_strength, proof_quality, output_refs, audit_refs, review_lineage,
		       degradation, recovery, payload, created_at
		FROM proof_artifacts
		WHERE tenant_id='default' AND id=$1`, strings.TrimSpace(id))
	item, err := scanProofArtifact(row)
	if err != nil {
		return protocol.ProofArtifactRecord{}, err
	}
	return item, nil
}

func scanExecutionContract(scanner interface{ Scan(dest ...any) error }) (protocol.ExecutionContractRecord, error) {
	var item protocol.ExecutionContractRecord
	var templateID, executionShape, status, executionStatus, validationSource, evidenceStrength, proofQuality string
	var outputRefs, auditRefs, reviewLineage, degradation, recovery []byte
	var createdAt, updatedAt time.Time
	var confirmedAt, completedAt sql.NullTime
	if err := scanner.Scan(
		&item.ID, &item.IntentProofID, &item.RunID, &templateID, &item.ResolvedIntent, &executionShape,
		&status, &executionStatus, &validationSource, &evidenceStrength, &proofQuality,
		&item.LatestProofArtifactID, &item.AuditEventID, &outputRefs, &auditRefs, &reviewLineage,
		&degradation, &recovery, &createdAt, &confirmedAt, &completedAt, &updatedAt,
	); err != nil {
		return item, err
	}
	item.TemplateID = protocol.TemplateID(templateID)
	item.ExecutionShape = protocol.ExecutionShape(executionShape)
	item.Status = protocol.ExecutionContractStatus(status)
	item.ExecutionStatus = protocol.ExecutionStatus(executionStatus)
	item.ValidationSource = protocol.TrustValidationSource(validationSource)
	item.EvidenceStrength = protocol.TrustEvidenceStrength(evidenceStrength)
	item.ProofQuality = protocol.TrustProofQuality(proofQuality)
	_ = json.Unmarshal(outputRefs, &item.OutputRefs)
	_ = json.Unmarshal(auditRefs, &item.AuditRefs)
	_ = json.Unmarshal(reviewLineage, &item.ReviewLineage)
	_ = json.Unmarshal(degradation, &item.Degradation)
	_ = json.Unmarshal(recovery, &item.Recovery)
	item.CreatedAt = formatTime(createdAt)
	item.ConfirmedAt = formatNullTime(confirmedAt)
	item.CompletedAt = formatNullTime(completedAt)
	item.UpdatedAt = formatTime(updatedAt)
	return item, nil
}

func scanProofArtifact(scanner interface{ Scan(dest ...any) error }) (protocol.ProofArtifactRecord, error) {
	var item protocol.ProofArtifactRecord
	var status, proofClass, validationSource, evidenceStrength, proofQuality string
	var outputRefs, auditRefs, reviewLineage, degradation, recovery, payload []byte
	var createdAt time.Time
	if err := scanner.Scan(
		&item.ID, &item.ContractID, &item.IntentProofID, &item.RunID, &item.ArtifactKind, &status,
		&proofClass, &validationSource, &evidenceStrength, &proofQuality, &outputRefs, &auditRefs,
		&reviewLineage, &degradation, &recovery, &payload, &createdAt,
	); err != nil {
		return item, err
	}
	item.Status = protocol.ProofArtifactStatus(status)
	item.ProofClass = protocol.ExecutionProofClass(proofClass)
	item.ValidationSource = protocol.TrustValidationSource(validationSource)
	item.EvidenceStrength = protocol.TrustEvidenceStrength(evidenceStrength)
	item.ProofQuality = protocol.TrustProofQuality(proofQuality)
	_ = json.Unmarshal(outputRefs, &item.OutputRefs)
	_ = json.Unmarshal(auditRefs, &item.AuditRefs)
	_ = json.Unmarshal(reviewLineage, &item.ReviewLineage)
	_ = json.Unmarshal(degradation, &item.Degradation)
	_ = json.Unmarshal(recovery, &item.Recovery)
	_ = json.Unmarshal(payload, &item.Payload)
	item.CreatedAt = formatTime(createdAt)
	return item, nil
}

func boundedTrustLimit(limit int) int {
	if limit <= 0 {
		return 20
	}
	if limit > 100 {
		return 100
	}
	return limit
}

func formatTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339Nano)
}

func formatNullTime(value sql.NullTime) string {
	if !value.Valid {
		return ""
	}
	return formatTime(value.Time)
}

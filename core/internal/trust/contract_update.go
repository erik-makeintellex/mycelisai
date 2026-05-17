package trust

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

func updateContractFromProof(ctx context.Context, exec SQLExecutor, proofID string, input ProofArtifactInput, status protocol.ProofArtifactStatus, validationSource protocol.TrustValidationSource, evidenceStrength protocol.TrustEvidenceStrength, proofQuality protocol.TrustProofQuality) error {
	contractStatus := protocol.ExecutionContractStatusCompleted
	executionStatus := protocol.ExecutionStatusCompleted
	if status == protocol.ProofArtifactStatusFailure || status == protocol.ProofArtifactStatusDegraded {
		contractStatus = protocol.ExecutionContractStatusFailed
		executionStatus = protocol.ExecutionStatusFailed
	}

	result, err := exec.ExecContext(ctx, `
		UPDATE execution_contracts
		SET run_id = COALESCE($2, run_id),
		    status = $3,
		    execution_status = $4,
		    validation_source = $5,
		    evidence_strength = $6,
		    proof_quality = $7,
		    latest_proof_artifact_id = $8,
		    audit_event_id = COALESCE($9, audit_event_id),
		    output_refs = $10,
		    audit_refs = $11,
		    review_lineage = $12,
		    degradation = $13,
		    recovery = $14,
		    confirmed_at = COALESCE(confirmed_at, NOW()),
		    completed_at = NOW(),
		    updated_at = NOW()
		WHERE id = $1`,
		nullableUUID(input.ContractID),
		nullableUUID(input.RunID),
		string(contractStatus),
		string(executionStatus),
		string(validationSource),
		string(evidenceStrength),
		string(proofQuality),
		nullableUUID(proofID),
		nullableUUID(input.AuditEventID),
		mustJSON(input.OutputRefs),
		mustJSON(auditRefs(input.AuditEventID)),
		mustJSON(input.ReviewLineage),
		mustJSON(input.Degradation),
		mustJSON(input.Recovery),
	)
	if err != nil {
		return fmt.Errorf("trust: update contract from proof: %w", err)
	}
	if rowsAffected(result) > 0 || strings.TrimSpace(input.IntentProofID) == "" {
		return nil
	}
	_, err = exec.ExecContext(ctx, `
		UPDATE execution_contracts
		SET run_id = COALESCE($2, run_id),
		    status = $3,
		    execution_status = $4,
		    validation_source = $5,
		    evidence_strength = $6,
		    proof_quality = $7,
		    latest_proof_artifact_id = $8,
		    audit_event_id = COALESCE($9, audit_event_id),
		    output_refs = $10,
		    audit_refs = $11,
		    review_lineage = $12,
		    degradation = $13,
		    recovery = $14,
		    confirmed_at = COALESCE(confirmed_at, NOW()),
		    completed_at = NOW(),
		    updated_at = NOW()
		WHERE intent_proof_id = $1`,
		nullableUUID(input.IntentProofID),
		nullableUUID(input.RunID),
		string(contractStatus),
		string(executionStatus),
		string(validationSource),
		string(evidenceStrength),
		string(proofQuality),
		nullableUUID(proofID),
		nullableUUID(input.AuditEventID),
		mustJSON(input.OutputRefs),
		mustJSON(auditRefs(input.AuditEventID)),
		mustJSON(input.ReviewLineage),
		mustJSON(input.Degradation),
		mustJSON(input.Recovery),
	)
	if err != nil {
		return fmt.Errorf("trust: update contract from proof by intent proof: %w", err)
	}
	return nil
}

func rowsAffected(result sql.Result) int64 {
	if result == nil {
		return 0
	}
	count, _ := result.RowsAffected()
	return count
}

package trust

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mycelis/core/pkg/protocol"
)

type SQLExecutor interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

type ContractInput struct {
	ID             string
	IntentProofID  string
	RunID          string
	TemplateID     protocol.TemplateID
	ResolvedIntent string
	AuditEventID   string
}

type ProofArtifactInput struct {
	ID               string
	ContractID       string
	IntentProofID    string
	RunID            string
	AuditEventID     string
	Status           protocol.ProofArtifactStatus
	ProofClass       protocol.ExecutionProofClass
	ValidationSource protocol.TrustValidationSource
	EvidenceStrength protocol.TrustEvidenceStrength
	ProofQuality     protocol.TrustProofQuality
	OutputRefs       any
	AuditRefs        any
	ReviewLineage    any
	Degradation      any
	Recovery         any
	Payload          any
}

func (s *Store) UpsertContract(ctx context.Context, input ContractInput) (string, error) {
	if s == nil || s.db == nil {
		return "", fmt.Errorf("trust: database not available")
	}
	return UpsertContract(ctx, s.db, input)
}

func UpsertContract(ctx context.Context, exec SQLExecutor, input ContractInput) (string, error) {
	if exec == nil {
		return "", fmt.Errorf("trust: database not available")
	}
	id := strings.TrimSpace(input.ID)
	if id == "" {
		id = uuid.NewString()
	}
	intentProofID := nullableUUID(input.IntentProofID)
	if intentProofID == nil {
		return "", fmt.Errorf("trust: intent_proof_id is required")
	}
	templateID := strings.TrimSpace(string(input.TemplateID))
	if templateID == "" {
		templateID = string(protocol.TemplateChatToProposal)
	}
	resolvedIntent := strings.TrimSpace(input.ResolvedIntent)

	err := exec.QueryRowContext(ctx, `
		INSERT INTO execution_contracts
			(id, intent_proof_id, run_id, template_id, resolved_intent, execution_shape, status,
			 execution_status, validation_source, evidence_strength, proof_quality, audit_event_id,
			 audit_refs, review_lineage, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, NOW())
		ON CONFLICT (intent_proof_id) WHERE intent_proof_id IS NOT NULL DO UPDATE
		SET run_id = COALESCE(excluded.run_id, execution_contracts.run_id),
		    template_id = excluded.template_id,
		    resolved_intent = CASE WHEN excluded.resolved_intent <> '' THEN excluded.resolved_intent ELSE execution_contracts.resolved_intent END,
		    audit_event_id = COALESCE(excluded.audit_event_id, execution_contracts.audit_event_id),
		    updated_at = NOW()
		RETURNING id::text`,
		id,
		intentProofID,
		nullableUUID(input.RunID),
		templateID,
		resolvedIntent,
		string(protocol.ExecutionShapeGuidedProposal),
		string(protocol.ExecutionContractStatusProposed),
		string(protocol.ExecutionStatusProposed),
		string(protocol.TrustValidationSourceIntentProof),
		string(protocol.TrustEvidenceStrengthIntentOnly),
		string(protocol.TrustProofQualityProposed),
		nullableUUID(input.AuditEventID),
		mustJSON(auditRefs(input.AuditEventID)),
		mustJSON([]map[string]string{{"event": "contract_created", "source": string(protocol.TrustValidationSourceIntentProof), "at": time.Now().UTC().Format(time.RFC3339)}}),
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("trust: upsert contract: %w", err)
	}
	return id, nil
}

func (s *Store) RecordProofArtifact(ctx context.Context, input ProofArtifactInput) (string, error) {
	if s == nil || s.db == nil {
		return "", fmt.Errorf("trust: database not available")
	}
	return RecordProofArtifact(ctx, s.db, input)
}

func RecordProofArtifact(ctx context.Context, exec SQLExecutor, input ProofArtifactInput) (string, error) {
	if exec == nil {
		return "", fmt.Errorf("trust: database not available")
	}
	id := strings.TrimSpace(input.ID)
	if id == "" {
		id = uuid.NewString()
	}
	status := input.Status
	if status == "" {
		status = protocol.ProofArtifactStatusSuccess
	}
	proofClass := input.ProofClass
	if proofClass == "" {
		proofClass = protocol.ExecutionProofClassRunAudit
	}
	validationSource := input.ValidationSource
	if validationSource == "" {
		validationSource = protocol.TrustValidationSourceConfirmAction
	}
	evidenceStrength := input.EvidenceStrength
	if evidenceStrength == "" {
		evidenceStrength = protocol.TrustEvidenceStrengthRunAudit
	}
	proofQuality := input.ProofQuality
	if proofQuality == "" {
		proofQuality = proofQualityForStatus(status)
	}
	auditRefsValue := input.AuditRefs
	if auditRefsValue == nil {
		auditRefsValue = auditRefs(input.AuditEventID)
	}

	err := exec.QueryRowContext(ctx, `
		INSERT INTO proof_artifacts
			(id, contract_id, intent_proof_id, run_id, artifact_kind, status, proof_class,
			 validation_source, evidence_strength, proof_quality, output_refs, audit_refs,
			 review_lineage, degradation, recovery, payload)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING id::text`,
		id,
		nullableUUID(input.ContractID),
		nullableUUID(input.IntentProofID),
		nullableUUID(input.RunID),
		"confirm_action",
		string(status),
		string(proofClass),
		string(validationSource),
		string(evidenceStrength),
		string(proofQuality),
		mustJSON(input.OutputRefs),
		mustJSON(auditRefsValue),
		mustJSON(input.ReviewLineage),
		mustJSON(input.Degradation),
		mustJSON(input.Recovery),
		mustJSON(input.Payload),
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("trust: insert proof artifact: %w", err)
	}

	if err := updateContractFromProof(ctx, exec, id, input, status, validationSource, evidenceStrength, proofQuality); err != nil {
		return "", err
	}
	return id, nil
}

func proofQualityForStatus(status protocol.ProofArtifactStatus) protocol.TrustProofQuality {
	switch status {
	case protocol.ProofArtifactStatusFailure, protocol.ProofArtifactStatusDegraded:
		return protocol.TrustProofQualityFailed
	default:
		return protocol.TrustProofQualityVerified
	}
}

func nullableUUID(raw string) any {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parsed, err := uuid.Parse(raw)
	if err != nil || parsed == uuid.Nil {
		return nil
	}
	return parsed
}

func mustJSON(value any) []byte {
	if value == nil {
		return []byte("{}")
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return []byte("{}")
	}
	return raw
}

func auditRefs(auditEventID string) []map[string]string {
	auditEventID = strings.TrimSpace(auditEventID)
	if auditEventID == "" {
		return []map[string]string{}
	}
	return []map[string]string{{"audit_event_id": auditEventID, "source": "log_entries"}}
}

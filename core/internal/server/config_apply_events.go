package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/mycelis/core/pkg/protocol"
)

// Audit sources for governed config mutations (RC-004G). The mission event
// types are protocol.EventTemplateCloned and protocol.EventConfigApplied.
const (
	auditSourceTemplateClonePreview = "template.clone.previewed"
	auditSourceConfigApplyPreview   = "config.apply.previewed"
	auditSourceTemplateCloned       = string(protocol.EventTemplateCloned)
	auditSourceConfigApplied        = string(protocol.EventConfigApplied)
	configMutationSourceChannel     = "api.config"
)

// configMutationRecord is what a confirmed clone or apply records.
type configMutationRecord struct {
	EventType    protocol.EventType
	AuditMessage string
	Kind         string
	ID           string
	DocumentSHA  string
	ProofID      string
	Actor        string
	Details      map[string]any
}

func (rec configMutationRecord) payload() map[string]any {
	payload := map[string]any{
		"kind":            rec.Kind,
		"id":              rec.ID,
		"document_sha256": rec.DocumentSHA,
		"intent_proof_id": rec.ProofID,
		"actor":           rec.Actor,
		"source_kind":     string(protocol.SourceKindWebAPI),
		"source_channel":  configMutationSourceChannel,
		"payload_kind":    string(protocol.PayloadKindEvent),
	}
	for key, value := range rec.Details {
		if _, reserved := payload[key]; !reserved {
			payload[key] = value
		}
	}
	return payload
}

// configMutationResult identifies the records a confirmed mutation wrote.
type configMutationResult struct {
	AuditEventID   string `json:"audit_event_id"`
	RunID          string `json:"run_id"`
	MissionEventID string `json:"mission_event_id"`
}

// auditConfigMutationTx writes the audit entry inside the mutation
// transaction, so an audit record exists only for a committed change.
func auditConfigMutationTx(tx *sql.Tx, rec configMutationRecord) (string, error) {
	return insertAuditEvent(tx, protocol.TemplateChatToProposal, string(rec.EventType), rec.AuditMessage, rec.payload())
}

// recordConfigMutationTx persists the governed mission trail for a confirmed
// mutation: a governance run keyed by the intent proof, the mutation event,
// run completion, and the confirmed proof. All writes share tx.
func (s *AdminServer) recordConfigMutationTx(tx *sql.Tx, rec configMutationRecord, auditID string) (configMutationResult, error) {
	result := configMutationResult{AuditEventID: auditID}
	runID, err := s.createExecutionRunTx(tx, rec.ProofID)
	if err != nil {
		return result, fmt.Errorf("create governance run: %w", err)
	}
	result.RunID = runID
	payload := rec.payload()
	payload["run_id"] = runID
	payload["audit_event_id"] = auditID
	payloadJSON, _ := json.Marshal(payload)
	eventID := uuid.New().String()
	if _, err := tx.Exec(
		`INSERT INTO mission_events
			(id, run_id, tenant_id, event_type, severity, source_agent, source_team, payload, audit_event_id, emitted_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		eventID, runID, "default", string(rec.EventType), string(protocol.SeverityInfo),
		"admin", "governance", payloadJSON, parseAuditUUID(auditID), time.Now(),
	); err != nil {
		return result, fmt.Errorf("persist %s event: %w", rec.EventType, err)
	}
	result.MissionEventID = eventID
	if err := s.markRunCompletedTx(tx, runID, rec.ProofID); err != nil {
		return result, fmt.Errorf("complete governance run: %w", err)
	}
	if err := s.confirmChatProofTx(tx, rec.ProofID); err != nil {
		return result, fmt.Errorf("confirm intent proof: %w", err)
	}
	return result, nil
}

// publishConfigMutationSignal sends the best-effort CTS signal for a
// committed mutation on the run's mission-events subject. The persisted
// mission event is the record; the signal only references it.
func (s *AdminServer) publishConfigMutationSignal(rec configMutationRecord, result configMutationResult) {
	if s.NC == nil {
		return
	}
	msg, err := json.Marshal(map[string]any{
		"mission_event_id": result.MissionEventID,
		"run_id":           result.RunID,
		"event_type":       string(rec.EventType),
		"source_agent":     "admin",
		"source_kind":      string(protocol.SourceKindWebAPI),
		"source_channel":   configMutationSourceChannel,
		"payload_kind":     string(protocol.PayloadKindEvent),
		"timestamp":        time.Now().UTC(),
		"kind":             rec.Kind,
		"id":               rec.ID,
	})
	if err != nil {
		return
	}
	if err := s.NC.Publish(fmt.Sprintf(protocol.TopicMissionEventsFmt, result.RunID), msg); err != nil {
		log.Printf("[config-apply] CTS publish failed for %s: %v", result.MissionEventID, err)
	}
}

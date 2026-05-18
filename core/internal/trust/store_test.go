package trust

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/pkg/protocol"
)

func TestExecutionTrustMigrationDefinesContractsAndProofArtifacts(t *testing.T) {
	raw, err := os.ReadFile(filepath.FromSlash("../../migrations/041_execution_trust_spine.up.sql"))
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	sql := string(raw)
	for _, want := range []string{
		"CREATE TABLE IF NOT EXISTS execution_contracts",
		"CREATE TABLE IF NOT EXISTS proof_artifacts",
		"validation_source",
		"evidence_strength",
		"proof_quality",
		"review_lineage",
		"degradation",
		"recovery",
	} {
		if !strings.Contains(sql, want) {
			t.Fatalf("migration missing %q", want)
		}
	}
}

func TestUpsertContractPersistsIntentProofLink(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("INSERT INTO execution_contracts").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("11111111-1111-1111-1111-111111111111"))

	id, err := NewStore(db).UpsertContract(context.Background(), ContractInput{
		IntentProofID:  "22222222-2222-2222-2222-222222222222",
		TemplateID:     protocol.TemplateChatToProposal,
		ResolvedIntent: "chat-action",
		AuditEventID:   "33333333-3333-3333-3333-333333333333",
	})
	if err != nil {
		t.Fatalf("upsert contract: %v", err)
	}
	if id != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("id = %q", id)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestRecordProofArtifactLinksContractAndUpdatesContract(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("INSERT INTO proof_artifacts").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("44444444-4444-4444-4444-444444444444"))
	mock.ExpectExec("UPDATE execution_contracts").
		WillReturnResult(sqlmock.NewResult(0, 1))

	id, err := NewStore(db).RecordProofArtifact(context.Background(), ProofArtifactInput{
		ContractID:       "11111111-1111-1111-1111-111111111111",
		IntentProofID:    "22222222-2222-2222-2222-222222222222",
		RunID:            "55555555-5555-5555-5555-555555555555",
		AuditEventID:     "33333333-3333-3333-3333-333333333333",
		Status:           protocol.ProofArtifactStatusSuccess,
		ValidationSource: protocol.TrustValidationSourceConfirmAction,
		OutputRefs: []protocol.ExecutionOutput{{
			ID:    "workspace/output.txt",
			Kind:  "file",
			Title: "workspace/output.txt",
		}},
		Recovery: protocol.AuditRecovery{RecoveryState: "verified"},
	})
	if err != nil {
		t.Fatalf("record proof artifact: %v", err)
	}
	if id != "44444444-4444-4444-4444-444444444444" {
		t.Fatalf("id = %q", id)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestListExecutionContractsReturnsDurableContractRecords(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	now := time.Now().UTC()
	mock.ExpectQuery("SELECT id::text, COALESCE\\(intent_proof_id::text").
		WithArgs("", "22222222-2222-2222-2222-222222222222", "", 7).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "intent_proof_id", "run_id", "template_id", "resolved_intent", "execution_shape",
			"status", "execution_status", "validation_source", "evidence_strength", "proof_quality",
			"latest_proof_artifact_id", "audit_event_id", "output_refs", "audit_refs", "review_lineage",
			"degradation", "recovery", "created_at", "confirmed_at", "completed_at", "updated_at",
		}).AddRow(
			"11111111-1111-1111-1111-111111111111",
			"22222222-2222-2222-2222-222222222222",
			"",
			string(protocol.TemplateChatToProposal),
			"ship it",
			string(protocol.ExecutionShapeGuidedProposal),
			string(protocol.ExecutionContractStatusCompleted),
			string(protocol.ExecutionStatusCompleted),
			string(protocol.TrustValidationSourceConfirmAction),
			string(protocol.TrustEvidenceStrengthRunAudit),
			string(protocol.TrustProofQualityVerified),
			"44444444-4444-4444-4444-444444444444",
			"33333333-3333-3333-3333-333333333333",
			[]byte(`[{"id":"workspace/output.md","kind":"file","title":"output"}]`),
			[]byte(`[{"audit_event_id":"33333333-3333-3333-3333-333333333333","source":"log_entries"}]`),
			[]byte(`[{"event":"contract_completed","source":"confirm_action"}]`),
			[]byte(`{}`),
			[]byte(`{"recovery_state":"verified"}`),
			now,
			now,
			now,
			now,
		))

	items, err := NewStore(db).ListExecutionContracts(context.Background(), ListOptions{
		IntentProofID: "22222222-2222-2222-2222-222222222222",
		Limit:         7,
	})
	if err != nil {
		t.Fatalf("list execution contracts: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 contract, got %d", len(items))
	}
	if items[0].LatestProofArtifactID != "44444444-4444-4444-4444-444444444444" {
		t.Fatalf("latest proof = %q", items[0].LatestProofArtifactID)
	}
	if len(items[0].OutputRefs) != 1 || items[0].OutputRefs[0].Title != "output" {
		t.Fatalf("unexpected output refs: %+v", items[0].OutputRefs)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestListProofArtifactsReturnsBoundedProofRecords(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	now := time.Now().UTC()
	mock.ExpectQuery("SELECT id::text, COALESCE\\(contract_id::text").
		WithArgs("11111111-1111-1111-1111-111111111111", "", "", string(protocol.ProofArtifactStatusSuccess), 100).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "contract_id", "intent_proof_id", "run_id", "artifact_kind", "status", "proof_class",
			"validation_source", "evidence_strength", "proof_quality", "output_refs", "audit_refs",
			"review_lineage", "degradation", "recovery", "payload", "created_at",
		}).AddRow(
			"44444444-4444-4444-4444-444444444444",
			"11111111-1111-1111-1111-111111111111",
			"22222222-2222-2222-2222-222222222222",
			"",
			"confirm_action",
			string(protocol.ProofArtifactStatusSuccess),
			string(protocol.ExecutionProofClassRunAudit),
			string(protocol.TrustValidationSourceConfirmAction),
			string(protocol.TrustEvidenceStrengthRunAudit),
			string(protocol.TrustProofQualityVerified),
			[]byte(`[]`),
			[]byte(`[]`),
			[]byte(`[]`),
			[]byte(`{}`),
			[]byte(`{"recovery_state":"verified"}`),
			[]byte(`{"summary":"completed"}`),
			now,
		))

	items, err := NewStore(db).ListProofArtifacts(context.Background(), ListOptions{
		ContractID: "11111111-1111-1111-1111-111111111111",
		Status:     string(protocol.ProofArtifactStatusSuccess),
		Limit:      1000,
	})
	if err != nil {
		t.Fatalf("list proof artifacts: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 proof artifact, got %d", len(items))
	}
	if items[0].Payload["summary"] != "completed" {
		t.Fatalf("payload = %+v", items[0].Payload)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

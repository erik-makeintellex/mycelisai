package trust

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

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

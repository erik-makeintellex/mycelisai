package server

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/configdocuments"
	"github.com/mycelis/core/pkg/protocol"
)

func TestConfigDocumentFixtureResourcesTracksActivationWithoutOwningRevision(t *testing.T) {
	historyID := "22222222-2222-2222-2222-222222222222"
	output, err := json.Marshal(map[string]any{
		"activation": configdocuments.ActivationResult{
			HistoryID:  historyID,
			ToRecordID: "33333333-3333-3333-3333-333333333333",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	resources, err := configDocumentFixtureResources([]plannedToolExecutionResult{{
		Name: "activate_config_document", Output: string(output),
	}})
	if err != nil || len(resources) != 1 {
		t.Fatalf("activation resources = %#v, %v", resources, err)
	}
	if resources[0].Ref != configDocumentActivationFixtureRef(historyID) {
		t.Fatalf("activation ref = %q", resources[0].Ref)
	}
}

func TestClaimConfirmedConfigDocumentsUsesConfirmationTransaction(t *testing.T) {
	withDatabase, mock := withDB(t)
	s := newTestServer(withDatabase)
	scopeID := "11111111-1111-1111-1111-111111111111"
	recordID := "22222222-2222-2222-2222-222222222222"
	now := time.Now().UTC()
	document := protocol.ConfigDocument{
		Metadata: protocol.ConfigDocumentMetadata{Name: "Launch brief", Version: "1.0.0"},
	}
	output, err := json.Marshal(map[string]any{
		"revision": configdocuments.RevisionRecord{RecordID: recordID, Document: document, Digest: "sha256:test"},
	})
	if err != nil {
		t.Fatal(err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id::text, tenant_id, owner_ref, execution_ref, status").
		WithArgs(scopeID, qaFixtureTenantID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "tenant_id", "owner_ref", "execution_ref", "status", "expires_at", "created_at", "updated_at",
		}).AddRow(scopeID, qaFixtureTenantID, "owner", "execution", "open", now.Add(time.Hour), now, now))
	mock.ExpectExec("INSERT INTO qa_fixture_resources").
		WithArgs(sqlmock.AnyArg(), scopeID, "config_document", recordID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	tx, err := s.getDB().BeginTx(t.Context(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.claimConfirmedConfigDocuments(t.Context(), tx, scopeID, []plannedToolExecutionResult{{
		Name: "store_config_document", Output: string(output),
	}}); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestFixtureConfigActivationBaselineRestoresBeforeMultipleFixtureRevisions(t *testing.T) {
	base := "11111111-1111-1111-1111-111111111111"
	first := "22222222-2222-2222-2222-222222222222"
	second := "33333333-3333-3333-3333-333333333333"
	now := time.Now().UTC()
	baseline, changed := fixtureConfigActivationBaseline(second, []fixtureConfigActivation{
		{FromRecord: base, ToRecord: first, CreatedAt: now.Add(-time.Minute)},
		{FromRecord: first, ToRecord: second, CreatedAt: now},
	})
	if !changed || baseline != base {
		t.Fatalf("baseline = %q, changed=%v; want legitimate revision %q", baseline, changed, base)
	}
}

func TestDeleteQAFixtureConfigDocumentsRestoresActivationBeforeDeletingOwnedRevision(t *testing.T) {
	withDatabase, mock := withDB(t)
	s := newTestServer(withDatabase)
	historyID := "22222222-2222-2222-2222-222222222222"
	baseline := "33333333-3333-3333-3333-333333333333"
	target := "44444444-4444-4444-4444-444444444444"
	now := time.Now().UTC()
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id::text, kind, document_id, scope_kind, scope_ref").
		WithArgs(qaFixtureTenantID, historyID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "kind", "document_id", "scope_kind", "scope_ref", "from_record_id", "to_record_id", "created_at",
		}).AddRow(historyID, "OutcomeTemplate", "launch", "workspace", "workspace-1", baseline, target, now))
	mock.ExpectQuery("SELECT id::text FROM config_document_activation_history.*created_at >= \\$6.*FOR UPDATE").
		WithArgs(qaFixtureTenantID, "OutcomeTemplate", "launch", "workspace", "workspace-1", now).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(historyID))
	mock.ExpectQuery("SELECT config_document_record_id::text").
		WithArgs(qaFixtureTenantID, "OutcomeTemplate", "launch", "workspace", "workspace-1").
		WillReturnRows(sqlmock.NewRows([]string{"config_document_record_id"}).AddRow(target))
	mock.ExpectQuery("SELECT id::text FROM config_document_activation_history.*to_record_id=\\$2::uuid").
		WithArgs(qaFixtureTenantID, target).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(historyID))
	mock.ExpectQuery("SELECT kind, document_id, scope_kind, scope_ref FROM config_document_activations").
		WithArgs(qaFixtureTenantID, target).
		WillReturnRows(sqlmock.NewRows([]string{"kind", "document_id", "scope_kind", "scope_ref"}).
			AddRow("OutcomeTemplate", "launch", "workspace", "workspace-1"))
	mock.ExpectExec("UPDATE config_document_activations").
		WithArgs(qaFixtureTenantID, "OutcomeTemplate", "launch", "workspace", "workspace-1", baseline, target).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM config_document_activation_history").
		WithArgs(qaFixtureTenantID, historyID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM config_documents").
		WithArgs(qaFixtureTenantID, target).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	tx, err := s.getDB().BeginTx(t.Context(), nil)
	if err != nil {
		t.Fatal(err)
	}
	deleted := make(map[string]int64)
	err = deleteQAFixtureConfigDocuments(
		t.Context(), tx, qaFixtureTenantID,
		[]string{configDocumentActivationFixtureRef(historyID), target}, deleted,
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	if deleted["config_document_activations_restored"] != 1 ||
		deleted["config_document_activation_history"] != 1 || deleted["config_documents"] != 1 {
		t.Fatalf("deleted rows = %v", deleted)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteQAFixtureConfigDocumentsRefusesLegitimateHistoryAfterFixture(t *testing.T) {
	withDatabase, mock := withDB(t)
	s := newTestServer(withDatabase)
	historyID := "22222222-2222-2222-2222-222222222222"
	legitimateHistoryID := "33333333-3333-3333-3333-333333333333"
	baseline := "44444444-4444-4444-4444-444444444444"
	target := "55555555-5555-5555-5555-555555555555"
	now := time.Now().UTC()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id::text, kind, document_id, scope_kind, scope_ref").
		WithArgs(qaFixtureTenantID, historyID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "kind", "document_id", "scope_kind", "scope_ref", "from_record_id", "to_record_id", "created_at",
		}).AddRow(historyID, "OutcomeTemplate", "launch", "workspace", "workspace-1", baseline, target, now))
	mock.ExpectQuery("SELECT id::text FROM config_document_activation_history.*created_at >= \\$6.*FOR UPDATE").
		WithArgs(qaFixtureTenantID, "OutcomeTemplate", "launch", "workspace", "workspace-1", now).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).
			AddRow(historyID).
			AddRow(legitimateHistoryID))
	mock.ExpectRollback()

	tx, err := s.getDB().BeginTx(t.Context(), nil)
	if err != nil {
		t.Fatal(err)
	}
	deleted := make(map[string]int64)
	err = deleteQAFixtureConfigDocuments(
		t.Context(), tx, qaFixtureTenantID,
		[]string{configDocumentActivationFixtureRef(historyID), target}, deleted,
	)
	if err == nil || !strings.Contains(err.Error(), "unowned history") {
		t.Fatalf("cleanup error = %v, want unowned post-fixture history refusal", err)
	}
	if len(deleted) != 0 {
		t.Fatalf("cleanup wrote deletion counters before refusal: %v", deleted)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteQAFixtureConfigDocumentsRefusesUnownedCurrentActivation(t *testing.T) {
	withDatabase, mock := withDB(t)
	s := newTestServer(withDatabase)
	historyID := "22222222-2222-2222-2222-222222222222"
	baseline := "33333333-3333-3333-3333-333333333333"
	target := "44444444-4444-4444-4444-444444444444"
	legitimateTarget := "55555555-5555-5555-5555-555555555555"
	now := time.Now().UTC()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id::text, kind, document_id, scope_kind, scope_ref").
		WithArgs(qaFixtureTenantID, historyID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "kind", "document_id", "scope_kind", "scope_ref", "from_record_id", "to_record_id", "created_at",
		}).AddRow(historyID, "OutcomeTemplate", "launch", "workspace", "workspace-1", baseline, target, now))
	mock.ExpectQuery("SELECT id::text FROM config_document_activation_history.*created_at >= \\$6.*FOR UPDATE").
		WithArgs(qaFixtureTenantID, "OutcomeTemplate", "launch", "workspace", "workspace-1", now).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(historyID))
	mock.ExpectQuery("SELECT config_document_record_id::text").
		WithArgs(qaFixtureTenantID, "OutcomeTemplate", "launch", "workspace", "workspace-1").
		WillReturnRows(sqlmock.NewRows([]string{"config_document_record_id"}).AddRow(legitimateTarget))
	mock.ExpectRollback()

	tx, err := s.getDB().BeginTx(t.Context(), nil)
	if err != nil {
		t.Fatal(err)
	}
	deleted := make(map[string]int64)
	err = deleteQAFixtureConfigDocuments(
		t.Context(), tx, qaFixtureTenantID,
		[]string{configDocumentActivationFixtureRef(historyID)}, deleted,
	)
	if err == nil || !strings.Contains(err.Error(), "not an owned suffix") {
		t.Fatalf("cleanup error = %v, want unowned current activation refusal", err)
	}
	if len(deleted) != 0 {
		t.Fatalf("cleanup wrote deletion counters before refusal: %v", deleted)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

package configdocuments

import (
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/pkg/protocol"
)

func TestActivateRevisionCreatesFirstScopedActivationAndHistory(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	document := validDocument()
	digest, _ := protocol.CanonicalConfigDocumentDigest(document)
	now := time.Now().UTC()
	expectLockedRevision(mock, "tenant-1", revisionID, document, digest, now)
	mock.ExpectQuery("SELECT config_document_record_id::text FROM config_document_activations.*FOR UPDATE").
		WithArgs("tenant-1", string(document.Kind), document.Metadata.ID, string(document.Metadata.Scope.Kind), document.Metadata.Scope.Ref).
		WillReturnRows(sqlmock.NewRows([]string{"config_document_record_id"}))
	mock.ExpectQuery("INSERT INTO config_document_activations").
		WithArgs("tenant-1", string(document.Kind), document.Metadata.ID, string(document.Metadata.Scope.Kind), document.Metadata.Scope.Ref, revisionID, "operator-2").
		WillReturnRows(sqlmock.NewRows([]string{"activated_at"}).AddRow(now))
	mock.ExpectExec("INSERT INTO config_document_activation_history").
		WithArgs(sqlmock.AnyArg(), "tenant-1", string(document.Kind), document.Metadata.ID, string(document.Metadata.Scope.Kind), document.Metadata.Scope.Ref,
			"", revisionID, string(ActivationActionActivate), "operator-2", auditEventID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	result, err := NewStore(db).ActivateRevision(
		t.Context(), "tenant-1", revisionID, "operator-2", auditEventID, ActivationActionActivate,
	)
	if err != nil {
		t.Fatalf("ActivateRevision: %v", err)
	}
	if result.HistoryID == "" || result.FromRecordID != "" || result.ToRecordID != revisionID || result.Action != ActivationActionActivate {
		t.Fatalf("unexpected activation result: %#v", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestActivateRevisionRollbackReplacesPointerAndRecordsHistory(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	document := validDocument()
	digest, _ := protocol.CanonicalConfigDocumentDigest(document)
	now := time.Now().UTC()
	expectLockedRevision(mock, "tenant-1", revisionID, document, digest, now)
	mock.ExpectQuery("SELECT config_document_record_id::text FROM config_document_activations.*FOR UPDATE").
		WithArgs("tenant-1", string(document.Kind), document.Metadata.ID, string(document.Metadata.Scope.Kind), document.Metadata.Scope.Ref).
		WillReturnRows(sqlmock.NewRows([]string{"config_document_record_id"}).AddRow(previousRevisionID))
	mock.ExpectQuery("INSERT INTO config_document_activations").
		WithArgs("tenant-1", string(document.Kind), document.Metadata.ID, string(document.Metadata.Scope.Kind), document.Metadata.Scope.Ref, revisionID, "operator-2").
		WillReturnRows(sqlmock.NewRows([]string{"activated_at"}).AddRow(now))
	mock.ExpectExec("INSERT INTO config_document_activation_history").
		WithArgs(sqlmock.AnyArg(), "tenant-1", string(document.Kind), document.Metadata.ID, string(document.Metadata.Scope.Kind), document.Metadata.Scope.Ref,
			previousRevisionID, revisionID, string(ActivationActionRollback), "operator-2", auditEventID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	result, err := NewStore(db).ActivateRevision(
		t.Context(), "tenant-1", revisionID, "operator-2", auditEventID, ActivationActionRollback,
	)
	if err != nil {
		t.Fatalf("ActivateRevision: %v", err)
	}
	if result.FromRecordID != previousRevisionID || result.ToRecordID != revisionID || result.Action != ActivationActionRollback {
		t.Fatalf("unexpected rollback result: %#v", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestActivateRevisionRejectsDisabledDocumentBeforePointerMutation(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	document := validDocument()
	document.Metadata.Enabled = false
	digest, _ := protocol.CanonicalConfigDocumentDigest(document)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT .*FROM config_documents.*FOR UPDATE").
		WithArgs("tenant-1", revisionID).
		WillReturnRows(revisionRows(revisionID, "tenant-1", document, digest, "valid", "operator-1", time.Now().UTC()))
	mock.ExpectRollback()

	_, err = NewStore(db).ActivateRevision(
		t.Context(), "tenant-1", revisionID, "operator-2", auditEventID, ActivationActionActivate,
	)
	if !errors.Is(err, ErrDocumentDisabled) {
		t.Fatalf("ActivateRevision error = %v, want ErrDocumentDisabled", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestActivateRevisionTxLeavesCommitToCaller(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	document := validDocument()
	digest, _ := protocol.CanonicalConfigDocumentDigest(document)
	now := time.Now().UTC()
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT .*FROM config_documents.*FOR UPDATE").
		WithArgs("tenant-1", revisionID).
		WillReturnRows(revisionRows(revisionID, "tenant-1", document, digest, "valid", "operator-1", now))
	mock.ExpectQuery("SELECT config_document_record_id::text FROM config_document_activations.*FOR UPDATE").
		WithArgs("tenant-1", string(document.Kind), document.Metadata.ID, string(document.Metadata.Scope.Kind), document.Metadata.Scope.Ref).
		WillReturnRows(sqlmock.NewRows([]string{"config_document_record_id"}))
	mock.ExpectQuery("INSERT INTO config_document_activations").
		WithArgs("tenant-1", string(document.Kind), document.Metadata.ID, string(document.Metadata.Scope.Kind), document.Metadata.Scope.Ref, revisionID, "operator-2").
		WillReturnRows(sqlmock.NewRows([]string{"activated_at"}).AddRow(now))
	mock.ExpectExec("INSERT INTO config_document_activation_history").
		WithArgs(sqlmock.AnyArg(), "tenant-1", string(document.Kind), document.Metadata.ID, string(document.Metadata.Scope.Kind), document.Metadata.Scope.Ref,
			"", revisionID, string(ActivationActionActivate), "operator-2", auditEventID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectRollback()

	tx, err := db.BeginTx(t.Context(), nil)
	if err != nil {
		t.Fatal(err)
	}
	result, err := NewStore(db).ActivateRevisionTx(
		t.Context(), tx, "tenant-1", revisionID, "operator-2", auditEventID, ActivationActionActivate,
	)
	if err != nil || result.ToRecordID != revisionID {
		t.Fatalf("ActivateRevisionTx = %#v, %v", result, err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func expectLockedRevision(mock sqlmock.Sqlmock, tenantID, recordID string, document protocol.ConfigDocument, digest string, now time.Time) {
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT .*FROM config_documents.*FOR UPDATE").
		WithArgs(tenantID, recordID).
		WillReturnRows(revisionRows(recordID, tenantID, document, digest, "valid", "operator-1", now))
}

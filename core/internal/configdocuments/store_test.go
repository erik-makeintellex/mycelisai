package configdocuments

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/pkg/protocol"
)

const (
	revisionID         = "11111111-1111-1111-1111-111111111111"
	previousRevisionID = "22222222-2222-2222-2222-222222222222"
	auditEventID       = "33333333-3333-3333-3333-333333333333"
)

func TestStoreRevisionInsertsValidatedDocumentAndRoundTripsJSON(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	document := validDocument()
	digest, err := protocol.CanonicalConfigDocumentDigest(document)
	if err != nil {
		t.Fatalf("digest: %v", err)
	}
	secretRefs, _ := json.Marshal(document.Metadata.SecretRefs)
	governance, _ := json.Marshal(document.Metadata.Governance)
	now := time.Now().UTC()
	mock.ExpectQuery("INSERT INTO config_documents").
		WithArgs(
			"tenant-1", document.Metadata.ID, document.APIVersion, string(document.Kind),
			document.Metadata.Name, document.Metadata.Version, document.Metadata.OwnerID,
			string(document.Metadata.Scope.Kind), document.Metadata.Scope.Ref, true,
			string(document.Metadata.Source.Kind), document.Metadata.Source.Ref,
			string(secretRefs), string(governance), string(document.Spec), digest, "operator-1",
		).
		WillReturnRows(revisionRows(revisionID, "tenant-1", document, digest, "valid", "operator-1", now))

	record, err := NewStore(db).StoreRevision(t.Context(), "tenant-1", "operator-1", document)
	if err != nil {
		t.Fatalf("StoreRevision: %v", err)
	}
	if record.RecordID != revisionID || record.Digest != digest || record.ValidationState != "valid" {
		t.Fatalf("unexpected record: %#v", record)
	}
	assertDocumentJSONEqual(t, record.Document, document)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestStoreRevisionValidationFailureDoesNotUseSQL(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	document := validDocument()
	document.Metadata.Version = ""
	_, err = NewStore(db).StoreRevision(t.Context(), "tenant-1", "operator-1", document)
	if !errors.Is(err, ErrInvalidDocument) {
		t.Fatalf("StoreRevision error = %v, want ErrInvalidDocument", err)
	}
	var validationErr *ValidationError
	if !errors.As(err, &validationErr) || len(validationErr.Issues) == 0 {
		t.Fatalf("StoreRevision error = %T %v, want validation issues", err, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected SQL: %v", err)
	}
}

func TestStoreRevisionTxUsesCallerTransaction(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	document := validDocument()
	digest, _ := protocol.CanonicalConfigDocumentDigest(document)
	secretRefs, _ := json.Marshal(document.Metadata.SecretRefs)
	governance, _ := json.Marshal(document.Metadata.Governance)
	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO config_documents").WithArgs(
		"tenant-1", document.Metadata.ID, document.APIVersion, string(document.Kind),
		document.Metadata.Name, document.Metadata.Version, document.Metadata.OwnerID,
		string(document.Metadata.Scope.Kind), document.Metadata.Scope.Ref, true,
		string(document.Metadata.Source.Kind), document.Metadata.Source.Ref,
		string(secretRefs), string(governance), string(document.Spec), digest, "operator-1",
	).WillReturnRows(revisionRows(revisionID, "tenant-1", document, digest, "valid", "operator-1", time.Now().UTC()))
	mock.ExpectRollback()

	tx, err := db.BeginTx(t.Context(), nil)
	if err != nil {
		t.Fatal(err)
	}
	record, err := NewStore(db).StoreRevisionTx(t.Context(), tx, "tenant-1", "operator-1", document)
	if err != nil || record.RecordID != revisionID {
		t.Fatalf("StoreRevisionTx = %#v, %v", record, err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetActiveRevisionReloadsValidatedActivation(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	document := validDocument()
	digest, _ := protocol.CanonicalConfigDocumentDigest(document)
	mock.ExpectQuery("FROM config_document_activations activation.*JOIN config_documents document").
		WithArgs(
			"tenant-1", string(document.Kind), document.Metadata.ID,
			string(document.Metadata.Scope.Kind), document.Metadata.Scope.Ref,
		).
		WillReturnRows(revisionRows(revisionID, "tenant-1", document, digest, "valid", "operator-1", time.Now().UTC()))

	// A fresh Store instance represents runtime reload against retained state.
	record, err := NewStore(db).GetActiveRevision(
		t.Context(), "tenant-1", document.Kind, document.Metadata.ID, document.Metadata.Scope,
	)
	if err != nil {
		t.Fatalf("GetActiveRevision: %v", err)
	}
	if record.RecordID != revisionID || record.Digest != digest {
		t.Fatalf("unexpected active revision: %#v", record)
	}
	assertDocumentJSONEqual(t, record.Document, document)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func revisionRows(recordID, tenantID string, document protocol.ConfigDocument, digest, validationState, createdBy string, createdAt time.Time) *sqlmock.Rows {
	secretRefs, _ := json.Marshal(document.Metadata.SecretRefs)
	governance, _ := json.Marshal(document.Metadata.Governance)
	return sqlmock.NewRows([]string{
		"record_id", "tenant_id", "document_id", "api_version", "kind", "name", "version",
		"owner_id", "scope_kind", "scope_ref", "enabled", "source_kind", "source_ref",
		"secret_refs", "governance", "spec", "digest", "validation_state", "created_by", "created_at",
	}).AddRow(
		recordID, tenantID, document.Metadata.ID, document.APIVersion, string(document.Kind), document.Metadata.Name,
		document.Metadata.Version, document.Metadata.OwnerID, string(document.Metadata.Scope.Kind), document.Metadata.Scope.Ref,
		document.Metadata.Enabled, string(document.Metadata.Source.Kind), document.Metadata.Source.Ref,
		string(secretRefs), string(governance), string(document.Spec), digest, validationState, createdBy, createdAt,
	)
}

func validDocument() protocol.ConfigDocument {
	return protocol.ConfigDocument{
		APIVersion: protocol.ConfigDocumentAPIVersionV1,
		Kind:       protocol.ConfigDocumentKindOutcomeTemplate,
		Metadata: protocol.ConfigDocumentMetadata{
			ID:      "browser-package",
			Name:    "Browser package",
			Version: "1.0.0",
			OwnerID: "operator-1",
			Scope: protocol.ConfigDocumentScope{
				Kind: protocol.ConfigDocumentScopeWorkspace,
				Ref:  "workspace-1",
			},
			Enabled: true,
			Source: protocol.ConfigDocumentSource{
				Kind: protocol.ConfigDocumentSourceFile,
				Ref:  "templates/browser-package.json",
			},
			SecretRefs: []string{"env:OPENAI_API_KEY"},
			Governance: protocol.ConfigDocumentGovernance{
				RiskLevel:       protocol.ConfigDocumentRiskMedium,
				ApprovalPosture: protocol.ApprovalPostureRequired,
			},
		},
		Spec: json.RawMessage(`{"deliverable":{"format":"browser_package"},"auth":{"token_ref":"env:OPENAI_API_KEY"}}`),
	}
}

func assertDocumentJSONEqual(t *testing.T, got, want protocol.ConfigDocument) {
	t.Helper()
	gotJSON, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("marshal got document: %v", err)
	}
	wantJSON, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("marshal want document: %v", err)
	}
	var gotValue, wantValue any
	if err := json.Unmarshal(gotJSON, &gotValue); err != nil {
		t.Fatalf("unmarshal got document: %v", err)
	}
	if err := json.Unmarshal(wantJSON, &wantValue); err != nil {
		t.Fatalf("unmarshal want document: %v", err)
	}
	if !reflect.DeepEqual(gotValue, wantValue) {
		t.Fatalf("document round trip mismatch:\n got  %s\n want %s", gotJSON, wantJSON)
	}
}

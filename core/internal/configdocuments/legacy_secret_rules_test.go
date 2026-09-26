package configdocuments

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/pkg/protocol"
)

// legacyAdmittedDocument was valid under the stored secret rules but carries
// text the strict rules now reject (an embedded bearer credential, a raw array
// token, and a URL password). The values are fake fixtures.
func legacyAdmittedDocument() protocol.ConfigDocument {
	document := validDocument()
	document.Spec = json.RawMessage(`{"deliverable":{"format":"browser_package"},"auth":{"token_ref":"env:OPENAI_API_KEY"},` +
		`"notes":"curl -H 'Authorization: Bearer LEGACYFIXTURE0123456789'","args":["--x","ghs_LEGACYFIXTUREabcdef012345"],` +
		`"dsn":"postgres://app:LEGACYFIXTUREpw@db/app"}`)
	return document
}

func TestLegacyAdmittedRevisionStillActivatesAndRollsBack(t *testing.T) {
	document := legacyAdmittedDocument()
	if len(protocol.ValidateNewConfigDocument(document)) == 0 {
		t.Fatal("fixture must be rejected by the strict rules to prove the compatibility path")
	}
	digest, err := protocol.CanonicalConfigDocumentDigest(document)
	if err != nil {
		t.Fatalf("stored-rule digest: %v", err)
	}
	for _, tc := range []struct {
		action ActivationAction
		from   string
	}{{ActivationActionActivate, ""}, {ActivationActionRollback, previousRevisionID}} {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("sqlmock: %v", err)
		}
		now := time.Now().UTC()
		expectLockedRevision(mock, "tenant-1", revisionID, document, digest, now)
		fromRows := sqlmock.NewRows([]string{"config_document_record_id"})
		if tc.from != "" {
			fromRows.AddRow(tc.from)
		}
		mock.ExpectQuery("SELECT config_document_record_id::text FROM config_document_activations.*FOR UPDATE").
			WithArgs("tenant-1", string(document.Kind), document.Metadata.ID, string(document.Metadata.Scope.Kind), document.Metadata.Scope.Ref).
			WillReturnRows(fromRows)
		mock.ExpectQuery("INSERT INTO config_document_activations").
			WithArgs("tenant-1", string(document.Kind), document.Metadata.ID, string(document.Metadata.Scope.Kind), document.Metadata.Scope.Ref, revisionID, "operator-2").
			WillReturnRows(sqlmock.NewRows([]string{"activated_at"}).AddRow(now))
		mock.ExpectExec("INSERT INTO config_document_activation_history").
			WithArgs(sqlmock.AnyArg(), "tenant-1", string(document.Kind), document.Metadata.ID, string(document.Metadata.Scope.Kind), document.Metadata.Scope.Ref,
				tc.from, revisionID, string(tc.action), "operator-2", auditEventID).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		result, err := NewStore(db).ActivateRevision(t.Context(), "tenant-1", revisionID, "operator-2", auditEventID, tc.action)
		if err != nil {
			t.Fatalf("%s legacy-admitted revision: %v", tc.action, err)
		}
		if result.ToRecordID != revisionID || result.FromRecordID != tc.from || result.Action != tc.action {
			t.Fatalf("%s result = %#v", tc.action, result)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("%s unmet expectations: %v", tc.action, err)
		}
		db.Close()
	}
}

func TestLegacyAdmittedRevisionStillReloadsAndCompiles(t *testing.T) {
	document := legacyAdmittedDocument()
	digest, _ := protocol.CanonicalConfigDocumentDigest(document)
	record := RevisionRecord{RecordID: revisionID, Document: document, Digest: digest, ValidationState: "valid"}
	if err := validateStoredRevision(record); err != nil {
		t.Fatalf("stored revision revalidation failed: %v", err)
	}
	if _, err := CompileDocument(document, protocol.MinimumSufficientBrief{}, protocol.MinimumSufficientBrief{}); err != nil {
		t.Fatalf("compile legacy-admitted revision: %v", err)
	}
}

func TestNewWriteOfLegacyAdmittedContentIsRejectedBeforeSQL(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	_, err = NewStore(db).StoreRevision(t.Context(), "tenant-1", "operator-1", legacyAdmittedDocument())
	var validation *ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("StoreRevision error = %v, want strict ValidationError", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("rejected write touched SQL: %v", err)
	}
}

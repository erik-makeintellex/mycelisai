package configdocuments

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

const seedLookupSQL = "SELECT record_id::text, digest FROM config_documents WHERE tenant_id = \\$1 AND document_id = \\$2 AND version = \\$3 AND scope_kind = 'built_in' AND scope_ref = ''"
const seedActiveSQL = "SELECT config_document_record_id::text FROM config_document_activations WHERE tenant_id = \\$1 AND kind = \\$2 AND document_id = \\$3 AND scope_kind = 'built_in' AND scope_ref = ''\\s*$"

func seedDir(t *testing.T, names ...string) string {
	t.Helper()
	dir := t.TempDir()
	for _, name := range names {
		if err := os.WriteFile(filepath.Join(dir, name), readDeliveryPostureFile(t, name), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func rewriteSeed(t *testing.T, dir, name, old, replacement string) {
	t.Helper()
	path := filepath.Join(dir, name)
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), old) {
		t.Fatalf("%s does not contain %q", name, old)
	}
	if err := os.WriteFile(path, []byte(strings.Replace(string(raw), old, replacement, 1)), 0o600); err != nil {
		t.Fatal(err)
	}
}

func loadSeeds(t *testing.T, dir string) []BuiltInSeedDocument {
	t.Helper()
	seeds, present, err := LoadBuiltInSeedDirectory(dir)
	if err != nil || !present {
		t.Fatalf("LoadBuiltInSeedDirectory = %v, %v", present, err)
	}
	return seeds
}

func seedRecordID(i int) string { return fmt.Sprintf("00000000-0000-0000-0000-%012d", i+1) }

func expectSeedStart(mock sqlmock.Sqlmock) {
	mock.ExpectBegin()
	mock.ExpectExec("SELECT pg_advisory_xact_lock").WithArgs(builtInSeedLockKey).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("FROM config_documents WHERE tenant_id = \\$1 AND \\(scope_kind = 'built_in' OR source_kind = 'built_in'\\) AND created_by <> \\$2").
		WithArgs(BuiltInSeedTenant, BuiltInSeedActor).
		WillReturnRows(sqlmock.NewRows([]string{"document_id", "record_id"}))
}

func expectSeedInsertAndActivate(mock sqlmock.Sqlmock, seed BuiltInSeedDocument, recordID, fromRecordID string) {
	document := seed.Document
	now := time.Now().UTC()
	mock.ExpectQuery(seedLookupSQL).WithArgs(BuiltInSeedTenant, document.Metadata.ID, document.Metadata.Version).
		WillReturnRows(sqlmock.NewRows([]string{"record_id", "digest"}))
	secretRefs, _ := json.Marshal(document.Metadata.SecretRefs)
	governance, _ := json.Marshal(document.Metadata.Governance)
	mock.ExpectQuery("INSERT INTO config_documents").WithArgs(
		BuiltInSeedTenant, document.Metadata.ID, document.APIVersion, string(document.Kind),
		document.Metadata.Name, document.Metadata.Version, document.Metadata.OwnerID,
		"built_in", "", true, "built_in", document.Metadata.Source.Ref,
		string(secretRefs), string(governance), sqlmock.AnyArg(), seed.Digest, BuiltInSeedActor,
	).WillReturnRows(revisionRows(recordID, BuiltInSeedTenant, document, seed.Digest, "valid", BuiltInSeedActor, now))
	activeRows := sqlmock.NewRows([]string{"config_document_record_id"})
	if fromRecordID != "" {
		activeRows.AddRow(fromRecordID)
	}
	mock.ExpectQuery(seedActiveSQL).WithArgs(BuiltInSeedTenant, string(document.Kind), document.Metadata.ID).WillReturnRows(activeRows)
	mock.ExpectQuery("SELECT .*FROM config_documents.*FOR UPDATE").WithArgs(BuiltInSeedTenant, recordID).
		WillReturnRows(revisionRows(recordID, BuiltInSeedTenant, document, seed.Digest, "valid", BuiltInSeedActor, now))
	lockedRows := sqlmock.NewRows([]string{"config_document_record_id"})
	if fromRecordID != "" {
		lockedRows.AddRow(fromRecordID)
	}
	mock.ExpectQuery("SELECT config_document_record_id::text FROM config_document_activations.*FOR UPDATE").
		WithArgs(BuiltInSeedTenant, string(document.Kind), document.Metadata.ID, "built_in", "").WillReturnRows(lockedRows)
	mock.ExpectQuery("INSERT INTO config_document_activations").
		WithArgs(BuiltInSeedTenant, string(document.Kind), document.Metadata.ID, "built_in", "", recordID, BuiltInSeedActor).
		WillReturnRows(sqlmock.NewRows([]string{"activated_at"}).AddRow(now))
	mock.ExpectExec("INSERT INTO config_document_activation_history").
		WithArgs(sqlmock.AnyArg(), BuiltInSeedTenant, string(document.Kind), document.Metadata.ID, "built_in", "",
			fromRecordID, recordID, string(ActivationActionActivate), BuiltInSeedActor, "").
		WillReturnResult(sqlmock.NewResult(0, 1))
}

func expectSeedUnchanged(mock sqlmock.Sqlmock, seed BuiltInSeedDocument, recordID, digest string) {
	document := seed.Document
	mock.ExpectQuery(seedLookupSQL).WithArgs(BuiltInSeedTenant, document.Metadata.ID, document.Metadata.Version).
		WillReturnRows(sqlmock.NewRows([]string{"record_id", "digest"}).AddRow(recordID, digest))
}

func TestSeedBuiltInRevisionsFirstBootWritesFourRevisionsAndActivations(t *testing.T) {
	dir := seedDir(t, deliveryPostureFiles...)
	seeds := loadSeeds(t, dir)
	if len(seeds) != 4 {
		t.Fatalf("seeds = %d, want 4", len(seeds))
	}
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	expectSeedStart(mock)
	for i, seed := range seeds {
		expectSeedInsertAndActivate(mock, seed, seedRecordID(i), "")
	}
	mock.ExpectCommit()

	result, err := NewStore(db).SeedBuiltInRevisions(t.Context(), dir)
	if err != nil {
		t.Fatalf("SeedBuiltInRevisions: %v", err)
	}
	if result.Inserted != 4 || result.Activated != 4 || result.Reused != 0 || result.Unchanged != 0 {
		t.Fatalf("result = %#v", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSeedBuiltInRevisionsReseedWritesNothing(t *testing.T) {
	dir := seedDir(t, deliveryPostureFiles...)
	seeds := loadSeeds(t, dir)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	expectSeedStart(mock)
	for i, seed := range seeds {
		expectSeedUnchanged(mock, seed, seedRecordID(i), seed.Digest)
		mock.ExpectQuery(seedActiveSQL).WithArgs(BuiltInSeedTenant, string(seed.Document.Kind), seed.Document.Metadata.ID).
			WillReturnRows(sqlmock.NewRows([]string{"config_document_record_id"}).AddRow(seedRecordID(i)))
	}
	mock.ExpectCommit()

	result, err := NewStore(db).SeedBuiltInRevisions(t.Context(), dir)
	if err != nil {
		t.Fatalf("SeedBuiltInRevisions: %v", err)
	}
	if result.Inserted != 0 || result.Activated != 0 || result.Unchanged != 4 {
		t.Fatalf("reseed result = %#v", result)
	}
	// Any INSERT or activation write would be an unexpected statement here.
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSeedBuiltInRevisionsSameVersionDifferentContentIsFatal(t *testing.T) {
	name := deliveryPostureFiles[0]
	dir := seedDir(t, name)
	seed := loadSeeds(t, dir)[0]
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	expectSeedStart(mock)
	expectSeedUnchanged(mock, seed, seedRecordID(0), "sha256:previously-seeded")
	mock.ExpectRollback()

	_, err = NewStore(db).SeedBuiltInRevisions(t.Context(), dir)
	if err == nil || !strings.Contains(err.Error(), "bump metadata.version") {
		t.Fatalf("error = %v, want bump-version failure", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSeedBuiltInRevisionsNewVersionMovesActivation(t *testing.T) {
	name := deliveryPostureFiles[0]
	dir := seedDir(t, name)
	rewriteSeed(t, dir, name, `version: "1.0.0"`, `version: "1.1.0"`)
	seed := loadSeeds(t, dir)[0]
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	expectSeedStart(mock)
	expectSeedInsertAndActivate(mock, seed, seedRecordID(1), seedRecordID(0))
	mock.ExpectCommit()

	result, err := NewStore(db).SeedBuiltInRevisions(t.Context(), dir)
	if err != nil || result.Inserted != 1 || result.Activated != 1 {
		t.Fatalf("SeedBuiltInRevisions = %#v, %v", result, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSeedBuiltInRevisionsRefusesNonBootstrapBuiltInRows(t *testing.T) {
	dir := seedDir(t, deliveryPostureFiles...)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectExec("SELECT pg_advisory_xact_lock").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("created_by <> \\$2").WithArgs(BuiltInSeedTenant, BuiltInSeedActor).
		WillReturnRows(sqlmock.NewRows([]string{"document_id", "record_id"}).AddRow("default-researcher", revisionID))
	mock.ExpectRollback()

	_, err = NewStore(db).SeedBuiltInRevisions(t.Context(), dir)
	if !errors.Is(err, ErrBuiltInReserved) || !strings.Contains(err.Error(), revisionID) {
		t.Fatalf("error = %v, want ErrBuiltInReserved naming the record", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSeedBuiltInRevisionsTamperedBatchFailsClosedWithoutSQL(t *testing.T) {
	good := deliveryPostureFiles[1]
	target := deliveryPostureFiles[0]
	stem := strings.TrimSuffix(target, ".yaml")
	cases := map[string]func(t *testing.T, dir string){
		"unknown field": func(t *testing.T, dir string) {
			rewriteSeed(t, dir, target, "  enabled: true", "  enabled: true\n  extra: 1")
		},
		"worker profile": func(t *testing.T, dir string) {
			rewriteSeed(t, dir, target, "kind: OutcomeTemplate", "kind: WorkerProfile")
		},
		"org scope": func(t *testing.T, dir string) {
			rewriteSeed(t, dir, target, "kind: built_in\n  enabled", "kind: organization\n    ref: org-1\n  enabled")
		},
		"id not file name": func(t *testing.T, dir string) { rewriteSeed(t, dir, target, "id: "+stem, "id: "+stem+"-x") },
		"wrong source ref": func(t *testing.T, dir string) {
			rewriteSeed(t, dir, target, "ref: core/config", "ref: workspace/config")
		},
		"raw secret": func(t *testing.T, dir string) {
			rewriteSeed(t, dir, target, "  governance:", "  secret_refs: [\"sk-live-abcdefghijklmnop\"]\n  governance:")
		},
		"duplicate id": func(t *testing.T, dir string) {
			raw := readDeliveryPostureFile(t, target)
			if err := os.WriteFile(filepath.Join(dir, "zz-copy.yaml"), raw, 0o600); err != nil {
				t.Fatal(err)
			}
		},
		"file source kind": func(t *testing.T, dir string) {
			rewriteSeed(t, dir, target, "kind: built_in\n    ref:", "kind: file\n    ref:")
		},
	}
	for name, tamper := range cases {
		t.Run(name, func(t *testing.T) {
			dir := seedDir(t, good, target)
			tamper(t, dir)
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			if _, err := NewStore(db).SeedBuiltInRevisions(t.Context(), dir); err == nil {
				t.Fatal("tampered batch seeded")
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("tampered batch reached SQL: %v", err)
			}
		})
	}
}

func TestSeedBuiltInRevisionsMissingDirectorySeedsNothing(t *testing.T) {
	result, err := NewStore(nil).SeedBuiltInRevisions(t.Context(), filepath.Join(t.TempDir(), "absent"))
	if err != nil || result.DirectoryOK || result.Documents != 0 {
		t.Fatalf("missing dir = %#v, %v", result, err)
	}
}

func TestSeedBuiltInRevisionsWithoutDatabaseStillValidates(t *testing.T) {
	dir := seedDir(t, deliveryPostureFiles...)
	if _, err := NewStore(nil).SeedBuiltInRevisions(t.Context(), dir); err == nil || !strings.Contains(err.Error(), "database not available") {
		t.Fatalf("error = %v, want database not available after validation", err)
	}
	rewriteSeed(t, dir, deliveryPostureFiles[0], "kind: OutcomeTemplate", "kind: WorkerProfile")
	if _, err := NewStore(nil).SeedBuiltInRevisions(t.Context(), dir); err == nil || strings.Contains(err.Error(), "database not available") {
		t.Fatalf("error = %v, want validation failure before DB check", err)
	}
}

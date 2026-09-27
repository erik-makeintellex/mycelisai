package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/configdocuments"
)

const shippedBuiltInSeedDir = "../../config/documents/templates"

func TestSeedBuiltInConfigDocumentsValidatesShippedFilesWithoutDatabase(t *testing.T) {
	message, err := seedBuiltInConfigDocumentsFrom(t.Context(), nil, shippedBuiltInSeedDir)
	if err != nil {
		t.Fatalf("seed without DB: %v", err)
	}
	if !strings.Contains(message, "skipped (database unavailable); 5 file(s) validated") {
		t.Fatalf("message = %q", message)
	}
}

func TestSeedBuiltInConfigDocumentsMissingDirectorySeedsNothing(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	message, err := seedBuiltInConfigDocumentsFrom(t.Context(), db, filepath.Join(t.TempDir(), "absent"))
	if err != nil || !strings.Contains(message, "nothing to seed") {
		t.Fatalf("missing dir = %q, %v", message, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSeedBuiltInConfigDocumentsTamperedFileFailsWithoutDatabase(t *testing.T) {
	dir := t.TempDir()
	raw, err := os.ReadFile(filepath.Join(shippedBuiltInSeedDir, "delivery-posture-operations-desk.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	tampered := strings.Replace(string(raw), "kind: OutcomeTemplate", "kind: WorkerProfile", 1)
	if err := os.WriteFile(filepath.Join(dir, "delivery-posture-operations-desk.yaml"), []byte(tampered), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := seedBuiltInConfigDocumentsFrom(t.Context(), nil, dir); err == nil {
		t.Fatal("tampered seed file validated without DB")
	}
}

func TestSeedBuiltInConfigDocumentsRefusesForgedBuiltInRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectExec("SELECT pg_advisory_xact_lock").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("created_by <> \\$2").
		WillReturnRows(sqlmock.NewRows([]string{"document_id", "record_id"}).AddRow("forged", "44444444-4444-4444-4444-444444444444"))
	mock.ExpectRollback()

	_, err = seedBuiltInConfigDocumentsFrom(t.Context(), db, shippedBuiltInSeedDir)
	if !errors.Is(err, configdocuments.ErrBuiltInReserved) {
		t.Fatalf("error = %v, want ErrBuiltInReserved", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

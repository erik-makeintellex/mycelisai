package artifacts

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestArtifactsService_Store(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to mock DB: %v", err)
	}
	defer db.Close()

	svc := NewService(db, "/data/artifacts")
	newID := uuid.New()
	now := time.Now()
	trust := 0.85

	mock.ExpectQuery("INSERT INTO artifacts").
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(),
			"agent-scanner-1", sqlmock.AnyArg(),
			ArtifactType("code"), "main.go", "text/x-go",
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), "pending",
		).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).
			AddRow(newID, now))

	input := Artifact{
		AgentID:      "agent-scanner-1",
		TraceID:      "trace-123",
		ArtifactType: TypeCode,
		Title:        "main.go",
		ContentType:  "text/x-go",
		Content:      "package main\n\nfunc main() {}",
		TrustScore:   &trust,
		Status:       "pending",
	}

	result, err := svc.Store(context.Background(), input)
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}
	if result.ID != newID {
		t.Errorf("Expected ID %s, got %s", newID, result.ID)
	}
	if result.Title != "main.go" {
		t.Errorf("Expected title 'main.go', got %q", result.Title)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet DB expectations: %v", err)
	}
}

func TestArtifactsService_StoreDefaultMetadata(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to mock DB: %v", err)
	}
	defer db.Close()

	svc := NewService(db, "/data/artifacts")
	newID := uuid.New()
	now := time.Now()

	mock.ExpectQuery("INSERT INTO artifacts").
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(),
			"agent-1", sqlmock.AnyArg(),
			ArtifactType("document"), "Report", "text/plain",
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), "pending",
		).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).
			AddRow(newID, now))

	input := Artifact{
		AgentID:      "agent-1",
		ArtifactType: TypeDocument,
		Title:        "Report",
		ContentType:  "text/plain",
		Status:       "pending",
	}

	result, err := svc.Store(context.Background(), input)
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}
	if result.Metadata == nil {
		t.Error("Expected Metadata to be initialized, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet DB expectations: %v", err)
	}
}

func TestArtifactsService_StoreTx(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to mock DB: %v", err)
	}
	defer db.Close()

	svc := NewService(db, "/data/artifacts")
	newID := uuid.New()
	now := time.Now()

	mock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	mock.ExpectQuery("INSERT INTO artifacts").
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(),
			"codex-package", sqlmock.AnyArg(),
			ArtifactType("document"), "Package Proof", "text/markdown",
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), "approved",
		).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).
			AddRow(newID, now))
	mock.ExpectCommit()

	result, err := svc.StoreTx(context.Background(), tx, Artifact{
		AgentID:       "codex-package",
		TraceID:       "run-123",
		ArtifactType:  TypeDocument,
		Title:         "Package Proof",
		ContentType:   "text/markdown",
		Content:       "# Proof",
		FilePath:      "codex-task-packages/myc-rc-001/proof.md",
		FileSizeBytes: 7,
		Status:        "approved",
	})
	if err != nil {
		t.Fatalf("StoreTx failed: %v", err)
	}
	if result.ID != newID {
		t.Fatalf("ID = %s, want %s", result.ID, newID)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit tx: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet DB expectations: %v", err)
	}
}

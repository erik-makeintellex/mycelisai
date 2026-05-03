package artifacts

import (
	"context"
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestArtifactsService_SaveImageToWorkspace(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to mock DB: %v", err)
	}
	defer db.Close()

	svc := NewService(db, "/data/artifacts")
	artID := uuid.New()
	now := time.Now()
	imgB64 := base64.StdEncoding.EncodeToString([]byte("png-bytes"))

	rows := sqlmock.NewRows(artColumns).
		AddRow(artID, nil, nil, "internal", nil, "image",
			"Generated Hero", "image/png", imgB64, nil, nil,
			[]byte(`{"cache_policy":"ephemeral","saved":false}`), nil, "completed", now)
	mock.ExpectQuery("SELECT .+ FROM artifacts\\s+WHERE id = \\$1").
		WithArgs(artID).
		WillReturnRows(rows)
	mock.ExpectExec("UPDATE artifacts").
		WithArgs(sqlmock.AnyArg(), int64(len([]byte("png-bytes"))), artID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	workspace := t.TempDir()
	relPath, err := svc.SaveImageToWorkspace(context.Background(), artID, workspace, "saved-media", "hero")
	if err != nil {
		t.Fatalf("SaveImageToWorkspace failed: %v", err)
	}
	if relPath != "saved-media/hero.png" {
		t.Fatalf("unexpected rel path: %s", relPath)
	}

	fullPath := filepath.Join(workspace, filepath.FromSlash(relPath))
	if _, err := os.Stat(fullPath); err != nil {
		t.Fatalf("expected saved file at %s: %v", fullPath, err)
	}
}

func TestArtifactsService_SaveImageToWorkspace_NonImage(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to mock DB: %v", err)
	}
	defer db.Close()

	svc := NewService(db, "/data/artifacts")
	artID := uuid.New()
	now := time.Now()

	rows := sqlmock.NewRows(artColumns).
		AddRow(artID, nil, nil, "internal", nil, "document",
			"Doc", "text/plain", "hello", nil, nil,
			[]byte(`{}`), nil, "completed", now)
	mock.ExpectQuery("SELECT .+ FROM artifacts\\s+WHERE id = \\$1").
		WithArgs(artID).
		WillReturnRows(rows)

	_, err = svc.SaveImageToWorkspace(context.Background(), artID, t.TempDir(), "", "")
	if err == nil {
		t.Fatal("expected error for non-image artifact")
	}
}

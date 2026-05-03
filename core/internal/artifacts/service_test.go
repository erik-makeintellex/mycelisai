package artifacts

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

var artColumns = []string{
	"id", "mission_id", "team_id", "agent_id", "trace_id", "artifact_type",
	"title", "content_type", "content", "file_path", "file_size_bytes",
	"metadata", "trust_score", "status", "created_at",
}

func TestArtifactsService_ListByMission(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to mock DB: %v", err)
	}
	defer db.Close()

	svc := NewService(db, "/data/artifacts")
	missionID := uuid.New()
	artID := uuid.New()
	now := time.Now()

	rows := sqlmock.NewRows(artColumns).
		AddRow(artID, &missionID, nil, "agent-1", nil, "code",
			"output.py", "text/x-python", "print('hello')", nil, nil,
			[]byte(`{}`), 0.9, "approved", now)

	mock.ExpectQuery("SELECT .+ FROM artifacts WHERE mission_id").
		WithArgs(missionID, 50).
		WillReturnRows(rows)

	results, err := svc.ListByMission(context.Background(), missionID, 0) // 0 defaults to 50
	if err != nil {
		t.Fatalf("ListByMission failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("Expected 1 artifact, got %d", len(results))
	}
	if results[0].Title != "output.py" {
		t.Errorf("Expected title 'output.py', got %q", results[0].Title)
	}
	if results[0].Status != "approved" {
		t.Errorf("Expected status 'approved', got %q", results[0].Status)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet DB expectations: %v", err)
	}
}

func TestArtifactsService_ListByTeam(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to mock DB: %v", err)
	}
	defer db.Close()

	svc := NewService(db, "/data/artifacts")
	teamID := uuid.New()

	rows := sqlmock.NewRows(artColumns)
	mock.ExpectQuery("SELECT .+ FROM artifacts WHERE team_id").
		WithArgs(teamID, 10).
		WillReturnRows(rows)

	results, err := svc.ListByTeam(context.Background(), teamID, 10)
	if err != nil {
		t.Fatalf("ListByTeam failed: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Expected empty list, got %d", len(results))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet DB expectations: %v", err)
	}
}

func TestArtifactsService_ListByAgent(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to mock DB: %v", err)
	}
	defer db.Close()

	svc := NewService(db, "/data/artifacts")
	artID1 := uuid.New()
	artID2 := uuid.New()
	now := time.Now()

	rows := sqlmock.NewRows(artColumns).
		AddRow(artID1, nil, nil, "scanner-1", nil, "image",
			"screenshot.png", "image/png", nil, "/data/artifacts/screenshot.png", int64(204800),
			[]byte(`{"width":1920,"height":1080}`), nil, "pending", now).
		AddRow(artID2, nil, nil, "scanner-1", nil, "data",
			"results.json", "application/json", `{"count":42}`, nil, nil,
			[]byte(`{}`), 0.95, "approved", now)

	mock.ExpectQuery("SELECT .+ FROM artifacts WHERE agent_id").
		WithArgs("scanner-1", 50).
		WillReturnRows(rows)

	results, err := svc.ListByAgent(context.Background(), "scanner-1", 0)
	if err != nil {
		t.Fatalf("ListByAgent failed: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("Expected 2 artifacts, got %d", len(results))
	}
	if results[0].ArtifactType != TypeImage {
		t.Errorf("Expected type 'image', got %q", results[0].ArtifactType)
	}
	if results[0].FileSizeBytes != 204800 {
		t.Errorf("Expected file size 204800, got %d", results[0].FileSizeBytes)
	}

	// Verify metadata was parsed
	var meta map[string]interface{}
	if err := json.Unmarshal(results[0].Metadata, &meta); err != nil {
		t.Errorf("Failed to parse metadata: %v", err)
	}
	if meta["width"] != float64(1920) {
		t.Errorf("Expected width 1920 in metadata, got %v", meta["width"])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet DB expectations: %v", err)
	}
}

func TestArtifactsService_ListRecent(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to mock DB: %v", err)
	}
	defer db.Close()

	svc := NewService(db, "/data/artifacts")

	rows := sqlmock.NewRows(artColumns)
	mock.ExpectQuery("SELECT .+ FROM artifacts ORDER BY created_at DESC").
		WithArgs(25).
		WillReturnRows(rows)

	results, err := svc.ListRecent(context.Background(), 25)
	if err != nil {
		t.Fatalf("ListRecent failed: %v", err)
	}
	if results != nil && len(results) != 0 {
		t.Errorf("Expected empty list, got %d", len(results))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet DB expectations: %v", err)
	}
}

func TestArtifactsService_UpdateStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to mock DB: %v", err)
	}
	defer db.Close()

	svc := NewService(db, "/data/artifacts")
	artID := uuid.New()

	mock.ExpectExec("UPDATE artifacts SET status").
		WithArgs("approved", artID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = svc.UpdateStatus(context.Background(), artID, "approved")
	if err != nil {
		t.Fatalf("UpdateStatus failed: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet DB expectations: %v", err)
	}
}

func TestArtifactsService_UpdateStatusNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to mock DB: %v", err)
	}
	defer db.Close()

	svc := NewService(db, "/data/artifacts")
	artID := uuid.New()

	mock.ExpectExec("UPDATE artifacts SET status").
		WithArgs("rejected", artID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = svc.UpdateStatus(context.Background(), artID, "rejected")
	if err == nil {
		t.Fatal("Expected error for non-existent artifact, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet DB expectations: %v", err)
	}
}

func TestArtifactType_Constants(t *testing.T) {
	types := []struct {
		name     string
		constant ArtifactType
	}{
		{"code", TypeCode},
		{"document", TypeDocument},
		{"image", TypeImage},
		{"audio", TypeAudio},
		{"data", TypeData},
		{"file", TypeFile},
		{"chart", TypeChart},
	}

	for _, tt := range types {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.constant) != tt.name {
				t.Errorf("Expected %q, got %q", tt.name, tt.constant)
			}
		})
	}
}

func TestArtifactsService_DeleteExpiredCachedImages(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to mock DB: %v", err)
	}
	defer db.Close()

	svc := NewService(db, "/data/artifacts")
	mock.ExpectExec("DELETE FROM artifacts").
		WithArgs(int64(3600)).
		WillReturnResult(sqlmock.NewResult(0, 2))

	deleted, err := svc.DeleteExpiredCachedImages(context.Background(), time.Hour)
	if err != nil {
		t.Fatalf("DeleteExpiredCachedImages failed: %v", err)
	}
	if deleted != 2 {
		t.Fatalf("expected 2 deletions, got %d", deleted)
	}
}

package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/artifacts"
)

func TestHandleSaveArtifactToFolder(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to mock DB: %v", err)
	}
	defer db.Close()

	workspace := t.TempDir()
	t.Setenv("MYCELIS_WORKSPACE", workspace)

	artSvc := artifacts.NewService(db, "/data/artifacts")
	s := &AdminServer{Artifacts: artSvc}

	artID := "dddddddd-dddd-dddd-dddd-dddddddddddd"
	now := time.Now()
	rows := sqlmock.NewRows(artTestColumns).
		AddRow(artID, nil, nil, "internal", nil, "image",
			"Generated", "image/png", "cG5n", nil, nil,
			[]byte(`{"cache_policy":"ephemeral","saved":false}`), nil, "completed", now)
	mock.ExpectQuery("SELECT .+ FROM artifacts\\s+WHERE id = \\$1").
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(rows)
	mock.ExpectExec("UPDATE artifacts").
		WithArgs(sqlmock.AnyArg(), int64(3), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/artifacts/{id}/save", s.handleSaveArtifactToFolder)

	req, _ := http.NewRequest("POST", "/api/v1/artifacts/"+artID+"/save", strings.NewReader(`{"folder":"saved-media","filename":"test-image"}`))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d. Body: %s", rr.Code, rr.Body.String())
	}

	var payload map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload["file_path"] == "" {
		t.Fatalf("expected file_path in response")
	}

	savedPath := filepath.Join(workspace, filepath.FromSlash(payload["file_path"]))
	if _, err := os.Stat(savedPath); err != nil {
		t.Fatalf("expected saved file to exist: %v", err)
	}
}

func TestHandleDownloadArtifact_FileBacked(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to mock DB: %v", err)
	}
	defer db.Close()

	workspace := t.TempDir()
	t.Setenv("MYCELIS_WORKSPACE", workspace)
	relativePath := filepath.ToSlash(filepath.Join("saved-media", "test-image.png"))
	absolutePath := filepath.Join(workspace, "saved-media", "test-image.png")
	if err := os.MkdirAll(filepath.Dir(absolutePath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(absolutePath, []byte("png"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	artSvc := artifacts.NewService(db, "/data/artifacts")
	s := &AdminServer{Artifacts: artSvc}

	artID := "eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee"
	now := time.Now()
	rows := sqlmock.NewRows(artTestColumns).
		AddRow(artID, nil, nil, "internal", nil, "image",
			"Generated", "image/png", nil, relativePath, int64(3),
			[]byte(`{"saved":true}`), nil, "completed", now)
	mock.ExpectQuery("SELECT .+ FROM artifacts\\s+WHERE id = \\$1").
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(rows)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/artifacts/{id}/download", s.handleDownloadArtifact)

	req, _ := http.NewRequest("GET", "/api/v1/artifacts/"+artID+"/download", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d. Body: %s", rr.Code, rr.Body.String())
	}
	if got := rr.Header().Get("Content-Disposition"); !strings.Contains(got, `attachment; filename="test-image.png"`) {
		t.Fatalf("unexpected Content-Disposition: %q", got)
	}
	if body := rr.Body.String(); body != "png" {
		t.Fatalf("expected downloaded body %q, got %q", "png", body)
	}
}

func TestHandleDownloadArtifact_InlineDocument(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to mock DB: %v", err)
	}
	defer db.Close()

	artSvc := artifacts.NewService(db, "/data/artifacts")
	s := &AdminServer{Artifacts: artSvc}

	artID := "ffffffff-ffff-ffff-ffff-ffffffffffff"
	now := time.Now()
	rows := sqlmock.NewRows(artTestColumns).
		AddRow(artID, nil, nil, "internal", nil, "document",
			"brief.md", "text/markdown", "# Brief", nil, nil,
			[]byte(`{}`), nil, "completed", now)
	mock.ExpectQuery("SELECT .+ FROM artifacts\\s+WHERE id = \\$1").
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(rows)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/artifacts/{id}/download", s.handleDownloadArtifact)

	req, _ := http.NewRequest("GET", "/api/v1/artifacts/"+artID+"/download", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d. Body: %s", rr.Code, rr.Body.String())
	}
	if got := rr.Header().Get("Content-Disposition"); !strings.Contains(got, `attachment; filename="brief.md"`) {
		t.Fatalf("unexpected Content-Disposition: %q", got)
	}
	if body := rr.Body.String(); body != "# Brief" {
		t.Fatalf("expected inline document body %q, got %q", "# Brief", body)
	}
}

package deploymentcontext

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/artifacts"
	"github.com/mycelis/core/internal/memory"
)

func TestIngest_ChunkInsertFailureRollsBackArtifact(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO artifacts").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", time.Now()))
	mock.ExpectQuery("INSERT INTO context_vectors").WillReturnError(errors.New("disk full"))
	mock.ExpectRollback()

	svc := NewService(artifacts.NewService(db, t.TempDir()), memory.NewServiceWithDB(db), nil)
	result, err := svc.Ingest(context.Background(), IngestRequest{Title: "Juniper & Rye Bakery", Content: bakeryContent})
	if err == nil || result != nil {
		t.Fatalf("a failed transaction must fail the save: result=%+v err=%v", result, err)
	}
	var saveErr *SaveError
	if !errors.As(err, &saveErr) || saveErr.Code != "deployment_context_save_failed" {
		t.Fatalf("failure must carry a normalized code, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("artifact insert must roll back with the chunk rows: %v", err)
	}
}

func TestNormalize_TaxonomyAliasesMapToWorklog(t *testing.T) {
	if got := normalizeSourceKind("diary_entry"); got != "worklog_entry" {
		t.Fatalf("diary_entry alias -> %q", got)
	}
	if got := normalizeSourceKind("worklog_entry"); got != "worklog_entry" {
		t.Fatalf("worklog_entry -> %q", got)
	}
	if got := normalizeContentDomain("diary"); got != "worklog" {
		t.Fatalf("diary alias -> %q", got)
	}
	if got := normalizeContentDomain("worklog"); got != "worklog" {
		t.Fatalf("worklog -> %q", got)
	}
	tags := normalizeTags([]string{"diary", "Diary_Entry", "worklog"})
	if len(tags) != 2 || tags[0] != "worklog" || tags[1] != "worklog_entry" {
		t.Fatalf("tag aliases must normalize and dedupe: %v", tags)
	}
}

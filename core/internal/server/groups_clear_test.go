package server

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestHandleClearGroup_ArchivesStandingGroupAndKeepsOutputs(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)
	mux := setupMux(t, "POST /api/v1/groups/{id}/clear", s.HandleClearGroup)

	now := time.Now().UTC()
	expectGroupLoad(mock, "group-keep", "active", now)
	mock.ExpectExec("INSERT INTO log_entries").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE collaboration_groups").
		WithArgs("group-keep", groupStatusArchived, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	expectGroupLoad(mock, "group-keep", groupStatusArchived, now)

	rr := doAuthenticatedRequest(t, mux, http.MethodPost, "/api/v1/groups/group-keep/clear", `{"include_outputs":false}`)
	assertStatus(t, rr, http.StatusOK)

	var payload map[string]any
	assertJSON(t, rr, &payload)
	data := payload["data"].(map[string]any)
	if data["outputs_cleared"] != false {
		t.Fatalf("outputs_cleared = %v, want false", data["outputs_cleared"])
	}
	group := data["group"].(map[string]any)
	if group["status"] != groupStatusArchived {
		t.Fatalf("status = %v, want %q", group["status"], groupStatusArchived)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestHandleClearGroup_WithOutputsRemovesWorkspaceAndArchivesArtifacts(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)
	mux := setupMux(t, "POST /api/v1/groups/{id}/clear", s.HandleClearGroup)
	root := t.TempDir()
	t.Setenv("MYCELIS_WORKSPACE", root)
	outputDir := filepath.Join(root, "groups", "cleanup-team", "generated")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(outputDir, "index.html"), []byte("<html></html>"), 0o644); err != nil {
		t.Fatalf("write output: %v", err)
	}

	now := time.Now().UTC()
	expectGroupLoad(mock, "group-clear", "active", now)
	mock.ExpectExec("INSERT INTO log_entries").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE collaboration_groups").
		WithArgs("group-clear", groupStatusArchived, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	expectGroupLoad(mock, "group-clear", groupStatusArchived, now)
	mock.ExpectExec("UPDATE artifacts SET status").
		WithArgs("archived", "cleanup-lead").
		WillReturnResult(sqlmock.NewResult(0, 2))

	rr := doAuthenticatedRequest(t, mux, http.MethodPost, "/api/v1/groups/group-clear/clear", `{"include_outputs":true}`)
	assertStatus(t, rr, http.StatusOK)

	var payload map[string]any
	assertJSON(t, rr, &payload)
	data := payload["data"].(map[string]any)
	if data["outputs_cleared"] != true || data["workspace_removed"] != true {
		t.Fatalf("clear result = %#v", data)
	}
	if data["artifacts_archived"] != float64(2) {
		t.Fatalf("artifacts_archived = %v, want 2", data["artifacts_archived"])
	}
	if _, err := os.Stat(filepath.Join(root, "groups", "cleanup-team")); !os.IsNotExist(err) {
		t.Fatalf("group workspace still exists or stat failed: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func expectGroupLoad(mock sqlmock.Sqlmock, id, status string, now time.Time) {
	mock.ExpectQuery("SELECT id::text, tenant_id, name, goal_statement, work_mode").
		WithArgs(id).
		WillReturnRows(sqlmock.NewRows(collaborationGroupColumns()).
			AddRow(
				id,
				"default",
				"Cleanup Team",
				"Produce and review retained outputs",
				"propose_only",
				[]byte(`["write_file"]`),
				[]byte(`["owner"]`),
				[]byte(`["cleanup-lead"]`),
				"groups/cleanup-team",
				"team-lead",
				"policy.default",
				status,
				"test-user-001",
				nil,
				"33333333-3333-3333-3333-333333333333",
				"33333333-3333-3333-3333-333333333333",
				now,
				now,
			))
}

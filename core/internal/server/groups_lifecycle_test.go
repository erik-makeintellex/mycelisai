package server

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestHandleGroupLifecycleReport_ClassifiesGroupReviewState(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)
	mux := setupMux(t, "GET /api/v1/groups/lifecycle", s.HandleGroupLifecycleReport)

	now := time.Now().UTC()
	expiredAt := now.Add(-2 * time.Hour)
	createdAt := now.Add(-5 * 24 * time.Hour)
	updatedAt := now.Add(-4 * 24 * time.Hour)

	mock.ExpectQuery("SELECT id::text, tenant_id, name, goal_statement, work_mode").
		WillReturnRows(sqlmock.NewRows(collaborationGroupColumns()).
			AddRow("group-expired", "default", "Expired QA Team", "Temporary proof run", "propose_only", []byte(`[]`), []byte(`[]`), []byte(`["qa-team"]`), "groups/qa-team", "lead", "standard", "active", "admin", expiredAt, "", "", createdAt, updatedAt).
			AddRow("group-work", "default", "Blocked Runtime Team", "Recover blocked work", "propose_only", []byte(`[]`), []byte(`[]`), []byte(`["runtime-team"]`), "groups/runtime-team", "lead", "standard", "active", "admin", nil, "", "", createdAt, updatedAt).
			AddRow("group-idle-output", "default", "Game Output Team", "Retained playable output", "propose_only", []byte(`[]`), []byte(`[]`), []byte(`["game-team"]`), "groups/game-team", "lead", "standard", "active", "admin", nil, "", "", createdAt, updatedAt).
			AddRow("group-stale", "default", "Old Standing Team", "No output yet", "propose_only", []byte(`[]`), []byte(`[]`), []byte(`["old-team"]`), "groups/old-team", "lead", "standard", "active", "admin", nil, "", "", createdAt, updatedAt))
	mock.ExpectQuery("SELECT\\s+COUNT\\(\\*\\)::int").
		WithArgs("qa-team").
		WillReturnRows(sqlmock.NewRows([]string{"count", "active", "output", "archived", "latest"}).
			AddRow(0, 0, 0, 0, nil))
	mock.ExpectQuery("SELECT\\s+COUNT\\(\\*\\)::int").
		WithArgs("runtime-team").
		WillReturnRows(sqlmock.NewRows([]string{"count", "active", "output", "archived", "latest"}).
			AddRow(2, 2, 0, 0, now.Add(-time.Hour)))
	mock.ExpectQuery("SELECT\\s+COUNT\\(\\*\\)::int").
		WithArgs("game-team").
		WillReturnRows(sqlmock.NewRows([]string{"count", "active", "output", "archived", "latest"}).
			AddRow(1, 0, 1, 0, now.Add(-time.Hour)))
	mock.ExpectQuery("SELECT\\s+COUNT\\(\\*\\)::int").
		WithArgs("old-team").
		WillReturnRows(sqlmock.NewRows([]string{"count", "active", "output", "archived", "latest"}).
			AddRow(0, 0, 0, 0, nil))

	rr := doAuthenticatedRequest(t, mux, http.MethodGet, "/api/v1/groups/lifecycle", "")
	assertStatus(t, rr, http.StatusOK)

	var payload struct {
		Data groupLifecycleReport `json:"data"`
	}
	assertJSON(t, rr, &payload)
	if payload.Data.Summary.ExpiredActiveGroups != 1 {
		t.Fatalf("expired groups = %d, want 1", payload.Data.Summary.ExpiredActiveGroups)
	}
	if payload.Data.Summary.TeamWorkNeedingAttention != 2 {
		t.Fatalf("team work needing attention = %d, want 2", payload.Data.Summary.TeamWorkNeedingAttention)
	}
	want := map[string]string{
		"group-expired":     "archive_expired",
		"group-work":        "review_work",
		"group-idle-output": "archive_completed",
		"group-stale":       "review_standing",
	}
	for _, item := range payload.Data.Items {
		if got := item.Recommendation; want[item.GroupID] != got {
			t.Fatalf("%s recommendation = %q, want %q", item.GroupID, got, want[item.GroupID])
		}
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestArchiveExpiredGroupsDB_OnlyArchivesExpiredActiveGroups(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)

	mock.ExpectQuery("UPDATE collaboration_groups").
		WithArgs(groupStatusArchived, sqlmock.AnyArg(), groupStatusActive).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).
			AddRow("group-expired-a").
			AddRow("group-expired-b"))

	ids, err := s.archiveExpiredGroupsDB(context.Background(), "11111111-1111-1111-1111-111111111111")
	if err != nil {
		t.Fatalf("archive expired groups: %v", err)
	}
	if len(ids) != 2 {
		t.Fatalf("archived ids = %v, want 2 ids", ids)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

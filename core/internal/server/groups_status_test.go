package server

import (
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestHandleUpdateGroupStatus_ArchivesTemporaryGroup(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)
	mux := setupMux(t, "PATCH /api/v1/groups/{id}/status", s.HandleUpdateGroupStatus)

	now := time.Now().UTC()
	auditID := "33333333-3333-3333-3333-333333333333"
	mock.ExpectExec("INSERT INTO log_entries").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE collaboration_groups").
		WithArgs("group-temp", groupStatusArchived, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("SELECT id::text, tenant_id, name, goal_statement, work_mode").
		WithArgs("group-temp").
		WillReturnRows(sqlmock.NewRows(collaborationGroupColumns()).
			AddRow(
				"group-temp",
				"default",
				"Temp Campaign",
				"Produce one campaign package",
				"execute_with_approval",
				[]byte(`["write_file"]`),
				[]byte(`["owner"]`),
				[]byte(`["11111111-1111-1111-1111-111111111111"]`),
				"marketing-lead",
				"",
				groupStatusArchived,
				"test-user-001",
				now.Add(2*time.Hour),
				auditID,
				auditID,
				now,
				now,
			))

	rr := doAuthenticatedRequest(t, mux, "PATCH", "/api/v1/groups/group-temp/status", `{"status":"archived"}`)
	assertStatus(t, rr, http.StatusOK)

	var payload map[string]any
	assertJSON(t, rr, &payload)
	data, ok := payload["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected data map, got %T", payload["data"])
	}
	if data["status"] != groupStatusArchived {
		t.Fatalf("status = %v, want %q", data["status"], groupStatusArchived)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestHandleUpdateGroupStatus_ArchivesGroup(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)
	mux := setupMux(t, "PATCH /api/v1/groups/{id}/status", s.HandleUpdateGroupStatus)

	now := time.Now().UTC()
	auditID := "33333333-3333-3333-3333-333333333333"

	mock.ExpectExec("INSERT INTO log_entries").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE collaboration_groups").
		WithArgs("group-temp", groupStatusArchived, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("SELECT id::text, tenant_id, name, goal_statement, work_mode").
		WithArgs("group-temp").
		WillReturnRows(sqlmock.NewRows(collaborationGroupColumns()).
			AddRow(
				"group-temp",
				"default",
				"Launch lane",
				"coordinate launch follow-through",
				"propose_only",
				[]byte(`["runs.read"]`),
				[]byte(`["u1"]`),
				[]byte(`["11111111-1111-1111-1111-111111111111"]`),
				"launch-profile",
				"policy.launch",
				groupStatusArchived,
				"test-user-001",
				nil,
				auditID,
				auditID,
				now,
				now,
			))

	rr := doAuthenticatedRequest(t, mux, "PATCH", "/api/v1/groups/group-temp/status", `{"status":"archived"}`)
	assertStatus(t, rr, http.StatusOK)

	var payload map[string]any
	assertJSON(t, rr, &payload)
	data, ok := payload["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected data object, got %T", payload["data"])
	}
	if got := data["status"]; got != groupStatusArchived {
		t.Fatalf("status = %v, want %s", got, groupStatusArchived)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestHandleUpdateGroupStatus_InvalidStatus(t *testing.T) {
	s := newTestServer()
	mux := setupMux(t, "PATCH /api/v1/groups/{id}/status", s.HandleUpdateGroupStatus)

	rr := doAuthenticatedRequest(t, mux, "PATCH", "/api/v1/groups/group-temp/status", `{"status":"gone"}`)
	assertStatus(t, rr, http.StatusBadRequest)
}

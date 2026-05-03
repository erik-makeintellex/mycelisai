package mcp

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func newTestToolSetService(t *testing.T) (*ToolSetService, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return NewToolSetService(db), mock
}

func toolSetColumns() []string {
	return []string{"id", "name", "description", "tool_refs", "tenant_id", "created_at", "updated_at"}
}

var (
	tsID1 = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	tsID2 = uuid.MustParse("22222222-2222-2222-2222-222222222222")
)

func TestToolSetService_List_HappyPath(t *testing.T) {
	svc, mock := newTestToolSetService(t)
	now := time.Now()

	rows := sqlmock.NewRows(toolSetColumns()).
		AddRow(tsID1, "workspace", "File I/O", `["mcp:filesystem/*"]`, "default", now, now).
		AddRow(tsID2, "research", "Web tools", `["mcp:fetch/*"]`, "default", now, now)
	mock.ExpectQuery("SELECT .+ FROM mcp_tool_sets").WillReturnRows(rows)

	sets, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(sets) != 2 {
		t.Fatalf("got %d sets, want 2", len(sets))
	}
	if sets[0].Name != "workspace" {
		t.Errorf("sets[0].Name = %q", sets[0].Name)
	}
	if len(sets[0].ToolRefs) != 1 || sets[0].ToolRefs[0] != "mcp:filesystem/*" {
		t.Errorf("sets[0].ToolRefs = %v", sets[0].ToolRefs)
	}
}

func TestToolSetService_List_Empty(t *testing.T) {
	svc, mock := newTestToolSetService(t)

	mock.ExpectQuery("SELECT .+ FROM mcp_tool_sets").
		WillReturnRows(sqlmock.NewRows(toolSetColumns()))

	sets, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	// nil or empty — both acceptable, document behavior
	if sets != nil && len(sets) != 0 {
		t.Errorf("expected nil or empty, got %d", len(sets))
	}
}

func TestToolSetService_Get_Found(t *testing.T) {
	svc, mock := newTestToolSetService(t)
	now := time.Now()

	mock.ExpectQuery("SELECT .+ FROM mcp_tool_sets WHERE id").
		WithArgs(tsID1).
		WillReturnRows(sqlmock.NewRows(toolSetColumns()).
			AddRow(tsID1, "workspace", "File I/O", `["mcp:filesystem/*"]`, "default", now, now))

	ts, err := svc.Get(context.Background(), tsID1)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if ts == nil {
		t.Fatal("expected non-nil")
	}
	if ts.Name != "workspace" {
		t.Errorf("Name = %q", ts.Name)
	}
}

func TestToolSetService_Get_NotFound(t *testing.T) {
	svc, mock := newTestToolSetService(t)

	mock.ExpectQuery("SELECT .+ FROM mcp_tool_sets WHERE id").
		WithArgs(tsID1).
		WillReturnError(sql.ErrNoRows)

	ts, err := svc.Get(context.Background(), tsID1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ts != nil {
		t.Error("expected nil for not found")
	}
}

func TestToolSetService_Create_HappyPath(t *testing.T) {
	svc, mock := newTestToolSetService(t)
	now := time.Now()

	mock.ExpectQuery("INSERT INTO mcp_tool_sets").
		WithArgs("development", "Dev tools", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow(tsID1, now, now))

	ts := ToolSet{Name: "development", Description: "Dev tools", ToolRefs: []string{"mcp:filesystem/*"}}
	created, err := svc.Create(context.Background(), ts)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.ID != tsID1 {
		t.Errorf("ID = %v", created.ID)
	}
	if created.TenantID != "default" {
		t.Errorf("TenantID = %q, want default", created.TenantID)
	}
}

func TestToolSetService_Create_NilRefs(t *testing.T) {
	svc, mock := newTestToolSetService(t)
	now := time.Now()

	mock.ExpectQuery("INSERT INTO mcp_tool_sets").
		WithArgs("empty", "", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow(tsID1, now, now))

	ts := ToolSet{Name: "empty", ToolRefs: nil}
	created, err := svc.Create(context.Background(), ts)
	if err != nil {
		t.Fatalf("Create with nil refs: %v", err)
	}
	// ToolRefs should default to empty slice
	if created.ToolRefs == nil {
		t.Error("expected non-nil ToolRefs after create")
	}
}

func TestToolSetService_Update_Found(t *testing.T) {
	svc, mock := newTestToolSetService(t)
	now := time.Now()

	mock.ExpectQuery("UPDATE mcp_tool_sets SET").
		WithArgs("updated", "New desc", sqlmock.AnyArg(), tsID1).
		WillReturnRows(sqlmock.NewRows(toolSetColumns()).
			AddRow(tsID1, "updated", "New desc", `["mcp:github/*"]`, "default", now, now))

	ts := ToolSet{Name: "updated", Description: "New desc", ToolRefs: []string{"mcp:github/*"}}
	result, err := svc.Update(context.Background(), tsID1, ts)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if result.Name != "updated" {
		t.Errorf("Name = %q", result.Name)
	}
}

func TestToolSetService_Update_NotFound(t *testing.T) {
	svc, mock := newTestToolSetService(t)

	mock.ExpectQuery("UPDATE mcp_tool_sets SET").
		WithArgs("x", "", sqlmock.AnyArg(), tsID1).
		WillReturnError(sql.ErrNoRows)

	_, err := svc.Update(context.Background(), tsID1, ToolSet{Name: "x"})
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestToolSetService_Delete_Found(t *testing.T) {
	svc, mock := newTestToolSetService(t)

	mock.ExpectExec("DELETE FROM mcp_tool_sets WHERE id").
		WithArgs(tsID1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := svc.Delete(context.Background(), tsID1); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

func TestToolSetService_Delete_NotFound(t *testing.T) {
	svc, mock := newTestToolSetService(t)

	mock.ExpectExec("DELETE FROM mcp_tool_sets WHERE id").
		WithArgs(tsID1).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := svc.Delete(context.Background(), tsID1)
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestToolSetService_FindByName_Found(t *testing.T) {
	svc, mock := newTestToolSetService(t)
	now := time.Now()

	mock.ExpectQuery("SELECT .+ FROM mcp_tool_sets WHERE name").
		WithArgs("workspace").
		WillReturnRows(sqlmock.NewRows(toolSetColumns()).
			AddRow(tsID1, "workspace", "File I/O", `["mcp:filesystem/*"]`, "default", now, now))

	ts, err := svc.FindByName(context.Background(), "workspace")
	if err != nil {
		t.Fatalf("FindByName: %v", err)
	}
	if ts == nil {
		t.Fatal("expected non-nil")
	}
	if ts.Name != "workspace" {
		t.Errorf("Name = %q", ts.Name)
	}
}

func TestToolSetService_FindByName_NotFound(t *testing.T) {
	svc, mock := newTestToolSetService(t)

	mock.ExpectQuery("SELECT .+ FROM mcp_tool_sets WHERE name").
		WithArgs("nonexistent").
		WillReturnError(sql.ErrNoRows)

	ts, err := svc.FindByName(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ts != nil {
		t.Error("expected nil for not found")
	}
}

package mcp

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

// ── Helpers ──────────────────────────────────────────────────

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

// ── List ─────────────────────────────────────────────────────

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

// ── Get ──────────────────────────────────────────────────────

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

// ── Create ───────────────────────────────────────────────────

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

// ── Update ───────────────────────────────────────────────────

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

// ── Delete ───────────────────────────────────────────────────

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

// ── FindByName ───────────────────────────────────────────────

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

// ── ResolveRefs ──────────────────────────────────────────────

func TestToolSetService_ResolveRefs_Mixed(t *testing.T) {
	svc, mock := newTestToolSetService(t)
	now := time.Now()

	// toolset:workspace → expand to mcp:filesystem/*
	mock.ExpectQuery("SELECT .+ FROM mcp_tool_sets WHERE name").
		WithArgs("workspace").
		WillReturnRows(sqlmock.NewRows(toolSetColumns()).
			AddRow(tsID1, "workspace", "File I/O", `["mcp:filesystem/*"]`, "default", now, now))

	tools := []string{"read_file", "mcp:github/*", "toolset:workspace", "consult_council"}
	resolved, err := svc.ResolveRefs(context.Background(), tools)
	if err != nil {
		t.Fatalf("ResolveRefs: %v", err)
	}
	// Expected: read_file, mcp:github/*, mcp:filesystem/*, consult_council
	if len(resolved) != 4 {
		t.Fatalf("got %d refs, want 4: %v", len(resolved), resolved)
	}
	if resolved[0] != "read_file" {
		t.Errorf("resolved[0] = %q", resolved[0])
	}
	if resolved[1] != "mcp:github/*" {
		t.Errorf("resolved[1] = %q", resolved[1])
	}
	if resolved[2] != "mcp:filesystem/*" {
		t.Errorf("resolved[2] = %q", resolved[2])
	}
	if resolved[3] != "consult_council" {
		t.Errorf("resolved[3] = %q", resolved[3])
	}
}

func TestToolSetService_ResolveRefs_MissingToolSet(t *testing.T) {
	svc, mock := newTestToolSetService(t)

	mock.ExpectQuery("SELECT .+ FROM mcp_tool_sets WHERE name").
		WithArgs("unknown").
		WillReturnError(sql.ErrNoRows)

	tools := []string{"mcp:filesystem/*", "toolset:unknown"}
	resolved, err := svc.ResolveRefs(context.Background(), tools)
	if err != nil {
		t.Fatalf("ResolveRefs: %v", err)
	}
	// toolset:unknown silently skipped
	if len(resolved) != 1 {
		t.Fatalf("got %d refs, want 1: %v", len(resolved), resolved)
	}
}

func TestToolSetService_ResolveRefs_NoToolSets(t *testing.T) {
	svc, _ := newTestToolSetService(t)

	tools := []string{"mcp:filesystem/*", "read_file", "mcp:github/create_issue"}
	resolved, err := svc.ResolveRefs(context.Background(), tools)
	if err != nil {
		t.Fatalf("ResolveRefs: %v", err)
	}
	if len(resolved) != 3 {
		t.Fatalf("got %d, want 3", len(resolved))
	}
}

// ── NilDB ────────────────────────────────────────────────────

func TestToolSetService_NilDB(t *testing.T) {
	svc := NewToolSetService(nil)

	if _, err := svc.List(context.Background()); err == nil {
		t.Error("List: expected error with nil DB")
	}
	if _, err := svc.Get(context.Background(), tsID1); err == nil {
		t.Error("Get: expected error with nil DB")
	}
	if _, err := svc.Create(context.Background(), ToolSet{Name: "x"}); err == nil {
		t.Error("Create: expected error with nil DB")
	}
	if _, err := svc.Update(context.Background(), tsID1, ToolSet{Name: "x"}); err == nil {
		t.Error("Update: expected error with nil DB")
	}
	if err := svc.Delete(context.Background(), tsID1); err == nil {
		t.Error("Delete: expected error with nil DB")
	}
	if _, err := svc.FindByName(context.Background(), "x"); err == nil {
		t.Error("FindByName: expected error with nil DB")
	}
}

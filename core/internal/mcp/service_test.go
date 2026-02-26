package mcp

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

// ── Helpers ──────────────────────────────────────────────────

func newTestService(t *testing.T) (*Service, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return NewService(db), mock
}

func serverColumns() []string {
	return []string{"id", "name", "transport", "command", "args", "env", "url", "headers", "status", "error_message", "created_at", "updated_at"}
}

func toolColumns() []string {
	return []string{"id", "server_id", "name", "description", "input_schema"}
}

func toolWithServerColumns() []string {
	return []string{"id", "server_id", "server_name", "name", "description", "input_schema"}
}

var (
	testServerID = uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	testToolID   = uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
)

// ── Install ──────────────────────────────────────────────────

func TestService_Install_HappyPath(t *testing.T) {
	svc, mock := newTestService(t)
	now := time.Now()

	mock.ExpectQuery("INSERT INTO mcp_servers").
		WithArgs("filesystem", "stdio", "npx", sqlmock.AnyArg(), sqlmock.AnyArg(), "", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows(serverColumns()).
			AddRow(testServerID, "filesystem", "stdio", "npx",
				`["-y","@mcp/server-fs"]`, `{}`, "", `{}`,
				"installed", nil, now, now))

	cfg := ServerConfig{
		Name:      "filesystem",
		Transport: "stdio",
		Command:   "npx",
		Args:      []string{"-y", "@mcp/server-fs"},
	}
	result, err := svc.Install(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Install: %v", err)
	}
	if result.ID != testServerID {
		t.Errorf("ID = %v, want %v", result.ID, testServerID)
	}
	if result.Name != "filesystem" {
		t.Errorf("Name = %q", result.Name)
	}
	if result.Status != "installed" {
		t.Errorf("Status = %q, want installed", result.Status)
	}
}

func TestService_Install_DuplicateName(t *testing.T) {
	svc, mock := newTestService(t)

	mock.ExpectQuery("INSERT INTO mcp_servers").
		WillReturnError(sql.ErrConnDone) // simulate constraint violation

	_, err := svc.Install(context.Background(), ServerConfig{Name: "dup", Transport: "stdio", Command: "echo"})
	if err == nil {
		t.Fatal("expected error for duplicate name")
	}
}

// ── List ─────────────────────────────────────────────────────

func TestService_List_HappyPath(t *testing.T) {
	svc, mock := newTestService(t)
	now := time.Now()

	rows := sqlmock.NewRows(serverColumns()).
		AddRow(testServerID, "filesystem", "stdio", "npx",
			`["-y","@mcp/server-fs"]`, `{}`, "", `{}`,
			"connected", nil, now, now)
	mock.ExpectQuery("SELECT .+ FROM mcp_servers").WillReturnRows(rows)

	servers, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(servers) != 1 {
		t.Fatalf("got %d servers, want 1", len(servers))
	}
	if servers[0].Name != "filesystem" {
		t.Errorf("Name = %q", servers[0].Name)
	}
	if len(servers[0].Args) != 2 {
		t.Errorf("Args len = %d, want 2", len(servers[0].Args))
	}
}

func TestService_List_Empty(t *testing.T) {
	svc, mock := newTestService(t)

	mock.ExpectQuery("SELECT .+ FROM mcp_servers").
		WillReturnRows(sqlmock.NewRows(serverColumns()))

	servers, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if servers == nil {
		// nil is acceptable (append to nil slice), but test documents behavior
		t.Log("List returns nil for empty result (callers should nil-guard)")
	}
}

// ── Get ──────────────────────────────────────────────────────

func TestService_Get_Found(t *testing.T) {
	svc, mock := newTestService(t)
	now := time.Now()

	mock.ExpectQuery("SELECT .+ FROM mcp_servers WHERE id").
		WithArgs(testServerID).
		WillReturnRows(sqlmock.NewRows(serverColumns()).
			AddRow(testServerID, "filesystem", "stdio", "npx",
				`[]`, `{"FOO":"bar"}`, "", `{}`,
				"connected", nil, now, now))

	srv, err := svc.Get(context.Background(), testServerID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if srv.Env["FOO"] != "bar" {
		t.Errorf("Env[FOO] = %q", srv.Env["FOO"])
	}
}

func TestService_Get_NotFound(t *testing.T) {
	svc, mock := newTestService(t)

	mock.ExpectQuery("SELECT .+ FROM mcp_servers WHERE id").
		WithArgs(testServerID).
		WillReturnError(sql.ErrNoRows)

	_, err := svc.Get(context.Background(), testServerID)
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

// ── Delete ───────────────────────────────────────────────────

func TestService_Delete_Found(t *testing.T) {
	svc, mock := newTestService(t)

	mock.ExpectExec("DELETE FROM mcp_servers WHERE id").
		WithArgs(testServerID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := svc.Delete(context.Background(), testServerID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

func TestService_Delete_NotFound(t *testing.T) {
	svc, mock := newTestService(t)

	mock.ExpectExec("DELETE FROM mcp_servers WHERE id").
		WithArgs(testServerID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := svc.Delete(context.Background(), testServerID)
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

// ── UpdateStatus ─────────────────────────────────────────────

func TestService_UpdateStatus(t *testing.T) {
	svc, mock := newTestService(t)

	mock.ExpectExec("UPDATE mcp_servers").
		WithArgs("connected", sqlmock.AnyArg(), testServerID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := svc.UpdateStatus(context.Background(), testServerID, "connected", ""); err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}
}

func TestService_UpdateStatus_NotFound(t *testing.T) {
	svc, mock := newTestService(t)

	mock.ExpectExec("UPDATE mcp_servers").
		WithArgs("error", sqlmock.AnyArg(), testServerID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := svc.UpdateStatus(context.Background(), testServerID, "error", "bad")
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

// ── CacheTools ───────────────────────────────────────────────

func TestService_CacheTools_HappyPath(t *testing.T) {
	svc, mock := newTestService(t)

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM mcp_tools WHERE server_id").
		WithArgs(testServerID).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO mcp_tools").
		WithArgs(testServerID, "read_file", "Read a file", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	tools := []ToolDef{
		{Name: "read_file", Description: "Read a file", InputSchema: json.RawMessage(`{"type":"object"}`)},
	}
	if err := svc.CacheTools(context.Background(), testServerID, tools); err != nil {
		t.Fatalf("CacheTools: %v", err)
	}
}

func TestService_CacheTools_NilSchema(t *testing.T) {
	svc, mock := newTestService(t)

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM mcp_tools WHERE server_id").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO mcp_tools").
		WithArgs(testServerID, "list_dir", "", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	tools := []ToolDef{
		{Name: "list_dir", InputSchema: nil}, // nil defaults to {}
	}
	if err := svc.CacheTools(context.Background(), testServerID, tools); err != nil {
		t.Fatalf("CacheTools with nil schema: %v", err)
	}
}

// ── ListTools ────────────────────────────────────────────────

func TestService_ListTools(t *testing.T) {
	svc, mock := newTestService(t)

	rows := sqlmock.NewRows(toolColumns()).
		AddRow(testToolID, testServerID, "read_file", "Read a file", []byte(`{"type":"object"}`)).
		AddRow(uuid.New(), testServerID, "write_file", nil, []byte(`{}`))
	mock.ExpectQuery("SELECT .+ FROM mcp_tools").
		WithArgs(testServerID).
		WillReturnRows(rows)

	tools, err := svc.ListTools(context.Background(), testServerID)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}
	if len(tools) != 2 {
		t.Fatalf("got %d tools, want 2", len(tools))
	}
	if tools[0].Description != "Read a file" {
		t.Errorf("tools[0].Description = %q", tools[0].Description)
	}
	// Null description should be empty string
	if tools[1].Description != "" {
		t.Errorf("tools[1].Description = %q, want empty", tools[1].Description)
	}
}

// ── ListAllTools ─────────────────────────────────────────────

func TestService_ListAllTools(t *testing.T) {
	svc, mock := newTestService(t)

	rows := sqlmock.NewRows(toolWithServerColumns()).
		AddRow(testToolID, testServerID, "filesystem", "read_file", "Read a file", []byte(`{}`)).
		AddRow(uuid.New(), uuid.New(), "github", "create_issue", nil, []byte(`{}`))
	mock.ExpectQuery("SELECT .+ FROM mcp_tools .+ JOIN mcp_servers").
		WillReturnRows(rows)

	tools, err := svc.ListAllTools(context.Background())
	if err != nil {
		t.Fatalf("ListAllTools: %v", err)
	}
	if len(tools) != 2 {
		t.Fatalf("got %d tools, want 2", len(tools))
	}
	if tools[0].ServerName != "filesystem" {
		t.Errorf("tools[0].ServerName = %q", tools[0].ServerName)
	}
	if tools[1].ServerName != "github" {
		t.Errorf("tools[1].ServerName = %q", tools[1].ServerName)
	}
}

// ── FindToolByName ───────────────────────────────────────────

func TestService_FindToolByName(t *testing.T) {
	svc, mock := newTestService(t)
	now := time.Now()

	findToolColumns := []string{
		"id", "server_id", "server_name", "name", "description", "input_schema",
		"srv_id", "srv_name", "transport", "command", "args", "env", "url", "headers", "status", "error_message", "created_at", "updated_at",
	}
	mock.ExpectQuery("SELECT .+ FROM mcp_tools .+ JOIN mcp_servers").
		WithArgs("read_file").
		WillReturnRows(sqlmock.NewRows(findToolColumns).
			AddRow(testToolID, testServerID, "filesystem", "read_file", "Read a file", []byte(`{}`),
				testServerID, "filesystem", "stdio", "npx", `[]`, `{}`, "", `{}`, "connected", nil, now, now))

	tool, srv, err := svc.FindToolByName(context.Background(), "read_file")
	if err != nil {
		t.Fatalf("FindToolByName: %v", err)
	}
	if tool.Name != "read_file" {
		t.Errorf("tool.Name = %q", tool.Name)
	}
	if srv.Name != "filesystem" {
		t.Errorf("srv.Name = %q", srv.Name)
	}
}

func TestService_FindToolByName_NotFound(t *testing.T) {
	svc, mock := newTestService(t)

	findToolColumns := []string{
		"id", "server_id", "server_name", "name", "description", "input_schema",
		"srv_id", "srv_name", "transport", "command", "args", "env", "url", "headers", "status", "error_message", "created_at", "updated_at",
	}
	mock.ExpectQuery("SELECT .+ FROM mcp_tools .+ JOIN mcp_servers").
		WithArgs("nonexistent").
		WillReturnRows(sqlmock.NewRows(findToolColumns)) // empty result → ErrNoRows from QueryRow

	_, _, err := svc.FindToolByName(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

// ── FindServerByName ─────────────────────────────────────────

func TestService_FindServerByName(t *testing.T) {
	svc, mock := newTestService(t)
	now := time.Now()

	mock.ExpectQuery("SELECT .+ FROM mcp_servers WHERE name").
		WithArgs("filesystem").
		WillReturnRows(sqlmock.NewRows(serverColumns()).
			AddRow(testServerID, "filesystem", "stdio", "npx",
				`[]`, `{}`, "", `{}`,
				"connected", nil, now, now))

	srv, err := svc.FindServerByName(context.Background(), "filesystem")
	if err != nil {
		t.Fatalf("FindServerByName: %v", err)
	}
	if srv == nil {
		t.Fatal("expected non-nil server")
	}
	if srv.Name != "filesystem" {
		t.Errorf("Name = %q", srv.Name)
	}
}

func TestService_FindServerByName_NotFound(t *testing.T) {
	svc, mock := newTestService(t)

	mock.ExpectQuery("SELECT .+ FROM mcp_servers WHERE name").
		WithArgs("nonexistent").
		WillReturnError(sql.ErrNoRows)

	srv, err := svc.FindServerByName(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if srv != nil {
		t.Error("expected nil for not found")
	}
}

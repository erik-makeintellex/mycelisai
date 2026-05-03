package mcp

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestService_FindToolByName(t *testing.T) {
	svc, mock := newTestService(t)
	now := time.Now()

	mock.ExpectQuery("SELECT .+ FROM mcp_tools .+ JOIN mcp_servers").
		WithArgs("read_file").
		WillReturnRows(sqlmock.NewRows(findToolColumns()).
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

	mock.ExpectQuery("SELECT .+ FROM mcp_tools .+ JOIN mcp_servers").
		WithArgs("nonexistent").
		WillReturnRows(sqlmock.NewRows(findToolColumns()))

	_, _, err := svc.FindToolByName(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestService_FindServerByName(t *testing.T) {
	svc, mock := newTestService(t)
	now := time.Now()

	mock.ExpectQuery("SELECT .+ FROM mcp_servers WHERE name").
		WithArgs("filesystem").
		WillReturnRows(sqlmock.NewRows(serverColumns()).
			AddRow(testServerID, "filesystem", "stdio", "npx", `[]`, `{}`, "", `{}`, "connected", nil, now, now))

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

func findToolColumns() []string {
	return []string{
		"id", "server_id", "server_name", "name", "description", "input_schema",
		"srv_id", "srv_name", "transport", "command", "args", "env", "url", "headers", "status", "error_message", "created_at", "updated_at",
	}
}

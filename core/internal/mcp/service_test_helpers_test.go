package mcp

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

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

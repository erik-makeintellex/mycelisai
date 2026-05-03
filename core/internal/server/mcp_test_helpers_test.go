package server

import (
	"path/filepath"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/exchange"
	"github.com/mycelis/core/internal/mcp"
)

func loadStandardMCPLibrary(t *testing.T) *mcp.Library {
	t.Helper()
	lib, err := mcp.LoadLibrary(filepath.Join("..", "..", "config", "mcp-library.yaml"))
	if err != nil {
		t.Fatalf("LoadLibrary: %v", err)
	}
	return lib
}

// withMCPStubs wires non-nil MCP + MCPPool so handlers pass the nil guard.
// The stubs have nil DB underneath but validation tests fail before DB calls.
func withMCPStubs() func(*AdminServer) {
	return func(s *AdminServer) {
		s.MCP = mcp.NewService(nil)
		s.MCPPool = mcp.NewClientPool(s.MCP)
	}
}

func withMCPDB(t *testing.T) (func(*AdminServer), sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return func(s *AdminServer) {
		s.MCP = mcp.NewService(db)
		s.MCPPool = mcp.NewClientPool(s.MCP)
	}, mock
}

func withExchangeDB(t *testing.T) (func(*AdminServer), sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return func(s *AdminServer) {
		s.Exchange = exchange.NewService(db, nil, nil)
	}, mock
}

func mcpServerColumns() []string {
	return []string{"id", "name", "transport", "command", "args", "env", "url", "headers", "status", "error_message", "created_at", "updated_at"}
}

func mcpToolColumns() []string {
	return []string{"id", "server_id", "name", "description", "input_schema"}
}

func mcpToolWithServerColumns() []string {
	return []string{"id", "server_id", "server_name", "name", "description", "input_schema"}
}

func exchangeItemColumns() []string {
	return []string{"id", "channel_id", "channel_name", "schema_id", "payload", "created_by", "addressed_to", "thread_id", "visibility", "sensitivity_class", "source_role", "source_team", "target_role", "target_team", "allowed_consumers", "capability_id", "trust_class", "review_required", "metadata", "summary", "created_at"}
}

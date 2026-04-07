package swarm

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/internal/exchange"
	"github.com/mycelis/core/pkg/protocol"
)

type mcpExchangeExecutor struct {
	serverID uuid.UUID
	result   string
	err      error
}

func (m *mcpExchangeExecutor) FindToolByName(_ context.Context, name string) (uuid.UUID, string, error) {
	return m.serverID, name, nil
}

func (m *mcpExchangeExecutor) CallTool(_ context.Context, _ uuid.UUID, _ string, _ map[string]any) (string, error) {
	return m.result, m.err
}

func TestExecuteToolIteration_PersistsMCPFailureToExchange(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	serverID := uuid.New()
	expectMCPExchangeWrite(t, mock, "browser.research.results")

	agent := NewAgent(context.Background(), protocol.AgentManifest{
		ID:    "soma-admin",
		Role:  "soma",
		Tools: []string{"browser_search"},
	}, "alpha", nil, nil, &mcpExchangeExecutor{serverID: serverID, err: context.DeadlineExceeded})
	agent.SetInternalTools(&InternalToolRegistry{exchange: exchange.NewService(db, nil, nil)})
	agent.runID = "run-42"

	ok := agent.executeToolIteration(
		0,
		&cognitive.InferRequest{Profile: "chat"},
		&toolCallPayload{Name: "browser_search", Arguments: map[string]any{"query": "governed MCP visibility"}},
		map[string]int{},
		func(string, string) bool { return true },
		&agentToolLoopResult{responseText: `{"tool_call":{"name":"browser_search"}}`},
	)
	if ok {
		t.Fatal("expected MCP failure iteration to return false")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestExecuteToolIteration_PersistsMCPCompletionToExchange(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	serverID := uuid.New()
	expectMCPExchangeWrite(t, mock, "browser.research.results")

	agent := NewAgent(context.Background(), protocol.AgentManifest{
		ID:    "soma-admin",
		Role:  "soma",
		Tools: []string{"browser_search"},
	}, "alpha", nil, newFakeBrain(fakeMemoryProvider{inferText: "final answer"}), &mcpExchangeExecutor{serverID: serverID, result: "Workspace brief loaded."})
	agent.SetInternalTools(&InternalToolRegistry{exchange: exchange.NewService(db, nil, nil)})
	agent.runID = "run-42"

	req := &cognitive.InferRequest{
		Profile: "chat",
		Messages: []cognitive.ChatMessage{
			{Role: "user", Content: "Summarize the MCP brief."},
		},
	}
	result := &agentToolLoopResult{responseText: `{"tool_call":{"name":"browser_search"}}`}
	ok := agent.executeToolIteration(
		0,
		req,
		&toolCallPayload{Name: "browser_search", Arguments: map[string]any{"query": "workspace brief"}},
		map[string]int{},
		func(string, string) bool { return true },
		result,
	)
	if !ok {
		t.Fatal("expected MCP completion iteration to succeed")
	}
	if result.responseText != "final answer" {
		t.Fatalf("responseText = %q, want final answer", result.responseText)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func expectMCPExchangeWrite(t *testing.T, mock sqlmock.Sqlmock, channelName string) {
	t.Helper()
	now := time.Now()
	channelID := uuid.New()
	mock.ExpectQuery("FROM exchange_channels").
		WithArgs(channelName).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "channel_type", "owner", "participants", "reviewers", "schema_id", "retention_policy", "visibility", "sensitivity_class", "description", "metadata", "created_at"}).
			AddRow(
				channelID.String(),
				channelName,
				"output",
				"mcp",
				`[{"role":"mcp","can_read":true,"can_write":true},{"role":"soma","can_read":true,"can_write":true}]`,
				`["review","admin"]`,
				"ToolResult",
				"30d",
				"advanced",
				"team_scoped",
				"Normalized MCP output",
				[]byte(`{}`),
				now,
			))
	mock.ExpectQuery("INSERT INTO exchange_items").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(uuid.New().String(), now))
}

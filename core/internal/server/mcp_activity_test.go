package server

import (
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestHandleMCPActivity_NilExchange(t *testing.T) {
	s := newTestServer()
	rr := doRequest(t, http.HandlerFunc(s.handleMCPActivity), "GET", "/api/v1/mcp/activity", "")
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

func TestHandleMCPActivity_ReturnsPersistedMCPUsage(t *testing.T) {
	opt, mock := withExchangeDB(t)
	s := newTestServer(opt)
	now := time.Now()
	channelID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	itemID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")

	payload := []byte(`{"summary":"Read workspace brief successfully.","state":"completed","server_id":"srv-001","server_name":"filesystem","tool_name":"read_file","result_preview":"Read workspace brief successfully.","run_id":"run-1","source_team":"alpha","agent_id":"soma-admin","created_at":"2026-04-06T12:00:00Z"}`)
	metadata := []byte(`{"source_kind":"mcp","mcp":{"server_id":"srv-001","server_name":"filesystem","tool_name":"read_file","state":"completed","run_id":"run-1","source_team":"alpha","agent_id":"soma-admin"}}`)

	mock.ExpectQuery("SELECT i.id, i.channel_id, c.name, i.schema_id, i.payload, i.created_by").
		WithArgs("browser.research.results", nil, 10).
		WillReturnRows(sqlmock.NewRows(exchangeItemColumns()).
			AddRow(itemID.String(), channelID.String(), "browser.research.results", "ToolResult", payload, "mcp:filesystem", "", nil, "advanced", "team_scoped", "mcp", "alpha", "soma", "", []byte(`[]`), "browser_research", "bounded_external", true, metadata, "Read workspace brief successfully.", now))
	mock.ExpectQuery("SELECT i.id, i.channel_id, c.name, i.schema_id, i.payload, i.created_by").
		WithArgs("media.image.output", nil, 10).
		WillReturnRows(sqlmock.NewRows(exchangeItemColumns()))
	mock.ExpectQuery("SELECT i.id, i.channel_id, c.name, i.schema_id, i.payload, i.created_by").
		WithArgs("api.data.output", nil, 10).
		WillReturnRows(sqlmock.NewRows(exchangeItemColumns()))

	rr := doAuthenticatedRequest(t, http.HandlerFunc(s.handleMCPActivity), "GET", "/api/v1/mcp/activity?limit=10", "")
	assertStatus(t, rr, http.StatusOK)

	var resp struct {
		OK   bool                     `json:"ok"`
		Data []map[string]interface{} `json:"data"`
	}
	assertJSON(t, rr, &resp)
	if !resp.OK {
		t.Fatal("expected ok=true")
	}
	if len(resp.Data) != 1 {
		t.Fatalf("len(data) = %d, want 1", len(resp.Data))
	}
	if resp.Data[0]["server_id"] != "srv-001" {
		t.Fatalf("server_id = %v, want srv-001", resp.Data[0]["server_id"])
	}
	if resp.Data[0]["tool_name"] != "read_file" {
		t.Fatalf("tool_name = %v, want read_file", resp.Data[0]["tool_name"])
	}
	if resp.Data[0]["state"] != "completed" {
		t.Fatalf("state = %v, want completed", resp.Data[0]["state"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

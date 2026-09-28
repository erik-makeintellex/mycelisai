package server

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/mycelis/core/internal/exchange"
	"github.com/mycelis/core/internal/mcp"
)

// MCPS harness: a real ClientPool connected over streamable HTTP to an
// in-test MCP server whose tools record every call they receive. The handler
// under test reads the server row and tool cache from mcpsH.mcpMock, audits
// through mcpsH.auditMock (s.DB) and retains through mcpsH.exMock.

const mcpsCallRoute = "POST /api/v1/mcp/servers/{id}/tools/{tool}/call"

type mcpsCall struct {
	Tool      string
	Args      map[string]any
	AuditDone bool // every expected audit insert had happened when the tool ran
}

type mcpsH struct {
	t          *testing.T
	s          *AdminServer
	mux        *http.ServeMux
	serverID   uuid.UUID
	serverName string
	status     string
	tools      []string
	mcpMock    sqlmock.Sqlmock
	auditMock  sqlmock.Sqlmock
	exMock     sqlmock.Sqlmock
	mu         sync.Mutex
	calls      []mcpsCall
}

// mcpsCapture is a sqlmock argument that records the value it matched.
type mcpsCapture struct {
	mu  sync.Mutex
	raw string
}

func (c *mcpsCapture) Match(v driver.Value) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	switch typed := v.(type) {
	case []byte:
		c.raw = string(typed)
	case string:
		c.raw = typed
	}
	return true
}

func (c *mcpsCapture) value() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.raw
}

func (c *mcpsCapture) object(t *testing.T) map[string]any {
	t.Helper()
	var out map[string]any
	if err := json.Unmarshal([]byte(c.value()), &out); err != nil {
		t.Fatalf("captured value is not a JSON object: %v (%q)", err, c.value())
	}
	return out
}

func mcpsNewMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db, mock
}

// newMCPSHarness connects serverName (offering tools) and wires the server.
// withAuditDB=false leaves s.DB nil (the audit store is unavailable).
func newMCPSHarness(t *testing.T, serverName string, withAuditDB bool, tools ...string) *mcpsH {
	t.Helper()
	h := &mcpsH{t: t, serverID: uuid.New(), serverName: serverName, status: "connected", tools: tools}
	fixture := mcpserver.NewMCPServer("mcps-fixture", "0")
	for _, name := range tools {
		fixture.AddTool(mcplib.NewTool(name), h.recordingTool(name))
	}
	endpoint := newLocalHTTPTestServer(t, mcpserver.NewStreamableHTTPServer(fixture))

	poolDB, poolMock := mcpsNewMock(t)
	poolMock.ExpectBegin()
	poolMock.ExpectExec("DELETE FROM mcp_tools").WillReturnResult(sqlmock.NewResult(0, 0))
	for range tools {
		poolMock.ExpectExec("INSERT INTO mcp_tools").WillReturnResult(sqlmock.NewResult(0, 1))
	}
	poolMock.ExpectCommit()
	poolMock.ExpectExec("UPDATE mcp_servers").WillReturnResult(sqlmock.NewResult(0, 1))
	pool := mcp.NewClientPool(mcp.NewService(poolDB))
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := pool.Connect(ctx, mcp.ServerConfig{ID: h.serverID, Name: serverName, Transport: "sse", URL: endpoint.URL}); err != nil {
		t.Fatalf("connect fixture MCP server: %v", err)
	}
	t.Cleanup(pool.ShutdownAll)

	mcpDB, mcpMock := mcpsNewMock(t)
	exDB, exMock := mcpsNewMock(t)
	h.mcpMock, h.exMock = mcpMock, exMock
	h.s = newTestServer(func(s *AdminServer) {
		s.MCP = mcp.NewService(mcpDB)
		s.MCPPool = pool
		s.Exchange = exchange.NewService(exDB, nil, nil)
	})
	if withAuditDB {
		auditDB, auditMock := mcpsNewMock(t)
		h.s.DB, h.auditMock = auditDB, auditMock
	}
	h.mux = setupMux(t, mcpsCallRoute, h.s.handleMCPToolCall)
	return h
}

func (h *mcpsH) recordingTool(name string) mcpserver.ToolHandlerFunc {
	return func(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
		auditDone := h.auditMock == nil || h.auditMock.ExpectationsWereMet() == nil
		h.mu.Lock()
		h.calls = append(h.calls, mcpsCall{Tool: name, Args: req.GetArguments(), AuditDone: auditDone})
		h.mu.Unlock()
		return mcplib.NewToolResultText(name + " ok"), nil
	}
}

func (h *mcpsH) recorded() []mcpsCall {
	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]mcpsCall(nil), h.calls...)
}

// expectResolve answers the server row and the discovered tool cache.
func (h *mcpsH) expectResolve() {
	h.expectServerRow()
	h.expectToolCache()
}

// expectServerRow answers the registry row for the harness server. A
// harness server named "filesystem" carries the curated library command,
// because MCPL2 identifies filesystem servers by package, not by name.
func (h *mcpsH) expectServerRow() {
	now := time.Now()
	command, args := "", `[]`
	if h.serverName == "filesystem" {
		command, args = "npx", `["-y","@modelcontextprotocol/server-filesystem","./workspace"]`
	}
	h.mcpMock.ExpectQuery("SELECT .+ FROM mcp_servers").WithArgs(h.serverID).
		WillReturnRows(sqlmock.NewRows(mcpServerColumns()).
			AddRow(h.serverID.String(), h.serverName, "sse", command, args, `{}`, "http://fixture", `{}`, h.status, nil, now, now))
}

// expectToolCache answers the discovered tool cache for the harness server.
func (h *mcpsH) expectToolCache() {
	rows := sqlmock.NewRows(mcpToolColumns())
	for _, name := range h.tools {
		rows.AddRow(uuid.New().String(), h.serverID.String(), name, "", []byte(`{}`))
	}
	h.mcpMock.ExpectQuery("SELECT .+ FROM mcp_tools").WithArgs(h.serverID).WillReturnRows(rows)
}

// expectAudit expects one log_entries insert and captures its context JSON.
func (h *mcpsH) expectAudit(fail bool) *mcpsCapture {
	capture := &mcpsCapture{}
	exp := h.auditMock.ExpectExec("INSERT INTO log_entries").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "audit", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), capture)
	if fail {
		exp.WillReturnError(errors.New("audit store down"))
	} else {
		exp.WillReturnResult(sqlmock.NewResult(0, 1))
	}
	return capture
}

// expectExchange expects one retained Exchange item and captures its payload.
func (h *mcpsH) expectExchange() *mcpsCapture {
	now := time.Now()
	h.exMock.ExpectQuery("FROM exchange_channels").WillReturnRows(sqlmock.NewRows([]string{
		"id", "name", "channel_type", "owner", "participants", "reviewers", "schema_id", "retention_policy", "visibility", "sensitivity_class", "description", "metadata", "created_at",
	}).AddRow(uuid.New().String(), "api.data.output", "output", "mcp", `[{"role":"mcp","can_read":true,"can_write":true}]`, `["admin"]`, "ToolResult", "30d", "advanced", "team_scoped", "", []byte(`{}`), now))
	payload := &mcpsCapture{}
	args := make([]driver.Value, 18)
	for i := range args {
		args[i] = sqlmock.AnyArg()
	}
	args[2] = payload
	h.exMock.ExpectQuery("INSERT INTO exchange_items").WithArgs(args...).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(uuid.New().String(), now))
	return payload
}

// call posts body to {tool} (a raw, already-escaped path segment) as identity.
func (h *mcpsH) call(identity *RequestIdentity, serverID, rawTool, body string) *httptest.ResponseRecorder {
	h.t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/mcp/servers/"+serverID+"/tools/"+rawTool+"/call", strings.NewReader(body))
	if identity != nil {
		req = req.WithContext(context.WithValue(req.Context(), ctxKeyIdentity, identity))
	}
	rr := httptest.NewRecorder()
	h.mux.ServeHTTP(rr, req)
	return rr
}

func (h *mcpsH) callTool(identity *RequestIdentity, tool, body string) *httptest.ResponseRecorder {
	return h.call(identity, h.serverID.String(), tool, body)
}

func (h *mcpsH) assertCalls(want int) []mcpsCall {
	h.t.Helper()
	got := h.recorded()
	if len(got) != want {
		h.t.Fatalf("pool calls = %d (%+v), want %d", len(got), got, want)
	}
	return got
}

func (h *mcpsH) assertMocksMet() {
	h.t.Helper()
	for name, mock := range map[string]sqlmock.Sqlmock{"mcp": h.mcpMock, "audit": h.auditMock, "exchange": h.exMock} {
		if mock == nil {
			continue
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			h.t.Fatalf("%s expectations: %v", name, err)
		}
	}
}

func mcpsAssertBlocker(t *testing.T, rr *httptest.ResponseRecorder, status int, code string) blockerEnvelope {
	t.Helper()
	assertStatus(t, rr, status)
	env := decodeBlocker(t, rr)
	if env.OK || env.Data["code"] != code {
		t.Fatalf("blocker = %+v, want code %q", env, code)
	}
	return env
}

func mcpsWebAdmin() *RequestIdentity {
	return &RequestIdentity{UserID: "u-admin", Username: "admin@example.com", Role: "admin", EffectiveRole: "owner",
		PrincipalType: "google_workspace_user", AuthSource: "web_google", Scopes: []string{"*"}}
}

func mcpsEmptyServerRows() *sqlmock.Rows {
	return sqlmock.NewRows(mcpServerColumns())
}

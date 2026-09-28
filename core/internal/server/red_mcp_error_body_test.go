package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/mcp"
)

// RED: MCP error bodies are valid JSON for hostile input, and toolset audits
// record the normalized scope that is stored.

func redAssertJSONError(t *testing.T, label string, status int, rr interface {
	Result() *http.Response
}, body []byte) {
	t.Helper()
	if !json.Valid(body) {
		t.Fatalf("%s: invalid JSON error body: %q", label, body)
	}
	if got := rr.Result().StatusCode; got != status {
		t.Fatalf("%s: status = %d, want %d (%s)", label, got, status, body)
	}
	if ct := rr.Result().Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Fatalf("%s: content type = %q", label, ct)
	}
}

// Port of TestZzQaMCPA_ErrorBodiesAreValidJSON.
func TestRedLibraryErrorBodiesAreValidJSON(t *testing.T) {
	cases := map[string]int{
		`{"name":"fetch" "x":1}`:     http.StatusBadRequest,
		`{"name":"\u0001zz\"q"}`:     http.StatusNotFound,
		`{"name":"a b</script>"}`:    http.StatusNotFound,
		`{"name":1}`:                 http.StatusBadRequest,
		`{"name":""}`:                http.StatusBadRequest,
		`{"name":"x\\\"}\n{\"a\":"}`: http.StatusNotFound,
	}
	for body, status := range cases {
		for label, h := range map[string]func(*AdminServer) http.HandlerFunc{
			"install": func(s *AdminServer) http.HandlerFunc { return s.handleMCPLibraryInstall },
			"inspect": func(s *AdminServer) http.HandlerFunc { return s.handleMCPLibraryInspect },
		} {
			opt, _ := mcpaSharedDB(t)
			s := newTestServer(opt, mcpaLibrary())
			identity := localAdminIdentityForTest()
			if label == "inspect" {
				identity = standardUserIdentity()
			}
			rr := doAuthenticatedRequestAs(t, h(s), "POST", "/x", body, identity)
			redAssertJSONError(t, label+" "+body, status, rr, rr.Body.Bytes())
		}
	}
}

func TestRedToolCallErrorBodiesAreValidJSON(t *testing.T) {
	h := newMCPSHarness(t, "filesystem", false, "read_text_file")
	rr := h.call(mcpsWebAdmin(), `bad%22id%01`, "read_text_file", `{}`)
	redAssertJSONError(t, "invalid server id", http.StatusBadRequest, rr, rr.Body.Bytes())
	rr = h.callTool(mcpsWebAdmin(), "read_text_file", `{"a" "b"}`)
	redAssertJSONError(t, "invalid JSON body", http.StatusBadRequest, rr, rr.Body.Bytes())
	h.assertCalls(0)
}

func TestRedListErrorBodiesAreValidJSONAndRedacted(t *testing.T) {
	for label, pick := range map[string]func(*AdminServer) http.HandlerFunc{
		"servers": func(s *AdminServer) http.HandlerFunc { return s.handleMCPList },
		"tools":   func(s *AdminServer) http.HandlerFunc { return s.handleMCPToolsList },
	} {
		db, mock := mcpsNewMock(t)
		mock.ExpectQuery("SELECT").WillReturnError(errors.New("pq: \"down\"\x01 password=redlist-secret-7"))
		s := newTestServer(func(s *AdminServer) { s.MCP = mcp.NewService(db) })
		rr := doAuthenticatedRequestAs(t, pick(s), "GET", "/x", "", standardUserIdentity())
		redAssertJSONError(t, "list "+label, http.StatusInternalServerError, rr, rr.Body.Bytes())
		redAssertNoLeak(t, "list "+label, rr.Body.String(), "redlist-secret-7")
	}
}

// Port of TestZzQaMCPA_ToolSetAuditRecordsStoredScope, plus update and the
// "all" scope whose stray ref is dropped on store.
func TestRedToolSetAuditRecordsStoredScope(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name, route, body, wantKind, wantRef string
	}{
		{"create group", "create", `{"name":"workspace","tool_refs":["mcp:*"],"scope_kind":" GROUP ","scope_ref":" team-a "}`, "group", "team-a"},
		{"create default", "create", `{"name":"workspace","tool_refs":["mcp:*"],"scope_ref":" stray "}`, "all", ""},
		{"update host", "update", `{"name":"workspace","tool_refs":["mcp:*"],"scope_kind":"Host","scope_ref":"\thost-1 "}`, "host", "host-1"},
	}
	for _, tc := range cases {
		opt, mock := mcpaSharedDB(t)
		s := newTestServer(opt)
		route := mcpaWriteRoutes()[3]
		if tc.route == "update" {
			route = mcpaWriteRoutes()[4]
			mcpaExpectToolSetGet(mock, true)
		}
		audit := mcpaExpectAudit(mock, false)
		if tc.route == "create" {
			mock.ExpectQuery("INSERT INTO mcp_tool_sets").WithArgs("workspace", "", sqlmock.AnyArg(), tc.wantKind, tc.wantRef).
				WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).AddRow(mcpaToolSetID, now, now))
		} else {
			mock.ExpectQuery("UPDATE mcp_tool_sets SET").WithArgs("workspace", "", sqlmock.AnyArg(), tc.wantKind, tc.wantRef, sqlmock.AnyArg()).
				WillReturnRows(sqlmock.NewRows(toolSetColumns()).AddRow(mcpaToolSetID, "workspace", "", `["mcp:*"]`, tc.wantKind, tc.wantRef, "default", now, now))
		}
		route.body = tc.body
		rr := route.do(t, s, mcpsWebAdmin())
		if rr.Code != http.StatusOK && rr.Code != http.StatusCreated {
			t.Fatalf("%s: status %d body %s", tc.name, rr.Code, rr.Body.String())
		}
		mcpaAssertMet(t, mock)
		if rec := audit.object(t); rec["scope_kind"] != tc.wantKind || rec["scope_ref"] != tc.wantRef {
			t.Fatalf("%s: audit scope = %q/%q, stored %q/%q", tc.name, rec["scope_kind"], rec["scope_ref"], tc.wantKind, tc.wantRef)
		}
	}
}

// An invalid scope is still refused by the store (authority unchanged); the
// audit keeps the caller's value, trimmed, since nothing was normalized.
func TestRedToolSetInvalidScopeStillRefused(t *testing.T) {
	opt, mock := mcpaSharedDB(t)
	s := newTestServer(opt)
	mcpaExpectAudit(mock, false)
	failed := mcpaExpectAudit(mock, false)
	route := mcpaWriteRoutes()[3]
	route.body = `{"name":"workspace","tool_refs":["mcp:*"],"scope_kind":"planet"}`
	rr := route.do(t, s, mcpsWebAdmin())
	assertStatus(t, rr, http.StatusBadRequest)
	mcpaAssertMet(t, mock)
	if rec := failed.object(t); rec["action"] != "mcp_toolset_create_failed" {
		t.Fatalf("failure record = %v", rec)
	}
}

package server

import (
	"database/sql/driver"
	"errors"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/mcp"
)

// MCPA (contract MCPA_MCP_CONFIGURATION_AUTHORITY_CONTRACT.md D5-D8): install
// and apply audit before any side effect, require_approval entries need an
// approver (self-approved tier 2, never a 202), and governance context comes
// from the identity only. One sqlmock backs the audit store and the MCP
// services, so the in-order expectations prove the audit comes first.

const (
	mcpaServerID    = "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
	mcpaSecretValue = "mcpa-secret-value-do-not-leak"
)

// mcpaSharedDB wires the audit store, MCP service, pool and toolsets to one mock.
func mcpaSharedDB(t *testing.T) (func(*AdminServer), sqlmock.Sqlmock) {
	t.Helper()
	db, mock := mcpsNewMock(t)
	return func(s *AdminServer) {
		s.DB = db
		s.MCP = mcp.NewService(db)
		s.MCPPool = mcp.NewClientPool(mcp.NewService(db))
		s.MCPToolSets = mcp.NewToolSetService(db)
	}, mock
}

// mcpaLibrary is the MCPA-H stub library plus a credentialed external entry
// named github. No transport spawns a process.
func mcpaLibrary() func(*AdminServer) {
	return func(s *AdminServer) {
		mcpahLibrary()(s)
		s.MCPLibrary.Categories[0].Servers = append(s.MCPLibrary.Categories[0].Servers, mcp.LibraryEntry{
			Name: "github", Transport: "unsupported", DeploymentBoundary: "external_saas",
			EnvironmentVariables: []mcp.LibraryEnvVar{{Name: "GITHUB_PERSONAL_ACCESS_TOKEN", Required: true, Secret: true}},
		})
	}
}

// mcpaExpectLookup answers the pre-audit replace-by-name lookup.
func mcpaExpectLookup(mock sqlmock.Sqlmock, name string, exists bool) {
	rows := sqlmock.NewRows(mcpServerColumns())
	if exists {
		now := time.Now()
		rows.AddRow(mcpaServerID, name, "unsupported", "", `[]`, `{}`, "", `{}`, "connected", nil, now, now)
	}
	mock.ExpectQuery("SELECT .+ FROM mcp_servers").WithArgs(name).WillReturnRows(rows)
}

// mcpaExpectAudit expects one audit insert and captures its context JSON.
func mcpaExpectAudit(mock sqlmock.Sqlmock, fail bool) *mcpsCapture {
	capture := &mcpsCapture{}
	exp := mock.ExpectExec("INSERT INTO log_entries").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "audit", "mcp-config", sqlmock.AnyArg(), sqlmock.AnyArg(), capture)
	if fail {
		exp.WillReturnError(errors.New("audit store down"))
	} else {
		exp.WillReturnResult(sqlmock.NewResult(0, 1))
	}
	return capture
}

// mcpaExpectConnectFailedInstall scripts lookup, audit, Install, a failed
// Connect (status error) and the connect-stage failure record.
func mcpaExpectConnectFailedInstall(mock sqlmock.Sqlmock, name string, envArg driver.Value) *mcpsCapture {
	mcpaExpectLookup(mock, name, false)
	mcpaExpectAudit(mock, false)
	now := time.Now()
	mock.ExpectQuery("INSERT INTO mcp_servers").
		WithArgs(name, "unsupported", "", sqlmock.AnyArg(), envArg, "", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows(mcpServerColumns()).
			AddRow(mcpaServerID, name, "unsupported", "", `[]`, `{}`, "", `{}`, "installed", nil, now, now))
	mock.ExpectExec("UPDATE mcp_servers").WithArgs("error", sqlmock.AnyArg(), mcpaServerID).WillReturnResult(sqlmock.NewResult(0, 1))
	return mcpaExpectAudit(mock, false)
}

func mcpaStrings(v any) []string {
	items, _ := v.([]any)
	out := make([]string, 0, len(items))
	for _, item := range items {
		s, _ := item.(string)
		out = append(out, s)
	}
	return out
}

func mcpaAssertMet(t *testing.T, mock sqlmock.Sqlmock) {
	t.Helper()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations (in order): %v", err)
	}
}

func TestMcpaInstall_AdminAuditsBeforeInstallWithEnvKeysOnly(t *testing.T) {
	url := mcpaFixtureURL(t)
	admins := map[string]*RequestIdentity{"web_admin": mcpsWebAdmin(), "api_key": localAdminIdentityForTest(), "break_glass": breakGlassIdentityForTest()}
	for label, identity := range admins {
		for _, route := range mcpahWriteRoutes() {
			opt, mock := mcpaSharedDB(t)
			s := newTestServer(opt, mcpaLiveLibrary(url))
			audit := mcpaExpectLiveInstall(t, s, mock, "tokened", mcpahEnvKeys{"FETCH_REGION", "FETCH_TOKEN"}, url, `{}`, `{}`, false)
			rr := doAuthenticatedRequestAs(t, route.handler(s), "POST", route.path, `{"name":"tokened","env":{"FETCH_TOKEN":"`+mcpaSecretValue+`"}}`, identity)
			if rr.Code != http.StatusOK {
				t.Fatalf("%s %s: status = %d body %s", label, route.name, rr.Code, rr.Body.String())
			}
			mcpaAssertMet(t, mock)
			record := audit.object(t)
			if record["action"] != "mcp_server_installed" || record["server_name"] != "tokened" || record["replaced_existing"] != false {
				t.Fatalf("%s %s: audit = %v", label, route.name, record)
			}
			if got := mcpaStrings(record["env_keys"]); !slices.Equal(got, []string{"FETCH_TOKEN"}) {
				t.Fatalf("%s %s: env_keys = %v", label, route.name, got)
			}
			if record["self_approved"] != false || record["authority"] != scopeMCPConfigWrite {
				t.Fatalf("%s %s: authority fields = %v", label, route.name, record)
			}
			if strings.Contains(audit.value(), mcpaSecretValue) || strings.Contains(rr.Body.String(), mcpaSecretValue) {
				t.Fatalf("%s %s: env value leaked into audit or response", label, route.name)
			}
		}
	}
}

func TestMcpaInstall_ReinstallRecordsReplacedExisting(t *testing.T) {
	opt, mock := mcpaSharedDB(t)
	url := mcpaFixtureURL(t)
	s := newTestServer(opt, mcpaLiveLibrary(url))
	audit := mcpaExpectLiveInstall(t, s, mock, "fetch", sqlmock.AnyArg(), url, `{}`, `{}`, true)
	rr := doAuthenticatedRequest(t, http.HandlerFunc(s.handleMCPLibraryInstall), "POST", "/api/v1/mcp/library/install", `{"name":"fetch"}`)
	assertStatus(t, rr, http.StatusOK)
	mcpaAssertMet(t, mock)
	if audit.object(t)["replaced_existing"] != true {
		t.Fatalf("audit = %s, want replaced_existing=true", audit.value())
	}
}

func TestMcpaInstall_ApproverInstallsCredentialedEntrySelfApproved(t *testing.T) {
	url := mcpaFixtureURL(t)
	for _, route := range mcpahWriteRoutes() {
		opt, mock := mcpaSharedDB(t)
		s := newTestServer(opt, mcpaLiveLibrary(url))
		audit := mcpaExpectLiveInstall(t, s, mock, "github", mcpahEnvKeys{"GITHUB_PERSONAL_ACCESS_TOKEN"}, url, `{}`, `{}`, false)
		body := `{"name":"github","env":{"GITHUB_PERSONAL_ACCESS_TOKEN":"` + mcpaSecretValue + `"}}`
		rr := doAuthenticatedRequestAs(t, route.handler(s), "POST", route.path, body, adminWithScopes(scopeMCPConfigWrite, scopeApprovalsDecide))
		if rr.Code != http.StatusOK {
			t.Fatalf("%s: status = %d, want 200 (no 202); body %s", route.name, rr.Code, rr.Body.String())
		}
		mcpaAssertMet(t, mock)
		var resp map[string]any
		assertJSON(t, rr, &resp)
		if resp["self_approved"] != true || resp["requires_approval"] == true || resp["audit_event_id"] == "" {
			t.Fatalf("%s: response = %v", route.name, resp)
		}
		record := audit.object(t)
		if record["self_approved"] != true || record["tier"] != float64(approverTierApprover) || record["authority"] != scopeApprovalsDecide {
			t.Fatalf("%s: audit = %v", route.name, record)
		}
		if strings.Contains(audit.value(), mcpaSecretValue) || strings.Contains(rr.Body.String(), mcpaSecretValue) {
			t.Fatalf("%s: token value leaked", route.name)
		}
	}
}

func TestMcpaInstall_RequireApprovalWithoutApproverIsForbidden(t *testing.T) {
	for _, route := range mcpahWriteRoutes() {
		opt, mock := mcpaSharedDB(t)
		s := newTestServer(opt, mcpaLibrary())
		for _, name := range []string{"github", "remote"} {
			if name == "remote" {
				s.MCPLibrary.Categories[0].Servers = append(s.MCPLibrary.Categories[0].Servers,
					mcp.LibraryEntry{Name: "remote", Transport: "sse", URL: "https://mcp.example.com/sse", Tags: []string{"remote"}})
			}
			rr := doAuthenticatedRequestAs(t, route.handler(s), "POST", route.path, `{"name":"`+name+`"}`, adminWithScopes(scopeMCPConfigWrite))
			env := mcpsAssertBlocker(t, rr, http.StatusForbidden, codeAdminRequired)
			if env.Data["required_scope"] != scopeApprovalsDecide {
				t.Fatalf("%s %s: required_scope = %v", route.name, name, env.Data["required_scope"])
			}
		}
		mcpaAssertMet(t, mock) // no lookup, audit or install ran
	}
}

func TestMcpaInstall_BodyGovernanceContextIsIgnored(t *testing.T) {
	opt, mock := mcpaSharedDB(t)
	url := mcpaFixtureURL(t)
	s := newTestServer(opt, mcpaLiveLibrary(url))
	audit := mcpaExpectLiveInstall(t, s, mock, "fetch", sqlmock.AnyArg(), url, `{}`, `{}`, false)
	identity := adminWithScopes(scopeMCPConfigWrite)
	body := `{"name":"fetch","governance_context":{"actor_role":"operator","owner_user_id":"mcpa-spoofed-owner"}}`
	rr := doAuthenticatedRequestAs(t, http.HandlerFunc(s.handleMCPLibraryApply), "POST", "/api/v1/mcp/library/apply", body, identity)
	assertStatus(t, rr, http.StatusOK)
	mcpaAssertMet(t, mock)
	if strings.Contains(audit.value()+rr.Body.String(), "mcpa-spoofed-owner") {
		t.Fatalf("body owner_user_id reached the audit or response: %s", audit.value())
	}
	var resp struct {
		Inspection struct {
			Context mcpGovernanceContext `json:"governance_context"`
		} `json:"inspection"`
	}
	assertJSON(t, rr, &resp)
	if resp.Inspection.Context.OwnerUserID != identity.UserID || resp.Inspection.Context.ActorRole != "owner" {
		t.Fatalf("governance_context = %+v, want identity %s/owner", resp.Inspection.Context, identity.UserID)
	}
}

func TestMcpaGovernanceContext_IdentityOnly(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, "/", nil)
	req = req.WithContext(withIdentityForMcpa(req, standardUserIdentity()))
	got := normalizeMCPGovernanceContext(req, mcpGovernanceContext{ActorRole: "owner", OwnerUserID: "root", SourceSurface: "soma"})
	if got.ActorRole != "operator" || got.OwnerUserID != "u-std" || got.SourceSurface != "soma" {
		t.Fatalf("context = %+v, want identity role/owner and the body label", got)
	}
	anon, _ := http.NewRequest(http.MethodPost, "/", nil)
	if got := normalizeMCPGovernanceContext(anon, mcpGovernanceContext{ActorRole: "owner", OwnerUserID: "root"}); got.ActorRole != "" || got.OwnerUserID != "" {
		t.Fatalf("anonymous context = %+v, want no body-supplied owner or role", got)
	}
}

func TestMcpaInspect_RequiredScopesAndStandardUserRead(t *testing.T) {
	cases := map[string][]string{
		"fetch":  {scopeMCPConfigWrite},
		"github": {scopeMCPConfigWrite, scopeApprovalsDecide},
	}
	for name, want := range cases {
		s := newTestServer(mcpaLibrary())
		rr := doAuthenticatedRequestAs(t, http.HandlerFunc(s.handleMCPLibraryInspect), "POST", "/api/v1/mcp/library/inspect", `{"name":"`+name+`"}`, standardUserIdentity())
		assertStatus(t, rr, http.StatusOK)
		var resp map[string]any
		assertJSON(t, rr, &resp)
		if got := mcpaStrings(resp["required_scopes"]); !slices.Equal(got, want) {
			t.Fatalf("%s: required_scopes = %v, want %v", name, got, want)
		}
		if name == "github" && resp["decision"] != "require_approval" {
			t.Fatalf("github decision = %v, want require_approval", resp["decision"])
		}
	}
}

func TestMcpaScopeString_IsNotAToolGrant(t *testing.T) {
	if mcp.IsMCPRef(scopeMCPConfigWrite) || mcp.IsToolSetRef(scopeMCPConfigWrite) {
		t.Fatalf("%s must not parse as an agent tool grant", scopeMCPConfigWrite)
	}
}

func TestMcpaInstall_ErrorsAreFixedAndRecordedAfterAudit(t *testing.T) {
	for _, route := range mcpahWriteRoutes() {
		opt, mock := mcpaSharedDB(t)
		s := newTestServer(opt, mcpaLibrary())
		mcpaExpectLookup(mock, "fetch", false)
		mcpaExpectAudit(mock, false)
		mock.ExpectQuery("INSERT INTO mcp_servers").WillReturnError(errors.New("pq: password=" + mcpaSecretValue + " rejected"))
		failed := mcpaExpectAudit(mock, false)
		rr := doAuthenticatedRequest(t, route.handler(s), "POST", route.path, `{"name":"fetch"}`)
		assertStatus(t, rr, http.StatusInternalServerError)
		if strings.Contains(rr.Body.String(), mcpaSecretValue) || strings.Contains(rr.Body.String(), "pq:") {
			t.Fatalf("%s: raw install error reached the response: %s", route.name, rr.Body.String())
		}
		mcpaAssertMet(t, mock)
		if record := failed.object(t); record["action"] != "mcp_server_install_failed" || strings.Contains(failed.value(), mcpaSecretValue) {
			t.Fatalf("%s: failure record = %s", route.name, failed.value())
		}
	}
}

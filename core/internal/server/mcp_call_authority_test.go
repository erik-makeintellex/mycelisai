package server

import (
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// MCPS acceptance (contract items 1-6). Every call goes through the real
// route, a real ClientPool and a recording MCP server.

const mcpsReadBody = `{"arguments":{"path":"notes/readme.md"}}`

func mcpsRetainedResult(t *testing.T, payload *mcpsCapture) map[string]any {
	t.Helper()
	result, _ := payload.object(t)["tool_result"].(map[string]any)
	if result == nil {
		t.Fatalf("exchange payload has no tool_result: %s", payload.value())
	}
	return result
}

// 1. A standard user reads the workspace directly; the Exchange item names them.
func TestMCPSStandardUserReadsFilesystemDirectly(t *testing.T) {
	for _, tool := range []string{"list_directory", "read_text_file"} {
		t.Run(tool, func(t *testing.T) {
			h := newMCPSHarness(t, "filesystem", true, "list_directory", "read_text_file", "write_file")
			h.expectResolve()
			payload := h.expectExchange()

			rr := h.callTool(standardUserIdentity(), tool, mcpsReadBody)

			assertStatus(t, rr, http.StatusOK)
			h.assertCalls(1)
			h.assertMocksMet()
			requestedBy, _ := mcpsRetainedResult(t, payload)["requested_by"].(map[string]any)
			if requestedBy["user_id"] != "u-std" || requestedBy["role"] != "operator" {
				t.Fatalf("requested_by = %v, want the standard user", requestedBy)
			}
		})
	}
}

// 1. Approvers (web admin and the local API key) call high-risk tools; the
// audit record is written before the pool call and marks self-approval.
func TestMCPSApproverHighCallAuditsBeforeCall(t *testing.T) {
	for name, c := range map[string]struct {
		identity     *RequestIdentity
		server, tool string
	}{
		"web admin write_file":   {mcpsWebAdmin(), "filesystem", "write_file"},
		"web admin github issue": {mcpsWebAdmin(), "github", "create_issue"},
		"api key write_file":     {localAdminIdentityForTest(), "filesystem", "write_file"},
		"api key github issue":   {localAdminIdentityForTest(), "github", "create_issue"},
	} {
		t.Run(name, func(t *testing.T) {
			h := newMCPSHarness(t, c.server, true, c.tool)
			h.expectResolve()
			audit := h.expectAudit(false)
			payload := h.expectExchange()

			rr := h.callTool(c.identity, c.tool, `{"arguments":{"path":"a.txt","content":"body"}}`)

			assertStatus(t, rr, http.StatusOK)
			if calls := h.assertCalls(1); !calls[0].AuditDone {
				t.Fatalf("pool call ran before the audit record was written")
			}
			h.assertMocksMet()
			ctx := audit.object(t)
			if ctx["action"] != "mcp_tool_called" || ctx["self_approved"] != true || ctx["authority"] != scopeApprovalsDecide ||
				ctx["risk"] != "high" || ctx["tier"] != float64(approverTierApprover) || ctx["server_name"] != c.server || ctx["tool"] != c.tool {
				t.Fatalf("audit context = %v", ctx)
			}
			if keys := ctx["argument_keys"]; !reflect.DeepEqual(keys, []any{"content", "path"}) {
				t.Fatalf("argument_keys = %v, want sorted keys", keys)
			}
			if strings.Contains(audit.value(), "a.txt") || strings.Contains(audit.value(), "body") {
				t.Fatalf("audit record holds argument values: %s", audit.value())
			}
			if id, _ := mcpsRetainedResult(t, payload)["audit_event_id"].(string); id == "" {
				t.Fatalf("exchange item does not carry audit_event_id: %s", payload.value())
			}
		})
	}
}

// 2. Non-approvers get 403 admin_required for any high-risk tool; nothing runs
// and a keys-only refusal audit is attempted.
func TestMCPSNonApproverHighCallIsRefused(t *testing.T) {
	cases := []struct{ server, tool string }{
		{"filesystem", "write_file"}, {"filesystem", "create_directory"}, {"filesystem", "move_file"},
		{"filesystem", "edit_file"}, {"github", "create_issue"}, {"slack", "slack_post_message"}, {"fetch", "fetch"},
	}
	identities := map[string]*RequestIdentity{
		"standard user":        standardUserIdentity(),
		"admin without decide": adminWithScopes("outputs:read"),
	}
	for who, identity := range identities {
		for _, c := range cases {
			t.Run(who+" "+c.server+"/"+c.tool, func(t *testing.T) {
				h := newMCPSHarness(t, c.server, true, c.tool)
				h.expectResolve()
				refusal := h.expectAudit(false)

				rr := h.callTool(identity, c.tool, `{"arguments":{"path":"secret-plan.txt"}}`)

				env := mcpsAssertBlocker(t, rr, http.StatusForbidden, codeAdminRequired)
				h.assertCalls(0)
				h.assertMocksMet()
				if identity.Role == "admin" && env.Data["required_scope"] != scopeApprovalsDecide {
					t.Fatalf("admin blocker required_scope = %v, want %s", env.Data["required_scope"], scopeApprovalsDecide)
				}
				assertUserSafe(t, "mcp approver blocker", userVisibleIfUser(identity, env)...)
				if ctx := refusal.object(t); ctx["action"] != "mcp_tool_call_refused" || strings.Contains(refusal.value(), "secret-plan") {
					t.Fatalf("refusal audit = %s", refusal.value())
				}
			})
		}
	}
}

func userVisibleIfUser(identity *RequestIdentity, env blockerEnvelope) []string {
	if identity.Role == "admin" {
		return nil
	}
	return userVisible(env)
}

// 2. Read tools need outputs:read; no identity is 401 before any lookup.
func TestMCPSReadNeedsOutputsReadAndIdentity(t *testing.T) {
	h := newMCPSHarness(t, "filesystem", true, "read_text_file")
	h.expectResolve()
	noRead := &RequestIdentity{UserID: "u-narrow", Role: "operator", Scopes: []string{"soma:work"}}
	env := mcpsAssertBlocker(t, h.callTool(noRead, "read_text_file", mcpsReadBody), http.StatusForbidden, codeMCPCallForbidden)
	assertUserSafe(t, "mcp_call_forbidden", userVisible(env)...)

	assertStatus(t, h.callTool(nil, "read_text_file", mcpsReadBody), http.StatusUnauthorized)
	h.assertCalls(0)
	h.assertMocksMet()
}

// 3. Resolution is exact: anything not in the discovered cache is 404.
func TestMCPSUnresolvedToolIsNotFound(t *testing.T) {
	h := newMCPSHarness(t, "filesystem", true, "read_text_file", "write_file")
	admin := mcpsWebAdmin()

	h.mcpMock.ExpectQuery("SELECT .+ FROM mcp_servers").WillReturnRows(mcpsEmptyServerRows())
	mcpsAssertBlocker(t, h.call(admin, uuid.NewString(), "read_text_file", mcpsReadBody), http.StatusNotFound, codeMCPToolNotFound)
	mcpsAssertBlocker(t, h.call(admin, uuid.Nil.String(), "read_text_file", mcpsReadBody), http.StatusNotFound, codeMCPToolNotFound)
	for _, raw := range []string{"list_directory", "Read_Text_File", "read_text_file%20", "read%2Ftext_file", "write_file%2F..%2Fread_text_file"} {
		h.expectResolve()
		env := mcpsAssertBlocker(t, h.callTool(admin, raw, mcpsReadBody), http.StatusNotFound, codeMCPToolNotFound)
		if raw == "Read_Text_File" {
			assertUserSafe(t, "mcp_tool_not_found", env.Error, env.Data["recommended_action"].(string))
		}
	}
	h.assertCalls(0)
	h.assertMocksMet()
}

// 3. A server that is not connected, or a missing pool, is 503.
func TestMCPSDisconnectedServerIsUnavailable(t *testing.T) {
	h := newMCPSHarness(t, "filesystem", true, "read_text_file")
	h.status = "stopped"
	h.expectServerRow() // a stopped server never reaches the tool cache
	mcpsAssertBlocker(t, h.callTool(mcpsWebAdmin(), "read_text_file", mcpsReadBody), http.StatusServiceUnavailable, codeServiceUnavailable)

	pool := h.s.MCPPool
	h.s.MCPPool = nil
	assertStatus(t, h.callTool(mcpsWebAdmin(), "read_text_file", mcpsReadBody), http.StatusServiceUnavailable)
	h.s.MCPPool = pool
	h.assertCalls(0)
	h.assertMocksMet()
}

// 4. The allowlist is keyed on the exact server name, and body fields never
// lower the class or claim authority.
func TestMCPSAllowlistIsExactAndBodyCannotClaimAuthority(t *testing.T) {
	for _, server := range []string{"filesystem2", "Filesystem", "my-reader"} {
		t.Run(server, func(t *testing.T) {
			h := newMCPSHarness(t, server, true, "read_text_file")
			h.expectResolve()
			h.expectAudit(false)
			mcpsAssertBlocker(t, h.callTool(standardUserIdentity(), "read_text_file", mcpsReadBody), http.StatusForbidden, codeAdminRequired)
			h.assertCalls(0)
		})
	}
	for _, body := range []string{
		`{"arguments":{"path":"a"},"actor":"admin","confirm_token":"x","risk":"low","self_approved":true}`,
		`{"path":"a","actor":"admin","confirm_token":"x","risk":"low","self_approved":true}`,
	} {
		h := newMCPSHarness(t, "filesystem", true, "write_file")
		h.expectResolve()
		h.expectAudit(false)
		mcpsAssertBlocker(t, h.callTool(standardUserIdentity(), "write_file", body), http.StatusForbidden, codeAdminRequired)
		h.assertCalls(0)
		h.assertMocksMet()
	}
}

// 5. The audit fails closed for high calls; reads do not depend on it.
func TestMCPSHighCallAuditFailsClosed(t *testing.T) {
	t.Run("audit store missing", func(t *testing.T) {
		h := newMCPSHarness(t, "filesystem", false, "write_file", "read_text_file")
		h.expectResolve()
		mcpsAssertBlocker(t, h.callTool(mcpsWebAdmin(), "write_file", `{"path":"a"}`), http.StatusServiceUnavailable, codeServiceUnavailable)
		h.assertCalls(0)

		h.expectResolve()
		h.expectExchange()
		assertStatus(t, h.callTool(standardUserIdentity(), "read_text_file", mcpsReadBody), http.StatusOK)
		h.assertCalls(1)
		h.assertMocksMet()
	})
	t.Run("audit insert fails", func(t *testing.T) {
		h := newMCPSHarness(t, "filesystem", true, "write_file")
		h.expectResolve()
		h.expectAudit(true)
		mcpsAssertBlocker(t, h.callTool(localAdminIdentityForTest(), "write_file", `{"path":"a"}`), http.StatusServiceUnavailable, codeServiceUnavailable)
		h.assertCalls(0)
		h.assertMocksMet()
	})
}

// 6. Retained arguments are redacted; the tool still receives raw values and
// neither the audit record nor the response echoes any argument value.
func TestMCPSRetainedArgumentsAreRedacted(t *testing.T) {
	h := newMCPSHarness(t, "filesystem", true, "write_file")
	h.expectResolve()
	audit := h.expectAudit(false)
	payload := h.expectExchange()
	body := `{"arguments":{"token":"tok-raw-1","env":{"GITHUB_PERSONAL_ACCESS_TOKEN":"ghp-raw-2"},"url":"https://u:p@h"}}`

	rr := h.callTool(mcpsWebAdmin(), "write_file", body)

	assertStatus(t, rr, http.StatusOK)
	got := h.assertCalls(1)[0].Args
	env, _ := got["env"].(map[string]any)
	if got["token"] != "tok-raw-1" || env["GITHUB_PERSONAL_ACCESS_TOKEN"] != "ghp-raw-2" || got["url"] != "https://u:p@h" {
		t.Fatalf("tool did not receive raw arguments: %v", got)
	}
	retained, _ := mcpsRetainedResult(t, payload)["arguments"].(map[string]any)
	retainedEnv, _ := retained["env"].(map[string]any)
	if retained["token"] != "[REDACTED]" || retainedEnv["GITHUB_PERSONAL_ACCESS_TOKEN"] != "[REDACTED]" {
		t.Fatalf("retained arguments not redacted: %v", retained)
	}
	for label, text := range map[string]string{"exchange": payload.value(), "audit": audit.value(), "response": rr.Body.String()} {
		for _, secret := range []string{"tok-raw-1", "ghp-raw-2", "u:p@"} {
			if strings.Contains(text, secret) {
				t.Fatalf("%s leaked %q: %s", label, secret, text)
			}
		}
	}
	h.assertMocksMet()
}

// MCPS blocker copy: plain user variants and actionable admin variants for
// both new codes and the two MCP-specific copies of existing codes.
func TestMCPSBlockerCopy_UserPlainAdminActionable(t *testing.T) {
	copies := map[string]roleBlockerText{
		codeMCPToolNotFound:  blockerCopies[codeMCPToolNotFound],
		codeMCPCallForbidden: blockerCopies[codeMCPCallForbidden],
		"approver":           mcpDirectCallNeedsApprover,
		"offline":            mcpServerOfflineCopy,
	}
	for name, c := range copies {
		user, admin := c.forViewer(false), c.forViewer(true)
		if user.Message == "" || user.Action == "" || admin.Action == "" {
			t.Fatalf("%s: incomplete copy %+v", name, c)
		}
		assertUserSafe(t, name, user.Message, user.Action)
		if user == admin {
			t.Fatalf("%s: admin reads the user copy", name)
		}
	}
}

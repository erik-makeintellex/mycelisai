package server

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// MCPL item 1: Core confines every filesystem MCP path argument to the
// MYCELIS_WORKSPACE root (lexically and through symlinks) before any pool
// call, whatever the upstream filesystem server would do.

// mcplWorkspace builds root (the workspace) next to an outside directory
// holding a secret, with a symlink out, a dangling symlink out and a symlink
// that stays inside. It returns the root and the outside directory.
func mcplWorkspace(t *testing.T) (string, string) {
	t.Helper()
	base := t.TempDir()
	root, outside := filepath.Join(base, "ws"), filepath.Join(base, "outside")
	for _, dir := range []string{filepath.Join(root, "notes"), outside} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
	}
	for path, body := range map[string]string{filepath.Join(outside, "secret.txt"): "s", filepath.Join(base, ".env"): "K=V", filepath.Join(root, "notes", "a.md"): "a"} {
		if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
			t.Fatalf("write: %v", err)
		}
	}
	for link, target := range map[string]string{"out": outside, "dangle": filepath.Join(outside, "new.txt"), "inner": filepath.Join(root, "notes")} {
		if err := os.Symlink(target, filepath.Join(root, link)); err != nil {
			t.Fatalf("symlink: %v", err)
		}
	}
	t.Setenv("MYCELIS_WORKSPACE", root)
	return root, outside
}

func mcplAssertOutside(t *testing.T, h *mcpsH, identity *RequestIdentity, tool, body string) blockerEnvelope {
	t.Helper()
	env := mcpsAssertBlocker(t, h.callTool(identity, tool, body), http.StatusForbidden, codeMCPPathOutsideWorkspace)
	h.assertCalls(0)
	return env
}

func TestMcplReadPathsOutsideWorkspaceAreRefused(t *testing.T) {
	_, outside := mcplWorkspace(t)
	cases := map[string]string{
		"dotdot alias":     `{"path":"workspace/../../etc/shadow"}`,
		"dotdot relative":  `{"path":"../outside/secret.txt"}`,
		"absolute":         `{"path":"/etc/passwd"}`,
		"proc":             `{"path":"/proc/self/environ"}`,
		"repo env":         `{"path":"/repo/.env"}`,
		"env beside root":  `{"path":"workspace/../.env"}`,
		"backslash dotdot": `{"path":"workspace\\..\\..\\etc\\shadow"}`,
		"symlink out":      `{"path":"out/secret.txt"}`,
		"symlink dir out":  `{"path":"out"}`,
		"absolute outside": `{"path":"` + filepath.Join(outside, "secret.txt") + `"}`,
		"one bad in paths": `{"paths":["notes/a.md","../outside/secret.txt"]}`,
		"source outside":   `{"source":"/etc/hosts"}`,
	}
	for name, args := range cases {
		t.Run(name, func(t *testing.T) {
			h := newMCPSHarness(t, "filesystem", true, "read_text_file")
			h.expectResolve()
			env := mcplAssertOutside(t, h, standardUserIdentity(), "read_text_file", `{"arguments":`+args+`}`)
			h.assertMocksMet() // no audit and no exchange item for a low-risk refusal
			assertUserSafe(t, "path outside", userVisible(env)...)
			if strings.Contains(env.Error+env.Data["recommended_action"].(string), "secret") {
				t.Fatalf("blocker echoes the path: %+v", env)
			}
		})
	}
}

func TestMcplNonStringPathArgumentIsRefused(t *testing.T) {
	mcplWorkspace(t)
	for name, args := range map[string]string{"number": `{"path":7}`, "object": `{"path":{"p":"/etc"}}`, "paths not list": `{"paths":"/etc/passwd"}`, "paths number": `{"paths":[1]}`} {
		t.Run(name, func(t *testing.T) {
			h := newMCPSHarness(t, "filesystem", true, "read_text_file")
			h.expectResolve()
			mcplAssertOutside(t, h, standardUserIdentity(), "read_text_file", `{"arguments":`+args+`}`)
			h.assertMocksMet()
		})
	}
}

func TestMcplReadPathsInsideWorkspaceAreForwardedAbsolute(t *testing.T) {
	root, _ := mcplWorkspace(t)
	cases := map[string]struct{ args, key, want string }{
		"relative":        {`{"path":"notes/a.md"}`, "path", filepath.Join(root, "notes", "a.md")},
		"alias":           {`{"path":"workspace/notes/a.md"}`, "path", filepath.Join(root, "notes", "a.md")},
		"slash alias":     {`{"path":"/workspace/notes"}`, "path", filepath.Join(root, "notes")},
		"absolute inside": {`{"path":"` + filepath.Join(root, "notes") + `"}`, "path", filepath.Join(root, "notes")},
		"inner symlink":   {`{"path":"inner/a.md"}`, "path", filepath.Join(root, "inner", "a.md")},
		"tilde stays in":  {`{"path":"~/.ssh/id_rsa"}`, "path", filepath.Join(root, "~", ".ssh", "id_rsa")},
		"empty is root":   {`{"path":""}`, "path", root},
		"new file":        {`{"path":"notes/new/deep.md"}`, "path", filepath.Join(root, "notes", "new", "deep.md")},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			h := newMCPSHarness(t, "filesystem", true, "read_text_file")
			h.expectResolve()
			h.expectExchange()
			assertStatus(t, h.callTool(standardUserIdentity(), "read_text_file", `{"arguments":`+c.args+`}`), http.StatusOK)
			if got := h.assertCalls(1)[0].Args[c.key]; got != c.want {
				t.Fatalf("forwarded %s = %v, want %s", c.key, got, c.want)
			}
			h.assertMocksMet()
		})
	}
}

// An approver's write outside the root is refused before the "called"
// audit: a best-effort refusal audit (keys only) and zero pool calls.
func TestMcplApproverWriteOutsideWorkspaceIsRefusedAndAudited(t *testing.T) {
	mcplWorkspace(t)
	cases := map[string]struct{ tool, args string }{
		"write dotdot":        {"write_file", `{"path":"../outside/x.txt","content":"c"}`},
		"write dangling link": {"write_file", `{"path":"dangle","content":"c"}`},
		"write through link":  {"write_file", `{"path":"out/new.txt","content":"c"}`},
		"move destination":    {"move_file", `{"source":"notes/a.md","destination":"/tmp/stolen.md"}`},
		"create dir proc":     {"create_directory", `{"path":"/proc/x"}`},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			h := newMCPSHarness(t, "filesystem", true, c.tool)
			h.expectResolve()
			refusal := h.expectAudit(false)
			env := mcplAssertOutside(t, h, mcpsWebAdmin(), c.tool, `{"arguments":`+c.args+`}`)
			h.assertMocksMet()
			ctx := refusal.object(t)
			if ctx["action"] != "mcp_tool_call_refused" || ctx["refusal_reason"] != codeMCPPathOutsideWorkspace || ctx["self_approved"] != false {
				t.Fatalf("refusal audit = %s", refusal.value())
			}
			for _, leak := range []string{"outside/", "stolen", "/proc"} {
				if strings.Contains(refusal.value(), leak) {
					t.Fatalf("refusal audit holds a path value %q: %s", leak, refusal.value())
				}
			}
			if env.Data["detail"] == nil {
				t.Fatalf("admin blocker has no detail: %+v", env)
			}
		})
	}
}

// The refusal audit is best-effort: an audit outage still refuses, 403.
func TestMcplApproverRefusalAuditFailureStillRefuses(t *testing.T) {
	mcplWorkspace(t)
	h := newMCPSHarness(t, "filesystem", true, "write_file")
	h.expectResolve()
	h.expectAudit(true)
	mcplAssertOutside(t, h, mcpsWebAdmin(), "write_file", `{"arguments":{"path":"/etc/cron.d/x"}}`)
	h.assertMocksMet()
}

// Authorization runs first: a non-approver's high-risk call with an outside
// path is admin_required, not a path answer.
func TestMcplNonApproverOutsidePathStaysAdminRequired(t *testing.T) {
	mcplWorkspace(t)
	h := newMCPSHarness(t, "filesystem", true, "write_file")
	h.expectResolve()
	h.expectAudit(false)
	mcpsAssertBlocker(t, h.callTool(standardUserIdentity(), "write_file", `{"arguments":{"path":"/etc/passwd"}}`), http.StatusForbidden, codeAdminRequired)
	h.assertCalls(0)
	h.assertMocksMet()
}

func TestMcplApproverWriteInsideWorkspaceRuns(t *testing.T) {
	root, _ := mcplWorkspace(t)
	h := newMCPSHarness(t, "filesystem", true, "write_file")
	h.expectResolve()
	h.expectAudit(false)
	h.expectExchange()
	assertStatus(t, h.callTool(mcpsWebAdmin(), "write_file", `{"arguments":{"path":"notes/b.md","content":"c"}}`), http.StatusOK)
	if got := h.assertCalls(1)[0].Args["path"]; got != filepath.Join(root, "notes", "b.md") {
		t.Fatalf("forwarded path = %v", got)
	}
	h.assertMocksMet()
}

// Other servers are not filesystem servers: their arguments pass unchanged.
func TestMcplNonFilesystemServerArgumentsUnchanged(t *testing.T) {
	mcplWorkspace(t)
	h := newMCPSHarness(t, "github", true, "create_issue")
	h.expectResolve()
	h.expectAudit(false)
	h.expectExchange()
	assertStatus(t, h.callTool(mcpsWebAdmin(), "create_issue", `{"arguments":{"path":"../../etc/passwd"}}`), http.StatusOK)
	if got := h.assertCalls(1)[0].Args["path"]; got != "../../etc/passwd" {
		t.Fatalf("non-filesystem path rewritten: %v", got)
	}
}

func TestMcplPathBlockerCopy(t *testing.T) {
	c := blockerCopies[codeMCPPathOutsideWorkspace]
	user, admin := c.forViewer(false), c.forViewer(true)
	if user.Message == "" || user.Action == "" || admin.Message == user.Message || admin.Action == "" {
		t.Fatalf("incomplete copy %+v", c)
	}
	assertUserSafe(t, codeMCPPathOutsideWorkspace, user.Message, user.Action)
}

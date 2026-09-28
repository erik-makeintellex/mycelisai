package server

import (
	"database/sql/driver"
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"testing"
)

// MCPA-H (contract D4): only the entry's declared env keys reach Install, for
// every principal, matched exactly and case-sensitively.

const mcpahHostileValue = "--require /tmp/mcpah-pwn.js"

func TestMcpahEnv_UndeclaredKeysRejectedForAdminWithZeroInstall(t *testing.T) {
	cases := []struct {
		label, name, env, rejected string
		standard                   bool // entry comes from core/config/mcp-library.yaml
	}{
		{"node_options", "fetch", `{"NODE_OPTIONS":"` + mcpahHostileValue + `"}`, "NODE_OPTIONS", false},
		{"ld_preload", "filesystem", `{"LD_PRELOAD":"` + mcpahHostileValue + `"}`, "LD_PRELOAD", false},
		{"path", "fetch", `{"PATH":"` + mcpahHostileValue + `"}`, "PATH", false},
		{"empty_key", "fetch", `{"":"` + mcpahHostileValue + `"}`, mcpEnvEmptyKeyLabel, false},
		{"declared_plus_hostile", "tokened", `{"FETCH_TOKEN":"ok","NODE_OPTIONS":"` + mcpahHostileValue + `"}`, "NODE_OPTIONS", false},
		{"case_variant", "github", `{"github_personal_access_token":"` + mcpahHostileValue + `"}`, "github_personal_access_token", true},
		{"whitespace_variant", "github", `{" GITHUB_PERSONAL_ACCESS_TOKEN":"` + mcpahHostileValue + `"}`, " GITHUB_PERSONAL_ACCESS_TOKEN", true},
		{"standard_filesystem_node_options", "filesystem", `{"NODE_OPTIONS":"` + mcpahHostileValue + `"}`, "NODE_OPTIONS", true},
	}
	for _, tc := range cases {
		for _, route := range mcpahWriteRoutes() {
			opt, mock := withMCPDB(t)
			s := newTestServer(opt, mcpahLibrary())
			if tc.standard {
				s.MCPLibrary = loadStandardMCPLibrary(t)
			}
			body := `{"name":"` + tc.name + `","env":` + tc.env + `}`
			rr := doAuthenticatedRequest(t, route.handler(s), "POST", route.path, body)
			if rr.Code != http.StatusBadRequest {
				t.Fatalf("%s %s: status = %d, want 400; body %s", tc.label, route.name, rr.Code, rr.Body.String())
			}
			if strings.Contains(rr.Body.String(), "mcpah-pwn") {
				t.Fatalf("%s %s: response echoed an env value: %s", tc.label, route.name, rr.Body.String())
			}
			env := decodeBlocker(t, rr)
			if env.OK || env.Data["code"] != codeMCPEnvRejected {
				t.Fatalf("%s %s: data = %v, want code %s", tc.label, route.name, env.Data, codeMCPEnvRejected)
			}
			if env.Data["rejected_env_keys"] != tc.rejected {
				t.Fatalf("%s %s: rejected_env_keys = %q, want %q", tc.label, route.name, env.Data["rejected_env_keys"], tc.rejected)
			}
			mcpahAssertNoDBCalls(t, mock)
		}
	}
}

func TestMcpahEnv_RejectionNamesAllUndeclaredKeysSorted(t *testing.T) {
	opt, mock := withMCPDB(t)
	s := newTestServer(opt, mcpahLibrary())
	body := `{"name":"tokened","env":{"PATH":"x","FETCH_TOKEN":"ok","LD_PRELOAD":"y","":"z"}}`
	rr := doAuthenticatedRequest(t, http.HandlerFunc(s.handleMCPLibraryInstall), "POST", "/api/v1/mcp/library/install", body)
	assertStatus(t, rr, http.StatusBadRequest)
	env := decodeBlocker(t, rr)
	want := strings.Join([]string{mcpEnvEmptyKeyLabel, "LD_PRELOAD", "PATH"}, ", ")
	if env.Data["rejected_env_keys"] != want {
		t.Fatalf("rejected_env_keys = %q, want %q", env.Data["rejected_env_keys"], want)
	}
	if env.Data["allowed_env_keys"] != "FETCH_REGION, FETCH_TOKEN" {
		t.Fatalf("allowed_env_keys = %q", env.Data["allowed_env_keys"])
	}
	mcpahAssertNoDBCalls(t, mock)
}

// mcpahEnvKeys matches the Install env JSON argument by its exact key set.
type mcpahEnvKeys []string

func (m mcpahEnvKeys) Match(v driver.Value) bool {
	raw, ok := v.([]byte)
	if !ok {
		return false
	}
	var env map[string]string
	if json.Unmarshal(raw, &env) != nil {
		return false
	}
	got := make([]string, 0, len(env))
	for k := range env {
		got = append(got, k)
	}
	sort.Strings(got)
	return strings.Join(got, ",") == strings.Join(m, ",")
}

func TestMcpahEnv_DeclaredKeysReachInstallAndStayRedacted(t *testing.T) {
	for _, route := range mcpahWriteRoutes() {
		opt, mock := withMCPDB(t)
		s := newTestServer(opt, mcpahLibrary())
		mcpahExpectInstall(mock, "tokened", mcpahEnvKeys{"FETCH_REGION", "FETCH_TOKEN"})
		body := `{"name":"tokened","env":{"FETCH_TOKEN":"mcpah-declared-secret","FETCH_REGION":"eu"}}`
		rr := doAuthenticatedRequest(t, route.handler(s), "POST", route.path, body)
		assertStatus(t, rr, http.StatusOK)
		if strings.Contains(rr.Body.String(), "mcpah-declared-secret") {
			t.Fatalf("%s: response leaked a declared env value: %s", route.name, rr.Body.String())
		}
		mcpahAssertNoDBCalls(t, mock)
	}
}

func TestMcpahEnv_DeclaredGitHubKeyPassesEnvCheck(t *testing.T) {
	for _, route := range mcpahWriteRoutes() {
		opt, mock := withMCPDB(t)
		s := newTestServer(opt, func(s *AdminServer) { s.MCPLibrary = loadStandardMCPLibrary(t) })
		body := `{"name":"github","env":{"GITHUB_PERSONAL_ACCESS_TOKEN":"mcpah-gh-value"}}`
		rr := doAuthenticatedRequest(t, route.handler(s), "POST", route.path, body)
		// The existing external_saas approval boundary (202) is unchanged by MCPA-H.
		assertStatus(t, rr, http.StatusAccepted)
		if strings.Contains(rr.Body.String(), "mcpah-gh-value") {
			t.Fatalf("%s: response echoed the token value", route.name)
		}
		mcpahAssertNoDBCalls(t, mock)
	}
}

func TestMcpahEnv_RejectedCopyIsRoleAwareAndPlain(t *testing.T) {
	user := mcpEnvRejectedCopy.User
	if user.Message == "" || user.Action == "" || mcpEnvRejectedCopy.Admin.Message == "" {
		t.Fatalf("mcp_env_rejected needs user message+action and admin message: %+v", mcpEnvRejectedCopy)
	}
	assertUserSafe(t, codeMCPEnvRejected, user.Message, user.Action)
	for _, bad := range []string{"env", "NODE_OPTIONS", "variable"} {
		if strings.Contains(strings.ToLower(user.Message+user.Action), strings.ToLower(bad)) {
			t.Fatalf("user copy names %q: %+v", bad, user)
		}
	}
}

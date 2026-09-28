package server

import (
	"net/http"
	"sort"
	"strings"

	"github.com/mycelis/core/internal/mcp"
)

// MCP configuration authority (MCPA-H hotfix). Installing or applying a
// curated MCP server launches a process inside Core, so it is a root-admin
// action and the caller may only set the env keys the library entry declares.

// scopeMCPConfigWrite gates MCP configuration writes. It is deliberately not
// "mcp:write", which parses as an agent tool grant (mcp.ParseToolRef).
const scopeMCPConfigWrite = "mcp_config:write"

// codeMCPEnvRejected is the 400 when a request carries env keys the library
// entry does not declare. Nothing is installed or launched.
const codeMCPEnvRejected = "mcp_env_rejected"

// mcpEnvEmptyKeyLabel stands in for an empty env key in rejected_env_keys.
const mcpEnvEmptyKeyLabel = "(empty)"

// mcpEnvRejectedCopy is local so blocker_copy.go stays untouched while MCPS
// edits it; respondBlockerText takes explicit copy.
var mcpEnvRejectedCopy = roleBlockerText{
	User: blockerText{"This tool can't be set up with those settings.",
		"Remove the settings this tool doesn't use and try again. Nothing was installed."},
	Admin: blockerText{"This install sets environment variables the library entry does not declare.",
		"Send only the keys in allowed_env_keys (exact spelling and case); rejected_env_keys lists the others. Nothing was installed or started."},
}

// requireMCPConfigWrite admits only a root admin holding mcp_config:write
// (web admins, the API key and break-glass hold "*"). No identity is 401 and
// everyone else is 403 admin_required. Call it before any other check.
func requireMCPConfigWrite(w http.ResponseWriter, r *http.Request) (*RequestIdentity, bool) {
	return requireRootAdminScope(w, r, scopeMCPConfigWrite)
}

// undeclaredMCPEnvKeys returns the sorted request env keys that the entry does
// not declare. Matching is exact and case-sensitive; an empty key is never
// declared.
func undeclaredMCPEnvKeys(entry *mcp.LibraryEntry, env map[string]string) []string {
	if len(env) == 0 {
		return nil
	}
	declared := map[string]struct{}{}
	if entry != nil {
		for _, key := range entry.DeclaredEnvKeys() {
			declared[key] = struct{}{}
		}
	}
	var rejected []string
	for key := range env {
		if key == "" {
			rejected = append(rejected, mcpEnvEmptyKeyLabel)
			continue
		}
		if _, ok := declared[key]; !ok {
			rejected = append(rejected, key)
		}
	}
	sort.Strings(rejected)
	return rejected
}

// requireDeclaredMCPEnv writes 400 mcp_env_rejected and returns false when
// the request env holds any undeclared key. The response names keys only,
// never values.
func requireDeclaredMCPEnv(w http.ResponseWriter, r *http.Request, entry *mcp.LibraryEntry, env map[string]string) bool {
	rejected := undeclaredMCPEnvKeys(entry, env)
	if len(rejected) == 0 {
		return true
	}
	allowed := []string{}
	if entry != nil {
		allowed = entry.DeclaredEnvKeys()
	}
	sort.Strings(allowed)
	respondBlockerText(w, r, http.StatusBadRequest, codeMCPEnvRejected, mcpEnvRejectedCopy, "", map[string]string{
		"rejected_env_keys": strings.Join(rejected, ", "),
		"allowed_env_keys":  strings.Join(allowed, ", "),
	})
	return false
}

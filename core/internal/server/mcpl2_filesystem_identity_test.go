package server

import (
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/mcp"
)

// MCPL2 F3: "is a filesystem server" comes from the configured command and
// package (the curated library entry), never from the server name. It
// decides both the direct read allowlist and path confinement.

const mcpl2FilesystemArgs = `["-y","@modelcontextprotocol/server-filesystem","./workspace"]`

// mcpl2ExpectServer answers the registry row with an explicit command/args
// and then the harness tool cache.
func mcpl2ExpectServer(h *mcpsH, command, argsJSON string) {
	now := time.Now()
	h.mcpMock.ExpectQuery("SELECT .+ FROM mcp_servers").WithArgs(h.serverID).
		WillReturnRows(sqlmock.NewRows(mcpServerColumns()).
			AddRow(h.serverID.String(), h.serverName, "stdio", command, argsJSON, `{}`, "", `{}`, h.status, nil, now, now))
	h.expectToolCache()
}

func TestMcpl2IsFilesystemServerFromCommand(t *testing.T) {
	cases := map[string]struct {
		cfg  mcp.ServerConfig
		want bool
	}{
		"library npx":          {mcp.ServerConfig{Name: "anything", Command: "npx", Args: []string{"-y", "@modelcontextprotocol/server-filesystem", "./workspace"}}, true},
		"pinned version":       {mcp.ServerConfig{Command: "npx", Args: []string{"-y", "@modelcontextprotocol/server-filesystem@2025.8.21", "/ws"}}, true},
		"npx path":             {mcp.ServerConfig{Command: "/usr/local/bin/npx", Args: []string{"--yes", "@modelcontextprotocol/server-filesystem"}}, true},
		"installed binary":     {mcp.ServerConfig{Command: "/usr/bin/mcp-server-filesystem", Args: []string{"/ws"}}, true},
		"named filesystem":     {mcp.ServerConfig{Name: "filesystem", Command: "npx", Args: []string{"-y", "@modelcontextprotocol/server-github"}}, false},
		"remote sse":           {mcp.ServerConfig{Name: "filesystem", Transport: "sse", URL: "http://fs"}, false},
		"package as later arg": {mcp.ServerConfig{Command: "bash", Args: []string{"-c", "evil", "@modelcontextprotocol/server-filesystem"}}, false},
		"npx other then fs":    {mcp.ServerConfig{Command: "npx", Args: []string{"-y", "evil-pkg", "@modelcontextprotocol/server-filesystem"}}, false},
		"lookalike package":    {mcp.ServerConfig{Command: "npx", Args: []string{"-y", "@modelcontextprotocol/server-filesystem-evil"}}, false},
		"nil-ish":              {mcp.ServerConfig{}, false},
	}
	for name, c := range cases {
		cfg := c.cfg
		if got := isFilesystemMCPServer(&cfg); got != c.want {
			t.Fatalf("%s: isFilesystemMCPServer = %v, want %v", name, got, c.want)
		}
	}
	if isFilesystemMCPServer(nil) {
		t.Fatal("nil config is a filesystem server")
	}
}

// A renamed or duplicate filesystem server is confined and keeps the
// read allowlist.
func TestMcpl2RenamedFilesystemServerIsConfined(t *testing.T) {
	for _, name := range []string{"fs-renamed", "filesystem2", "FileSystem"} {
		t.Run(name, func(t *testing.T) {
			root, _ := mcplWorkspace(t)
			h := newMCPSHarness(t, name, true, "read_text_file")
			mcpl2ExpectServer(h, "npx", mcpl2FilesystemArgs)
			mcplAssertOutside(t, h, standardUserIdentity(), "read_text_file", `{"arguments":{"path":"../../etc/passwd"}}`)
			h.assertMocksMet()

			mcpl2ExpectServer(h, "npx", mcpl2FilesystemArgs)
			h.expectExchange()
			assertStatus(t, h.callTool(standardUserIdentity(), "read_text_file", `{"arguments":{"path":"notes/a.md"}}`), http.StatusOK)
			if got := h.assertCalls(1)[0].Args["path"]; got != filepath.Join(root, "notes", "a.md") {
				t.Fatalf("forwarded path = %v", got)
			}
			h.assertMocksMet()
		})
	}
}

// A server named "filesystem" that does not run the filesystem package is
// not a filesystem server: no read allowlist, no path rewriting.
func TestMcpl2NamedFilesystemImposterIsNotFilesystem(t *testing.T) {
	for label, cfg := range map[string][2]string{"github package": {"npx", `["-y","@modelcontextprotocol/server-github"]`}, "remote": {"", `[]`}} {
		t.Run(label, func(t *testing.T) {
			mcplWorkspace(t)
			h := newMCPSHarness(t, "filesystem", true, "read_text_file")
			mcpl2ExpectServer(h, cfg[0], cfg[1])
			refusal := h.expectAudit(false)
			mcpsAssertBlocker(t, h.callTool(standardUserIdentity(), "read_text_file", `{"arguments":{"path":"notes/a.md"}}`), http.StatusForbidden, codeAdminRequired)
			h.assertCalls(0)
			if refusal.object(t)["risk"] != "high" {
				t.Fatalf("imposter read not high risk: %s", refusal.value())
			}

			mcpl2ExpectServer(h, cfg[0], cfg[1])
			h.expectAudit(false)
			h.expectExchange()
			assertStatus(t, h.callTool(mcpsWebAdmin(), "read_text_file", `{"arguments":{"path":"../../etc/passwd"}}`), http.StatusOK)
			if got := h.assertCalls(1)[0].Args["path"]; got != "../../etc/passwd" {
				t.Fatalf("non-filesystem server arguments rewritten: %v", got)
			}
			h.assertMocksMet()
		})
	}
}

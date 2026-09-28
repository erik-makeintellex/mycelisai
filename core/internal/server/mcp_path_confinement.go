package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/mycelis/core/internal/mcp"
)

// MCPL: Core-side filesystem path confinement for direct MCP calls. The
// upstream filesystem MCP server also confines paths; Core does not rely on
// it. Whether a server is a filesystem server comes from its configured
// command and package (MCPL2), never from its name.

// mcpFilesystemPackages are the npm packages of the curated filesystem
// library entry (core/config/mcp-library.yaml).
var mcpFilesystemPackages = []string{"@modelcontextprotocol/server-filesystem"}

// mcpFilesystemBinaries are installed binaries of those packages.
var mcpFilesystemBinaries = []string{"mcp-server-filesystem"}

// mcpPackageRunners launch a package named by their first non-flag argument.
var mcpPackageRunners = map[string]bool{"npx": true, "bunx": true, "pnpx": true}

// isFilesystemMCPServer: the command is a filesystem binary, or a package
// runner whose first non-flag argument is the filesystem package (optionally
// pinned with @version). A package named anywhere else does not count.
func isFilesystemMCPServer(cfg *mcp.ServerConfig) bool {
	if cfg == nil {
		return false
	}
	command := strings.TrimSuffix(strings.ToLower(path.Base(strings.ReplaceAll(strings.TrimSpace(cfg.Command), "\\", "/"))), ".cmd")
	for _, binary := range mcpFilesystemBinaries {
		if command == binary {
			return true
		}
	}
	if !mcpPackageRunners[command] {
		return false
	}
	for _, arg := range cfg.Args {
		arg = strings.TrimSpace(arg)
		if strings.HasPrefix(arg, "-") {
			continue
		}
		for _, pkg := range mcpFilesystemPackages {
			if arg == pkg || strings.HasPrefix(arg, pkg+"@") {
				return true
			}
		}
		return false
	}
	return false
}

// mcpFilesystemValueKeys are the non-path keys the curated filesystem tools
// declare (lower-cased); they pass unchanged.
var mcpFilesystemValueKeys = map[string]bool{"content": true, "edits": true, "dryrun": true, "head": true,
	"tail": true, "sortby": true, "pattern": true, "excludepatterns": true}

// mcpPathKeyTokens mark a key as path-bearing (matched case-insensitively
// as substrings, so "Path", "filePath", "rootDir" and "paths" all match).
var mcpPathKeyTokens = []string{"path", "file", "dir", "folder", "root", "uri", "cwd", "target", "dest", "source", "location", "mount"}

func mcpPathBearingKey(lower string) bool {
	if mcpFilesystemValueKeys[lower] {
		return false
	}
	for _, token := range mcpPathKeyTokens {
		if strings.Contains(lower, token) {
			return true
		}
	}
	return false
}

// errMCPWorkspaceRoot marks a workspace root Core cannot resolve at all.
var errMCPWorkspaceRoot = errors.New("the workspace root could not be resolved")

// mcpPathError names the refused argument key, never its value. Unknown
// marks an undeclared key whose value looks like a path (400); otherwise the
// path leaves the workspace or is malformed (403).
type mcpPathError struct {
	Key, Reason string
	Unknown     bool
}

func (e *mcpPathError) Error() string { return "argument " + strconv.Quote(e.Key) + " " + e.Reason }

func (e *mcpPathError) code() string {
	if e.Unknown {
		return codeMCPPathArgumentUnknown
	}
	return codeMCPPathOutsideWorkspace
}

// normalizeMCPToolCallArgumentsForServer confines a filesystem server's
// arguments; other servers' arguments pass unchanged. Every path-bearing key
// (case-insensitive; declared in toolSchema or matching mcpPathKeyTokens)
// becomes an absolute path under the MYCELIS_WORKSPACE root; "paths"-style
// keys need a list and are confined per element. Known value keys and keys
// the cached tool schema declares pass; any other key whose value looks like
// a path is refused.
func normalizeMCPToolCallArgumentsForServer(filesystem bool, toolSchema json.RawMessage, args map[string]any) (map[string]any, error) {
	if !filesystem || len(args) == 0 {
		return args, nil
	}
	root, err := resolveMCPWorkspaceRoot()
	if err != nil {
		return nil, err
	}
	declared := mcpSchemaKeys(toolSchema)
	out := make(map[string]any, len(args))
	for key, value := range args {
		lower := strings.ToLower(strings.TrimSpace(key))
		switch {
		case mcpPathBearingKey(lower):
			if out[key], err = root.confineValue(key, lower, value); err != nil {
				return nil, err
			}
		case mcpFilesystemValueKeys[lower] || declared[lower]:
			out[key] = value
		case mcpLooksLikePath(value):
			return nil, &mcpPathError{Key: key, Reason: "is not a known argument of this tool and looks like a path", Unknown: true}
		default:
			out[key] = value
		}
	}
	return out, nil
}

// confineValue: a "paths"-style key needs a list of path strings; other path
// keys take one path string.
func (ws mcpWorkspaceRoot) confineValue(key, lower string, value any) (any, error) {
	if !strings.HasSuffix(lower, "paths") {
		return ws.confine(key, value)
	}
	list, ok := value.([]any)
	if !ok {
		return nil, &mcpPathError{Key: key, Reason: "is not a list of paths"}
	}
	paths := make([]any, len(list))
	for i, item := range list {
		confined, err := ws.confine(key, item)
		if err != nil {
			return nil, err
		}
		paths[i] = confined
	}
	return paths, nil
}

// mcpSchemaKeys are the lower-cased property names of a cached tool input
// schema; an unreadable schema declares nothing (the code-owned lists apply).
func mcpSchemaKeys(schema json.RawMessage) map[string]bool {
	var parsed struct {
		Properties map[string]json.RawMessage `json:"properties"`
	}
	keys := map[string]bool{}
	if len(schema) == 0 || json.Unmarshal(schema, &parsed) != nil {
		return keys
	}
	for name := range parsed.Properties {
		keys[strings.ToLower(strings.TrimSpace(name))] = true
	}
	return keys
}

// mcpLooksLikePath reports whether any string in value (recursively) has a
// path shape: absolute, drive or UNC, home-relative, dot-led, a file: URL, or
// any separator outside a non-file URL.
func mcpLooksLikePath(value any) bool {
	switch typed := value.(type) {
	case string:
		s := strings.TrimSpace(typed)
		lower := strings.ToLower(s)
		if strings.HasPrefix(lower, "file:") {
			return true
		}
		if strings.Contains(s, "://") {
			return false
		}
		return strings.HasPrefix(s, ".") || strings.HasPrefix(s, "~") || strings.ContainsAny(s, `/\`) ||
			(len(s) >= 2 && s[1] == ':')
	case []any:
		for _, item := range typed {
			if mcpLooksLikePath(item) {
				return true
			}
		}
	case map[string]any:
		for _, item := range typed {
			if mcpLooksLikePath(item) {
				return true
			}
		}
	}
	return false
}

// mcpWorkspaceRoot is the configured root, absolute (abs) and with the
// symlinks of its existing prefix resolved (real).
type mcpWorkspaceRoot struct{ abs, real string }

func resolveMCPWorkspaceRoot() (mcpWorkspaceRoot, error) {
	abs, err := filepath.Abs(strings.TrimSpace(mcp.ResolveFilesystemWorkspaceRoot()))
	if err != nil {
		return mcpWorkspaceRoot{}, errMCPWorkspaceRoot
	}
	real, err := resolveExistingMCPPath(abs)
	if err != nil {
		return mcpWorkspaceRoot{}, errMCPWorkspaceRoot
	}
	return mcpWorkspaceRoot{abs: abs, real: real}, nil
}

// confine returns the absolute path the upstream server receives, after
// checking it lexically against the root and, following symlinks, against
// the root's real path. Non-strings and dangling symlinks are refused.
func (ws mcpWorkspaceRoot) confine(key string, raw any) (string, error) {
	text, ok := raw.(string)
	if !ok {
		return "", &mcpPathError{Key: key, Reason: "is not a path string"}
	}
	candidate := ws.lexical(text)
	if !mcpPathWithin(ws.abs, candidate) {
		return "", &mcpPathError{Key: key, Reason: "resolves outside the workspace root"}
	}
	real, err := resolveExistingMCPPath(candidate)
	if err != nil {
		return "", &mcpPathError{Key: key, Reason: "could not be resolved inside the workspace root"}
	}
	if !mcpPathWithin(ws.real, real) {
		return "", &mcpPathError{Key: key, Reason: "follows a symlink outside the workspace root"}
	}
	return candidate, nil
}

// lexical maps a caller path onto an absolute, cleaned path: the workspace
// aliases and relative paths join the root; absolute paths stay absolute.
func (ws mcpWorkspaceRoot) lexical(raw string) string {
	p := strings.TrimPrefix(strings.ReplaceAll(strings.TrimSpace(raw), "\\", "/"), "./")
	switch {
	case p == "" || p == "." || p == "workspace" || p == "/workspace":
		return ws.abs
	case strings.HasPrefix(p, "workspace/"):
		p = strings.TrimPrefix(p, "workspace/")
	case strings.HasPrefix(p, "/workspace/"):
		p = strings.TrimPrefix(p, "/workspace/")
	case filepath.IsAbs(filepath.FromSlash(p)):
		return filepath.Clean(filepath.FromSlash(p))
	}
	return filepath.Join(ws.abs, filepath.FromSlash(p))
}

// resolveExistingMCPPath evaluates symlinks on the longest existing prefix
// of the absolute, cleaned path p and appends the missing remainder. An
// entry that exists but does not resolve (a dangling symlink) is an error,
// because a write through it would land wherever it points.
func resolveExistingMCPPath(p string) (string, error) {
	current, rest := p, []string{}
	for {
		real, err := filepath.EvalSymlinks(current)
		if err == nil {
			return filepath.Join(append([]string{real}, rest...)...), nil
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return "", err
		}
		if _, lerr := os.Lstat(current); lerr == nil {
			return "", fmt.Errorf("dangling symlink")
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", err
		}
		rest = append([]string{filepath.Base(current)}, rest...)
		current = parent
	}
}

func mcpPathWithin(root, p string) bool {
	rel, err := filepath.Rel(root, p)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
}

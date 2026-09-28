package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"path/filepath"
	"reflect"
	"testing"
)

// MCPL2 F2: on a filesystem server every path-bearing key is confined,
// matched case-insensitively; unknown keys whose values look like paths are
// refused; known value keys pass unchanged.

func mcpl2Normalize(schema string, args map[string]any) (map[string]any, error) {
	return normalizeMCPToolCallArgumentsForServer(true, json.RawMessage(schema), args)
}

func mcpl2Code(err error) string {
	var pathErr *mcpPathError
	if errors.As(err, &pathErr) {
		return pathErr.code()
	}
	return ""
}

func TestMcpl2PathKeyVariantsAreConfined(t *testing.T) {
	root, _ := mcplWorkspace(t)
	keys := []string{"Path", "PATH", " path", "path ", "Source", "Destination", "file", "filePath", "file_path",
		"target", "dir", "directory", "root", "uri", "cwd", "folder", "rootPath", "location"}
	for _, key := range keys {
		if _, err := mcpl2Normalize(`{}`, map[string]any{key: "/etc/passwd"}); mcpl2Code(err) != codeMCPPathOutsideWorkspace {
			t.Fatalf("key %q outside path: err = %v, want %s", key, err, codeMCPPathOutsideWorkspace)
		}
		out, err := mcpl2Normalize(`{}`, map[string]any{key: "notes/a.md"})
		if err != nil || out[key] != filepath.Join(root, "notes", "a.md") {
			t.Fatalf("key %q inside path: %v %v", key, out, err)
		}
	}
}

func TestMcpl2PathsVariantsConfinedPerElement(t *testing.T) {
	root, _ := mcplWorkspace(t)
	for _, key := range []string{"paths", "Paths", "PATHS"} {
		out, err := mcpl2Normalize(`{}`, map[string]any{key: []any{"notes/a.md", "workspace/b.md"}})
		if err != nil || !reflect.DeepEqual(out[key], []any{filepath.Join(root, "notes", "a.md"), filepath.Join(root, "b.md")}) {
			t.Fatalf("%s: %v %v", key, out, err)
		}
		for name, bad := range map[string]any{"string": "/etc/passwd", "outside element": []any{"notes/a.md", "/etc/passwd"},
			"nested list": []any{[]any{"/etc"}}, "map element": []any{map[string]any{"path": "/etc"}}, "nil": nil} {
			if _, err := mcpl2Normalize(`{}`, map[string]any{key: bad}); err == nil {
				t.Fatalf("%s %s accepted", key, name)
			}
		}
	}
}

func TestMcpl2ValueKeysPassUnchanged(t *testing.T) {
	mcplWorkspace(t)
	args := map[string]any{"path": "notes/a.md", "content": "/etc/passwd is not a path here", "pattern": "**/*.md",
		"excludePatterns": []any{"node_modules/**"}, "edits": []any{map[string]any{"oldText": "../a", "newText": "/b"}},
		"dryRun": true, "head": 3.0, "tail": 2.0, "sortBy": "size", "token": "abc", "url": "https://u:p@h"}
	out, err := mcpl2Normalize(`{}`, args)
	if err != nil {
		t.Fatalf("value keys refused: %v", err)
	}
	for key, want := range args {
		if key != "path" && !reflect.DeepEqual(out[key], want) {
			t.Fatalf("value key %q changed: %v", key, out[key])
		}
	}
}

func TestMcpl2UnknownPathLikeKeysAreRefused(t *testing.T) {
	mcplWorkspace(t)
	for name, value := range map[string]any{
		"absolute": "/etc/passwd", "dotdot": "../outside", "tilde": "~/.ssh/id_rsa", "relative slash": "notes/a.md",
		"backslash": `..\outside`, "drive": `C:\Windows`, "file url": "file:///etc/passwd", "in list": []any{"ok", "/etc"},
		"nested map": map[string]any{"inner": "/etc/shadow"},
	} {
		_, err := mcpl2Normalize(`{}`, map[string]any{"mystery": value})
		if mcpl2Code(err) != codeMCPPathArgumentUnknown {
			t.Fatalf("%s: err = %v, want %s", name, err, codeMCPPathArgumentUnknown)
		}
	}
	for name, value := range map[string]any{"word": "hello", "number": 4.0, "bool": false, "url": "https://example.com/x", "nil": nil} {
		if _, err := mcpl2Normalize(`{}`, map[string]any{"mystery": value}); err != nil {
			t.Fatalf("%s: non-path unknown value refused: %v", name, err)
		}
	}
}

// A key the cached tool schema declares is a known key: path-ish names are
// confined, other declared keys pass even with slashes in the value.
func TestMcpl2SchemaDeclaredKeys(t *testing.T) {
	root, _ := mcplWorkspace(t)
	schema := `{"type":"object","properties":{"mode":{"type":"string"},"baseDir":{"type":"string"}}}`
	out, err := mcpl2Normalize(schema, map[string]any{"mode": "a/b", "baseDir": "notes"})
	if err != nil || out["mode"] != "a/b" || out["baseDir"] != filepath.Join(root, "notes") {
		t.Fatalf("declared keys: %v %v", out, err)
	}
	if _, err := mcpl2Normalize(schema, map[string]any{"baseDir": "/etc"}); mcpl2Code(err) != codeMCPPathOutsideWorkspace {
		t.Fatalf("declared path key outside: %v", err)
	}
	if _, err := mcpl2Normalize(`not json`, map[string]any{"mode": "a/b"}); mcpl2Code(err) != codeMCPPathArgumentUnknown {
		t.Fatalf("unreadable schema must fall back to the code-owned list: %v", err)
	}
}

func TestMcpl2NonFilesystemArgumentsUntouched(t *testing.T) {
	args := map[string]any{"Path": "/etc/passwd", "mystery": "../x"}
	out, err := normalizeMCPToolCallArgumentsForServer(false, nil, args)
	if err != nil || !reflect.DeepEqual(out, args) {
		t.Fatalf("non-filesystem args changed: %v %v", out, err)
	}
}

// Route: the 400 and 403 are normalized blockers; nothing runs; a high call
// gets a keys-only refusal audit naming the code.
func TestMcpl2RouteRefusesKeyVariantsAndUnknownPaths(t *testing.T) {
	mcplWorkspace(t)
	cases := map[string]struct {
		identity *RequestIdentity
		tool     string
		body     string
		status   int
		code     string
	}{
		"read Path outside":    {standardUserIdentity(), "read_text_file", `{"arguments":{"Path":"/etc/passwd"}}`, http.StatusForbidden, codeMCPPathOutsideWorkspace},
		"read unknown key":     {standardUserIdentity(), "read_text_file", `{"arguments":{"path":"notes/a.md","file_ref":"x","other":"/etc/passwd"}}`, http.StatusBadRequest, codeMCPPathArgumentUnknown},
		"write root outside":   {mcpsWebAdmin(), "write_file", `{"arguments":{"path":"notes/a.md","root":"/"}}`, http.StatusForbidden, codeMCPPathOutsideWorkspace},
		"write unknown path":   {mcpsWebAdmin(), "write_file", `{"arguments":{"path":"notes/a.md","content":"c","extra":"../../x"}}`, http.StatusBadRequest, codeMCPPathArgumentUnknown},
		"top-level PATH input": {standardUserIdentity(), "read_text_file", `{"PATH":"/etc/passwd"}`, http.StatusForbidden, codeMCPPathOutsideWorkspace},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			h := newMCPSHarness(t, "filesystem", true, c.tool)
			h.expectResolve()
			var refusal *mcpsCapture
			if c.identity.Role == "admin" {
				refusal = h.expectAudit(false)
			}
			env := mcpsAssertBlocker(t, h.callTool(c.identity, c.tool, c.body), c.status, c.code)
			h.assertCalls(0)
			h.assertMocksMet()
			if c.identity.Role != "admin" {
				assertUserSafe(t, c.code, userVisible(env)...)
			}
			if refusal != nil && refusal.object(t)["refusal_reason"] != c.code {
				t.Fatalf("refusal audit = %s", refusal.value())
			}
		})
	}
}

func TestMcpl2UnknownKeyBlockerCopy(t *testing.T) {
	user, admin := mcpPathArgumentUnknownCopy.forViewer(false), mcpPathArgumentUnknownCopy.forViewer(true)
	if user.Message == "" || user.Action == "" || admin.Message == user.Message || admin.Action == "" {
		t.Fatalf("incomplete copy %+v", mcpPathArgumentUnknownCopy)
	}
	assertUserSafe(t, codeMCPPathArgumentUnknown, user.Message, user.Action)
}

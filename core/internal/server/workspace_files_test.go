package server

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandleWorkspaceFileView_ServesSandboxedHTML(t *testing.T) {
	workspace := t.TempDir()
	t.Setenv("MYCELIS_WORKSPACE", workspace)
	target := filepath.Join(workspace, "logs", "game.html")
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(target, []byte("<!doctype html><button id=coin>Coin</button>"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	s := newTestServer()
	mux := setupMux(t, "GET /api/v1/workspace/files/view", s.HandleWorkspaceFileView)
	rr := doRequest(t, mux, http.MethodGet, "/api/v1/workspace/files/view?path=workspace/logs/game.html", "")

	assertStatus(t, rr, http.StatusOK)
	if !strings.Contains(rr.Body.String(), "coin") {
		t.Fatalf("body = %q", rr.Body.String())
	}
	if got := rr.Header().Get("Content-Security-Policy"); !strings.Contains(got, "sandbox allow-scripts") {
		t.Fatalf("content security policy = %q", got)
	}
	if got := rr.Header().Get("X-Mycelis-Workspace-Path"); got != "logs/game.html" {
		t.Fatalf("workspace path header = %q", got)
	}
}

func TestHandleWorkspaceFileView_RejectsEscapingPath(t *testing.T) {
	t.Setenv("MYCELIS_WORKSPACE", t.TempDir())

	s := newTestServer()
	mux := setupMux(t, "GET /api/v1/workspace/files/view", s.HandleWorkspaceFileView)
	rr := doRequest(t, mux, http.MethodGet, "/api/v1/workspace/files/view?path=../secret.txt", "")

	assertStatus(t, rr, http.StatusBadRequest)
}

func TestHandleWorkspaceFileReveal_OpensContainingFolderWithinWorkspace(t *testing.T) {
	workspace := t.TempDir()
	t.Setenv("MYCELIS_WORKSPACE", workspace)
	t.Setenv("MYCELIS_WORKSPACE_REVEAL_DRY_RUN", "1")
	target := filepath.Join(workspace, "logs", "generated.html")
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(target, []byte("<!doctype html>"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	s := newTestServer()
	mux := setupMux(t, "POST /api/v1/workspace/files/reveal", s.HandleWorkspaceFileReveal)
	rr := doAuthenticatedRequest(t, mux, http.MethodPost, "/api/v1/workspace/files/reveal?path=workspace/logs/generated.html", "")

	assertStatus(t, rr, http.StatusOK)
	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["ok"] != true {
		t.Fatalf("body = %#v", body)
	}
	data, ok := body["data"].(map[string]any)
	if !ok {
		t.Fatalf("data = %#v", body["data"])
	}
	if data["workspace_path"] != "logs/generated.html" {
		t.Fatalf("workspace_path = %#v", data["workspace_path"])
	}
	if data["folder_path"] != filepath.Join(workspace, "logs") {
		t.Fatalf("folder_path = %#v", data["folder_path"])
	}
}

func TestHandleWorkspaceFileReveal_OpensWorkspaceRoot(t *testing.T) {
	workspace := t.TempDir()
	t.Setenv("MYCELIS_WORKSPACE", workspace)
	t.Setenv("MYCELIS_WORKSPACE_REVEAL_DRY_RUN", "1")

	s := newTestServer()
	mux := setupMux(t, "POST /api/v1/workspace/files/reveal", s.HandleWorkspaceFileReveal)
	rr := doAuthenticatedRequest(t, mux, http.MethodPost, "/api/v1/workspace/files/reveal?path=workspace", "")

	assertStatus(t, rr, http.StatusOK)
	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	data, ok := body["data"].(map[string]any)
	if !ok {
		t.Fatalf("data = %#v", body["data"])
	}
	if data["workspace_path"] != "." {
		t.Fatalf("workspace_path = %#v", data["workspace_path"])
	}
	if data["folder_path"] != workspace {
		t.Fatalf("folder_path = %#v", data["folder_path"])
	}
}

func TestHandleWorkspaceFileReveal_RejectsEscapingPath(t *testing.T) {
	t.Setenv("MYCELIS_WORKSPACE", t.TempDir())
	t.Setenv("MYCELIS_WORKSPACE_REVEAL_DRY_RUN", "1")

	s := newTestServer()
	mux := setupMux(t, "POST /api/v1/workspace/files/reveal", s.HandleWorkspaceFileReveal)
	rr := doAuthenticatedRequest(t, mux, http.MethodPost, "/api/v1/workspace/files/reveal?path=../secret.txt", "")

	assertStatus(t, rr, http.StatusBadRequest)
}

func TestHandleWorkspaceFileView_ServesTextOutputsAsReadableTypes(t *testing.T) {
	workspace := t.TempDir()
	t.Setenv("MYCELIS_WORKSPACE", workspace)
	s := newTestServer()
	mux := setupMux(t, "GET /api/v1/workspace/files/view", s.HandleWorkspaceFileView)
	htmlCSP := ""
	cases := []struct{ name, want string }{
		{"game.html", "text/html; charset=utf-8"},
		{"team-brief.md", "text/plain; charset=utf-8"},
		{"notes.markdown", "text/plain; charset=utf-8"},
		{"notes.txt", "text/plain; charset=utf-8"},
		{"rows.csv", "text/plain; charset=utf-8"},
		{"run.log", "text/plain; charset=utf-8"},
		{"conf.yaml", "text/plain; charset=utf-8"},
		{"conf.yml", "text/plain; charset=utf-8"},
		{"conf.toml", "text/plain; charset=utf-8"},
		{"feed.xml", "text/plain; charset=utf-8"},
		{"data.json", "application/json"},
		{"blob.unknownext", "application/octet-stream"},
	}
	for _, tc := range cases {
		if err := os.WriteFile(filepath.Join(workspace, tc.name), []byte("# body\n"), 0o644); err != nil {
			t.Fatalf("write %s: %v", tc.name, err)
		}
		rr := doRequest(t, mux, http.MethodGet, "/api/v1/workspace/files/view?path=workspace/"+tc.name, "")
		assertStatus(t, rr, http.StatusOK)
		if got := rr.Header().Get("Content-Type"); got != tc.want {
			t.Fatalf("%s content type = %q, want %q", tc.name, got, tc.want)
		}
		if got := rr.Header().Get("X-Content-Type-Options"); got != "nosniff" {
			t.Fatalf("%s nosniff = %q", tc.name, got)
		}
		csp := rr.Header().Get("Content-Security-Policy")
		if htmlCSP == "" {
			htmlCSP = csp
		}
		if !strings.Contains(csp, "sandbox allow-scripts") || csp != htmlCSP {
			t.Fatalf("%s CSP drifted: %q vs html %q", tc.name, csp, htmlCSP)
		}
	}
}

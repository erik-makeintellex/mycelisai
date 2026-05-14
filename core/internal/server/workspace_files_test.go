package server

import (
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

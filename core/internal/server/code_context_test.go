package server

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mycelis/core/internal/codecontext"
)

func TestHandleCodeContextQueryReturnsRefs(t *testing.T) {
	root := t.TempDir()
	writeCodeContextFile(t, filepath.Join(root, "pkg", "service.go"), "package pkg\n\nfunc BuildWidget() string { return \"ok\" }\n")
	s := newTestServer(func(s *AdminServer) {
		s.CodeContext = codecontext.NewService(codecontext.Config{SourceRoots: []string{root}})
	})
	mux := setupMux(t, "POST /api/v1/code-context/query", s.HandleCodeContextQuery)

	rr := doAuthenticatedRequest(t, mux, http.MethodPost, "/api/v1/code-context/query", `{"query":"BuildWidget","limit":3}`)

	assertStatus(t, rr, http.StatusOK)
	if !strings.Contains(rr.Body.String(), `"file_path":"pkg/service.go"`) ||
		!strings.Contains(rr.Body.String(), `"raw_graph_internals":"not_exposed"`) {
		t.Fatalf("body = %s", rr.Body.String())
	}
}

func TestHandleCodeContextSourcesRegistersConfigDocument(t *testing.T) {
	root := t.TempDir()
	repo := filepath.Join(root, "repo")
	writeCodeContextFile(t, filepath.Join(repo, "main.go"), "package main\n\nfunc main() {}\n")
	s := newTestServer(func(s *AdminServer) {
		s.CodeContext = codecontext.NewService(codecontext.Config{SourceRoots: []string{root}})
	})
	mux := setupMux(t, "POST /api/v1/code-context/sources", s.HandleCodeContextSources)
	body := `{"document":{
		"apiVersion":"mycelis.ai/v1",
		"kind":"CodeContextSource",
		"metadata":{
			"id":"code-context-source",
			"name":"Repo source",
			"version":"1.0.0",
			"owner_id":"operator-1",
			"scope":{"kind":"workspace","ref":"workspace-1"},
			"enabled":true,
			"source":{"kind":"file","ref":"code-context/repo.yaml"},
			"governance":{"risk_level":"medium","approval_posture":"required"}
		},
		"spec":{"source_id":"repo-source","source_type":"repository","root_path":"` + filepath.ToSlash(repo) + `"}
	}}`

	rr := doAuthenticatedRequest(t, mux, http.MethodPost, "/api/v1/code-context/sources", body)

	assertStatus(t, rr, http.StatusCreated)
	if !strings.Contains(rr.Body.String(), `"id":"repo-source"`) ||
		!strings.Contains(rr.Body.String(), `"config_digest":"sha256:`) {
		t.Fatalf("body = %s", rr.Body.String())
	}
}

func writeCodeContextFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
}

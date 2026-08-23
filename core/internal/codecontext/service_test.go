package codecontext

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestServiceQueryReturnsBoundedSourceRefs(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "core", "widget.go"), "package core\n\nfunc BuildWidget() string { return \"ok\" }\n")

	svc := NewService(Config{SourceRoots: []string{root}})
	resp, err := svc.Query(context.Background(), Request{Query: "BuildWidget", Limit: 3})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if resp.Status != "ok" || resp.Count == 0 {
		t.Fatalf("resp = %+v", resp)
	}
	if resp.Refs[0].FilePath != "core/widget.go" {
		t.Fatalf("ref path = %q", resp.Refs[0].FilePath)
	}
	if resp.Source == nil || resp.Source.Root != "" {
		t.Fatalf("source should not expose root: %+v", resp.Source)
	}
	if resp.Metadata["raw_graph_internals"] != "not_exposed" {
		t.Fatalf("metadata = %+v", resp.Metadata)
	}
}

func TestServiceImpactSeparatesInferredRelationships(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "api.go"), "package app\n\nfunc Handler() {}\n")
	writeFile(t, filepath.Join(root, "api_test.go"), "package app\n\nfunc TestHandler() { Handler() }\n")

	svc := NewService(Config{SourceRoots: []string{root}})
	resp, err := svc.Impact(context.Background(), Request{Target: "Handler", Limit: 5})
	if err != nil {
		t.Fatalf("Impact: %v", err)
	}
	if len(resp.InferredRelationships) == 0 {
		t.Fatalf("expected inferred relationships: %+v", resp)
	}
	if resp.InferredRelationships[0].Provenance == "" || resp.InferredRelationships[0].Confidence == "" {
		t.Fatalf("inferred impact missing provenance/confidence: %+v", resp.InferredRelationships[0])
	}
}

func TestServiceExplainExtractsFactsWithoutEscapingBoundary(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "pkg", "service.go"), "package pkg\n\nimport \"context\"\n\ntype Service struct{}\n\nfunc Query(ctx context.Context) {}\n")

	svc := NewService(Config{SourceRoots: []string{root}})
	resp, err := svc.Explain(context.Background(), Request{Path: "pkg/service.go"})
	if err != nil {
		t.Fatalf("Explain: %v", err)
	}
	if resp.Status != "ok" || len(resp.ExtractedFacts) == 0 {
		t.Fatalf("resp = %+v", resp)
	}
	escaped, err := svc.Explain(context.Background(), Request{Path: "../outside.go"})
	if err != nil {
		t.Fatalf("Explain escaped: %v", err)
	}
	if escaped.Status != "blocked" || escaped.Blocker == nil {
		t.Fatalf("escaped resp = %+v", escaped)
	}
}

func TestServiceRegistersSourceInsideApprovedRoot(t *testing.T) {
	root := t.TempDir()
	sourceRoot := filepath.Join(root, "repo")
	writeFile(t, filepath.Join(sourceRoot, "main.go"), "package main\n\nfunc main() {}\n")

	svc := NewService(Config{SourceRoots: []string{root}})
	source, err := svc.RegisterSource(context.Background(), SourceInput{
		ID:         "repo-source",
		Name:       "Repo source",
		SourceType: "repository",
		RootPath:   sourceRoot,
		ScopeKind:  "workspace",
	})
	if err != nil {
		t.Fatalf("RegisterSource: %v", err)
	}
	if source.Root != "" || source.ID != "repo-source" || source.SnapshotRef == "" {
		t.Fatalf("public source = %+v", source)
	}
	index, err := svc.Index(context.Background(), "repo-source")
	if err != nil {
		t.Fatalf("Index: %v", err)
	}
	if index.Status != "ok" || index.Metadata["storage_model"] == "" {
		t.Fatalf("index = %+v", index)
	}

	outside := t.TempDir()
	if _, err := svc.RegisterSource(context.Background(), SourceInput{ID: "outside-source", Name: "Outside", SourceType: "repository", RootPath: outside}); err == nil {
		t.Fatal("RegisterSource outside approved root succeeded")
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
}

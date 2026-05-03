package mcp

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestToolSetService_ResolveRefs_Mixed(t *testing.T) {
	svc, mock := newTestToolSetService(t)
	now := time.Now()

	mock.ExpectQuery("SELECT .+ FROM mcp_tool_sets WHERE name").
		WithArgs("workspace").
		WillReturnRows(sqlmock.NewRows(toolSetColumns()).
			AddRow(tsID1, "workspace", "File I/O", `["mcp:filesystem/*"]`, "default", now, now))

	tools := []string{"read_file", "mcp:github/*", "toolset:workspace", "consult_council"}
	resolved, err := svc.ResolveRefs(context.Background(), tools)
	if err != nil {
		t.Fatalf("ResolveRefs: %v", err)
	}
	want := []string{"read_file", "mcp:github/*", "mcp:filesystem/*", "consult_council"}
	if len(resolved) != len(want) {
		t.Fatalf("got %d refs, want %d: %v", len(resolved), len(want), resolved)
	}
	for i, expected := range want {
		if resolved[i] != expected {
			t.Errorf("resolved[%d] = %q, want %q", i, resolved[i], expected)
		}
	}
}

func TestToolSetService_ResolveRefs_MissingToolSet(t *testing.T) {
	svc, mock := newTestToolSetService(t)

	mock.ExpectQuery("SELECT .+ FROM mcp_tool_sets WHERE name").
		WithArgs("unknown").
		WillReturnError(sql.ErrNoRows)

	resolved, err := svc.ResolveRefs(context.Background(), []string{"mcp:filesystem/*", "toolset:unknown"})
	if err != nil {
		t.Fatalf("ResolveRefs: %v", err)
	}
	if len(resolved) != 1 {
		t.Fatalf("got %d refs, want 1: %v", len(resolved), resolved)
	}
}

func TestToolSetService_ResolveRefs_NoToolSets(t *testing.T) {
	svc, _ := newTestToolSetService(t)

	tools := []string{"mcp:filesystem/*", "read_file", "mcp:github/create_issue"}
	resolved, err := svc.ResolveRefs(context.Background(), tools)
	if err != nil {
		t.Fatalf("ResolveRefs: %v", err)
	}
	if len(resolved) != 3 {
		t.Fatalf("got %d, want 3", len(resolved))
	}
}

func TestToolSetService_NilDB(t *testing.T) {
	svc := NewToolSetService(nil)

	checks := []struct {
		name string
		err  error
	}{
		{"List", firstErr(svc.List(context.Background()))},
		{"Get", firstErr(svc.Get(context.Background(), tsID1))},
		{"Create", firstErr(svc.Create(context.Background(), ToolSet{Name: "x"}))},
		{"Update", firstErr(svc.Update(context.Background(), tsID1, ToolSet{Name: "x"}))},
		{"Delete", svc.Delete(context.Background(), tsID1)},
		{"FindByName", firstErr(svc.FindByName(context.Background(), "x"))},
	}
	for _, check := range checks {
		if check.err == nil {
			t.Errorf("%s: expected error with nil DB", check.name)
		}
	}
}

func firstErr[T any](_ T, err error) error {
	return err
}

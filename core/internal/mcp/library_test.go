package mcp

import (
	"os"
	"path/filepath"
	"testing"
)

// ── LoadLibrary ──────────────────────────────────────────────

func TestLoadLibrary_ValidYAML(t *testing.T) {
	// Load the real config file used in production.
	lib, err := LoadLibrary(filepath.Join("..", "..", "config", "mcp-library.yaml"))
	if err != nil {
		t.Fatalf("LoadLibrary: %v", err)
	}
	if len(lib.Categories) == 0 {
		t.Fatal("expected at least one category")
	}
	// Spot-check a known entry.
	fs := lib.FindByName("filesystem")
	if fs == nil {
		t.Fatal("expected to find 'filesystem' entry")
	}
	if fs.Transport != "stdio" {
		t.Errorf("filesystem transport = %q, want stdio", fs.Transport)
	}
}

func TestLoadLibrary_FileNotFound(t *testing.T) {
	_, err := LoadLibrary("/nonexistent/path/library.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadLibrary_InvalidYAML(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "bad.yaml")
	if err := os.WriteFile(tmp, []byte("{{not valid yaml"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := LoadLibrary(tmp)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

// ── FindByName ───────────────────────────────────────────────

func TestLibrary_FindByName_Found(t *testing.T) {
	lib := &Library{
		Categories: []LibraryCategory{
			{Name: "Dev", Servers: []LibraryEntry{
				{Name: "filesystem", Transport: "stdio", Command: "npx", Args: []string{"-y", "mcp-fs"}},
				{Name: "github", Transport: "stdio", Command: "npx"},
			}},
		},
	}
	entry := lib.FindByName("filesystem")
	if entry == nil {
		t.Fatal("expected to find filesystem")
	}
	if entry.Command != "npx" {
		t.Errorf("command = %q, want npx", entry.Command)
	}
}

func TestLibrary_FindByName_NotFound(t *testing.T) {
	lib := &Library{Categories: []LibraryCategory{
		{Name: "Empty", Servers: []LibraryEntry{}},
	}}
	if lib.FindByName("nonexistent") != nil {
		t.Error("expected nil for unknown server")
	}
}

// ── ToServerConfig ───────────────────────────────────────────

func TestLibraryEntry_ToServerConfig_Basic(t *testing.T) {
	entry := LibraryEntry{
		Name:      "filesystem",
		Transport: "stdio",
		Command:   "npx",
		Args:      []string{"-y", "@mcp/server-fs", "./workspace"},
		Env:       map[string]string{"FOO": "bar"},
	}
	cfg := entry.ToServerConfig(nil)
	if cfg.Name != "filesystem" {
		t.Errorf("name = %q", cfg.Name)
	}
	if cfg.Transport != "stdio" {
		t.Errorf("transport = %q", cfg.Transport)
	}
	if cfg.Env["FOO"] != "bar" {
		t.Errorf("env FOO = %q", cfg.Env["FOO"])
	}
	if len(cfg.Args) != 3 {
		t.Errorf("args len = %d, want 3", len(cfg.Args))
	}
}

func TestLibraryEntry_ToServerConfig_EnvOverride(t *testing.T) {
	entry := LibraryEntry{
		Name:      "github",
		Transport: "stdio",
		Command:   "npx",
		Env:       map[string]string{"TOKEN": "", "OTHER": "keep"},
	}
	cfg := entry.ToServerConfig(map[string]string{"TOKEN": "my-secret"})
	if cfg.Env["TOKEN"] != "my-secret" {
		t.Errorf("TOKEN = %q, want my-secret", cfg.Env["TOKEN"])
	}
	if cfg.Env["OTHER"] != "keep" {
		t.Errorf("OTHER = %q, want keep", cfg.Env["OTHER"])
	}
}

func TestLibraryEntry_ToServerConfig_NoEnv(t *testing.T) {
	entry := LibraryEntry{
		Name:      "fetch",
		Transport: "stdio",
		Command:   "npx",
	}
	cfg := entry.ToServerConfig(nil)
	if cfg.Env == nil {
		t.Fatal("expected non-nil env map")
	}
	if len(cfg.Env) != 0 {
		t.Errorf("expected empty env, got %d entries", len(cfg.Env))
	}
}

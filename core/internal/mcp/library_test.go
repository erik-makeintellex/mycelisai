package mcp

import (
	"os"
	"path/filepath"
	"testing"
)

func loadStandardLibraryForTest(t *testing.T) *Library {
	t.Helper()
	lib, err := LoadLibrary(filepath.Join("..", "..", "config", "mcp-library.yaml"))
	if err != nil {
		t.Fatalf("LoadLibrary: %v", err)
	}
	return lib
}

// ── LoadLibrary ──────────────────────────────────────────────

func TestLoadLibrary_ValidYAML(t *testing.T) {
	// Load the real config file used in production.
	lib := loadStandardLibraryForTest(t)
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
	if fs.Version == "" {
		t.Fatal("expected filesystem version metadata")
	}
	if len(fs.Packages) == 0 {
		t.Fatal("expected filesystem package metadata")
	}
}

func TestLoadLibrary_StandardEntriesRemainLocalFirst(t *testing.T) {
	lib := loadStandardLibraryForTest(t)

	for _, name := range []string{"filesystem", "fetch"} {
		entry := lib.FindByName(name)
		if entry == nil {
			t.Fatalf("expected to find %q in standard library", name)
		}
		if entry.Transport != "stdio" {
			t.Fatalf("%s transport = %q, want stdio", name, entry.Transport)
		}
		if entry.URL != "" {
			t.Fatalf("%s url = %q, want empty for local-first stdio entry", name, entry.URL)
		}
		if entry.Command == "" {
			t.Fatalf("%s command is empty", name)
		}
	}
}

func TestLoadLibrary_StandardCredentialedEntriesDeclareExternalBoundary(t *testing.T) {
	lib := loadStandardLibraryForTest(t)

	for _, name := range []string{"github", "brave-search", "slack", "flux", "elevenlabs", "replicate", "dall-e"} {
		entry := lib.FindByName(name)
		if entry == nil {
			t.Fatalf("expected to find %q in standard library", name)
		}
		if entry.DeploymentBoundary != "external_saas" {
			t.Fatalf("%s deployment_boundary = %q, want external_saas", name, entry.DeploymentBoundary)
		}
		if !entry.HasRequiredSecretEnvVar() {
			t.Fatalf("%s should declare a required secret credential", name)
		}
	}
}

func TestLoadLibrary_StandardToolSetLinksRemainStable(t *testing.T) {
	lib := loadStandardLibraryForTest(t)

	filesystem := lib.FindByName("filesystem")
	if filesystem == nil || filesystem.ToolSet != "workspace" {
		t.Fatalf("filesystem tool_set = %q, want workspace", filesystem.ToolSet)
	}

	fetch := lib.FindByName("fetch")
	if fetch == nil || fetch.ToolSet != "research" {
		t.Fatalf("fetch tool_set = %q, want research", fetch.ToolSet)
	}
}

func TestLoadLibrary_StandardEntriesExposeServerJSONStyleMetadata(t *testing.T) {
	lib := loadStandardLibraryForTest(t)

	github := lib.FindByName("github")
	if github == nil {
		t.Fatal("expected to find github entry")
	}
	if github.Title != "GitHub" {
		t.Fatalf("title = %q, want GitHub", github.Title)
	}
	if github.Version == "" {
		t.Fatal("expected github version metadata")
	}
	if len(github.Packages) != 1 {
		t.Fatalf("packages len = %d, want 1", len(github.Packages))
	}
	if github.Packages[0].Identifier != "@modelcontextprotocol/server-github" {
		t.Fatalf("identifier = %q, want @modelcontextprotocol/server-github", github.Packages[0].Identifier)
	}
	if github.Packages[0].Transport.Type != "stdio" {
		t.Fatalf("transport.type = %q, want stdio", github.Packages[0].Transport.Type)
	}
	if github.Repository == "" {
		t.Fatal("expected github repository metadata")
	}
	if github.Homepage == "" {
		t.Fatal("expected github homepage metadata")
	}
	if len(github.EnvironmentVariables) != 1 {
		t.Fatalf("environment_variables len = %d, want 1", len(github.EnvironmentVariables))
	}
	if !github.EnvironmentVariables[0].Required || !github.EnvironmentVariables[0].Secret {
		t.Fatalf("expected github token to be required and secret")
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

func TestLibraryEntry_ToServerConfig_CopiesURL(t *testing.T) {
	entry := LibraryEntry{
		Name:      "remote-knowledge",
		Transport: "sse",
		URL:       "https://mcp.example.com/sse",
	}

	cfg := entry.ToServerConfig(nil)
	if cfg.URL != "https://mcp.example.com/sse" {
		t.Fatalf("url = %q, want https://mcp.example.com/sse", cfg.URL)
	}
}

func TestLibraryEntry_ToServerConfig_UsesEnvironmentVariableDefaults(t *testing.T) {
	entry := LibraryEntry{
		Name:      "stable-diffusion",
		Transport: "stdio",
		Command:   "npx",
		EnvironmentVariables: []LibraryEnvVar{
			{Name: "SD_API_URL", DefaultValue: "http://127.0.0.1:7860"},
		},
	}

	cfg := entry.ToServerConfig(nil)
	if cfg.Env["SD_API_URL"] != "http://127.0.0.1:7860" {
		t.Fatalf("SD_API_URL = %q, want default value", cfg.Env["SD_API_URL"])
	}
}

func TestLibraryEntry_DeclaredEnvKeys_DeduplicatesLegacyAndTypedEnv(t *testing.T) {
	entry := LibraryEntry{
		Env: map[string]string{
			"TOKEN": "",
			"URL":   "http://127.0.0.1",
		},
		EnvironmentVariables: []LibraryEnvVar{
			{Name: "TOKEN", Required: true, Secret: true},
			{Name: "EXTRA"},
		},
	}

	keys := entry.DeclaredEnvKeys()
	if len(keys) != 3 {
		t.Fatalf("len(keys) = %d, want 3", len(keys))
	}
}

func TestLibraryEntry_HasRequiredSecretEnvVar(t *testing.T) {
	entry := LibraryEntry{
		EnvironmentVariables: []LibraryEnvVar{
			{Name: "PUBLIC_URL", Required: true},
			{Name: "API_TOKEN", Required: true, Secret: true},
		},
	}

	if !entry.HasRequiredSecretEnvVar() {
		t.Fatal("expected required secret env var to be detected")
	}
}

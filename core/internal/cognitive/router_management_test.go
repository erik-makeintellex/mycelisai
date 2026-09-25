package cognitive

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSaveConfig_NeverWritesRawAPIKey proves the L-F8 negative case: an
// api_key populated in memory (e.g. from MYCELIS_PROVIDER_X_API_KEY via
// env_overrides.go, or a UI/API save) must never be written back into
// cognitive.yaml, a tracked file. Only api_key_env may round-trip.
func TestSaveConfig_NeverWritesRawAPIKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cognitive.yaml")

	r := &Router{
		ConfigPath: path,
		Config: &BrainConfig{
			Providers: map[string]ProviderConfig{
				"openai": {
					Type:       "openai",
					ModelID:    "gpt-4.1-mini",
					AuthKey:    "sk-super-secret-should-never-persist",
					AuthKeyEnv: "OPENAI_API_KEY",
					Enabled:    true,
				},
			},
			Profiles: map[string]string{"chat": "openai"},
		},
	}

	if err := r.SaveConfig(); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read saved config: %v", err)
	}
	written := string(data)
	if strings.Contains(written, "sk-super-secret-should-never-persist") {
		t.Fatalf("raw api_key leaked into cognitive.yaml:\n%s", written)
	}
	if !strings.Contains(written, "OPENAI_API_KEY") {
		t.Fatalf("expected api_key_env to still round-trip, got:\n%s", written)
	}

	// In-memory config must be unaffected by the redaction copy.
	if r.Config.Providers["openai"].AuthKey != "sk-super-secret-should-never-persist" {
		t.Fatal("SaveConfig must not mutate the live in-memory AuthKey")
	}
}

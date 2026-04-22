package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveDatabaseConfigDefaults(t *testing.T) {
	t.Setenv("DB_HOST", "")
	t.Setenv("DB_PORT", "")
	t.Setenv("DB_USER", "")
	t.Setenv("DB_PASSWORD", "")
	t.Setenv("DB_NAME", "")

	cfg := resolveDatabaseConfig()

	if cfg.Host != "mycelis-core-postgresql" {
		t.Fatalf("Host = %q, want default", cfg.Host)
	}
	if cfg.Port != "5432" {
		t.Fatalf("Port = %q, want default", cfg.Port)
	}
	if cfg.User != "mycelis" {
		t.Fatalf("User = %q, want default", cfg.User)
	}
	if cfg.Password != "password" {
		t.Fatalf("Password = %q, want default", cfg.Password)
	}
	if cfg.Name != "cortex" {
		t.Fatalf("Name = %q, want default", cfg.Name)
	}
}

func TestResolveDatabaseConfigUsesEnvironment(t *testing.T) {
	t.Setenv("DB_HOST", "postgres")
	t.Setenv("DB_PORT", "5544")
	t.Setenv("DB_USER", "compose-user")
	t.Setenv("DB_PASSWORD", "compose-pass")
	t.Setenv("DB_NAME", "compose-db")

	cfg := resolveDatabaseConfig()

	if cfg.Host != "postgres" || cfg.Port != "5544" || cfg.User != "compose-user" || cfg.Password != "compose-pass" || cfg.Name != "compose-db" {
		t.Fatalf("unexpected config: %#v", cfg)
	}

	got := cfg.connectionString()
	want := "postgres://compose-user:compose-pass@postgres:5544/compose-db"
	if got != want {
		t.Fatalf("connectionString = %q, want %q", got, want)
	}
}

func TestDefaultMCPBootstrapEnabledDefaultsToTrue(t *testing.T) {
	t.Setenv("MYCELIS_DISABLE_DEFAULT_MCP_BOOTSTRAP", "")

	if !defaultMCPBootstrapEnabled() {
		t.Fatal("defaultMCPBootstrapEnabled() = false, want true")
	}
}

func TestDefaultMCPBootstrapEnabledHonorsDisableFlag(t *testing.T) {
	for _, value := range []string{"1", "true", "yes", "on"} {
		t.Run(value, func(t *testing.T) {
			t.Setenv("MYCELIS_DISABLE_DEFAULT_MCP_BOOTSTRAP", value)
			if defaultMCPBootstrapEnabled() {
				t.Fatalf("defaultMCPBootstrapEnabled() = true for %q, want false", value)
			}
		})
	}
}

func TestResolveLocalAuthRuntimeConfigDefaults(t *testing.T) {
	t.Setenv("MYCELIS_API_KEY", "primary")
	t.Setenv("MYCELIS_LOCAL_ADMIN_USERNAME", "")
	t.Setenv("MYCELIS_LOCAL_ADMIN_USER_ID", "")
	t.Setenv("MYCELIS_BREAK_GLASS_API_KEY", "")
	t.Setenv("MYCELIS_BREAK_GLASS_USERNAME", "")
	t.Setenv("MYCELIS_BREAK_GLASS_USER_ID", "")

	cfg := resolveLocalAuthRuntimeConfig()

	if cfg.PrimaryAPIKey != "primary" {
		t.Fatalf("PrimaryAPIKey = %q, want primary", cfg.PrimaryAPIKey)
	}
	if cfg.PrimaryUsername != "admin" {
		t.Fatalf("PrimaryUsername = %q, want default admin", cfg.PrimaryUsername)
	}
	if cfg.PrimaryUserID != "00000000-0000-0000-0000-000000000000" {
		t.Fatalf("PrimaryUserID = %q, want default", cfg.PrimaryUserID)
	}
	if cfg.BreakGlassAPIKey != "" {
		t.Fatalf("BreakGlassAPIKey = %q, want empty", cfg.BreakGlassAPIKey)
	}
	if cfg.BreakGlassUsername != "recovery-admin" {
		t.Fatalf("BreakGlassUsername = %q, want default recovery-admin", cfg.BreakGlassUsername)
	}
	if cfg.BreakGlassUserID != "00000000-0000-0000-0000-000000000001" {
		t.Fatalf("BreakGlassUserID = %q, want default", cfg.BreakGlassUserID)
	}
}

func TestLocalAuthRuntimeConfigWarnsOnPartialBreakGlassConfig(t *testing.T) {
	t.Setenv("MYCELIS_API_KEY", "primary")
	t.Setenv("MYCELIS_BREAK_GLASS_API_KEY", "break")
	t.Setenv("MYCELIS_BREAK_GLASS_USERNAME", "recovery-admin")
	t.Setenv("MYCELIS_BREAK_GLASS_USER_ID", "")

	cfg := resolveLocalAuthRuntimeConfig()
	warnings := cfg.breakGlassWarnings()

	if len(warnings) == 0 {
		t.Fatal("expected warnings for partial break-glass config")
	}
}

func TestLocalAuthRuntimeConfigWarnsOnDuplicateBreakGlassKey(t *testing.T) {
	t.Setenv("MYCELIS_API_KEY", "same-key")
	t.Setenv("MYCELIS_BREAK_GLASS_API_KEY", "same-key")
	t.Setenv("MYCELIS_BREAK_GLASS_USERNAME", "recovery-admin")
	t.Setenv("MYCELIS_BREAK_GLASS_USER_ID", "00000000-0000-0000-0000-000000000001")

	cfg := resolveLocalAuthRuntimeConfig()
	warnings := cfg.breakGlassWarnings()

	found := false
	for _, warning := range warnings {
		if warning == "break-glass API key matches MYCELIS_API_KEY; use a distinct recovery credential" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected duplicate-key warning, got %#v", warnings)
	}
}

func TestLocalAuthRuntimeConfigWarnsWhenDeploymentContractRequiresBreakGlass(t *testing.T) {
	t.Setenv("MYCELIS_API_KEY", "primary")
	t.Setenv("MYCELIS_IDENTITY_MODE", "hybrid")
	t.Setenv("MYCELIS_BREAK_GLASS_API_KEY", "")

	cfg := resolveLocalAuthRuntimeConfig()
	warnings := cfg.breakGlassWarnings()

	found := false
	for _, warning := range warnings {
		if warning == "deployment auth contract requires MYCELIS_BREAK_GLASS_API_KEY for enterprise-like recovery posture" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected deployment-contract warning, got %#v", warnings)
	}
}

func TestResolveLocalAuthRuntimeConfigReadsDeploymentContractPath(t *testing.T) {
	contractPath := filepath.Join(t.TempDir(), "deployment-contract.json")
	if err := os.WriteFile(contractPath, []byte(`{"identity_mode":"federated"}`), 0o644); err != nil {
		t.Fatalf("write deployment contract: %v", err)
	}
	t.Setenv("MYCELIS_DEPLOYMENT_CONTRACT_PATH", contractPath)
	t.Setenv("MYCELIS_API_KEY", "primary")
	t.Setenv("MYCELIS_BREAK_GLASS_API_KEY", "")

	cfg := resolveLocalAuthRuntimeConfig()
	warnings := cfg.breakGlassWarnings()

	found := false
	for _, warning := range warnings {
		if warning == "deployment auth contract requires MYCELIS_BREAK_GLASS_API_KEY for enterprise-like recovery posture" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected path-driven deployment-contract warning, got %#v", warnings)
	}
}

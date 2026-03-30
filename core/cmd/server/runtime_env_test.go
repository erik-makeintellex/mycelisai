package main

import "testing"

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

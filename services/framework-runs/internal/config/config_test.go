package config

import "testing"

func TestFromEnvRequiresDatabaseAndCanonicalCredential(t *testing.T) {
	t.Setenv("FRAMEWORK_RUNS_DATABASE_URL", "")
	t.Setenv("FRAMEWORK_RUNS_CORE_TOKEN", "")
	if _, err := FromEnv(); err == nil {
		t.Fatal("missing database and credential were accepted")
	}
	t.Setenv("FRAMEWORK_RUNS_DATABASE_URL", "postgres://service.invalid/runs")
	t.Setenv("FRAMEWORK_RUNS_CORE_TOKEN", " short credential with whitespace ")
	if _, err := FromEnv(); err == nil {
		t.Fatal("noncanonical credential was accepted")
	}
}

func TestFromEnvAppliesBoundedOperationalSettings(t *testing.T) {
	t.Setenv("FRAMEWORK_RUNS_DATABASE_URL", "postgres://service.invalid/runs")
	t.Setenv("FRAMEWORK_RUNS_CORE_TOKEN", "0123456789abcdef0123456789abcdef")
	t.Setenv("FRAMEWORK_RUNS_MAX_RUNS", "64")
	t.Setenv("FRAMEWORK_RUNS_LEASE_SECONDS", "45")
	settings, err := FromEnv()
	if err != nil || settings.MaxRuns != 64 || settings.LeaseDuration.Seconds() != 45 {
		t.Fatalf("settings = %#v, %v", settings, err)
	}
	for key, value := range map[string]string{
		"FRAMEWORK_RUNS_MAX_RUNS": "0", "FRAMEWORK_RUNS_LEASE_SECONDS": "3601",
	} {
		t.Setenv("FRAMEWORK_RUNS_MAX_RUNS", "64")
		t.Setenv("FRAMEWORK_RUNS_LEASE_SECONDS", "45")
		t.Setenv(key, value)
		if _, err := FromEnv(); err == nil {
			t.Fatalf("invalid %s was accepted", key)
		}
	}
}

package cognitive

import (
	"strings"
	"testing"
)

// UX1: RecommendedAction and Summary are read by anyone, so they never name
// an API path, env var, URL, permission string, provider id or internal term.
// The admin remedy stays in AdminAction; codes are unchanged.

var forbiddenAvailabilityCopy = []string{
	"/api/", "mycelis_", "://", "cognitive:", ".env", "override", "provider", "routed", "profile", "degraded", "swarm", "token",
}

func TestExecutionAvailability_UserCopyIsPlainAdminCopyActionable(t *testing.T) {
	providers := map[string]ProviderConfig{
		"off":     {Type: "openai_compatible", Enabled: false, ModelID: "m"},
		"nomodel": {Type: "openai_compatible", Enabled: true},
		"down":    {Type: "openai_compatible", Enabled: true, ModelID: "m"},
	}
	cases := []struct {
		name, provider, origin, wantCode, adminHint string
	}{
		{"none", "", "", ExecutionNoProviders, "AI Engines"},
		{"missing/root", "ghost", "", ExecutionProviderMissing, "Enable provider ghost"},
		{"disabled/db override", "off", ProfileOriginDB, ExecutionProviderDisabled, "DELETE /api/v1/cognitive/profiles/chat/override"},
		{"model missing/env pin", "nomodel", ProfileOriginEnv, ExecutionModelMissing, "MYCELIS_PROFILE_CHAT_PROVIDER"},
		{"not running", "down", "", ExecutionProviderOffline, "Enable provider down"},
	}
	for _, c := range cases {
		cfg := &BrainConfig{Providers: providers, Profiles: map[string]string{}, ProfileOverrideOrigins: map[string]string{}}
		if c.provider != "" {
			cfg.Profiles["chat"] = c.provider
		}
		if c.origin != "" {
			cfg.ProfileOverrideOrigins["chat"] = c.origin
		}
		r := &Router{Config: cfg, Adapters: map[string]LLMProvider{}}
		got := r.ExecutionAvailability("chat", "")
		if got.Available || got.Code != c.wantCode {
			t.Fatalf("%s: availability = %+v, want %s", c.name, got, c.wantCode)
		}
		assertPlainAvailability(t, c.name, got.Summary, got.RecommendedAction)
		if got.RecommendedAction != UserEngineSetupAction || !strings.Contains(got.AdminAction, c.adminHint) {
			t.Fatalf("%s: user=%q admin=%q, want admin hint %q", c.name, got.RecommendedAction, got.AdminAction, c.adminHint)
		}
	}

	var nilRouter *Router
	offline := nilRouter.ExecutionAvailability("chat", "")
	if offline.Code != ExecutionRouterUnavailable || offline.Summary != SummaryRouterUnavailable {
		t.Fatalf("router offline = %+v", offline)
	}
	assertPlainAvailability(t, "router offline", offline.Summary, offline.RecommendedAction)
}

func assertPlainAvailability(t *testing.T, label string, texts ...string) {
	t.Helper()
	for _, text := range texts {
		if strings.TrimSpace(text) == "" {
			t.Fatalf("%s: empty user copy", label)
		}
		lower := strings.ToLower(text)
		for _, bad := range forbiddenAvailabilityCopy {
			if strings.Contains(lower, bad) {
				t.Fatalf("%s: user copy %q contains %q", label, text, bad)
			}
		}
	}
}

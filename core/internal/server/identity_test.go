package server

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ── GET /api/v1/user/me ────────────────────────────────────────────

func TestHandleMe(t *testing.T) {
	s := newTestServer()
	rr := doAuthenticatedRequest(t, http.HandlerFunc(s.HandleMe), "GET", "/api/v1/user/me", "")
	assertStatus(t, rr, http.StatusOK)

	var user User
	assertJSON(t, rr, &user)
	if user.Username != "admin" {
		t.Errorf("Expected username 'admin', got %q", user.Username)
	}
	if user.Role != "admin" {
		t.Errorf("Expected role 'admin', got %q", user.Role)
	}
	if user.EffectiveRole != "owner" {
		t.Errorf("Expected effective_role 'owner', got %q", user.EffectiveRole)
	}
	if user.PrincipalType != "local_admin" {
		t.Errorf("Expected principal_type 'local_admin', got %q", user.PrincipalType)
	}
	if user.AuthSource != "local_api_key" {
		t.Errorf("Expected auth_source 'local_api_key', got %q", user.AuthSource)
	}
	if user.BreakGlass {
		t.Error("Expected break_glass false")
	}
	if user.ID == "" {
		t.Error("Expected non-empty user ID")
	}
	var settings map[string]any
	if err := json.Unmarshal(user.Settings, &settings); err != nil {
		t.Fatalf("unmarshal settings: %v", err)
	}
	for k, want := range map[string]string{
		"assistant_name":                 "Soma",
		"access_management_tier":         "release",
		"product_edition":                "self_hosted_release",
		"identity_mode":                  "local_only",
		"shared_agent_specificity_owner": "root_admin",
	} {
		if settings[k] != want {
			t.Errorf("Expected %s %s, got %#v", k, want, settings[k])
		}
	}
}

func TestHandleMe_HybridModeReportsBreakGlassPrincipal(t *testing.T) {
	s := newTestServer()
	rr := doAuthenticatedRequestAs(t, http.HandlerFunc(s.HandleMe), "GET", "/api/v1/user/me", "", breakGlassIdentityForTest())
	assertStatus(t, rr, http.StatusOK)

	var user User
	assertJSON(t, rr, &user)
	if user.PrincipalType != "break_glass_admin" || user.AuthSource != "local_break_glass" || !user.BreakGlass || user.EffectiveRole != "owner" {
		t.Errorf("unexpected break glass user: %+v", user)
	}
}

func TestHandleMe_UsesDeploymentContractSettings(t *testing.T) {
	s := newTestServer()
	contractPath := filepath.Join(t.TempDir(), "deployment-contract.json")
	if err := os.WriteFile(contractPath, []byte(`{
  "access_management_tier": "enterprise",
  "product_edition": "self_hosted_enterprise",
  "identity_mode": "hybrid",
  "shared_agent_specificity_owner": "delegated_owner"
}`), 0o644); err != nil {
		t.Fatalf("write deployment contract: %v", err)
	}
	t.Setenv("MYCELIS_DEPLOYMENT_CONTRACT_PATH", contractPath)

	rr := doAuthenticatedRequest(t, http.HandlerFunc(s.HandleMe), "GET", "/api/v1/user/me", "")
	assertStatus(t, rr, http.StatusOK)

	var user User
	assertJSON(t, rr, &user)

	var settings map[string]any
	if err := json.Unmarshal(user.Settings, &settings); err != nil {
		t.Fatalf("unmarshal settings: %v", err)
	}
	for k, want := range map[string]string{
		"access_management_tier":         "enterprise",
		"product_edition":                "self_hosted_enterprise",
		"identity_mode":                  "hybrid",
		"shared_agent_specificity_owner": "delegated_owner",
	} {
		if settings[k] != want {
			t.Errorf("Expected %s %s, got %#v", k, want, settings[k])
		}
	}
}

func TestHandleMe_Unauthenticated(t *testing.T) {
	s := newTestServer()
	rr := doRequest(t, http.HandlerFunc(s.HandleMe), "GET", "/api/v1/user/me", "")
	assertStatus(t, rr, http.StatusUnauthorized)
}

// ── GET /api/v1/teams ──────────────────────────────────────────────

func TestHandleTeams_NilSoma(t *testing.T) {
	s := newTestServer()
	rr := doRequest(t, http.HandlerFunc(s.HandleTeams), "GET", "/api/v1/teams", "")
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

func TestHandleTeams_POST_NilSoma(t *testing.T) {
	s := newTestServer()
	rr := doRequest(t, http.HandlerFunc(s.HandleTeams), "POST", "/api/v1/teams", `{"name":"test"}`)
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

// ── GET/PUT /api/v1/user/settings ──────────────────────────────────

func TestHandleUserSettings_GET(t *testing.T) {
	s := newTestServer()
	t.Setenv("MYCELIS_USER_SETTINGS_PATH", t.TempDir()+"/user-settings.json")
	rr := doRequest(t, http.HandlerFunc(s.HandleUserSettings), "GET", "/api/v1/user/settings", "")
	assertStatus(t, rr, http.StatusOK)

	var settings map[string]any
	assertJSON(t, rr, &settings)
	for k, want := range map[string]string{
		"theme":                          "system",
		"assistant_name":                 "Soma",
		"access_management_tier":         "release",
		"product_edition":                "self_hosted_release",
		"identity_mode":                  "local_only",
		"shared_agent_specificity_owner": "root_admin",
	} {
		if settings[k] != want {
			t.Errorf("Expected %s %s, got %#v", k, want, settings[k])
		}
	}
}

func TestHandleUpdateSettings(t *testing.T) {
	s := newTestServer()
	t.Setenv("MYCELIS_USER_SETTINGS_PATH", t.TempDir()+"/user-settings.json")
	rr := doRequest(t, http.HandlerFunc(s.HandleUserSettings), "PUT", "/api/v1/user/settings", `{"theme":"midnight-cortex"}`)
	assertStatus(t, rr, http.StatusOK)

	var settings map[string]any
	assertJSON(t, rr, &settings)
	if settings["theme"] != "midnight-cortex" {
		t.Errorf("Expected theme midnight-cortex, got %#v", settings["theme"])
	}
}

func TestHandleUpdateSettings_AssistantNamePersists(t *testing.T) {
	s := newTestServer()
	t.Setenv("MYCELIS_USER_SETTINGS_PATH", t.TempDir()+"/user-settings.json")

	rr := doRequest(t, http.HandlerFunc(s.HandleUserSettings), "PUT", "/api/v1/user/settings", `{"assistant_name":"Mycelis Prime"}`)
	assertStatus(t, rr, http.StatusOK)

	me := doAuthenticatedRequest(t, http.HandlerFunc(s.HandleMe), "GET", "/api/v1/user/me", "")
	assertStatus(t, me, http.StatusOK)

	var user User
	assertJSON(t, me, &user)

	var settings map[string]any
	if err := json.Unmarshal(user.Settings, &settings); err != nil {
		t.Fatalf("unmarshal settings: %v", err)
	}
	if settings["assistant_name"] != "Mycelis Prime" {
		t.Errorf("Expected assistant_name Mycelis Prime, got %#v", settings["assistant_name"])
	}
}

func TestHandleUpdateSettings_DoesNotPersistDeploymentContractOwnedFields(t *testing.T) {
	s := newTestServer()
	settingsPath := filepath.Join(t.TempDir(), "user-settings.json")
	t.Setenv("MYCELIS_USER_SETTINGS_PATH", settingsPath)

	body := `{"theme":"midnight-cortex","access_management_tier":"enterprise","product_edition":"hosted_control_plane","identity_mode":"federated","shared_agent_specificity_owner":"delegated_owner"}`
	rr := doRequest(t, http.HandlerFunc(s.HandleUserSettings), "PUT", "/api/v1/user/settings", body)
	assertStatus(t, rr, http.StatusOK)

	var settings map[string]any
	assertJSON(t, rr, &settings)
	if settings["theme"] != "midnight-cortex" || settings["access_management_tier"] != "release" || settings["product_edition"] != "self_hosted_release" || settings["identity_mode"] != "local_only" || settings["shared_agent_specificity_owner"] != "root_admin" {
		t.Fatalf("unexpected settings after PUT: %#v", settings)
	}

	me := doAuthenticatedRequest(t, http.HandlerFunc(s.HandleMe), "GET", "/api/v1/user/me", "")
	assertStatus(t, me, http.StatusOK)

	var user User
	assertJSON(t, me, &user)
	if err := json.Unmarshal(user.Settings, &settings); err != nil {
		t.Fatalf("unmarshal settings: %v", err)
	}
	if settings["theme"] != "midnight-cortex" || settings["access_management_tier"] != "release" || settings["product_edition"] != "self_hosted_release" || settings["identity_mode"] != "local_only" || settings["shared_agent_specificity_owner"] != "root_admin" {
		t.Fatalf("unexpected settings from HandleMe: %#v", settings)
	}

	raw, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("read persisted settings: %v", err)
	}
	persisted := string(raw)
	for _, forbidden := range []string{
		"access_management_tier",
		"product_edition",
		"identity_mode",
		"shared_agent_specificity_owner",
	} {
		if strings.Contains(persisted, forbidden) {
			t.Fatalf("persisted settings unexpectedly contain %s: %s", forbidden, persisted)
		}
	}
}

func TestIdentityAndSettings_RoleEnforcement(t *testing.T) {
	s := newTestServer()
	t.Setenv("MYCELIS_USER_SETTINGS_PATH", filepath.Join(t.TempDir(), "user-settings.json"))

	viewer := &RequestIdentity{
		UserID:        "00000000-0000-0000-0000-000000000003",
		Username:      "viewer",
		Role:          "viewer",
		EffectiveRole: "viewer",
		PrincipalType: "local_viewer",
		AuthSource:    "local_api_key",
		Scopes:        DefaultScopesForRole("viewer"),
	}
	operator := &RequestIdentity{
		UserID:        "00000000-0000-0000-0000-000000000002",
		Username:      "operator",
		Role:          "operator",
		EffectiveRole: "operator",
		PrincipalType: "local_operator",
		AuthSource:    "local_api_key",
		Scopes:        DefaultScopesForRole("operator"),
	}

	assertStatus(t, doAuthenticatedRequestAs(t, http.HandlerFunc(s.HandleMe), "GET", "/api/v1/user/me", "", viewer), http.StatusOK)
	assertStatus(t, doAuthenticatedRequestAs(t, http.HandlerFunc(s.HandleUserSettings), "GET", "/api/v1/user/settings", "", viewer), http.StatusOK)
	assertStatus(t, doAuthenticatedRequestAs(t, http.HandlerFunc(s.HandleUserSettings), "PUT", "/api/v1/user/settings", `{"theme":"midnight-cortex"}`, viewer), http.StatusForbidden)
	assertStatus(t, doAuthenticatedRequestAs(t, http.HandlerFunc(s.HandleTeams), "POST", "/api/v1/teams", `{"name":"ops"}`, viewer), http.StatusForbidden)

	assertStatus(t, doAuthenticatedRequestAs(t, http.HandlerFunc(s.HandleUserSettings), "PUT", "/api/v1/user/settings", `{"theme":"midnight-cortex"}`, operator), http.StatusOK)
	assertStatus(t, doAuthenticatedRequestAs(t, http.HandlerFunc(s.HandleUserSettings), "PUT", "/api/v1/user/settings", `{"identity_mode":"hybrid"}`, operator), http.StatusForbidden)
	assertStatus(t, doAuthenticatedRequest(t, http.HandlerFunc(s.HandleUserSettings), "PUT", "/api/v1/user/settings", `{"identity_mode":"hybrid","theme":"aero-light"}`), http.StatusOK)

	if got := teamRoleForIdentity(viewer); got != "viewer" {
		t.Fatalf("expected viewer team role, got %q", got)
	}
	if got := teamRoleForIdentity(operator); got != "operator" {
		t.Fatalf("expected operator team role, got %q", got)
	}
	if got := teamRoleForIdentity(localAdminIdentityForTest()); got != "owner" {
		t.Fatalf("expected owner team role, got %q", got)
	}
}

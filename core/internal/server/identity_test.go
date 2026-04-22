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
	if settings["assistant_name"] != "Soma" {
		t.Errorf("Expected assistant_name Soma, got %#v", settings["assistant_name"])
	}
	if settings["access_management_tier"] != "release" {
		t.Errorf("Expected access_management_tier release, got %#v", settings["access_management_tier"])
	}
	if settings["product_edition"] != "self_hosted_release" {
		t.Errorf("Expected product_edition self_hosted_release, got %#v", settings["product_edition"])
	}
	if settings["identity_mode"] != "local_only" {
		t.Errorf("Expected identity_mode local_only, got %#v", settings["identity_mode"])
	}
	if settings["shared_agent_specificity_owner"] != "root_admin" {
		t.Errorf("Expected shared_agent_specificity_owner root_admin, got %#v", settings["shared_agent_specificity_owner"])
	}
}

func TestHandleMe_HybridModeReportsBreakGlassPrincipal(t *testing.T) {
	s := newTestServer()
	rr := doAuthenticatedRequestAs(t, http.HandlerFunc(s.HandleMe), "GET", "/api/v1/user/me", "", breakGlassIdentityForTest())
	assertStatus(t, rr, http.StatusOK)

	var user User
	assertJSON(t, rr, &user)
	if user.PrincipalType != "break_glass_admin" {
		t.Errorf("Expected principal_type break_glass_admin, got %q", user.PrincipalType)
	}
	if user.AuthSource != "local_break_glass" {
		t.Errorf("Expected auth_source local_break_glass, got %q", user.AuthSource)
	}
	if !user.BreakGlass {
		t.Error("Expected break_glass true")
	}
	if user.EffectiveRole != "owner" {
		t.Errorf("Expected effective_role owner, got %q", user.EffectiveRole)
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
	if settings["access_management_tier"] != "enterprise" {
		t.Errorf("Expected access_management_tier enterprise, got %#v", settings["access_management_tier"])
	}
	if settings["product_edition"] != "self_hosted_enterprise" {
		t.Errorf("Expected product_edition self_hosted_enterprise, got %#v", settings["product_edition"])
	}
	if settings["identity_mode"] != "hybrid" {
		t.Errorf("Expected identity_mode hybrid, got %#v", settings["identity_mode"])
	}
	if settings["shared_agent_specificity_owner"] != "delegated_owner" {
		t.Errorf("Expected shared_agent_specificity_owner delegated_owner, got %#v", settings["shared_agent_specificity_owner"])
	}
}

func TestHandleMe_Unauthenticated(t *testing.T) {
	s := newTestServer()
	rr := doRequest(t, http.HandlerFunc(s.HandleMe), "GET", "/api/v1/user/me", "")
	assertStatus(t, rr, http.StatusUnauthorized)
}

// ── GET /api/v1/teams ──────────────────────────────────────────────

func TestHandleTeams_NilSoma(t *testing.T) {
	s := newTestServer() // No Soma
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
	if settings["theme"] != "aero-light" {
		t.Errorf("Expected theme aero-light, got %#v", settings["theme"])
	}
	if settings["assistant_name"] != "Soma" {
		t.Errorf("Expected assistant_name Soma, got %#v", settings["assistant_name"])
	}
	if settings["access_management_tier"] != "release" {
		t.Errorf("Expected access_management_tier release, got %#v", settings["access_management_tier"])
	}
	if settings["product_edition"] != "self_hosted_release" {
		t.Errorf("Expected product_edition self_hosted_release, got %#v", settings["product_edition"])
	}
	if settings["identity_mode"] != "local_only" {
		t.Errorf("Expected identity_mode local_only, got %#v", settings["identity_mode"])
	}
	if settings["shared_agent_specificity_owner"] != "root_admin" {
		t.Errorf("Expected shared_agent_specificity_owner root_admin, got %#v", settings["shared_agent_specificity_owner"])
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
	if settings["theme"] != "midnight-cortex" {
		t.Fatalf("Expected theme midnight-cortex, got %#v", settings["theme"])
	}
	if settings["access_management_tier"] != "release" {
		t.Fatalf("Expected access_management_tier release, got %#v", settings["access_management_tier"])
	}
	if settings["product_edition"] != "self_hosted_release" {
		t.Fatalf("Expected product_edition self_hosted_release, got %#v", settings["product_edition"])
	}
	if settings["identity_mode"] != "local_only" {
		t.Fatalf("Expected identity_mode local_only, got %#v", settings["identity_mode"])
	}
	if settings["shared_agent_specificity_owner"] != "root_admin" {
		t.Fatalf("Expected shared_agent_specificity_owner root_admin, got %#v", settings["shared_agent_specificity_owner"])
	}

	me := doAuthenticatedRequest(t, http.HandlerFunc(s.HandleMe), "GET", "/api/v1/user/me", "")
	assertStatus(t, me, http.StatusOK)

	var user User
	assertJSON(t, me, &user)
	if err := json.Unmarshal(user.Settings, &settings); err != nil {
		t.Fatalf("unmarshal settings: %v", err)
	}
	if settings["theme"] != "midnight-cortex" {
		t.Errorf("Expected theme midnight-cortex from HandleMe, got %#v", settings["theme"])
	}
	if settings["access_management_tier"] != "release" {
		t.Errorf("Expected access_management_tier release from HandleMe, got %#v", settings["access_management_tier"])
	}
	if settings["product_edition"] != "self_hosted_release" {
		t.Errorf("Expected product_edition self_hosted_release from HandleMe, got %#v", settings["product_edition"])
	}
	if settings["identity_mode"] != "local_only" {
		t.Errorf("Expected identity_mode local_only from HandleMe, got %#v", settings["identity_mode"])
	}
	if settings["shared_agent_specificity_owner"] != "root_admin" {
		t.Errorf("Expected shared_agent_specificity_owner root_admin from HandleMe, got %#v", settings["shared_agent_specificity_owner"])
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

// ── GET /api/v1/teams/detail ───────────────────────────────────────

func TestHandleTeamsDetail_NilSoma(t *testing.T) {
	s := newTestServer() // No Soma → returns empty array
	rr := doRequest(t, http.HandlerFunc(s.HandleTeamsDetail), "GET", "/api/v1/teams/detail", "")
	assertStatus(t, rr, http.StatusOK)

	var entries []TeamDetailEntry
	assertJSON(t, rr, &entries)
	if len(entries) != 0 {
		t.Errorf("Expected empty array, got %d entries", len(entries))
	}
}

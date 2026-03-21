package server

import (
	"encoding/json"
	"net/http"
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

// ── PUT /api/v1/user/settings ──────────────────────────────────────

func TestHandleUpdateSettings(t *testing.T) {
	s := newTestServer()
	rr := doRequest(t, http.HandlerFunc(s.HandleUpdateSettings), "PUT", "/api/v1/user/settings", `{"theme":"dark"}`)
	assertStatus(t, rr, http.StatusOK)
}

func TestHandleUpdateSettings_AssistantNamePersists(t *testing.T) {
	s := newTestServer()
	t.Setenv("MYCELIS_USER_SETTINGS_PATH", t.TempDir()+"/user-settings.json")

	rr := doRequest(t, http.HandlerFunc(s.HandleUpdateSettings), "PUT", "/api/v1/user/settings", `{"assistant_name":"Mycelis Prime"}`)
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

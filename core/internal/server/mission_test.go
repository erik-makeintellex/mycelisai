package server

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/pkg/protocol"
)

// ── GET /api/v1/missions ───────────────────────────────────────────

func TestHandleListMissions(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "directive", "status", "created_at", "teams", "agents"}).
		AddRow("m-1", "Build a web scraper", "active", now, 2, 5).
		AddRow("m-2", "Monitor weather", "active", now, 1, 3)

	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	rr := doRequest(t, http.HandlerFunc(s.handleListMissions), "GET", "/api/v1/missions", "")
	assertStatus(t, rr, http.StatusOK)

	var missions []map[string]any
	assertJSON(t, rr, &missions)
	if len(missions) != 2 {
		t.Fatalf("Expected 2 missions, got %d", len(missions))
	}
	if missions[0]["intent"] != "Build a web scraper" {
		t.Errorf("Expected intent 'Build a web scraper', got %v", missions[0]["intent"])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet DB expectations: %v", err)
	}
}

func TestHandleListMissions_Empty(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)

	mock.ExpectQuery("SELECT").
		WillReturnRows(sqlmock.NewRows([]string{"id", "directive", "status", "created_at", "teams", "agents"}))

	rr := doRequest(t, http.HandlerFunc(s.handleListMissions), "GET", "/api/v1/missions", "")
	assertStatus(t, rr, http.StatusOK)

	var missions []map[string]any
	assertJSON(t, rr, &missions)
	if len(missions) != 0 {
		t.Errorf("Expected empty array, got %d missions", len(missions))
	}
}

func TestHandleListMissions_NilDB(t *testing.T) {
	s := newTestServer() // No Registry → getDB() returns nil
	rr := doRequest(t, http.HandlerFunc(s.handleListMissions), "GET", "/api/v1/missions", "")
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

// ── GET /api/v1/missions/{id} ──────────────────────────────────────

func TestHandleGetMission(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)

	now := time.Now()
	missionID := "11111111-1111-1111-1111-111111111111"
	teamID := "22222222-2222-2222-2222-222222222222"

	// 1. Mission query
	mock.ExpectQuery("SELECT .+ FROM missions WHERE").
		WithArgs(missionID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "directive", "status", "created_at"}).
			AddRow(missionID, "Build a scraper", "active", now))

	// 2. Teams query
	mock.ExpectQuery("SELECT .+ FROM teams WHERE").
		WithArgs(missionID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "role"}).
			AddRow(teamID, "scraper-team", "action"))

	// 3. Agents query
	mock.ExpectQuery("SELECT .+ FROM service_manifests WHERE").
		WithArgs(teamID).
		WillReturnRows(sqlmock.NewRows([]string{"team_id", "manifest"}).
			AddRow(teamID, []byte(`{"id":"web-scraper","role":"cognitive","system_prompt":"Scrape the web"}`)))

	mux := setupMux(t, "GET /api/v1/missions/{id}", s.handleGetMission)
	rr := doRequest(t, mux, "GET", "/api/v1/missions/"+missionID, "")
	assertStatus(t, rr, http.StatusOK)

	var result map[string]any
	assertJSON(t, rr, &result)
	if result["id"] != missionID {
		t.Errorf("Expected ID %q, got %v", missionID, result["id"])
	}
	teams, ok := result["teams"].([]any)
	if !ok || len(teams) != 1 {
		t.Fatalf("Expected 1 team, got %v", result["teams"])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet DB expectations: %v", err)
	}
}

func TestHandleGetMission_NotFound(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)

	mock.ExpectQuery("SELECT .+ FROM missions WHERE").
		WithArgs("nonexistent").
		WillReturnRows(sqlmock.NewRows([]string{"id", "directive", "status", "created_at"}))

	mux := setupMux(t, "GET /api/v1/missions/{id}", s.handleGetMission)
	rr := doRequest(t, mux, "GET", "/api/v1/missions/nonexistent", "")
	assertStatus(t, rr, http.StatusNotFound)
}

func TestHandleGetMission_NilDB(t *testing.T) {
	s := newTestServer()
	mux := setupMux(t, "GET /api/v1/missions/{id}", s.handleGetMission)
	rr := doRequest(t, mux, "GET", "/api/v1/missions/some-id", "")
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

// ── PUT /api/v1/missions/{id}/agents/{name} ────────────────────────

func TestHandleUpdateMissionAgent(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)

	mock.ExpectExec("UPDATE service_manifests").
		WithArgs(sqlmock.AnyArg(), "web-scraper", "m-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	mux := setupMux(t, "PUT /api/v1/missions/{id}/agents/{name}", s.handleUpdateMissionAgent)
	body := `{"id":"web-scraper","role":"cognitive","system_prompt":"Updated prompt"}`
	rr := doRequest(t, mux, "PUT", "/api/v1/missions/m-1/agents/web-scraper", body)
	assertStatus(t, rr, http.StatusOK)

	var result map[string]string
	assertJSON(t, rr, &result)
	if result["status"] != "updated" {
		t.Errorf("Expected status 'updated', got %q", result["status"])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet DB expectations: %v", err)
	}
}

func TestHandleUpdateMissionAgent_NotFound(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)

	mock.ExpectExec("UPDATE service_manifests").
		WithArgs(sqlmock.AnyArg(), "ghost", "m-1").
		WillReturnResult(sqlmock.NewResult(0, 0))

	mux := setupMux(t, "PUT /api/v1/missions/{id}/agents/{name}", s.handleUpdateMissionAgent)
	body := `{"id":"ghost","role":"cognitive"}`
	rr := doRequest(t, mux, "PUT", "/api/v1/missions/m-1/agents/ghost", body)
	assertStatus(t, rr, http.StatusNotFound)
}

func TestHandleUpdateMissionAgent_InvalidJSON(t *testing.T) {
	dbOpt, _ := withDB(t)
	s := newTestServer(dbOpt)

	mux := setupMux(t, "PUT /api/v1/missions/{id}/agents/{name}", s.handleUpdateMissionAgent)
	rr := doRequest(t, mux, "PUT", "/api/v1/missions/m-1/agents/x", "not-json")
	assertStatus(t, rr, http.StatusBadRequest)
}

// ── DELETE /api/v1/missions/{id}/agents/{name} ─────────────────────

func TestHandleDeleteMissionAgent(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)

	mock.ExpectExec("DELETE FROM service_manifests").
		WithArgs("web-scraper", "m-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	mux := setupMux(t, "DELETE /api/v1/missions/{id}/agents/{name}", s.handleDeleteMissionAgent)
	rr := doRequest(t, mux, "DELETE", "/api/v1/missions/m-1/agents/web-scraper", "")
	assertStatus(t, rr, http.StatusOK)

	var result map[string]string
	assertJSON(t, rr, &result)
	if result["status"] != "deleted" {
		t.Errorf("Expected status 'deleted', got %q", result["status"])
	}
}

func TestHandleDeleteMissionAgent_NotFound(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)

	mock.ExpectExec("DELETE FROM service_manifests").
		WithArgs("ghost", "m-1").
		WillReturnResult(sqlmock.NewResult(0, 0))

	mux := setupMux(t, "DELETE /api/v1/missions/{id}/agents/{name}", s.handleDeleteMissionAgent)
	rr := doRequest(t, mux, "DELETE", "/api/v1/missions/m-1/agents/ghost", "")
	assertStatus(t, rr, http.StatusNotFound)
}

// ── DELETE /api/v1/missions/{id} ───────────────────────────────────

func TestHandleDeleteMission(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)

	mock.ExpectExec("DELETE FROM missions").
		WithArgs("m-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	mux := setupMux(t, "DELETE /api/v1/missions/{id}", s.handleDeleteMission)
	rr := doRequest(t, mux, "DELETE", "/api/v1/missions/m-1", "")
	assertStatus(t, rr, http.StatusOK)

	var result map[string]any
	assertJSON(t, rr, &result)
	if result["status"] != "deleted" {
		t.Errorf("Expected status 'deleted', got %v", result["status"])
	}
}

func TestHandleDeleteMission_NotFound(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)

	mock.ExpectExec("DELETE FROM missions").
		WithArgs("nonexistent").
		WillReturnResult(sqlmock.NewResult(0, 0))

	mux := setupMux(t, "DELETE /api/v1/missions/{id}", s.handleDeleteMission)
	rr := doRequest(t, mux, "DELETE", "/api/v1/missions/nonexistent", "")
	assertStatus(t, rr, http.StatusNotFound)
}

func TestHandleDeleteMission_NilDB(t *testing.T) {
	s := newTestServer()
	mux := setupMux(t, "DELETE /api/v1/missions/{id}", s.handleDeleteMission)
	rr := doRequest(t, mux, "DELETE", "/api/v1/missions/m-1", "")
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

// ── POST /api/v1/intent/commit ─────────────────────────────────────
// CE-1: handleIntentCommit now requires a confirm_token.

func TestHandleIntentCommit_MissingToken(t *testing.T) {
	// CE-1 directive test: commit without confirm_token → 403
	dbOpt, _ := withDB(t)
	s := newTestServer(dbOpt)

	body := `{
		"intent": "Build a scraper",
		"teams": [{"name": "t", "role": "r", "agents": []}]
	}`
	rr := doRequest(t, http.HandlerFunc(s.handleIntentCommit), "POST", "/api/v1/intent/commit", body)
	assertStatus(t, rr, http.StatusForbidden)
}

func TestHandleIntentCommit_InvalidToken(t *testing.T) {
	// CE-1 directive test: commit with invalid token format → 403
	dbOpt, _ := withDB(t)
	s := newTestServer(dbOpt)

	body := `{
		"intent": "Build a scraper",
		"confirm_token": "not-a-uuid",
		"teams": [{"name": "t", "role": "r", "agents": []}]
	}`
	rr := doRequest(t, http.HandlerFunc(s.handleIntentCommit), "POST", "/api/v1/intent/commit", body)
	assertStatus(t, rr, http.StatusForbidden)
}

func TestHandleIntentCommit_TokenNotFound(t *testing.T) {
	// CE-1 directive test: commit with nonexistent valid-format token → 403
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)

	mock.ExpectQuery("SELECT .+ FROM confirm_tokens").
		WillReturnRows(sqlmock.NewRows([]string{"intent_proof_id", "consumed", "expires_at"}))

	body := `{
		"intent": "Build a scraper",
		"confirm_token": "11111111-1111-1111-1111-111111111111",
		"teams": [{"name": "t", "role": "r", "agents": []}]
	}`
	rr := doRequest(t, http.HandlerFunc(s.handleIntentCommit), "POST", "/api/v1/intent/commit", body)
	assertStatus(t, rr, http.StatusForbidden)
}

func TestHandleIntentCommit_MissingIntent(t *testing.T) {
	dbOpt, _ := withDB(t)
	s := newTestServer(dbOpt)

	body := `{"teams":[{"name":"t","role":"r","agents":[]}]}`
	rr := doRequest(t, http.HandlerFunc(s.handleIntentCommit), "POST", "/api/v1/intent/commit", body)
	assertStatus(t, rr, http.StatusBadRequest)
}

func TestHandleIntentCommit_InvalidJSON(t *testing.T) {
	dbOpt, _ := withDB(t)
	s := newTestServer(dbOpt)

	rr := doRequest(t, http.HandlerFunc(s.handleIntentCommit), "POST", "/api/v1/intent/commit", "not-json")
	assertStatus(t, rr, http.StatusBadRequest)
}

// ── POST /api/v1/intent/negotiate ──────────────────────────────────

func TestHandleIntentNegotiate_NilArchitect(t *testing.T) {
	s := newTestServer() // No MetaArchitect
	rr := doRequest(t, http.HandlerFunc(s.handleIntentNegotiate), "POST", "/api/v1/intent/negotiate", `{"intent":"Build something"}`)
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

func TestHandleIntentNegotiate_MissingIntent(t *testing.T) {
	// Need a non-nil MetaArchitect to pass the nil check, but we can't easily
	// construct one without a real cognitive router. Instead, test that an empty
	// intent returns 400 by simulating via the handler directly.
	// Since MetaArchitect is nil, the 503 check fires first — so this test
	// verifies that ordering: nil MetaArchitect takes precedence.
	s := newTestServer()
	rr := doRequest(t, http.HandlerFunc(s.handleIntentNegotiate), "POST", "/api/v1/intent/negotiate", `{"intent":""}`)
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

// ── buildSensorConfigs (unit test) ─────────────────────────────────

func TestBuildSensorConfigs(t *testing.T) {
	bp := &protocol.MissionBlueprint{
		Teams: []protocol.BlueprintTeam{
			{
				Name: "sensors",
				Agents: []protocol.AgentManifest{
					{ID: "weather-sensor", Role: "sensory"},
					{ID: "gmail-sensor", Role: "Sensor Agent"},
					{ID: "worker", Role: "cognitive"},
				},
			},
		},
	}

	configs := buildSensorConfigs(bp)
	if len(configs) != 2 {
		t.Fatalf("Expected 2 sensor configs, got %d", len(configs))
	}
	if _, ok := configs["weather-sensor"]; !ok {
		t.Error("Expected 'weather-sensor' in configs")
	}
	if _, ok := configs["gmail-sensor"]; !ok {
		t.Error("Expected 'gmail-sensor' in configs")
	}
	if _, ok := configs["worker"]; ok {
		t.Error("'worker' should not be in sensor configs")
	}
}

func TestBuildSensorConfigs_Empty(t *testing.T) {
	bp := &protocol.MissionBlueprint{
		Teams: []protocol.BlueprintTeam{
			{
				Name:   "no-sensors",
				Agents: []protocol.AgentManifest{{ID: "worker", Role: "cognitive"}},
			},
		},
	}

	configs := buildSensorConfigs(bp)
	if len(configs) != 0 {
		t.Errorf("Expected 0 sensor configs, got %d", len(configs))
	}
}

// ── extractBlueprintFromResponse (unit test) ───────────────────────

func TestExtractBlueprintFromResponse_PlainJSON(t *testing.T) {
	json := `{"mission_id":"m-1","intent":"test","teams":[{"name":"t1","agents":[]}]}`
	bp, err := extractBlueprintFromResponse(json)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if bp.MissionID != "m-1" {
		t.Errorf("Expected mission_id 'm-1', got %q", bp.MissionID)
	}
}

func TestExtractBlueprintFromResponse_MarkdownFenced(t *testing.T) {
	response := "Here is the blueprint:\n```json\n" +
		`{"mission_id":"m-2","intent":"test","teams":[{"name":"t1","agents":[]}]}` +
		"\n```\nDone."
	bp, err := extractBlueprintFromResponse(response)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if bp.MissionID != "m-2" {
		t.Errorf("Expected mission_id 'm-2', got %q", bp.MissionID)
	}
}

func TestExtractBlueprintFromResponse_NoJSON(t *testing.T) {
	_, err := extractBlueprintFromResponse("no json here")
	if err == nil {
		t.Error("Expected error for response with no JSON")
	}
	if !strings.Contains(err.Error(), "no JSON object found") {
		t.Errorf("Expected 'no JSON object found' error, got: %v", err)
	}
}

func TestExtractBlueprintFromResponse_InvalidBlueprint(t *testing.T) {
	// Valid JSON but missing required fields
	_, err := extractBlueprintFromResponse(`{"foo":"bar"}`)
	if err == nil {
		t.Error("Expected error for invalid blueprint")
	}
}

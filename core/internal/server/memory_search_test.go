package server

import (
	"net/http"
	"testing"
)

// ── GET /api/v1/memory/search ──────────────────────────────────────

func TestHandleMemorySearch_MissingQuery(t *testing.T) {
	s := newTestServer() // Cognitive and Mem nil — but missing query fires first
	rr := doRequest(t, http.HandlerFunc(s.HandleMemorySearch), "GET", "/api/v1/memory/search", "")
	// Method check fires first (GET is correct), then nil Cognitive → 503
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

func TestHandleMemorySearch_NilCognitive(t *testing.T) {
	s := newTestServer()
	rr := doRequest(t, http.HandlerFunc(s.HandleMemorySearch), "GET", "/api/v1/memory/search?q=hello", "")
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

func TestHandleMemorySearch_MethodNotAllowed(t *testing.T) {
	s := newTestServer()
	rr := doRequest(t, http.HandlerFunc(s.HandleMemorySearch), "POST", "/api/v1/memory/search", "")
	assertStatus(t, rr, http.StatusMethodNotAllowed)
}

// ── GET /api/v1/memory/sitreps ─────────────────────────────────────

func TestHandleListSitReps_NilMem(t *testing.T) {
	s := newTestServer()
	rr := doRequest(t, http.HandlerFunc(s.HandleListSitReps), "GET", "/api/v1/memory/sitreps", "")
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

func TestHandleListSitReps_MethodNotAllowed(t *testing.T) {
	s := newTestServer()
	rr := doRequest(t, http.HandlerFunc(s.HandleListSitReps), "POST", "/api/v1/memory/sitreps", "")
	assertStatus(t, rr, http.StatusMethodNotAllowed)
}

// ── GET /api/v1/sensors ────────────────────────────────────────────

func TestHandleSensors_BaseSensors(t *testing.T) {
	s := newTestServer() // Mem nil → returns base sensors only
	rr := doRequest(t, http.HandlerFunc(s.HandleSensors), "GET", "/api/v1/sensors", "")
	assertStatus(t, rr, http.StatusOK)

	var result map[string]any
	assertJSON(t, rr, &result)
	count := result["count"].(float64)
	if count != 7 {
		t.Errorf("Expected 7 base sensors, got %v", count)
	}
}

func TestHandleSensors_MethodNotAllowed(t *testing.T) {
	s := newTestServer()
	rr := doRequest(t, http.HandlerFunc(s.HandleSensors), "POST", "/api/v1/sensors", "")
	assertStatus(t, rr, http.StatusMethodNotAllowed)
}

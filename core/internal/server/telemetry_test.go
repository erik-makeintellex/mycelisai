package server

import (
	"net/http"
	"testing"
)

// ── GET /api/v1/telemetry/compute ──────────────────────────────────

func TestHandleTelemetry(t *testing.T) {
	s := newTestServer()
	rr := doRequest(t, http.HandlerFunc(s.HandleTelemetry), "GET", "/api/v1/telemetry/compute", "")
	assertStatus(t, rr, http.StatusOK)

	var snap TelemetrySnapshot
	assertJSON(t, rr, &snap)
	if snap.Goroutines <= 0 {
		t.Errorf("Expected goroutines > 0, got %d", snap.Goroutines)
	}
	if snap.Timestamp == "" {
		t.Error("Expected non-empty timestamp")
	}
}

func TestHandleTelemetry_MethodNotAllowed(t *testing.T) {
	s := newTestServer()
	rr := doRequest(t, http.HandlerFunc(s.HandleTelemetry), "POST", "/api/v1/telemetry/compute", "")
	assertStatus(t, rr, http.StatusMethodNotAllowed)
}

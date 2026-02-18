package server

import (
	"net/http"
	"testing"

	"github.com/mycelis/core/internal/overseer"
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

// ── GET/PUT /api/v1/trust/threshold ────────────────────────────────

func TestHandleTrustThreshold_GET(t *testing.T) {
	ov := overseer.NewEngine(nil) // nil NATS is fine for threshold ops
	s := newTestServer(func(s *AdminServer) { s.Overseer = ov })

	rr := doRequest(t, http.HandlerFunc(s.HandleTrustThreshold), "GET", "/api/v1/trust/threshold", "")
	assertStatus(t, rr, http.StatusOK)

	var result map[string]float64
	assertJSON(t, rr, &result)
	if result["threshold"] != 0.7 {
		t.Errorf("Expected default threshold 0.7, got %f", result["threshold"])
	}
}

func TestHandleTrustThreshold_PUT(t *testing.T) {
	ov := overseer.NewEngine(nil)
	s := newTestServer(func(s *AdminServer) { s.Overseer = ov })

	rr := doRequest(t, http.HandlerFunc(s.HandleTrustThreshold), "PUT", "/api/v1/trust/threshold", `{"threshold":0.5}`)
	assertStatus(t, rr, http.StatusOK)

	// Verify the threshold was actually updated
	if ov.GetAutoExecuteThreshold() != 0.5 {
		t.Errorf("Expected threshold 0.5, got %f", ov.GetAutoExecuteThreshold())
	}
}

func TestHandleTrustThreshold_OutOfRange(t *testing.T) {
	ov := overseer.NewEngine(nil)
	s := newTestServer(func(s *AdminServer) { s.Overseer = ov })

	// Above 1.0
	rr := doRequest(t, http.HandlerFunc(s.HandleTrustThreshold), "PUT", "/api/v1/trust/threshold", `{"threshold":1.5}`)
	assertStatus(t, rr, http.StatusBadRequest)

	// Below 0.0
	rr = doRequest(t, http.HandlerFunc(s.HandleTrustThreshold), "PUT", "/api/v1/trust/threshold", `{"threshold":-0.1}`)
	assertStatus(t, rr, http.StatusBadRequest)
}

func TestHandleTrustThreshold_NilOverseer(t *testing.T) {
	s := newTestServer()
	rr := doRequest(t, http.HandlerFunc(s.HandleTrustThreshold), "GET", "/api/v1/trust/threshold", "")
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

func TestHandleTrustThreshold_InvalidJSON(t *testing.T) {
	ov := overseer.NewEngine(nil)
	s := newTestServer(func(s *AdminServer) { s.Overseer = ov })

	rr := doRequest(t, http.HandlerFunc(s.HandleTrustThreshold), "PUT", "/api/v1/trust/threshold", "not-json")
	assertStatus(t, rr, http.StatusBadRequest)
}

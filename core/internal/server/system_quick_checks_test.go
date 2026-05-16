package server

import (
	"net/http"
	"testing"
)

func TestHandleSystemQuickCheck_SchedulerHealthy(t *testing.T) {
	s := newTestServer()
	s.LoopScheduler = NewLoopScheduler(s)
	s.LoopScheduler.Start(t.Context())
	defer s.LoopScheduler.Stop()

	mux := setupMux(t, "GET /api/v1/system/quick-checks/{id}", s.HandleSystemQuickCheck)
	rr := doRequest(t, mux, "GET", "/api/v1/system/quick-checks/scheduler", "")

	assertStatus(t, rr, http.StatusOK)
	var resp map[string]any
	assertJSON(t, rr, &resp)
	data := resp["data"].(map[string]any)
	if data["status"] != "healthy" {
		t.Fatalf("expected scheduler healthy, got %v", data["status"])
	}
}

func TestHandleSystemQuickCheck_Unknown(t *testing.T) {
	s := newTestServer()

	mux := setupMux(t, "GET /api/v1/system/quick-checks/{id}", s.HandleSystemQuickCheck)
	rr := doRequest(t, mux, "GET", "/api/v1/system/quick-checks/not-real", "")

	assertStatus(t, rr, http.StatusNotFound)
}

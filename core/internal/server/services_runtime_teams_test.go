package server

import (
	"net/http"
	"testing"

	"github.com/mycelis/core/internal/swarm"
)

func TestHandleServicesStatus_RuntimeTeamsOnlineWhenNothingDegraded(t *testing.T) {
	s := newTestServer()
	s.Soma = swarm.NewSoma(s.NC, nil, nil, nil, nil, nil, nil)

	mux := setupMux(t, "GET /api/v1/services/status", s.HandleServicesStatus)
	rr := doRequest(t, mux, "GET", "/api/v1/services/status", "")
	assertStatus(t, rr, http.StatusOK)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	data, ok := resp["data"].([]any)
	if !ok {
		t.Fatalf("expected data array, got %T", resp["data"])
	}
	for _, item := range data {
		svc := item.(map[string]any)
		if svc["name"] == "runtime_teams" {
			if svc["status"] != "online" {
				t.Fatalf("runtime_teams status = %v, want online", svc["status"])
			}
			return
		}
	}
	t.Fatal("runtime_teams service missing from status payload")
}

func TestHandleServicesStatus_NoRuntimeTeamsEntryWithoutSoma(t *testing.T) {
	s := newTestServer()

	mux := setupMux(t, "GET /api/v1/services/status", s.HandleServicesStatus)
	rr := doRequest(t, mux, "GET", "/api/v1/services/status", "")

	var resp map[string]any
	assertJSON(t, rr, &resp)
	for _, item := range resp["data"].([]any) {
		if item.(map[string]any)["name"] == "runtime_teams" {
			t.Fatal("runtime_teams must not be reported when Soma is not wired")
		}
	}
}

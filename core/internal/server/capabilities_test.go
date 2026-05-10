package server

import (
	"net/http"
	"testing"
	"time"

	"github.com/mycelis/core/internal/capabilities"
	"github.com/mycelis/core/internal/exchange"
)

func withCapabilityService() func(*AdminServer) {
	return func(s *AdminServer) {
		s.Capabilities = capabilities.NewService(capabilities.Dependencies{
			ExchangeCapabilities: []exchange.CapabilityDefinition{{
				ID:        "planning",
				Label:     "Planning",
				Source:    "internal_tool",
				RiskClass: "low-risk",
			}},
			HostCommands: func() []string { return nil },
			Now:          func() time.Time { return time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC) },
		})
	}
}

func TestHandleListCapabilities(t *testing.T) {
	s := newTestServer(withCapabilityService())
	rr := doRequest(t, http.HandlerFunc(s.HandleListCapabilities), "GET", "/api/v1/capabilities", "")
	assertStatus(t, rr, http.StatusOK)

	var resp struct {
		OK   bool `json:"ok"`
		Data struct {
			Count     int `json:"count"`
			Manifests []struct {
				ID   string `json:"id"`
				Kind string `json:"kind"`
			} `json:"manifests"`
		} `json:"data"`
	}
	assertJSON(t, rr, &resp)
	if !resp.OK {
		t.Fatal("response ok = false")
	}
	if resp.Data.Count != 2 {
		t.Fatalf("count = %d, want 2", resp.Data.Count)
	}
	foundPlanning := false
	for _, manifest := range resp.Data.Manifests {
		if manifest.ID == "planning" {
			foundPlanning = true
		}
	}
	if !foundPlanning {
		t.Fatalf("planning manifest not found: %#v", resp.Data.Manifests)
	}
}

func TestHandleGetCapability(t *testing.T) {
	s := newTestServer(withCapabilityService())
	mux := setupMux(t, "GET /api/v1/capabilities/{id}", s.HandleGetCapability)
	rr := doRequest(t, mux, "GET", "/api/v1/capabilities/planning", "")
	assertStatus(t, rr, http.StatusOK)

	var resp struct {
		OK   bool `json:"ok"`
		Data struct {
			ID   string `json:"id"`
			Kind string `json:"kind"`
		} `json:"data"`
	}
	assertJSON(t, rr, &resp)
	if resp.Data.ID != "planning" || resp.Data.Kind != "exchange_capability" {
		t.Fatalf("unexpected manifest: %#v", resp.Data)
	}
}

func TestHandleGetCapabilityNotFound(t *testing.T) {
	s := newTestServer(withCapabilityService())
	mux := setupMux(t, "GET /api/v1/capabilities/{id}", s.HandleGetCapability)
	rr := doRequest(t, mux, "GET", "/api/v1/capabilities/missing", "")
	assertStatus(t, rr, http.StatusNotFound)
}

func TestHandleRefreshCapabilities(t *testing.T) {
	s := newTestServer(withCapabilityService())
	rr := doRequest(t, http.HandlerFunc(s.HandleRefreshCapabilities), "POST", "/api/v1/capabilities/refresh", "")
	assertStatus(t, rr, http.StatusOK)
}

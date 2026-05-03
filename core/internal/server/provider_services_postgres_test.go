package server

import (
	"fmt"
	"net/http"
	"testing"
)

func TestHandleServicesStatus_PostgresOnline(t *testing.T) {
	dbOpt, mock := withDirectDBPing(t)
	s := newTestServer(dbOpt)

	mock.ExpectPing()

	mux := setupMux(t, "GET /api/v1/services/status", s.HandleServicesStatus)
	rr := doRequest(t, mux, "GET", "/api/v1/services/status", "")

	assertStatus(t, rr, http.StatusOK)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	data := resp["data"].([]any)

	for _, item := range data {
		svc := item.(map[string]any)
		if svc["name"] == "postgres" {
			if svc["status"] != "online" {
				t.Errorf("expected postgres=online, got %v", svc["status"])
			}
			return
		}
	}
	t.Error("postgres service entry not found")
}

func TestHandleServicesStatus_PostgresPingFails(t *testing.T) {
	dbOpt, mock := withDirectDBPing(t)
	s := newTestServer(dbOpt)

	mock.ExpectPing().WillReturnError(fmt.Errorf("connection refused"))

	mux := setupMux(t, "GET /api/v1/services/status", s.HandleServicesStatus)
	rr := doRequest(t, mux, "GET", "/api/v1/services/status", "")

	assertStatus(t, rr, http.StatusOK)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	data := resp["data"].([]any)

	for _, item := range data {
		svc := item.(map[string]any)
		if svc["name"] == "postgres" {
			if svc["status"] != "offline" {
				t.Errorf("expected postgres=offline on ping failure, got %v", svc["status"])
			}
			return
		}
	}
	t.Error("postgres service entry not found")
}

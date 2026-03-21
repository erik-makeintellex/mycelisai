package server

import (
	"net/http"
	"testing"
)

func TestHandleHostStatus_AuthRequired(t *testing.T) {
	s := newTestServer()
	rr := doRequest(t, http.HandlerFunc(s.HandleHostStatus), "GET", "/api/v1/host/status", "")
	assertStatus(t, rr, 401)
}

func TestHandleHostStatus_HappyPath(t *testing.T) {
	s := newTestServer()
	mux := setupMux(t, "GET /api/v1/host/status", s.HandleHostStatus)
	rr := doAuthenticatedRequest(t, mux, "GET", "/api/v1/host/status", "")
	assertStatus(t, rr, 200)

	var resp struct {
		OK   bool `json:"ok"`
		Data struct {
			LocalCommandV0 bool     `json:"local_command_v0"`
			Allowed        []string `json:"allowed_commands"`
		} `json:"data"`
	}
	assertJSON(t, rr, &resp)
	if !resp.OK {
		t.Fatal("expected ok=true")
	}
	if !resp.Data.LocalCommandV0 {
		t.Fatal("expected local_command_v0=true")
	}
	if len(resp.Data.Allowed) == 0 {
		t.Fatal("expected allowed commands list")
	}
}

func TestHandleHostActions_HappyPath(t *testing.T) {
	s := newTestServer()
	mux := setupMux(t, "GET /api/v1/host/actions", s.HandleHostActions)
	rr := doAuthenticatedRequest(t, mux, "GET", "/api/v1/host/actions", "")
	assertStatus(t, rr, 200)

	var resp struct {
		OK   bool                     `json:"ok"`
		Data []map[string]interface{} `json:"data"`
	}
	assertJSON(t, rr, &resp)
	if !resp.OK || len(resp.Data) == 0 {
		t.Fatal("expected at least one host action")
	}
	if resp.Data[0]["action_id"] != "local-command" {
		t.Fatalf("action_id = %v, want local-command", resp.Data[0]["action_id"])
	}
}

func TestHandleInvokeHostAction_UnknownAction(t *testing.T) {
	s := newTestServer()
	mux := setupMux(t, "POST /api/v1/host/actions/{id}/invoke", s.HandleInvokeHostAction)
	rr := doAuthenticatedRequest(t, mux, "POST", "/api/v1/host/actions/unknown/invoke", `{"command":"hostname"}`)
	assertStatus(t, rr, 404)
}

func TestHandleInvokeHostAction_Disallowed(t *testing.T) {
	t.Setenv("MYCELIS_LOCAL_COMMAND_ALLOWLIST", "hostname")
	s := newTestServer()
	mux := setupMux(t, "POST /api/v1/host/actions/{id}/invoke", s.HandleInvokeHostAction)
	rr := doAuthenticatedRequest(t, mux, "POST", "/api/v1/host/actions/local-command/invoke", `{"command":"whoami"}`)
	assertStatus(t, rr, 403)
}

func TestHandleInvokeHostAction_HappyPath(t *testing.T) {
	t.Setenv("MYCELIS_LOCAL_COMMAND_ALLOWLIST", "hostname")
	s := newTestServer()
	mux := setupMux(t, "POST /api/v1/host/actions/{id}/invoke", s.HandleInvokeHostAction)
	rr := doAuthenticatedRequest(t, mux, "POST", "/api/v1/host/actions/local-command/invoke", `{"command":"hostname"}`)
	assertStatus(t, rr, 200)

	var resp struct {
		OK   bool `json:"ok"`
		Data struct {
			ActionID string `json:"action_id"`
			Result   struct {
				Status string `json:"status"`
			} `json:"result"`
		} `json:"data"`
	}
	assertJSON(t, rr, &resp)
	if !resp.OK {
		t.Fatal("expected ok=true")
	}
	if resp.Data.ActionID != "local-command" {
		t.Fatalf("action_id = %q, want local-command", resp.Data.ActionID)
	}
	if resp.Data.Result.Status != "success" {
		t.Fatalf("result.status = %q, want success", resp.Data.Result.Status)
	}
}

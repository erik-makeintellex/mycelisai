package server

import (
	"net/http"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

func TestRespondAPIJSON_Success(t *testing.T) {
	s := newTestServer(withCouncilSoma())
	rr := doRequest(t, http.HandlerFunc(s.HandleListCouncilMembers), "GET", "/api/v1/council/members", "")

	assertJSONContentType(t, rr.Header().Get("Content-Type"))
	var resp protocol.APIResponse
	assertJSON(t, rr, &resp)
	if !resp.OK {
		t.Error("Expected ok=true in success response")
	}
}

func TestRespondAPIJSON_Error(t *testing.T) {
	s := newTestServer()
	rr := doRequest(t, http.HandlerFunc(s.HandleListCouncilMembers), "GET", "/api/v1/council/members", "")

	assertJSONContentType(t, rr.Header().Get("Content-Type"))
	var resp protocol.APIResponse
	assertJSON(t, rr, &resp)
	if resp.OK {
		t.Error("Expected ok=false in error response")
	}
	if resp.Error == "" {
		t.Error("Expected non-empty error in error response")
	}
}

func assertJSONContentType(t *testing.T, contentType string) {
	t.Helper()
	if contentType != "application/json" {
		t.Errorf("Content-Type: got %q, want %q", contentType, "application/json")
	}
}

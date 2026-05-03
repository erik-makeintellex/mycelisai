package server

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

func TestHandleCreateOrganization_StartEmpty(t *testing.T) {
	s := newTestServer(withTemplateBundlesPath(writeStarterBundle(t)))

	rr := doRequest(t, http.HandlerFunc(s.handleCreateOrganization), "POST", "/api/v1/organizations", `{"name":"Blank Canvas","purpose":"Shape a new AI Organization","start_mode":"empty"}`)
	assertStatus(t, rr, http.StatusCreated)

	var resp protocol.APIResponse
	assertJSON(t, rr, &resp)
	data, ok := resp.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected object data, got %T", resp.Data)
	}
	if data["department_count"] != float64(0) || data["specialist_count"] != float64(0) {
		t.Fatalf("unexpected empty-start counts: %+v", data)
	}
	if data["ai_engine_settings_summary"] != "Set up later in Advanced mode" {
		t.Fatalf("unexpected empty-start AI Engine summary: %+v", data)
	}
}

func TestHandleCreateOrganization_TriggersEventDrivenReviews(t *testing.T) {
	s := newTestServer(withTemplateBundlesPath(writeStarterBundle(t)))

	rr := doRequest(t, http.HandlerFunc(s.handleCreateOrganization), "POST", "/api/v1/organizations", `{"name":"Northstar Labs","purpose":"Ship a focused AI engineering organization","start_mode":"template","template_id":"engineering-starter"}`)
	assertStatus(t, rr, http.StatusCreated)

	var resp protocol.APIResponse
	assertJSON(t, rr, &resp)
	data, ok := resp.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected object data, got %T", resp.Data)
	}
	id, _ := data["id"].(string)
	if id == "" {
		t.Fatalf("expected organization id, got %+v", data)
	}

	results := s.loopResultStore().List(id)
	if len(results) != 2 {
		t.Fatalf("expected default event-driven review activity after create, got %+v", results)
	}
	for _, result := range results {
		if result.Trigger != "event:organization_created" {
			t.Fatalf("expected organization-created trigger label, got %+v", result)
		}
	}
}

func TestHandleListOrganizations_ReturnsCreatedSummaries(t *testing.T) {
	s := newTestServer(withTemplateBundlesPath(writeStarterBundle(t)))

	createBody := `{"name":"Atlas","purpose":"Resume me later","start_mode":"empty"}`
	createRR := doRequest(t, http.HandlerFunc(s.handleCreateOrganization), "POST", "/api/v1/organizations", createBody)
	assertStatus(t, createRR, http.StatusCreated)

	rr := doRequest(t, http.HandlerFunc(s.handleListOrganizations), "GET", "/api/v1/organizations?view=summary", "")
	assertStatus(t, rr, http.StatusOK)

	var resp protocol.APIResponse
	assertJSON(t, rr, &resp)
	list, ok := resp.Data.([]any)
	if !ok {
		raw, _ := json.Marshal(resp.Data)
		t.Fatalf("expected list data, got %T (%s)", resp.Data, raw)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 organization summary, got %d", len(list))
	}
}

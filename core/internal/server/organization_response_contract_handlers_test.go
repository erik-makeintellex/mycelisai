package server

import (
	"net/http"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

func TestHandleUpdateResponseContract_StoresCuratedProfile(t *testing.T) {
	s := newTestServer(withTemplateBundlesPath(writeStarterBundle(t)))
	created := s.organizationStore().Save(OrganizationHomePayload{
		OrganizationSummary: OrganizationSummary{
			ID:                        "org-123",
			Name:                      "Northstar Labs",
			Purpose:                   "Ship a focused AI engineering organization",
			StartMode:                 OrganizationStartModeTemplate,
			TeamLeadLabel:             "Team Lead",
			DepartmentCount:           1,
			SpecialistCount:           1,
			ResponseContractProfileID: "clear_balanced",
			ResponseContractSummary:   "Clear & Balanced",
			Status:                    "ready",
		},
		Departments: []OrganizationDepartmentSummary{
			{
				ID:              "platform",
				Name:            "Platform Department",
				SpecialistCount: 1,
				AgentTypeProfiles: []OrganizationAgentTypeProfileSummary{
					{ID: "planner", Name: "Planner", HelpsWith: "Keeps priorities clear."},
				},
			},
		},
	})

	mux := http.NewServeMux()
	mux.HandleFunc("PATCH /api/v1/organizations/{id}/response-contract", s.handleUpdateResponseContract)
	mux.HandleFunc("GET /api/v1/organizations/{id}/home", s.handleGetOrganizationHome)

	updateRR := doRequest(t, mux, "PATCH", "/api/v1/organizations/"+created.ID+"/response-contract", `{"profile_id":"warm_supportive"}`)
	assertStatus(t, updateRR, http.StatusOK)

	var updateResp protocol.APIResponse
	assertJSON(t, updateRR, &updateResp)
	data, ok := updateResp.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected object data, got %T", updateResp.Data)
	}
	if data["response_contract_profile_id"] != "warm_supportive" || data["response_contract_summary"] != "Warm & Supportive" {
		t.Fatalf("unexpected updated response contract: %+v", data)
	}

	homeRR := doRequest(t, mux, "GET", "/api/v1/organizations/"+created.ID+"/home", "")
	assertStatus(t, homeRR, http.StatusOK)

	var homeResp protocol.APIResponse
	assertJSON(t, homeRR, &homeResp)
	homeData, ok := homeResp.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected object home data, got %T", homeResp.Data)
	}
	if homeData["response_contract_profile_id"] != "warm_supportive" || homeData["response_contract_summary"] != "Warm & Supportive" {
		t.Fatalf("unexpected persisted response contract: %+v", homeData)
	}
	results := s.loopResultStore().List(created.ID)
	if len(results) != 2 {
		t.Fatalf("expected event-driven review activity after Response Style update, got %+v", results)
	}
	for _, result := range results {
		if result.Trigger != "event:response_contract_changed" {
			t.Fatalf("expected response-style event trigger label, got %+v", results)
		}
	}
}

func TestHandleUpdateResponseContract_RejectsInvalidProfile(t *testing.T) {
	s := newTestServer(withTemplateBundlesPath(writeStarterBundle(t)))
	created := s.organizationStore().Save(OrganizationHomePayload{
		OrganizationSummary: OrganizationSummary{
			ID:                        "org-123",
			Name:                      "Northstar Labs",
			Purpose:                   "Ship a focused AI engineering organization",
			StartMode:                 OrganizationStartModeTemplate,
			TeamLeadLabel:             "Team Lead",
			ResponseContractProfileID: "clear_balanced",
			ResponseContractSummary:   "Clear & Balanced",
			Status:                    "ready",
		},
	})

	mux := http.NewServeMux()
	mux.HandleFunc("PATCH /api/v1/organizations/{id}/response-contract", s.handleUpdateResponseContract)

	rr := doRequest(t, mux, "PATCH", "/api/v1/organizations/"+created.ID+"/response-contract", `{"profile_id":"raw_prompt_override"}`)
	assertStatus(t, rr, http.StatusBadRequest)
}

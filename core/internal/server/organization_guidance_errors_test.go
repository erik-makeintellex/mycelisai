package server

import (
	"net/http"
	"testing"
)

func TestNormalizeOrganizationHome_AgentTypeProfilesResolveInheritanceAndTypeBindings(t *testing.T) {
	home := normalizeOrganizationHome(OrganizationHomePayload{
		OrganizationSummary: OrganizationSummary{
			ID:                        "org-123",
			Name:                      "Northstar Labs",
			Purpose:                   "Ship a focused AI engineering organization",
			StartMode:                 OrganizationStartModeTemplate,
			TeamLeadLabel:             "Team Lead",
			DepartmentCount:           1,
			SpecialistCount:           2,
			AIEngineProfileID:         "balanced",
			AIEngineSettingsSummary:   "Balanced",
			ResponseContractProfileID: "warm_supportive",
			ResponseContractSummary:   "Warm & Supportive",
			Status:                    "ready",
		},
		Departments: []OrganizationDepartmentSummary{
			{
				ID:              "platform",
				Name:            "Platform Department",
				SpecialistCount: 2,
				AgentTypeProfiles: []OrganizationAgentTypeProfileSummary{
					{
						ID:                               "planner",
						Name:                             "Planner",
						HelpsWith:                        "Keeps priorities and sequencing clear.",
						AIEngineBindingProfileID:         "high_reasoning",
						ResponseContractBindingProfileID: "structured_analytical",
					},
					{
						ID:        "reviewer",
						Name:      "Reviewer",
						HelpsWith: "Checks quality before work moves forward.",
					},
				},
			},
		},
	})

	if len(home.Departments) != 1 || len(home.Departments[0].AgentTypeProfiles) != 2 {
		t.Fatalf("expected normalized agent type profiles, got %+v", home.Departments)
	}

	planner := home.Departments[0].AgentTypeProfiles[0]
	if planner.InheritsDepartmentAIEngine || planner.AIEngineEffectiveSummary != "High Reasoning" {
		t.Fatalf("expected planner AI Engine binding to stay type-specific, got %+v", planner)
	}
	if planner.InheritsDefaultResponseContract || planner.ResponseContractEffectiveSummary != "Structured & Analytical" {
		t.Fatalf("expected planner Response Style binding to stay type-specific, got %+v", planner)
	}

	reviewer := home.Departments[0].AgentTypeProfiles[1]
	if !reviewer.InheritsDepartmentAIEngine || reviewer.AIEngineEffectiveSummary != "Balanced" {
		t.Fatalf("expected reviewer AI Engine to inherit department default, got %+v", reviewer)
	}
	if !reviewer.InheritsDefaultResponseContract || reviewer.ResponseContractEffectiveSummary != "Warm & Supportive" {
		t.Fatalf("expected reviewer Response Style to inherit organization default, got %+v", reviewer)
	}
}

func TestHandleTeamLeadGuidedAction_RejectsUnknownAction(t *testing.T) {
	s := newTestServer(withTemplateBundlesPath(writeStarterBundle(t)))
	created := s.organizationStore().Save(OrganizationHomePayload{
		OrganizationSummary: OrganizationSummary{
			ID:            "org-123",
			Name:          "Northstar Labs",
			Purpose:       "Ship a focused AI engineering organization",
			StartMode:     OrganizationStartModeEmpty,
			TeamLeadLabel: "Team Lead",
			Status:        "ready",
		},
	})

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/organizations/{id}/workspace/actions", s.handleTeamLeadGuidedAction)

	rr := doRequest(t, mux, "POST", "/api/v1/organizations/"+created.ID+"/workspace/actions", `{"action":"launch_agents"}`)
	assertStatus(t, rr, http.StatusBadRequest)
}

func TestHandleTeamLeadGuidedAction_RejectsMalformedRequest(t *testing.T) {
	s := newTestServer(withTemplateBundlesPath(writeStarterBundle(t)))
	created := s.organizationStore().Save(OrganizationHomePayload{
		OrganizationSummary: OrganizationSummary{
			ID:            "org-123",
			Name:          "Northstar Labs",
			Purpose:       "Ship a focused AI engineering organization",
			StartMode:     OrganizationStartModeEmpty,
			TeamLeadLabel: "Team Lead",
			Status:        "ready",
		},
	})

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/organizations/{id}/workspace/actions", s.handleTeamLeadGuidedAction)

	rr := doRequest(t, mux, "POST", "/api/v1/organizations/"+created.ID+"/workspace/actions", `{"action":`)
	assertStatus(t, rr, http.StatusBadRequest)
}

func TestHandleTeamLeadGuidedAction_ReturnsNotFoundForMissingOrganization(t *testing.T) {
	s := newTestServer(withTemplateBundlesPath(writeStarterBundle(t)))

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/organizations/{id}/workspace/actions", s.handleTeamLeadGuidedAction)

	rr := doRequest(t, mux, "POST", "/api/v1/organizations/org-missing/workspace/actions", `{"action":"plan_next_steps"}`)
	assertStatus(t, rr, http.StatusNotFound)
}

func TestHandleTeamLeadGuidedAction_TriggersEventDrivenReview(t *testing.T) {
	s := newTestServer(withTemplateBundlesPath(writeStarterBundle(t)))
	created := s.organizationStore().Save(OrganizationHomePayload{
		OrganizationSummary: OrganizationSummary{
			ID:                        "org-123",
			Name:                      "Northstar Labs",
			Purpose:                   "Ship a focused AI engineering organization",
			StartMode:                 OrganizationStartModeTemplate,
			TeamLeadLabel:             "Team Lead",
			AdvisorCount:              1,
			DepartmentCount:           1,
			SpecialistCount:           2,
			AIEngineProfileID:         "balanced",
			AIEngineSettingsSummary:   "Balanced",
			ResponseContractProfileID: "clear_balanced",
			ResponseContractSummary:   "Clear & Balanced",
			Status:                    "ready",
		},
		Departments: []OrganizationDepartmentSummary{
			{
				ID:              "platform",
				Name:            "Platform Department",
				SpecialistCount: 2,
				AgentTypeProfiles: []OrganizationAgentTypeProfileSummary{
					{ID: "planner", Name: "Planner", HelpsWith: "Keeps priorities clear."},
				},
			},
		},
	})
	s.loopProfileStore().EnsureDefaults(created)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/organizations/{id}/workspace/actions", s.handleTeamLeadGuidedAction)

	rr := doRequest(t, mux, "POST", "/api/v1/organizations/"+created.ID+"/workspace/actions", `{"action":"plan_next_steps"}`)
	assertStatus(t, rr, http.StatusOK)

	results := s.loopResultStore().List(created.ID)
	if len(results) != 1 || results[0].Trigger != "event:team_lead_action_completed" {
		t.Fatalf("expected event-driven review activity after Team Lead action, got %+v", results)
	}
}

func TestBuildTeamLeadGuidance_UsesReadableFallbacksForPartialHome(t *testing.T) {
	response, err := buildTeamLeadGuidance(OrganizationHomePayload{
		OrganizationSummary: OrganizationSummary{
			StartMode: OrganizationStartModeEmpty,
			Status:    "ready",
		},
	}, TeamLeadGuidedActionPlanNextSteps, "")
	if err != nil {
		t.Fatalf("buildTeamLeadGuidance returned error: %v", err)
	}

	if response.Headline != "Team Lead plan for this AI Organization" {
		t.Fatalf("unexpected fallback headline: %+v", response)
	}
	if response.Summary != "Team Lead recommends moving this AI Organization from setup into a focused first delivery loop." {
		t.Fatalf("unexpected fallback summary: %+v", response)
	}
	if len(response.PrioritySteps) != 3 {
		t.Fatalf("expected fallback priority steps, got %+v", response)
	}
	if response.PrioritySteps[0] != "Align the first outcome with this purpose: the current AI Organization priorities." {
		t.Fatalf("unexpected fallback priority steps: %+v", response.PrioritySteps)
	}
	if len(response.SuggestedFollowUps) != 3 {
		t.Fatalf("expected fallback follow-ups, got %+v", response)
	}
}

func TestBuildTeamLeadGuidance_ResumeRetainedPackageUsesReadableFallbacksForPartialHome(t *testing.T) {
	response, err := buildTeamLeadGuidance(OrganizationHomePayload{
		OrganizationSummary: OrganizationSummary{
			StartMode:     OrganizationStartModeEmpty,
			TeamLeadLabel: "Team Lead",
		},
	}, TeamLeadGuidedActionResumeRetainedPackage, "")
	if err != nil {
		t.Fatalf("buildTeamLeadGuidance returned error: %v", err)
	}

	if response.Headline != "Retained package continuity for this AI Organization" {
		t.Fatalf("unexpected fallback headline: %+v", response)
	}
	if response.Summary != "Team Lead resumes the retained package for this AI Organization so completed work stays durable and the next step stays explicit after a reboot or reload." {
		t.Fatalf("unexpected fallback summary: %+v", response)
	}
	if len(response.PrioritySteps) != 3 {
		t.Fatalf("expected fallback priority steps, got %+v", response)
	}
	if response.PrioritySteps[0] != "Open the retained package and confirm the latest durable outputs." {
		t.Fatalf("unexpected fallback priority steps: %+v", response.PrioritySteps)
	}
	if response.ExecutionContract == nil {
		t.Fatal("expected execution contract")
	}
	if response.ExecutionContract.ExecutionMode != TeamLeadExecutionModeContinuityResume {
		t.Fatalf("expected continuity resume execution mode, got %+v", response.ExecutionContract)
	}
	if response.ExecutionContract.ContinuityLabel != "Retained package continuity" {
		t.Fatalf("expected continuity label, got %+v", response.ExecutionContract)
	}
	if len(response.ExecutionContract.Workstreams) != 3 {
		t.Fatalf("expected continuity workstreams, got %+v", response.ExecutionContract)
	}
	if response.ExecutionContract.Workstreams[2].Label != "Next-step handoff lane" {
		t.Fatalf("unexpected continuity handoff lane: %+v", response.ExecutionContract.Workstreams)
	}
	if response.ExecutionContract.WorkflowGroup == nil {
		t.Fatal("expected workflow group draft")
	}
	if response.ExecutionContract.WorkflowGroup.GoalStatement != "Resume the retained package for this AI Organization after a reboot or reload." {
		t.Fatalf("unexpected fallback goal statement: %+v", response.ExecutionContract.WorkflowGroup)
	}
}

package server

func testReviewLoopHome() OrganizationHomePayload {
	return normalizeOrganizationHome(OrganizationHomePayload{
		OrganizationSummary: OrganizationSummary{
			ID:                        "org-review",
			Name:                      "Northstar Labs",
			Purpose:                   "Keep delivery reviews safe and structured",
			StartMode:                 OrganizationStartModeTemplate,
			Status:                    "ready",
			TeamLeadLabel:             "Team Lead",
			AdvisorCount:              1,
			DepartmentCount:           1,
			SpecialistCount:           2,
			AIEngineProfileID:         string(OrganizationAIEngineProfileBalanced),
			AIEngineSettingsSummary:   "Balanced",
			ResponseContractProfileID: string(ResponseContractProfileClearBalanced),
			ResponseContractSummary:   "Clear & Balanced",
			MemoryPersonalitySummary:  "Prepared for guided work",
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
						HelpsWith:                        "Turns organization goals into practical next steps.",
						AIEngineBindingProfileID:         string(OrganizationAIEngineProfileHighReasoning),
						ResponseContractBindingProfileID: string(ResponseContractProfileStructuredAnalytical),
					},
					{
						ID:                               "reviewer",
						Name:                             "Reviewer",
						HelpsWith:                        "Checks work for quality and readiness.",
						AIEngineBindingProfileID:         string(OrganizationAIEngineProfileHighReasoning),
						ResponseContractBindingProfileID: string(ResponseContractProfileStructuredAnalytical),
					},
				},
			},
		},
	})
}

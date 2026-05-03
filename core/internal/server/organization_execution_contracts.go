package server

import (
	"fmt"
	"strings"
)

func firstDepartmentStep(home OrganizationHomePayload) string {
	if home.DepartmentCount > 0 {
		return fmt.Sprintf("Use %d Department%s as the first routing layer for work.", home.DepartmentCount, pluralSuffix(home.DepartmentCount))
	}
	return "Define the first Department so the Team Lead has a clear execution lane."
}

func firstSpecialistStep(home OrganizationHomePayload) string {
	if home.SpecialistCount > 0 {
		return fmt.Sprintf("Bring %d Specialist%s in after the Team Lead confirms the plan.", home.SpecialistCount, pluralSuffix(home.SpecialistCount))
	}
	return "Add Specialists only after the Team Lead confirms the first Department-level plan."
}

func firstAdvisorStep(home OrganizationHomePayload) string {
	if home.AdvisorCount > 0 {
		return fmt.Sprintf("Use %d Advisor%s when the Team Lead needs review or decision support.", home.AdvisorCount, pluralSuffix(home.AdvisorCount))
	}
	return "Decide whether advisor guidance is needed before the next planning cycle."
}

func firstFocusSummary(home OrganizationHomePayload) string {
	if home.StartMode == OrganizationStartModeTemplate && strings.TrimSpace(home.TemplateName) != "" {
		return fmt.Sprintf("Start by using %s as the first working shape, then let the Team Lead confirm which part of the organization should lead.", home.TemplateName)
	}
	return fmt.Sprintf("Start by confirming the first outcome this AI Organization should deliver around %s, then let the Team Lead shape the initial structure around that goal.", safePurposeText(home.Purpose))
}

func buildTeamLeadExecutionContract(home OrganizationHomePayload, requestContext string) *TeamLeadExecutionContract {
	normalized := strings.TrimSpace(strings.ToLower(requestContext))
	if normalized == "" {
		return nil
	}

	if referencesExternalWorkflowContract(normalized) {
		return &TeamLeadExecutionContract{
			ExecutionMode:  TeamLeadExecutionModeExternalWorkflowContract,
			OwnerLabel:     "External workflow contract",
			ExternalTarget: externalWorkflowTarget(normalized),
			Summary:        "This request is best handled as an external workflow contract so Mycelis can keep invocation posture, governance, and normalized result return clear without pretending the external graph is a native team.",
			TargetOutputs: []string{
				"Normalized workflow result",
				"Linked artifact or execution note",
			},
			Workstreams: buildExternalWorkflowWorkstreams(externalWorkflowTarget(normalized)),
		}
	}

	if isBroadCoordinationRequest(normalized) {
		teamName := "Program Orchestration Team"
		outputs := []string{
			"Program orchestration brief",
			"Per-team delivery plans",
			"Cross-team coordination summary",
		}
		return &TeamLeadExecutionContract{
			ExecutionMode:              TeamLeadExecutionModeNativeTeam,
			OwnerLabel:                 "Soma and Council orchestration",
			TeamName:                   teamName,
			CoordinationModel:          "multi_team_orchestration",
			RecommendedTeamShape:       "Several small teams coordinated by Soma and Council over NATS/exchange, with no single team exceeding the member cap.",
			RecommendedTeamCount:       3,
			RecommendedTeamMemberLimit: 5,
			Summary:                    fmt.Sprintf("This request is broad enough to split into several compact teams instead of one oversized group. Use Soma and Council to coordinate the lanes over NATS/exchange, keep each team small, and return one orchestration summary plus the team-level outputs for %s.", safeOrganizationName(home.Name)),
			TargetOutputs:              outputs,
			Workstreams:                buildMultiTeamExecutionWorkstreams(outputs),
			WorkflowGroup:              buildWorkflowGroupDraft(home, teamName, requestContext, "propose_only", outputs, []string{"team.coordinate", "artifact.review", "broadcast"}, 5),
		}
	}

	if referencesImageTeamOutput(normalized) {
		teamName := "Creative Delivery Team"
		outputs := []string{
			"Reviewable image artifact",
			"Short concept note",
		}
		return &TeamLeadExecutionContract{
			ExecutionMode:              TeamLeadExecutionModeNativeTeam,
			OwnerLabel:                 "Native Mycelis team",
			TeamName:                   teamName,
			CoordinationModel:          "compact_team",
			RecommendedTeamShape:       "One focused team with a small specialist roster.",
			RecommendedTeamCount:       1,
			RecommendedTeamMemberLimit: 6,
			Summary:                    fmt.Sprintf("Use a bounded creative team inside %s so Soma can shape the work, route it through the right specialists, and return the generated image as a managed artifact.", safeOrganizationName(home.Name)),
			TargetOutputs:              outputs,
			Workstreams:                buildCreativeExecutionWorkstreams(teamName, outputs),
			WorkflowGroup:              buildWorkflowGroupDraft(home, teamName, requestContext, "propose_only", outputs, []string{"content.plan", "artifact.review"}, 6),
		}
	}

	if teamName, outputs, ok := inferBusinessTeamExecution(normalized); ok {
		return &TeamLeadExecutionContract{
			ExecutionMode:              TeamLeadExecutionModeNativeTeam,
			OwnerLabel:                 "Native Mycelis team",
			TeamName:                   teamName,
			CoordinationModel:          "compact_team",
			RecommendedTeamShape:       "One focused team with a small specialist roster.",
			RecommendedTeamCount:       1,
			RecommendedTeamMemberLimit: 6,
			Summary:                    fmt.Sprintf("Use a bounded %s lane inside %s so Soma can stand up a focused delivery group, coordinate the right specialists, and keep the resulting outputs reviewable in one place.", strings.ToLower(teamName), safeOrganizationName(home.Name)),
			TargetOutputs:              outputs,
			Workstreams:                buildCompactExecutionWorkstreams(teamName, outputs),
			WorkflowGroup:              buildWorkflowGroupDraft(home, teamName, requestContext, "propose_only", outputs, []string{"team.coordinate", "artifact.review"}, 6),
		}
	}

	return nil
}

func buildRetainedPackageContinuityContract(home OrganizationHomePayload, requestContext string) *TeamLeadExecutionContract {
	organizationName := safeOrganizationName(home.Name)
	resumeGoal := strings.TrimSpace(requestContext)
	if resumeGoal == "" {
		resumeGoal = fmt.Sprintf("Resume the retained package for %s after a reboot or reload.", organizationName)
	}

	targetOutputs := []string{
		"Retained package continuity summary",
		"Completed work snapshot",
		"Remaining work checklist",
	}
	return &TeamLeadExecutionContract{
		ExecutionMode:     TeamLeadExecutionModeContinuityResume,
		OwnerLabel:        "Team Lead continuity",
		Summary:           fmt.Sprintf("Resume the retained package for %s, confirm completed work, and keep the remaining steps reviewable after a reboot or reload.", organizationName),
		ContinuityLabel:   "Retained package continuity",
		ContinuitySummary: fmt.Sprintf("Continuity resumes from the last durable outputs for %s without rebuilding finished work.", organizationName),
		ResumeCheckpoint:  "Continue from the last retained package after reload or reboot.",
		TargetOutputs:     targetOutputs,
		Workstreams:       buildContinuityExecutionWorkstreams(),
		WorkflowGroup:     buildWorkflowGroupDraft(home, "Retained Package Continuity", resumeGoal, "resume_continuity", targetOutputs, []string{"artifact.review", "team.coordinate"}, 4),
	}
}

func inferBusinessTeamExecution(normalized string) (string, []string, bool) {
	switch {
	case strings.Contains(normalized, "marketing"):
		return "Marketing Launch Team", []string{
			"Launch plan",
			"Messaging brief",
			"Campaign asset list",
		}, true
	case strings.Contains(normalized, "customer research") || strings.Contains(normalized, "user research"):
		return "Customer Research Team", []string{
			"Research brief",
			"Interview guide",
			"Findings summary",
		}, true
	case strings.Contains(normalized, "revops") || strings.Contains(normalized, "revenue operations") || strings.Contains(normalized, "lead routing"):
		return "Revenue Operations Team", []string{
			"Workflow recommendation",
			"Operational checklist",
			"Handoff summary",
		}, true
	case strings.Contains(normalized, "security"):
		return "Security Review Team", []string{
			"Risk review",
			"Mitigation checklist",
			"Approval notes",
		}, true
	case strings.Contains(normalized, "creative team"):
		return "Creative Delivery Team", []string{
			"Creative brief",
			"Artifact package",
			"Review notes",
		}, true
	default:
		return "", nil, false
	}
}

func isBroadCoordinationRequest(normalized string) bool {
	breadthSignals := 0
	for _, marker := range []string{
		"company-wide",
		"organization-wide",
		"cross-functional",
		"multi-team",
		"multiple teams",
		"several teams",
		"several workstreams",
		"whole organization",
		"enterprise-wide",
		"all teams",
		"across teams",
		"across functions",
	} {
		if strings.Contains(normalized, marker) {
			breadthSignals++
		}
	}
	if breadthSignals >= 1 {
		return true
	}
	return strings.Contains(normalized, "program") && (strings.Contains(normalized, "across") || strings.Contains(normalized, "multiple") || strings.Contains(normalized, "several") || strings.Contains(normalized, "cross"))
}

func buildWorkflowGroupDraft(home OrganizationHomePayload, teamName, requestContext, workMode string, targetOutputs, allowedCapabilities []string, recommendedMemberLimit int) *TeamLeadWorkflowGroupDraft {
	organizationName := safeOrganizationName(home.Name)
	goal := strings.TrimSpace(requestContext)
	if goal == "" {
		goal = fmt.Sprintf("Coordinate a focused %s workflow inside %s.", strings.ToLower(teamName), organizationName)
	}
	return &TeamLeadWorkflowGroupDraft{
		Name:                   fmt.Sprintf("%s temporary workflow", teamName),
		GoalStatement:          goal,
		WorkMode:               workMode,
		CoordinatorProfile:     fmt.Sprintf("%s lead", teamName),
		AllowedCapabilities:    normalizeExecutionCapabilityList(allowedCapabilities),
		RecommendedMemberLimit: recommendedMemberLimit,
		ExpiryHours:            72,
		Summary:                fmt.Sprintf("Launch a temporary workflow group for %s, keep coordination bounded to at most %d members, and retain outputs like %s after the lane is archived.", teamName, recommendedMemberLimit, humanJoin(targetOutputs)),
	}
}

package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

func (s *AdminServer) handleTeamLeadGuidedAction(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		respondAPIError(w, "organization id is required", http.StatusBadRequest)
		return
	}

	home, ok := s.organizationStore().Get(id)
	if !ok {
		respondAPIError(w, "organization not found", http.StatusNotFound)
		return
	}

	var req TeamLeadGuidanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondAPIError(w, "invalid Team Lead action request", http.StatusBadRequest)
		return
	}

	response, err := s.buildTeamLeadGuidanceResponse(r.Context(), home, req.Action, req.RequestContext)
	if err != nil {
		respondAPIError(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.emitReviewLoopEvent(home.ID, ReviewLoopEventTeamLeadActionCompleted)
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(response))
}

func (s *AdminServer) buildTeamLeadGuidanceResponse(ctx context.Context, home OrganizationHomePayload, action TeamLeadGuidedAction, requestContext string) (TeamLeadGuidanceResponse, error) {
	response, err := buildTeamLeadGuidance(home, action, requestContext)
	if err != nil {
		return TeamLeadGuidanceResponse{}, err
	}
	if action != TeamLeadGuidedActionResumeRetainedPackage {
		return response, nil
	}

	enriched, ok := s.buildRetainedPackageContinuityFromState(ctx, home, requestContext)
	if !ok {
		return response, nil
	}
	return enriched, nil
}

func buildTeamLeadGuidance(home OrganizationHomePayload, action TeamLeadGuidedAction, requestContext string) (TeamLeadGuidanceResponse, error) {
	organizationName := safeOrganizationName(home.Name)
	teamLeadLabel := safeTeamLeadLabel(home.TeamLeadLabel)
	purposeText := safePurposeText(home.Purpose)

	switch action {
	case TeamLeadGuidedActionPlanNextSteps:
		executionContract := buildTeamLeadExecutionContract(home, requestContext)
		steps := []string{
			fmt.Sprintf("Align the first outcome with this purpose: %s.", purposeText),
			firstDepartmentStep(home),
			firstSpecialistStep(home),
		}
		return TeamLeadGuidanceResponse{
			Action:        action,
			RequestLabel:  "Plan next steps for this organization",
			Headline:      fmt.Sprintf("Team Lead plan for %s", organizationName),
			Summary:       fmt.Sprintf("%s recommends moving %s from setup into a focused first delivery loop.", teamLeadLabel, organizationName),
			PrioritySteps: steps,
			SuggestedFollowUps: []string{
				"Review my organization setup",
				"What should I focus on first?",
				templateSpecificSuggestion(home),
			},
			ExecutionContract: executionContract,
		}, nil
	case TeamLeadGuidedActionFocusFirst:
		executionContract := buildTeamLeadExecutionContract(home, requestContext)
		return TeamLeadGuidanceResponse{
			Action:       action,
			RequestLabel: "What should I focus on first?",
			Headline:     fmt.Sprintf("First focus for %s", organizationName),
			Summary:      firstFocusSummary(home),
			PrioritySteps: []string{
				firstDepartmentStep(home),
				firstAdvisorStep(home),
				"Keep the Team Lead as the primary working counterpart while the organization takes shape.",
			},
			SuggestedFollowUps: []string{
				"Plan next steps for this organization",
				"Review my organization setup",
				"Review the Team Lead guidance before expanding into deeper structure.",
			},
			ExecutionContract: executionContract,
		}, nil
	case TeamLeadGuidedActionReviewSetup:
		executionContract := buildTeamLeadExecutionContract(home, requestContext)
		return TeamLeadGuidanceResponse{
			Action:       action,
			RequestLabel: "Review my organization setup",
			Headline:     fmt.Sprintf("Organization setup review for %s", organizationName),
			Summary:      fmt.Sprintf("%s is ready to review the current AI Organization shape before the next action begins.", teamLeadLabel),
			PrioritySteps: []string{
				fmt.Sprintf("Advisors: %s.", formatConfiguredCountForGuidance(home.AdvisorCount, "advisor")),
				fmt.Sprintf("Departments: %s.", formatConfiguredCountForGuidance(home.DepartmentCount, "department")),
				fmt.Sprintf("Specialists: %s.", formatConfiguredCountForGuidance(home.SpecialistCount, "specialist")),
			},
			SuggestedFollowUps: []string{
				"Plan next steps for this organization",
				"What should I focus on first?",
				fmt.Sprintf("Review the %s summary and confirm the Team Lead has what it needs.", home.startingPointLabel()),
			},
			ExecutionContract: executionContract,
		}, nil
	case TeamLeadGuidedActionResumeRetainedPackage:
		executionContract := buildRetainedPackageContinuityContract(home, requestContext)
		return TeamLeadGuidanceResponse{
			Action:       action,
			RequestLabel: "Resume retained package continuity",
			Headline:     fmt.Sprintf("Retained package continuity for %s", organizationName),
			Summary:      fmt.Sprintf("%s resumes the retained package for %s so completed work stays durable and the next step stays explicit after a reboot or reload.", teamLeadLabel, organizationName),
			PrioritySteps: []string{
				"Open the retained package and confirm the latest durable outputs.",
				"Record what is already complete, what remains, and who owns the next step.",
				"Continue from the retained package without rebuilding finished work.",
			},
			SuggestedFollowUps: []string{
				"Plan next steps for this organization",
				"Review my organization setup",
				"Use the retained package continuity contract as the starting point for UI or live test automation.",
			},
			ExecutionContract: executionContract,
		}, nil
	default:
		return TeamLeadGuidanceResponse{}, fmt.Errorf("action must be plan_next_steps, focus_first, review_setup, or resume_retained_package")
	}
}

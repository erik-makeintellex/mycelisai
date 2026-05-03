package server

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
)

func defaultReviewLoopProfiles(home OrganizationHomePayload) []LoopProfile {
	home = normalizeOrganizationHome(home)
	profiles := make([]LoopProfile, 0, 2)

	if len(home.Departments) > 0 {
		firstDepartment := home.Departments[0]
		profiles = append(profiles, LoopProfile{
			ID:          DefaultDepartmentReviewLoopID,
			Name:        "Department readiness review",
			Type:        LoopProfileTypeReview,
			Description: "Reviews the current Department structure and operating readiness without taking action.",
			Owner: LoopOwnerRef{
				Type: LoopOwnerTypeTeam,
				ID:   firstDepartment.ID,
			},
			IntervalSeconds: 60,
			EventTriggers: []ReviewLoopEventKind{
				ReviewLoopEventOrganizationCreated,
				ReviewLoopEventTeamLeadActionCompleted,
				ReviewLoopEventOrganizationAIEngineChanged,
				ReviewLoopEventResponseContractChanged,
			},
		})

		if len(firstDepartment.AgentTypeProfiles) > 0 {
			firstProfile := firstDepartment.AgentTypeProfiles[0]
			profiles = append(profiles, LoopProfile{
				ID:          DefaultAgentTypeReviewLoopID,
				Name:        "Agent type readiness review",
				Type:        LoopProfileTypeReview,
				Description: "Reviews a specialist profile and its inherited defaults without taking action.",
				Owner: LoopOwnerRef{
					Type: LoopOwnerTypeAgentType,
					ID:   firstProfile.ID,
				},
				EventTriggers: []ReviewLoopEventKind{
					ReviewLoopEventOrganizationCreated,
					ReviewLoopEventOrganizationAIEngineChanged,
					ReviewLoopEventResponseContractChanged,
				},
			})
		}
	}

	return profiles
}

func resolveLoopOwner(home OrganizationHomePayload, profile LoopProfile) (LoopOwnerResolution, error) {
	switch profile.Owner.Type {
	case LoopOwnerTypeTeam:
		for _, department := range home.Departments {
			if department.ID == profile.Owner.ID {
				return LoopOwnerResolution{
					Type:      LoopOwnerTypeTeam,
					ID:        department.ID,
					Name:      department.Name,
					HelpsWith: fmt.Sprintf("%s organizes %d Specialist%s for this AI Organization.", department.Name, department.SpecialistCount, pluralSuffix(department.SpecialistCount)),
				}, nil
			}
		}
		return LoopOwnerResolution{}, fmt.Errorf("loop owner team not found")
	case LoopOwnerTypeAgentType:
		for _, department := range home.Departments {
			for _, agentType := range department.AgentTypeProfiles {
				if agentType.ID == profile.Owner.ID {
					return LoopOwnerResolution{
						Type:      LoopOwnerTypeAgentType,
						ID:        agentType.ID,
						Name:      agentType.Name,
						HelpsWith: agentType.HelpsWith,
					}, nil
				}
			}
		}
		return LoopOwnerResolution{}, fmt.Errorf("loop owner agent type not found")
	default:
		return LoopOwnerResolution{}, fmt.Errorf("loop owner type must be team or agent_type")
	}
}

func buildReviewLoopOutput(home OrganizationHomePayload, owner LoopOwnerResolution) ReviewLoopStructuredOutput {
	findings := []string{
		fmt.Sprintf("%s is currently operating with %d Department%s, %d Specialist%s, and %d Advisor%s.", safeOrganizationName(home.Name), home.DepartmentCount, pluralSuffix(home.DepartmentCount), home.SpecialistCount, pluralSuffix(home.SpecialistCount), home.AdvisorCount, pluralSuffix(home.AdvisorCount)),
		fmt.Sprintf("%s is the active review owner for this loop.", owner.Name),
	}
	suggestions := []string{
		"Keep this loop read-only until Automations visibility and policy surfaces are in place.",
	}

	status := "healthy"
	flaggedItems := 0
	if home.DepartmentCount == 0 {
		status = "attention_needed"
		flaggedItems++
		findings = append(findings, "No Departments are configured yet, so the Team Lead does not have a clear execution lane.")
		suggestions = append(suggestions, "Add at least one Department before broadening beyond Team Lead guidance.")
	} else {
		findings = append(findings, fmt.Sprintf("The active AI Engine default is %s.", strings.TrimSpace(home.AIEngineSettingsSummary)))
		suggestions = append(suggestions, "Use the review findings to confirm whether the current Department structure still matches the organization's first priorities.")
	}

	if home.SpecialistCount == 0 {
		status = "attention_needed"
		flaggedItems++
		findings = append(findings, "No Specialists are currently attached, so follow-through still depends on Team Lead planning alone.")
		suggestions = append(suggestions, "Add Specialists only after the Team Lead confirms the first execution lane.")
	}

	if strings.TrimSpace(home.ResponseContractSummary) == "" {
		status = "attention_needed"
		flaggedItems++
		findings = append(findings, "No Response Style is visible yet for this AI Organization.")
		suggestions = append(suggestions, "Set a Response Style so future guidance stays consistent and reviewable.")
	} else {
		findings = append(findings, fmt.Sprintf("The current Response Style is %s.", strings.TrimSpace(home.ResponseContractSummary)))
	}

	if owner.Type == LoopOwnerTypeAgentType {
		findings = append(findings, fmt.Sprintf("%s currently helps with: %s.", owner.Name, owner.HelpsWith))
		suggestions = append(suggestions, "Review whether this specialist profile should keep inheriting defaults or eventually receive a bounded override.")
	} else {
		suggestions = append(suggestions, "Review the Department summary before widening into live loop execution or external actions.")
	}

	return ReviewLoopStructuredOutput{
		Status:       status,
		FlaggedItems: flaggedItems,
		Findings:     findings,
		Suggestions:  suggestions,
	}
}

func (s *AdminServer) executeReviewLoop(home OrganizationHomePayload, profile LoopProfile, trigger string) (ReviewLoopResult, error) {
	if err := validateLoopProfile(profile); err != nil {
		return ReviewLoopResult{}, err
	}

	home = normalizeOrganizationHome(home)
	owner, err := resolveLoopOwner(home, profile)
	if err != nil {
		return ReviewLoopResult{}, err
	}

	result := ReviewLoopResult{
		ID:               uuid.NewString(),
		LoopID:           profile.ID,
		LoopName:         profile.Name,
		LoopType:         profile.Type,
		OrganizationID:   home.ID,
		OrganizationName: safeOrganizationName(home.Name),
		Trigger:          trigger,
		Owner:            owner,
		Review:           buildReviewLoopOutput(home, owner),
		ReviewedAt:       time.Now().UTC().Format(time.RFC3339),
	}
	result.ActivityStatus, result.ActivitySummary = summarizeActivityFromReview(result.Review)

	s.loopResultStore().Add(home.ID, result)
	log.Printf("[review-loop] organization=%s loop=%s owner_type=%s owner_id=%s status=%s findings=%d suggestions=%d", home.ID, profile.ID, owner.Type, owner.ID, result.Review.Status, len(result.Review.Findings), len(result.Review.Suggestions))
	return result, nil
}

func summarizeActivityFromReview(review ReviewLoopStructuredOutput) (LoopActivityStatus, string) {
	if review.Status == "attention_needed" {
		flagged := review.FlaggedItems
		if flagged <= 0 {
			flagged = 1
		}
		return LoopActivityStatusWarning, fmt.Sprintf("%d item%s flagged", flagged, pluralSuffix(flagged))
	}
	return LoopActivityStatusSuccess, "No issues detected"
}

func summarizeActivityFromError(err error) (LoopActivityStatus, string) {
	if err == nil {
		return LoopActivityStatusFailed, "Review unavailable"
	}
	message := strings.TrimSpace(err.Error())
	switch {
	case strings.Contains(message, "owner"):
		return LoopActivityStatusFailed, "Review owner unavailable"
	case strings.Contains(message, "interval_seconds"):
		return LoopActivityStatusFailed, "Review timing needs attention"
	default:
		return LoopActivityStatusFailed, "Review unavailable"
	}
}

func validateLoopProfile(profile LoopProfile) error {
	if profile.Type != LoopProfileTypeReview {
		return fmt.Errorf("loop profile must be a review loop")
	}
	if strings.TrimSpace(profile.ID) == "" {
		return fmt.Errorf("loop profile id is required")
	}
	if profile.Owner.Type != LoopOwnerTypeTeam && profile.Owner.Type != LoopOwnerTypeAgentType {
		return fmt.Errorf("loop owner type must be team or agent_type")
	}
	if strings.TrimSpace(profile.Owner.ID) == "" {
		return fmt.Errorf("loop owner id is required")
	}
	if profile.IntervalSeconds < 0 {
		return fmt.Errorf("interval_seconds must be greater than or equal to zero")
	}
	for _, trigger := range profile.EventTriggers {
		if !isAllowedReviewLoopEventKind(trigger) {
			return fmt.Errorf("event_triggers must use allowed internal review events only")
		}
	}
	return nil
}

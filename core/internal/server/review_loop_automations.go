package server

import (
	"fmt"
	"strings"
)

func safeAutomationName(profile LoopProfile) string {
	name := strings.TrimSpace(profile.Name)
	if name == "" {
		return "Organization review"
	}
	return name
}

func safeAutomationPurpose(profile LoopProfile, owner LoopOwnerResolution) string {
	description := strings.TrimSpace(profile.Description)
	if description != "" {
		return description
	}
	switch owner.Type {
	case LoopOwnerTypeTeam:
		return fmt.Sprintf("Keeps %s under review so the Team Lead can see whether the current structure is still ready.", owner.Name)
	case LoopOwnerTypeAgentType:
		return fmt.Sprintf("Keeps the %s specialist role under review so inherited defaults stay safe and understandable.", owner.Name)
	default:
		return "Keeps part of the AI Organization under review without taking action."
	}
}

func safeAutomationOwnerLabel(owner LoopOwnerResolution) string {
	switch owner.Type {
	case LoopOwnerTypeTeam:
		return fmt.Sprintf("Team: %s", owner.Name)
	case LoopOwnerTypeAgentType:
		return fmt.Sprintf("Specialist role: %s", owner.Name)
	default:
		if strings.TrimSpace(owner.Name) != "" {
			return owner.Name
		}
		return "Owner unavailable"
	}
}

func safeAutomationWatches(home OrganizationHomePayload, owner LoopOwnerResolution) string {
	switch owner.Type {
	case LoopOwnerTypeTeam:
		return fmt.Sprintf("Watches %s structure, specialist coverage, and current organization defaults inside AI Organization %s.", owner.Name, safeOrganizationName(home.Name))
	case LoopOwnerTypeAgentType:
		return fmt.Sprintf("Watches the %s specialist role, its working focus, and the defaults it inherits inside AI Organization %s.", owner.Name, safeOrganizationName(home.Name))
	default:
		return fmt.Sprintf("Watches the current organization setup and guided defaults inside AI Organization %s.", safeOrganizationName(home.Name))
	}
}

func automationTriggerType(profile LoopProfile) AutomationTriggerType {
	if profile.IntervalSeconds > 0 {
		return AutomationTriggerTypeScheduled
	}
	return AutomationTriggerTypeEventDriven
}

func humanizeLoopInterval(seconds int) string {
	if seconds <= 0 {
		return ""
	}
	if seconds < 60 {
		if seconds == 1 {
			return "every second"
		}
		return fmt.Sprintf("every %d seconds", seconds)
	}
	if seconds%3600 == 0 {
		hours := seconds / 3600
		if hours == 1 {
			return "every hour"
		}
		return fmt.Sprintf("every %d hours", hours)
	}
	if seconds%60 == 0 {
		minutes := seconds / 60
		if minutes == 1 {
			return "every minute"
		}
		return fmt.Sprintf("every %d minutes", minutes)
	}
	return fmt.Sprintf("every %d seconds", seconds)
}

func safeAutomationEventLabel(eventKind ReviewLoopEventKind) string {
	switch eventKind {
	case ReviewLoopEventOrganizationCreated:
		return "organization setup"
	case ReviewLoopEventTeamLeadActionCompleted:
		return "Team Lead guidance"
	case ReviewLoopEventOrganizationAIEngineChanged:
		return "AI Engine changes"
	case ReviewLoopEventResponseContractChanged:
		return "Response Style changes"
	default:
		return ""
	}
}

func joinWithOr(items []string) string {
	switch len(items) {
	case 0:
		return ""
	case 1:
		return items[0]
	case 2:
		return items[0] + " or " + items[1]
	default:
		return strings.Join(items[:len(items)-1], ", ") + ", or " + items[len(items)-1]
	}
}

func safeAutomationTriggerSummary(profile LoopProfile) string {
	eventLabels := make([]string, 0, len(profile.EventTriggers))
	for _, trigger := range profile.EventTriggers {
		if label := safeAutomationEventLabel(trigger); label != "" {
			eventLabels = append(eventLabels, label)
		}
	}

	switch {
	case profile.IntervalSeconds > 0 && len(eventLabels) > 0:
		return fmt.Sprintf("Runs %s and also after %s.", humanizeLoopInterval(profile.IntervalSeconds), joinWithOr(eventLabels))
	case profile.IntervalSeconds > 0:
		return fmt.Sprintf("Runs %s.", humanizeLoopInterval(profile.IntervalSeconds))
	case len(eventLabels) > 0:
		return fmt.Sprintf("Runs after %s.", joinWithOr(eventLabels))
	default:
		return "Runs when this review is triggered from the organization workspace."
	}
}

func safeAutomationOwner(home OrganizationHomePayload, profile LoopProfile) LoopOwnerResolution {
	owner, err := resolveLoopOwner(home, profile)
	if err == nil {
		return owner
	}

	switch profile.Owner.Type {
	case LoopOwnerTypeTeam:
		return LoopOwnerResolution{
			Type:      LoopOwnerTypeTeam,
			ID:        profile.Owner.ID,
			Name:      "Team owner unavailable",
			HelpsWith: "This Automation is waiting for its Team owner to become available again.",
		}
	case LoopOwnerTypeAgentType:
		return LoopOwnerResolution{
			Type:      LoopOwnerTypeAgentType,
			ID:        profile.Owner.ID,
			Name:      "Specialist role unavailable",
			HelpsWith: "This Automation is waiting for its Specialist role to become available again.",
		}
	default:
		return LoopOwnerResolution{
			Type:      profile.Owner.Type,
			ID:        profile.Owner.ID,
			Name:      "Owner unavailable",
			HelpsWith: "This Automation is waiting for its owner to become available again.",
		}
	}
}

func recentAutomationOutcomes(results []ReviewLoopResult, loopID string, limit int) []AutomationOutcomeItem {
	if limit <= 0 {
		limit = 3
	}

	items := make([]AutomationOutcomeItem, 0, limit)
	for _, result := range results {
		if result.LoopID != loopID {
			continue
		}
		items = append(items, AutomationOutcomeItem{
			Summary:    strings.TrimSpace(result.ActivitySummary),
			OccurredAt: result.ReviewedAt,
		})
		if len(items) == limit {
			break
		}
	}
	return items
}

func automationStatusForLoop(results []ReviewLoopResult, loopID string) LoopActivityStatus {
	for _, result := range results {
		if result.LoopID == loopID {
			if result.ActivityStatus == "" {
				return LoopActivityStatusSuccess
			}
			return result.ActivityStatus
		}
	}
	return LoopActivityStatusSuccess
}

func organizationAutomations(home OrganizationHomePayload, profiles []LoopProfile, results []ReviewLoopResult) []OrganizationAutomationItem {
	if len(profiles) == 0 {
		return nil
	}

	items := make([]OrganizationAutomationItem, 0, len(profiles))
	for _, profile := range profiles {
		owner := safeAutomationOwner(home, profile)
		items = append(items, OrganizationAutomationItem{
			ID:             profile.ID,
			Name:           safeAutomationName(profile),
			Purpose:        safeAutomationPurpose(profile, owner),
			TriggerType:    automationTriggerType(profile),
			OwnerLabel:     safeAutomationOwnerLabel(owner),
			Status:         automationStatusForLoop(results, profile.ID),
			Watches:        safeAutomationWatches(home, owner),
			TriggerSummary: safeAutomationTriggerSummary(profile),
			RecentOutcomes: recentAutomationOutcomes(results, profile.ID, 3),
		})
	}

	return items
}

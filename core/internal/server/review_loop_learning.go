package server

import (
	"fmt"
	"strings"
)

func learningInsightStrength(results []ReviewLoopResult, loopID string) LearningInsightStrength {
	count := 0
	for _, result := range results {
		if result.LoopID == loopID {
			count++
		}
	}

	switch {
	case count >= 3:
		return LearningInsightStrengthStrong
	case count >= 2:
		return LearningInsightStrengthConsistent
	default:
		return LearningInsightStrengthEmerging
	}
}

func learningInsightSourceLabel(owner LoopOwnerResolution) string {
	label := strings.TrimSpace(safeAutomationOwnerLabel(owner))
	if label != "" && label != "Owner unavailable" {
		return label
	}
	return "Organization review"
}

func lowerFirst(input string) string {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return ""
	}
	return strings.ToLower(trimmed[:1]) + trimmed[1:]
}

func leadingVerbToGerund(input string) string {
	parts := strings.Fields(strings.TrimSpace(input))
	if len(parts) == 0 {
		return ""
	}

	word := strings.ToLower(parts[0])
	switch {
	case strings.HasSuffix(word, "ies") && len(word) > 3:
		parts[0] = parts[0][:len(parts[0])-3] + "ying"
	case strings.HasSuffix(word, "es") && len(word) > 2:
		parts[0] = parts[0][:len(parts[0])-2] + "ing"
	case strings.HasSuffix(word, "s") && len(word) > 1:
		parts[0] = parts[0][:len(parts[0])-1] + "ing"
	default:
		parts[0] = parts[0] + "ing"
	}

	return strings.Join(parts, " ")
}

func learningInsightAction(owner LoopOwnerResolution) string {
	action := strings.TrimSpace(owner.HelpsWith)
	action = strings.TrimSuffix(action, ".")
	if action == "" {
		return "the current work"
	}
	return lowerFirst(leadingVerbToGerund(action))
}

func learningInsightRoleSubject(name string) string {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "Specialists"
	}
	if strings.HasSuffix(strings.ToLower(trimmed), "specialist") {
		return trimmed + "s"
	}
	return trimmed + " specialists"
}

func summarizeLearningInsight(result ReviewLoopResult) string {
	switch result.Owner.Type {
	case LoopOwnerTypeTeam:
		teamName := strings.TrimSpace(result.Owner.Name)
		if teamName == "" {
			teamName = "This team"
		}
		switch result.ActivityStatus {
		case LoopActivityStatusWarning:
			return fmt.Sprintf("%s is highlighting recurring readiness gaps that still need attention.", teamName)
		case LoopActivityStatusFailed:
			return fmt.Sprintf("%s has not reported a clear learning update yet.", teamName)
		default:
			return fmt.Sprintf("%s is building a steadier execution lane for the organization.", teamName)
		}
	case LoopOwnerTypeAgentType:
		subject := learningInsightRoleSubject(result.Owner.Name)
		action := learningInsightAction(result.Owner)
		switch result.ActivityStatus {
		case LoopActivityStatusWarning:
			return fmt.Sprintf("%s are identifying recurring gaps while %s.", subject, action)
		case LoopActivityStatusFailed:
			return fmt.Sprintf("%s have not reported a clear learning update yet.", subject)
		default:
			return fmt.Sprintf("%s are getting more consistent at %s.", subject, action)
		}
	default:
		switch result.ActivityStatus {
		case LoopActivityStatusWarning:
			return "The organization is identifying recurring gaps that still need attention."
		case LoopActivityStatusFailed:
			return "The organization has not reported a clear learning update yet."
		default:
			return "The organization is getting more consistent at carrying its work forward."
		}
	}
}

func organizationLearningInsights(results []ReviewLoopResult, limit int) []OrganizationLearningInsightItem {
	if len(results) == 0 {
		return nil
	}
	if limit <= 0 {
		limit = 4
	}

	items := make([]OrganizationLearningInsightItem, 0, limit)
	seenLoopIDs := make(map[string]bool)
	for _, result := range results {
		if seenLoopIDs[result.LoopID] {
			continue
		}
		seenLoopIDs[result.LoopID] = true

		items = append(items, OrganizationLearningInsightItem{
			ID:         result.ID,
			Summary:    summarizeLearningInsight(result),
			Source:     learningInsightSourceLabel(result.Owner),
			ObservedAt: result.ReviewedAt,
			Strength:   learningInsightStrength(results, result.LoopID),
		})
		if len(items) == limit {
			break
		}
	}

	return items
}

package server

import (
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

func teamWorkActionCopy(action protocol.TeamWorkAction) (string, string, string) {
	switch action {
	case protocol.TeamWorkActionStartWork:
		return "Team work started", "The operator moved this durable work item into active execution.", "Watch for team status, retained output, or proof."
	case protocol.TeamWorkActionPause:
		return "Team work paused", "The operator paused this durable work item.", "Resume or archive this work when ready."
	case protocol.TeamWorkActionResume:
		return "Resume requested", "The operator returned this durable work item to the queue.", "Wait for the team to continue or add steering guidance."
	case protocol.TeamWorkActionArchive:
		return "Team work archived", "The operator archived this durable work item. Retained outputs and proof remain inspectable.", "Review retained proof or start a new work item."
	case protocol.TeamWorkActionSteer:
		return "Team steering recorded", "The operator added steering guidance to this durable work item.", "Review the guidance and continue from the retained work state."
	case protocol.TeamWorkActionRecover:
		return "Recovery requested", "The operator moved this degraded work item back to the queue for safe continuation.", "Watch for new status, retained output, or proof."
	case protocol.TeamWorkActionVerifyExternalOutcome:
		return "External outcome verified", "The operator attested to the external mutation outcome.", "Review the recorded evidence and resulting posture."
	default:
		return "Team work updated", "The operator updated this durable work item.", "Review the latest state."
	}
}

func teamWorkExternalOutcomeCopy(result string) (string, string, string) {
	switch result {
	case protocol.WorkExternalOutcomeCommitted:
		return "External outcome verified as committed", "The operator attested that the external mutation committed.", "Review the retained outcome and proof."
	case protocol.WorkExternalOutcomeNotCommitted:
		return "External outcome verified as not committed", "The operator attested that the external mutation did not commit.", "Ask Soma to create a new governed proposal before attempting the mutation again."
	default:
		return "External outcome remains unknown", "The operator could not determine whether the external mutation committed.", "Continue external verification or archive this work."
	}
}

func teamWorkActionPayloadKind(action protocol.TeamWorkAction) string {
	if action == protocol.TeamWorkActionVerifyExternalOutcome {
		return "external_outcome_verification"
	}
	return "team_work_action"
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}

func mergeStrings(left, right []string) []string {
	seen := map[string]struct{}{}
	merged := make([]string, 0, len(left)+len(right))
	for _, value := range append(left, right...) {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		merged = append(merged, trimmed)
	}
	return merged
}

func stringSliceFromAny(value any) []string {
	values, ok := value.([]any)
	if !ok {
		stringValues, stringsOK := value.([]string)
		if stringsOK {
			return stringValues
		}
		return nil
	}
	result := make([]string, 0, len(values))
	for _, value := range values {
		if text, ok := value.(string); ok {
			result = append(result, text)
		}
	}
	return result
}

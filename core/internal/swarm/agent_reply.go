package swarm

import (
	"fmt"
	"strings"
)

func teamAgentRequestReply(result ProcessResult) string {
	if result.Availability != nil && !result.Availability.Available {
		code := strings.TrimSpace(result.Availability.Code)
		if code == "" {
			code = "unavailable"
		}
		summary := strings.TrimSpace(result.Availability.Summary)
		if summary == "" {
			summary = "The team could not complete the requested work."
		}
		reply := fmt.Sprintf("Work unavailable (%s): %s", code, summary)
		if recovery := strings.TrimSpace(result.Availability.RecommendedAction); recovery != "" {
			reply += " Recovery: " + recovery
		}
		return reply
	}
	return result.Text
}

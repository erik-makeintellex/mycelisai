package server

import (
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

func executableMutationPlan(isMutation bool, agentResult chatAgentResult, latestRequest string, mutTools []string) (bool, []string, []protocol.PlannedToolCall) {
	if !isMutation {
		return false, mutTools, nil
	}
	planned := buildPlannedToolCalls(agentResult, latestRequest, mutTools)
	if len(planned) == 0 {
		return false, nil, nil
	}
	return true, mutTools, planned
}

func isToolPostureGuidanceRequest(lower string) bool {
	trimmed := strings.TrimSpace(lower)
	if trimmed == "" {
		return false
	}
	if strings.Contains(trimmed, "enable now") || strings.Contains(trimmed, "install now") || strings.Contains(trimmed, "connect now") {
		return false
	}
	guidanceCues := []string{
		"show me",
		"check ",
		"list ",
		"review ",
		"what ",
		"which ",
		"tell me",
		"walk me through",
		"currently configured",
		"available tools",
	}
	for _, cue := range guidanceCues {
		if strings.HasPrefix(trimmed, cue) || strings.Contains(trimmed, " "+cue) {
			return true
		}
	}
	return false
}

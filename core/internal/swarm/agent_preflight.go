package swarm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

func extractNATSSubject(input string) string {
	for _, tok := range strings.Fields(input) {
		clean := strings.Trim(tok, " \t\r\n,.;:()[]{}<>\"'")
		if clean == "" {
			continue
		}
		if strings.HasPrefix(clean, "swarm.") {
			return clean
		}
	}
	return ""
}

func normalizeCouncilMember(member string) string {
	m := strings.ToLower(strings.TrimSpace(member))
	if m == "" {
		return ""
	}
	switch m {
	case "architect", "council architect", "council-architect":
		return "council-architect"
	case "coder", "council coder", "council-coder":
		return "council-coder"
	case "creative", "council creative", "council-creative":
		return "council-creative"
	case "sentry", "council sentry", "council-sentry":
		return "council-sentry"
	default:
		if strings.HasPrefix(m, "council-") {
			return m
		}
		return member
	}
}

func inferCouncilMemberFromInput(input string) string {
	lower := strings.ToLower(input)
	switch {
	case strings.Contains(lower, "architect"):
		return "council-architect"
	case strings.Contains(lower, "coder"), strings.Contains(lower, "code"):
		return "council-coder"
	case strings.Contains(lower, "creative"), strings.Contains(lower, "design"), strings.Contains(lower, "image"):
		return "council-creative"
	case strings.Contains(lower, "sentry"), strings.Contains(lower, "security"), strings.Contains(lower, "risk"):
		return "council-sentry"
	default:
		return ""
	}
}

func shouldCouncilPreflight(toolName string) bool {
	switch strings.TrimSpace(toolName) {
	case "create_team", "delegate_task", "local_command":
		return true
	default:
		return false
	}
}

func councilPreflightMember(toolName string) string {
	switch strings.TrimSpace(toolName) {
	case "create_team", "delegate_task":
		return "council-architect"
	case "local_command":
		return "council-coder"
	default:
		return ""
	}
}

func formatCouncilPreflightQuestion(userInput string, call *toolCallPayload) string {
	if call == nil {
		return userInput
	}
	args := "{}"
	if call.Arguments != nil {
		if b, err := json.Marshal(call.Arguments); err == nil {
			args = string(b)
		}
	}
	return fmt.Sprintf("Preflight review before executing tool '%s' with args %s. User request: %s. Provide concise execution guidance and edge cases.", call.Name, args, strings.TrimSpace(userInput))
}

func (a *Agent) runCouncilPreflight(userMember string, userInput string, call *toolCallPayload) (string, error) {
	if a == nil || a.toolExecutor == nil {
		return "", fmt.Errorf("tool executor unavailable")
	}
	serverID, toolName, err := a.toolExecutor.FindToolByName(a.ctx, "consult_council")
	if err != nil {
		return "", err
	}
	preflightCtx, cancel := context.WithTimeout(a.ctx, 10*time.Second)
	defer cancel()
	args := map[string]any{
		"member":   userMember,
		"question": formatCouncilPreflightQuestion(userInput, call),
	}
	return a.toolExecutor.CallTool(preflightCtx, serverID, toolName, args)
}

func toolCallFingerprint(call *toolCallPayload) string {
	if call == nil {
		return ""
	}
	args := "{}"
	if call.Arguments != nil {
		if b, err := json.Marshal(call.Arguments); err == nil {
			args = string(b)
		}
	}
	return fmt.Sprintf("%s:%s", strings.TrimSpace(call.Name), args)
}

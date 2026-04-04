package swarm

import "strings"

func preferDirectDraftResponse(input string) bool {
	lower := strings.ToLower(strings.TrimSpace(input))
	if lower == "" {
		return false
	}
	informationalLead := []string{"summarize ", "summarise ", "explain ", "describe ", "recap ", "give me a summary", "provide a summary", "walk me through "}
	requestsInformationalAnswer := false
	for _, marker := range informationalLead {
		if strings.HasPrefix(lower, marker) {
			requestsInformationalAnswer = true
			break
		}
	}
	if requestsInformationalAnswer {
		informationalBoundary := []string{"file", "path", "workspace/", "workspace\\", "folder", "directory", "save", "persist", "store", "read ", "open ", "command", "run ", "execute", "shell", "terminal", "team", "agent", "council", "signal", "nats", "api", "http", "url", "image", "diagram"}
		hitsBoundary := false
		for _, token := range informationalBoundary {
			if strings.Contains(lower, token) {
				hitsBoundary = true
				break
			}
		}
		if !hitsBoundary {
			return true
		}
	}
	explicitAction := []string{"file", "path", "workspace", "folder", "directory", "save", "persist", "store", "read ", "open ", "inspect", "command", "run ", "execute", "shell", "terminal", "team", "agent", "council", "signal", "nats", "api", "http", "url", "image", "diagram", "code", "blueprint", "mission", "workflow"}
	for _, token := range explicitAction {
		if strings.Contains(lower, token) {
			return false
		}
	}
	for _, marker := range []string{"write ", "draft ", "compose ", "create ", "make ", "generate "} {
		if strings.Contains(lower, marker) {
			for _, target := range []string{"letter", "email", "message", "note", "reply", "paragraph", "announcement", "bio", "summary", "caption", "introduction"} {
				if strings.Contains(lower, target) {
					return true
				}
			}
			break
		}
	}
	return false
}

func shouldAvoidToolsForDirectDraft(toolName string) bool {
	switch strings.TrimSpace(toolName) {
	case "write_file", "read_file", "local_command", "consult_council", "delegate_task",
		"research_for_blueprint", "generate_blueprint", "load_deployment_context", "remember", "recall",
		"store_artifact", "list_teams", "list_missions", "get_system_status",
		"list_available_tools", "list_catalogue", "publish_signal", "read_signals",
		"broadcast", "create_team":
		return true
	default:
		return false
	}
}

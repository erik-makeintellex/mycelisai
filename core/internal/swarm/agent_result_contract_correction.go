package swarm

import "strings"

func focusedResultContractCorrectionIssues(issues []string) []string {
	for _, issue := range issues {
		if strings.Contains(issue, "missing successful write") ||
			strings.Contains(issue, "no successful retained-output write") ||
			strings.Contains(issue, "missing tool-backed entrypoint") ||
			strings.Contains(issue, "missing tool-backed project package") {
			return []string{issue}
		}
	}
	entrypointIssues := []string{}
	for _, issue := range issues {
		if strings.Contains(issue, "entrypoint") || strings.Contains(issue, "animation loop") {
			entrypointIssues = append(entrypointIssues, issue)
		}
	}
	if len(entrypointIssues) > 0 {
		return entrypointIssues
	}
	for _, issue := range issues {
		if strings.Contains(issue, "readback") {
			return []string{issue}
		}
	}
	if len(issues) > 0 {
		return issues[:1]
	}
	return nil
}

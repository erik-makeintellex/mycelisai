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
	markupIssues := []string{}
	for _, issue := range issues {
		if strings.Contains(issue, "validation target") ||
			strings.Contains(issue, "primary interaction") ||
			strings.Contains(issue, "visible control instructions") ||
			strings.Contains(issue, "text-change observation target") ||
			strings.Contains(issue, "visual-change observation target") {
			markupIssues = append(markupIssues, issue)
		}
	}
	if len(markupIssues) > 0 {
		return markupIssues
	}
	gameplayIssues := []string{}
	for _, issue := range issues {
		if isFunctionalGameCorrectionIssue(issue) {
			gameplayIssues = append(gameplayIssues, issue)
		}
	}
	if len(gameplayIssues) > 0 {
		return gameplayIssues
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

func isFunctionalGameCorrectionIssue(issue string) bool {
	for _, fragment := range []string{
		"game canvas", "game entrypoint", "game-state loop", "game-state model",
		"movement controls", "attack action", "hazard action", "key action", "win action",
		"fail and restart", "animation loop",
	} {
		if strings.Contains(issue, fragment) {
			return true
		}
	}
	return false
}

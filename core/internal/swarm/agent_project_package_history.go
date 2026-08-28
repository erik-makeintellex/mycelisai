package swarm

import (
	"fmt"
	"strings"

	"github.com/mycelis/core/internal/cognitive"
)

const (
	maxProjectPackageLedgerEntries = 24
	maxProjectPackageLedgerPath    = 256
	maxProjectPackageLatestChars   = 1200
)

func projectPackageHistoryEnabled(requirement *teamResultRequirement) bool {
	return requirement != nil && strings.EqualFold(strings.TrimSpace(requirement.Kind), "project_package")
}

func projectPackageInferenceBase(messages []cognitive.ChatMessage, requirement *teamResultRequirement) []cognitive.ChatMessage {
	if !projectPackageHistoryEnabled(requirement) {
		return nil
	}
	base := make([]cognitive.ChatMessage, 0, len(messages))
	for _, message := range messages {
		if message.Role == "system" {
			base = append(base, message)
		}
	}
	for index := len(messages) - 1; index >= 0; index-- {
		if messages[index].Role == "user" {
			base = append(base, messages[index])
			break
		}
	}
	return base
}

func compactProjectPackageInferenceHistory(base []cognitive.ChatMessage, evidence []successfulToolEvidence, latest cognitive.ChatMessage) []cognitive.ChatMessage {
	messages := append([]cognitive.ChatMessage(nil), base...)
	if ledger := compactProjectPackageEvidenceLedger(evidence); ledger != "" {
		messages = append(messages, cognitive.ChatMessage{Role: "system", Content: ledger})
	}
	latest.Content = strings.TrimSpace(latest.Content)
	if len(latest.Content) > maxProjectPackageLatestChars {
		latest.Content = latest.Content[:maxProjectPackageLatestChars] + "..."
	}
	if latest.Content != "" {
		messages = append(messages, latest)
	}
	return messages
}

func compactProjectPackageEvidenceLedger(evidence []successfulToolEvidence) string {
	if len(evidence) == 0 {
		return "Successful project-package evidence ledger: none yet."
	}
	var builder strings.Builder
	builder.WriteString("Successful project-package evidence ledger (paths only; file contents are intentionally omitted):")
	seen := map[string]bool{}
	count := 0
	for _, item := range evidence {
		path := cleanEvidencePath(item.Path)
		if path == "" || seen[item.ToolName+":"+path] {
			continue
		}
		seen[item.ToolName+":"+path] = true
		if len(path) > maxProjectPackageLedgerPath {
			path = path[:maxProjectPackageLedgerPath] + "..."
		}
		builder.WriteString(fmt.Sprintf("\n- tool=%s path=%s", strings.TrimSpace(item.ToolName), path))
		count++
		if count == maxProjectPackageLedgerEntries {
			break
		}
	}
	return builder.String()
}

func compactProjectPackageToolOutcome(toolName string, succeeded bool) cognitive.ChatMessage {
	state := "failed and needs a different bounded correction"
	if succeeded {
		state = "completed; continue with the next missing contract evidence"
	}
	return cognitive.ChatMessage{Role: "user", Content: fmt.Sprintf("Latest tool outcome: tool=%s state=%s. Do not repeat completed evidence.", strings.TrimSpace(toolName), state)}
}

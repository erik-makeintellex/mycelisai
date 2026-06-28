package server

import (
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

func shouldUseRequestedWriteFilePlan(latestRequest string, fileCall protocol.PlannedToolCall) bool {
	if requestHasExplicitWriteFileContent(latestRequest) {
		return true
	}
	path := firstNonEmptyString(fileCall.Arguments["path"])
	if filepathExt(path) == "" {
		return false
	}
	lower := strings.ToLower(strings.TrimSpace(latestRequest))
	if strings.Contains(lower, "project_package") || strings.Contains(lower, "project package") {
		return false
	}
	return requestContainsAny(lower, []string{
		"write", "create", "save", "register", "report", "reaction", "markdown", "md file", "retained output",
	})
}

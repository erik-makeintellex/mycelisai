package swarm

import (
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

func evidencePaths(evidence []successfulToolEvidence, toolName string) []string {
	paths := []string{}
	for _, item := range evidence {
		if item.ToolName == toolName && item.Path != "" {
			paths = append(paths, item.Path)
		}
	}
	return uniqueResultContractStrings(paths)
}

func evidenceHasTool(evidence []successfulToolEvidence, toolName string) bool {
	for _, item := range evidence {
		if item.ToolName == toolName {
			return true
		}
	}
	return false
}

func cleanEvidencePath(value string) string {
	return strings.Trim(strings.ReplaceAll(strings.TrimSpace(value), "\\", "/"), "/")
}

func comparableEvidencePath(value string) string {
	return strings.ToLower(cleanEvidencePath(value))
}

func evidenceContainsPath(paths []string, candidate string) bool {
	return matchingEvidencePath(paths, candidate) != ""
}

func matchingEvidencePath(paths []string, candidate string) string {
	candidateComparable := comparableEvidencePath(candidate)
	for _, path := range paths {
		pathComparable := comparableEvidencePath(path)
		if pathComparable == candidateComparable || strings.HasSuffix(pathComparable, "/"+candidateComparable) || strings.HasSuffix(candidateComparable, "/"+pathComparable) {
			return cleanEvidencePath(path)
		}
	}
	return ""
}

func hasReadbackEvidence(writes, reads []string) bool {
	for _, read := range reads {
		if evidenceContainsPath(writes, read) {
			return true
		}
	}
	return false
}

func firstHTMLWrite(writes []string) string {
	for _, path := range writes {
		if strings.HasSuffix(strings.ToLower(path), ".html") {
			return path
		}
	}
	return ""
}

func pathWithinFolder(path, folder string) bool {
	return strings.HasPrefix(comparableEvidencePath(path), strings.TrimSuffix(comparableEvidencePath(folder), "/")+"/")
}

func relativeEvidencePath(path, folder string) string {
	path = cleanEvidencePath(path)
	prefix := strings.TrimSuffix(cleanEvidencePath(folder), "/") + "/"
	if len(path) >= len(prefix) && strings.EqualFold(path[:len(prefix)], prefix) {
		return path[len(prefix):]
	}
	return path
}

func hasProjectPackageArtifact(artifacts []protocol.ChatArtifactRef) bool {
	return firstProjectPackageArtifact(artifacts) != nil
}

func firstProjectPackageArtifact(artifacts []protocol.ChatArtifactRef) *protocol.ChatArtifactRef {
	for index := range artifacts {
		if strings.EqualFold(strings.TrimSpace(artifacts[index].Type), "project_package") {
			return &artifacts[index]
		}
	}
	return nil
}

func uniqueResultContractStrings(items []string) []string {
	seen := map[string]struct{}{}
	unique := make([]string, 0, len(items))
	for _, item := range items {
		if _, ok := seen[item]; ok || strings.TrimSpace(item) == "" {
			continue
		}
		seen[item] = struct{}{}
		unique = append(unique, item)
	}
	return unique
}

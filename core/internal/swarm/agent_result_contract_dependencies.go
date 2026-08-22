package swarm

import (
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

var resultContractLocalAssetRefPattern = regexp.MustCompile(`(?i)(?:src|href)\s*=\s*["']([^"']+)["']`)

func resultContractLocalDependencyIssues(artifacts []protocol.ChatArtifactRef, evidence []successfulToolEvidence) []string {
	artifact := firstProjectPackageArtifact(artifacts)
	if artifact == nil || strings.TrimSpace(artifact.Entrypoint) == "" || strings.TrimSpace(artifact.Folder) == "" {
		return nil
	}
	entrypoint := cleanEvidencePath(artifact.Entrypoint)
	content := latestEntrypointEvidenceContent(evidence, entrypoint)
	if strings.TrimSpace(content) == "" {
		return nil
	}
	writes := evidencePaths(evidence, "write_file")
	issues := []string{}
	for _, match := range resultContractLocalAssetRefPattern.FindAllStringSubmatch(content, -1) {
		assetRef := strings.TrimSpace(match[1])
		if !isLocalResultContractAsset(assetRef) {
			continue
		}
		if cut := strings.IndexAny(assetRef, "?#"); cut >= 0 {
			assetRef = assetRef[:cut]
		}
		assetPath := cleanEvidencePath(path.Join(path.Dir(entrypoint), assetRef))
		if !pathWithinFolder(assetPath, artifact.Folder) {
			issues = append(issues, "local dependency escapes the retained package folder: "+assetRef)
			continue
		}
		if !evidenceContainsPath(writes, assetPath) {
			issues = append(issues, "missing successful write for local dependency "+assetRef)
		}
	}
	return uniqueResultContractStrings(issues)
}

func resultContractPackageValidationContent(artifacts []protocol.ChatArtifactRef, evidence []successfulToolEvidence) string {
	artifact := firstProjectPackageArtifact(artifacts)
	if artifact == nil || strings.TrimSpace(artifact.Entrypoint) == "" {
		return ""
	}
	entrypoint := cleanEvidencePath(artifact.Entrypoint)
	content := latestEntrypointEvidenceContent(evidence, entrypoint)
	for _, dependency := range resultContractLocalDependencyPaths(artifact, evidence) {
		if !strings.EqualFold(path.Ext(dependency), ".js") {
			continue
		}
		if dependencyContent := latestSuccessfulWriteContent(evidence, dependency); dependencyContent != "" {
			content += "\n" + dependencyContent
		}
	}
	return content
}

func resultContractLocalDependencyPaths(artifact *protocol.ChatArtifactRef, evidence []successfulToolEvidence) []string {
	if artifact == nil {
		return nil
	}
	entrypoint := cleanEvidencePath(artifact.Entrypoint)
	content := latestEntrypointEvidenceContent(evidence, entrypoint)
	dependencies := []string{}
	for _, match := range resultContractLocalAssetRefPattern.FindAllStringSubmatch(content, -1) {
		assetRef := strings.TrimSpace(match[1])
		if !isLocalResultContractAsset(assetRef) {
			continue
		}
		if cut := strings.IndexAny(assetRef, "?#"); cut >= 0 {
			assetRef = assetRef[:cut]
		}
		dependencies = append(dependencies, cleanEvidencePath(path.Join(path.Dir(entrypoint), assetRef)))
	}
	return uniqueResultContractStrings(dependencies)
}

func latestSuccessfulWriteContent(evidence []successfulToolEvidence, candidate string) string {
	content := ""
	for _, item := range evidence {
		if item.ToolName == "write_file" && evidenceContainsPath([]string{item.Path}, candidate) {
			content = item.Content
		}
	}
	return content
}

func isLocalResultContractAsset(value string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "//") {
		return false
	}
	parsed, err := url.Parse(trimmed)
	return err == nil && parsed.Scheme == "" && parsed.Host == ""
}

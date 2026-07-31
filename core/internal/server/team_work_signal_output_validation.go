package server

import (
	"html"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

var (
	localHTMLAssetRefPattern      = regexp.MustCompile(`(?i)(?:src|href)\s*=\s*["']([^"']+)["']`)
	interactiveHandlerPattern     = regexp.MustCompile(`(?i)(?:addEventListener\s*\(\s*["'](?:click|pointerdown|touchstart|keydown|keyup)|on(?:click|pointerdown|touchstart|keydown)\s*=)`)
	actionableButtonPattern       = regexp.MustCompile(`(?is)<button\b[^>]*>(.*?)</button>`)
	visibleControlLanguagePattern = regexp.MustCompile(`(?i)\b(?:click|tap|press|use|move|drag|select|arrow|space|wasd|control)\b`)
	scriptOrStylePattern          = regexp.MustCompile(`(?is)<(?:script|style)\b[^>]*>.*?</(?:script|style)>`)
	htmlTagPattern                = regexp.MustCompile(`(?s)<[^>]+>`)
)

func deliverableResultMissingOutputs(item protocol.TeamWorkItem, payloadKind protocol.SignalPayloadKind, outputRefs []protocol.TeamOutputRef) bool {
	return deliverableResultOutputIssue(item, payloadKind, outputRefs) != ""
}

func deliverableResultOutputIssue(item protocol.TeamWorkItem, payloadKind protocol.SignalPayloadKind, outputRefs []protocol.TeamOutputRef) string {
	if payloadKind != protocol.PayloadKindResult ||
		(item.ExecutionShape != protocol.TeamExecutionShapeDeliverable && item.ExecutionShape != protocol.TeamExecutionShapeDelegatedWork) ||
		len(item.ExpectedOutputs) == 0 {
		return ""
	}
	if len(outputRefs) == 0 {
		return "missing_retained_output"
	}
	if !teamWorkExpectsProjectPackage(item) {
		return ""
	}
	teamRoot := "groups/" + strings.Trim(strings.TrimSpace(item.TeamID), "/") + "/"
	for _, ref := range outputRefs {
		storageRef := strings.Trim(strings.ReplaceAll(strings.TrimSpace(ref.StorageRef), "\\", "/"), "/")
		if strings.EqualFold(strings.TrimSpace(ref.Kind), "project_package") &&
			strings.TrimSpace(ref.Entrypoint) != "" &&
			strings.HasPrefix(storageRef+"/", teamRoot) {
			return projectPackageFileIssue(item, ref)
		}
	}
	return "invalid_deliverable_shape"
}

func projectPackageFileIssue(item protocol.TeamWorkItem, ref protocol.TeamOutputRef) string {
	storageRef := strings.Trim(strings.ReplaceAll(strings.TrimSpace(ref.StorageRef), "\\", "/"), "/")
	entrypoint := strings.Trim(strings.ReplaceAll(strings.TrimSpace(ref.Entrypoint), "\\", "/"), "/")
	entryPath := entrypoint
	if !strings.HasPrefix(entrypoint+"/", storageRef+"/") {
		entryPath = storageRef + "/" + entrypoint
	}
	target, _, err := resolveWorkspaceFilePath(entryPath)
	if err != nil {
		return "invalid_deliverable_shape"
	}
	info, err := os.Stat(target)
	if err != nil || info.IsDir() {
		return "incomplete_deliverable_files"
	}
	if !strings.EqualFold(filepath.Ext(target), ".html") {
		return ""
	}
	content, err := os.ReadFile(target)
	if err != nil || len(content) > maxWorkspaceViewBytes {
		return "incomplete_deliverable_files"
	}
	packageFolder, _, err := resolveWorkspacePath(storageRef, false)
	if err != nil {
		return "invalid_deliverable_shape"
	}
	combinedExecutableContent := string(content)
	for _, match := range localHTMLAssetRefPattern.FindAllStringSubmatch(string(content), -1) {
		assetRef := strings.TrimSpace(match[1])
		lower := strings.ToLower(assetRef)
		if assetRef == "" || strings.HasPrefix(assetRef, "#") || strings.HasPrefix(lower, "data:") ||
			strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") || strings.HasPrefix(lower, "//") {
			continue
		}
		if cut := strings.IndexAny(assetRef, "?#"); cut >= 0 {
			assetRef = assetRef[:cut]
		}
		assetTarget := filepath.Clean(filepath.Join(filepath.Dir(target), filepath.FromSlash(assetRef)))
		rel, relErr := filepath.Rel(packageFolder, assetTarget)
		if relErr != nil || pathEscapesWorkspace(rel) {
			return "incomplete_deliverable_files"
		}
		assetInfo, statErr := os.Stat(assetTarget)
		if statErr != nil || assetInfo.IsDir() {
			return "incomplete_deliverable_files"
		}
		if strings.EqualFold(filepath.Ext(assetTarget), ".js") {
			assetContent, readErr := os.ReadFile(assetTarget)
			if readErr != nil || len(assetContent) > maxWorkspaceViewBytes {
				return "incomplete_deliverable_files"
			}
			combinedExecutableContent += "\n" + string(assetContent)
		}
	}
	if teamWorkRequiresPrimaryInteraction(item) &&
		(!interactiveHandlerPattern.MatchString(combinedExecutableContent) || !htmlExposesPrimaryControl(string(content))) {
		return "unverified_primary_interaction"
	}
	return ""
}

func htmlExposesPrimaryControl(content string) bool {
	for _, match := range actionableButtonPattern.FindAllStringSubmatch(content, -1) {
		label := strings.ToLower(strings.TrimSpace(html.UnescapeString(htmlTagPattern.ReplaceAllString(match[1], " "))))
		if label != "" && !strings.Contains(label, "restart") && !strings.Contains(label, "reset") {
			return true
		}
	}
	visibleText := scriptOrStylePattern.ReplaceAllString(content, " ")
	visibleText = html.UnescapeString(htmlTagPattern.ReplaceAllString(visibleText, " "))
	return visibleControlLanguagePattern.MatchString(visibleText)
}

func teamWorkRequiresPrimaryInteraction(item protocol.TeamWorkItem) bool {
	values := append([]string{item.Objective}, item.ExpectedOutputs...)
	if item.WorkIntent != nil {
		values = append(values, item.WorkIntent.Objective)
		if item.WorkIntent.OutputContract != nil {
			values = append(values, item.WorkIntent.OutputContract.Validation...)
		}
	}
	for _, value := range values {
		lower := strings.ToLower(value)
		if strings.Contains(lower, "playable") || strings.Contains(lower, "browser game") ||
			strings.Contains(lower, "controls respond") || strings.Contains(lower, "primary user workflow") {
			return true
		}
	}
	return false
}

func teamWorkExpectsProjectPackage(item protocol.TeamWorkItem) bool {
	if item.WorkIntent != nil && item.WorkIntent.OutputContract != nil &&
		strings.EqualFold(strings.TrimSpace(item.WorkIntent.OutputContract.Shape), "app_package") {
		return true
	}
	for _, expected := range item.ExpectedOutputs {
		lower := strings.ToLower(expected)
		if strings.Contains(lower, "project package") || strings.Contains(lower, "application package") ||
			strings.Contains(lower, "browser game") || strings.Contains(lower, "playable app") {
			return true
		}
	}
	return false
}

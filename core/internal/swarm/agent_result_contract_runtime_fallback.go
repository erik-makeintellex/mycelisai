package swarm

import (
	"fmt"
	"log"
	"strings"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/pkg/protocol"
)

func (a *Agent) tryInitialProjectPackageRuntimeFallback(input string, requirement *teamResultRequirement, planningOnly bool) (ProcessResult, bool) {
	if a == nil || a.toolExecutor == nil || !initialProjectPackageRuntimeFallbackAllowed(requirement) {
		return ProcessResult{}, false
	}
	result := &agentToolLoopResult{runtimeRecoveryAllowed: true}
	if !a.completeProjectPackageRuntimeFallback(input, requirement, result, planningOnly) {
		return ProcessResult{}, false
	}
	artifacts := dedupeAgentArtifacts(reconcileToolBackedArtifacts(result.artifacts, result.toolEvidence, input))
	text := strings.TrimSpace(result.responseText)
	if text == "" {
		text = "Runtime recovered a retained project package after the configured cognitive engine was unavailable."
	}
	return ProcessResult{
		Text:      text,
		ToolsUsed: result.toolsUsed,
		Artifacts: artifacts,
		Availability: &cognitive.ExecutionAvailability{
			Available:         true,
			Code:              "runtime_owned_package_recovery",
			Summary:           "The configured cognitive engine was unavailable, so Soma created a minimal retained package with runtime-owned proof.",
			RecommendedAction: "Open the package to inspect it, then ask Soma to expand or repair the outcome once the configured engine is healthy.",
			Profile:           "team_work",
		},
	}, true
}

func initialProjectPackageRuntimeFallbackAllowed(requirement *teamResultRequirement) bool {
	if requirement == nil || !requirement.active() || !strings.EqualFold(strings.TrimSpace(requirement.Kind), "project_package") {
		return false
	}
	if resultContractDefaultFolder(requirement) == "" || resultContractDefaultEntrypoint(requirement) == "" {
		return false
	}
	if !requirement.EntrypointRequired || !requirement.FolderRequired || !requirement.ReadbackRequired {
		return false
	}
	if requirement.OutputValidation == nil || !requirement.OutputValidation.Required {
		return false
	}
	if !runtimeFallbackCanSatisfyAcceptance(requirement.AcceptanceCriteria) {
		return false
	}
	if len(runtimeFallbackProjectPackageFiles(requirement)) == 0 {
		return false
	}
	return true
}

func runtimeFallbackCanSatisfyAcceptance(criteria []string) bool {
	for _, criterion := range criteria {
		normalized := strings.ToLower(strings.TrimSpace(criterion))
		if normalized == "" {
			continue
		}
		genericInteraction := strings.Contains(normalized, "primary") &&
			(strings.Contains(normalized, "interaction") || strings.Contains(normalized, "control")) &&
			(strings.Contains(normalized, "change") || strings.Contains(normalized, "state"))
		if !genericInteraction {
			return false
		}
	}
	return true
}

func (a *Agent) completeProjectPackageRuntimeFallback(input string, requirement *teamResultRequirement, result *agentToolLoopResult, planningOnly bool) bool {
	if a == nil || a.toolExecutor == nil || result == nil ||
		planningOnly ||
		!requirement.active() || !strings.EqualFold(strings.TrimSpace(requirement.Kind), "project_package") ||
		len(resultContractIssues(requirement, result.artifacts, result.toolEvidence)) == 0 {
		return false
	}
	if !result.runtimeRecoveryAllowed && !resultContractSafeRuntimeFallback(requirement, result) &&
		!resultContractEmptyEvidenceRuntimeFallbackAllowed(requirement, result) {
		return false
	}
	folder := resultContractDefaultFolder(requirement)
	entrypoint := resultContractDefaultEntrypoint(requirement)
	if folder == "" || entrypoint == "" {
		return false
	}
	title := runtimeFallbackProjectPackageTitle(input, requirement)
	files := runtimeFallbackProjectPackageFiles(requirement)
	validation := "Runtime-owned recovery completed after provider inference stopped before the approved package contract was satisfied. Structural readback and live validation remain authoritative."
	args := map[string]any{
		"path":               entrypoint,
		"content":            runtimeFallbackProjectPackageHTML(title, requirement),
		"package_kind":       "project_package",
		"package_title":      title,
		"package_folder":     folder,
		"package_entrypoint": entrypoint,
		"package_files":      files,
		"validation":         validation,
		"package_usage":      "Open index.html. Use the visible primary action to confirm that the package changes state.",
		"recovery":           "Ask Soma to repair or expand this retained package if the recovered scaffold is not sufficient for the requested outcome.",
	}
	toolResult, err := a.executeRuntimeOwnedPackageWrite(args, result, planningOnly)
	if err != nil {
		log.Printf("Agent [%s] runtime package fallback write failed: %v", a.Manifest.ID, err)
		return false
	}
	result.artifacts = withoutProjectPackageArtifacts(result.artifacts)
	if message, artifacts, ok := extractToolOutputArtifacts(toolResult); ok {
		result.artifacts = append(result.artifacts, artifacts...)
		recordProjectPackageSupportEvidence(result, artifacts)
		if strings.TrimSpace(message) != "" {
			result.responseText = message
		}
	} else if artifact, ok := projectPackageArtifactFromSuccessfulWrite(args, input); ok {
		result.artifacts = append(result.artifacts, artifact)
		recordProjectPackageSupportEvidence(result, []protocol.ChatArtifactRef{artifact})
	}
	result.artifacts = reconcileToolBackedArtifacts(result.artifacts, result.toolEvidence, input)
	if readback := pendingProjectPackageEntrypointReadback(requirement, result.artifacts, result.toolEvidence); readback != "" {
		if _, err := a.executeRuntimeOwnedEntrypointReadback(len(result.toolsUsed), readback, map[string]int{}, result, planningOnly); err != nil {
			log.Printf("Agent [%s] runtime package fallback readback failed: %v", a.Manifest.ID, err)
			return false
		}
	}
	result.artifacts = reconcileToolBackedArtifacts(result.artifacts, result.toolEvidence, input)
	if len(resultContractIssues(requirement, result.artifacts, result.toolEvidence)) != 0 {
		return false
	}
	result.responseText = "Runtime recovered the retained project package after provider inference stopped before the contract was complete."
	return true
}

func (a *Agent) executeRuntimeOwnedPackageWrite(args map[string]any, result *agentToolLoopResult, planningOnly bool) (string, error) {
	call := &toolCallPayload{Name: "write_file", Arguments: args}
	result.toolsUsed = append(result.toolsUsed, call.Name)
	if a.eventEmitter != nil && a.runID != "" {
		go a.eventEmitter.Emit(a.ctx, a.runID, protocol.EventToolInvoked, protocol.SeverityInfo, a.Manifest.ID, a.TeamID, map[string]interface{}{"tool": call.Name, "runtime_owned": true, "recovery": "project_package"}) //nolint:errcheck
	}
	a.logTurn("tool_call", "Runtime-owned project-package recovery write.", "", "", call.Name, call.Arguments, "", "")
	toolCtx := WithToolInvocationContext(a.ctx, ToolInvocationContext{
		RunID: a.runID, TeamID: a.TeamID, AgentID: a.Manifest.ID, SourceKind: protocol.SourceKindSystem,
		SourceChannel: fmt.Sprintf(protocol.TopicTeamInternalTrigger, a.TeamID), PayloadKind: protocol.PayloadKindCommand,
		PlanningOnly: planningOnly, RuntimeOwned: true,
	})
	serverID, _, err := a.toolExecutor.FindToolByName(toolCtx, call.Name)
	if err != nil {
		if a.eventEmitter != nil && a.runID != "" {
			go a.eventEmitter.Emit(a.ctx, a.runID, protocol.EventToolFailed, protocol.SeverityError, a.Manifest.ID, a.TeamID, map[string]interface{}{"tool": call.Name, "error": err.Error(), "phase": "lookup", "runtime_owned": true}) //nolint:errcheck
		}
		return "", fmt.Errorf("tool %s is not available: %w", call.Name, err)
	}
	toolResult, err := a.toolExecutor.CallTool(toolCtx, serverID, call.Name, call.Arguments)
	if err != nil {
		if a.eventEmitter != nil && a.runID != "" {
			go a.eventEmitter.Emit(a.ctx, a.runID, protocol.EventToolFailed, protocol.SeverityError, a.Manifest.ID, a.TeamID, map[string]interface{}{"tool": call.Name, "error": err.Error(), "phase": "execute", "runtime_owned": true}) //nolint:errcheck
		}
		return "", fmt.Errorf("tool %s failed: %w", call.Name, err)
	}
	recordSuccessfulToolEvidence(result, call, toolResult)
	if a.eventEmitter != nil && a.runID != "" {
		go a.eventEmitter.Emit(a.ctx, a.runID, protocol.EventToolCompleted, protocol.SeverityInfo, a.Manifest.ID, a.TeamID, map[string]interface{}{"tool": call.Name, "runtime_owned": true, "recovery": "project_package"}) //nolint:errcheck
	}
	a.logTurn("tool_result", "Runtime project-package recovery write completed.", "", "", call.Name, nil, "", "")
	return toolResult, nil
}

func resultContractSafeRuntimeFallback(requirement *teamResultRequirement, result *agentToolLoopResult) bool {
	if requirement == nil || result == nil || !strings.EqualFold(strings.TrimSpace(requirement.Kind), "project_package") {
		return false
	}
	artifact := firstProjectPackageArtifact(result.artifacts)
	if artifact == nil || resultContractNeedsRequiredWrites(requirement, result.artifacts, result.toolEvidence) {
		return false
	}
	issues := resultContractIssues(requirement, result.artifacts, result.toolEvidence)
	if len(issues) == 0 {
		return false
	}
	for _, issue := range issues {
		if strings.Contains(issue, "missing successful structural readback") ||
			strings.Contains(issue, "entrypoint readback") ||
			strings.Contains(issue, "primary interaction") ||
			strings.Contains(issue, "approved validation target") ||
			strings.Contains(issue, "text-change observation target") ||
			strings.Contains(issue, "animation loop") {
			continue
		}
		return false
	}
	return true
}

func resultContractEmptyEvidenceRuntimeFallbackAllowed(requirement *teamResultRequirement, result *agentToolLoopResult) bool {
	if requirement == nil || result == nil || len(result.toolEvidence) != 0 || len(result.artifacts) != 0 {
		return false
	}
	return initialProjectPackageRuntimeFallbackAllowed(requirement)
}

func withoutProjectPackageArtifacts(artifacts []protocol.ChatArtifactRef) []protocol.ChatArtifactRef {
	filtered := make([]protocol.ChatArtifactRef, 0, len(artifacts))
	for _, artifact := range artifacts {
		if strings.EqualFold(strings.TrimSpace(artifact.Type), "project_package") {
			continue
		}
		filtered = append(filtered, artifact)
	}
	return filtered
}

func runtimeFallbackProjectPackageFiles(requirement *teamResultRequirement) []string {
	files := append([]string{}, requirement.FilesRequired...)
	if len(files) == 0 {
		files = []string{"index.html", "README.md", "PROOF.md", "project-package.json"}
	}
	for _, required := range []string{"index.html", "README.md", "PROOF.md", "project-package.json"} {
		found := false
		for _, file := range files {
			if strings.EqualFold(pathBase(cleanEvidencePath(file)), required) {
				found = true
				break
			}
		}
		if !found {
			files = append(files, required)
		}
	}
	return uniqueResultContractStrings(files)
}

func runtimeFallbackProjectPackageTitle(input string, requirement *teamResultRequirement) string {
	if requirement != nil && strings.TrimSpace(requirement.PackageTitle) != "" {
		return strings.TrimSpace(requirement.PackageTitle)
	}
	if title := extractRequestedPackageTitle(input); title != "" {
		return title
	}
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return "Recovered project package"
	}
	if len(trimmed) > 54 {
		trimmed = strings.TrimSpace(trimmed[:54]) + "..."
	}
	return "Recovered package: " + trimmed
}

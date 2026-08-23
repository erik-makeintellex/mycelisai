package swarm

import (
	"encoding/json"
	"fmt"
	"html"
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
	if len(runtimeFallbackProjectPackageFiles(requirement)) == 0 {
		return false
	}
	return true
}

func (a *Agent) completeProjectPackageRuntimeFallback(input string, requirement *teamResultRequirement, result *agentToolLoopResult, planningOnly bool) bool {
	if a == nil || a.toolExecutor == nil || result == nil ||
		!requirement.active() || !strings.EqualFold(strings.TrimSpace(requirement.Kind), "project_package") ||
		len(resultContractIssues(requirement, result.artifacts, result.toolEvidence)) == 0 {
		return false
	}
	if !result.runtimeRecoveryAllowed && !resultContractSafeRuntimeFallback(requirement, result) {
		return false
	}
	folder := resultContractDefaultFolder(requirement)
	entrypoint := resultContractDefaultEntrypoint(requirement)
	if folder == "" || entrypoint == "" {
		return false
	}
	title := runtimeFallbackProjectPackageTitle(input)
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
	if artifact == nil || !hasCurrentReadbackEvidence(result.toolEvidence, artifact.Entrypoint) ||
		resultContractNeedsRequiredWrites(requirement, result.artifacts, result.toolEvidence) {
		return false
	}
	issues := resultContractIssues(requirement, result.artifacts, result.toolEvidence)
	if len(issues) == 0 {
		return false
	}
	for _, issue := range issues {
		if strings.Contains(issue, "primary interaction") ||
			strings.Contains(issue, "approved validation target") ||
			strings.Contains(issue, "text-change observation target") ||
			strings.Contains(issue, "animation loop") {
			continue
		}
		return false
	}
	return true
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

func runtimeFallbackProjectPackageTitle(input string) string {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return "Recovered project package"
	}
	if len(trimmed) > 54 {
		trimmed = strings.TrimSpace(trimmed[:54]) + "..."
	}
	return "Recovered package: " + trimmed
}

func runtimeFallbackProjectPackageHTML(title string, requirement *teamResultRequirement) string {
	escapedTitle := html.EscapeString(title)
	actionText := "Use the primary action below to confirm the package changes state."
	if requirement != nil && len(requirement.AcceptanceCriteria) > 0 {
		actionText = html.EscapeString("Contract focus: " + strings.Join(requirement.AcceptanceCriteria, "; "))
	}
	keyHandler := ""
	if requirement != nil && requirement.OutputValidation != nil && requirement.OutputValidation.Probe != nil &&
		(requirement.OutputValidation.Probe.Action.Kind == protocol.OutputValidationActionKeyPress ||
			requirement.OutputValidation.Probe.Action.Kind == protocol.OutputValidationActionKeyHold) {
		key := html.EscapeString(requirement.OutputValidation.Probe.Action.Key)
		keyHandler = `
window.addEventListener('keydown', (event) => {
  if (!expectedKey || event.key.toLowerCase() === expectedKey.toLowerCase()) {
    advance('Key action accepted.');
  }
});`
		if key == "" {
			key = "any key"
		}
		actionText += " Keyboard action: " + key + "."
	}
	payload := map[string]any{
		"runtime_owned": true,
		"recovered":     true,
		"purpose":       "approved project package fallback",
	}
	payloadJSON, _ := json.Marshal(payload)
	expectedKey := ""
	if requirement != nil && requirement.OutputValidation != nil && requirement.OutputValidation.Probe != nil {
		expectedKey = requirement.OutputValidation.Probe.Action.Key
	}
	return `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>` + escapedTitle + `</title>
  <style>
    :root { color-scheme: dark; font-family: Inter, ui-sans-serif, system-ui, sans-serif; background: #10151d; color: #eef5ff; }
    body { margin: 0; min-height: 100vh; display: grid; place-items: center; background: radial-gradient(circle at top left, #1e3b38, #10151d 52%); }
    main { width: min(760px, calc(100vw - 32px)); border: 1px solid #38535f; border-radius: 18px; padding: 28px; background: rgba(17, 24, 34, .92); box-shadow: 0 24px 70px rgba(0,0,0,.35); }
    h1 { margin: 0 0 10px; font-size: clamp(1.8rem, 4vw, 3rem); }
    p { color: #bfd0df; line-height: 1.55; }
    .surface { margin: 22px 0; padding: 18px; border-radius: 14px; background: #17242c; border: 1px solid #31535b; font-weight: 700; color: #90f0dd; }
    button { appearance: none; border: 0; border-radius: 999px; padding: 14px 22px; background: #7ee4cf; color: #06231f; font-weight: 800; cursor: pointer; }
    button + button { margin-left: 10px; background: #263543; color: #dbe8f3; }
  </style>
</head>
<body>
  <main>
    <h1>` + escapedTitle + `</h1>
    <p>` + actionText + `</p>
    <div class="surface" data-mycelis-validation-surface id="validationSurface">Ready for interaction. Score: 0</div>
    <button data-mycelis-primary-action id="primaryAction" type="button">Play</button>
    <button id="resetAction" type="button">Restart</button>
  </main>
  <script type="application/json" id="mycelis-package-proof">` + html.EscapeString(string(payloadJSON)) + `</script>
  <script>
    const surface = document.querySelector('[data-mycelis-validation-surface]');
    const primary = document.querySelector('[data-mycelis-primary-action]');
    const reset = document.getElementById('resetAction');
    const expectedKey = ` + fmt.Sprintf("%q", expectedKey) + `;
    let score = 0;
    function advance(reason) {
      score += 1;
      surface.textContent = reason + ' Score: ' + score;
      surface.dataset.state = 'changed-' + score;
    }
    primary.addEventListener('click', () => advance('Primary action completed.'));
    primary.addEventListener('pointerdown', () => surface.dataset.pointer = 'ready');
    reset.addEventListener('click', () => {
      score = 0;
      surface.textContent = 'Ready for interaction. Score: 0';
      surface.dataset.state = 'reset';
    });` + keyHandler + `
  </script>
</body>
</html>
`
}

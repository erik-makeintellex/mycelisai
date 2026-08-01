package swarm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

var (
	resultContractInteractiveHandlerPattern = regexp.MustCompile(`(?i)(?:addEventListener\s*\(\s*["'](?:click|pointerdown|touchstart|keydown|keyup)|on(?:click|pointerdown|touchstart|keydown)\s*=)`)
	resultContractVisibleControlPattern     = regexp.MustCompile(`(?i)\b(?:click|tap|press|use|move|drag|select|arrow|space|wasd|control)\b`)
	resultContractScriptOrStylePattern      = regexp.MustCompile(`(?is)<(?:script|style)\b[^>]*>.*?</(?:script|style)>`)
	resultContractHTMLTagPattern            = regexp.MustCompile(`(?s)<[^>]+>`)
)

const (
	maxResultContractCorrections    = 2
	maxResultContractToolIterations = 14
)

type teamResultRequirement struct {
	Kind               string
	FilesRequired      []string
	ExpectedOutputs    []string
	AcceptanceCriteria []string
	ProofRequirements  []string
	EntrypointRequired bool
	FolderRequired     bool
	ReadbackRequired   bool
	DownstreamProofRef bool
	RepairChannel      string
}

type successfulToolEvidence struct {
	ToolName string
	Path     string
	Content  string
}

func renderTeamAskContractList(sb *strings.Builder, label string, items []string) {
	if len(items) == 0 {
		return
	}
	sb.WriteString(label + ":\n")
	for _, item := range items {
		sb.WriteString(fmt.Sprintf("- %s\n", item))
	}
}

func teamResultRequirementFromTrigger(data []byte, planningOnly bool) *teamResultRequirement {
	if planningOnly {
		return nil
	}
	var ask protocol.TeamAsk
	if err := json.Unmarshal(bytes.TrimSpace(data), &ask); err != nil || ask.IsZero() {
		return nil
	}
	contract, ok := ask.Context["result_contract"].(map[string]any)
	if !ok || len(contract) == 0 {
		return nil
	}
	requirement := &teamResultRequirement{
		Kind:               strings.TrimSpace(stringValue(contract["kind"])),
		FilesRequired:      stringSlice(contract["files_required"]),
		ExpectedOutputs:    stringSlice(contract["expected_outputs"]),
		AcceptanceCriteria: stringSlice(contract["acceptance_criteria"]),
		ProofRequirements:  stringSlice(contract["proof_required"]),
		EntrypointRequired: boolValue(contract["entrypoint_required"]),
		FolderRequired:     boolValue(contract["folder_required"]),
		ReadbackRequired:   boolValue(contract["validation_required"]),
		DownstreamProofRef: boolValue(contract["proof_ref_required"]),
		RepairChannel:      strings.TrimSpace(stringValue(contract["repair_channel"])),
	}
	if len(requirement.AcceptanceCriteria) == 0 {
		requirement.AcceptanceCriteria = append([]string(nil), ask.ExitCriteria...)
	}
	if len(requirement.ProofRequirements) == 0 {
		requirement.ProofRequirements = append([]string(nil), ask.EvidenceRequired...)
	}
	if !requirement.active() {
		return nil
	}
	return requirement
}

func (requirement *teamResultRequirement) active() bool {
	return requirement != nil && (requirement.Kind != "" || len(requirement.FilesRequired) > 0 ||
		len(requirement.ExpectedOutputs) > 0 || len(requirement.AcceptanceCriteria) > 0 ||
		len(requirement.ProofRequirements) > 0 || requirement.EntrypointRequired ||
		requirement.FolderRequired || requirement.ReadbackRequired || requirement.DownstreamProofRef)
}

func resultContractLoopLimit(current int, requirement *teamResultRequirement) int {
	if !requirement.active() {
		return current
	}
	if current >= maxResultContractToolIterations {
		return current
	}
	return maxResultContractToolIterations
}

func resultContractEvidenceToolAllowed(requirement *teamResultRequirement, toolName string) bool {
	if !requirement.active() || !strings.EqualFold(requirement.Kind, "project_package") {
		return true
	}
	switch strings.TrimSpace(toolName) {
	case "write_file", "read_file", "read_text_file":
		return true
	default:
		return false
	}
}

func recordSuccessfulToolEvidence(result *agentToolLoopResult, call *toolCallPayload, toolResult string) {
	if result == nil || call == nil {
		return
	}
	evidence := successfulToolEvidence{ToolName: strings.TrimSpace(call.Name)}
	switch evidence.ToolName {
	case "write_file", "read_file", "read_text_file":
		evidence.Path = cleanEvidencePath(stringValue(call.Arguments["path"]))
	}
	if evidence.ToolName == "write_file" {
		evidence.Content = stringValue(call.Arguments["content"])
	} else if evidence.ToolName == "read_file" || evidence.ToolName == "read_text_file" {
		evidence.Content = toolResult
	}
	result.toolEvidence = append(result.toolEvidence, evidence)
}

func reconcileToolBackedArtifacts(artifacts []protocol.ChatArtifactRef, evidence []successfulToolEvidence, input string) []protocol.ChatArtifactRef {
	writes := evidencePaths(evidence, "write_file")
	reads := append(evidencePaths(evidence, "read_file"), evidencePaths(evidence, "read_text_file")...)
	if !hasProjectPackageArtifact(artifacts) {
		for _, path := range writes {
			if artifact, ok := projectPackageArtifactFromSuccessfulWrite(map[string]any{"path": path}, input); ok {
				artifacts = append(artifacts, artifact)
				break
			}
		}
	}
	for index := range artifacts {
		if !strings.EqualFold(strings.TrimSpace(artifacts[index].Type), "project_package") {
			continue
		}
		artifacts[index] = reconcileProjectPackageArtifact(artifacts[index], writes, reads)
	}
	return artifacts
}

func reconcileProjectPackageArtifact(artifact protocol.ChatArtifactRef, writes, reads []string) protocol.ChatArtifactRef {
	entrypoint := firstHTMLWrite(writes)
	if candidate := cleanEvidencePath(artifact.Entrypoint); candidate != "" && evidenceContainsPath(writes, candidate) {
		entrypoint = matchingEvidencePath(writes, candidate)
	}
	folder := ""
	if entrypoint != "" {
		folder = strings.ReplaceAll(filepath.Dir(entrypoint), "\\", "/")
	}
	files := make([]string, 0, len(writes))
	for _, path := range writes {
		if folder == "" || !pathWithinFolder(path, folder) {
			continue
		}
		files = append(files, relativeEvidencePath(path, folder))
	}
	sort.Strings(files)
	validation := ""
	if entrypoint != "" && evidenceContainsPath(reads, entrypoint) {
		validation = fmt.Sprintf("Structural readback completed for %s; semantic acceptance is evaluated by server/live validation.", entrypoint)
	}
	artifact.Entrypoint = entrypoint
	artifact.Folder = folder
	artifact.SavedPath = folder
	artifact.Files = files
	artifact.Validation = validation
	payload, _ := json.Marshal(map[string]any{
		"title": artifact.Title, "kind": "project_package", "entrypoint": entrypoint,
		"folder": folder, "files": files, "validation": validation,
	})
	artifact.Content = string(payload)
	artifact.ContentType = "application/vnd.mycelis.project+json"
	return artifact
}

func resultContractIssues(requirement *teamResultRequirement, artifacts []protocol.ChatArtifactRef, evidence []successfulToolEvidence) []string {
	if !requirement.active() {
		return nil
	}
	writes := evidencePaths(evidence, "write_file")
	reads := append(evidencePaths(evidence, "read_file"), evidencePaths(evidence, "read_text_file")...)
	stored := evidenceHasTool(evidence, "store_artifact")
	issues := make([]string, 0, 5)
	if len(writes) == 0 && !(stored && len(artifacts) > 0) {
		issues = append(issues, "no successful retained-output write or artifact store")
	}
	packageArtifact := firstProjectPackageArtifact(artifacts)
	for _, required := range requirement.FilesRequired {
		requiredPath := cleanEvidencePath(required)
		if packageArtifact != nil && strings.TrimSpace(packageArtifact.Folder) != "" && !pathWithinFolder(requiredPath, packageArtifact.Folder) {
			requiredPath = strings.TrimRight(cleanEvidencePath(packageArtifact.Folder), "/") + "/" + strings.TrimLeft(requiredPath, "/")
		}
		if !evidenceContainsPath(writes, requiredPath) {
			issues = append(issues, "missing successful write for "+required)
		}
	}
	if strings.EqualFold(requirement.Kind, "project_package") && packageArtifact == nil {
		issues = append(issues, "missing tool-backed project package")
	}
	if requirement.EntrypointRequired && (packageArtifact == nil || packageArtifact.Entrypoint == "" || !evidenceContainsPath(writes, packageArtifact.Entrypoint)) {
		issues = append(issues, "missing tool-backed entrypoint")
	}
	if requirement.FolderRequired && (packageArtifact == nil || strings.TrimSpace(packageArtifact.Folder) == "") {
		issues = append(issues, "missing retained output folder")
	}
	// Acceptance criteria guide the worker and downstream validator. This gate
	// only asserts inspectable tool evidence; it does not grade semantic quality.
	if requirement.ReadbackRequired || requirement.DownstreamProofRef || len(requirement.ProofRequirements) > 0 {
		readbackPresent := hasReadbackEvidence(writes, reads)
		if packageArtifact != nil && strings.TrimSpace(packageArtifact.Entrypoint) != "" {
			readbackPresent = evidenceContainsPath(reads, packageArtifact.Entrypoint)
		}
		if !readbackPresent {
			issues = append(issues, "missing successful structural readback of a written output")
		}
	}
	if resultContractRequiresPrimaryInteraction(requirement) && packageArtifact != nil && packageArtifact.Entrypoint != "" {
		content := latestEntrypointEvidenceContent(evidence, packageArtifact.Entrypoint)
		if !resultContractInteractiveHandlerPattern.MatchString(content) || !resultContractExposesPrimaryControl(content) {
			issues = append(issues, "entrypoint readback does not expose an inspectable primary interaction and visible control instructions")
		}
	}
	return uniqueResultContractStrings(issues)
}

func resultContractRequiresPrimaryInteraction(requirement *teamResultRequirement) bool {
	if requirement == nil {
		return false
	}
	values := append(append([]string{}, requirement.ExpectedOutputs...), requirement.AcceptanceCriteria...)
	for _, value := range values {
		lower := strings.ToLower(value)
		if strings.Contains(lower, "playable") || strings.Contains(lower, "browser game") ||
			strings.Contains(lower, "controls respond") || strings.Contains(lower, "primary user workflow") ||
			strings.Contains(lower, "primary control") {
			return true
		}
	}
	return false
}

func latestEntrypointEvidenceContent(evidence []successfulToolEvidence, entrypoint string) string {
	content := ""
	for _, item := range evidence {
		if (item.ToolName == "write_file" || item.ToolName == "read_file" || item.ToolName == "read_text_file") &&
			evidenceContainsPath([]string{item.Path}, entrypoint) && strings.TrimSpace(item.Content) != "" {
			content = item.Content
		}
	}
	return content
}

func resultContractExposesPrimaryControl(content string) bool {
	visibleText := resultContractScriptOrStylePattern.ReplaceAllString(content, " ")
	visibleText = html.UnescapeString(resultContractHTMLTagPattern.ReplaceAllString(visibleText, " "))
	return resultContractVisibleControlPattern.MatchString(visibleText)
}

func resultContractExecutionPrompt(requirement *teamResultRequirement) string {
	files := strings.Join(requirement.FilesRequired, ", ")
	if files == "" {
		files = "the approved retained outputs"
	}
	prompt := "Approved result contract is active. Execute it through tool calls before returning final prose. " +
		"Use write_file once for each required physical output (" + files + "). For project packages, do not call store_artifact and do not read a missing support file to discover it; write each required file directly. " +
		"After writing the entrypoint, read back that entrypoint only unless the contract explicitly requires other readbacks."
	if len(requirement.AcceptanceCriteria) > 0 {
		prompt += " Implement these acceptance criteria in the retained output: " + strings.Join(requirement.AcceptanceCriteria, "; ") + "."
	}
	if strings.EqualFold(requirement.Kind, "project_package") {
		prompt += " When the requested application is interactive, its entrypoint must visibly explain the primary control and implement that control with a standard click, pointer, touch, keydown, or keyup handler that retained validation can inspect."
	}
	return prompt + " Continue until the evidence is complete or a concrete tool blocker prevents progress."
}

func resultContractCorrectionPrompt(requirement *teamResultRequirement, issues []string) string {
	prompt := "Result-contract correction: the approved delivery is not complete. Missing: " + strings.Join(issues, "; ") +
		". Emit exactly one available tool_call that advances the missing evidence. Use write_file to create required retained outputs and read_file/read_text_file on the entrypoint to establish structural readback. Do not use store_artifact as a substitute for project-package files or read missing support files before writing them. Readback does not prove semantic acceptance; server/live validation remains authoritative. Do not claim files, proof, or completion from prose or metadata."
	for _, issue := range issues {
		if strings.Contains(issue, "visible control instructions") {
			prompt += " Overwrite the existing entrypoint with write_file so visible page text names the primary control (for example, 'Controls: Hold ArrowRight to move') while retaining its matching interaction handler, then read the entrypoint back."
			break
		}
	}
	return prompt
}

func resultContractRecoveryAction(requirement *teamResultRequirement, issues []string) string {
	channel := "Soma"
	if requirement != nil && strings.TrimSpace(requirement.RepairChannel) != "" {
		channel = requirement.RepairChannel
	}
	return fmt.Sprintf("Use %s to retry the approved work after ensuring the team can write and read retained outputs. Required recovery: %s.", channel, strings.Join(issues, "; "))
}

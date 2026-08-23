package swarm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

const (
	maxResultContractCorrections    = 4
	maxResultContractToolIterations = 20
)

type teamResultRequirement struct {
	Kind               string
	TeamID             string
	PackageTitle       string
	PackageFolder      string
	PackageEntrypoint  string
	FilesRequired      []string
	ExpectedOutputs    []string
	AcceptanceCriteria []string
	ProofRequirements  []string
	EntrypointRequired bool
	FolderRequired     bool
	ReadbackRequired   bool
	DownstreamProofRef bool
	RepairChannel      string
	OutputValidation   *protocol.OutputValidationPlan
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
		TeamID:             strings.TrimSpace(stringValue(contract["team_id"])),
		PackageTitle:       strings.TrimSpace(stringValue(contract["package_title"])),
		PackageFolder:      cleanEvidencePath(stringValue(contract["package_folder"])),
		PackageEntrypoint:  cleanEvidencePath(stringValue(contract["package_entrypoint"])),
		FilesRequired:      stringSlice(contract["files_required"]),
		ExpectedOutputs:    stringSlice(contract["expected_outputs"]),
		AcceptanceCriteria: stringSlice(contract["acceptance_criteria"]),
		ProofRequirements:  stringSlice(contract["proof_required"]),
		EntrypointRequired: boolValue(contract["entrypoint_required"]),
		FolderRequired:     boolValue(contract["folder_required"]),
		ReadbackRequired:   boolValue(contract["validation_required"]),
		DownstreamProofRef: boolValue(contract["proof_ref_required"]),
		RepairChannel:      strings.TrimSpace(stringValue(contract["repair_channel"])),
		OutputValidation:   outputValidationRequirement(contract["output_validation"]),
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
		requirement.FolderRequired || requirement.ReadbackRequired || requirement.DownstreamProofRef || requirement.OutputValidation != nil)
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

func resultContractEvidenceToolAllowed(requirement *teamResultRequirement, toolName string, artifacts []protocol.ChatArtifactRef, evidence []successfulToolEvidence) bool {
	if !requirement.active() || !strings.EqualFold(requirement.Kind, "project_package") {
		return true
	}
	switch strings.TrimSpace(toolName) {
	case "write_file":
		return true
	case "read_file", "read_text_file":
		return !resultContractNeedsRequiredWrites(requirement, artifacts, evidence) &&
			!resultContractEntrypointNeedsRepair(requirement, artifacts, evidence)
	default:
		return false
	}
}

func resultContractEntrypointNeedsRepair(requirement *teamResultRequirement, artifacts []protocol.ChatArtifactRef, evidence []successfulToolEvidence) bool {
	packageArtifact := firstProjectPackageArtifact(artifacts)
	if packageArtifact == nil || strings.TrimSpace(packageArtifact.Entrypoint) == "" {
		return false
	}
	content := resultContractPackageValidationContent(artifacts, evidence)
	if resultContractRequiresPrimaryInteraction(requirement) &&
		!resultContractExposesInspectablePrimaryInteraction(content) {
		return true
	}
	return len(outputValidationTargetIssues(requirement.OutputValidation, content)) > 0 ||
		len(outputValidationAnimationLoopIssues(requirement.OutputValidation, content)) > 0
}

func pendingProjectPackageEntrypointReadback(requirement *teamResultRequirement, artifacts []protocol.ChatArtifactRef, evidence []successfulToolEvidence) string {
	if !requirement.active() || !strings.EqualFold(requirement.Kind, "project_package") ||
		resultContractNeedsRequiredWrites(requirement, artifacts, evidence) ||
		resultContractEntrypointNeedsRepair(requirement, artifacts, evidence) {
		return ""
	}
	packageArtifact := firstProjectPackageArtifact(artifacts)
	if packageArtifact == nil || strings.TrimSpace(packageArtifact.Entrypoint) == "" ||
		hasCurrentReadbackEvidence(evidence, packageArtifact.Entrypoint) {
		return ""
	}
	return cleanEvidencePath(packageArtifact.Entrypoint)
}

func currentProjectPackageEntrypointReadback(requirement *teamResultRequirement, artifacts []protocol.ChatArtifactRef, evidence []successfulToolEvidence) string {
	if !requirement.active() || !strings.EqualFold(requirement.Kind, "project_package") {
		return ""
	}
	packageArtifact := firstProjectPackageArtifact(artifacts)
	if packageArtifact == nil || !hasCurrentReadbackEvidence(evidence, packageArtifact.Entrypoint) {
		return ""
	}
	return cleanEvidencePath(packageArtifact.Entrypoint)
}

func resultContractNeedsRequiredWrites(requirement *teamResultRequirement, artifacts []protocol.ChatArtifactRef, evidence []successfulToolEvidence) bool {
	if requirement == nil {
		return false
	}
	writes := evidencePaths(evidence, "write_file")
	packageArtifact := firstProjectPackageArtifact(artifacts)
	if requirement.EntrypointRequired && (packageArtifact == nil || packageArtifact.Entrypoint == "" || !evidenceContainsPath(writes, packageArtifact.Entrypoint)) {
		return true
	}
	for _, required := range requirement.FilesRequired {
		requiredPath := cleanEvidencePath(required)
		if packageArtifact != nil && strings.TrimSpace(packageArtifact.Folder) != "" && !pathWithinFolder(requiredPath, packageArtifact.Folder) {
			requiredPath = strings.TrimRight(cleanEvidencePath(packageArtifact.Folder), "/") + "/" + strings.TrimLeft(requiredPath, "/")
		}
		if !evidenceContainsPath(writes, requiredPath) {
			return true
		}
	}
	return len(resultContractLocalDependencyIssues(artifacts, evidence)) > 0
}

func reconcileToolBackedArtifacts(artifacts []protocol.ChatArtifactRef, evidence []successfulToolEvidence, input string) []protocol.ChatArtifactRef {
	writes := evidencePaths(evidence, "write_file")
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
		artifacts[index] = reconcileProjectPackageArtifact(artifacts[index], writes, evidence)
	}
	return artifacts
}

func reconcileProjectPackageArtifact(artifact protocol.ChatArtifactRef, writes []string, evidence []successfulToolEvidence) protocol.ChatArtifactRef {
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
	if entrypoint != "" && hasCurrentReadbackEvidence(evidence, entrypoint) {
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
	issues = append(issues, resultContractLocalDependencyIssues(artifacts, evidence)...)
	// Acceptance criteria guide the worker and downstream validator. This gate
	// only asserts inspectable tool evidence; it does not grade semantic quality.
	if requirement.ReadbackRequired || requirement.DownstreamProofRef || len(requirement.ProofRequirements) > 0 {
		readbackPresent := hasCurrentReadbackEvidence(evidence, "")
		if packageArtifact != nil && strings.TrimSpace(packageArtifact.Entrypoint) != "" {
			readbackPresent = hasCurrentReadbackEvidence(evidence, packageArtifact.Entrypoint)
		}
		if !readbackPresent {
			issues = append(issues, "missing successful structural readback of a written output")
		}
	}
	if packageArtifact != nil && packageArtifact.Entrypoint != "" {
		content := resultContractPackageValidationContent(artifacts, evidence)
		if resultContractRequiresPrimaryInteraction(requirement) &&
			!resultContractExposesInspectablePrimaryInteraction(content) {
			issues = append(issues, "entrypoint readback does not expose an inspectable primary interaction and visible control instructions")
		}
		issues = append(issues, outputValidationTargetIssues(requirement.OutputValidation, content)...)
		issues = append(issues, outputValidationTextChangeIssues(requirement.OutputValidation, content)...)
		issues = append(issues, outputValidationAnimationLoopIssues(requirement.OutputValidation, content)...)
	}
	return uniqueResultContractStrings(issues)
}

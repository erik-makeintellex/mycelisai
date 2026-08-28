package swarm

import (
	"fmt"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

func resultContractExecutionPrompt(requirement *teamResultRequirement) string {
	prompt := "Execution policy for the governed package ask: the structured user message contains the sole authoritative PACKAGE CONTRACT v1. " +
		"Do not repeat, relax, or replace its targets or acceptance rubric. Emit exactly one tool_call JSON object per response until its evidence is complete. " +
		"Use write_file once for each missing physical output. For project packages, do not call store_artifact and do not read a missing support file to discover it; write each required file directly. " +
		"After writing the entrypoint, read back that entrypoint only unless the contract explicitly requires other readbacks."
	if strings.EqualFold(requirement.Kind, "project_package") {
		prompt += " When the requested application is interactive, its entrypoint must visibly explain the primary control and implement that control with a standard click, pointer, touch, keydown, or keyup handler that retained validation can inspect."
		prompt += " On the entrypoint write, include the package_kind, package_folder, package_entrypoint, and package_files metadata named by the contract."
	}
	prompt += outputValidationExecutionInstruction(requirement.OutputValidation)
	return prompt + " Continue until the evidence is complete or a concrete tool blocker prevents progress."
}

func resultContractCorrectionPrompt(requirement *teamResultRequirement, issues []string, artifacts []protocol.ChatArtifactRef, evidence []successfulToolEvidence) string {
	focusedIssues := focusedResultContractCorrectionIssues(issues)
	prompt := "Result-contract correction: complete only this next gap: " + strings.Join(focusedIssues, "; ") +
		". Emit exactly one tool_call JSON object and no second tool call; the runtime executes only the first object in a response. Use write_file for a missing or invalid retained output, or read_file/read_text_file only when entrypoint readback is the named gap. Do not use store_artifact as a substitute for project-package files. Do not claim files, proof, or completion from prose or metadata."
	if instruction := resultContractTargetInstruction(requirement, focusedIssues, artifacts, evidence); instruction != "" {
		prompt += " " + instruction
	}
	for _, issue := range focusedIssues {
		if strings.Contains(issue, "visible control instructions") {
			prompt += " Overwrite the existing entrypoint with write_file so visible page text names the primary control (for example, 'Controls: Hold ArrowRight to move') while retaining its matching interaction handler, then read the entrypoint back."
			break
		}
	}
	if strings.Contains(strings.Join(focusedIssues, " "), "missing semantic validation target") {
		prompt += " Overwrite the entrypoint to include every named semantic validation target as a visible functional control or state surface, wire each control to the requested gameplay state, preserve the rest of the package, then read the entrypoint back."
	}
	prompt += outputValidationCorrectionInstruction(requirement.OutputValidation, focusedIssues)
	return prompt
}

func resultContractDefaultFolder(requirement *teamResultRequirement) string {
	if requirement == nil {
		return ""
	}
	if folder := cleanEvidencePath(requirement.PackageFolder); folder != "" {
		return folder
	}
	if teamID := strings.TrimSpace(requirement.TeamID); teamID != "" {
		return "groups/" + teamID + "/generated/package"
	}
	return ""
}

func resultContractDefaultEntrypoint(requirement *teamResultRequirement) string {
	if requirement == nil {
		return ""
	}
	if entrypoint := cleanEvidencePath(requirement.PackageEntrypoint); entrypoint != "" {
		return entrypoint
	}
	if folder := resultContractDefaultFolder(requirement); folder != "" {
		return strings.TrimRight(folder, "/") + "/index.html"
	}
	return ""
}

func resultContractTargetInstruction(requirement *teamResultRequirement, issues []string, artifacts []protocol.ChatArtifactRef, evidence []successfulToolEvidence) string {
	if !requirement.active() || !strings.EqualFold(requirement.Kind, "project_package") || len(issues) == 0 {
		return ""
	}
	artifact := firstProjectPackageArtifact(artifacts)
	folder := resultContractDefaultFolder(requirement)
	if artifact != nil && strings.TrimSpace(artifact.Folder) != "" {
		folder = cleanEvidencePath(artifact.Folder)
	}
	entrypoint := resultContractDefaultEntrypoint(requirement)
	if artifact != nil && strings.TrimSpace(artifact.Entrypoint) != "" {
		entrypoint = cleanEvidencePath(artifact.Entrypoint)
	}
	for _, issue := range issues {
		if strings.Contains(issue, "structural readback") && entrypoint != "" {
			return "Read back the current entrypoint with read_file at " + entrypoint + "."
		}
		if strings.Contains(issue, "missing tool-backed project package") || strings.Contains(issue, "missing tool-backed entrypoint") {
			if entrypoint == "" {
				return ""
			}
			return "Write the entrypoint with write_file at " + entrypoint + " and include package_kind=project_package, package_folder=" + resultContractDefaultFolder(requirement) + ", package_entrypoint=" + entrypoint + ", package_files=[index.html, README.md, PROOF.md, project-package.json]."
		}
		if strings.HasPrefix(issue, "missing successful write for ") {
			return resultContractMissingWriteInstruction(issue, folder)
		}
	}
	_ = evidence
	return ""
}

func resultContractMissingWriteInstruction(issue string, folder string) string {
	required := cleanEvidencePath(strings.TrimPrefix(issue, "missing successful write for "))
	if required == "" {
		return ""
	}
	target := required
	if folder != "" && !pathWithinFolder(target, folder) {
		target = strings.TrimRight(folder, "/") + "/" + strings.TrimLeft(target, "/")
	}
	if strings.HasSuffix(strings.ToLower(target), ".html") {
		parent := strings.TrimSuffix(target, "/"+pathBase(target))
		return "Write the entrypoint with write_file at " + target + " and include package_kind=project_package, package_folder=" + parent + ", package_entrypoint=" + target + ", package_files=[index.html, README.md, PROOF.md, project-package.json]."
	}
	return "Write the missing package file with write_file at " + target + "."
}

func resultContractRecoveryAction(requirement *teamResultRequirement, issues []string) string {
	channel := "Soma"
	if requirement != nil && strings.TrimSpace(requirement.RepairChannel) != "" {
		channel = requirement.RepairChannel
	}
	return fmt.Sprintf("Use %s to retry the approved work after ensuring the team can write and read retained outputs. Required recovery: %s.", channel, strings.Join(issues, "; "))
}

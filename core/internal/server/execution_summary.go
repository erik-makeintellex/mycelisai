package server

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

func boolPtr(value bool) *bool {
	return &value
}

func buildDirectChatExecutionSummary(originalIntent, replyText, auditEventID string, agentResult chatAgentResult) *protocol.ExecutionSummary {
	summaryText := firstNonEmptyString(replyText, "Soma answered directly.")
	shape := protocol.ExecutionShapeDirectSoma
	executionSummary := "Soma completed a direct response."
	if len(agentResult.ToolsUsed) > 0 || len(agentResult.Artifacts) > 0 {
		shape = protocol.ExecutionShapeToolAssistedWork
		executionSummary = "Soma completed tool-assisted work."
	}
	outputs := []protocol.ExecutionOutput{{
		Kind:           "answer",
		Title:          "Soma answer",
		Summary:        summaryText,
		Retained:       boolPtr(false),
		RetentionClass: protocol.ExecutionRetentionClassNonRetained,
	}}
	outputs = append(outputs, executionOutputsFromArtifacts(agentResult.Artifacts)...)

	return &protocol.ExecutionSummary{
		Intent: protocol.ExecutionIntent{
			Original: strings.TrimSpace(originalIntent),
			Resolved: "answer",
		},
		Understanding: protocol.ExecutionUnderstanding{
			Summary: summaryText,
		},
		Execution: protocol.ExecutionState{
			Shape:   shape,
			Status:  protocol.ExecutionStatusCompleted,
			Summary: executionSummary,
		},
		CapabilityUse: capabilityUseFromChat(agentResult.ToolsUsed, agentResult.Consultations, ""),
		Outputs:       outputs,
		Proof: protocol.ExecutionProof{
			RunClass:     protocol.ExecutionRunClassNoRun,
			NoRunReason:  "Direct Soma answers are audit-only and do not create execution runs.",
			ProofClass:   protocol.ExecutionProofClassAuditOnly,
			AuditEventID: auditEventID,
			Verified:     boolPtr(strings.TrimSpace(auditEventID) != ""),
		},
		AuditRecovery: protocol.AuditRecovery{
			ApprovalStatus: "allow",
			RecoveryState:  "completed",
			Retryable:      boolPtr(true),
		},
		NextStep: &protocol.ExecutionNextStep{
			Label:  "Ask a follow-up",
			Action: "chat",
		},
	}
}

func buildProposalExecutionSummary(originalIntent string, planned []protocol.PlannedToolCall, mutTools []string, display proposalDisplayContract, proofID, contractID, auditEventID string, approval *protocol.ApprovalPolicy) *protocol.ExecutionSummary {
	approvalStatus := approvalStatusValue(approval)
	recoveryState := "awaiting_confirmation"
	if approvalStatus == "approval_required" {
		recoveryState = "approval_required"
	}
	blocker := approvalReasonValue(approval)

	return &protocol.ExecutionSummary{
		ContractID: contractID,
		Intent: protocol.ExecutionIntent{
			Original: strings.TrimSpace(originalIntent),
			Resolved: "chat-action",
		},
		Understanding: protocol.ExecutionUnderstanding{
			Summary: display.OperatorSummary,
		},
		Execution: protocol.ExecutionState{
			Shape:   protocol.ExecutionShapeGuidedProposal,
			Status:  protocol.ExecutionStatusProposed,
			Summary: display.ExpectedResult,
		},
		CapabilityUse: capabilityUseFromPlannedCalls(planned, mutTools, chatToolRisk(mutTools)),
		Outputs: []protocol.ExecutionOutput{{
			ID:             proofID,
			Kind:           "proposal",
			Title:          "Guided proposal",
			Summary:        display.OperatorSummary,
			Retained:       boolPtr(true),
			RetentionClass: protocol.ExecutionRetentionClassRetained,
		}},
		Proof: protocol.ExecutionProof{
			ContractID:    contractID,
			RunClass:      protocol.ExecutionRunClassNoRun,
			NoRunReason:   "Guided proposals create intent proof first; no execution run exists until confirmation.",
			ProofClass:    protocol.ExecutionProofClassIntentOnly,
			AuditEventID:  auditEventID,
			IntentProofID: proofID,
			Verified:      boolPtr(strings.TrimSpace(proofID) != "" && strings.TrimSpace(auditEventID) != ""),
		},
		AuditRecovery: protocol.AuditRecovery{
			ApprovalStatus: approvalStatus,
			RecoveryState:  recoveryState,
			Blocker:        blocker,
			Retryable:      boolPtr(true),
		},
		NextStep: &protocol.ExecutionNextStep{
			Label:  "Confirm proposal",
			Action: "confirm_action",
			Href:   "/api/v1/intent/confirm-action",
		},
	}
}

func buildConfirmActionExecutionSummary(proofID, contractID, proofArtifactID, runID, auditID string, scope *protocol.ScopeValidation, results []plannedToolExecutionResult) *protocol.ExecutionSummary {
	capabilities := []protocol.CapabilityUse{}
	if scope != nil {
		capabilities = capabilityUseFromPlannedCalls(scope.PlannedToolCalls, scope.Tools, scope.RiskLevel)
	}
	outputs := executionOutputsFromToolResults(results)
	outputs = attachConfirmActionOutputProofs(outputs, proofArtifactID, runID, contractID, results)
	understandingSummary := "Confirmed proposal execution completed."
	executionStateSummary := "Soma executed the confirmed proposal and recorded durable proof."
	nextStep := &protocol.ExecutionNextStep{
		Label:  "Review run",
		Action: "view_run",
		Href:   "/api/v1/runs/" + runID,
	}
	if toolResultExists(results, "create_team") && toolResultExists(results, "write_file") {
		understandingSummary = "Team created and its first retained deliverable completed."
		executionStateSummary = "Soma created the governed team, produced the first reviewable output, and recorded durable proof."
		nextStep = &protocol.ExecutionNextStep{
			Label:  "Open the output, then ask Soma for the next team task.",
			Action: "chat",
			Href:   "/api/v1/runs/" + runID,
		}
	} else if toolResultExists(results, "create_team") {
		understandingSummary = "Team created. No work item has started yet."
		executionStateSummary = "Soma created the governed team and recorded proof. Ask Soma for the team's first task to begin visible work."
		nextStep = &protocol.ExecutionNextStep{
			Label:  "Ask Soma to start the team's first work item.",
			Action: "chat",
			Href:   "/api/v1/runs/" + runID,
		}
	}

	return &protocol.ExecutionSummary{
		ContractID: contractID,
		ProofID:    proofArtifactID,
		Intent: protocol.ExecutionIntent{
			Resolved: "chat-action",
		},
		Understanding: protocol.ExecutionUnderstanding{
			Summary: understandingSummary,
		},
		Execution: protocol.ExecutionState{
			Shape:   protocol.ExecutionShapeGuidedProposal,
			Status:  protocol.ExecutionStatusCompleted,
			Summary: executionStateSummary,
		},
		CapabilityUse: capabilities,
		Outputs:       outputs,
		Proof: protocol.ExecutionProof{
			ContractID:    contractID,
			ProofID:       proofArtifactID,
			RunID:         runID,
			RunClass:      protocol.ExecutionRunClassLinked,
			ProofClass:    protocol.ExecutionProofClassRunAudit,
			AuditEventID:  auditID,
			IntentProofID: proofID,
			Verified:      boolPtr(true),
		},
		AuditRecovery: protocol.AuditRecovery{
			ApprovalStatus: "confirmed",
			RecoveryState:  "verified",
			Retryable:      boolPtr(false),
		},
		NextStep: nextStep,
	}
}

func attachConfirmActionOutputProofs(outputs []protocol.ExecutionOutput, proofArtifactID, runID, contractID string, results []plannedToolExecutionResult) []protocol.ExecutionOutput {
	for i := range outputs {
		output := &outputs[i]
		output.ProofArtifactID = proofArtifactID
		output.OpenURL = firstNonEmptyString(output.OpenURL, output.Href)
		proof := &protocol.OutputProofEnvelope{
			ProofID:            proofArtifactID,
			OutputRefID:        firstNonEmptyString(output.ID, output.Title),
			ArtifactID:         output.ArtifactID,
			StorageRef:         firstNonEmptyString(output.Href, output.Folder, output.Entrypoint, output.ID),
			SourceRunID:        runID,
			SourceContractID:   contractID,
			ExecutionStatus:    "verified",
			PathBoundaryStatus: "verified",
			ReadbackStatus:     "verified",
		}
		if checksum, bytes, contentType := outputContentProof(output, results); checksum != "" {
			proof.Checksum = checksum
			proof.ChecksumAlgorithm = "sha256"
			proof.Bytes = bytes
			proof.ContentType = contentType
		}
		if proof.StorageRef == "" {
			proof.PathBoundaryStatus = "not_applicable"
		}
		output.Proof = proof
	}
	return outputs
}

func outputContentProof(output *protocol.ExecutionOutput, results []plannedToolExecutionResult) (string, int64, string) {
	if output == nil {
		return "", 0, ""
	}
	outputRef := firstNonEmptyString(output.ID, output.Entrypoint, output.Folder, output.Title)
	for _, result := range results {
		path := firstNonEmptyString(result.Arguments["path"], result.Arguments["file_path"], result.Arguments["target_path"])
		if path != "" && (path == outputRef || path == output.ID || path == output.Entrypoint) {
			content := firstNonEmptyString(result.Arguments["content"], result.Arguments["body"], result.Arguments["text"])
			if content != "" {
				sum := sha256.Sum256([]byte(content))
				return hex.EncodeToString(sum[:]), int64(len([]byte(content))), contentTypeForOutput(output)
			}
		}
		for _, artifact := range result.Artifacts {
			if artifact.Content == "" {
				continue
			}
			if artifact.ID == output.ID || artifact.Title == output.Title || artifact.Entrypoint == output.Entrypoint {
				sum := sha256.Sum256([]byte(artifact.Content))
				return hex.EncodeToString(sum[:]), int64(len([]byte(artifact.Content))), firstNonEmptyString(artifact.ContentType, contentTypeForOutput(output))
			}
		}
	}
	return "", 0, ""
}

func contentTypeForOutput(output *protocol.ExecutionOutput) string {
	if output == nil {
		return ""
	}
	switch strings.ToLower(strings.TrimSpace(output.Kind)) {
	case "code", "file", "document":
		return "text/plain"
	case "project_package":
		return "application/vnd.mycelis.project-package+json"
	default:
		return ""
	}
}

func toolResultExists(results []plannedToolExecutionResult, name string) bool {
	for _, result := range results {
		if strings.TrimSpace(result.Name) == name {
			return true
		}
	}
	return false
}

func buildConfirmActionFailureExecutionSummary(proofID, contractID, proofArtifactID, runID, auditID string, err error) *protocol.ExecutionSummary {
	blocker := "The approved execution failed."
	if err != nil {
		blocker = firstNonEmptyString(strings.TrimSpace(err.Error()), blocker)
	}
	runClass := protocol.ExecutionRunClassNoRun
	noRunReason := "The approved execution did not retain a run id."
	if strings.TrimSpace(runID) != "" {
		runClass = protocol.ExecutionRunClassLinked
		noRunReason = ""
	}
	var nextStep *protocol.ExecutionNextStep
	if strings.TrimSpace(runID) != "" {
		nextStep = &protocol.ExecutionNextStep{
			Label:  "Review failed run",
			Action: "view_run",
			Href:   "/api/v1/runs/" + runID,
		}
	}
	return &protocol.ExecutionSummary{
		ContractID: contractID,
		ProofID:    proofArtifactID,
		Intent: protocol.ExecutionIntent{
			Resolved: "chat-action",
		},
		Understanding: protocol.ExecutionUnderstanding{
			Summary: "Soma accepted the approval, then execution degraded before completion.",
		},
		Execution: protocol.ExecutionState{
			Shape:   protocol.ExecutionShapeGuidedProposal,
			Status:  protocol.ExecutionStatusFailed,
			Summary: "Soma could not complete the approved proposal.",
		},
		Proof: protocol.ExecutionProof{
			ContractID:    contractID,
			ProofID:       proofArtifactID,
			RunID:         runID,
			RunClass:      runClass,
			NoRunReason:   noRunReason,
			ProofClass:    protocol.ExecutionProofClassRunAudit,
			AuditEventID:  auditID,
			IntentProofID: proofID,
			Verified:      boolPtr(false),
		},
		AuditRecovery: protocol.AuditRecovery{
			ApprovalStatus: "confirmed",
			RecoveryState:  "failed",
			Blocker:        blocker,
			Retryable:      boolPtr(true),
			Degradation: &protocol.ExecutionDegradation{
				Code:              "approved_execution_failed",
				WhatFailed:        blocker,
				TrustedState:      "The approval, intent proof, failed run record, and audit event remain trusted.",
				InvalidatedProof:  "No completed execution proof or retained output should be trusted for this attempt.",
				SafeContinuation:  "Review the failed run, adjust the request or runtime dependency, then retry the proposal.",
				RequiresAttention: true,
			},
		},
		NextStep: nextStep,
	}
}

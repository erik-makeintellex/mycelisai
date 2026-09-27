package server

import (
	"net/url"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

const workspaceFileViewPrefix = "/api/v1/workspace/files/view?"

// confirmActionReadback aggregates Core's readback of every workspace output a
// confirmed action claims to have written.
type confirmActionReadback struct {
	Checked int
	Failed  []outputReadback
}

func (r confirmActionReadback) verified() bool {
	return len(r.Failed) == 0
}

func attachConfirmActionOutputProofs(outputs []protocol.ExecutionOutput, proofArtifactID, runID, contractID string, results []plannedToolExecutionResult, requestEchoes []string) ([]protocol.ExecutionOutput, confirmActionReadback) {
	readback := confirmActionReadback{}
	for i := range outputs {
		output := &outputs[i]
		output.ProofArtifactID = proofArtifactID
		output.OpenURL = firstNonEmptyString(output.OpenURL, output.Href)
		proof := &protocol.OutputProofEnvelope{
			ProofID:          proofArtifactID,
			OutputRefID:      firstNonEmptyString(output.ID, output.Title),
			ArtifactID:       output.ArtifactID,
			StorageRef:       firstNonEmptyString(output.Href, output.Folder, output.Entrypoint, output.ID),
			SourceRunID:      runID,
			SourceContractID: contractID,
			ExecutionStatus:  string(protocol.ExecutionStatusCompleted),
		}
		storagePath := confirmActionOutputPath(output)
		if storagePath == "" {
			// No workspace file to read back (team, config revision, external
			// reference): the proof rests on the run and audit records only.
			proof.PathBoundaryStatus = outputReadbackNotApplicable
			proof.ReadbackStatus = outputReadbackNotApplicable
			output.Proof = proof
			continue
		}
		result := readbackWorkspaceOutput(storagePath, requestedWriteChecksum(storagePath, results), requestEchoes)
		readback.Checked++
		proof.StorageRef = firstNonEmptyString(result.Path, proof.StorageRef)
		proof.PathBoundaryStatus = result.PathBoundary
		proof.ReadbackStatus = result.Status
		if result.Checksum != "" {
			proof.Checksum = result.Checksum
			proof.ChecksumAlgorithm = "sha256"
			proof.Bytes = result.Bytes
			proof.ContentType = contentTypeForOutput(output)
		}
		if !result.verified() {
			proof.RecoveryHint = result.Detail
			readback.Failed = append(readback.Failed, result)
		}
		output.Proof = proof
	}
	return outputs, readback
}

// confirmActionOutputPath returns the workspace path an output claims, from
// its workspace view link or its entrypoint.
func confirmActionOutputPath(output *protocol.ExecutionOutput) string {
	if output == nil {
		return ""
	}
	href := strings.TrimSpace(output.Href)
	if strings.HasPrefix(href, workspaceFileViewPrefix) {
		if values, err := url.ParseQuery(strings.TrimPrefix(href, workspaceFileViewPrefix)); err == nil {
			if path := strings.TrimSpace(values.Get("path")); path != "" {
				return path
			}
		}
	}
	return strings.TrimSpace(output.Entrypoint)
}

// requestedWriteChecksum is the checksum of the exact bytes an internal
// write_file call reported writing to path; the disk must match it.
func requestedWriteChecksum(path string, results []plannedToolExecutionResult) string {
	want := normalizeWorkspacePath(path)
	for _, result := range results {
		if !strings.EqualFold(strings.TrimSpace(result.Name), "write_file") || strings.TrimSpace(result.ToolRef) != "" {
			continue
		}
		written := firstNonEmptyString(result.Arguments["path"], result.Arguments["file_path"], result.Arguments["target_path"])
		content, ok := result.Arguments["content"].(string)
		if ok && content != "" && written != "" && normalizeWorkspacePath(written) == want {
			return sha256Hex(content)
		}
	}
	return ""
}

func confirmActionRequestEchoes(scope *protocol.ScopeValidation) []string {
	if scope == nil || scope.WorkIntent == nil {
		return nil
	}
	return normalizeStringSlice([]string{scope.WorkIntent.Objective})
}

// applyConfirmActionReadbackFailure turns a completed run whose outputs did not
// read back into an honest unverified result with a recovery path.
func applyConfirmActionReadbackFailure(summary *protocol.ExecutionSummary, readback confirmActionReadback, runID string) {
	if summary == nil || readback.verified() {
		return
	}
	first := readback.Failed[0]
	whatFailed := "Output readback " + first.Status + " for " + firstNonEmptyString(first.Path, "the output") + ": " + first.Detail + "."
	summary.Understanding.Summary = "Soma ran the approved action, but Core could not verify the written output."
	summary.Execution.Summary = whatFailed
	summary.Proof.Verified = boolPtr(false)
	summary.AuditRecovery.RecoveryState = "unverified"
	summary.AuditRecovery.Blocker = whatFailed
	summary.AuditRecovery.Retryable = boolPtr(true)
	summary.AuditRecovery.Degradation = &protocol.ExecutionDegradation{
		Code:              "output_readback_" + first.Status,
		WhatFailed:        whatFailed,
		TrustedState:      "The approval, intent proof, run record, and audit event remain trusted.",
		InvalidatedProof:  "The written output was not verified. Do not treat it as a delivered result.",
		SafeContinuation:  "Review the run and the workspace file, then ask Soma to retry the proposal.",
		RequiresAttention: true,
	}
	summary.NextStep = &protocol.ExecutionNextStep{Label: "Review run", Action: "view_run", Href: "/api/v1/runs/" + runID}
}

// confirmActionProofQuality reads the confirmed action's workspace outputs
// back and returns the proof quality its completion may claim.
func confirmActionProofQuality(scope *protocol.ScopeValidation, results []plannedToolExecutionResult) protocol.TrustProofQuality {
	_, readback := attachConfirmActionOutputProofs(executionOutputsFromToolResults(results), "", "", "", results, confirmActionRequestEchoes(scope))
	if readback.verified() {
		return protocol.TrustProofQualityVerified
	}
	return protocol.TrustProofQualityFailed
}

// runCompletionProofQuality fails closed: anything but verified or failed is unverified.
func runCompletionProofQuality(quality protocol.TrustProofQuality) protocol.TrustProofQuality {
	switch quality {
	case protocol.TrustProofQualityVerified, protocol.TrustProofQualityFailed:
		return quality
	default:
		return protocol.TrustProofQualityUnverified
	}
}

func runCompletionExecutionState(quality protocol.TrustProofQuality) string {
	if quality == protocol.TrustProofQualityVerified {
		return "verified"
	}
	return "unverified"
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

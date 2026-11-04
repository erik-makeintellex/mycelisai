package server

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/artifacts"
	"github.com/mycelis/core/pkg/protocol"
)

func buildGroupOutputsExecutionSummary(group *CollaborationGroup, items []artifacts.Artifact) *protocol.ExecutionSummary {
	if group == nil || len(items) == 0 {
		return nil
	}
	outputRefs := make([]protocol.OutputRef, 0, len(items))
	outputs := make([]protocol.ExecutionOutput, 0, len(items))
	teamRefs := make([]protocol.TeamOutputRef, 0, len(items))
	runID := ""

	for index, item := range items {
		role := "primary"
		if index > 0 {
			role = "supporting"
		}
		if isPackageProofFile(item.FilePath, item.Title) {
			role = "proof"
		}
		teamRef := teamOutputRefFromGroupArtifact(item, group.ID, role)
		outRef := protocol.OutputRefFromTeamOutputRef(teamRef)
		if runID == "" && outRef.SourceRunID != "" {
			runID = outRef.SourceRunID
		}
		teamRefs = append(teamRefs, teamRef)
		outputRefs = append(outputRefs, outRef)
		outputs = append(outputs, executionOutputFromOutputRef(
			outRef,
			fmt.Sprintf("Retained group output from %s.", firstNonEmptyString(item.AgentID, group.Name)),
			true,
		))
	}

	return &protocol.ExecutionSummary{
		Title:          firstNonEmptyString(group.Name, "Group retained outputs"),
		Summary:        fmt.Sprintf("%d retained output(s) recorded for %s.", len(outputRefs), firstNonEmptyString(group.Name, group.ID)),
		Status:         outputRefs[0].Status,
		RunID:          runID,
		ProofID:        outputRefs[0].ProofArtifactID,
		OutputRefs:     outputRefs,
		Outputs:        outputs,
		TeamOutputRefs: teamRefs,
	}
}

func teamOutputRefFromGroupArtifact(item artifacts.Artifact, groupID, role string) protocol.TeamOutputRef {
	artifactID := item.ID.String()
	meta := decodeArtifactMetadata(item.Metadata)
	runID := firstNonEmptyString(item.TraceID, lookupString([]map[string]any{meta}, "run_id"), uuidPtrString(item.MissionID))
	contractID := lookupString([]map[string]any{meta}, "source_contract_id", "proof_id")
	storageRef := firstNonEmptyString(item.FilePath, lookupString([]map[string]any{meta}, "storage_ref"), "artifact://"+artifactID)
	openURL := firstNonEmptyString(artifactDownloadURL(artifactID), workspaceFileOpenURL(item.FilePath))
	checksum := lookupString([]map[string]any{meta}, "sha256", "checksum")
	sizeBytes := int(item.FileSizeBytes)
	if sizeBytes == 0 && len(item.Content) > 0 {
		sizeBytes = len(item.Content)
	}
	if checksum == "" && len(item.Content) > 0 {
		sum := sha256.Sum256([]byte(item.Content))
		checksum = fmt.Sprintf("%x", sum)
	}
	status := firstNonEmptyString(strings.TrimSpace(item.Status), "approved")
	proof := &protocol.OutputProofEnvelope{
		ProofID:            firstNonEmptyString(contractID, artifactID),
		OutputRefID:        artifactID,
		ArtifactID:         artifactID,
		StorageRef:         storageRef,
		SourceRunID:        runID,
		SourceContractID:   contractID,
		ExecutionStatus:    status,
		PathBoundaryStatus: proofStatusVerified,
		Checksum:           checksum,
		ChecksumAlgorithm:  "sha256",
		Bytes:              sizeBytes,
		ContentType:        item.ContentType,
		CreatedAt:          item.CreatedAt.UTC().Format(time.RFC3339),
		ReadbackStatus:     proofStatusVerified,
	}
	return protocol.TeamOutputRef{
		OutputID:    artifactID,
		ArtifactID:  artifactID,
		TeamID:      firstNonEmptyString(uuidPtrString(item.TeamID), groupID),
		RunID:       runID,
		Kind:        firstNonEmptyString(string(item.ArtifactType), outputKindForContentType(item.ContentType)),
		Label:       firstNonEmptyString(item.Title, path.Base(storageRef), artifactID),
		StorageRef:  storageRef,
		OpenURL:     openURL,
		RevealURL:   openURL,
		ProofRef:    firstNonEmptyString(contractID, artifactID),
		ContractID:  contractID,
		ProofID:     firstNonEmptyString(contractID, artifactID),
		PackageID:   firstNonEmptyString(lookupString([]map[string]any{meta}, "package_id"), groupID),
		Role:        firstNonEmptyString(role, "primary"),
		Status:      status,
		ContentType: item.ContentType,
		Bytes:       sizeBytes,
		SHA256:      checksum,
		Proof:       proof,
	}
}

func buildTeamLeadGuidanceExecutionSummary(response TeamLeadGuidanceResponse, requestContext, runID, contractID string) *protocol.ExecutionSummary {
	contract := response.ExecutionContract
	if contract == nil || len(contract.TargetOutputs) == 0 {
		return nil
	}
	groupID := ""
	if contract.WorkflowGroup != nil {
		groupID = firstNonEmptyString(contract.WorkflowGroup.GroupID, contract.WorkflowGroup.Name)
	}
	packageID := firstNonEmptyString(groupID, contract.TeamName, "team-lead-guidance")
	status := "proposed"
	if contract.ExecutionMode == TeamLeadExecutionModeContinuityResume {
		status = "retained"
	}

	outputRefs := make([]protocol.OutputRef, 0, len(contract.TargetOutputs))
	outputs := make([]protocol.ExecutionOutput, 0, len(contract.TargetOutputs))
	teamRefs := make([]protocol.TeamOutputRef, 0, len(contract.TargetOutputs))

	for index, target := range contract.TargetOutputs {
		label := strings.TrimSpace(target)
		if label == "" {
			continue
		}
		role := "primary"
		if len(outputRefs) > 0 {
			role = "supporting"
		}
		refID := fmt.Sprintf("%s:output-%d", packageID, index+1)
		storageRef := fmt.Sprintf("guidance://%s/%d", url.PathEscape(packageID), index+1)
		openURL := "/deployments?group=" + url.QueryEscape(packageID)
		sum := sha256.Sum256([]byte(label))
		checksum := fmt.Sprintf("%x", sum)
		proof := &protocol.OutputProofEnvelope{
			ProofID:           firstNonEmptyString(contractID, refID),
			OutputRefID:       refID,
			StorageRef:        storageRef,
			SourceRunID:       runID,
			SourceContractID:  contractID,
			ExecutionStatus:   status,
			Checksum:          checksum,
			ChecksumAlgorithm: "sha256",
			Bytes:             len(label),
			ContentType:       "text/markdown",
			CreatedAt:         time.Now().UTC().Format(time.RFC3339),
			ReadbackStatus:    status,
		}
		teamRef := protocol.TeamOutputRef{
			OutputID:    refID,
			TeamID:      firstNonEmptyString(contract.TeamName, groupID),
			RunID:       runID,
			Kind:        "document",
			Label:       label,
			StorageRef:  storageRef,
			OpenURL:     openURL,
			RevealURL:   openURL,
			ProofRef:    firstNonEmptyString(contractID, refID),
			ContractID:  contractID,
			ProofID:     firstNonEmptyString(contractID, refID),
			PackageID:   packageID,
			Role:        role,
			Status:      status,
			ContentType: "text/markdown",
			Bytes:       len(label),
			SHA256:      checksum,
			Proof:       proof,
		}
		outRef := protocol.OutputRefFromTeamOutputRef(teamRef)
		teamRefs = append(teamRefs, teamRef)
		outputRefs = append(outputRefs, outRef)
		outputs = append(outputs, executionOutputFromOutputRef(
			outRef,
			contract.Summary,
			contract.ExecutionMode == TeamLeadExecutionModeContinuityResume,
		))
	}
	if len(outputRefs) == 0 {
		return nil
	}
	return &protocol.ExecutionSummary{
		Title:          firstNonEmptyString(response.Headline, contract.OwnerLabel, "Team Lead execution guidance"),
		Summary:        firstNonEmptyString(contract.Summary, response.Summary),
		Status:         status,
		RunID:          runID,
		ProofID:        contractID,
		OutputRefs:     outputRefs,
		Outputs:        outputs,
		TeamOutputRefs: teamRefs,
	}
}

func buildGroupBroadcastExecutionSummary(group *CollaborationGroup, message, auditEventID, runID string, subjects []string) *protocol.ExecutionSummary {
	if group == nil {
		return nil
	}
	targets := append([]string{group.ID}, group.TeamIDs...)
	outputRefs := make([]protocol.OutputRef, 0, len(targets))
	outputs := make([]protocol.ExecutionOutput, 0, len(targets))
	teamRefs := make([]protocol.TeamOutputRef, 0, len(targets))

	for index, targetID := range targets {
		trimmed := strings.TrimSpace(targetID)
		if trimmed == "" {
			continue
		}
		role := "primary"
		kind := "group_broadcast"
		label := fmt.Sprintf("Group broadcast to %s", firstNonEmptyString(group.Name, group.ID))
		if index > 0 {
			role = "supporting"
			kind = "team_signal"
			label = fmt.Sprintf("Team command fanout: %s", trimmed)
		}
		refID := fmt.Sprintf("%s:%s", firstNonEmptyString(auditEventID, group.ID), trimmed)
		storageRef := fmt.Sprintf("nats://groups/%s/%s", group.ID, trimmed)
		openURL := "/deployments?group=" + url.QueryEscape(group.ID)
		proof := &protocol.OutputProofEnvelope{
			ProofID:          firstNonEmptyString(auditEventID, refID),
			OutputRefID:      refID,
			StorageRef:       storageRef,
			SourceRunID:      runID,
			SourceContractID: auditEventID,
			ExecutionStatus:  "queued",
			CreatedAt:        time.Now().UTC().Format(time.RFC3339),
			ReadbackStatus:   proofStatusVerified,
		}
		teamRef := protocol.TeamOutputRef{
			OutputID:   refID,
			TeamID:     trimmed,
			RunID:      runID,
			Kind:       kind,
			Label:      label,
			StorageRef: storageRef,
			OpenURL:    openURL,
			RevealURL:  openURL,
			ProofRef:   firstNonEmptyString(auditEventID, refID),
			ContractID: auditEventID,
			ProofID:    firstNonEmptyString(auditEventID, refID),
			PackageID:  group.ID,
			Role:       role,
			Status:     "queued",
			Proof:      proof,
		}
		outRef := protocol.OutputRefFromTeamOutputRef(teamRef)
		teamRefs = append(teamRefs, teamRef)
		outputRefs = append(outputRefs, outRef)
		outputs = append(outputs, executionOutputFromOutputRef(outRef, strings.TrimSpace(message), false))
	}

	return &protocol.ExecutionSummary{
		Title:          fmt.Sprintf("Broadcast to %s", firstNonEmptyString(group.Name, group.ID)),
		Summary:        fmt.Sprintf("Queued broadcast across %d subject(s).", len(subjects)),
		Status:         "queued",
		RunID:          runID,
		ProofID:        auditEventID,
		OutputRefs:     outputRefs,
		Outputs:        outputs,
		TeamOutputRefs: teamRefs,
	}
}

func decodeArtifactMetadata(raw json.RawMessage) map[string]any {
	if len(raw) == 0 {
		return map[string]any{}
	}
	out := map[string]any{}
	_ = json.Unmarshal(raw, &out)
	return out
}

func uuidPtrString(id *uuid.UUID) string {
	if id == nil {
		return ""
	}
	return id.String()
}

package server

import (
	"crypto/sha256"
	"fmt"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/mycelis/core/pkg/protocol"
)

func buildTeamWorkConfirmExecutionSummary(planned []protocol.PlannedToolCall, teamID, workItemID, runID, proofID string) *protocol.ExecutionSummary {
	if len(planned) == 0 {
		return nil
	}
	fileRefs, fileOutputs, artifacts := confirmActionOutputsFromPlannedCalls(planned, runID, proofID)
	teamRefs := make([]protocol.TeamOutputRef, 0, len(planned))
	outputRefs := make([]protocol.OutputRef, 0, len(planned))
	outputs := make([]protocol.ExecutionOutput, 0, len(planned))

	outputRefs = append(outputRefs, fileRefs...)
	outputs = append(outputs, fileOutputs...)
	for _, ref := range fileRefs {
		teamRefs = append(teamRefs, protocol.TeamOutputRef{
			OutputID:    ref.ID,
			ArtifactID:  ref.ArtifactID,
			TeamID:      teamID,
			WorkItemID:  workItemID,
			RunID:       runID,
			Kind:        ref.Kind,
			Label:       ref.Label,
			StorageRef:  ref.StorageRef,
			OpenURL:     ref.OpenURL,
			RevealURL:   ref.RevealURL,
			ProofRef:    ref.ProofArtifactID,
			ContractID:  proofID,
			ProofID:     ref.ProofArtifactID,
			PackageID:   firstNonEmptyString(ref.PackageID, teamID),
			Role:        ref.Role,
			Status:      ref.Status,
			ContentType: ref.ContentType,
			Bytes:       ref.Bytes,
			SHA256:      ref.SHA256,
			Proof:       ref.Proof,
		})
	}

	for _, call := range planned {
		toolName := strings.TrimSpace(call.Name)
		if toolName == "" || toolName == "write_file" {
			continue
		}
		role := "primary"
		if len(outputRefs) > 0 {
			role = "supporting"
		}
		teamRef, execOut, outRef := teamWorkOutputFromPlannedCall(call, teamID, workItemID, runID, proofID, role)
		teamRefs = append(teamRefs, teamRef)
		outputs = append(outputs, execOut)
		outputRefs = append(outputRefs, outRef)
	}

	if len(outputRefs) == 0 {
		return nil
	}
	primary := outputRefs[0]
	return &protocol.ExecutionSummary{
		Title:          firstNonEmptyString(primary.Label, "Confirmed team work"),
		Summary:        fmt.Sprintf("Confirmed execution recorded %d output(s).", len(outputRefs)),
		Status:         firstNonEmptyString(primary.Status, "verified"),
		RunID:          runID,
		ProofID:        proofID,
		OutputRefs:     outputRefs,
		Outputs:        outputs,
		ArtifactRefs:   artifacts,
		TeamOutputRefs: teamRefs,
	}
}

func teamWorkOutputFromPlannedCall(call protocol.PlannedToolCall, teamID, workItemID, runID, proofID, role string) (protocol.TeamOutputRef, protocol.ExecutionOutput, protocol.OutputRef) {
	toolName := firstNonEmptyString(strings.TrimSpace(call.Name), "tool")
	resolvedTeamID := firstNonEmptyString(teamID, firstNonEmptyString(call.Arguments["team_id"], call.Arguments["id"], call.Arguments["name"]))
	label := firstNonEmptyString(call.Arguments["title"], call.Arguments["name"], call.Arguments["goal"], call.Arguments["objective"], toolName)
	kind := teamWorkKindForTool(toolName)
	storageRef := teamWorkStorageRefForTool(toolName, resolvedTeamID, workItemID, call.Arguments)
	openURL := teamWorkOpenURLForTool(toolName, resolvedTeamID, storageRef)
	refID := fmt.Sprintf("%s:%s:%s", firstNonEmptyString(runID, proofID, "confirm"), toolName, firstNonEmptyString(resolvedTeamID, label))
	sum := sha256.Sum256([]byte(storageRef + "|" + label))
	checksum := fmt.Sprintf("%x", sum)
	proof := &protocol.OutputProofEnvelope{
		ProofID:            firstNonEmptyString(proofID, refID),
		OutputRefID:        refID,
		StorageRef:         storageRef,
		SourceRunID:        runID,
		SourceContractID:   proofID,
		ExecutionStatus:    proofStatusVerified,
		PathBoundaryStatus: proofStatusVerified,
		Checksum:           checksum,
		ChecksumAlgorithm:  "sha256",
		Bytes:              len(label),
		ContentType:        "application/json",
		CreatedAt:          time.Now().UTC().Format(time.RFC3339),
		ReadbackStatus:     proofStatusVerified,
	}
	teamRef := protocol.TeamOutputRef{
		OutputID:    refID,
		TeamID:      resolvedTeamID,
		WorkItemID:  workItemID,
		RunID:       runID,
		Kind:        kind,
		Label:       label,
		StorageRef:  storageRef,
		OpenURL:     openURL,
		RevealURL:   openURL,
		ProofRef:    firstNonEmptyString(proofID, refID),
		ContractID:  proofID,
		ProofID:     firstNonEmptyString(proofID, refID),
		PackageID:   resolvedTeamID,
		Role:        firstNonEmptyString(role, "primary"),
		Status:      "recorded",
		ContentType: "application/json",
		Bytes:       len(label),
		SHA256:      checksum,
		Proof:       proof,
	}
	outRef := protocol.OutputRefFromTeamOutputRef(teamRef)
	execOut := executionOutputFromOutputRef(outRef, fmt.Sprintf("%s completed for %s.", toolName, firstNonEmptyString(resolvedTeamID, label)), true)
	return teamRef, execOut, outRef
}

func executionOutputFromOutputRef(outRef protocol.OutputRef, summary string, retained bool) protocol.ExecutionOutput {
	return protocol.ExecutionOutput{
		ID:               outRef.ID,
		ArtifactID:       outRef.ArtifactID,
		ProofArtifactID:  outRef.ProofArtifactID,
		Kind:             outRef.Kind,
		Title:            outRef.Label,
		Summary:          summary,
		StorageRef:       outRef.StorageRef,
		Href:             outRef.OpenURL,
		OpenURL:          outRef.OpenURL,
		RevealURL:        outRef.RevealURL,
		SourceRunID:      outRef.SourceRunID,
		SourceContractID: outRef.SourceContractID,
		PackageID:        outRef.PackageID,
		Role:             outRef.Role,
		Status:           outRef.Status,
		ContentType:      outRef.ContentType,
		Bytes:            outRef.Bytes,
		SHA256:           outRef.SHA256,
		Retained:         retained,
		Proof:            outRef.Proof,
	}
}

func outputRefsFromTeamOutputRefs(teamRefs []protocol.TeamOutputRef) []protocol.OutputRef {
	if len(teamRefs) == 0 {
		return nil
	}
	refs := make([]protocol.OutputRef, 0, len(teamRefs))
	for index, item := range teamRefs {
		if strings.TrimSpace(item.Role) == "" {
			if index == 0 {
				item.Role = "primary"
			} else {
				item.Role = "supporting"
			}
		}
		refs = append(refs, protocol.OutputRefFromTeamOutputRef(item))
	}
	return refs
}

func teamWorkKindForTool(toolName string) string {
	switch strings.TrimSpace(toolName) {
	case "create_team":
		return "team"
	case "delegate", "delegate_task":
		return "team_Ask"
	case "store_artifact":
		return "document"
	case "generate_image", "save_cached_image":
		return "image"
	default:
		return "tool_result"
	}
}

func teamWorkStorageRefForTool(toolName, teamID, workItemID string, args map[string]any) string {
	if rawPath := firstNonEmptyString(args["path"], args["file_path"], args["target_path"]); rawPath != "" {
		return rawPath
	}
	if teamID != "" && workItemID != "" {
		return fmt.Sprintf("team://%s/work/%s", teamID, workItemID)
	}
	if teamID != "" {
		return fmt.Sprintf("team://%s/%s", teamID, firstNonEmptyString(toolName, "work"))
	}
	return fmt.Sprintf("tool://%s", firstNonEmptyString(toolName, "action"))
}

func teamWorkOpenURLForTool(toolName, teamID, storageRef string) string {
	if strings.TrimSpace(toolName) == "create_team" || teamID != "" {
		return "/deployments?team=" + url.QueryEscape(firstNonEmptyString(teamID, storageRef))
	}
	return workspaceFileOpenURL(storageRef)
}

func outputKindForWrittenFile(filePath, contentType string) string {
	switch strings.ToLower(path.Ext(strings.TrimSpace(filePath))) {
	case ".go", ".ts", ".tsx", ".js", ".jsx", ".py", ".css", ".html", ".sql":
		return "code"
	case ".json", ".yaml", ".yml", ".csv":
		return "data"
	case ".md", ".txt":
		return "document"
	default:
		return outputKindForContentType(contentType)
	}
}

func contentTypeForPath(filePath string) string {
	switch strings.ToLower(path.Ext(strings.TrimSpace(filePath))) {
	case ".md":
		return "text/markdown"
	case ".json":
		return "application/json"
	case ".html":
		return "text/html"
	default:
		return "text/plain"
	}
}

func workspaceFileOpenURL(filePath string) string {
	if trimmed := strings.TrimSpace(filePath); trimmed != "" {
		return "/resources?tab=files&path=" + url.QueryEscape(trimmed)
	}
	return ""
}

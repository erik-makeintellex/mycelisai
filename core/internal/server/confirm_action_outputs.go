package server

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/mycelis/core/pkg/protocol"
)

func outputRefsFromPackageSummary(summary *packageOutputSummary) []protocol.OutputRef {
	if summary == nil {
		return nil
	}
	refs := make([]protocol.OutputRef, 0, len(summary.Files))
	for index, file := range summary.Files {
		role := packageOutputFileRole(summary.Files, index)
		refID := fmt.Sprintf("%s:%s", summary.PackageID, path.Base(file.Path))
		refs = append(refs, protocol.OutputRef{
			ID:               refID,
			ArtifactID:       file.ArtifactID,
			Kind:             outputKindForContentType(file.ContentType),
			Label:            firstNonEmptyString(file.Title, path.Base(file.Path)),
			StorageRef:       file.Path,
			OpenURL:          artifactDownloadURL(file.ArtifactID),
			SourceRunID:      summary.RunID,
			SourceContractID: summary.ProofID,
			ProofArtifactID:  proofArtifactIDForRole(summary, role),
			PackageID:        summary.PackageID,
			Role:             role,
			Status:           "final",
			ContentType:      file.ContentType,
			Bytes:            file.Bytes,
			SHA256:           file.SHA256,
			Proof:            file.Proof,
			ProofEnvelope:    file.Proof,
		})
	}
	return refs
}

func deliveryPackageFromPackageSummary(summary *packageOutputSummary) protocol.DeliveryPackageSummary {
	if summary == nil {
		return protocol.DeliveryPackageSummary{}
	}
	refs := summary.OutputRefs
	if len(refs) == 0 {
		refs = outputRefsFromPackageSummary(summary)
	}
	primaryRef, proofRef := outputRefPointers(refs)
	return protocol.DeliveryPackageSummary{
		PackageID:        summary.PackageID,
		RunID:            summary.RunID,
		Title:            "Codex task bundle",
		Summary:          fmt.Sprintf("Codex task bundle created under %s with proof at %s.", summary.BasePath, summary.ProofPath),
		PackageKind:      summary.PackageKind,
		Status:           "verified",
		PrimaryOutputRef: primaryRef,
		Items:            refs,
		NextActions: []string{
			"Open the implementation task.",
			"Run the validation checklist.",
			"Review package proof before promotion.",
		},
		ProofRef:    proofRef,
		InspectRefs: refs,
	}
}

func executionSummaryFromPackageSummary(summary *packageOutputSummary) protocol.ExecutionSummary {
	if summary == nil {
		return protocol.ExecutionSummary{}
	}
	deliveryPackage := deliveryPackageFromPackageSummary(summary)
	return protocol.ExecutionSummary{
		Title:           "Codex task bundle",
		Summary:         deliveryPackage.Summary,
		Status:          "verified",
		RunID:           summary.RunID,
		ProofID:         summary.ProofID,
		OutputRefs:      deliveryPackage.Items,
		DeliveryPackage: &deliveryPackage,
	}
}

func confirmActionOutputsFromPlannedCalls(planned []protocol.PlannedToolCall, runID, proofID string) ([]protocol.OutputRef, []protocol.ExecutionOutput, []protocol.ChatArtifactRef) {
	if len(planned) == 0 {
		return nil, nil, nil
	}
	outputRefs := make([]protocol.OutputRef, 0, len(planned))
	outputs := make([]protocol.ExecutionOutput, 0, len(planned))
	artifacts := make([]protocol.ChatArtifactRef, 0, len(planned))

	for _, call := range planned {
		toolName := strings.TrimSpace(call.Name)
		if toolName != "write_file" {
			continue
		}
		file, artifact, ok := writtenFileFromPlannedCall(call)
		if !ok {
			continue
		}
		role := "primary"
		if len(outputRefs) > 0 {
			role = "supporting"
		}
		ref := outputRefFromWrittenFile(file, runID, proofID, role)
		outputRefs = append(outputRefs, ref)
		outputs = append(outputs, protocol.ExecutionOutput{
			ID:               ref.ID,
			ArtifactID:       ref.ArtifactID,
			ProofArtifactID:  ref.ProofArtifactID,
			Kind:             ref.Kind,
			Title:            ref.Label,
			Summary:          fmt.Sprintf("Wrote %s (%d bytes).", ref.StorageRef, ref.Bytes),
			StorageRef:       ref.StorageRef,
			Href:             ref.OpenURL,
			OpenURL:          ref.OpenURL,
			RevealURL:        ref.RevealURL,
			SourceRunID:      ref.SourceRunID,
			SourceContractID: ref.SourceContractID,
			Role:             ref.Role,
			Status:           ref.Status,
			ContentType:      ref.ContentType,
			Bytes:            ref.Bytes,
			SHA256:           ref.SHA256,
			Retained:         true,
			Proof:            ref.Proof,
		})
		artifacts = append(artifacts, artifact)
	}
	return outputRefs, outputs, artifacts
}

func executionSummaryFromPlannedToolCalls(planned []protocol.PlannedToolCall, runID, proofID string) *protocol.ExecutionSummary {
	outputRefs, outputs, artifacts := confirmActionOutputsFromPlannedCalls(planned, runID, proofID)
	if len(outputRefs) == 0 {
		return nil
	}
	primary := outputRefs[0]
	title := firstNonEmptyString(primary.Label, "Confirmed workspace output")
	summaryText := fmt.Sprintf("Confirmed file output written to %s.", primary.StorageRef)
	if len(outputRefs) > 1 {
		title = "Confirmed workspace outputs"
		summaryText = fmt.Sprintf("%d confirmed file outputs written to workspace.", len(outputRefs))
	}
	return &protocol.ExecutionSummary{
		Title:        title,
		Summary:      summaryText,
		Status:       firstNonEmptyString(primary.Status, "verified"),
		RunID:        runID,
		ProofID:      proofID,
		OutputRefs:   outputRefs,
		Outputs:      outputs,
		ArtifactRefs: artifacts,
	}
}

func writtenFileFromPlannedCall(call protocol.PlannedToolCall) (packageOutputFile, protocol.ChatArtifactRef, bool) {
	rawPath := firstNonEmptyString(call.Arguments["path"], call.Arguments["file_path"], call.Arguments["target_path"], call.Arguments["filename"])
	if rawPath == "" {
		return packageOutputFile{}, protocol.ChatArtifactRef{}, false
	}
	normalizedPath := filepath.ToSlash(path.Clean(strings.TrimPrefix(strings.ReplaceAll(rawPath, "\\", "/"), "./")))
	title := firstNonEmptyString(call.Arguments["title"], call.Arguments["artifact_role"], path.Base(normalizedPath), normalizedPath)
	contentType := firstNonEmptyString(call.Arguments["content_type"], contentTypeForPath(normalizedPath))
	data := []byte(firstNonEmptyString(call.Arguments["content"], call.Arguments["body"], call.Arguments["text"]))
	if absPath, err := workspacePathForServer(normalizedPath); err == nil {
		if diskData, readErr := os.ReadFile(absPath); readErr == nil {
			data = diskData
		}
	}
	file := packageOutputFileFromBytes(normalizedPath, title, data)
	file.ContentType = contentType
	file.ArtifactID = firstNonEmptyString(call.Arguments["artifact_id"], call.Arguments["id"])
	openURL := firstNonEmptyString(artifactDownloadURL(file.ArtifactID), workspaceFileOpenURL(normalizedPath))
	return file, protocol.ChatArtifactRef{
		ID:          file.ArtifactID,
		Type:        outputKindForWrittenFile(normalizedPath, contentType),
		Title:       title,
		ContentType: contentType,
		Content:     string(data),
		URL:         openURL,
		SavedPath:   normalizedPath,
	}, true
}

func outputRefFromWrittenFile(file packageOutputFile, runID, proofID, role string) protocol.OutputRef {
	refID := fmt.Sprintf("%s:%s", firstNonEmptyString(runID, proofID, "write_file"), path.Base(file.Path))
	proof := proofEnvelopeForWrittenFile(file, runID, proofID, refID)
	status := "verified"
	if proof != nil && proof.ReadbackStatus != "" && proof.ReadbackStatus != proofStatusVerified {
		status = proof.ReadbackStatus
	}
	openURL := firstNonEmptyString(artifactDownloadURL(file.ArtifactID), workspaceFileOpenURL(file.Path))
	return protocol.NormalizeOutputRef(protocol.OutputRef{
		ID:               refID,
		ArtifactID:       file.ArtifactID,
		Kind:             outputKindForWrittenFile(file.Path, file.ContentType),
		Label:            firstNonEmptyString(file.Title, path.Base(file.Path), file.Path),
		StorageRef:       file.Path,
		OpenURL:          openURL,
		RevealURL:        workspaceFileOpenURL(file.Path),
		SourceRunID:      runID,
		SourceContractID: proofID,
		ProofArtifactID:  firstNonEmptyString(file.ArtifactID, proofID),
		Role:             firstNonEmptyString(role, "primary"),
		Status:           status,
		ContentType:      file.ContentType,
		Bytes:            file.Bytes,
		SHA256:           file.SHA256,
		Proof:            proof,
		ProofEnvelope:    proof,
	})
}

func proofEnvelopeForWrittenFile(file packageOutputFile, runID, proofID, refID string) *protocol.OutputProofEnvelope {
	boundary, readback, hint := packageFileReadback(file)
	execStatus := proofStatusVerified
	if readback != proofStatusVerified {
		execStatus = proofStatusDegraded
	}
	return &protocol.OutputProofEnvelope{
		ProofID:            firstNonEmptyString(file.ArtifactID, proofID, refID),
		OutputRefID:        refID,
		ArtifactID:         file.ArtifactID,
		StorageRef:         file.Path,
		SourceRunID:        runID,
		SourceContractID:   proofID,
		ExecutionStatus:    execStatus,
		PathBoundaryStatus: boundary,
		Checksum:           file.SHA256,
		ChecksumAlgorithm:  "sha256",
		Bytes:              file.Bytes,
		ContentType:        file.ContentType,
		CreatedAt:          time.Now().UTC().Format(time.RFC3339),
		ReadbackStatus:     readback,
		RecoveryHint:       hint,
	}
}

func outputRefPointers(refs []protocol.OutputRef) (*protocol.OutputRef, *protocol.OutputRef) {
	var primaryRef, proofRef *protocol.OutputRef
	for _, ref := range refs {
		current := ref
		if ref.Role == "primary" && primaryRef == nil {
			primaryRef = &current
		} else if ref.Role == "proof" && proofRef == nil {
			proofRef = &current
		}
	}
	return primaryRef, proofRef
}

func isPackageProofFile(filePath, title string) bool {
	return strings.HasSuffix(strings.ToLower(strings.TrimSpace(filePath)), "proof.md") ||
		strings.Contains(strings.ToLower(strings.TrimSpace(title)), "proof")
}

func outputKindForContentType(contentType string) string {
	lower := strings.ToLower(strings.TrimSpace(contentType))
	switch {
	case strings.Contains(lower, "markdown"), strings.HasPrefix(lower, "text/"):
		return "document"
	case strings.Contains(lower, "json"):
		return "data"
	default:
		return "file"
	}
}

func proofArtifactIDForRole(summary *packageOutputSummary, role string) string {
	if summary == nil || role == "proof" {
		return ""
	}
	return packageProofArtifactID(summary)
}

func artifactDownloadURL(artifactID string) string {
	if strings.TrimSpace(artifactID) == "" {
		return ""
	}
	return "/api/v1/artifacts/" + artifactID + "/download"
}

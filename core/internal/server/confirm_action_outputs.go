package server

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
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
		role, refID := packageOutputFileRole(summary.Files, index), fmt.Sprintf("%s:%s", summary.PackageID, path.Base(file.Path))
		refs = append(refs, protocol.OutputRef{
			ID: refID, ArtifactID: file.ArtifactID, Kind: outputKindForContentType(file.ContentType),
			Label: firstNonEmptyString(file.Title, path.Base(file.Path)), StorageRef: file.Path,
			OpenURL: artifactDownloadURL(file.ArtifactID), SourceRunID: summary.RunID, SourceContractID: summary.ProofID,
			ProofArtifactID: proofArtifactIDForRole(summary, role), PackageID: summary.PackageID, Role: role,
			Status: "final", ContentType: file.ContentType, Bytes: file.Bytes, SHA256: file.SHA256, Proof: file.Proof, ProofEnvelope: file.Proof,
		})
	}
	return refs
}

func packageSummaryTitleAndMessage(summary *packageOutputSummary) (string, string) {
	if summary == nil {
		return "Deliverable package", ""
	}
	if summary.PackageKind == "" || summary.PackageKind == "codex_task_bundle" {
		return firstNonEmptyString(summary.TeamName, "Codex task bundle"),
			fmt.Sprintf("Codex task bundle created under %s with proof at %s.", summary.BasePath, summary.ProofPath)
	}
	kindLabel := strings.TrimSuffix(deliverablePackageProofTitle(summary.PackageKind), " Proof")
	title := firstNonEmptyString(summary.TeamName, kindLabel, titleCaseTeamID(summary.PackageSlug), "Deliverable package")
	if summary.TeamName != "" {
		return title, fmt.Sprintf("Team %s activated and %s created under %s (proof verified at %s).", summary.TeamName, strings.ToLower(kindLabel), summary.BasePath, summary.ProofPath)
	}
	return title, fmt.Sprintf("%s created under %s with proof at %s.", kindLabel, summary.BasePath, summary.ProofPath)
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
	title, summaryMsg := packageSummaryTitleAndMessage(summary)
	return protocol.DeliveryPackageSummary{
		PackageID: summary.PackageID, RunID: summary.RunID, TeamID: summary.TeamID, TeamName: summary.TeamName, Title: title,
		Summary:     summaryMsg,
		PackageKind: summary.PackageKind, Status: "verified", PrimaryOutputRef: primaryRef, Items: refs,
		NextActions: []string{"Inspect the primary deliverable.", "Review supporting files and channel outputs.", "Verify the SHA-256 package proof."},
		ProofRef:    proofRef, InspectRefs: refs,
	}
}

func executionSummaryFromPackageSummary(summary *packageOutputSummary) protocol.ExecutionSummary {
	if summary == nil {
		return protocol.ExecutionSummary{}
	}
	dp := deliveryPackageFromPackageSummary(summary)
	return protocol.ExecutionSummary{Title: dp.Title, Summary: dp.Summary, Status: "verified", RunID: summary.RunID, ProofID: summary.ProofID, OutputRefs: dp.Items, DeliveryPackage: &dp}
}

func confirmActionOutputsFromPlannedCalls(planned []protocol.PlannedToolCall, runID, proofID string) ([]protocol.OutputRef, []protocol.ExecutionOutput, []protocol.ChatArtifactRef) {
	if len(planned) == 0 {
		return nil, nil, nil
	}
	outputRefs, outputs, artifacts := make([]protocol.OutputRef, 0, len(planned)), make([]protocol.ExecutionOutput, 0, len(planned)), make([]protocol.ChatArtifactRef, 0, len(planned))
	for idx, call := range planned {
		toolName, role := strings.TrimSpace(call.Name), "primary"
		if len(outputRefs) > 0 {
			role = "supporting"
		}
		if toolName == "write_file" {
			if file, artifact, ok := writtenFileFromPlannedCall(call); ok {
				ref := outputRefFromWrittenFile(file, runID, proofID, role)
				outputRefs = append(outputRefs, ref)
				outputs = append(outputs, executionOutputFromOutputRef(ref, fmt.Sprintf("Wrote %s (%d bytes).", ref.StorageRef, ref.Bytes), true))
				artifacts = append(artifacts, artifact)
			}
			continue
		}
		if ref, out, artifact, ok := extractOutputsFromToolResult(call.Arguments, runID, proofID, role); ok {
			outputRefs, outputs, artifacts = append(outputRefs, ref), append(outputs, out), append(artifacts, artifact)
			planned[idx].Name = ""
		}
	}
	return outputRefs, outputs, artifacts
}

func extractOutputsFromToolResult(raw any, runID, proofID, role string) (protocol.OutputRef, protocol.ExecutionOutput, protocol.ChatArtifactRef, bool) {
	topMap, art := unwrapToolArtifactMap(raw)
	if len(art) == 0 {
		return protocol.OutputRef{}, protocol.ExecutionOutput{}, protocol.ChatArtifactRef{}, false
	}
	artType, savedPath := strings.ToLower(firstNonEmptyString(art["type"], topMap["type"])), filepath.ToSlash(strings.TrimSpace(firstNonEmptyString(art["saved_path"], topMap["saved_path"])))
	if artType != "image" && savedPath == "" {
		return protocol.OutputRef{}, protocol.ExecutionOutput{}, protocol.ChatArtifactRef{}, false
	}
	id, title := firstNonEmptyString(art["id"], art["artifact_id"], topMap["artifact_id"], topMap["id"]), firstNonEmptyString(art["title"], topMap["title"], path.Base(savedPath), "Generated image")
	contentType, content := firstNonEmptyString(art["content_type"], topMap["content_type"], "image/png"), firstNonEmptyString(art["content"], topMap["content"])
	sizeBytes, sumHex := resolveMediaDigest(savedPath, content, intFromAny(art["bytes"], topMap["bytes"]), firstNonEmptyString(art["sha256"], topMap["sha256"]))
	storageRef := firstNonEmptyString(savedPath, "artifact://"+firstNonEmptyString(id, "cached-image"))
	cached := art["cached"] == true || topMap["cached"] == true || savedPath == ""
	status, readback, boundary := "cached", "cached", "local_only"
	if savedPath != "" && sumHex != "" {
		status, readback, boundary, cached = proofStatusVerified, proofStatusVerified, proofStatusVerified, false
	}
	refID, openURL := firstNonEmptyString(id, fmt.Sprintf("%s:%s", firstNonEmptyString(runID, proofID, "media"), path.Base(storageRef))), firstNonEmptyString(artifactDownloadURL(id), workspaceFileOpenURL(savedPath))
	proof := &protocol.OutputProofEnvelope{
		ProofID: firstNonEmptyString(id, proofID, refID), OutputRefID: refID, ArtifactID: id, StorageRef: storageRef,
		SourceRunID: runID, SourceContractID: proofID, ExecutionStatus: status, PathBoundaryStatus: boundary,
		Checksum: sumHex, ChecksumAlgorithm: "sha256", Bytes: sizeBytes, ContentType: contentType,
		CreatedAt: time.Now().UTC().Format(time.RFC3339), ReadbackStatus: readback,
	}
	ref := protocol.NormalizeOutputRef(protocol.OutputRef{
		ID: refID, ArtifactID: id, Kind: "image", Label: title, StorageRef: storageRef, OpenURL: openURL,
		RevealURL: firstNonEmptyString(workspaceFileOpenURL(savedPath), openURL), SourceRunID: runID, SourceContractID: proofID,
		ProofArtifactID: firstNonEmptyString(id, proofID), Role: firstNonEmptyString(role, "primary"), Status: status,
		ContentType: contentType, Bytes: sizeBytes, SHA256: sumHex, Proof: proof, ProofEnvelope: proof,
	})
	artifact := protocol.ChatArtifactRef{
		ID: id, Type: "image", Title: title, ContentType: contentType, Content: content, URL: openURL,
		Cached: cached, ExpiresAt: firstNonEmptyString(art["expires_at"], topMap["expires_at"]), SavedPath: savedPath,
	}
	return ref, executionOutputFromOutputRef(ref, fmt.Sprintf("Media output %s (%d bytes).", storageRef, sizeBytes), !cached), artifact, true
}

func unwrapToolArtifactMap(raw any) (map[string]any, map[string]any) {
	var top map[string]any
	switch v := raw.(type) {
	case string:
		_ = json.Unmarshal([]byte(v), &top)
	case []byte:
		_ = json.Unmarshal(v, &top)
	case map[string]any:
		top = v
	}
	if len(top) == 0 {
		return nil, nil
	}
	if art, ok := top["artifact"].(map[string]any); ok {
		return top, art
	}
	for _, key := range []string{"result", "tool_result", "output"} {
		if top[key] != nil {
			if _, art := unwrapToolArtifactMap(top[key]); len(art) > 0 {
				return top, art
			}
		}
	}
	if firstNonEmptyString(top["saved_path"]) != "" || strings.EqualFold(firstNonEmptyString(top["type"]), "image") {
		return top, top
	}
	return top, nil
}

func resolveMediaDigest(savedPath, content string, sizeBytes int, sumHex string) (int, string) {
	if (sizeBytes == 0 || sumHex == "") && savedPath != "" {
		if abs, err := workspacePathForServer(savedPath); err == nil {
			if disk, err := os.ReadFile(abs); err == nil {
				sizeBytes, sumHex = len(disk), fmt.Sprintf("%x", sha256.Sum256(disk))
			}
		}
	}
	if (sizeBytes == 0 || sumHex == "") && content != "" {
		rawBytes, err := base64.StdEncoding.DecodeString(content)
		if err != nil {
			rawBytes = []byte(content)
		}
		sizeBytes, sumHex = len(rawBytes), fmt.Sprintf("%x", sha256.Sum256(rawBytes))
	}
	return sizeBytes, sumHex
}

func intFromAny(values ...any) int {
	for _, v := range values {
		switch n := v.(type) {
		case int:
			return n
		case int64:
			return int(n)
		case float64:
			return int(n)
		}
	}
	return 0
}

func executionSummaryFromPlannedToolCalls(planned []protocol.PlannedToolCall, runID, proofID string) *protocol.ExecutionSummary {
	outputRefs, outputs, artifacts := confirmActionOutputsFromPlannedCalls(planned, runID, proofID)
	if len(outputRefs) == 0 {
		return nil
	}
	primary := outputRefs[0]
	title, summaryText := firstNonEmptyString(primary.Label, "Confirmed workspace output"), fmt.Sprintf("Confirmed file output written to %s.", primary.StorageRef)
	if len(outputRefs) > 1 {
		title, summaryText = "Confirmed workspace outputs", fmt.Sprintf("%d confirmed file outputs written to workspace.", len(outputRefs))
	}
	return &protocol.ExecutionSummary{Title: title, Summary: summaryText, Status: firstNonEmptyString(primary.Status, "verified"), RunID: runID, ProofID: proofID, OutputRefs: outputRefs, Outputs: outputs, ArtifactRefs: artifacts}
}

func writtenFileFromPlannedCall(call protocol.PlannedToolCall) (packageOutputFile, protocol.ChatArtifactRef, bool) {
	rawPath := firstNonEmptyString(call.Arguments["path"], call.Arguments["file_path"], call.Arguments["target_path"], call.Arguments["filename"])
	if rawPath == "" {
		return packageOutputFile{}, protocol.ChatArtifactRef{}, false
	}
	normalizedPath := filepath.ToSlash(path.Clean(strings.TrimPrefix(strings.ReplaceAll(rawPath, "\\", "/"), "./")))
	title := firstNonEmptyString(call.Arguments["title"], call.Arguments["artifact_role"], path.Base(normalizedPath), normalizedPath)
	contentType, data := firstNonEmptyString(call.Arguments["content_type"], contentTypeForPath(normalizedPath)), []byte(firstNonEmptyString(call.Arguments["content"], call.Arguments["body"], call.Arguments["text"]))
	if absPath, err := workspacePathForServer(normalizedPath); err == nil {
		if diskData, readErr := os.ReadFile(absPath); readErr == nil {
			data = diskData
		}
	}
	file := packageOutputFileFromBytes(normalizedPath, title, data)
	file.ContentType, file.ArtifactID = contentType, firstNonEmptyString(call.Arguments["artifact_id"], call.Arguments["id"])
	openURL := firstNonEmptyString(artifactDownloadURL(file.ArtifactID), workspaceFileOpenURL(normalizedPath))
	return file, protocol.ChatArtifactRef{ID: file.ArtifactID, Type: outputKindForWrittenFile(normalizedPath, contentType), Title: title, ContentType: contentType, Content: string(data), URL: openURL, SavedPath: normalizedPath}, true
}

func outputRefFromWrittenFile(file packageOutputFile, runID, proofID, role string) protocol.OutputRef {
	refID := fmt.Sprintf("%s:%s", firstNonEmptyString(runID, proofID, "write_file"), path.Base(file.Path))
	proof, status := proofEnvelopeForWrittenFile(file, runID, proofID, refID), "verified"
	if proof != nil && proof.ReadbackStatus != "" && proof.ReadbackStatus != proofStatusVerified {
		status = proof.ReadbackStatus
	}
	openURL := firstNonEmptyString(artifactDownloadURL(file.ArtifactID), workspaceFileOpenURL(file.Path))
	return protocol.NormalizeOutputRef(protocol.OutputRef{ID: refID, ArtifactID: file.ArtifactID, Kind: outputKindForWrittenFile(file.Path, file.ContentType), Label: firstNonEmptyString(file.Title, path.Base(file.Path), file.Path), StorageRef: file.Path, OpenURL: openURL, RevealURL: workspaceFileOpenURL(file.Path), SourceRunID: runID, SourceContractID: proofID, ProofArtifactID: firstNonEmptyString(file.ArtifactID, proofID), Role: firstNonEmptyString(role, "primary"), Status: status, ContentType: file.ContentType, Bytes: file.Bytes, SHA256: file.SHA256, Proof: proof, ProofEnvelope: proof})
}

func proofEnvelopeForWrittenFile(file packageOutputFile, runID, proofID, refID string) *protocol.OutputProofEnvelope {
	boundary, readback, hint := packageFileReadback(file)
	execStatus := proofStatusVerified
	if readback != proofStatusVerified {
		execStatus = proofStatusDegraded
	}
	return &protocol.OutputProofEnvelope{ProofID: firstNonEmptyString(file.ArtifactID, proofID, refID), OutputRefID: refID, ArtifactID: file.ArtifactID, StorageRef: file.Path, SourceRunID: runID, SourceContractID: proofID, ExecutionStatus: execStatus, PathBoundaryStatus: boundary, Checksum: file.SHA256, ChecksumAlgorithm: "sha256", Bytes: file.Bytes, ContentType: file.ContentType, CreatedAt: time.Now().UTC().Format(time.RFC3339), ReadbackStatus: readback, RecoveryHint: hint}
}

func outputRefPointers(refs []protocol.OutputRef) (primaryRef, proofRef *protocol.OutputRef) {
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
	return strings.HasSuffix(strings.ToLower(strings.TrimSpace(filePath)), "proof.md") || strings.Contains(strings.ToLower(strings.TrimSpace(title)), "proof")
}

func outputKindForContentType(contentType string) string {
	lower := strings.ToLower(strings.TrimSpace(contentType))
	if strings.Contains(lower, "markdown") || strings.HasPrefix(lower, "text/") {
		return "document"
	} else if strings.Contains(lower, "json") {
		return "data"
	}
	return "file"
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

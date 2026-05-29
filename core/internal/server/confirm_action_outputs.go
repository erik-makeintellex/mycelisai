package server

import (
	"fmt"
	"path"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

func outputRefsFromPackageSummary(summary *packageOutputSummary) []protocol.OutputRef {
	if summary == nil {
		return nil
	}
	refs := make([]protocol.OutputRef, 0, len(summary.Files))
	for index, file := range summary.Files {
		role := "supporting"
		if isPackageProofFile(file.Path, file.Title) {
			role = "proof"
		} else if index == 0 {
			role = "primary"
		}
		refs = append(refs, protocol.OutputRef{
			ID:               fmt.Sprintf("%s:%s", summary.PackageID, path.Base(file.Path)),
			Kind:             outputKindForContentType(file.ContentType),
			Label:            firstNonEmptyString(file.Title, path.Base(file.Path)),
			StorageRef:       file.Path,
			SourceRunID:      summary.RunID,
			SourceContractID: summary.ProofID,
			ProofArtifactID:  proofArtifactIDForRole(summary, role),
			PackageID:        summary.PackageID,
			Role:             role,
			Status:           "final",
			ContentType:      file.ContentType,
			Bytes:            file.Bytes,
			SHA256:           file.SHA256,
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

func outputRefPointers(refs []protocol.OutputRef) (*protocol.OutputRef, *protocol.OutputRef) {
	var primaryRef *protocol.OutputRef
	var proofRef *protocol.OutputRef
	for _, ref := range refs {
		current := ref
		switch ref.Role {
		case "primary":
			if primaryRef == nil {
				primaryRef = &current
			}
		case "proof":
			if proofRef == nil {
				proofRef = &current
			}
		}
	}
	return primaryRef, proofRef
}

func isPackageProofFile(filePath, title string) bool {
	lowerPath := strings.ToLower(strings.TrimSpace(filePath))
	lowerTitle := strings.ToLower(strings.TrimSpace(title))
	return strings.HasSuffix(lowerPath, "proof.md") || strings.Contains(lowerTitle, "proof")
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
	return fmt.Sprintf("%s:%s", summary.PackageID, path.Base(summary.ProofPath))
}

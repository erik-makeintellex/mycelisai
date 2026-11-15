package server

import (
	"fmt"
	"path"
	"strings"

	"github.com/mycelis/core/internal/artifacts"
	"github.com/mycelis/core/pkg/protocol"
)

const (
	TeamPackageImplementationBundle = "implementation_bundle"
	TeamPackageMediaBundle          = "media_bundle"
	TeamPackageCreativeBundle       = "creative_bundle"
	TeamPackageDocumentBundle       = "document_bundle"

	VerificationPostureVerified      = "verified"
	VerificationPosturePendingReview = "pending_review"
	VerificationPostureFailed        = "failed"
)

// AssembleTeamSignalDeliveryPackage aggregates multi-agent team artifact outputs into a
// governed ExecutionSummary with an attached DeliveryPackageSummary.
func AssembleTeamSignalDeliveryPackage(teamID, runID, title string, items []artifacts.Artifact) *protocol.ExecutionSummary {
	if len(items) == 0 {
		return nil
	}

	trimmedTeamID, resolvedRunID := strings.TrimSpace(teamID), strings.TrimSpace(runID)
	collisions := DetectTeamPackageCollisions(items)
	packageKind, posture := ClassifyTeamDeliveryPackageKind(items), ResolveTeamPackageVerificationPosture(items)
	packageID := fmt.Sprintf("%s:%s", packageKind, firstNonEmptyString(trimmedTeamID, resolvedRunID, "team"))
	applyTeamPackageCollisionGovernance(trimmedTeamID, resolvedRunID, collisions)

	outputRefs := make([]protocol.OutputRef, 0, len(items))
	outputs := make([]protocol.ExecutionOutput, 0, len(items))
	teamRefs := make([]protocol.TeamOutputRef, 0, len(items))
	agentSet := make(map[string]struct{}, len(items))
	agents := make([]string, 0, len(items))

	for index, item := range items {
		role := "primary"
		if index > 0 {
			role = "supporting"
		}
		if isPackageProofFile(item.FilePath, item.Title) {
			role = "proof"
		}

		teamRef := teamOutputRefFromGroupArtifact(item, firstNonEmptyString(trimmedTeamID, "team"), role)
		teamRef.TeamID = firstNonEmptyString(trimmedTeamID, teamRef.TeamID)
		teamRef.PackageID = packageID
		if resolvedRunID != "" {
			teamRef.RunID = resolvedRunID
			if teamRef.Proof != nil {
				teamRef.Proof.SourceRunID = resolvedRunID
			}
		} else if teamRef.RunID != "" {
			resolvedRunID = teamRef.RunID
		}

		normalizedStatus := normalizeArtifactPostureStatus(item.Status)
		normPath := NormalizeTeamPackageFilePath(item.FilePath)
		if collidingAgents, hasCollision := collisions[normPath]; hasCollision {
			normalizedStatus = VerificationPosturePendingReview
			if teamRef.Proof != nil {
				teamRef.Proof.PathBoundaryStatus = "collision_detected"
				teamRef.Proof.ReadbackStatus = "degraded"
				teamRef.Proof.RecoveryHint = fmt.Sprintf(
					"Multi-agent write collision on %s across agents [%s]; assign disjoint file ownership before promoting.",
					normPath,
					strings.Join(collidingAgents, ", "),
				)
			}
		}
		teamRef.Status = normalizedStatus
		if teamRef.Proof != nil {
			teamRef.Proof.ExecutionStatus = normalizedStatus
		}

		outRef := protocol.OutputRefFromTeamOutputRef(teamRef)
		outRef.PackageID, outRef.Status = packageID, normalizedStatus

		if agentID := strings.TrimSpace(item.AgentID); agentID != "" {
			if _, seen := agentSet[agentID]; !seen {
				agentSet[agentID] = struct{}{}
				agents = append(agents, agentID)
			}
		}

		teamRefs = append(teamRefs, teamRef)
		outputRefs = append(outputRefs, outRef)
		outputs = append(outputs, executionOutputFromOutputRef(
			outRef,
			fmt.Sprintf("Multi-agent output from %s (%s).", firstNonEmptyString(item.AgentID, trimmedTeamID, "team"), packageKind),
			true,
		))
	}

	primaryRef, proofRef := outputRefPointers(outputRefs)
	if primaryRef == nil && len(outputRefs) > 0 {
		first := outputRefs[0]
		primaryRef = &first
	}
	if proofRef == nil && primaryRef != nil {
		proofRef = primaryRef
	}

	resolvedTitle := firstNonEmptyString(strings.TrimSpace(title), defaultTeamPackageTitle(packageKind, trimmedTeamID))
	agentSummary := "team agents"
	if len(agents) > 0 {
		agentSummary = strings.Join(agents, ", ")
	}
	summaryText := fmt.Sprintf("Assembled %s with %d artifact(s) from %s (posture: %s).", packageKind, len(outputRefs), agentSummary, posture)

	pkgSummary := &protocol.DeliveryPackageSummary{
		PackageID: packageID, RunID: resolvedRunID, Title: resolvedTitle, Summary: summaryText,
		PackageKind: packageKind, Status: posture, PrimaryOutputRef: primaryRef, Items: outputRefs,
		NextActions: nextActionsForTeamPackage(packageKind, posture), ProofRef: proofRef, InspectRefs: outputRefs,
	}

	proofID := ""
	if proofRef != nil {
		proofID = firstNonEmptyString(proofRef.ProofArtifactID, proofRef.ArtifactID, proofRef.ID)
	}

	return &protocol.ExecutionSummary{
		Title: resolvedTitle, Summary: summaryText, Status: posture, RunID: resolvedRunID,
		ProofID: proofID, OutputRefs: outputRefs, Outputs: outputs, TeamOutputRefs: teamRefs,
		DeliveryPackage: pkgSummary,
	}
}

// ClassifyTeamDeliveryPackageKind classifies a multi-agent artifact set into one of the
// canonical delivery package kinds: implementation_bundle, media_bundle, creative_bundle, or document_bundle.
func ClassifyTeamDeliveryPackageKind(items []artifacts.Artifact) string {
	var hasCode, hasMedia, hasDoc, hasChart bool
	for _, item := range items {
		aType := strings.ToLower(strings.TrimSpace(string(item.ArtifactType)))
		mime := strings.ToLower(strings.TrimSpace(item.ContentType))
		ext := strings.ToLower(path.Ext(strings.TrimSpace(item.FilePath)))
		switch {
		case aType == string(artifacts.TypeCode) || isCodeExtensionOrMime(ext, mime):
			hasCode = true
		case aType == string(artifacts.TypeImage) || aType == string(artifacts.TypeAudio) ||
			strings.HasPrefix(mime, "image/") || strings.HasPrefix(mime, "audio/") || strings.HasPrefix(mime, "video/"):
			hasMedia = true
		case aType == string(artifacts.TypeChart):
			hasChart = true
		default:
			hasDoc = true
		}
	}
	switch {
	case hasCode:
		return TeamPackageImplementationBundle
	case (hasMedia && (hasDoc || hasChart)) || hasChart:
		return TeamPackageCreativeBundle
	case hasMedia:
		return TeamPackageMediaBundle
	default:
		return TeamPackageDocumentBundle
	}
}

// ResolveTeamPackageVerificationPosture determines the aggregate governance verification posture.
func ResolveTeamPackageVerificationPosture(items []artifacts.Artifact) string {
	if len(items) == 0 {
		return VerificationPosturePendingReview
	}
	hasPending := false
	for _, item := range items {
		switch strings.ToLower(strings.TrimSpace(item.Status)) {
		case "rejected", "failed", "error":
			return VerificationPostureFailed
		case "pending", "pending_review", "proposed", "draft":
			hasPending = true
		}
	}
	if hasPending || len(DetectTeamPackageCollisions(items)) > 0 {
		return VerificationPosturePendingReview
	}
	return VerificationPostureVerified
}

func normalizeArtifactPostureStatus(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "rejected", "failed", "error":
		return VerificationPostureFailed
	case "pending", "pending_review", "proposed", "draft":
		return VerificationPosturePendingReview
	default:
		return VerificationPostureVerified
	}
}

func isCodeExtensionOrMime(ext, mime string) bool {
	switch ext {
	case ".go", ".ts", ".tsx", ".js", ".jsx", ".py", ".rs", ".sql", ".html", ".css", ".sh":
		return true
	}
	return strings.Contains(mime, "javascript") || strings.Contains(mime, "typescript") || strings.Contains(mime, "html") || strings.Contains(mime, "x-go")
}

func defaultTeamPackageTitle(packageKind, teamID string) string {
	if teamID != "" {
		return fmt.Sprintf("Team %s Delivery Package (%s)", teamID, packageKind)
	}
	return fmt.Sprintf("Team Delivery Package (%s)", packageKind)
}

func nextActionsForTeamPackage(packageKind, posture string) []string {
	switch posture {
	case VerificationPostureFailed:
		return []string{"Inspect rejected or failed team artifact.", "Re-run failed agent slice or request revision."}
	case VerificationPosturePendingReview:
		return []string{"Review pending team artifacts in operator console.", "Approve or reject pending package outputs."}
	default:
		return []string{fmt.Sprintf("Inspect verified %s outputs.", packageKind), "Promote package deliverables to downstream workflow."}
	}
}

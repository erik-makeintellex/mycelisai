package server

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/artifacts"
)

func TestAssembleTeamSignalDeliveryPackage_MultiAgentRoles(t *testing.T) {
	teamID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	artID1 := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa1")
	artID2 := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa2")
	meta, _ := json.Marshal(map[string]any{"source_contract_id": "contract-team-1"})

	items := []artifacts.Artifact{
		{
			ID:           artID1,
			TeamID:       &teamID,
			AgentID:      "coder-1",
			TraceID:      "run-team-1",
			ArtifactType: artifacts.TypeCode,
			Title:        "Handler Patch",
			ContentType:  "text/x-go",
			Content:      "package server\n",
			FilePath:     "core/internal/server/patch.go",
			Metadata:     meta,
			Status:       "approved",
			CreatedAt:    time.Now().UTC(),
		},
		{
			ID:           artID2,
			TeamID:       &teamID,
			AgentID:      "verifier-1",
			TraceID:      "run-team-1",
			ArtifactType: artifacts.TypeDocument,
			Title:        "Package Proof",
			ContentType:  "text/markdown",
			Content:      "# Proof\nverified\n",
			FilePath:     "workspace/deliverables/proof.md",
			Metadata:     meta,
			Status:       "approved",
			CreatedAt:    time.Now().UTC(),
		},
	}

	summary := AssembleTeamSignalDeliveryPackage(teamID.String(), "run-team-1", "Retained team outputs", items)
	if summary == nil {
		t.Fatal("expected non-nil ExecutionSummary")
	}
	if summary.DeliveryPackage == nil {
		t.Fatal("expected non-nil DeliveryPackage on ExecutionSummary")
	}
	dp := summary.DeliveryPackage
	if dp.PackageKind != TeamPackageImplementationBundle {
		t.Fatalf("PackageKind = %q, want %q", dp.PackageKind, TeamPackageImplementationBundle)
	}
	if dp.Status != VerificationPostureVerified {
		t.Fatalf("Status = %q, want %q", dp.Status, VerificationPostureVerified)
	}
	if dp.PackageID == "" || dp.RunID != "run-team-1" {
		t.Fatalf("unexpected DeliveryPackage identity: %#v", dp)
	}
	if len(dp.Items) != 2 {
		t.Fatalf("len(dp.Items) = %d, want 2", len(dp.Items))
	}
	if dp.PrimaryOutputRef == nil || dp.PrimaryOutputRef.ID != artID1.String() {
		t.Fatalf("PrimaryOutputRef = %#v", dp.PrimaryOutputRef)
	}
	if dp.ProofRef == nil || dp.ProofRef.ID != artID2.String() || dp.ProofRef.Role != "proof" {
		t.Fatalf("ProofRef = %#v", dp.ProofRef)
	}
}

func TestBuildTeamOutputsExecutionSummary(t *testing.T) {
	teamID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	artID1 := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbb1")
	artID2 := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbb2")
	meta, _ := json.Marshal(map[string]any{"source_contract_id": "contract-team-2", "run_id": "run-team-2"})

	items := []artifacts.Artifact{
		{
			ID:           artID1,
			TeamID:       &teamID,
			AgentID:      "writer-agent",
			TraceID:      "run-team-2",
			ArtifactType: artifacts.TypeDocument,
			Title:        "Architecture Spec",
			ContentType:  "text/markdown",
			Content:      "# Architecture Spec",
			FilePath:     "docs/spec.md",
			Metadata:     meta,
			Status:       "pending_review",
			CreatedAt:    time.Now().UTC(),
		},
		{
			ID:           artID2,
			TeamID:       &teamID,
			AgentID:      "lead-agent",
			TraceID:      "run-team-2",
			ArtifactType: artifacts.TypeDocument,
			Title:        "Delivery Proof",
			ContentType:  "text/markdown",
			Content:      "# Proof",
			FilePath:     "docs/proof.md",
			Metadata:     meta,
			Status:       "approved",
			CreatedAt:    time.Now().UTC(),
		},
	}

	summary := buildTeamOutputsExecutionSummary(teamID.String(), items)
	if summary == nil {
		t.Fatal("expected non-nil ExecutionSummary from buildTeamOutputsExecutionSummary")
	}
	if len(summary.OutputRefs) != 2 || len(summary.TeamOutputRefs) != 2 || len(summary.Outputs) != 2 {
		t.Fatalf("unexpected output slices: refs=%d teamRefs=%d outputs=%d", len(summary.OutputRefs), len(summary.TeamOutputRefs), len(summary.Outputs))
	}
	if summary.DeliveryPackage == nil {
		t.Fatal("expected DeliveryPackage to be populated on buildTeamOutputsExecutionSummary")
	}
	dp := summary.DeliveryPackage
	if dp.PackageID != "document_bundle:"+teamID.String() {
		t.Fatalf("PackageID = %q, want %q", dp.PackageID, "document_bundle:"+teamID.String())
	}
	if dp.PackageKind != TeamPackageDocumentBundle {
		t.Fatalf("PackageKind = %q, want %q", dp.PackageKind, TeamPackageDocumentBundle)
	}
	if dp.Status != VerificationPosturePendingReview {
		t.Fatalf("VerificationPosture (Status) = %q, want %q", dp.Status, VerificationPosturePendingReview)
	}
	if len(dp.Items) != 2 {
		t.Fatalf("len(dp.Items) = %d, want 2", len(dp.Items))
	}
	if dp.Items[0].StorageRef != "docs/spec.md" || dp.Items[1].Role != "proof" {
		t.Fatalf("unexpected DeliveryPackage items: %#v", dp.Items)
	}
	if buildTeamOutputsExecutionSummary(teamID.String(), nil) != nil {
		t.Fatal("expected nil summary for empty items")
	}
}

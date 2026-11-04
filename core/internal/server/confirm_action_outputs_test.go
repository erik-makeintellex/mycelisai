package server

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/artifacts"
	"github.com/mycelis/core/pkg/protocol"
)

func TestExecutionSummaryFromPlannedToolCalls_NormalizesWriteFileOutputs(t *testing.T) {
	workspace := t.TempDir()
	t.Setenv("MYCELIS_WORKSPACE", workspace)

	targetRel := "notes/release_brief.md"
	content := "# Release Brief\n\nNormalized output ref proof.\n"
	targetAbs := filepath.Join(workspace, filepath.FromSlash(targetRel))
	if err := os.MkdirAll(filepath.Dir(targetAbs), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(targetAbs, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	planned := []protocol.PlannedToolCall{
		{
			Name: "write_file",
			Arguments: map[string]any{
				"path":    targetRel,
				"title":   "Release Brief",
				"content": content,
			},
		},
	}

	summary := executionSummaryFromPlannedToolCalls(planned, "run-confirm-1", "proof-confirm-1")
	if summary == nil {
		t.Fatal("expected execution summary for non-package write_file")
	}
	if len(summary.OutputRefs) != 1 || len(summary.Outputs) != 1 || len(summary.ArtifactRefs) != 1 {
		t.Fatalf("counts = refs:%d outputs:%d artifacts:%d, want 1/1/1", len(summary.OutputRefs), len(summary.Outputs), len(summary.ArtifactRefs))
	}

	ref := summary.OutputRefs[0]
	if ref.ID == "" || ref.Kind != "document" || ref.Label != "Release Brief" {
		t.Fatalf("output ref identity = %#v", ref)
	}
	if ref.StorageRef != targetRel || ref.OpenURL == "" || ref.SourceRunID != "run-confirm-1" || ref.SourceContractID != "proof-confirm-1" {
		t.Fatalf("output ref linkage = %#v", ref)
	}
	if ref.Role != "primary" || ref.Status != "verified" {
		t.Fatalf("output ref role/status = %s/%s, want primary/verified", ref.Role, ref.Status)
	}
	if ref.Proof == nil || ref.ProofEnvelope == nil {
		t.Fatalf("expected proof and proof_envelope on output ref: %#v", ref)
	}
	if ref.ProofEnvelope.ReadbackStatus != "verified" || ref.ProofEnvelope.PathBoundaryStatus != "verified" || ref.ProofEnvelope.Checksum == "" {
		t.Fatalf("unexpected proof envelope: %#v", ref.ProofEnvelope)
	}
	if summary.Outputs[0].StorageRef != targetRel || summary.ArtifactRefs[0].SavedPath != targetRel {
		t.Fatalf("compatibility fields mismatch: output=%#v artifact=%#v", summary.Outputs[0], summary.ArtifactRefs[0])
	}
}

func TestBuildMCPToolCallAndActivityExecutionSummary_PopulatesNormalizedOutputRefs(t *testing.T) {
	toolSummary := buildMCPToolCallExecutionSummary(
		"srv-1", "filesystem", "read_file", "Read workspace brief successfully.", "exch-item-1", "run-mcp-1",
		map[string]any{"summary": "Read workspace brief successfully.", "bytes": 128},
	)
	if toolSummary == nil || len(toolSummary.OutputRefs) != 1 || len(toolSummary.Outputs) != 1 {
		t.Fatalf("mcp tool summary = %#v", toolSummary)
	}
	ref := toolSummary.OutputRefs[0]
	if ref.ID != "exch-item-1" || ref.Kind != "document" || ref.Label != "read_file" || ref.StorageRef != "mcp://filesystem/read_file/exch-item-1" || ref.OpenURL == "" || ref.SourceRunID != "run-mcp-1" || ref.Role != "primary" || ref.Status != "recorded" || ref.ProofEnvelope == nil {
		t.Fatalf("mcp tool output ref = %#v", ref)
	}

	activitySummary := executionSummaryFromMCPActivityEntries([]mcpActivityEntry{
		{ID: "act-1", ServerID: "srv-1", ServerName: "filesystem", ToolName: "read_file", State: "completed", Summary: "Read architecture summary", Message: "Read architecture summary", ChannelName: "browser.research.results", RunID: "run-act-1", TeamID: "team-alpha", Timestamp: "2026-10-25T12:00:00Z"},
		{ID: "act-2", ServerID: "srv-2", ServerName: "media", ToolName: "generate_image", State: "completed", Summary: "Generated concept image", Message: "Generated concept image", ChannelName: "media.image.output", RunID: "run-act-1", TeamID: "team-alpha", Timestamp: "2026-10-25T12:01:00Z"},
	})
	if activitySummary == nil || len(activitySummary.OutputRefs) != 2 || len(activitySummary.Outputs) != 2 {
		t.Fatalf("unexpected activity summary: %#v", activitySummary)
	}
	if activitySummary.OutputRefs[0].Role != "primary" || activitySummary.OutputRefs[1].Role != "supporting" || activitySummary.OutputRefs[1].Kind != "image" {
		t.Fatalf("unexpected activity output refs: %#v", activitySummary.OutputRefs)
	}
}

func TestBuildTeamWorkConfirmExecutionSummary_PopulatesOutputRefsAndTeamOutputRefs(t *testing.T) {
	workspace := t.TempDir()
	t.Setenv("MYCELIS_WORKSPACE", workspace)

	targetRel := "deliverables/plan.go"
	targetAbs := filepath.Join(workspace, filepath.FromSlash(targetRel))
	if err := os.MkdirAll(filepath.Dir(targetAbs), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(targetAbs, []byte("package deliverables\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	planned := []protocol.PlannedToolCall{
		{Name: "write_file", Arguments: map[string]any{"path": targetRel, "title": "Deliverable Plan"}},
		{Name: "create_team", Arguments: map[string]any{"name": "Delivery Squad", "goal": "Ship normalized output refs"}},
	}

	summary := buildTeamWorkConfirmExecutionSummary(planned, "team-delivery", "work-1", "run-team-1", "proof-team-1")
	if summary == nil || len(summary.OutputRefs) != 2 || len(summary.Outputs) != 2 || len(summary.TeamOutputRefs) != 2 || len(summary.ArtifactRefs) != 1 {
		t.Fatalf("unexpected summary slice lengths: %#v", summary)
	}
	if summary.OutputRefs[0].Kind != "code" || summary.OutputRefs[0].Role != "primary" || summary.OutputRefs[0].ProofEnvelope == nil {
		t.Fatalf("first team confirm output ref = %#v", summary.OutputRefs[0])
	}
	if summary.OutputRefs[1].Kind != "team" || summary.OutputRefs[1].Role != "supporting" || summary.OutputRefs[1].SourceRunID != "run-team-1" {
		t.Fatalf("second team confirm output ref = %#v", summary.OutputRefs[1])
	}
}

func TestConfirmActionMediaOutputs_ProjectsEphemeralAndSavedMediaWithProofEnvelope(t *testing.T) {
	imgBytes := []byte("verified-media-image")
	imgB64 := base64.StdEncoding.EncodeToString(imgBytes)
	wantSHA := fmt.Sprintf("%x", sha256.Sum256(imgBytes))

	ephemeralRaw := `{"message":"Image generated","artifact":{"id":"art-eph-1","type":"image","title":"Generated: Hero","content_type":"image/png","content":"` + imgB64 + `","cached":true}}`
	ephRef, _, ephArt, ok := extractOutputsFromToolResult(ephemeralRaw, "run-media-1", "proof-media-1", "primary")
	if !ok {
		t.Fatal("expected ephemeral generate_image output extraction")
	}
	if ephRef.Kind != "image" || ephRef.StorageRef != "artifact://art-eph-1" || ephRef.ContentType != "image/png" || ephRef.Bytes != len(imgBytes) || ephRef.SHA256 != wantSHA {
		t.Fatalf("unexpected ephemeral OutputRef: %#v", ephRef)
	}
	if ephRef.Proof == nil || ephRef.Proof.ExecutionStatus != "cached" || !ephArt.Cached {
		t.Fatalf("expected cached proof posture on ephemeral media: ref=%#v art=%#v", ephRef, ephArt)
	}

	planned := []protocol.PlannedToolCall{
		{
			Name: "generate_image",
			Arguments: map[string]any{
				"artifact": map[string]any{"id": "art-eph-2", "type": "image", "title": "Concept Sketch", "content_type": "image/png", "content": imgB64, "cached": true},
			},
		},
		{
			Name: "save_cached_image",
			Arguments: map[string]any{
				"artifact": map[string]any{
					"id": "art-saved-1", "type": "image", "title": "hero.png",
					"saved_path": "saved-media/hero.png", "content_type": "image/png",
					"bytes": len(imgBytes), "sha256": wantSHA, "verified": true,
				},
			},
		},
	}

	summary := buildTeamWorkConfirmExecutionSummary(planned, "team-media", "work-media-1", "run-media-2", "proof-media-2")
	if summary == nil || len(summary.OutputRefs) != 2 || len(summary.ArtifactRefs) != 2 {
		t.Fatalf("unexpected media confirm summary: %#v", summary)
	}
	savedRef := summary.OutputRefs[1]
	if savedRef.Kind != "image" || savedRef.StorageRef != "saved-media/hero.png" || savedRef.ContentType != "image/png" || savedRef.Bytes != len(imgBytes) || savedRef.SHA256 != wantSHA {
		t.Fatalf("unexpected saved media OutputRef: %#v", savedRef)
	}
	if savedRef.Proof == nil || savedRef.Proof.ExecutionStatus != "verified" || savedRef.Proof.ReadbackStatus != "verified" || savedRef.Proof.Checksum != wantSHA {
		t.Fatalf("unexpected saved media OutputProofEnvelope: %#v", savedRef.Proof)
	}
}

func TestBuildTeamAndGroupExecutionSummaries_PopulateNormalizedOutputRefs(t *testing.T) {
	teamID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	artID1 := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa1")
	artID2 := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa2")
	meta, _ := json.Marshal(map[string]any{"source_contract_id": "contract-grp-1", "package_id": "pkg-grp-1"})
	group := &CollaborationGroup{ID: "group-alpha", Name: "Alpha Launch Group", TeamIDs: []string{teamID.String()}}
	items := []artifacts.Artifact{
		{ID: artID1, TeamID: &teamID, AgentID: "lead-agent", TraceID: "run-grp-1", ArtifactType: artifacts.TypeDocument, Title: "Launch Brief", ContentType: "text/markdown", Content: "# Launch Brief", FilePath: "groups/alpha/launch_brief.md", Metadata: meta, Status: "approved", CreatedAt: time.Now().UTC()},
		{ID: artID2, TeamID: &teamID, AgentID: "lead-agent", TraceID: "run-grp-1", ArtifactType: artifacts.TypeDocument, Title: "Package Proof", ContentType: "text/markdown", Content: "# Proof", FilePath: "groups/alpha/proof.md", Metadata: meta, Status: "approved", CreatedAt: time.Now().UTC()},
	}

	groupSummary := buildGroupOutputsExecutionSummary(group, items)
	if groupSummary == nil || len(groupSummary.OutputRefs) != 2 || len(groupSummary.TeamOutputRefs) != 2 || len(groupSummary.Outputs) != 2 {
		t.Fatalf("unexpected group outputs summary: %#v", groupSummary)
	}
	if groupSummary.OutputRefs[0].ID != artID1.String() || groupSummary.OutputRefs[0].Role != "primary" || groupSummary.OutputRefs[0].OpenURL != "/api/v1/artifacts/"+artID1.String()+"/download" {
		t.Fatalf("first group output ref = %#v", groupSummary.OutputRefs[0])
	}
	if groupSummary.OutputRefs[1].Role != "proof" || groupSummary.OutputRefs[1].ProofEnvelope == nil {
		t.Fatalf("proof group output ref = %#v", groupSummary.OutputRefs[1])
	}

	guidance := TeamLeadGuidanceResponse{
		Action: TeamLeadGuidedActionResumeRetainedPackage, Headline: "Resume Alpha Launch Group", Summary: "Resume from retained outputs.",
		ExecutionContract: &TeamLeadExecutionContract{
			ExecutionMode: TeamLeadExecutionModeContinuityResume, OwnerLabel: "Team Lead", TeamName: "Alpha Team",
			Summary: "Continue from retained launch package.", TargetOutputs: []string{"Launch Brief", "Remaining Checklist"},
			WorkflowGroup: &TeamLeadWorkflowGroupDraft{GroupID: "group-alpha", Name: "Alpha Launch Group"},
		},
	}
	guidanceSummary := buildTeamLeadGuidanceExecutionSummary(guidance, "Resume launch", "run-guide-1", "contract-guide-1")
	if guidanceSummary == nil || len(guidanceSummary.OutputRefs) != 2 || len(guidanceSummary.TeamOutputRefs) != 2 {
		t.Fatalf("unexpected guidance summary: %#v", guidanceSummary)
	}
	if guidanceSummary.OutputRefs[0].SourceRunID != "run-guide-1" || guidanceSummary.OutputRefs[0].Status != "retained" || guidanceSummary.OutputRefs[0].ProofEnvelope == nil {
		t.Fatalf("first guidance output ref = %#v", guidanceSummary.OutputRefs[0])
	}

	broadcastSummary := buildGroupBroadcastExecutionSummary(group, "Sync launch status", "audit-bcast-1", "run-bcast-1", []string{"s1", "s2"})
	if broadcastSummary == nil || len(broadcastSummary.OutputRefs) != 2 || len(broadcastSummary.TeamOutputRefs) != 2 {
		t.Fatalf("unexpected broadcast summary: %#v", broadcastSummary)
	}
	if broadcastSummary.OutputRefs[0].Kind != "group_broadcast" || broadcastSummary.OutputRefs[1].Kind != "team_signal" {
		t.Fatalf("unexpected broadcast output refs: %#v", broadcastSummary.OutputRefs)
	}
}

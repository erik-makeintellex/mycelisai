package server

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFullDeliveryPipeline_AllBundleKindsProduceVerifiedOutputRefsAndProofEnvelopes(t *testing.T) {
	workspace := t.TempDir()
	t.Setenv("MYCELIS_WORKSPACE", workspace)

	cases := []struct {
		name            string
		prompt          string
		wantPackageKind string
		wantMinTeams    int
		wantFiles       int // including generated proof.md
		wantPrimaryType string
	}{
		{
			name:            "dev_project_bundle (casual build ask)",
			prompt:          "make me a quick timer app",
			wantPackageKind: "dev_project_bundle",
			wantMinTeams:    0,
			wantFiles:       4,
			wantPrimaryType: "text/html",
		},
		{
			name:            "dev_project_bundle (team-assigned build ask)",
			prompt:          "have a team build me a quick calculator app project",
			wantPackageKind: "dev_project_bundle",
			wantMinTeams:    1,
			wantFiles:       4,
			wantPrimaryType: "text/html",
		},
		{
			name:            "multi_app_bundle (interacting primary + companion apps)",
			prompt:          "Build a storefront app and have a team work on another application for inventory that interacts with the initial ask.",
			wantPackageKind: "multi_app_bundle",
			wantMinTeams:    3,
			wantFiles:       4,
			wantPrimaryType: "text/html",
		},
		{
			name:            "mcp_addition_bundle (Standard Deployment MCP addition)",
			prompt:          "Add Standard Deployment MCP Filesystem (mcp_id: filesystem, mcp_group: standard, mcp_title: Filesystem) to team Mission Delivery Team (team_id: mission-delivery-team) as an MCP addition and generate its deployment binding manifest.",
			wantPackageKind: "mcp_addition_bundle",
			wantMinTeams:    1,
			wantFiles:       3,
			wantPrimaryType: "text/yaml",
		},
		{
			name:            "mcp_addition_bundle (Custom Developed MCP addition + scaffold)",
			prompt:          "Add Custom Developed MCP Custom Python FastMCP Server (mcp_id: custom-python-fastmcp, mcp_group: custom, mcp_title: Custom Python FastMCP Server) to team Agui Design Architect (team_id: agui-design-architect) as an MCP addition and generate its deployment binding manifest.",
			wantPackageKind: "mcp_addition_bundle",
			wantMinTeams:    1,
			wantFiles:       4,
			wantPrimaryType: "text/yaml",
		},
		{
			name:            "team_execution_bundle (plan & execute next steps)",
			prompt:          "Plan for team Alpha Strike (team_id: alpha-strike) and execute next steps to produce its execution plan, action matrix, and handoff deliverable package.",
			wantPackageKind: "team_execution_bundle",
			wantMinTeams:    2,
			wantFiles:       4,
			wantPrimaryType: "text/markdown",
		},
		{
			name:            "chained_pipeline_bundle (inter-team & MCP channel handoff)",
			prompt:          "Chain the output published on channel swarm.team.mcp-upstream-discovery-team.signal.result into a downstream team that listens on swarm.team.mcp-upstream-discovery-team.signal.result and generates the next deliverable package.",
			wantPackageKind: "chained_pipeline_bundle",
			wantMinTeams:    2,
			wantFiles:       4,
			wantPrimaryType: "text/yaml",
		},
		{
			name:            "interactive_game_bundle (Street Fighter II Tetris)",
			prompt:          "Put a team together to create a playable Street Fighter II arcade Tetris game in the browser.",
			wantPackageKind: "interactive_game_bundle",
			wantMinTeams:    2,
			wantFiles:       4,
			wantPrimaryType: "text/html",
		},
		{
			name:            "mission_blueprint_bundle (multi-team mission blueprint)",
			prompt:          "Create a multi-agent mission blueprint with team lanes for plan, develop, verify, and deliver.",
			wantPackageKind: "mission_blueprint_bundle",
			wantMinTeams:    0,
			wantFiles:       4,
			wantPrimaryType: "text/yaml",
		},
	}

	s := newTestServer()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			review := s.buildSomaReferentialReview(context.Background(), []chatRequestMessage{{Role: "user", Content: tc.prompt}})
			if review.NeedsConfirmation {
				t.Fatalf("expected direct proposal fast-path without text confirmation gate, got NeedsConfirmation=true (%+v)", review)
			}

			mutTools := uniqueOrderedTools(append(inferMutationToolsFromText(tc.prompt), review.MutationTools...))
			planned, fastPathOK := hasDeterministicFastPathPlan(tc.prompt, mutTools)
			if !fastPathOK || len(planned) == 0 {
				t.Fatalf("expected deterministic fast-path plan for %q (mutTools=%v)", tc.prompt, mutTools)
			}

			display := buildProposalDisplayContract(planned, tc.prompt, mutTools)
			replyText := buildFastPathProposalReplyText(display)
			if !strings.Contains(replyText, "I'm ready to start:") || !strings.Contains(replyText, "Approve & Execute") {
				t.Fatalf("unexpected fast-path reply text: %q", replyText)
			}

			teamCount := 0
			for _, call := range planned {
				switch call.Name {
				case "create_team":
					teamCount++
					members, _ := call.Arguments["members"].(map[string]any)
					hasReviewer := false
					for role, agent := range members {
						hay := strings.ToLower(role + " " + firstNonEmptyString(agent))
						if strings.Contains(hay, "review") || strings.Contains(hay, "sentry") {
							hasReviewer = true
							break
						}
					}
					if !hasReviewer {
						t.Fatalf("planned team %v missing Ideal Delivery Reviewer in members: %#v", call.Arguments["team_id"], members)
					}
				case "write_file":
					relPath, _ := call.Arguments["path"].(string)
					content, _ := call.Arguments["content"].(string)
					if strings.TrimSpace(relPath) == "" || strings.TrimSpace(content) == "" {
						t.Fatalf("write_file call has empty path or content: %#v", call.Arguments)
					}
					absPath, err := workspacePathForServer(relPath)
					if err != nil {
						t.Fatalf("workspacePathForServer(%s): %v", relPath, err)
					}
					if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
						t.Fatalf("mkdir: %v", err)
					}
					if err := os.WriteFile(absPath, []byte(content), 0o644); err != nil {
						t.Fatalf("write %s: %v", relPath, err)
					}
				}
			}

			if teamCount < tc.wantMinTeams {
				t.Fatalf("planned create_team count = %d, want >= %d", teamCount, tc.wantMinTeams)
			}

			summary, err := s.finalizeCodexTaskBundleTx(context.Background(), nil, planned, "run-full-1", "proof-full-1")
			if err != nil {
				t.Fatalf("finalizeCodexTaskBundleTx error: %v", err)
			}
			if summary == nil {
				t.Fatal("expected non-nil packageOutputSummary")
			}
			if summary.PackageKind != tc.wantPackageKind {
				t.Fatalf("PackageKind = %q, want %q", summary.PackageKind, tc.wantPackageKind)
			}
			if len(summary.Files) != tc.wantFiles || len(summary.OutputRefs) != tc.wantFiles {
				t.Fatalf("files=%d outputRefs=%d, want %d", len(summary.Files), len(summary.OutputRefs), tc.wantFiles)
			}
			if summary.PrimaryRef == nil || summary.ProofRef == nil {
				t.Fatalf("expected PrimaryRef and ProofRef on packageOutputSummary: %+v", summary)
			}
			if summary.PrimaryRef.ContentType != tc.wantPrimaryType {
				t.Fatalf("PrimaryRef.ContentType = %q, want %q", summary.PrimaryRef.ContentType, tc.wantPrimaryType)
			}

			for i, ref := range summary.OutputRefs {
				if ref.Proof == nil {
					t.Fatalf("OutputRefs[%d] (%s) missing OutputProofEnvelope", i, ref.StorageRef)
				}
				if ref.Proof.PathBoundaryStatus != "verified" || ref.Proof.ReadbackStatus != "verified" {
					t.Fatalf("OutputRefs[%d] (%s) boundary/readback = %s/%s, want verified/verified",
						i, ref.StorageRef, ref.Proof.PathBoundaryStatus, ref.Proof.ReadbackStatus)
				}
				if len(ref.SHA256) != 64 || ref.Bytes <= 0 {
					t.Fatalf("OutputRefs[%d] (%s) invalid sha256/bytes: sha=%q bytes=%d", i, ref.StorageRef, ref.SHA256, ref.Bytes)
				}
			}
		})
	}
}

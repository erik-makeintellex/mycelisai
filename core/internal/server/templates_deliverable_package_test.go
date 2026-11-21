package server

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

func TestInferUserDeliverableBundlePlanFromRequest_MarketingAndGame(t *testing.T) {
	t.Run("marketing launch bundle", func(t *testing.T) {
		calls, ok := inferUserDeliverableBundlePlanFromRequest("Create a marketing launch kit with positioning brief, launch email, and customer FAQ.")
		if !ok {
			t.Fatal("expected marketing deliverable bundle plan")
		}
		if len(calls) != 3 {
			t.Fatalf("calls = %d, want 3", len(calls))
		}
		wantPaths := []string{
			"deliverables/marketing-launch-kit/positioning_brief.md",
			"deliverables/marketing-launch-kit/launch_email_and_copy.md",
			"deliverables/marketing-launch-kit/customer_faq.md",
		}
		for i, call := range calls {
			if call.Name != "write_file" {
				t.Fatalf("call[%d].Name = %q, want write_file", i, call.Name)
			}
			if call.Arguments["path"] != wantPaths[i] {
				t.Fatalf("call[%d].path = %v, want %s", i, call.Arguments["path"], wantPaths[i])
			}
			if call.Arguments["package_kind"] != "marketing_launch_bundle" {
				t.Fatalf("call[%d].package_kind = %v", i, call.Arguments["package_kind"])
			}
			if call.Arguments["content_type"] != "text/markdown" {
				t.Fatalf("call[%d].content_type = %v", i, call.Arguments["content_type"])
			}
			content, _ := call.Arguments["content"].(string)
			if strings.TrimSpace(content) == "" {
				t.Fatalf("call[%d] content is empty", i)
			}
		}
	})

	t.Run("interactive puzzle game bundle", func(t *testing.T) {
		calls, ok := inferUserDeliverableBundlePlanFromRequest("Build a playable sliding-tile puzzle game in the browser.")
		if !ok {
			t.Fatal("expected interactive game bundle plan")
		}
		if len(calls) != 3 {
			t.Fatalf("calls = %d, want 3", len(calls))
		}
		wantPaths := []string{
			"deliverables/games/sliding-tile-puzzle/index.html",
			"deliverables/games/sliding-tile-puzzle/game_design_and_rules.md",
			"deliverables/games/sliding-tile-puzzle/qa_and_controls.md",
		}
		for i, call := range calls {
			if call.Arguments["path"] != wantPaths[i] {
				t.Fatalf("call[%d].path = %v, want %s", i, call.Arguments["path"], wantPaths[i])
			}
			if call.Arguments["package_kind"] != "interactive_game_bundle" {
				t.Fatalf("call[%d].package_kind = %v", i, call.Arguments["package_kind"])
			}
		}
		htmlContent, _ := calls[0].Arguments["content"].(string)
		for _, snippet := range []string{
			"<h1>Sliding Tile Puzzle</h1>",
			`<span id="move-counter">Moves: 0</span>`,
			`<div id="puzzle-grid">`,
			`data-testid="puzzle-tile"`,
			`<button id="shuffle-btn"`,
			`<div id="win-banner" hidden>Puzzle Solved!</div>`,
		} {
			if !strings.Contains(htmlContent, snippet) {
				t.Fatalf("index.html missing required snippet %q", snippet)
			}
		}
	})
}

func TestInferUserDeliverableBundlePlanFromRequest_OnboardingKPIDashboardAndWireframe(t *testing.T) {
	t.Run("customer onboarding bundle", func(t *testing.T) {
		calls, ok := inferUserDeliverableBundlePlanFromRequest("Create a customer onboarding welcome kit with kickoff guide, checklist, and support FAQ.")
		if !ok || len(calls) != 3 {
			t.Fatalf("ok=%v len=%d, want true & 3", ok, len(calls))
		}
		wantPaths := []string{
			"deliverables/customer-onboarding-kit/welcome_and_kickoff_guide.md",
			"deliverables/customer-onboarding-kit/first_30_days_checklist.md",
			"deliverables/customer-onboarding-kit/escalation_and_faq.md",
		}
		wantRoles := []string{"primary", "supporting", "supporting"}
		for i, call := range calls {
			if call.Arguments["path"] != wantPaths[i] || call.Arguments["package_kind"] != "customer_onboarding_bundle" || call.Arguments["package_slug"] != "customer-onboarding-kit" {
				t.Fatalf("call[%d] args = %#v", i, call.Arguments)
			}
			if call.Arguments["artifact_role"] != wantRoles[i] || call.Arguments["content_type"] != "text/markdown" {
				t.Fatalf("call[%d] role/type = %v/%v", i, call.Arguments["artifact_role"], call.Arguments["content_type"])
			}
		}
	})

	t.Run("kpi dashboard bundle", func(t *testing.T) {
		calls, ok := inferUserDeliverableBundlePlanFromRequest("Build an executive KPI dashboard with interactive period filters and metric definitions.")
		if !ok || len(calls) != 3 {
			t.Fatalf("ok=%v len=%d, want true & 3", ok, len(calls))
		}
		wantPaths := []string{
			"deliverables/dashboards/executive-kpi-dashboard/index.html",
			"deliverables/dashboards/executive-kpi-dashboard/metrics_definitions.md",
			"deliverables/dashboards/executive-kpi-dashboard/data_sources_and_governance.md",
		}
		for i, call := range calls {
			if call.Arguments["path"] != wantPaths[i] || call.Arguments["package_kind"] != "kpi_dashboard_bundle" || call.Arguments["package_slug"] != "executive-kpi-dashboard" {
				t.Fatalf("call[%d] args = %#v", i, call.Arguments)
			}
		}
		if calls[0].Arguments["content_type"] != "text/html" || calls[0].Arguments["artifact_role"] != "primary" {
			t.Fatalf("call[0] html metadata = %#v", calls[0].Arguments)
		}
		htmlContent, _ := calls[0].Arguments["content"].(string)
		for _, snippet := range []string{
			"<h1>Executive KPI Dashboard</h1>",
			`<span id="active-period">Period: 30D</span>`,
			`data-testid="kpi-period-qtd"`,
		} {
			if !strings.Contains(htmlContent, snippet) {
				t.Fatalf("index.html missing required snippet %q", snippet)
			}
		}
	})

	t.Run("product wireframe bundle", func(t *testing.T) {
		calls, ok := inferUserDeliverableBundlePlanFromRequest("Create an interactive prototype and product wireframe with product spec and user flows.")
		if !ok || len(calls) != 3 {
			t.Fatalf("ok=%v len=%d, want true & 3", ok, len(calls))
		}
		wantPaths := []string{
			"deliverables/product-wireframe/index.html",
			"deliverables/product-wireframe/product_spec.md",
			"deliverables/product-wireframe/user_flows_and_acceptance.md",
		}
		for i, call := range calls {
			if call.Arguments["path"] != wantPaths[i] || call.Arguments["package_kind"] != "product_wireframe_bundle" || call.Arguments["package_slug"] != "product-wireframe" {
				t.Fatalf("call[%d] args = %#v", i, call.Arguments)
			}
		}
		htmlContent, _ := calls[0].Arguments["content"].(string)
		for _, snippet := range []string{
			"<h1>Interactive Product Wireframe</h1>",
			`data-testid="wireframe-tab-overview"`,
			`data-testid="wireframe-tab-workflow"`,
			`<h2 id="active-screen">Screen: Workspace Overview</h2>`,
		} {
			if !strings.Contains(htmlContent, snippet) {
				t.Fatalf("wireframe index.html missing snippet %q", snippet)
			}
		}
	})
}

func TestBuildPlannedToolCalls_FallbackDeliverableWritePlan(t *testing.T) {
	calls := buildPlannedToolCalls(
		chatAgentResult{Text: "Here is a structured launch summary covering goals, audience, and next steps."},
		"Create a one-page launch summary",
		[]string{"write_file"},
	)
	if len(calls) != 1 {
		t.Fatalf("planned calls = %d, want 1", len(calls))
	}
	if calls[0].Name != "write_file" {
		t.Fatalf("call name = %q, want write_file", calls[0].Name)
	}
	if calls[0].Arguments["path"] != "deliverables/soma-draft.md" {
		t.Fatalf("path = %v, want deliverables/soma-draft.md", calls[0].Arguments["path"])
	}
	content, _ := calls[0].Arguments["content"].(string)
	if !strings.Contains(content, "structured launch summary") {
		t.Fatalf("fallback content = %q, want agent reply text included", content)
	}
}

func TestFinalizeCodexTaskBundleTx_SupportsGameAndWireframeBundles(t *testing.T) {
	workspace := t.TempDir()
	t.Setenv("MYCELIS_WORKSPACE", workspace)

	for _, tc := range []struct {
		prompt    string
		wantKind  string
		wantPkgID string
		wantProof string
	}{
		{
			prompt:    "Create a playable sliding-tile puzzle game for the browser.",
			wantKind:  "interactive_game_bundle",
			wantPkgID: "interactive_game_bundle:sliding-tile-puzzle",
			wantProof: "deliverables/games/sliding-tile-puzzle/proof.md",
		},
		{
			prompt:    "Build an interactive product wireframe and product spec.",
			wantKind:  "product_wireframe_bundle",
			wantPkgID: "product_wireframe_bundle:product-wireframe",
			wantProof: "deliverables/product-wireframe/proof.md",
		},
	} {
		planned, ok := inferUserDeliverableBundlePlanFromRequest(tc.prompt)
		if !ok {
			t.Fatalf("expected bundle plan for %q", tc.prompt)
		}
		for _, call := range planned {
			relPath, _ := call.Arguments["path"].(string)
			content, _ := call.Arguments["content"].(string)
			absPath := filepath.Join(workspace, filepath.FromSlash(relPath))
			if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
				t.Fatalf("mkdir: %v", err)
			}
			if err := os.WriteFile(absPath, []byte(content), 0o644); err != nil {
				t.Fatalf("write %s: %v", relPath, err)
			}
		}

		s := &AdminServer{}
		summary, err := s.finalizeCodexTaskBundleTx(context.Background(), nil, planned, "run-1", "proof-1")
		if err != nil {
			t.Fatalf("finalizeCodexTaskBundleTx(%s): %v", tc.wantKind, err)
		}
		if summary == nil || summary.PackageKind != tc.wantKind || summary.PackageID != tc.wantPkgID {
			t.Fatalf("summary = %+v, want kind=%s id=%s", summary, tc.wantKind, tc.wantPkgID)
		}
		if len(summary.Files) != 4 || len(summary.Artifacts) != 4 {
			t.Fatalf("files/artifacts = %d/%d, want 4/4", len(summary.Files), len(summary.Artifacts))
		}
		if _, err := os.Stat(filepath.Join(workspace, filepath.FromSlash(tc.wantProof))); err != nil {
			t.Fatalf("expected %s on disk: %v", tc.wantProof, err)
		}
	}

	_ = protocol.PlannedToolCall{}
}

func TestInferMutationToolsFromText_IncludesWriteFileForPuzzleAndMarketingBundles(t *testing.T) {
	for _, prompt := range []string{
		"Create a simple browser sliding-tile puzzle game with playable HTML, game rules, and controls.",
		"Build a playable sliding-tile puzzle game in the browser.",
		"Create marketing documentation for our product launch including a positioning brief, launch email, and customer FAQ.",
		"Prepare a customer onboarding welcome kit for enterprise accounts.",
		"Generate an executive KPI dashboard for governance metrics.",
		"Create an interactive product wireframe and product spec.",
	} {
		if tools := inferMutationToolsFromText(prompt); !containsString(tools, "write_file") {
			t.Fatalf("inferMutationToolsFromText(%q) = %v, want write_file included", prompt, tools)
		}
	}
}

func TestInferUserDeliverableBundlePlanFromRequest_TeamExecutionBundle(t *testing.T) {
	prompt := "Plan for team Alpha Strike (team_id: alpha-strike) and execute next steps to produce its deliverable package."
	calls, ok := inferUserDeliverableBundlePlanFromRequest(prompt)
	if !ok || len(calls) != 3 {
		t.Fatalf("ok=%v len=%d, want true & 3", ok, len(calls))
	}
	wantPaths := []string{
		"deliverables/teams/alpha-strike/team_execution_plan.md",
		"deliverables/teams/alpha-strike/next_steps_action_matrix.md",
		"deliverables/teams/alpha-strike/team_handoff_summary.json",
	}
	for i, call := range calls {
		if call.Arguments["path"] != wantPaths[i] || call.Arguments["package_kind"] != "team_execution_bundle" || call.Arguments["package_slug"] != "alpha-strike" {
			t.Fatalf("call[%d] args = %#v", i, call.Arguments)
		}
		if call.Arguments["team_id"] != "alpha-strike" || call.Arguments["team_name"] != "Alpha Strike" {
			t.Fatalf("call[%d] team metadata = %v / %v", i, call.Arguments["team_id"], call.Arguments["team_name"])
		}
	}
	if tools := inferMutationToolsFromText(prompt); !containsString(tools, "create_team") || !containsString(tools, "write_file") {
		t.Fatalf("inferMutationToolsFromText(%q) = %v, want create_team and write_file", prompt, tools)
	}
	teamCalls := inferCreateTeamPlanCallsFromRequest(prompt, "")
	if len(teamCalls) != 2 || teamCalls[0].Arguments["team_id"] != "alpha-strike" || teamCalls[1].Arguments["team_id"] != "alpha-strike-review-team" {
		t.Fatalf("teamCalls = %#v, want primary team + companion review team", teamCalls)
	}
	if members, _ := teamCalls[0].Arguments["members"].(map[string]any); members["delivery-reviewer"] != "council-sentry" {
		t.Fatalf("primary team members = %#v, want delivery-reviewer: council-sentry", members)
	}
}

func TestInferUserDeliverableBundlePlanFromRequest_GeneralDevProjectBundle(t *testing.T) {
	calls, ok := inferUserDeliverableBundlePlanFromRequest("Create a clean todo tracker web app project with an interactive UI.")
	if !ok || len(calls) != 3 || calls[0].Arguments["package_kind"] != "dev_project_bundle" || calls[0].Arguments["path"] != "deliverables/projects/todo-app/index.html" {
		t.Fatalf("unexpected dev_project_bundle calls: ok=%v calls=%#v", ok, calls)
	}
	s := newTestServer()
	if review := s.buildSomaReferentialReview(t.Context(), []chatRequestMessage{{Role: "user", Content: "Build a team to create a calculator app project."}}); review.NeedsConfirmation {
		t.Fatalf("expected project/team development ask to go directly to proposal card: %+v", review)
	}
}

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

func TestFinalizeCodexTaskBundleTx_SupportsMarketingAndGameBundles(t *testing.T) {
	workspace := t.TempDir()
	t.Setenv("MYCELIS_WORKSPACE", workspace)

	planned, ok := inferUserDeliverableBundlePlanFromRequest("Create a playable sliding-tile puzzle game for the browser.")
	if !ok {
		t.Fatal("expected game bundle plan")
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
	summary, err := s.finalizeCodexTaskBundleTx(context.Background(), nil, planned, "run-game-1", "proof-game-1")
	if err != nil {
		t.Fatalf("finalizeCodexTaskBundleTx: %v", err)
	}
	if summary == nil {
		t.Fatal("expected non-nil package summary")
	}
	if summary.PackageKind != "interactive_game_bundle" {
		t.Fatalf("package_kind = %q, want interactive_game_bundle", summary.PackageKind)
	}
	if summary.PackageID != "interactive_game_bundle:sliding-tile-puzzle" {
		t.Fatalf("package_id = %q, want interactive_game_bundle:sliding-tile-puzzle", summary.PackageID)
	}
	if len(summary.Files) != 4 || len(summary.Artifacts) != 4 {
		t.Fatalf("files/artifacts = %d/%d, want 4/4", len(summary.Files), len(summary.Artifacts))
	}
	if summary.Files[0].ContentType != "text/html" || summary.Artifacts[0].Type != "code" {
		t.Fatalf("html file content_type/type = %q/%q, want text/html & code", summary.Files[0].ContentType, summary.Artifacts[0].Type)
	}
	if summary.Files[1].ContentType != "text/markdown" || summary.Artifacts[1].Type != "document" {
		t.Fatalf("md file content_type/type = %q/%q, want text/markdown & document", summary.Files[1].ContentType, summary.Artifacts[1].Type)
	}
	if _, err := os.Stat(filepath.Join(workspace, "deliverables", "games", "sliding-tile-puzzle", "proof.md")); err != nil {
		t.Fatalf("expected proof.md on disk: %v", err)
	}

	_ = protocol.PlannedToolCall{}
}

func TestInferMutationToolsFromText_IncludesWriteFileForPuzzleAndMarketingBundles(t *testing.T) {
	prompts := []string{
		"Create a simple browser sliding-tile puzzle game with playable HTML, game rules, and controls.",
		"Build a playable sliding-tile puzzle game in the browser.",
		"Create marketing documentation for our product launch including a positioning brief, launch email, and customer FAQ.",
	}
	for _, prompt := range prompts {
		tools := inferMutationToolsFromText(prompt)
		if !containsString(tools, "write_file") {
			t.Fatalf("inferMutationToolsFromText(%q) = %v, want write_file included", prompt, tools)
		}
	}
}

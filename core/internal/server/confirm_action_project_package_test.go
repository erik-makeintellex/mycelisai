package server

import (
	"strings"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

func TestExecutionOutputsFromToolResultsAddsReadmeFromProjectPackageContract(t *testing.T) {
	outputs := executionOutputsFromToolResults([]plannedToolExecutionResult{{
		Name: "write_file",
		Arguments: map[string]any{
			"path":               "workspace/generated/coin-runner/index.html",
			"content":            "<!doctype html><p>The package metadata must include README.md.</p>",
			"package_kind":       "project_package",
			"package_title":      "Coin Runner Game",
			"package_folder":     "workspace/generated/coin-runner",
			"package_entrypoint": "workspace/generated/coin-runner/index.html",
			"package_files":      []any{"index.html"},
			"validation_summary": "Browser opened and score increased after click.",
		},
		Output: "wrote playable game package",
	}})

	if len(outputs) != 1 {
		t.Fatalf("outputs = %#v, want 1", outputs)
	}
	files := outputs[0].Files
	if len(files) != 2 || files[0] != "index.html" || files[1] != "README.md" {
		t.Fatalf("files = %#v, want index.html and README.md", files)
	}
}

func TestExecutionOutputsFromToolResultsUsesProjectPackageTitleFromContract(t *testing.T) {
	outputs := executionOutputsFromToolResults([]plannedToolExecutionResult{{
		Name: "write_file",
		Arguments: map[string]any{
			"type":          "project_package",
			"package_title": "First Demo Game Team First Playable",
			"package_files": []any{"index.html", "README.md"},
		},
		Output: "Artifact stored.",
		Artifacts: []protocol.ChatArtifactRef{{
			ID:    "artifact-project-package",
			Type:  "project_package",
			Title: "first-demo-game-team-123 First Playable",
		}},
	}})

	if len(outputs) != 1 {
		t.Fatalf("outputs = %#v, want 1", outputs)
	}
	if outputs[0].Title != "First Demo Game Team First Playable" {
		t.Fatalf("title = %q, want package contract title", outputs[0].Title)
	}
}

func TestExecutionOutputsFromToolResultsRetainsStoredProjectPackageArtifact(t *testing.T) {
	outputs := executionOutputsFromToolResults([]plannedToolExecutionResult{{
		Name: "store_artifact",
		Arguments: map[string]any{
			"type":          "project_package",
			"title":         "Coin Runner Game",
			"package_files": []any{"index.html", "game.js", "README.md"},
		},
		Output: "Artifact stored.",
		Artifacts: []protocol.ChatArtifactRef{{
			ID:         "artifact-game",
			Type:       "project_package",
			Title:      "Coin Runner Game",
			Entrypoint: "workspace/generated/coin-runner/index.html",
			Folder:     "workspace/generated/coin-runner",
			Files:      []string{"index.html", "game.js"},
			Validation: "Browser opened and score increased after click.",
		}},
	}})

	if len(outputs) != 1 {
		t.Fatalf("outputs = %#v, want 1", outputs)
	}
	output := outputs[0]
	if output.ID != "artifact-game" || output.Kind != "project_package" {
		t.Fatalf("output identity = %#v", output)
	}
	if output.Href != "/api/v1/workspace/files/view?path=workspace%2Fgenerated%2Fcoin-runner%2Findex.html" {
		t.Fatalf("href = %q", output.Href)
	}
	if output.Summary != "Artifact stored." {
		t.Fatalf("summary = %q", output.Summary)
	}
	if len(output.Files) != 3 || output.Files[2] != "README.md" {
		t.Fatalf("files = %#v, want artifact files plus README.md from package metadata", output.Files)
	}
	if output.Retained == nil || !*output.Retained || output.RetentionClass != protocol.ExecutionRetentionClassRetained {
		t.Fatalf("retention = retained:%v class:%q", output.Retained, output.RetentionClass)
	}
}

func TestOutputRefsForTeamWorkProjectPackageUsesFolderStorageRef(t *testing.T) {
	outputs := executionOutputsForResult(plannedToolExecutionResult{
		Name: "write_file",
		Arguments: map[string]any{
			"team_id":            "first-demo-game-team",
			"path":               "groups/first-demo-game-team/generated/first-game/index.html",
			"content":            "<!doctype html><p>play</p>",
			"package_kind":       "project_package",
			"package_title":      "First Demo Game Team First Playable",
			"package_folder":     "groups/first-demo-game-team/generated/first-game",
			"package_entrypoint": "groups/first-demo-game-team/generated/first-game/index.html",
			"package_files":      []any{"index.html", "README.md"},
			"validation":         "Browser opened and score increased after click.",
		},
		Output: "wrote playable game package",
	})
	if len(outputs) != 1 {
		t.Fatalf("outputs = %#v, want 1", outputs)
	}
	if outputs[0].Href == "" || !strings.HasPrefix(outputs[0].Href, "/api/v1/workspace/files/view") {
		t.Fatalf("execution output href = %q, want viewer URL preserved", outputs[0].Href)
	}

	refs := outputRefsForTeamWork(testConfirmedActionTeamWorkLink(nil), "work-1", "first-demo-game-team", outputs)
	if len(refs) != 1 {
		t.Fatalf("refs = %#v, want 1", refs)
	}
	ref := refs[0]
	if ref.StorageRef != "groups/first-demo-game-team/generated/first-game" {
		t.Fatalf("storage_ref = %q, want project package folder", ref.StorageRef)
	}
	if strings.HasPrefix(ref.StorageRef, "/api/v1/workspace/files/view") {
		t.Fatalf("storage_ref = %q, must not be a viewer URL", ref.StorageRef)
	}
	if ref.Entrypoint != "index.html" {
		t.Fatalf("entrypoint = %q, want path relative to storage_ref", ref.Entrypoint)
	}
}

package server

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/pkg/protocol"
)

func TestEnsureGroupForCreatedTeamMirrorsConfirmedCreateTeam(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)
	mock.MatchExpectationsInOrder(true)

	auditID := "44444444-4444-4444-4444-444444444444"
	mock.ExpectQuery("FROM collaboration_groups").
		WithArgs("Research Team").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery("INSERT INTO collaboration_groups").
		WithArgs(
			sqlmock.AnyArg(), "default", "Research Team",
			"Map optimal agentry architecture.",
			"propose_only",
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			"researcher lead", "confirmed-chat-proposal", groupStatusActive,
			"test-user", nil, sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnRows(sqlmock.NewRows([]string{"created_at", "updated_at"}).AddRow(time.Now(), time.Now()))

	err := s.ensureGroupForCreatedTeam(t.Context(), auditID, "test-user", map[string]any{
		"team_id": "research-team",
		"name":    "Research Team",
		"role":    "researcher",
		"goal":    "Map optimal agentry architecture.",
	})
	if err != nil {
		t.Fatalf("ensure group for created team: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet db expectations: %v", err)
	}
}

func TestExecutionOutputsFromToolResultsRetainsTeamAndCodeFile(t *testing.T) {
	outputs := executionOutputsFromToolResults([]plannedToolExecutionResult{
		{
			Name: "create_team",
			Arguments: map[string]any{
				"team_id": "qa-browser-game-team",
				"name":    "QA Browser Game Team",
			},
			Output: "team created",
		},
		{
			Name: "write_file",
			Arguments: map[string]any{
				"path": "workspace/logs/qa_team_game.html",
			},
			Output: "wrote game file",
		},
	})

	if len(outputs) != 2 {
		t.Fatalf("outputs = %#v, want 2", outputs)
	}
	if outputs[0].Kind != "team" || outputs[0].ID != "qa-browser-game-team" || outputs[0].Href != "/groups" {
		t.Fatalf("team output = %#v", outputs[0])
	}
	if outputs[0].Retained == nil || !*outputs[0].Retained {
		t.Fatalf("team retained = %#v", outputs[0].Retained)
	}
	if outputs[1].Kind != "code" || outputs[1].ID != "workspace/logs/qa_team_game.html" {
		t.Fatalf("file output = %#v", outputs[1])
	}
	if outputs[1].Title != "workspace/logs/qa_team_game.html" {
		t.Fatalf("file title = %q", outputs[1].Title)
	}
	if outputs[1].Href != "/api/v1/workspace/files/view?path=workspace%2Flogs%2Fqa_team_game.html" {
		t.Fatalf("file href = %q", outputs[1].Href)
	}
	if outputs[1].Retained == nil || !*outputs[1].Retained {
		t.Fatalf("file retained = %#v", outputs[1].Retained)
	}
	if outputs[1].RetentionClass != protocol.ExecutionRetentionClassRetained {
		t.Fatalf("file retention = %q", outputs[1].RetentionClass)
	}
}

func TestExecutionOutputsFromToolResultsRetainsProjectPackage(t *testing.T) {
	outputs := executionOutputsFromToolResults([]plannedToolExecutionResult{
		{
			Name: "write_file",
			Arguments: map[string]any{
				"path":               "workspace/generated/coin-runner/index.html",
				"package_kind":       "project_package",
				"package_title":      "Coin Runner Game",
				"package_folder":     "workspace/generated/coin-runner",
				"package_entrypoint": "workspace/generated/coin-runner/index.html",
				"package_files":      []any{"index.html", "game.js", "styles.css", "README.md"},
				"validation_summary": "Browser opened and score increased after click.",
			},
			Output: "wrote playable game package",
		},
	})

	if len(outputs) != 1 {
		t.Fatalf("outputs = %#v, want 1", outputs)
	}
	output := outputs[0]
	if output.Kind != "project_package" {
		t.Fatalf("kind = %q, want project_package", output.Kind)
	}
	if output.Title != "Coin Runner Game" {
		t.Fatalf("title = %q", output.Title)
	}
	if output.Href != "/api/v1/workspace/files/view?path=workspace%2Fgenerated%2Fcoin-runner%2Findex.html" {
		t.Fatalf("href = %q", output.Href)
	}
	if output.Entrypoint != "workspace/generated/coin-runner/index.html" || output.Folder != "workspace/generated/coin-runner" {
		t.Fatalf("package paths = entry %q folder %q", output.Entrypoint, output.Folder)
	}
	if len(output.Files) != 4 || output.Files[1] != "game.js" {
		t.Fatalf("files = %#v", output.Files)
	}
	if output.Validation != "Browser opened and score increased after click." {
		t.Fatalf("validation = %q", output.Validation)
	}
	if output.Retained == nil || !*output.Retained {
		t.Fatalf("retained = %#v", output.Retained)
	}
}

func TestExecutionOutputsFromToolResultsRetainsStoredProjectPackageArtifact(t *testing.T) {
	outputs := executionOutputsFromToolResults([]plannedToolExecutionResult{
		{
			Name: "store_artifact",
			Arguments: map[string]any{
				"type":          "project_package",
				"title":         "Coin Runner Game",
				"package_files": []any{"index.html", "game.js"},
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
		},
	})

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
	if output.Retained == nil || !*output.Retained || output.RetentionClass != protocol.ExecutionRetentionClassRetained {
		t.Fatalf("retention = retained:%v class:%q", output.Retained, output.RetentionClass)
	}
}

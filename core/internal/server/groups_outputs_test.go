package server

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/mycelis/core/internal/artifacts"
)

func artifactColumns() []string {
	return []string{
		"id", "mission_id", "team_id", "agent_id", "trace_id", "artifact_type",
		"title", "content_type", "content", "file_path", "file_size_bytes",
		"metadata", "trust_score", "status", "created_at",
	}
}

func TestHandleGroupOutputs_ReturnsRetainedArtifacts(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt, func(s *AdminServer) {
		s.Artifacts = artifacts.NewService(s.DB, "/data/artifacts")
	})
	mux := setupMux(t, "GET /api/v1/groups/{id}/outputs", s.HandleGroupOutputs)

	now := time.Now().UTC()
	teamID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	mock.ExpectQuery("SELECT id::text, tenant_id, name, goal_statement, work_mode").
		WithArgs("group-temp").
		WillReturnRows(sqlmock.NewRows(collaborationGroupColumns()).
			AddRow(
				"group-temp",
				"default",
				"Temp Campaign",
				"Produce one campaign package",
				"execute_with_approval",
				[]byte(`["write_file"]`),
				[]byte(`["owner"]`),
				[]byte(`["11111111-1111-1111-1111-111111111111"]`),
				"marketing-lead",
				"",
				groupStatusArchived,
				"test-user-001",
				now.Add(2*time.Hour),
				"44444444-4444-4444-4444-444444444444",
				"55555555-5555-5555-5555-555555555555",
				now,
				now,
			))
	mock.ExpectQuery("SELECT .+ FROM artifacts WHERE team_id").
		WithArgs(teamID, 8).
		WillReturnRows(sqlmock.NewRows(artTestColumns).
			AddRow(
				"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
				nil,
				teamID,
				"marketing-lead",
				nil,
				"document",
				"Launch Brief",
				"text/markdown",
				"Campaign summary",
				nil,
				nil,
				[]byte(`{}`),
				0.9,
				"approved",
				now,
			))

	rr := doAuthenticatedRequest(t, mux, "GET", "/api/v1/groups/group-temp/outputs?limit=8", "")
	assertStatus(t, rr, http.StatusOK)

	var payload map[string]any
	assertJSON(t, rr, &payload)
	items, ok := payload["data"].([]any)
	if !ok {
		t.Fatalf("expected data array, got %T", payload["data"])
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 output, got %d", len(items))
	}
	first, ok := items[0].(map[string]any)
	if !ok {
		t.Fatalf("expected first output map, got %T", items[0])
	}
	if first["title"] != "Launch Brief" {
		t.Fatalf("title = %v, want Launch Brief", first["title"])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestHandleGroupOutputs_ReturnsArtifactsForArchivedGroup(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(
		dbOpt,
		func(s *AdminServer) {
			s.Artifacts = artifacts.NewService(s.DB, "")
		},
	)
	mux := setupMux(t, "GET /api/v1/groups/{id}/outputs", s.HandleGroupOutputs)

	now := time.Now().UTC()
	teamID1 := uuid.New()
	teamID2 := uuid.New()
	artifactID1 := uuid.New()
	artifactID2 := uuid.New()

	mock.ExpectQuery("SELECT id::text, tenant_id, name, goal_statement, work_mode").
		WithArgs("group-temp").
		WillReturnRows(sqlmock.NewRows(collaborationGroupColumns()).
			AddRow(
				"group-temp",
				"default",
				"Launch lane",
				"coordinate launch follow-through",
				"propose_only",
				[]byte(`["runs.read"]`),
				[]byte(`["u1"]`),
				[]byte(fmt.Sprintf(`["%s","%s"]`, teamID1.String(), teamID2.String())),
				"launch-profile",
				"policy.launch",
				groupStatusArchived,
				"test-user-001",
				nil,
				"",
				"",
				now,
				now,
			))

	mock.ExpectQuery("SELECT .+ FROM artifacts\\s+WHERE team_id = \\$1").
		WithArgs(teamID1, 3).
		WillReturnRows(sqlmock.NewRows(artifactColumns()).
			AddRow(
				artifactID1,
				nil,
				&teamID1,
				"launch-lead",
				nil,
				"document",
				"Launch summary",
				"text/markdown",
				"# Launch summary",
				nil,
				nil,
				[]byte(`{}`),
				nil,
				"approved",
				now.Add(-2*time.Minute),
			))

	mock.ExpectQuery("SELECT .+ FROM artifacts\\s+WHERE team_id = \\$1").
		WithArgs(teamID2, 3).
		WillReturnRows(sqlmock.NewRows(artifactColumns()).
			AddRow(
				artifactID2,
				nil,
				&teamID2,
				"review-lead",
				nil,
				"document",
				"Review checklist",
				"text/markdown",
				"# Review checklist",
				nil,
				nil,
				[]byte(`{}`),
				nil,
				"approved",
				now.Add(-1*time.Minute),
			))

	rr := doAuthenticatedRequest(t, mux, "GET", "/api/v1/groups/group-temp/outputs?limit=3", "")
	assertStatus(t, rr, http.StatusOK)

	var payload map[string]any
	assertJSON(t, rr, &payload)
	data, ok := payload["data"].([]any)
	if !ok {
		t.Fatalf("expected data array, got %T", payload["data"])
	}
	if len(data) != 2 {
		t.Fatalf("expected 2 outputs, got %d", len(data))
	}

	first, ok := data[0].(map[string]any)
	if !ok {
		t.Fatalf("expected first artifact object, got %T", data[0])
	}
	second, ok := data[1].(map[string]any)
	if !ok {
		t.Fatalf("expected second artifact object, got %T", data[1])
	}
	if first["title"] != "Review checklist" || second["title"] != "Launch summary" {
		t.Fatalf("unexpected artifact order: %#v", data)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

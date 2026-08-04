package swarm

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/catalogue"
)

var profileColumns = []string{
	"id", "name", "role", "system_prompt", "model", "tools", "inputs", "outputs",
	"verification_strategy", "verification_rubric", "validation_command", "profile_key",
	"description", "source", "locked", "capability_refs", "context_bindings", "usage_policy",
	"created_at", "updated_at",
}

func TestHydrateCreateTeamProfiles(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	now := time.Now()
	mock.ExpectQuery("SELECT .+ FROM agent_catalogue").
		WithArgs("default.researcher").
		WillReturnRows(sqlmock.NewRows(profileColumns).AddRow(
			"11111111-1111-1111-1111-111111111111", "Research Specialist", "researcher",
			"Research trusted sources.", "", []byte(`["web_search"]`), []byte(`[]`), []byte(`["source_summary"]`),
			"semantic", []byte(`["Cite evidence"]`), nil, "default.researcher",
			"Finds evidence", "built_in", true, []byte(`["web_search","read_text_file"]`),
			[]byte(`[{"kind":"public_web","access":"read"},{"kind":"approved_mount","access":"read"}]`),
			[]byte(`{"selection":"soma_or_manual","scope":"workspace"}`), now, now,
		))

	registry := &InternalToolRegistry{catalogue: catalogue.NewService(db)}
	member := map[string]any{
		"profile_ref": "default.researcher",
		"role":        "evidence-reviewer",
	}
	args := map[string]any{"agents": []map[string]any{member}}
	if err := registry.hydrateCreateTeamProfiles(context.Background(), args); err != nil {
		t.Fatalf("hydrate profiles: %v", err)
	}

	if member["role"] != "evidence-reviewer" {
		t.Fatalf("explicit role was overwritten: %#v", member["role"])
	}
	tools, ok := member["tools"].([]string)
	if !ok || len(tools) != 2 || tools[0] != "web_search" {
		t.Fatalf("capabilities not hydrated: %#v", member["tools"])
	}
	if member["context_bindings"] == nil || member["usage_policy"] == nil {
		t.Fatalf("context and usage policy must be hydrated: %#v", member)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestHydrateCreateTeamProfilesRejectsUnknownProfile(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectQuery("SELECT .+ FROM agent_catalogue").
		WithArgs("missing.profile").
		WillReturnRows(sqlmock.NewRows(profileColumns))

	registry := &InternalToolRegistry{catalogue: catalogue.NewService(db)}
	err = registry.hydrateCreateTeamProfiles(context.Background(), map[string]any{"profile_ref": "missing.profile"})
	if err == nil {
		t.Fatal("expected unknown profile to block team creation")
	}
}

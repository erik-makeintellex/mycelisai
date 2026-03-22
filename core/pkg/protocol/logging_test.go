package protocol

import "testing"

func TestOperationalLogContextNormalizeDefaultsCentralReview(t *testing.T) {
	ctx := OperationalLogContext{
		SourceChannel: "swarm.team.alpha.signal.status",
		TeamID:        "alpha",
		RunID:         "run-42",
		Summary:       "Alpha team reported a checkpoint.",
	}

	got := ctx.Normalize()

	if got.SchemaVersion != OperationalLogSchemaVersion {
		t.Fatalf("schema_version = %q, want %q", got.SchemaVersion, OperationalLogSchemaVersion)
	}
	if got.ReviewScope != LogReviewScopeCentralReview {
		t.Fatalf("review_scope = %q, want %q", got.ReviewScope, LogReviewScopeCentralReview)
	}
	if !got.CentralizedReview {
		t.Fatalf("centralized_review = false, want true")
	}
	if len(got.Audiences) != 3 {
		t.Fatalf("audiences len = %d, want 3", len(got.Audiences))
	}
	if got.ReviewChannels[0] != "memory.stream" {
		t.Fatalf("first review channel = %q, want memory.stream", got.ReviewChannels[0])
	}
	if got.ReviewChannels[1] != "swarm.team.alpha.signal.status" {
		t.Fatalf("second review channel = %q, want team signal status", got.ReviewChannels[1])
	}
	if got.ReviewChannels[2] != "swarm.mission.events.run-42" {
		t.Fatalf("third review channel = %q, want mission event path", got.ReviewChannels[2])
	}
}

func TestOperationalLogContextNormalizeAuditDefaults(t *testing.T) {
	ctx := OperationalLogContext{
		MissionEventID: "evt-1",
		Summary:        "Proposal approval recorded.",
	}

	got := ctx.Normalize()

	if got.ReviewScope != LogReviewScopeAudit {
		t.Fatalf("review_scope = %q, want %q", got.ReviewScope, LogReviewScopeAudit)
	}
	if len(got.Audiences) != 4 {
		t.Fatalf("audiences len = %d, want 4", len(got.Audiences))
	}
}

func TestParseOperationalLogContextNormalizesLooseMap(t *testing.T) {
	got := ParseOperationalLogContext(map[string]any{
		"summary":      "  Team lead review available  ",
		"review_scope": "team_local",
		"review_channels": []any{
			"swarm.team.alpha.signal.result",
			"swarm.team.alpha.signal.result",
		},
	})

	if got.Summary != "Team lead review available" {
		t.Fatalf("summary = %q", got.Summary)
	}
	if got.ReviewScope != LogReviewScopeTeamLocal {
		t.Fatalf("review_scope = %q", got.ReviewScope)
	}
	if len(got.ReviewChannels) != 2 {
		t.Fatalf("review_channels len = %d, want 2", len(got.ReviewChannels))
	}
	if got.ReviewChannels[0] != "swarm.team.alpha.signal.result" {
		t.Fatalf("first review channel = %q", got.ReviewChannels[0])
	}
	if got.ReviewChannels[1] != "memory.stream" {
		t.Fatalf("second review channel = %q, want memory.stream", got.ReviewChannels[1])
	}
}

package protocol

import "testing"

func TestNormalizeTeamAsk_DefaultsStructuredContract(t *testing.T) {
	ask := NormalizeTeamAsk(TeamAsk{
		Intent:     "  Review the current architecture docs  ",
		OwnedScope: []string{" docs/architecture-library ", ""},
	})

	if ask.SchemaVersion != "v1" {
		t.Fatalf("schema_version = %q, want v1", ask.SchemaVersion)
	}
	if ask.AskKind != TeamAskKindImplementation {
		t.Fatalf("ask_kind = %q, want %q", ask.AskKind, TeamAskKindImplementation)
	}
	if ask.LaneRole != TeamLaneRoleImplementer {
		t.Fatalf("lane_role = %q, want %q", ask.LaneRole, TeamLaneRoleImplementer)
	}
	if ask.Goal != "Review the current architecture docs" {
		t.Fatalf("goal = %q", ask.Goal)
	}
	if len(ask.OwnedScope) != 1 || ask.OwnedScope[0] != "docs/architecture-library" {
		t.Fatalf("owned_scope = %#v", ask.OwnedScope)
	}
}

func TestNormalizeTeamAsk_PreservesKnownKindsAndRoles(t *testing.T) {
	ask := NormalizeTeamAsk(TeamAsk{
		AskKind:      TeamAskKindValidation,
		LaneRole:     TeamLaneRoleValidator,
		Goal:         "Prove the first slice",
		ExitCriteria: []string{" tests green "},
	})

	if ask.AskKind != TeamAskKindValidation {
		t.Fatalf("ask_kind = %q, want %q", ask.AskKind, TeamAskKindValidation)
	}
	if ask.LaneRole != TeamLaneRoleValidator {
		t.Fatalf("lane_role = %q, want %q", ask.LaneRole, TeamLaneRoleValidator)
	}
	if len(ask.ExitCriteria) != 1 || ask.ExitCriteria[0] != "tests green" {
		t.Fatalf("exit_criteria = %#v", ask.ExitCriteria)
	}
}

func TestTeamAskIsZeroTreatsEmptyAskAsZero(t *testing.T) {
	if !(TeamAsk{}).IsZero() {
		t.Fatal("expected zero-value ask to be zero")
	}
	if (TeamAsk{Goal: "run checks"}).IsZero() {
		t.Fatal("expected goal-bearing ask to be non-zero")
	}
}

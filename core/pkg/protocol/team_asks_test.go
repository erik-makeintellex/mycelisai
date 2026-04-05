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

func TestTeamAskFromMap_NormalizesAliasFields(t *testing.T) {
	ask := TeamAskFromMap(map[string]any{
		"kind":                  "research",
		"role":                  "researcher",
		"intent":                "Map the governed browser proof",
		"owned_scope":           []any{" interface/e2e ", ""},
		"required_capabilities": []any{"read_file", "search_memory"},
	})

	if ask.AskKind != TeamAskKindResearch {
		t.Fatalf("ask_kind = %q", ask.AskKind)
	}
	if ask.LaneRole != TeamLaneRoleResearcher {
		t.Fatalf("lane_role = %q", ask.LaneRole)
	}
	if ask.Goal != "Map the governed browser proof" {
		t.Fatalf("goal = %q", ask.Goal)
	}
	if len(ask.OwnedScope) != 1 || ask.OwnedScope[0] != "interface/e2e" {
		t.Fatalf("owned_scope = %#v", ask.OwnedScope)
	}
	if len(ask.RequiredCapabilities) != 2 {
		t.Fatalf("required_capabilities = %#v", ask.RequiredCapabilities)
	}
}

func TestSummarizeTeamAsk_PrefersRoleAndGoal(t *testing.T) {
	summary := SummarizeTeamAsk(TeamAsk{
		AskKind:  TeamAskKindValidation,
		LaneRole: TeamLaneRoleValidator,
		Goal:     "Prove the release candidate from committed state",
	})

	if summary != "Validator ask: Prove the release candidate from committed state" {
		t.Fatalf("summary = %q", summary)
	}
}

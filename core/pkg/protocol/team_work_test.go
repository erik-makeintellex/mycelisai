package protocol

import "testing"

func TestNormalizeTeamWorkItem_DefaultsCreateTeamToNew(t *testing.T) {
	item := NormalizeTeamWorkItem(TeamWorkItem{
		TeamID:         "  research-team ",
		Objective:      "  Create the team shell  ",
		ExecutionShape: TeamExecutionShapeCreateTeam,
		Scope:          []string{" docs ", ""},
	})

	if item.TeamID != "research-team" {
		t.Fatalf("team_id = %q", item.TeamID)
	}
	if item.Objective != "Create the team shell" {
		t.Fatalf("objective = %q", item.Objective)
	}
	if item.State != TeamWorkStateNew {
		t.Fatalf("state = %q, want %q", item.State, TeamWorkStateNew)
	}
	if item.Version != "v1" {
		t.Fatalf("version = %q, want v1", item.Version)
	}
	if len(item.Scope) != 1 || item.Scope[0] != "docs" {
		t.Fatalf("scope = %#v", item.Scope)
	}
}

func TestValidateTeamWorkItem_RejectsWorkingCreateTeamState(t *testing.T) {
	item := NormalizeTeamWorkItem(TeamWorkItem{
		TeamID:         "research-team",
		Objective:      "Create the team shell",
		ExecutionShape: TeamExecutionShapeCreateTeam,
		State:          TeamWorkStateRunning,
	})

	if err := ValidateTeamWorkItem(item); err == nil {
		t.Fatal("expected running create_team work to be rejected")
	}
}

func TestNormalizeTeamWorkItem_DefaultsDelegatedWorkToQueued(t *testing.T) {
	item := NormalizeTeamWorkItem(TeamWorkItem{
		TeamID:         "implementation-team",
		Objective:      "Build the output package",
		ExecutionShape: TeamExecutionShapeDeliverable,
	})

	if item.State != TeamWorkStateQueued {
		t.Fatalf("state = %q, want %q", item.State, TeamWorkStateQueued)
	}
	if err := ValidateTeamWorkItem(item); err != nil {
		t.Fatalf("ValidateTeamWorkItem: %v", err)
	}
}

func TestValidateTeamWorkItem_AllowsManagedExecutionControlStates(t *testing.T) {
	for _, state := range []TeamWorkState{
		TeamWorkStateNeedsOperator,
		TeamWorkStateReviewing,
		TeamWorkStatePaused,
		TeamWorkStateArchived,
	} {
		item := NormalizeTeamWorkItem(TeamWorkItem{
			TeamID:         "implementation-team",
			Objective:      "Manage the output package",
			ExecutionShape: TeamExecutionShapeDelegatedWork,
			State:          state,
		})
		if err := ValidateTeamWorkItem(item); err != nil {
			t.Fatalf("ValidateTeamWorkItem(%q): %v", state, err)
		}
	}
}

func TestValidateTeamWorkItem_RejectsUnknownExecutionShape(t *testing.T) {
	item := NormalizeTeamWorkItem(TeamWorkItem{
		TeamID:         "implementation-team",
		Objective:      "Build the output package",
		ExecutionShape: TeamExecutionShape("unknown_shape"),
	})

	if err := ValidateTeamWorkItem(item); err == nil {
		t.Fatal("expected invalid execution_shape to be rejected")
	}
}

func TestValidateTeamInteraction_RequiresDurableSourceFields(t *testing.T) {
	interaction := NormalizeTeamInteraction(TeamInteraction{
		TeamID:        "qa-team",
		WorkItemID:    "11111111-1111-1111-1111-111111111111",
		Verb:          "brief",
		Summary:       "Review the acceptance criteria",
		PayloadKind:   "team_brief",
		SourceKind:    "workspace_ui",
		SourceChannel: "soma.team_work",
	})

	if err := ValidateTeamInteraction(interaction); err != nil {
		t.Fatalf("ValidateTeamInteraction: %v", err)
	}
	if interaction.Version != "v1" {
		t.Fatalf("version = %q, want v1", interaction.Version)
	}
}

func TestApplyTeamWorkAction_AllowsProductionControlTransitions(t *testing.T) {
	cases := []struct {
		name   string
		state  TeamWorkState
		action TeamWorkAction
		want   TeamWorkState
	}{
		{"start queued work", TeamWorkStateQueued, TeamWorkActionStartWork, TeamWorkStateRunning},
		{"pause running work", TeamWorkStateRunning, TeamWorkActionPause, TeamWorkStatePaused},
		{"resume paused work", TeamWorkStatePaused, TeamWorkActionResume, TeamWorkStateQueued},
		{"archive output ready work", TeamWorkStateOutputReady, TeamWorkActionArchive, TeamWorkStateArchived},
		{"steer running work", TeamWorkStateRunning, TeamWorkActionSteer, TeamWorkStateRunning},
		{"recover degraded work", TeamWorkStateDegraded, TeamWorkActionRecover, TeamWorkStateQueued},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			item := NormalizeTeamWorkItem(TeamWorkItem{
				TeamID:         "delivery-team",
				Objective:      "Produce the release package",
				ExecutionShape: TeamExecutionShapeDeliverable,
				State:          tc.state,
			})
			got, err := ApplyTeamWorkAction(item, tc.action)
			if err != nil {
				t.Fatalf("ApplyTeamWorkAction: %v", err)
			}
			if got != tc.want {
				t.Fatalf("state = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestApplyTeamWorkAction_RejectsInvalidTransitions(t *testing.T) {
	item := NormalizeTeamWorkItem(TeamWorkItem{
		TeamID:         "delivery-team",
		Objective:      "Produce the release package",
		ExecutionShape: TeamExecutionShapeDeliverable,
		State:          TeamWorkStateOutputReady,
	})
	if _, err := ApplyTeamWorkAction(item, TeamWorkActionStartWork); err == nil {
		t.Fatal("expected output_ready start_work to be rejected")
	}

	createTeam := NormalizeTeamWorkItem(TeamWorkItem{
		TeamID:         "delivery-team",
		Objective:      "Create the team shell",
		ExecutionShape: TeamExecutionShapeCreateTeam,
		State:          TeamWorkStateNew,
	})
	if _, err := ApplyTeamWorkAction(createTeam, TeamWorkActionArchive); err == nil {
		t.Fatal("expected create_team archive to be rejected")
	}

	if _, err := ApplyTeamWorkAction(item, TeamWorkActionRecover); err == nil {
		t.Fatal("expected output_ready recover to be rejected")
	}
}

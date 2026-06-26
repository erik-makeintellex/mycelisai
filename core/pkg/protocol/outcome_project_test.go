package protocol

import "testing"

func TestNormalizeOutcomeProjectDefaultsOwnershipFields(t *testing.T) {
	item := NormalizeOutcomeProject(OutcomeProject{
		ProjectID: "project-1",
		OutcomeID: "outcome-1",
		Title:     "Launch package",
	})

	if item.ExecutionMode != "project" {
		t.Fatalf("execution_mode = %q, want project", item.ExecutionMode)
	}
	if item.Status != OutcomeProjectStatusActive {
		t.Fatalf("status = %q, want active", item.Status)
	}
	if item.RetentionPolicy != "retained" {
		t.Fatalf("retention_policy = %q, want retained", item.RetentionPolicy)
	}
	if item.Version != "v1" {
		t.Fatalf("version = %q, want v1", item.Version)
	}
	if item.TargetRef == nil {
		t.Fatal("expected target_ref")
	}
	if item.TargetRef.Type != "outcome_project" || item.TargetRef.ID != "project-1" {
		t.Fatalf("target_ref = %#v, want outcome_project target", item.TargetRef)
	}
	if err := ValidateOutcomeProject(item); err != nil {
		t.Fatalf("ValidateOutcomeProject: %v", err)
	}
}

func TestValidateTeamRegistryEntryRequiresOwnerTarget(t *testing.T) {
	item := NormalizeTeamRegistryEntry(TeamRegistryEntry{
		ProjectID: "project-1",
		Role:      "lead",
	})

	if err := ValidateTeamRegistryEntry(item); err == nil {
		t.Fatalf("ValidateTeamRegistryEntry accepted missing team_id/agent_id")
	}
	item.TeamID = "qa-team"
	if err := ValidateTeamRegistryEntry(item); err != nil {
		t.Fatalf("ValidateTeamRegistryEntry with team_id: %v", err)
	}
}

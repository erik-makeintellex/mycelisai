package server

import (
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

func TestDeliverableResultOutputIssue_RejectsGenericFileForPackageOutcome(t *testing.T) {
	item := protocol.TeamWorkItem{
		TeamID:          "game-team",
		ExecutionShape:  protocol.TeamExecutionShapeDelegatedWork,
		ExpectedOutputs: []string{"openable browser game package"},
	}
	refs := []protocol.TeamOutputRef{{
		Kind:       "file",
		StorageRef: "generated/how-do-you-think-you-could-improve-it.py",
	}}
	if got := deliverableResultOutputIssue(item, protocol.PayloadKindResult, refs); got != "invalid_deliverable_shape" {
		t.Fatalf("issue = %q, want invalid_deliverable_shape", got)
	}
}

func TestDeliverableResultOutputIssue_AcceptsIsolatedOpenablePackage(t *testing.T) {
	item := protocol.TeamWorkItem{
		TeamID:          "game-team",
		ExecutionShape:  protocol.TeamExecutionShapeDelegatedWork,
		ExpectedOutputs: []string{"openable browser game package"},
	}
	refs := []protocol.TeamOutputRef{{
		Kind:       "project_package",
		StorageRef: "groups/game-team/generated/first-game",
		Entrypoint: "index.html",
	}}
	if got := deliverableResultOutputIssue(item, protocol.PayloadKindResult, refs); got != "" {
		t.Fatalf("issue = %q, want accepted package", got)
	}
}

package server

import (
	"testing"

	"github.com/mycelis/core/internal/runs"
	"github.com/mycelis/core/pkg/protocol"
)

func TestConfirmActionResponseDataIncludesTeamWorkRefs(t *testing.T) {
	refs := []confirmActionTeamWorkRef{
		{
			WorkItemID: "work-1",
			TeamID:     "qa-team",
			State:      protocol.TeamWorkStateQueued,
			RunID:      "run-1",
		},
	}

	data := confirmActionResponseData("proof-1", "contract-1", "artifact-1", "run-1", "audit-1", &protocol.ScopeValidation{}, nil, refs, nil)

	if data["confirmed"] != true || data["verified"] != true {
		t.Fatalf("response flags = confirmed:%v verified:%v", data["confirmed"], data["verified"])
	}
	if data["run_status"] != runs.StatusCompleted {
		t.Fatalf("run_status = %v, want %s", data["run_status"], runs.StatusCompleted)
	}
	got, ok := data["team_work_refs"].([]confirmActionTeamWorkRef)
	if !ok {
		t.Fatalf("team_work_refs = %T, want []confirmActionTeamWorkRef", data["team_work_refs"])
	}
	if len(got) != 1 || got[0].WorkItemID != "work-1" || got[0].TeamID != "qa-team" || got[0].State != protocol.TeamWorkStateQueued || got[0].RunID != "run-1" {
		t.Fatalf("team_work_refs = %#v", got)
	}
}

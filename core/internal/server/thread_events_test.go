package server

import (
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

func TestFirstThreadEventTeamWorkRefAggregatesOutputRefs(t *testing.T) {
	teamID, workItemID, outputs := firstThreadEventTeamWorkRef([]confirmActionTeamWorkRef{
		{
			TeamID:     "marketing-team",
			WorkItemID: "create-work",
			State:      protocol.TeamWorkStateNew,
		},
		{
			TeamID:     "source-team",
			WorkItemID: "source-output",
			State:      protocol.TeamWorkStateOutputReady,
			OutputRefs: []protocol.TeamOutputRef{{OutputID: "source-evidence", StorageRef: "groups/source-team/proof/CHANGE_EXAMPLES.md"}},
		},
		{
			TeamID:     "marketing-team",
			WorkItemID: "handoff-output",
			State:      protocol.TeamWorkStateOutputReady,
			OutputRefs: []protocol.TeamOutputRef{{OutputID: "marketing-handoff", StorageRef: "groups/marketing-team/marketing/MARKETING_HANDOFF.md"}},
		},
	})

	if teamID != "marketing-team" || workItemID != "create-work" {
		t.Fatalf("thread target = %s/%s, want first durable work ref", teamID, workItemID)
	}
	if len(outputs) != 2 {
		t.Fatalf("outputs = %#v, want source and downstream refs", outputs)
	}
	if outputs[0].StorageRef != "groups/source-team/proof/CHANGE_EXAMPLES.md" || outputs[1].StorageRef != "groups/marketing-team/marketing/MARKETING_HANDOFF.md" {
		t.Fatalf("outputs = %#v, want both retained refs in order", outputs)
	}
}

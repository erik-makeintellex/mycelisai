package server

import (
	"strings"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

func TestTeamWorkResultThreadEvent_ReturnsDirectPackageLink(t *testing.T) {
	item := protocol.TeamWorkItem{
		TeamID:     "game-team",
		WorkItemID: "work-1",
		RunID:      "run-1",
		State:      protocol.TeamWorkStateOutputReady,
		OutputRefs: []protocol.TeamOutputRef{{
			Kind:       "project_package",
			Label:      "Playable game",
			StorageRef: "groups/game-team/generated/first-game",
			Entrypoint: "index.html",
		}},
	}
	event := teamWorkResultThreadEvent(item, protocol.TeamStatusEvent{Details: "The team built and validated the playable game."})
	if event.Payload.Kind != protocol.ThreadEventResultReady || event.Payload.Label != "Work complete" {
		t.Fatalf("payload = %#v", event.Payload)
	}
	if event.Payload.HrefLabel != "Open app" || !strings.Contains(event.Payload.Href, "groups%2Fgame-team%2Fgenerated%2Ffirst-game%2Findex.html") {
		t.Fatalf("open action = %q %q", event.Payload.HrefLabel, event.Payload.Href)
	}
	if !strings.Contains(event.Payload.Detail, "One deliverable is ready to open") {
		t.Fatalf("detail = %q", event.Payload.Detail)
	}
}

func TestTeamWorkResultThreadEvent_PrefersAuthoritativeCompletionProof(t *testing.T) {
	item := protocol.TeamWorkItem{
		ProofID:   "dispatch-proof",
		ProofRefs: []string{"dispatch-proof", "completion-proof"},
		State:     protocol.TeamWorkStateOutputReady,
		OutputRefs: []protocol.TeamOutputRef{{
			Kind: "project_package", ProofID: "completion-proof", ProofRef: "completion-proof",
		}},
	}

	event := teamWorkResultThreadEvent(item, protocol.TeamStatusEvent{})
	if event.Payload.ProofID != "completion-proof" {
		t.Fatalf("proof id = %q, want authoritative completion proof", event.Payload.ProofID)
	}
}

func TestTeamWorkResultThreadEvent_ExplainsInvalidResultWithoutRunTimeline(t *testing.T) {
	item := protocol.TeamWorkItem{
		TeamID:           "game-team",
		WorkItemID:       "work-1",
		RunID:            "run-1",
		State:            protocol.TeamWorkStateDegraded,
		DegradationState: "invalid_deliverable_shape",
		OutputRefs: []protocol.TeamOutputRef{{
			Kind:       "project_package",
			Label:      "Unvalidated game",
			StorageRef: "groups/game-team/generated/first-game",
			Entrypoint: "index.html",
		}},
	}
	event := teamWorkResultThreadEvent(item, protocol.TeamStatusEvent{Details: "The expected package entrypoint was not retained."})
	if event.Payload.Kind != protocol.ThreadEventAttentionNeeded || event.Payload.Href != "" || event.Payload.HrefLabel != "" {
		t.Fatalf("payload = %#v, want attention without raw run link", event.Payload)
	}
}

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

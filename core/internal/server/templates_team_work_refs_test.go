package server

import (
	"strings"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

func assertTeamWorkRefsForConfirmAction(t *testing.T, data map[string]any, teamID, runID string) {
	t.Helper()
	teamWorkRefs, ok := data["team_work_refs"].([]any)
	if !ok {
		t.Fatalf("expected team_work_refs list, got %T", data["team_work_refs"])
	}
	if len(teamWorkRefs) != 2 {
		t.Fatalf("team_work_refs = %#v, want create-team and deliverable refs", teamWorkRefs)
	}
	if !teamWorkRefsHas(teamWorkRefs, teamID, string(protocol.TeamWorkStateNew), runID, false) {
		t.Fatalf("team_work_refs missing create-team ref: %#v", teamWorkRefs)
	}
	if !teamWorkRefsHas(teamWorkRefs, teamID, string(protocol.TeamWorkStateOutputReady), runID, true) {
		t.Fatalf("team_work_refs missing output-ready ref with output_refs: %#v", teamWorkRefs)
	}
}

func teamWorkRefsHas(refs []any, teamID, state, runID string, requiresOutputRefs bool) bool {
	for _, raw := range refs {
		ref, ok := raw.(map[string]any)
		if !ok || ref["team_id"] != teamID || ref["state"] != state || ref["run_id"] != runID {
			continue
		}
		workItemID, _ := ref["work_item_id"].(string)
		if strings.TrimSpace(workItemID) == "" {
			return false
		}
		if !requiresOutputRefs {
			return true
		}
		return teamWorkOutputRefsHas(ref, teamID, workItemID, runID)
	}
	return false
}

func teamWorkOutputRefsHas(ref map[string]any, teamID, workItemID, runID string) bool {
	outputRefs, ok := ref["output_refs"].([]any)
	if !ok || len(outputRefs) == 0 {
		return false
	}
	for _, rawOutputRef := range outputRefs {
		outputRef, ok := rawOutputRef.(map[string]any)
		if ok && outputRef["team_id"] == teamID && outputRef["work_item_id"] == workItemID && outputRef["run_id"] == runID {
			return true
		}
	}
	return false
}

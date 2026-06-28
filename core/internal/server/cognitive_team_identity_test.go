package server

import (
	"strings"
	"testing"
)

func TestInferCreateTeamPlanFromRequest_GeneratedTeamIDIsStableForProposalReplay(t *testing.T) {
	request := strings.Join([]string{
		"Soma, generate an original console-era action adventure game as a hard team-delivery test.",
		"Use research and council review before choosing the team and implementation path.",
		"The final user output should be a playable browser app with code-generated graphics, keyboard controls, enemies, pickups, audio, direct launch access, and proof notes.",
		"If this needs a team, create the smallest useful team and have them produce the game.",
	}, " ")
	first, ok := inferCreateTeamPlanFromRequest(request)
	if !ok {
		t.Fatal("expected first create team plan")
	}
	second, ok := inferCreateTeamPlanFromRequest(request)
	if !ok {
		t.Fatal("expected second create team plan")
	}
	if first.Arguments["team_id"] != second.Arguments["team_id"] {
		t.Fatalf("team ids drifted across proposal replay: %#v != %#v", first.Arguments["team_id"], second.Arguments["team_id"])
	}
}

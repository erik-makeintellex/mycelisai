package server

import (
	"testing"

	"github.com/mycelis/core/internal/swarm"
	"github.com/mycelis/core/pkg/protocol"
)

func TestExecutePlannedToolCallsWiresSomaForCreateTeam(t *testing.T) {
	wireNATS := withNATS(t)
	s := newTestServer(wireNATS, func(s *AdminServer) {
		s.Soma = swarm.NewSoma(s.NC, nil, nil, nil, nil, nil, nil)
		t.Cleanup(s.Soma.Shutdown)
	})

	scope := &protocol.ScopeValidation{
		Tools: []string{"create_team"},
		PlannedToolCalls: []protocol.PlannedToolCall{
			{
				Name: "create_team",
				Arguments: map[string]any{
					"team_id": "research-team",
					"name":    "Research Team",
					"role":    "researcher",
				},
			},
		},
		CapabilityIDs: []string{"team_orchestration"},
	}

	results, err := s.executePlannedToolCalls(t.Context(), scope, "test-user")
	if err != nil {
		t.Fatalf("execute planned create_team: %v", err)
	}
	if len(results) != 1 || results[0].Name != "create_team" {
		t.Fatalf("expected create_team execution result, got %#v", results)
	}

	for _, team := range s.Soma.ListTeams() {
		if team.ID == "research-team" {
			return
		}
	}
	t.Fatalf("expected confirmed create_team to spawn research-team, got %#v", s.Soma.ListTeams())
}

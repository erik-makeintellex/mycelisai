package state_test

import (
	"testing"

	"github.com/mycelis/core/internal/state"
)

func TestUpdateHeartbeat_Teams(t *testing.T) {
	reg := state.NewRegistry()

	// 1. Register agents joined to teams
	reg.UpdateHeartbeat("agent-a", "marketing", "swarm:base", state.StatusIdle)
	reg.UpdateHeartbeat("agent-b", "sensors", "ros2:lidar", state.StatusBusy)
	reg.UpdateHeartbeat("agent-c", "marketing", "swarm:llm", state.StatusIdle)

	// 2. Query Marketing Team
	teamMkt := reg.GetAgentsByTeam("marketing")
	if len(teamMkt) != 2 {
		t.Errorf("Expected 2 agents in marketing, got %d", len(teamMkt))
	}

	// 3. Query Sensors Team
	teamSens := reg.GetAgentsByTeam("sensors")
	if len(teamSens) != 1 {
		t.Errorf("Expected 1 agent in sensors, got %d", len(teamSens))
	}
	if teamSens[0].SourceURI != "ros2:lidar" {
		t.Errorf("Expected SourceURI 'ros2:lidar', got %s", teamSens[0].SourceURI)
	}

	// 4. Update an agent's team (Migration)
	// Move agent-a to 'engineering'
	reg.UpdateHeartbeat("agent-a", "engineering", "swarm:base", state.StatusIdle)

	newMkt := reg.GetAgentsByTeam("marketing")
	if len(newMkt) != 1 {
		t.Errorf("Expected 1 agent remaining in marketing, got %d", len(newMkt))
	}

	eng := reg.GetAgentsByTeam("engineering")
	if len(eng) != 1 {
		t.Errorf("Expected 1 agent in engineering, got %d", len(eng))
	}
}

func TestActiveThreshold(t *testing.T) {
	// Optional: verify timeout logic still works with teams
	reg := state.NewRegistry()
	reg.UpdateHeartbeat("ghost", "shadow", "void", state.StatusIdle)

	// Manually age the record (requires accessing map or waiting,
	// for unit test we rely on public API availability or simple check)
	// Since we can't easily mock time in simple version, we skip complex time manip
	// unless we inject a clock.

	agents := reg.GetAgentsByTeam("shadow")
	if len(agents) != 1 {
		t.Error("Should see active agent")
	}
}

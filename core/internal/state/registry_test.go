package state_test

import (
	"sync"
	"testing"
	"time"

	"github.com/mycelis/core/internal/state"
)

// consC3TeamCount counts active agents on a team through the production read path.
func consC3TeamCount(reg *state.Registry, teamID string) int {
	count := 0
	for _, agent := range reg.GetActiveAgents() {
		if agent.TeamID == teamID {
			count++
		}
	}
	return count
}

func TestUpdateHeartbeat_Teams(t *testing.T) {
	reg := &state.Registry{}

	reg.UpdateHeartbeat("agent-a", "marketing", "swarm:base", state.StatusIdle)
	reg.UpdateHeartbeat("agent-b", "sensors", "ros2:lidar", state.StatusBusy)
	reg.UpdateHeartbeat("agent-c", "marketing", "swarm:llm", state.StatusIdle)

	if got := consC3TeamCount(reg, "marketing"); got != 2 {
		t.Errorf("Expected 2 agents in marketing, got %d", got)
	}
	if got := consC3TeamCount(reg, "sensors"); got != 1 {
		t.Errorf("Expected 1 agent in sensors, got %d", got)
	}
	if b, _ := reg.Get("agent-b"); b.SourceURI != "ros2:lidar" {
		t.Errorf("Expected SourceURI 'ros2:lidar', got %s", b.SourceURI)
	}

	// Moving agent-a to engineering updates its team.
	reg.UpdateHeartbeat("agent-a", "engineering", "swarm:base", state.StatusIdle)
	if got := consC3TeamCount(reg, "marketing"); got != 1 {
		t.Errorf("Expected 1 agent remaining in marketing, got %d", got)
	}
	if got := consC3TeamCount(reg, "engineering"); got != 1 {
		t.Errorf("Expected 1 agent in engineering, got %d", got)
	}

	// A partial heartbeat (empty team/source) keeps the existing values.
	reg.UpdateHeartbeat("agent-b", "", "", state.StatusIdle)
	b, ok := reg.Get("agent-b")
	if !ok || b.TeamID != "sensors" {
		t.Error("Agent should persist in team on partial heartbeat")
	}
	if b.SourceURI != "ros2:lidar" {
		t.Errorf("SourceURI should persist. Got %s", b.SourceURI)
	}
}

func TestActiveThreshold(t *testing.T) {
	reg := &state.Registry{}
	reg.UpdateHeartbeat("ghost", "shadow", "void", state.StatusIdle)
	if got := consC3TeamCount(reg, "shadow"); got != 1 {
		t.Error("Should see active agent")
	}
}

func TestRefreshKnownNeverCreatesOrRewrites(t *testing.T) {
	reg := &state.Registry{}
	if reg.RefreshKnown("unknown") || reg.RefreshKnown("") {
		t.Fatal("refresh must not succeed for an unknown or empty agent")
	}
	if _, ok := reg.Get("unknown"); ok {
		t.Fatal("refresh must never create an agent")
	}
	reg.UpdateHeartbeat("known", "alpha", "swarm:base", state.StatusIdle)
	before, _ := reg.Get("known")
	time.Sleep(2 * time.Millisecond)
	if !reg.RefreshKnown("known") {
		t.Fatal("refresh must succeed for a registered agent")
	}
	after, _ := reg.Get("known")
	if !after.LastHeartbeat.After(before.LastHeartbeat) {
		t.Fatal("last seen must advance")
	}
	if after.TeamID != "alpha" || after.SourceURI != "swarm:base" || after.Status != before.Status {
		t.Fatalf("refresh must not rewrite metadata: %+v", after)
	}
}

func TestRefreshKnownConcurrent(t *testing.T) {
	reg := &state.Registry{}
	reg.UpdateHeartbeat("known", "alpha", "swarm:base", state.StatusIdle)
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(2)
		go func() { defer wg.Done(); reg.RefreshKnown("known") }()
		go func() { defer wg.Done(); _ = reg.GetActiveAgents() }()
	}
	wg.Wait()
	if got, _ := reg.Get("known"); got.TeamID != "alpha" {
		t.Fatalf("team changed: %+v", got)
	}
}

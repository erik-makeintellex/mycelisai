package swarm

import (
	"slices"
	"testing"
)

// Owner decision (H1 Q2): hand_off and read_handoff_input go to team leads
// only. In the shipped config every team has at most one designated
// team_lead; each lead that holds the handoff tools also holds
// store_artifact; no non-lead holds them; and the optional dev-swarm teams
// (prime-architect, prime-development) each ship exactly one such lead in both
// the team files and the optional bundle.
func TestShippedHandoffToolsAreDeclaredOnLeadsOnly(t *testing.T) {
	type teamKey struct{ file, team string }
	leads := map[teamKey]int{}
	holders := map[teamKey]int{}
	for _, agent := range shippedAgents(t) {
		key := teamKey{agent.file, agent.teamID}
		isLead := handoffLeadAuthority(ToolInvocationContext{AgentID: agent.id, TeamID: agent.teamID, AgentRole: agent.role})
		if isLead {
			leads[key]++
		}
		hasHandOff := slices.Contains(agent.tools, "hand_off")
		hasRead := slices.Contains(agent.tools, "read_handoff_input")
		if !hasHandOff && !hasRead {
			continue
		}
		if !isLead {
			t.Errorf("%s %s/%s declares handoff tools but is not the team lead (role %q)", agent.file, agent.teamID, agent.id, agent.role)
		}
		if !hasHandOff || !hasRead || !slices.Contains(agent.tools, "store_artifact") {
			t.Errorf("%s %s/%s must declare store_artifact, hand_off and read_handoff_input together", agent.file, agent.teamID, agent.id)
		}
		holders[key]++
	}
	for key, count := range leads {
		if count > 1 {
			t.Errorf("%s team %s has %d team leads, want at most one", key.file, key.team, count)
		}
	}
	for _, key := range []teamKey{
		{"prime-architect.yaml", "prime-architect"}, {"prime-development.yaml", "prime-development"},
		{"mycelis-dev-swarm-optional.yaml", "prime-architect"}, {"mycelis-dev-swarm-optional.yaml", "prime-development"},
	} {
		if leads[key] != 1 || holders[key] != 1 {
			t.Errorf("%s team %s: leads=%d handoff holders=%d, want exactly one lead holding the tools", key.file, key.team, leads[key], holders[key])
		}
	}
	// BOOT-B: default boot (mycelis-runtime-core) ships no team lead, so
	// hand_off has no shipped holder unless the optional bundle is selected.
	runtimeCoreAgents := 0
	for _, agent := range shippedAgents(t) {
		if agent.file != "mycelis-runtime-core.yaml" {
			continue
		}
		runtimeCoreAgents++
		if slices.Contains(agent.tools, "hand_off") || slices.Contains(agent.tools, "read_handoff_input") {
			t.Errorf("default boot bundle %s/%s declares handoff tools", agent.teamID, agent.id)
		}
	}
	if runtimeCoreAgents == 0 {
		t.Error("mycelis-runtime-core.yaml not found among shipped templates")
	}
}

func TestHandoffLeadAuthorityExcludesSomaAndSpecialists(t *testing.T) {
	for _, inv := range []ToolInvocationContext{
		{AgentID: "admin", TeamID: "admin-core", AgentRole: "admin"},
		{AgentID: "research-analyst", TeamID: "research", AgentRole: "researcher"},
		{AgentID: "observer", TeamID: "telemetry-core", AgentRole: "data_visualizer"},
		{AgentID: "", TeamID: "research", AgentRole: "lead"},
		// Specialist roles that isLeadAgent treats as leads for prompt context
		// are not designated team leads.
		{AgentID: "prime-development-agent", TeamID: "prime-development", AgentRole: "coder"},
		{AgentID: "prime-architect-agent", TeamID: "prime-architect", AgentRole: "architect"},
		{AgentID: "comic-layout", TeamID: "comic", AgentRole: "creative"},
		{AgentID: "council-sentry", TeamID: "council-core", AgentRole: "sentry"},
		{AgentID: "comic-story", TeamID: "comic", AgentRole: "story lead"},
		{AgentID: "research-lead", TeamID: "research", AgentRole: "coder"},
		{AgentID: "research-lead", TeamID: "", AgentRole: "lead"},
	} {
		if handoffLeadAuthority(inv) {
			t.Fatalf("%#v must not have handoff lead authority", inv)
		}
	}
	for _, inv := range []ToolInvocationContext{
		{AgentID: "research-lead", TeamID: "research", AgentRole: "lead"},
		{AgentID: "marketing-agent", TeamID: "marketing", AgentRole: "team lead"},
		{AgentID: "research-agent", TeamID: "research", AgentRole: "team_lead"},
	} {
		if !handoffLeadAuthority(inv) {
			t.Fatalf("%#v must have handoff lead authority", inv)
		}
	}
}

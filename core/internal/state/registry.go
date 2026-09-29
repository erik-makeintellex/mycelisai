package state

import (
	"sync"
	"time"
)

// AgentStatus represents the simplified high-level state of an agent
type AgentStatus int

const (
	StatusOffline AgentStatus = iota
	StatusIdle
	StatusBusy // Processing / Thinking
	StatusError
)

// AgentState holds the runtime metadata for a single agent
type AgentState struct {
	ID            string      `json:"id"`
	Name          string      `json:"name"`
	TeamID        string      `json:"team_id"`
	SourceURI     string      `json:"source_uri"`
	Status        AgentStatus `json:"status"`
	LastHeartbeat time.Time   `json:"last_heartbeat"`
}

// Registry is the thread-safe store for all active agents
type Registry struct {
	agents sync.Map
}

// Global registry instance. The zero value is ready to use.
var GlobalRegistry = &Registry{}

// UpdateHeartbeat refreshes the state of an agent based on an incoming signal
func (r *Registry) UpdateHeartbeat(agentID, teamID, sourceURI string, status AgentStatus) {
	if agentID == "" {
		return
	}
	now := time.Now()

	val, loaded := r.agents.LoadOrStore(agentID, &AgentState{
		ID:            agentID,
		Name:          agentID,
		TeamID:        teamID,
		SourceURI:     sourceURI,
		Status:        status,
		LastHeartbeat: now,
	})

	state := val.(*AgentState)

	if loaded {
		state.Status = status
		state.LastHeartbeat = now
		if teamID != "" {
			state.TeamID = teamID
		}
		if sourceURI != "" {
			state.SourceURI = sourceURI
		}
	}
}

// RefreshKnown advances last-seen for an agent that is already registered and
// reports whether it did. It never creates an agent and never changes TeamID or
// SourceURI (A2b: the degraded-governance heartbeat path). The entry is
// replaced copy-on-write so readers holding the old pointer are unaffected.
func (r *Registry) RefreshKnown(agentID string) bool {
	if agentID == "" {
		return false
	}
	for {
		val, ok := r.agents.Load(agentID)
		if !ok {
			return false
		}
		next := *val.(*AgentState)
		next.LastHeartbeat = time.Now()
		if r.agents.CompareAndSwap(agentID, val, &next) {
			return true
		}
	}
}

// Get returns a copy of a registered agent's state.
func (r *Registry) Get(agentID string) (AgentState, bool) {
	val, ok := r.agents.Load(agentID)
	if !ok {
		return AgentState{}, false
	}
	return *val.(*AgentState), true
}

// GetActiveAgents returns a list of agents seen in the last 30 seconds
func (r *Registry) GetActiveAgents() []*AgentState {
	active := []*AgentState{}
	threshold := time.Now().Add(-30 * time.Second)

	r.agents.Range(func(key, value interface{}) bool {
		state := value.(*AgentState)
		if state.LastHeartbeat.After(threshold) {
			active = append(active, state)
		}
		return true
	})

	return active
}

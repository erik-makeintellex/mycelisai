package state

import (
	"sync"
	"time"

	pb "github.com/mycelis/core/pkg/pb/swarm"
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
	ID            string
	Name          string
	TeamID        string // Logical Group (e.g. "marketing")
	SourceURI     string // Power Source (e.g. "swarm:base")
	Status        AgentStatus
	LastHeartbeat time.Time
	// We can extend this with CurrentTaskID, capabilities, etc.
}

// Registry is the thread-safe store for all active agents
type Registry struct {
	// agents maps AgentID (string) -> *AgentState
	agents sync.Map
}

// Global registry instance (could be injected, keeping it simple for V1)
var GlobalRegistry = &Registry{}

// NewRegistry creates a fresh registry (useful for tests)
func NewRegistry() *Registry {
	return &Registry{}
}

// UpdateHeartbeat refreshes the state of an agent based on an incoming signal
func (r *Registry) UpdateHeartbeat(agentID, teamID, sourceURI string, status AgentStatus) {
	now := time.Now()

	// LoadOrStore handles the race of creating a new entry
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
		// Update existing
		state.Status = status
		state.LastHeartbeat = now
		// Allow updating metadata if it changes
		if teamID != "" {
			state.TeamID = teamID
		}
		if sourceURI != "" {
			state.SourceURI = sourceURI
		}
	}
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

// GetAgentsByTeam filters active agents by TeamID
func (r *Registry) GetAgentsByTeam(teamID string) []*AgentState {
	active := []*AgentState{}
	threshold := time.Now().Add(-30 * time.Second)

	r.agents.Range(func(key, value interface{}) bool {
		state := value.(*AgentState)
		if state.LastHeartbeat.After(threshold) && state.TeamID == teamID {
			active = append(active, state)
		}
		return true
	})

	return active
}

// ToProtoStatus converts internal status to Protobuf if needed
func ToProtoStatus(s AgentStatus) pb.AgentConfig {
	// Placeholder for mapping to proto enum if strictly required
	return pb.AgentConfig{}
}

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
func (r *Registry) UpdateHeartbeat(agentID string, status AgentStatus) {
	now := time.Now()
	
	// LoadOrStore handles the race of creating a new entry
	// We optimize for read-heavy/update-heavy mixed workloads
	val, loaded := r.agents.LoadOrStore(agentID, &AgentState{
		ID:            agentID,
		Name:          agentID, // Default to ID if name unknown, update later
		Status:        status,
		LastHeartbeat: now,
	})

	state := val.(*AgentState)

	// If it already existed, we just update the fields.
	// Note: In a pure struct usage with sync.Map, we'd need to re-store.
	// Since we stored a pointer, we can be atomic on the pointer, 
	// but strictly speaking we might want a localized Mutex on the AgentState 
	// if we had multiple writers to the *same* agent. 
	// For Heartbeats, usually only one source (the agent) writes, so simple assignment is fine 
	// or we can swap the struct. 
	
	if loaded {
		// Update existing
		state.Status = status
		state.LastHeartbeat = now
	}
}

// GetActiveAgents returns a list of agents seen in the last 30 seconds
func (r *Registry) GetActiveAgents() []*AgentState {
	active := []*AgentState{}
	threshold := time.Now().Add(-30 * time.Second)

	r.agents.Range(func(key, value interface{}) bool {
		state := value.(*AgentState)
		if state.LastHeartbeat.After(threshold) {
			// Return a copy or the pointer? Pointer is cheaper.
			active = append(active, state)
		}
		return true // continue iteration
	})

	return active
}

// ToProtoStatus converts internal status to Protobuf if needed
func ToProtoStatus(s AgentStatus) pb.AgentConfig {
	// Placeholder for mapping to proto enum if strictly required
	return pb.AgentConfig{} 
}

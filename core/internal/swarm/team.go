package swarm

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
)

type TeamType string

const (
	TeamTypeAction     TeamType = "action"
	TeamTypeExpression TeamType = "expression"
)

// TeamManifest defines the configuration for a Swarm Team.
type TeamManifest struct {
	ID          string                   `yaml:"id"`
	Name        string                   `yaml:"name"`
	Type        TeamType                 `yaml:"type"`
	Description string                   `yaml:"description"`
	// Members are the Agents (and their roles) that form this team
	Members []protocol.AgentManifest `yaml:"members"`
	// Inputs are the NATS subjects this team listens to (Triggers)
	Inputs []string `yaml:"inputs"`
	// Deliveries are the output channels
	Deliveries []string `yaml:"deliveries"`
}

// Team represents a running instance of a TeamManifest.
// It acts as a "Group Chat" container using an internal NATS subject space.
type Team struct {
	Manifest      *TeamManifest
	nc            *nats.Conn
	brain         *cognitive.Router
	toolExecutor  MCPToolExecutor
	ctx           context.Context
	cancel        context.CancelFunc
	mu            sync.Mutex
	sensorConfigs map[string]SensorConfig // agent.ID → config; nil = all cognitive
}

// NewTeam creates a new Team instance.
// toolExec may be nil if MCP tools are not available.
func NewTeam(manifest *TeamManifest, nc *nats.Conn, brain *cognitive.Router, toolExec MCPToolExecutor) *Team {
	ctx, cancel := context.WithCancel(context.Background())
	return &Team{
		Manifest:     manifest,
		nc:           nc,
		brain:        brain,
		toolExecutor: toolExec,
		ctx:          ctx,
		cancel:       cancel,
	}
}

// Start activates the Team's subscriptions.
func (t *Team) Start() error {
	log.Printf("Team [%s] (%s) Online.", t.Manifest.Name, t.Manifest.Type)

	// 1. Subscribe to defined Inputs (Triggers)
	for _, subject := range t.Manifest.Inputs {
		if _, err := t.nc.Subscribe(subject, t.handleTrigger); err != nil {
			log.Printf("Team [%s] Failed to subscribe to input [%s]: %v", t.Manifest.Name, subject, err)
		} else {
			log.Printf("Team [%s] Listening on [%s]", t.Manifest.Name, subject)
		}
	}

	// 2. Spawn Agents (sensor-aware: SensorAgent for poll-based, Agent for cognitive)
	for _, manifest := range t.Manifest.Members {
		if cfg, isSensor := t.sensorConfigs[manifest.ID]; isSensor {
			sensor := NewSensorAgent(t.ctx, manifest, cfg, t.Manifest.ID, t.nc)
			go sensor.Start()
		} else {
			agent := NewAgent(t.ctx, manifest, t.Manifest.ID, t.nc, t.brain, t.toolExecutor)
			go agent.Start()
		}
	}

	// 3. Subscribe to Internal Responses (to forward to Outputs)
	internalResponse := fmt.Sprintf(protocol.TopicTeamInternalRespond, t.Manifest.ID)
	t.nc.Subscribe(internalResponse, t.handleResponse)

	return nil
}

// handleTrigger receives an external signal and broadens it to the internal team bus.
func (t *Team) handleTrigger(msg *nats.Msg) {
	log.Printf("Team [%s] Triggered by [%s]", t.Manifest.Name, msg.Subject)

	// Forward to Internal Bus for Agents to react
	internalSubject := fmt.Sprintf(protocol.TopicTeamInternalTrigger, t.Manifest.ID)
	t.nc.Publish(internalSubject, msg.Data)
}

// handleResponse receives an internal signal and broadens it to the external team bus (Deliveries).
func (t *Team) handleResponse(msg *nats.Msg) {
	log.Printf("Team [%s] Response: %s", t.Manifest.Name, string(msg.Data))
	for _, subject := range t.Manifest.Deliveries {
		t.nc.Publish(subject, msg.Data)
	}
}

// SetSensorConfigs marks specific agents as sensor agents.
// Agents whose ID appears in the map will be spawned as SensorAgents
// (poll-based) instead of cognitive Agents (LLM-based).
func (t *Team) SetSensorConfigs(configs map[string]SensorConfig) {
	t.sensorConfigs = configs
}

// Stop shuts down the team.
func (t *Team) Stop() {
	t.cancel()
}

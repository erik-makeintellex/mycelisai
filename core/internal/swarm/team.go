package swarm

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

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
	// Schedule defines optional periodic auto-triggering (V1: interval-based)
	Schedule *protocol.ScheduleConfig `yaml:"schedule,omitempty"`
}

// Team represents a running instance of a TeamManifest.
// It acts as a "Group Chat" container using an internal NATS subject space.
type Team struct {
	Manifest      *TeamManifest
	nc            *nats.Conn
	brain         *cognitive.Router
	toolExecutor  MCPToolExecutor
	toolDescs     map[string]string        // tool name → description for agent prompt injection
	internalTools *InternalToolRegistry    // live system state + context builder
	ctx           context.Context
	cancel        context.CancelFunc
	mu            sync.Mutex
	sensorConfigs map[string]SensorConfig  // agent.ID → config; nil = all cognitive
	scheduler     *TeamScheduler           // Optional scheduled auto-trigger
	// V7 Event Spine: optional event emitter + run_id for tool audit trail.
	// Set by Soma.ActivateBlueprint BEFORE team.Start() so agents receive them.
	eventEmitter protocol.EventEmitter
	runID        string
	// V7 Conversation Log: optional full-fidelity turn logger.
	conversationLogger protocol.ConversationLogger
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
			// Inject tool descriptions for the agent's bound tools
			if len(manifest.Tools) > 0 && len(t.toolDescs) > 0 {
				agentDescs := make(map[string]string, len(manifest.Tools))
				for _, name := range manifest.Tools {
					if desc, ok := t.toolDescs[name]; ok {
						agentDescs[name] = desc
					}
				}
				agent.SetToolDescriptions(agentDescs)
			}
			// Inject internal tool registry (for runtime context) and team topology
			if t.internalTools != nil {
				agent.SetInternalTools(t.internalTools)
			}
			// V7: wire event emitter + run_id so agent can emit tool events
			if t.eventEmitter != nil && t.runID != "" {
				agent.SetEventEmitter(t.eventEmitter, t.runID)
			}
			// V7: wire conversation logger so agent can persist full transcripts
			if t.conversationLogger != nil {
				agent.SetConversationLogger(t.conversationLogger)
			}
			agent.SetTeamTopology(t.Manifest.Inputs, t.Manifest.Deliveries)
			go agent.Start()
		}
	}

	// 3. Subscribe to Internal Responses (to forward to Outputs)
	internalResponse := fmt.Sprintf(protocol.TopicTeamInternalRespond, t.Manifest.ID)
	t.nc.Subscribe(internalResponse, t.handleResponse)

	// 4. Start scheduler if configured (Phase 19: scheduled team execution)
	if t.Manifest.Schedule != nil && t.Manifest.Schedule.Interval != "" {
		interval, err := time.ParseDuration(t.Manifest.Schedule.Interval)
		if err != nil {
			log.Printf("Team [%s] invalid schedule interval %q: %v", t.Manifest.Name, t.Manifest.Schedule.Interval, err)
		} else if interval > 0 {
			// Phase 0 safety: enforce minimum 30s interval
			const minInterval = 30 * time.Second
			if interval < minInterval {
				log.Printf("WARN: Team [%s] schedule interval %s below minimum, clamping to %s", t.Manifest.Name, interval, minInterval)
				interval = minInterval
			}
			schedCtx, schedCancel := context.WithCancel(t.ctx)
			t.scheduler = &TeamScheduler{
				teamID:   t.Manifest.ID,
				interval: interval,
				nc:       t.nc,
				ctx:      schedCtx,
				cancel:   schedCancel,
			}
			go t.scheduler.Start()
		}
	}

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
func (t *Team) SetSensorConfigs(configs map[string]SensorConfig) {
	t.sensorConfigs = configs
}

// SetToolDescriptions provides tool name→description pairs for agent prompt injection.
func (t *Team) SetToolDescriptions(descs map[string]string) {
	t.toolDescs = descs
}

// SetInternalTools provides the internal tool registry for runtime context building.
func (t *Team) SetInternalTools(tools *InternalToolRegistry) {
	t.internalTools = tools
}

// SetEventEmitter wires the V7 event emitter + run_id into this team.
// Must be called before Start() so agents receive the emitter on creation.
func (t *Team) SetEventEmitter(emitter protocol.EventEmitter, runID string) {
	t.eventEmitter = emitter
	t.runID = runID
}

// SetConversationLogger wires the V7 conversation logger into this team.
// Must be called before Start() so agents receive the logger on creation.
func (t *Team) SetConversationLogger(logger protocol.ConversationLogger) {
	t.conversationLogger = logger
}

// Stop shuts down the team and its scheduler (if any).
func (t *Team) Stop() {
	if t.scheduler != nil {
		t.scheduler.Stop()
	}
	t.cancel()
}

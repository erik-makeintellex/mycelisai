package swarm

import (
	"context"
	"sync"

	"github.com/google/uuid"
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
	Provider    string                   `yaml:"provider,omitempty"`
	Members     []protocol.AgentManifest `yaml:"members"`
	Inputs      []string                 `yaml:"inputs"`
	Deliveries  []string                 `yaml:"deliveries"`
	Schedule    *protocol.ScheduleConfig `yaml:"schedule,omitempty"`
}

// Team represents a running instance of a TeamManifest.
type Team struct {
	Manifest           *TeamManifest
	nc                 *nats.Conn
	brain              *cognitive.Router
	toolExecutor       MCPToolExecutor
	toolDescs          map[string]string
	internalTools      *InternalToolRegistry
	ctx                context.Context
	cancel             context.CancelFunc
	mu                 sync.Mutex
	sensorConfigs      map[string]SensorConfig
	scheduler          *TeamScheduler
	eventEmitter       protocol.EventEmitter
	runID              string
	conversationLogger protocol.ConversationLogger
	compositeExec      *CompositeToolExecutor
	mcpServerNames     map[uuid.UUID]string
	mcpToolDescs       map[string]string
}

// NewTeam creates a new Team instance.
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

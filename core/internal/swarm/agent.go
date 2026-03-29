package swarm

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
)

// MCPToolExecutor resolves and invokes MCP tools by name.
type MCPToolExecutor interface {
	FindToolByName(ctx context.Context, name string) (serverID uuid.UUID, toolName string, err error)
	CallTool(ctx context.Context, serverID uuid.UUID, toolName string, args map[string]any) (string, error)
}

// Agent represents a single node in a Swarm Team.
type Agent struct {
	Manifest           protocol.AgentManifest
	TeamID             string
	TeamInputs         []string
	TeamDeliveries     []string
	nc                 *nats.Conn
	brain              *cognitive.Router
	toolExecutor       MCPToolExecutor
	toolDescs          map[string]string
	internalTools      *InternalToolRegistry
	ctx                context.Context
	cancel             context.CancelFunc
	eventEmitter       protocol.EventEmitter
	runID              string
	conversationLogger protocol.ConversationLogger
	sessionID          string
	turnIndex          int
	interjectionMu     sync.Mutex
	interjection       string
	interjectionSub    *nats.Subscription
}

// NewAgent creates a new Agent instance with lifecycle context.
func NewAgent(ctx context.Context, manifest protocol.AgentManifest, teamID string, nc *nats.Conn, brain *cognitive.Router, toolExec MCPToolExecutor) *Agent {
	agentCtx, cancel := context.WithCancel(ctx)
	return &Agent{
		Manifest:     manifest,
		TeamID:       teamID,
		nc:           nc,
		brain:        brain,
		toolExecutor: toolExec,
		ctx:          agentCtx,
		cancel:       cancel,
	}
}

func (a *Agent) SetToolDescriptions(descs map[string]string) { a.toolDescs = descs }

func (a *Agent) SetInternalTools(tools *InternalToolRegistry) { a.internalTools = tools }

func (a *Agent) SetEventEmitter(emitter protocol.EventEmitter, runID string) {
	a.eventEmitter = emitter
	a.runID = runID
}

func (a *Agent) SetConversationLogger(logger protocol.ConversationLogger) {
	a.conversationLogger = logger
}

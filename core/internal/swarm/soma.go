package swarm

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/internal/governance"
	"github.com/mycelis/core/internal/signal"
	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
)

// Soma is the executive cell body of the swarm.
type Soma struct {
	id                 string
	nc                 *nats.Conn
	guard              *governance.Guard
	axon               *Axon
	teams              map[string]*Team
	mu                 sync.RWMutex
	ctx                context.Context
	cancel             context.CancelFunc
	registry           *Registry
	brain              *cognitive.Router
	toolExecutor       MCPToolExecutor
	compositeExec      *CompositeToolExecutor
	internalTools      *InternalToolRegistry
	mcpServerNames     map[uuid.UUID]string
	mcpToolDescs       map[string]string
	runsManager        protocol.RunsManager
	eventEmitter       protocol.EventEmitter
	conversationLogger protocol.ConversationLogger
	providerPolicy     ProviderPolicy
}

// NewSoma creates a new Soma instance with composite tool support.
func NewSoma(nc *nats.Conn, guard *governance.Guard, registry *Registry, brain *cognitive.Router, stream *signal.StreamHandler, mcpExec MCPToolExecutor, internalTools *InternalToolRegistry) *Soma {
	var composite *CompositeToolExecutor
	if internalTools != nil || mcpExec != nil {
		composite = NewCompositeToolExecutor(internalTools, mcpExec)
	}

	ctx, cancel := context.WithCancel(context.Background())
	s := &Soma{
		id:            "soma-core",
		nc:            nc,
		guard:         guard,
		registry:      registry,
		brain:         brain,
		toolExecutor:  composite,
		compositeExec: composite,
		internalTools: internalTools,
		teams:         make(map[string]*Team),
		ctx:           ctx,
		cancel:        cancel,
	}
	if internalTools != nil {
		internalTools.SetSoma(s)
	}
	s.axon = NewAxon(nc, s, stream)
	return s
}

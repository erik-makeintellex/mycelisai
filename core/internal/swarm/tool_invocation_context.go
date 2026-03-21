package swarm

import (
	"context"
	"time"

	"github.com/mycelis/core/pkg/protocol"
)

type toolInvocationContextKey struct{}

// ToolInvocationContext captures execution provenance for internal tool calls.
// It is attached to context during agent execution and consumed by internal tools
// when publishing governed product signals.
type ToolInvocationContext struct {
	RunID         string
	TeamID        string
	AgentID       string
	SourceKind    protocol.SignalSourceKind
	SourceChannel string
	PayloadKind   protocol.SignalPayloadKind
	Timestamp     time.Time
}

// WithToolInvocationContext stores invocation metadata in context.
func WithToolInvocationContext(ctx context.Context, meta ToolInvocationContext) context.Context {
	return context.WithValue(ctx, toolInvocationContextKey{}, meta)
}

// ToolInvocationContextFromContext returns invocation metadata, if available.
func ToolInvocationContextFromContext(ctx context.Context) (ToolInvocationContext, bool) {
	meta, ok := ctx.Value(toolInvocationContextKey{}).(ToolInvocationContext)
	return meta, ok
}

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
	RunID   string
	TeamID  string
	AgentID string
	// AgentRole is the calling agent's manifest role. Core sets it from the
	// agent identity; tool arguments never supply it.
	AgentRole      string
	UserLabel      string
	OperatorID     string
	WorkspaceID    string
	OrganizationID string
	SourceKind     protocol.SignalSourceKind
	SourceChannel  string
	PayloadKind    protocol.SignalPayloadKind
	Timestamp      time.Time
	// PlanningOnly marks an invocation that is part of proposal generation and
	// must not produce any mutation side effects before confirmation.
	PlanningOnly bool
	// RuntimeOwned marks a mechanical proof step executed by Core rather than
	// a model-selected tool call. Internal tools may use it to return complete
	// evidence without expanding ordinary model-visible tool output.
	RuntimeOwned bool
	// Recall is the requesting user's saved-memory read scope (SRU). Core
	// sets it from the request identity; the zero value reads org-wide only.
	Recall RecallAccess
	// confirmedDispatch marks Core's execution of an operator-approved plan
	// (TPD). It is unexported, so no request, model, config or JSON path can
	// set it: only WithConfirmedDispatchToolContext does, and
	// WithToolInvocationContext always clears it. delegate_task keeps Core's
	// run/proof/work-item correlation only under this marker
	// (delegateAskAuthority).
	confirmedDispatch bool
}

// WithToolInvocationContext stores invocation metadata in context. It never
// carries the confirmed-dispatch marker, even when meta was read back from a
// confirmed-dispatch context.
func WithToolInvocationContext(ctx context.Context, meta ToolInvocationContext) context.Context {
	meta.confirmedDispatch = false
	return context.WithValue(ctx, toolInvocationContextKey{}, meta)
}

// WithRecallAccess replaces only the invocation's Recall scope and keeps
// every other field, including the Core-only confirmed-dispatch marker.
// Re-wrapping with WithToolInvocationContext would clear that marker, which
// would quietly drop a confirmed plan's delegate correlation (SRU x TPD merge).
// It cannot grant the marker: it only carries what ctx already holds.
func WithRecallAccess(ctx context.Context, recall RecallAccess) context.Context {
	meta, ok := ctx.Value(toolInvocationContextKey{}).(ToolInvocationContext)
	if !ok {
		return ctx
	}
	meta.Recall = recall
	return context.WithValue(ctx, toolInvocationContextKey{}, meta)
}

// WithConfirmedDispatchToolContext stores meta with the Core-only
// confirmed-dispatch marker. Only server.confirmedActionToolContext calls it,
// when Core executes the planned calls of a confirmed proposal.
func WithConfirmedDispatchToolContext(ctx context.Context, meta ToolInvocationContext) context.Context {
	meta.confirmedDispatch = true
	return context.WithValue(ctx, toolInvocationContextKey{}, meta)
}

// ToolInvocationContextFromContext returns invocation metadata, if available.
func ToolInvocationContextFromContext(ctx context.Context) (ToolInvocationContext, bool) {
	meta, ok := ctx.Value(toolInvocationContextKey{}).(ToolInvocationContext)
	return meta, ok
}

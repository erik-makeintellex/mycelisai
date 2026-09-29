package server

import (
	"context"
	"strings"
	"time"

	"github.com/mycelis/core/internal/swarm"
	"github.com/mycelis/core/pkg/protocol"
)

// confirmedActionToolContext is the tool context for Core executing the
// planned calls of a confirmed proposal. It carries the Core-only
// confirmed-dispatch marker (swarm.WithConfirmedDispatchToolContext, TPD);
// the source channel below is descriptive metadata, not the authority check.
func confirmedActionToolContext(ctx context.Context, auditUser, runID string, boundary *protocol.ConfigDocumentRequestBoundary) context.Context {
	actorID := strings.TrimSpace(auditUser)
	trustedBoundary := protocol.ConfigDocumentRequestBoundary{OperatorID: actorID}
	if boundary != nil {
		trustedBoundary = *boundary
		trustedBoundary.OperatorID = actorID
	}
	return swarm.WithConfirmedDispatchToolContext(ctx, swarm.ToolInvocationContext{
		SourceKind:     protocol.SourceKindWebAPI,
		SourceChannel:  "api.intent.confirm-action",
		PayloadKind:    protocol.PayloadKindCommand,
		Timestamp:      time.Now(),
		UserLabel:      actorID,
		OperatorID:     trustedBoundary.OperatorID,
		WorkspaceID:    strings.TrimSpace(trustedBoundary.WorkspaceID),
		OrganizationID: strings.TrimSpace(trustedBoundary.OrganizationID),
		AgentID:        actorID,
		RunID:          strings.TrimSpace(runID),
		PlanningOnly:   false,
	})
}

package server

import (
	"context"
	"strings"
	"time"

	"github.com/mycelis/core/internal/swarm"
	"github.com/mycelis/core/pkg/protocol"
)

func confirmedActionToolContext(ctx context.Context, auditUser, runID string, boundary *protocol.ConfigDocumentRequestBoundary) context.Context {
	actorID := strings.TrimSpace(auditUser)
	trustedBoundary := protocol.ConfigDocumentRequestBoundary{OperatorID: actorID}
	if boundary != nil {
		trustedBoundary = *boundary
		trustedBoundary.OperatorID = actorID
	}
	return swarm.WithToolInvocationContext(ctx, swarm.ToolInvocationContext{
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

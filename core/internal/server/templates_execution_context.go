package server

import (
	"context"
	"strings"
	"time"

	"github.com/mycelis/core/internal/swarm"
	"github.com/mycelis/core/pkg/protocol"
)

func confirmedActionToolContext(ctx context.Context, auditUser, runID string) context.Context {
	actorID := strings.TrimSpace(auditUser)
	return swarm.WithToolInvocationContext(ctx, swarm.ToolInvocationContext{
		SourceKind:    protocol.SourceKindWebAPI,
		SourceChannel: "api.intent.confirm-action",
		PayloadKind:   protocol.PayloadKindCommand,
		Timestamp:     time.Now(),
		UserLabel:     actorID,
		AgentID:       actorID,
		RunID:         strings.TrimSpace(runID),
		PlanningOnly:  false,
	})
}

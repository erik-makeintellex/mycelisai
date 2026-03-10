package swarm

import (
	"context"
	"fmt"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

func (r *InternalToolRegistry) wrapGovernedSignalPayload(
	ctx context.Context,
	sourceChannel string,
	teamID string,
	payloadKind protocol.SignalPayloadKind,
	raw []byte,
) ([]byte, error) {
	runID := ""
	agentID := ""
	if inv, ok := ToolInvocationContextFromContext(ctx); ok {
		if strings.TrimSpace(inv.TeamID) != "" && strings.TrimSpace(teamID) == "" {
			teamID = strings.TrimSpace(inv.TeamID)
		}
		runID = strings.TrimSpace(inv.RunID)
		agentID = strings.TrimSpace(inv.AgentID)
	}

	return protocol.WrapSignalPayloadWithMeta(
		protocol.SourceKindInternalTool,
		sourceChannel,
		payloadKind,
		runID,
		strings.TrimSpace(teamID),
		agentID,
		raw,
	)
}

func inferPayloadKindFromSubject(subject string) (protocol.SignalPayloadKind, bool) {
	trimmed := strings.TrimSpace(subject)
	switch {
	case strings.HasSuffix(trimmed, ".internal.command"):
		return protocol.PayloadKindCommand, true
	case strings.HasSuffix(trimmed, ".signal.status"):
		return protocol.PayloadKindStatus, true
	case strings.HasSuffix(trimmed, ".signal.result"):
		return protocol.PayloadKindResult, true
	case strings.HasSuffix(trimmed, ".telemetry"):
		return protocol.PayloadKindTelemetry, true
	case strings.HasPrefix(trimmed, fmt.Sprintf(protocol.TopicMissionEventsFmt, "")):
		return protocol.PayloadKindEvent, true
	default:
		return "", false
	}
}

func inferTeamIDFromSubject(subject string) string {
	parts := strings.Split(strings.TrimSpace(subject), ".")
	if len(parts) < 3 {
		return ""
	}
	if parts[0] != "swarm" || parts[1] != "team" {
		return ""
	}
	return strings.TrimSpace(parts[2])
}

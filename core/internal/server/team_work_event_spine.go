package server

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mycelis/core/pkg/protocol"
)

func (s *AdminServer) insertTeamWorkMissionEventDB(ctx context.Context, event *protocol.TeamStatusEvent) error {
	if strings.TrimSpace(event.RunID) == "" {
		return nil
	}
	db := s.getDB()
	if db == nil {
		return errors.New("database not available")
	}
	return s.insertTeamWorkMissionEventExec(ctx, db, event)
}

func (s *AdminServer) insertTeamWorkMissionEventExec(ctx context.Context, exec teamWorkSQLExecutor, event *protocol.TeamStatusEvent) error {
	if strings.TrimSpace(event.RunID) == "" {
		return nil
	}
	if exec == nil {
		return errors.New("database not available")
	}
	emittedAt := event.Timestamp
	if emittedAt.IsZero() {
		emittedAt = time.Now()
	}
	payload, _ := json.Marshal(teamWorkMissionEventPayload(event, emittedAt))
	_, err := exec.ExecContext(ctx, `
		INSERT INTO mission_events
			(id, run_id, tenant_id, event_type, severity, source_agent, source_team, payload, emitted_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		uuid.NewString(), nullableUUID(event.RunID), "default", string(protocol.EventTeamWorkStatus),
		string(teamWorkMissionEventSeverity(event.State)), "soma", event.TeamID, payload, emittedAt,
	)
	return err
}

func teamWorkMissionEventPayload(event *protocol.TeamStatusEvent, emittedAt time.Time) map[string]any {
	targetRef := event.TargetRef
	if targetRef == nil {
		targetRef = protocol.TargetRefForTeamStatusEvent(*event)
	}
	return map[string]any{
		"team_status_event_id":    event.EventID,
		"team_work_event_version": "v1",
		"team_id":                 event.TeamID,
		"work_item_id":            event.WorkItemID,
		"run_id":                  event.RunID,
		"intent_proof_id":         event.IntentProofID,
		"contract_id":             event.ContractID,
		"proof_id":                event.ProofID,
		"state":                   string(event.State),
		"headline":                event.Headline,
		"details":                 event.Details,
		"confidence_posture":      event.ConfidencePosture,
		"blocked_by":              event.BlockedBy,
		"next_action":             event.NextAction,
		"source_kind":             event.SourceKind,
		"source_channel":          event.SourceChannel,
		"payload_kind":            event.PayloadKind,
		"audit_refs":              event.AuditRefs,
		"target_ref":              targetRef,
		"timestamp":               emittedAt,
	}
}

func teamWorkMissionEventSeverity(state protocol.TeamWorkState) protocol.EventSeverity {
	switch state {
	case protocol.TeamWorkStateDegraded:
		return protocol.SeverityError
	case protocol.TeamWorkStateNeedsOperator, protocol.TeamWorkStatePaused:
		return protocol.SeverityWarn
	default:
		return protocol.SeverityInfo
	}
}

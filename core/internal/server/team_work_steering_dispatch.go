package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/dispatchoutbox"
	"github.com/mycelis/core/pkg/protocol"
)

const (
	teamWorkSteeringDispatchKind = "team_work_steering"
	teamWorkSteeringMaxAttempts  = 3
)

type teamWorkSteeringDispatchPayload struct {
	Action         protocol.TeamWorkAction `json:"action"`
	Summary        string                  `json:"summary"`
	ActorRef       string                  `json:"actor_ref"`
	IdempotencyKey string                  `json:"idempotency_key"`
}

func (s *AdminServer) stageTeamWorkSteeringTx(
	ctx context.Context,
	tx *sql.Tx,
	item protocol.TeamWorkItem,
	req teamWorkActionRequest,
) (string, error) {
	if s.DispatchOutbox == nil {
		return "", dispatchoutbox.ErrUnavailable
	}
	steeringID := strings.TrimSpace(req.IdempotencyKey)
	if steeringID == "" {
		steeringID = uuid.NewString()
	}
	key := "team-steer:" + item.WorkItemID + ":" + steeringID
	payload := teamWorkSteeringDispatchPayload{
		Action:         protocol.TeamWorkActionSteer,
		Summary:        strings.TrimSpace(req.Summary),
		ActorRef:       defaultString(req.ActorRef, "operator"),
		IdempotencyKey: key,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	_, err = s.DispatchOutbox.EnqueueTx(ctx, tx, dispatchoutbox.Item{
		ID:             uuid.NewString(),
		IdempotencyKey: key,
		DispatchKind:   teamWorkSteeringDispatchKind,
		RunID:          item.RunID,
		IntentProofID:  item.IntentProofID,
		ContractID:     item.ContractID,
		TeamID:         item.TeamID,
		WorkItemID:     item.WorkItemID,
		SourceKind:     defaultString(req.SourceKind, string(protocol.SourceKindWorkspaceUI)),
		SourceChannel:  defaultString(req.SourceChannel, "soma.active_work.steer"),
		PayloadKind:    string(protocol.PayloadKindCommand),
		Payload:        raw,
		Recovery:       json.RawMessage(`{"action":"retry_team_steering","operator_required":false}`),
	})
	return key, err
}

func (s *AdminServer) dispatchClaimedTeamWorkSteering(ctx context.Context, item *dispatchoutbox.Item) error {
	var payload teamWorkSteeringDispatchPayload
	if err := json.Unmarshal(item.Payload, &payload); err != nil {
		_ = s.DispatchOutbox.MarkFailed(ctx, item.ID, err)
		return fmt.Errorf("decode team steering %s: %w", item.ID, err)
	}
	if s.NC == nil || !s.NC.IsConnected() {
		return s.retryOrFailTeamWorkSteering(ctx, item, errors.New("team bus unavailable"))
	}
	command, err := json.Marshal(map[string]any{
		"goal":     "Continue the current work with this operator guidance: " + strings.TrimSpace(payload.Summary),
		"guidance": strings.TrimSpace(payload.Summary),
		"context": map[string]any{
			"action":          string(protocol.TeamWorkActionSteer),
			"work_item_id":    item.WorkItemID,
			"team_id":         item.TeamID,
			"run_id":          item.RunID,
			"idempotency_key": payload.IdempotencyKey,
			"actor_ref":       payload.ActorRef,
		},
	})
	if err != nil {
		return s.retryOrFailTeamWorkSteering(ctx, item, err)
	}
	wrapper, err := protocol.WrapSignalPayloadWithMeta(
		protocol.SourceKindWorkspaceUI,
		"soma.active_work.steer",
		protocol.PayloadKindCommand,
		item.RunID,
		item.TeamID,
		"",
		command,
	)
	if err != nil {
		return s.retryOrFailTeamWorkSteering(ctx, item, err)
	}
	subject := fmt.Sprintf(protocol.TopicTeamInternalCommand, item.TeamID)
	if err := s.NC.Publish(subject, wrapper); err != nil {
		return s.retryOrFailTeamWorkSteering(ctx, item, err)
	}
	if err := s.NC.FlushTimeout(2 * time.Second); err != nil {
		return s.retryOrFailTeamWorkSteering(ctx, item, err)
	}
	return s.DispatchOutbox.MarkCompleted(ctx, item.ID)
}

func (s *AdminServer) retryOrFailTeamWorkSteering(ctx context.Context, item *dispatchoutbox.Item, cause error) error {
	if item.AttemptCount < teamWorkSteeringMaxAttempts {
		return s.DispatchOutbox.MarkRetry(ctx, item.ID, cause, time.Duration(item.AttemptCount)*time.Second)
	}
	return s.DispatchOutbox.MarkFailed(ctx, item.ID, cause)
}

package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mycelis/core/pkg/protocol"
)

const (
	teamWorkAskSourceChannel   = "api.teams.work.ask"
	defaultTeamAskTimeout      = 15 * time.Second
	maxTeamAskTimeout          = 60 * time.Second
	teamAskFollowupTimeout     = 5 * time.Second
	asyncTeamAskRecoveryWindow = 15 * time.Minute
)

type teamWorkAskRequest struct {
	Message                string                   `json:"message"`
	Summary                string                   `json:"summary,omitempty"`
	ActorRef               string                   `json:"actor_ref,omitempty"`
	Ask                    *protocol.TeamAsk        `json:"ask,omitempty"`
	TimeoutSeconds         int                      `json:"timeout_seconds,omitempty"`
	ExpectedOutputs        []string                 `json:"expected_outputs,omitempty"`
	ExpectedProof          []string                 `json:"expected_proof,omitempty"`
	CapabilityRequirements []string                 `json:"capability_requirements,omitempty"`
	GovernancePosture      protocol.ApprovalPosture `json:"governance_posture,omitempty"`
	Payload                map[string]any           `json:"payload,omitempty"`
	Async                  bool                     `json:"async,omitempty"`
}

type teamWorkAskResult struct {
	WorkItem      protocol.TeamWorkItem    `json:"work_item"`
	Event         protocol.TeamStatusEvent `json:"event"`
	Reply         string                   `json:"reply,omitempty"`
	Subject       string                   `json:"subject"`
	Accepted      bool                     `json:"accepted,omitempty"`
	DispatchState string                   `json:"dispatch_state,omitempty"`
}

// HandleTeamWorkAsk submits one bounded request to a team and records either
// output or degradation as durable Active Work state.
func (s *AdminServer) HandleTeamWorkAsk(w http.ResponseWriter, r *http.Request) {
	teamID := strings.TrimSpace(r.PathValue("id"))
	if teamID == "" {
		respondAPIError(w, "team_id is required", http.StatusBadRequest)
		return
	}
	var req teamWorkAskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondAPIError(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}
	req.Message = strings.TrimSpace(req.Message)
	if req.Message == "" && (req.Ask == nil || req.Ask.IsZero()) {
		respondAPIError(w, "message or ask is required", http.StatusBadRequest)
		return
	}

	item := newTeamWorkAskItem(teamID, req)
	queued := teamWorkAskStatusEvent(item, protocol.TeamWorkStateQueued, "Team ask queued", "Soma queued a bounded ask on the team's response lane.", "pending_team_response", "Wait for a team reply or degradation proof.", nil)
	interaction := teamWorkAskInteraction(item, req, "ask", teamWorkAskSummary(req), nil)
	if err := s.persistTeamWorkItemWithLifecycle(r.Context(), &item, []protocol.TeamStatusEvent{queued}, interaction); err != nil {
		respondAPIError(w, "Failed to create team work ask: "+err.Error(), http.StatusInternalServerError)
		return
	}

	raw, err := teamWorkAskPayload(req)
	if err != nil {
		respondAPIError(w, "Failed to serialize team ask: "+err.Error(), http.StatusInternalServerError)
		return
	}
	timeout := boundedTeamAskTimeout(req.TimeoutSeconds)
	if req.Async {
		subject := fmt.Sprintf(protocol.TopicTeamInternalCommand, teamID)
		dispatchState, err := s.dispatchTeamWorkAsk(item, req, subject)
		if err != nil {
			followupCtx, followupCancel := teamWorkAskFollowupContext(r.Context())
			defer followupCancel()
			s.respondTeamWorkAskDegraded(w, followupCtx, item, subject, dispatchState, err.Error(), http.StatusAccepted)
			return
		}
		followupCtx, followupCancel := teamWorkAskFollowupContext(r.Context())
		defer followupCancel()
		dispatched, err := s.recordTeamWorkAskDispatched(followupCtx, &item, subject, dispatchState)
		if err != nil {
			respondAPIError(w, "Failed to record team ask dispatch: "+err.Error(), http.StatusInternalServerError)
			return
		}
		respondAPIJSON(w, http.StatusAccepted, protocol.NewAPISuccess(teamWorkAskResult{
			WorkItem:      item,
			Event:         dispatched,
			Subject:       subject,
			Accepted:      true,
			DispatchState: dispatchState,
		}))
		return
	}
	followupCtx, followupCancel := teamWorkAskFollowupContext(r.Context())
	defer followupCancel()
	subject := fmt.Sprintf(protocol.TopicTeamInternalTrigger, teamID)
	if s.NC == nil || !s.NC.IsConnected() {
		s.respondTeamWorkAskDegraded(w, followupCtx, item, subject, "nats_offline", "NATS connection offline; the team ask was recorded but not sent.", http.StatusAccepted)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()
	msg, err := s.NC.RequestWithContext(ctx, subject, raw)
	if err != nil {
		s.respondTeamWorkAskDegraded(w, followupCtx, item, subject, "team_response_timeout", "The team did not return a response within "+timeout.String()+": "+err.Error(), http.StatusAccepted)
		return
	}
	reply := string(msg.Data)
	if !teamWorkAskReplyReadable(reply) {
		s.respondTeamWorkAskDegraded(w, followupCtx, item, subject, "team_response_unreadable", "The team returned a response, but it was not suitable as operator-visible output.", http.StatusAccepted)
		return
	}
	s.respondTeamWorkAskOutput(w, followupCtx, item, subject, reply)
}

func teamWorkAskFollowupContext(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.WithoutCancel(parent), teamAskFollowupTimeout)
}

func newTeamWorkAskItem(teamID string, req teamWorkAskRequest) protocol.TeamWorkItem {
	return protocol.NormalizeTeamWorkItem(protocol.TeamWorkItem{
		WorkItemID:             uuid.NewString(),
		TeamID:                 teamID,
		Objective:              firstNonEmptyString(req.Summary, req.Message, protocol.SummarizeTeamAsk(req.AskValue()), "Bounded team ask"),
		Owner:                  "Soma",
		ExecutionShape:         protocol.TeamExecutionShapeDelegatedWork,
		State:                  protocol.TeamWorkStateQueued,
		ExpectedOutputs:        defaultStringSlice(req.ExpectedOutputs, "Team response or retained output"),
		ExpectedProof:          defaultStringSlice(req.ExpectedProof, "Team response event or degraded timeout proof"),
		CapabilityRequirements: req.CapabilityRequirements,
		GovernancePosture:      req.GovernancePosture,
	})
}

func (req teamWorkAskRequest) AskValue() protocol.TeamAsk {
	if req.Ask == nil {
		return protocol.TeamAsk{}
	}
	return req.Ask.Normalize()
}

func teamWorkAskPayload(req teamWorkAskRequest) ([]byte, error) {
	if req.Ask != nil && !req.Ask.IsZero() {
		ask := req.Ask.Normalize()
		if strings.TrimSpace(req.Message) != "" && strings.TrimSpace(ask.Message) == "" {
			ask.Message = req.Message
		}
		return json.Marshal(ask)
	}
	return []byte(req.Message), nil
}

func (s *AdminServer) dispatchTeamWorkAsk(item protocol.TeamWorkItem, req teamWorkAskRequest, subject string) (string, error) {
	if s.NC == nil || !s.NC.IsConnected() {
		return "nats_offline", fmt.Errorf("NATS connection offline; the team ask was recorded but not sent.")
	}
	payload, err := teamWorkAskCommandEnvelope(item, req)
	if err != nil {
		return "team_ask_payload_invalid", fmt.Errorf("team ask command payload could not be prepared: %w", err)
	}
	if err := s.NC.Publish(subject, payload); err != nil {
		return "team_ask_publish_failed", fmt.Errorf("team ask command could not be published: %w", err)
	}
	return "published", nil
}

func teamWorkAskCommandEnvelope(item protocol.TeamWorkItem, req teamWorkAskRequest) ([]byte, error) {
	ask := req.AskValue()
	if ask.IsZero() {
		ask = protocol.TeamAsk{Message: req.Message}
	}
	ask = ask.Normalize()
	ask.Context = map[string]any{
		"work_item_id":             item.WorkItemID,
		"team_id":                  item.TeamID,
		"details":                  req.Payload,
		"expected_outputs":         item.ExpectedOutputs,
		"expected_proof":           item.ExpectedProof,
		"capability_requirements":  item.CapabilityRequirements,
		"governance_posture":       item.GovernancePosture,
		"source_active_work_state": item.State,
		"source_channel":           teamWorkAskSourceChannel,
	}
	raw, err := json.Marshal(ask)
	if err != nil {
		return nil, err
	}
	return protocol.WrapSignalPayload(
		protocol.SourceKindWebAPI,
		teamWorkAskSourceChannel,
		protocol.PayloadKindCommand,
		item.TeamID,
		raw,
	)
}

func boundedTeamAskTimeout(seconds int) time.Duration {
	if seconds <= 0 {
		return defaultTeamAskTimeout
	}
	timeout := time.Duration(seconds) * time.Second
	if timeout > maxTeamAskTimeout {
		return maxTeamAskTimeout
	}
	return timeout
}

func teamWorkAskSummary(req teamWorkAskRequest) string {
	return firstNonEmptyString(req.Summary, protocol.SummarizeTeamAsk(req.AskValue()), req.Message, "Bounded team ask")
}

func teamWorkAskStatusEvent(item protocol.TeamWorkItem, state protocol.TeamWorkState, headline, details, confidence, next string, blockedBy []string) protocol.TeamStatusEvent {
	return protocol.TeamStatusEvent{
		EventID:           uuid.NewString(),
		TeamID:            item.TeamID,
		WorkItemID:        item.WorkItemID,
		State:             state,
		Headline:          headline,
		Details:           details,
		ConfidencePosture: confidence,
		BlockedBy:         blockedBy,
		NextAction:        next,
		SourceKind:        string(protocol.SourceKindWebAPI),
		SourceChannel:     teamWorkAskSourceChannel,
		PayloadKind:       string(protocol.PayloadKindStatus),
		Version:           "v1",
	}
}

func teamWorkAskInteraction(item protocol.TeamWorkItem, req teamWorkAskRequest, verb, summary string, payload map[string]any) protocol.TeamInteraction {
	if payload == nil {
		payload = map[string]any{"message": req.Message, "ask": req.AskValue(), "details": req.Payload}
	}
	return protocol.NormalizeTeamInteraction(protocol.TeamInteraction{
		InteractionID: uuid.NewString(),
		TeamID:        item.TeamID,
		WorkItemID:    item.WorkItemID,
		SourceKind:    string(protocol.SourceKindWebAPI),
		SourceChannel: teamWorkAskSourceChannel,
		ActorRef:      defaultString(req.ActorRef, "Soma"),
		Verb:          verb,
		Summary:       summary,
		PayloadKind:   string(protocol.PayloadKindCommand),
		Payload:       payload,
		Version:       "v1",
	})
}

func (s *AdminServer) respondTeamWorkAskDegraded(w http.ResponseWriter, ctx context.Context, item protocol.TeamWorkItem, subject, degradation, details string, status int) {
	event, err := s.recordTeamWorkAskDegraded(ctx, &item, subject, degradation, details)
	if err != nil {
		respondAPIError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respondAPIJSON(w, status, protocol.NewAPISuccess(teamWorkAskResult{WorkItem: item, Event: event, Subject: subject}))
}

func (s *AdminServer) respondTeamWorkAskOutput(w http.ResponseWriter, ctx context.Context, item protocol.TeamWorkItem, subject, reply string) {
	event, err := s.recordTeamWorkAskOutput(ctx, &item, subject, reply)
	if err != nil {
		respondAPIError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(teamWorkAskResult{WorkItem: item, Event: event, Reply: reply, Subject: subject}))
}

func defaultStringSlice(items []string, fallback string) []string {
	normalized := normalizeStringSlice(items)
	if len(normalized) > 0 {
		return normalized
	}
	return []string{fallback}
}

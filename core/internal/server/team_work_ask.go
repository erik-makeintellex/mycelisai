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
	teamWorkAskSourceChannel = "api.teams.work.ask"
	defaultTeamAskTimeout    = 15 * time.Second
	maxTeamAskTimeout        = 60 * time.Second
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
}

type teamWorkAskResult struct {
	WorkItem protocol.TeamWorkItem    `json:"work_item"`
	Event    protocol.TeamStatusEvent `json:"event"`
	Reply    string                   `json:"reply,omitempty"`
	Subject  string                   `json:"subject"`
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

	subject := fmt.Sprintf(protocol.TopicTeamInternalTrigger, teamID)
	raw, err := teamWorkAskPayload(req)
	if err != nil {
		respondAPIError(w, "Failed to serialize team ask: "+err.Error(), http.StatusInternalServerError)
		return
	}
	timeout := boundedTeamAskTimeout(req.TimeoutSeconds)
	if s.NC == nil || !s.NC.IsConnected() {
		s.respondTeamWorkAskDegraded(w, r.Context(), item, subject, "nats_offline", "NATS connection offline; the team ask was recorded but not sent.", http.StatusAccepted)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()
	msg, err := s.NC.RequestWithContext(ctx, subject, raw)
	if err != nil {
		s.respondTeamWorkAskDegraded(w, r.Context(), item, subject, "team_response_timeout", "The team did not return a response within "+timeout.String()+": "+err.Error(), http.StatusAccepted)
		return
	}
	s.respondTeamWorkAskOutput(w, r.Context(), item, subject, string(msg.Data))
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
	item.State = protocol.TeamWorkStateDegraded
	item.NeedsOperator = true
	item.DegradationState = degradation
	item.RecoveryOptions = []string{"Recover the work item after NATS/team availability is restored.", "Add steering guidance before retrying.", "Archive if the work is no longer needed."}
	event := teamWorkAskStatusEvent(item, protocol.TeamWorkStateDegraded, "Team ask degraded", details, "operator_attention", "Recover or steer this work item before retrying.", []string{degradation})
	if err := s.insertTeamStatusEventDB(ctx, &event); err != nil {
		respondAPIError(w, "Failed to record degraded team ask: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err := s.updateTeamWorkItemLastEventDB(ctx, &item, event); err != nil {
		respondAPIError(w, "Failed to update degraded team ask: "+err.Error(), http.StatusInternalServerError)
		return
	}
	interaction := protocol.NormalizeTeamInteraction(protocol.TeamInteraction{
		InteractionID: uuid.NewString(),
		TeamID:        item.TeamID,
		WorkItemID:    item.WorkItemID,
		SourceKind:    string(protocol.SourceKindWebAPI),
		SourceChannel: teamWorkAskSourceChannel,
		ActorRef:      "Soma",
		Verb:          "degraded",
		Summary:       details,
		PayloadKind:   string(protocol.PayloadKindError),
		Payload:       map[string]any{"subject": subject, "degradation_state": degradation},
		Version:       "v1",
	})
	if err := s.insertTeamInteractionDB(ctx, &interaction); err != nil {
		respondAPIError(w, "Failed to record degraded team interaction: "+err.Error(), http.StatusInternalServerError)
		return
	}
	respondAPIJSON(w, status, protocol.NewAPISuccess(teamWorkAskResult{WorkItem: item, Event: event, Subject: subject}))
}

func (s *AdminServer) respondTeamWorkAskOutput(w http.ResponseWriter, ctx context.Context, item protocol.TeamWorkItem, subject, reply string) {
	item.State = protocol.TeamWorkStateOutputReady
	item.NeedsOperator = false
	item.DegradationState = ""
	event := teamWorkAskStatusEvent(item, protocol.TeamWorkStateOutputReady, "Team response ready", "The team returned a bounded response for this ask.", "team_response", "Review the response and decide whether to retain, steer, or ask for follow-up.", nil)
	if err := s.insertTeamStatusEventDB(ctx, &event); err != nil {
		respondAPIError(w, "Failed to record team response: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err := s.updateTeamWorkItemLastEventDB(ctx, &item, event); err != nil {
		respondAPIError(w, "Failed to update team response: "+err.Error(), http.StatusInternalServerError)
		return
	}
	interaction := protocol.NormalizeTeamInteraction(protocol.TeamInteraction{
		InteractionID: uuid.NewString(),
		TeamID:        item.TeamID,
		WorkItemID:    item.WorkItemID,
		SourceKind:    string(protocol.SourceKindWebAPI),
		SourceChannel: teamWorkAskSourceChannel,
		ActorRef:      item.TeamID,
		Verb:          "response",
		Summary:       firstNonEmptyString(reply, "Team returned an empty response."),
		PayloadKind:   string(protocol.PayloadKindResult),
		Payload:       map[string]any{"subject": subject, "reply": reply},
		Version:       "v1",
	})
	if err := s.insertTeamInteractionDB(ctx, &interaction); err != nil {
		respondAPIError(w, "Failed to record team response interaction: "+err.Error(), http.StatusInternalServerError)
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

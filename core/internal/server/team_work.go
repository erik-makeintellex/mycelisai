package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

type teamWorkActionRequest struct {
	Action        protocol.TeamWorkAction `json:"action"`
	Summary       string                  `json:"summary,omitempty"`
	ActorRef      string                  `json:"actor_ref,omitempty"`
	SourceKind    string                  `json:"source_kind,omitempty"`
	SourceChannel string                  `json:"source_channel,omitempty"`
	PayloadKind   string                  `json:"payload_kind,omitempty"`
	Payload       map[string]any          `json:"payload,omitempty"`
	AuditRefs     []string                `json:"audit_refs,omitempty"`
}

// HandleListTeamWork returns durable work items for one runtime or persisted team.
// GET /api/v1/teams/{id}/work
func (s *AdminServer) HandleListTeamWork(w http.ResponseWriter, r *http.Request) {
	teamID := strings.TrimSpace(r.PathValue("id"))
	if teamID == "" {
		respondAPIError(w, "team_id is required", http.StatusBadRequest)
		return
	}
	items, err := s.listTeamWorkItemsDB(r.Context(), teamID, parseLimit(r.URL.Query().Get("limit"), 20))
	if err != nil {
		respondAPIError(w, "Failed to list team work: "+err.Error(), http.StatusServiceUnavailable)
		return
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(items))
}

// HandleCreateTeamWork records a durable team-work item without starting runtime execution.
// POST /api/v1/teams/{id}/work
func (s *AdminServer) HandleCreateTeamWork(w http.ResponseWriter, r *http.Request) {
	teamID := strings.TrimSpace(r.PathValue("id"))
	if teamID == "" {
		respondAPIError(w, "team_id is required", http.StatusBadRequest)
		return
	}
	var req protocol.TeamWorkItem
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondAPIError(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}
	req.TeamID = teamID
	item := protocol.NormalizeTeamWorkItem(req)
	if err := protocol.ValidateTeamWorkItem(item); err != nil {
		respondAPIError(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := s.insertTeamWorkItemDB(r.Context(), &item); err != nil {
		respondAPIError(w, "Failed to create team work: "+err.Error(), http.StatusInternalServerError)
		return
	}
	respondAPIJSON(w, http.StatusCreated, protocol.NewAPISuccess(item))
}

// HandleListTeamInteractions returns the durable exchange history for one work item.
// GET /api/v1/teams/{id}/work/{workItemId}/interactions
func (s *AdminServer) HandleListTeamInteractions(w http.ResponseWriter, r *http.Request) {
	teamID, workItemID := teamWorkPathIDs(r)
	if teamID == "" || workItemID == "" {
		respondAPIError(w, "team_id and work_item_id are required", http.StatusBadRequest)
		return
	}
	if err := validateOptionalUUID("work_item_id", workItemID); err != nil {
		respondAPIError(w, err.Error(), http.StatusBadRequest)
		return
	}
	items, err := s.listTeamInteractionsDB(r.Context(), teamID, workItemID, parseLimit(r.URL.Query().Get("limit"), 50))
	if err != nil {
		respondAPIError(w, "Failed to list team interactions: "+err.Error(), http.StatusServiceUnavailable)
		return
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(items))
}

// HandleListTeamStatusEvents returns the durable operator-readable status timeline for one work item.
// GET /api/v1/teams/{id}/work/{workItemId}/status-events
func (s *AdminServer) HandleListTeamStatusEvents(w http.ResponseWriter, r *http.Request) {
	teamID, workItemID := teamWorkPathIDs(r)
	if teamID == "" || workItemID == "" {
		respondAPIError(w, "team_id and work_item_id are required", http.StatusBadRequest)
		return
	}
	if err := validateOptionalUUID("work_item_id", workItemID); err != nil {
		respondAPIError(w, err.Error(), http.StatusBadRequest)
		return
	}
	items, err := s.listTeamStatusEventsDB(r.Context(), teamID, workItemID, parseLimit(r.URL.Query().Get("limit"), 50))
	if err != nil {
		respondAPIError(w, "Failed to list team status events: "+err.Error(), http.StatusServiceUnavailable)
		return
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(items))
}

// HandleCreateTeamInteraction appends a durable Soma/Council/operator/team-lead exchange.
// POST /api/v1/teams/{id}/work/{workItemId}/interactions
func (s *AdminServer) HandleCreateTeamInteraction(w http.ResponseWriter, r *http.Request) {
	teamID, workItemID := teamWorkPathIDs(r)
	if teamID == "" || workItemID == "" {
		respondAPIError(w, "team_id and work_item_id are required", http.StatusBadRequest)
		return
	}
	var req protocol.TeamInteraction
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondAPIError(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}
	req.TeamID = teamID
	req.WorkItemID = workItemID
	item := protocol.NormalizeTeamInteraction(req)
	if err := protocol.ValidateTeamInteraction(item); err != nil {
		respondAPIError(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := s.insertTeamInteractionDB(r.Context(), &item); err != nil {
		respondAPIError(w, "Failed to create team interaction: "+err.Error(), http.StatusInternalServerError)
		return
	}
	respondAPIJSON(w, http.StatusCreated, protocol.NewAPISuccess(item))
}

// HandleTeamWorkAction applies a durable operator control to an existing work item.
// POST /api/v1/teams/{id}/work/{workItemId}/actions
func (s *AdminServer) HandleTeamWorkAction(w http.ResponseWriter, r *http.Request) {
	teamID, workItemID := teamWorkPathIDs(r)
	if teamID == "" || workItemID == "" {
		respondAPIError(w, "team_id and work_item_id are required", http.StatusBadRequest)
		return
	}
	if err := validateOptionalUUID("work_item_id", workItemID); err != nil {
		respondAPIError(w, err.Error(), http.StatusBadRequest)
		return
	}
	var req teamWorkActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondAPIError(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}
	action := protocol.NormalizeTeamWorkAction(req.Action)
	if action == "" {
		respondAPIError(w, "action is required", http.StatusBadRequest)
		return
	}

	item, err := s.getTeamWorkItemDB(r.Context(), teamID, workItemID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondAPIError(w, "team work item not found", http.StatusNotFound)
			return
		}
		respondAPIError(w, "Failed to load team work: "+err.Error(), http.StatusServiceUnavailable)
		return
	}
	targetState, err := protocol.ApplyTeamWorkAction(item, action)
	if err != nil {
		respondAPIError(w, err.Error(), http.StatusBadRequest)
		return
	}

	item.State = targetState
	if targetState == protocol.TeamWorkStateQueued || targetState == protocol.TeamWorkStateRunning {
		item.NeedsOperator = false
		item.DegradationState = ""
	}
	if targetState == protocol.TeamWorkStatePaused || targetState == protocol.TeamWorkStateArchived {
		item.NeedsOperator = false
	}
	if err := protocol.ValidateTeamWorkItem(item); err != nil {
		respondAPIError(w, err.Error(), http.StatusBadRequest)
		return
	}

	event := teamWorkActionStatusEvent(item, req, action)
	if err := s.insertTeamStatusEventDB(r.Context(), &event); err != nil {
		respondAPIError(w, "Failed to record team work status: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err := s.updateTeamWorkItemLastEventDB(r.Context(), &item, event); err != nil {
		respondAPIError(w, "Failed to update team work: "+err.Error(), http.StatusInternalServerError)
		return
	}
	interaction := teamWorkActionInteraction(item, req, action)
	if err := s.insertTeamInteractionDB(r.Context(), &interaction); err != nil {
		respondAPIError(w, "Failed to record team work interaction: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if !event.Timestamp.IsZero() {
		item.UpdatedAt = event.Timestamp
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(item))
}

func teamWorkPathIDs(r *http.Request) (string, string) {
	return strings.TrimSpace(r.PathValue("id")), strings.TrimSpace(r.PathValue("workItemId"))
}

func teamWorkActionStatusEvent(item protocol.TeamWorkItem, req teamWorkActionRequest, action protocol.TeamWorkAction) protocol.TeamStatusEvent {
	headline, details, nextAction := teamWorkActionCopy(action)
	if summary := strings.TrimSpace(req.Summary); summary != "" {
		details = summary
	}
	return protocol.TeamStatusEvent{
		TeamID:            item.TeamID,
		WorkItemID:        item.WorkItemID,
		RunID:             item.RunID,
		IntentProofID:     item.IntentProofID,
		ContractID:        item.ContractID,
		ProofID:           item.ProofID,
		State:             item.State,
		Headline:          headline,
		Details:           details,
		ConfidencePosture: "operator_recorded",
		NextAction:        nextAction,
		SourceKind:        defaultString(req.SourceKind, "workspace_ui"),
		SourceChannel:     defaultString(req.SourceChannel, "teams.active_work"),
		PayloadKind:       defaultString(req.PayloadKind, "team_work_action"),
		AuditRefs:         mergeStrings(item.AuditRefs, req.AuditRefs),
		Version:           "v1",
	}
}

func teamWorkActionInteraction(item protocol.TeamWorkItem, req teamWorkActionRequest, action protocol.TeamWorkAction) protocol.TeamInteraction {
	summary := strings.TrimSpace(req.Summary)
	if summary == "" {
		headline, _, _ := teamWorkActionCopy(action)
		summary = headline
	}
	return protocol.NormalizeTeamInteraction(protocol.TeamInteraction{
		TeamID:        item.TeamID,
		WorkItemID:    item.WorkItemID,
		RunID:         item.RunID,
		IntentProofID: item.IntentProofID,
		ContractID:    item.ContractID,
		ProofID:       item.ProofID,
		SourceKind:    defaultString(req.SourceKind, "workspace_ui"),
		SourceChannel: defaultString(req.SourceChannel, "teams.active_work"),
		ActorRef:      defaultString(req.ActorRef, "operator"),
		Verb:          string(action),
		Summary:       summary,
		PayloadKind:   defaultString(req.PayloadKind, "team_work_action"),
		Payload:       req.Payload,
		AuditRefs:     mergeStrings(item.AuditRefs, req.AuditRefs),
		Version:       "v1",
	})
}

func teamWorkActionCopy(action protocol.TeamWorkAction) (string, string, string) {
	switch action {
	case protocol.TeamWorkActionStartWork:
		return "Team work started", "The operator moved this durable work item into active execution.", "Watch for team status, retained output, or proof."
	case protocol.TeamWorkActionPause:
		return "Team work paused", "The operator paused this durable work item.", "Resume or archive this work when ready."
	case protocol.TeamWorkActionResume:
		return "Resume requested", "The operator returned this durable work item to the queue.", "Wait for the team to continue or add steering guidance."
	case protocol.TeamWorkActionArchive:
		return "Team work archived", "The operator archived this durable work item. Retained outputs and proof remain inspectable.", "Review retained proof or start a new work item."
	default:
		return "Team work updated", "The operator updated this durable work item.", "Review the latest state."
	}
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}

func mergeStrings(left, right []string) []string {
	seen := map[string]struct{}{}
	merged := make([]string, 0, len(left)+len(right))
	for _, value := range append(left, right...) {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		merged = append(merged, trimmed)
	}
	return merged
}

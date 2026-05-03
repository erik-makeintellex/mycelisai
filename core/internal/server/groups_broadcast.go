package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/mycelis/core/pkg/protocol"
)

// POST /api/v1/groups/{id}/broadcast
// Publishes in parallel to the group collab channel and each team's internal command lane.
func (s *AdminServer) HandleGroupBroadcast(w http.ResponseWriter, r *http.Request) {
	identity, ok := requireRootAdminScope(w, r, "groups:broadcast")
	if !ok {
		return
	}
	if s.NC == nil || !s.NC.IsConnected() {
		respondAPIError(w, "NATS connection offline", http.StatusServiceUnavailable)
		return
	}

	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		respondAPIError(w, "Missing group ID", http.StatusBadRequest)
		return
	}

	group, err := s.getGroupDB(r.Context(), id)
	if err != nil {
		respondAPIError(w, "Failed to load group: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if group == nil {
		respondAPIError(w, "Group not found", http.StatusNotFound)
		return
	}
	if group.Status != groupStatusActive {
		respondAPIError(w, "Group is not active", http.StatusConflict)
		return
	}

	var req groupBroadcastRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondAPIError(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}
	msg := strings.TrimSpace(req.Message)
	if msg == "" {
		respondAPIError(w, "message is required", http.StatusBadRequest)
		return
	}

	subjects := groupBroadcastSubjects(group)
	raw, err := json.Marshal(groupBroadcastPayload(group, identity.UserID, msg))
	if err != nil {
		respondAPIError(w, "Failed to serialize payload", http.StatusInternalServerError)
		return
	}

	if err := s.publishGroupBroadcast(subjects, raw); err != nil {
		if s.GroupBus != nil {
			s.GroupBus.RecordFailure(err)
		}
		respondAPIError(w, "Failed to publish group broadcast: "+err.Error(), http.StatusBadGateway)
		return
	}

	if s.GroupBus != nil {
		s.GroupBus.RecordSuccess(group.ID, identity.UserID, msg, subjects)
	}

	auditEventID, _ := s.createAuditEvent(
		protocol.TemplateChatToProposal,
		"groups.broadcast",
		"Broadcast to collaboration group",
		map[string]any{
			"group_id":   group.ID,
			"subjects":   subjects,
			"message":    msg,
			"actor_id":   identity.UserID,
			"team_ids":   group.TeamIDs,
			"team_count": len(group.TeamIDs),
		},
	)

	respondAPIJSON(w, http.StatusAccepted, protocol.NewAPISuccess(map[string]any{
		"group_id":       group.ID,
		"status":         "queued",
		"subjects":       subjects,
		"team_count":     len(group.TeamIDs),
		"audit_event_id": auditEventID,
	}))
}

func groupBroadcastSubjects(group *CollaborationGroup) []string {
	subjects := []string{fmt.Sprintf(protocol.TopicGroupCollabFmt, group.ID)}
	for _, teamID := range group.TeamIDs {
		teamID = strings.TrimSpace(teamID)
		if teamID == "" {
			continue
		}
		subjects = append(subjects, fmt.Sprintf(protocol.TopicTeamInternalCommand, teamID))
	}
	return subjects
}

func groupBroadcastPayload(group *CollaborationGroup, actorID, message string) map[string]any {
	return map[string]any{
		"group_id":   group.ID,
		"tenant_id":  group.TenantID,
		"message":    message,
		"actor_id":   actorID,
		"emitted_at": time.Now().UTC().Format(time.RFC3339Nano),
	}
}

func (s *AdminServer) publishGroupBroadcast(subjects []string, raw []byte) error {
	var wg sync.WaitGroup
	errCh := make(chan error, len(subjects))
	for _, subject := range subjects {
		subject := subject
		wg.Add(1)
		go func() {
			defer wg.Done()
			if pubErr := s.NC.Publish(subject, raw); pubErr != nil {
				errCh <- fmt.Errorf("%s: %w", subject, pubErr)
			}
		}()
	}
	wg.Wait()
	close(errCh)

	for pubErr := range errCh {
		return pubErr
	}
	return nil
}

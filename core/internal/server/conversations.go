package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/mycelis/core/pkg/protocol"
)

// GET /api/v1/runs/{id}/conversation
// Returns all conversation turns for a run, ordered chronologically.
// Optional query param: ?agent=X to filter by agent_id.
func (s *AdminServer) HandleGetRunConversation(w http.ResponseWriter, r *http.Request) {
	runID := r.PathValue("id")
	if runID == "" {
		respondAPIError(w, "Missing run ID", http.StatusBadRequest)
		return
	}

	if s.Conversations == nil {
		respondAPIError(w, "Conversation store not available", http.StatusServiceUnavailable)
		return
	}

	agentFilter := r.URL.Query().Get("agent")
	turns, err := s.Conversations.GetRunConversation(r.Context(), runID, agentFilter)
	if err != nil {
		log.Printf("[conversations] GetRunConversation failed: %v", err)
		respondAPIError(w, "Failed to fetch conversation", http.StatusInternalServerError)
		return
	}

	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(map[string]interface{}{
		"run_id": runID,
		"turns":  turns,
	}))
}

// GET /api/v1/conversations/{session_id}
// Returns all turns for a specific session (standing-team chats without run_id).
func (s *AdminServer) HandleGetSessionConversation(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("session_id")
	if sessionID == "" {
		respondAPIError(w, "Missing session ID", http.StatusBadRequest)
		return
	}

	if s.Conversations == nil {
		respondAPIError(w, "Conversation store not available", http.StatusServiceUnavailable)
		return
	}

	turns, err := s.Conversations.GetSessionTurns(r.Context(), sessionID)
	if err != nil {
		log.Printf("[conversations] GetSessionTurns failed: %v", err)
		respondAPIError(w, "Failed to fetch session conversation", http.StatusInternalServerError)
		return
	}

	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(map[string]interface{}{
		"session_id": sessionID,
		"turns":      turns,
	}))
}

// POST /api/v1/runs/{id}/interject
// Publishes a user interjection to the active agents in a run.
// The message is buffered by agents and injected between ReAct iterations.
func (s *AdminServer) HandleRunInterject(w http.ResponseWriter, r *http.Request) {
	runID := r.PathValue("id")
	if runID == "" {
		respondAPIError(w, "Missing run ID", http.StatusBadRequest)
		return
	}

	if s.NC == nil {
		respondAPIError(w, "NATS offline — cannot deliver interjection", http.StatusServiceUnavailable)
		return
	}

	var req struct {
		Message string `json:"message"`
		AgentID string `json:"agent_id,omitempty"` // optional: target a specific agent
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondAPIError(w, "Bad JSON", http.StatusBadRequest)
		return
	}
	if req.Message == "" {
		respondAPIError(w, "Message is required", http.StatusBadRequest)
		return
	}

	// If a specific agent_id is provided, publish to that agent only.
	// Otherwise, broadcast to all agents in the run's teams.
	published := 0
	if req.AgentID != "" {
		subject := fmt.Sprintf(protocol.TopicAgentInterjectionFmt, req.AgentID)
		if err := s.NC.Publish(subject, []byte(req.Message)); err != nil {
			log.Printf("[interject] publish to %s failed: %v", req.AgentID, err)
			respondAPIError(w, "Failed to deliver interjection", http.StatusInternalServerError)
			return
		}
		published = 1
	} else {
		// Broadcast to all council/standing team agents
		if s.Soma != nil {
			for _, tm := range s.Soma.ListTeams() {
				for _, member := range tm.Members {
					subject := fmt.Sprintf(protocol.TopicAgentInterjectionFmt, member.ID)
					if err := s.NC.Publish(subject, []byte(req.Message)); err != nil {
						log.Printf("[interject] publish to %s failed: %v", member.ID, err)
						continue
					}
					published++
				}
			}
		}
	}
	s.NC.Flush()

	// Also log the interjection as a conversation turn if store is available
	if s.Conversations != nil {
		go func() {
			agentID := req.AgentID
			if agentID == "" {
				agentID = "operator"
			}
			s.Conversations.LogTurn(r.Context(), protocol.ConversationTurnData{ //nolint:errcheck
				RunID:    runID,
				TenantID: "default",
				AgentID:  agentID,
				Role:     "interjection",
				Content:  req.Message,
			})
		}()
	}

	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(map[string]interface{}{
		"run_id":         runID,
		"agents_reached": published,
		"message":        "Interjection delivered",
	}))
}

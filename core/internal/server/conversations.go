package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/mycelis/core/internal/conversations"
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
	if len(turns) == 0 && s.Events != nil {
		turns = s.synthesizeRunConversationFromEvents(r.Context(), runID, agentFilter)
	}

	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(map[string]interface{}{
		"run_id": runID,
		"turns":  turns,
	}))
}

func (s *AdminServer) logConfirmedRunConversationTurns(ctx context.Context, runID, proofID string, scope *protocol.ScopeValidation, finalSummary string) {
	if s.Conversations == nil || strings.TrimSpace(runID) == "" || scope == nil {
		return
	}
	teamID, teamName := plannedTeamIdentity(scope.PlannedToolCalls)
	if teamID == "" {
		teamID = "admin-core"
	}
	turnIdx := 0
	_, _ = s.Conversations.LogTurn(ctx, protocol.ConversationTurnData{
		RunID: runID, SessionID: runID, TenantID: "default", AgentID: "soma", TeamID: teamID,
		TurnIndex: turnIdx, Role: "system",
		Content: fmt.Sprintf("Approved proposal (proof %s) started execution across %d planned step(s).", proofID, len(scope.PlannedToolCalls)),
	})
	for _, call := range scope.PlannedToolCalls {
		toolName := strings.TrimSpace(call.Name)
		if toolName == "" {
			continue
		}
		turnIdx++
		callTeamID := firstNonEmptyString(call.Arguments["team_id"], teamID)
		summary := fmt.Sprintf("Executing %s", toolName)
		resultMsg := fmt.Sprintf("%s completed", toolName)
		switch toolName {
		case "create_team":
			cName := firstNonEmptyString(call.Arguments["name"], callTeamID, teamName)
			summary = fmt.Sprintf("Spawning team %s (%s)", cName, callTeamID)
			resultMsg = fmt.Sprintf("Team %s (%s) active and listening on swarm.team.%s.internal.command", cName, callTeamID, callTeamID)
		case "write_file":
			p := firstNonEmptyString(call.Arguments["path"], call.Arguments["file_path"])
			t := firstNonEmptyString(call.Arguments["title"], p)
			summary = fmt.Sprintf("Writing deliverable %s (%s)", t, p)
			resultMsg = fmt.Sprintf("Verified workspace artifact written to %s", p)
		}
		_, _ = s.Conversations.LogTurn(ctx, protocol.ConversationTurnData{
			RunID: runID, SessionID: runID, TenantID: "default", AgentID: "soma", TeamID: callTeamID,
			TurnIndex: turnIdx, Role: "tool_call", ToolName: toolName, ToolArgs: call.Arguments, Content: summary,
		})
		turnIdx++
		_, _ = s.Conversations.LogTurn(ctx, protocol.ConversationTurnData{
			RunID: runID, SessionID: runID, TenantID: "default", AgentID: "soma", TeamID: callTeamID,
			TurnIndex: turnIdx, Role: "tool_result", ToolName: toolName, Content: resultMsg,
		})
	}
	if strings.TrimSpace(finalSummary) != "" {
		turnIdx++
		_, _ = s.Conversations.LogTurn(ctx, protocol.ConversationTurnData{
			RunID: runID, SessionID: runID, TenantID: "default", AgentID: "soma", TeamID: teamID,
			TurnIndex: turnIdx, Role: "assistant", Content: finalSummary,
		})
	}
}

func (s *AdminServer) synthesizeRunConversationFromEvents(ctx context.Context, runID, agentFilter string) []conversations.ConversationTurn {
	timeline, err := s.Events.GetRunTimeline(ctx, runID)
	if err != nil || len(timeline) == 0 {
		return []conversations.ConversationTurn{}
	}
	turns := make([]conversations.ConversationTurn, 0, len(timeline))
	for idx, ev := range timeline {
		agentID := firstNonEmptyString(ev.SourceAgent, "soma")
		if agentFilter != "" && !strings.EqualFold(agentID, agentFilter) {
			continue
		}
		teamID := firstNonEmptyString(ev.SourceTeam, "governance")
		role := "assistant"
		toolName := ""
		content := fmt.Sprintf("Event `%s` recorded.", ev.EventType)
		switch string(ev.EventType) {
		case string(protocol.EventToolCompleted), "signal.published":
			role = "tool_result"
			tName := firstNonEmptyString(ev.Payload["team_name"], ev.Payload["team_id"], teamID)
			toolName = "create_team"
			content = fmt.Sprintf("Team **%s** (`%s`) activated and published readiness on `%s`.", tName, firstNonEmptyString(ev.Payload["team_id"], teamID), firstNonEmptyString(ev.Payload["speaks_on"], "swarm.team."+teamID+".signal.result"))
		case string(protocol.EventArtifactCreated):
			role = "tool_result"
			toolName = "write_file"
			basePath := firstNonEmptyString(ev.Payload["base_path"], "workspace")
			pkgKind := firstNonEmptyString(ev.Payload["package_kind"], "deliverable_package")
			content = fmt.Sprintf("Verified deliverable package (`%s`) written under `%s` with SHA-256 proof at `%s`.", pkgKind, basePath, firstNonEmptyString(ev.Payload["proof_path"], basePath+"/proof.md"))
		case string(protocol.EventMissionCompleted):
			role = "assistant"
			content = fmt.Sprintf("Run `%s` completed with execution state **%s**.", runID, firstNonEmptyString(ev.Payload["execution_state"], "verified"))
		case string(protocol.EventMissionFailed):
			role = "system"
			content = fmt.Sprintf("Run `%s` failed: %s", runID, firstNonEmptyString(ev.Payload["reason"], "execution error"))
		}
		turns = append(turns, conversations.ConversationTurn{
			ID:        ev.ID,
			RunID:     runID,
			SessionID: runID,
			TenantID:  firstNonEmptyString(ev.TenantID, "default"),
			AgentID:   agentID,
			TeamID:    teamID,
			TurnIndex: idx,
			Role:      role,
			ToolName:  toolName,
			ToolArgs:  ev.Payload,
			Content:   content,
			CreatedAt: ev.EmittedAt,
		})
	}
	return turns
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

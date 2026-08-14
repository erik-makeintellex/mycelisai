package server

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/mycelis/core/pkg/protocol"
)

func shouldSteerActiveWork(ctx *chatActiveWorkContext, request string) bool {
	if ctx == nil {
		return false
	}
	lower := normalizeIntentText(request)
	if lower == "" {
		return false
	}
	if requestContainsAny(lower, []string{"tell the team ", "ask the team ", "have the team "}) {
		return true
	}
	for _, prefix := range []string{"what ", "why ", "how ", "which ", "where ", "who ", "explain ", "tell me "} {
		if strings.HasPrefix(lower, prefix) {
			return false
		}
	}
	for _, phrase := range []string{
		"also ", "add ", "remove ", "change ", "revise ", "update ", "fix ", "improve ",
		"instead ", "make sure ", "ensure ", "include ", "do not ", "don't ", "tell the team ",
		"while you are working", "while they're working",
	} {
		if strings.Contains(lower, phrase) {
			return true
		}
	}
	return false
}

func (s *AdminServer) respondActiveWorkSteering(
	w http.ResponseWriter,
	r *http.Request,
	ctx *chatActiveWorkContext,
	latestUserText string,
	sessionID string,
	focusedTeamID string,
	sessionTurnIndex int,
) {
	request := teamWorkActionRequest{
		Action:         protocol.TeamWorkActionSteer,
		Summary:        latestUserText,
		ActorRef:       auditUserLabelFromRequest(r),
		SourceKind:     string(protocol.SourceKindWorkspaceUI),
		SourceChannel:  "soma.active_work.steer",
		PayloadKind:    "team_work_steering",
		IdempotencyKey: ctx.SteeringID,
		ExpectedRunID:  ctx.RunID,
	}
	item, steeringKey, err := s.applyTeamWorkAction(r.Context(), ctx.TeamID, ctx.WorkItemID, request, protocol.TeamWorkActionSteer)
	if err != nil {
		status := http.StatusConflict
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			status = http.StatusServiceUnavailable
		}
		respondAPIError(w, "Soma could not attach that guidance to the active work: "+err.Error(), status)
		return
	}
	if steeringKey != "" {
		if err := s.DispatchOutbox.Activate(r.Context(), steeringKey); err != nil {
			log.Printf("active work steering %s remains staged for dispatch recovery: %v", steeringKey, err)
		}
	}

	replyText := "I passed that guidance to the team. The current work is still running, and I will bring the result back here."
	logSomaConversationTurn(r.Context(), s.Conversations, sessionID, focusedTeamID, sessionTurnIndex, "user", latestUserText, chatAgentResult{})
	logSomaConversationTurn(r.Context(), s.Conversations, sessionID, focusedTeamID, sessionTurnIndex+1, "assistant", replyText, chatAgentResult{})
	auditID, _ := s.createAuditEvent(
		protocol.TemplateChatToAnswer,
		"active-work-steering",
		"Soma passed guidance to active team work",
		attachActorIdentity(map[string]any{
			"action":        "team_work_steered",
			"result_status": "queued",
			"run_id":        item.RunID,
			"team_id":       item.TeamID,
			"work_item_id":  item.WorkItemID,
		}, r),
	)
	payload := protocol.ChatResponsePayload{
		Text:          replyText,
		AskClass:      protocol.AskClassDirectAnswer,
		ResponseDepth: protocol.ResponseDepthQuickBox,
		Provenance: &protocol.AnswerProvenance{
			ResolvedIntent:  "steer_active_work",
			PermissionCheck: "pass",
			PolicyDecision:  "allow_within_active_contract",
			AuditEventID:    auditID,
		},
	}
	payloadBytes, _ := json.Marshal(payload)
	envelope := protocol.CTSEnvelope{
		Meta:       protocol.CTSMeta{SourceNode: "admin", Timestamp: time.Now()},
		SignalType: protocol.SignalChatResponse,
		TrustScore: protocol.TrustScoreCognitive,
		Payload:    payloadBytes,
		TemplateID: protocol.TemplateChatToAnswer,
		Mode:       protocol.ModeAnswer,
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(envelope))
}

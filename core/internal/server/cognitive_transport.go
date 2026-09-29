package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
)

func (s *AdminServer) requestChatAgent(parent context.Context, subject string, messages []chatRequestMessage) (chatAgentResult, error) {
	payload, err := json.Marshal(messages)
	if err != nil {
		return chatAgentResult{}, err
	}

	reqCtx, cancel := context.WithTimeout(parent, chatAgentRequestTimeout())
	defer cancel()

	// SRU: the agent recalls saved memory as the requesting user.
	request, release := s.recallTurnMsg(parent, subject, payload)
	defer release()
	msg, err := s.NC.RequestMsgWithContext(reqCtx, request)
	if err != nil {
		return chatAgentResult{}, err
	}
	return decodeChatAgentResult(msg.Data), nil
}

func chatAgentRequestTimeout() time.Duration {
	raw := strings.TrimSpace(os.Getenv("MYCELIS_CHAT_AGENT_TIMEOUT_SECONDS"))
	if raw == "" {
		return 120 * time.Second
	}
	seconds, err := strconv.Atoi(raw)
	if err != nil || seconds < 30 {
		return 120 * time.Second
	}
	if seconds > 300 {
		seconds = 300
	}
	return time.Duration(seconds) * time.Second
}

func applyBrainProvenance(s *AdminServer, chatPayload *protocol.ChatResponsePayload, agentResult chatAgentResult) {
	// Context sources travel with every reply, with or without brain provenance.
	chatPayload.ContextSources = agentResult.ContextSources
	if agentResult.ProviderID == "" || s.Cognitive == nil {
		return
	}
	brain := &protocol.BrainProvenance{
		ProviderID: agentResult.ProviderID,
		ModelID:    agentResult.ModelUsed,
	}
	if s.Cognitive.Config != nil {
		if pCfg, ok := s.Cognitive.ProviderSnapshot(agentResult.ProviderID); ok {
			brain.ProviderName = agentResult.ProviderID
			brain.Location = pCfg.Location
			brain.DataBoundary = pCfg.DataBoundary
			if brain.Location == "" {
				brain.Location = "local"
			}
			if brain.DataBoundary == "" {
				brain.DataBoundary = "local_only"
			}
		}
	}
	chatPayload.Brain = brain
}

func respondStructuredChatBlocker(w http.ResponseWriter, agentResult chatAgentResult) {
	blocker := buildChatBlocker(agentResult, "Soma could not produce a readable reply for that request.")
	status := http.StatusBadGateway
	switch blocker.Code {
	case emptyProviderOutputCode:
	case cognitive.TokenBudgetExhaustedCode:
		status = http.StatusTooManyRequests // honest budget stop, never an approval request
	default:
		status = http.StatusServiceUnavailable
	}
	respondAPIJSON(w, status, protocol.APIResponse{
		OK:    false,
		Error: blocker.Summary,
		Data:  blocker,
	})
}

// buildTransportChatBlocker classifies a chat transport failure. Summary and
// RecommendedAction are user-safe; AdminAction names the runtime to inspect.
func buildTransportChatBlocker(targetLabel string, err error) (int, cognitive.ExecutionAvailability) {
	lower := strings.ToLower(strings.TrimSpace(err.Error()))
	blocker := func(code, summary, action, adminAction string) cognitive.ExecutionAvailability {
		return cognitive.ExecutionAvailability{Available: false, Code: code, Summary: fmt.Sprintf(summary, targetLabel),
			RecommendedAction: action, AdminAction: adminAction}
	}
	const askAdmin = " If it keeps happening, ask an admin to check that Soma's services are running."
	switch {
	case errors.Is(err, context.DeadlineExceeded) || strings.Contains(lower, "deadline exceeded") || strings.Contains(lower, "timeout"):
		return http.StatusGatewayTimeout, blocker("transport_timeout", "%s took too long to answer.",
			"Try again."+askAdmin,
			"Retry once. If the timeout repeats, inspect NATS connectivity and the target agent runtime.")
	case strings.Contains(lower, "outbound buffer limit exceeded"):
		return http.StatusServiceUnavailable, blocker("transport_backpressure", "%s is busy right now and couldn't take the request.",
			"Wait a moment, then try again."+askAdmin,
			"Retry once. If this repeats, inspect NATS backpressure and recent agent traffic.")
	case errors.Is(err, nats.ErrNoResponders) || strings.Contains(lower, "no responders") || strings.Contains(lower, "not connected") || strings.Contains(lower, "connection closed") || strings.Contains(lower, "disconnected"):
		return http.StatusServiceUnavailable, blocker("transport_unavailable", "%s can't be reached right now.",
			"Try again in a moment."+askAdmin,
			"Inspect NATS connectivity and confirm the target agent runtime is online before retrying.")
	default:
		return http.StatusBadGateway, blocker("transport_unavailable", "%s didn't finish the request.",
			"Try again."+askAdmin,
			"Retry once. If it persists, inspect NATS connectivity and recent runtime logs.")
	}
}

func respondChatTransportBlocker(w http.ResponseWriter, r *http.Request, targetLabel string, err error) {
	status, availability := buildTransportChatBlocker(targetLabel, err)
	respondAPIJSON(w, status, protocol.APIResponse{
		OK:    false,
		Error: availability.Summary,
		Data:  availabilityForViewer(r, availability),
	})
}

func (s *AdminServer) chatExecutionAvailability() cognitive.ExecutionAvailability {
	if s == nil || s.Cognitive == nil {
		return cognitive.ExecutionAvailability{
			Available:         false,
			Code:              cognitive.ExecutionRouterUnavailable,
			Summary:           cognitive.SummaryRouterUnavailable,
			RecommendedAction: cognitive.UserEngineSetupAction,
			AdminAction:       cognitive.AdminEngineSetupAction,
			Profile:           "chat",
			SetupRequired:     true,
			SetupPath:         cognitive.DefaultExecutionSetupPath,
		}
	}
	return s.Cognitive.ExecutionAvailability("chat", "")
}

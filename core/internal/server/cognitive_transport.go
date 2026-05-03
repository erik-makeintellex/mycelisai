package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
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

	reqCtx, cancel := context.WithTimeout(parent, 60*time.Second)
	defer cancel()

	msg, err := s.NC.RequestWithContext(reqCtx, subject, payload)
	if err != nil {
		return chatAgentResult{}, err
	}
	return decodeChatAgentResult(msg.Data), nil
}

func applyBrainProvenance(s *AdminServer, chatPayload *protocol.ChatResponsePayload, agentResult chatAgentResult) {
	if agentResult.ProviderID == "" || s.Cognitive == nil {
		return
	}
	brain := &protocol.BrainProvenance{
		ProviderID: agentResult.ProviderID,
		ModelID:    agentResult.ModelUsed,
	}
	if s.Cognitive.Config != nil {
		if pCfg, ok := s.Cognitive.Config.Providers[agentResult.ProviderID]; ok {
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
	if blocker.Code != emptyProviderOutputCode {
		status = http.StatusServiceUnavailable
	}
	respondAPIJSON(w, status, protocol.APIResponse{
		OK:    false,
		Error: blocker.Summary,
		Data:  blocker,
	})
}

func buildTransportChatBlocker(targetLabel string, err error) (int, cognitive.ExecutionAvailability) {
	lower := strings.ToLower(strings.TrimSpace(err.Error()))
	switch {
	case errors.Is(err, context.DeadlineExceeded) || strings.Contains(lower, "deadline exceeded") || strings.Contains(lower, "timeout"):
		return http.StatusGatewayTimeout, cognitive.ExecutionAvailability{
			Available:         false,
			Code:              "transport_timeout",
			Summary:           fmt.Sprintf("%s did not respond before the request deadline.", targetLabel),
			RecommendedAction: "Retry once. If the timeout repeats, inspect NATS connectivity and the target agent runtime.",
		}
	case strings.Contains(lower, "outbound buffer limit exceeded"):
		return http.StatusServiceUnavailable, cognitive.ExecutionAvailability{
			Available:         false,
			Code:              "transport_backpressure",
			Summary:           fmt.Sprintf("%s is overloaded right now and could not process the request.", targetLabel),
			RecommendedAction: "Retry once. If this repeats, inspect NATS backpressure and recent swarm traffic.",
		}
	case errors.Is(err, nats.ErrNoResponders) || strings.Contains(lower, "no responders") || strings.Contains(lower, "not connected") || strings.Contains(lower, "connection closed") || strings.Contains(lower, "disconnected"):
		return http.StatusServiceUnavailable, cognitive.ExecutionAvailability{
			Available:         false,
			Code:              "transport_unavailable",
			Summary:           fmt.Sprintf("%s is currently unreachable from the workspace runtime.", targetLabel),
			RecommendedAction: "Inspect NATS connectivity and confirm the target agent runtime is online before retrying.",
		}
	default:
		return http.StatusBadGateway, cognitive.ExecutionAvailability{
			Available:         false,
			Code:              "transport_unavailable",
			Summary:           fmt.Sprintf("%s could not complete the request because the agent runtime did not respond cleanly.", targetLabel),
			RecommendedAction: "Retry once. If it persists, inspect NATS connectivity and recent runtime logs.",
		}
	}
}

func respondChatTransportBlocker(w http.ResponseWriter, targetLabel string, err error) {
	status, availability := buildTransportChatBlocker(targetLabel, err)
	respondAPIJSON(w, status, protocol.APIResponse{
		OK:    false,
		Error: availability.Summary,
		Data:  availability,
	})
}

func (s *AdminServer) chatExecutionAvailability() cognitive.ExecutionAvailability {
	if s == nil || s.Cognitive == nil {
		return cognitive.ExecutionAvailability{
			Available:         false,
			Code:              cognitive.ExecutionRouterUnavailable,
			Summary:           "Soma does not have an available cognitive engine right now.",
			RecommendedAction: "Open Settings and verify that at least one AI Engine is enabled and reachable for Soma.",
			Profile:           "chat",
			SetupRequired:     true,
			SetupPath:         cognitive.DefaultExecutionSetupPath,
		}
	}
	return s.Cognitive.ExecutionAvailability("chat", "")
}

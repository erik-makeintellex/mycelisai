package swarm

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/mycelis/core/internal/cognitive"
)

var agentInferenceTimeout = 120 * time.Second

func (a *Agent) inferWithExecutionBounds(req cognitive.InferRequest, reason string, attempt int) (*cognitive.InferResponse, error) {
	parent := a.ctx
	if parent == nil {
		parent = context.Background()
	}
	ctx, cancel := context.WithTimeout(parent, agentInferenceTimeout)
	defer cancel()
	req.Messages = systemMessagesFirst(req.Messages)
	messageChars := 0
	for _, message := range req.Messages {
		messageChars += len(message.Content)
	}
	providerID, modelID, maxOutput := a.inferenceLogConfig(req)
	started := time.Now()
	resp, err := a.brain.InferWithContract(ctx, req)
	status := "completed"
	if err != nil {
		status = "failed"
	}
	log.Printf("Agent [%s] inference attempt=%d reason=%s provider=%s model=%s messages=%d chars=%d max_output_tokens=%d elapsed_ms=%d status=%s",
		a.Manifest.ID, attempt, reason, providerID, modelID, len(req.Messages), messageChars, maxOutput, time.Since(started).Milliseconds(), status)
	if err != nil && ctx.Err() == context.DeadlineExceeded {
		return nil, fmt.Errorf("provider inference timeout after %s: %w", agentInferenceTimeout, context.DeadlineExceeded)
	}
	return resp, err
}

func systemMessagesFirst(messages []cognitive.ChatMessage) []cognitive.ChatMessage {
	ordered := make([]cognitive.ChatMessage, 0, len(messages))
	for _, message := range messages {
		if message.Role == "system" {
			ordered = append(ordered, message)
		}
	}
	for _, message := range messages {
		if message.Role != "system" {
			ordered = append(ordered, message)
		}
	}
	return ordered
}

func (a *Agent) inferenceLogConfig(req cognitive.InferRequest) (string, string, int) {
	providerID := strings.TrimSpace(req.Provider)
	if providerID == "" && a.brain != nil && a.brain.Config != nil {
		providerID = strings.TrimSpace(a.brain.Config.Profiles[req.Profile])
	}
	modelID, maxOutput := "unknown", 0
	if a.brain != nil && a.brain.Config != nil {
		if config, ok := a.brain.Config.Providers[providerID]; ok {
			config = cognitive.NormalizeProviderTokenDefaults(config)
			modelID = firstNonEmptyString(strings.TrimSpace(config.ModelID), "unknown")
			maxOutput = config.MaxOutputTokens
		}
	}
	return firstNonEmptyString(providerID, "auto"), modelID, maxOutput
}

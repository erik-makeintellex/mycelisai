package swarm

import (
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/pkg/protocol"
)

// ProcessResult holds the structured output of a processMessage call.
type ProcessResult struct {
	Text          string                           `json:"text"`
	ToolsUsed     []string                         `json:"tools_used,omitempty"`
	Artifacts     []protocol.ChatArtifactRef       `json:"artifacts,omitempty"`
	Availability  *cognitive.ExecutionAvailability `json:"availability,omitempty"`
	ProviderID    string                           `json:"provider_id,omitempty"`
	ModelUsed     string                           `json:"model_used,omitempty"`
	Consultations []protocol.ConsultationEntry     `json:"consultations,omitempty"`
}

func (a *Agent) processMessage(input string, priorHistory []cognitive.ChatMessage) string {
	return a.processMessageStructured(input, priorHistory).Text
}

func (a *Agent) processMessageStructured(input string, priorHistory []cognitive.ChatMessage) ProcessResult {
	if a.brain == nil {
		log.Printf("Agent [%s] has no brain. Skipping inference.", a.Manifest.ID)
		return ProcessResult{Availability: &cognitive.ExecutionAvailability{
			Available: false, Code: cognitive.ExecutionRouterUnavailable, Summary: "Soma does not have an available cognitive engine right now.",
			RecommendedAction: "Open Settings and verify that at least one AI Engine is enabled and reachable for Soma.", Profile: "chat", SetupRequired: true, SetupPath: cognitive.DefaultExecutionSetupPath,
		}}
	}
	if a.conversationLogger != nil {
		a.sessionID = uuid.New().String()
		a.turnIndex = 0
	}

	req, profile := a.buildInferRequest(input, priorHistory)
	resp, err := a.brain.InferWithContract(a.ctx, req)
	if err != nil {
		log.Printf("Agent [%s] brain freeze: %v", a.Manifest.ID, err)
		availability := a.brain.ExecutionAvailability(profile, a.Manifest.Provider)
		if availability.Summary == "" {
			availability.Summary = "Soma does not have an available cognitive engine right now."
		}
		return ProcessResult{Availability: &availability}
	}

	loop := a.runToolLoop(input, priorHistory, &req, resp, profile)
	responseText := stripToolCallJSON(loop.responseText)
	if a.internalTools != nil && len(priorHistory) > 0 && len(priorHistory)%15 == 0 {
		histCopy := make([]cognitive.ChatMessage, len(priorHistory))
		copy(histCopy, priorHistory)
		go a.internalTools.AutoSummarize(a.ctx, a.Manifest.ID, a.TeamID, histCopy)
	}

	providerID, modelUsed := "", ""
	if loop.resp != nil {
		providerID = loop.resp.Provider
		modelUsed = loop.resp.ModelUsed
	}
	if strings.TrimSpace(responseText) == "" {
		summary := "Soma could not produce a readable reply for that request."
		if len(loop.toolsUsed) > 0 {
			summary = fmt.Sprintf("Soma captured tool intent (%s) but the provider did not return a readable reply.", strings.Join(loop.toolsUsed, ", "))
		}
		availability := cognitive.ExecutionAvailability{
			Available: false, Code: "empty_provider_output", Summary: summary,
			RecommendedAction: "Retry the request. If the issue persists, inspect the configured provider output or switch to a different engine.",
			Profile:           profile, ProviderID: providerID, ModelID: modelUsed,
		}
		a.logTurn("assistant", availability.Summary, providerID, modelUsed, "", nil, "", "")
		return ProcessResult{ToolsUsed: loop.toolsUsed, Artifacts: loop.artifacts, Availability: &availability, ProviderID: providerID, ModelUsed: modelUsed, Consultations: loop.consultations}
	}

	a.logTurn("assistant", responseText, providerID, modelUsed, "", nil, "", "")
	return ProcessResult{Text: responseText, ToolsUsed: loop.toolsUsed, Artifacts: loop.artifacts, ProviderID: providerID, ModelUsed: modelUsed, Consultations: loop.consultations}
}

func (a *Agent) buildInferRequest(input string, priorHistory []cognitive.ChatMessage) (cognitive.InferRequest, string) {
	sys := a.Manifest.SystemPrompt
	if sys == "" {
		sys = fmt.Sprintf("You are a %s in the %s team.", a.Manifest.Role, a.TeamID)
	}
	if a.internalTools != nil {
		sys += a.internalTools.BuildContext(a.Manifest.ID, a.TeamID, a.Manifest.Role, a.TeamInputs, a.TeamDeliveries, input)
	}
	sys += a.buildToolsBlock()

	messages := []cognitive.ChatMessage{{Role: "system", Content: sys}}
	if len(priorHistory) > 0 {
		messages = append(messages, priorHistory...)
	}
	messages = append(messages, cognitive.ChatMessage{Role: "user", Content: input})
	a.logTurn("system", sys, "", "", "", nil, "", "")
	a.logTurn("user", input, "", "", "", nil, "", "")

	profile := "chat"
	if a.Manifest.Model != "" {
		profile = a.Manifest.Model
	}
	return cognitive.InferRequest{Profile: profile, Provider: a.Manifest.Provider, Messages: messages}, profile
}

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
		availability.Available = false
		if availability.Code == "" || availability.Code == cognitive.ExecutionAvailable {
			availability.Code = cognitive.ExecutionProviderOffline
		}
		if availability.Summary == "" || availability.Summary == "Soma has an available cognitive engine." {
			availability.Summary = "Soma's configured AI Engine (Ollama / local provider) is unreachable right now."
		}
		if availability.RecommendedAction == "" {
			availability.RecommendedAction = "Verify that your local AI Engine (e.g. `ollama serve` on port 11434) is running and reachable, or configure an alternative AI Engine in Settings."
		}
		availability.SetupRequired = true
		if availability.SetupPath == "" {
			availability.SetupPath = cognitive.DefaultExecutionSetupPath
		}
		return ProcessResult{Availability: &availability}
	}

	loop := a.runToolLoop(input, priorHistory, &req, resp, profile)
	responseText := stripToolCallJSON(loop.responseText)
	if isRefusalAssistantReply(responseText) && len(loop.toolsUsed) == 0 {
		cleanInput, workspaceCtx, _, _ := extractDirectAnswerPayload(input, priorHistory)
		recoverySys := "You are Soma, the helpful central orchestrator of the Mycelis AI Organization platform. Answer the user's request directly, constructively, and concisely in plain text."
		if workspaceCtx != "" {
			recoverySys += "\n\n" + workspaceCtx
		}
		recoveryReq := cognitive.InferRequest{
			Profile:  profile,
			Provider: a.Manifest.Provider,
			Messages: []cognitive.ChatMessage{
				{Role: "system", Content: recoverySys},
				{Role: "user", Content: cleanInput},
			},
		}
		if recovered, recErr := a.brain.InferWithContract(a.ctx, recoveryReq); recErr == nil && recovered != nil {
			if candidate := strings.TrimSpace(stripToolCallJSON(recovered.Text)); candidate != "" && !isRefusalAssistantReply(candidate) {
				responseText = candidate
				loop.resp = recovered
			}
		}
	}
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
	cleanInput, workspaceCtx, cleanHistory, isDirect := extractDirectAnswerPayload(input, priorHistory)
	var sys string
	if isDirect {
		sys = "You are Soma, the central orchestrator and executive coordinator of the Mycelis AI Organization platform.\n" +
			"Answer the user's request directly, helpfully, and concisely in well-formatted GitHub-Flavored Markdown (use clear `##` section headings, bulleted or numbered lists with **bold** lead-ins, `[Title](url)` links, and markdown tables or fenced code blocks where appropriate).\n" +
			"Use the live runtime context below when answering questions about Mycelis, active teams, services, or current work.\n" +
			"If asked for the secret codeword, reply with \"BLUEBERRY\".\n" +
			"Do not output tool_call JSON in this response."
		if a.internalTools != nil {
			sys += a.internalTools.BuildDirectAnswerContext(a.Manifest.ID, a.TeamID, a.Manifest.Role, a.TeamInputs, a.TeamDeliveries, cleanInput)
		}
	} else {
		sys = sanitizeSystemPromptForModelSafety(a.Manifest.SystemPrompt)
		if sys == "" {
			sys = fmt.Sprintf("You are a %s in the %s team.", a.Manifest.Role, a.TeamID)
		}
		sys += "\nWhen replying to the user with final results, format your response in clean GitHub-Flavored Markdown (bulleted/numbered lists with **bold** labels, `##` headings, and `[Title](url)` links)."
		if a.internalTools != nil {
			sys += a.internalTools.BuildContext(a.Manifest.ID, a.TeamID, a.Manifest.Role, a.TeamInputs, a.TeamDeliveries, cleanInput)
		}
		sys += a.buildToolsBlock()
	}
	if workspaceCtx != "" {
		sys += "\n\n" + workspaceCtx
	}

	messages := []cognitive.ChatMessage{{Role: "system", Content: sys}}
	if len(cleanHistory) > 0 {
		messages = append(messages, cleanHistory...)
	}
	messages = append(messages, cognitive.ChatMessage{Role: "user", Content: cleanInput})
	a.logTurn("system", sys, "", "", "", nil, "", "")
	a.logTurn("user", cleanInput, "", "", "", nil, "", "")

	profile := "chat"
	if a.Manifest.Model != "" {
		profile = a.Manifest.Model
	}
	return cognitive.InferRequest{Profile: profile, Provider: a.Manifest.Provider, Messages: messages}, profile
}

package swarm

import (
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/pkg/protocol"
)

// ProcessResult holds the structured output of a processMessage call.
type ProcessResult struct {
	Text             string                           `json:"text"`
	ToolsUsed        []string                         `json:"tools_used,omitempty"`
	PlannedToolCalls []protocol.PlannedToolCall       `json:"planned_tool_calls,omitempty"`
	Artifacts        []protocol.ChatArtifactRef       `json:"artifacts,omitempty"`
	Availability     *cognitive.ExecutionAvailability `json:"availability,omitempty"`
	ProviderID       string                           `json:"provider_id,omitempty"`
	ModelUsed        string                           `json:"model_used,omitempty"`
	Consultations    []protocol.ConsultationEntry     `json:"consultations,omitempty"`
}

func (a *Agent) processMessage(input string, priorHistory []cognitive.ChatMessage) string {
	return a.processMessageStructured(input, priorHistory).Text
}

func (a *Agent) processMessageStructured(input string, priorHistory []cognitive.ChatMessage) ProcessResult {
	return a.processMessageStructuredWithPosture(input, priorHistory, true)
}

func (a *Agent) processMessageStructuredWithPosture(input string, priorHistory []cognitive.ChatMessage, planningOnly bool) ProcessResult {
	return a.processMessageStructuredWithRequirement(input, priorHistory, planningOnly, nil)
}

func (a *Agent) processMessageStructuredWithRequirement(input string, priorHistory []cognitive.ChatMessage, planningOnly bool, requirement *teamResultRequirement) ProcessResult {
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
	if requirement.active() {
		req.Messages = append([]cognitive.ChatMessage{{Role: "system", Content: resultContractExecutionPrompt(requirement)}}, req.Messages...)
	}
	resp, err := a.inferWithExecutionBounds(req, "initial", 1)
	if err != nil {
		log.Printf("Agent [%s] brain freeze: %v", a.Manifest.ID, err)
		if fallback, ok := a.tryInitialProjectPackageRuntimeFallback(input, requirement, planningOnly); ok {
			return fallback
		}
		availability := a.brain.ExecutionAvailability(profile, a.Manifest.Provider)
		availability.Available = false
		availability.Code = "provider_inference_failed"
		availability.Summary = "The team could not complete inference with its configured cognitive engine."
		availability.RecommendedAction = "Retry the work after checking the configured engine. If the issue persists, choose another available engine."
		if availability.Profile == "" {
			availability.Profile = profile
		}
		return ProcessResult{Availability: &availability}
	}
	if resp != nil && strings.TrimSpace(resp.Text) == "" {
		req.Messages = append(req.Messages,
			cognitive.ChatMessage{Role: "system", Content: "Recovery correction: the previous response was empty. Return a concise direct answer, emit exactly one available tool_call JSON needed to complete the ask, or state one concrete blocker. Do not return an empty response."},
			cognitive.ChatMessage{Role: "user", Content: "Retry the latest request now under the recovery correction."},
		)
		recovered, recoverErr := a.inferWithExecutionBounds(req, "empty_response", 2)
		if recoverErr != nil {
			log.Printf("Agent [%s] empty-response recovery failed: %v", a.Manifest.ID, recoverErr)
		} else if recovered != nil {
			resp = recovered
		}
	}

	loop := a.runToolLoop(input, priorHistory, &req, resp, profile, planningOnly, requirement)
	loop.artifacts = reconcileToolBackedArtifacts(loop.artifacts, loop.toolEvidence, input)
	loop.artifacts = dedupeAgentArtifacts(loop.artifacts)
	if len(resultContractIssues(requirement, loop.artifacts, loop.toolEvidence)) > 0 &&
		a.completeProjectPackageRuntimeFallback(input, requirement, &loop, planningOnly) {
		loop.artifacts = reconcileToolBackedArtifacts(loop.artifacts, loop.toolEvidence, input)
		loop.artifacts = dedupeAgentArtifacts(loop.artifacts)
	}
	responseText := stripToolCallJSON(loop.responseText)
	if strings.TrimSpace(responseText) == "" && len(loop.plannedCalls) > 0 {
		responseText = "Soma prepared the requested governed action for approval."
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
	if issues := resultContractIssues(requirement, loop.artifacts, loop.toolEvidence); len(issues) > 0 {
		summary := "The team exhausted its bounded correction attempts without satisfying the approved result contract: " + strings.Join(issues, "; ") + "."
		availability := cognitive.ExecutionAvailability{
			Available: false, Code: "result_contract_unsatisfied", Summary: summary,
			RecommendedAction: resultContractRecoveryAction(requirement, issues),
			Profile:           profile, ProviderID: providerID, ModelID: modelUsed,
		}
		a.logTurn("assistant", availability.Summary, providerID, modelUsed, "", nil, "", "")
		responseText = resultContractDegradedResponseText(requirement)
		return ProcessResult{Text: responseText, ToolsUsed: loop.toolsUsed, PlannedToolCalls: loop.plannedCalls, Artifacts: loop.artifacts, Availability: &availability, ProviderID: providerID, ModelUsed: modelUsed, Consultations: loop.consultations}
	}
	if strings.TrimSpace(responseText) == "" && len(loop.artifacts) > 0 {
		responseText = retainedArtifactCompletionSummary(loop.artifacts)
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
	return ProcessResult{Text: responseText, ToolsUsed: loop.toolsUsed, PlannedToolCalls: loop.plannedCalls, Artifacts: loop.artifacts, ProviderID: providerID, ModelUsed: modelUsed, Consultations: loop.consultations}
}

func dedupeAgentArtifacts(artifacts []protocol.ChatArtifactRef) []protocol.ChatArtifactRef {
	if len(artifacts) < 2 {
		return artifacts
	}
	seen := make(map[string]struct{}, len(artifacts))
	unique := make([]protocol.ChatArtifactRef, 0, len(artifacts))
	for _, artifact := range artifacts {
		if strings.EqualFold(strings.TrimSpace(artifact.Type), "project_package") {
			merged := false
			for index := range unique {
				if !sameLogicalProjectPackage(unique[index], artifact) {
					continue
				}
				if projectPackageCompleteness(artifact) > projectPackageCompleteness(unique[index]) {
					unique[index] = artifact
				}
				merged = true
				break
			}
			if merged {
				continue
			}
		}
		comparable := artifact
		comparable.ID = ""
		raw, err := json.Marshal(comparable)
		key := string(raw)
		if err != nil {
			key = strings.Join([]string{artifact.Type, artifact.Title, artifact.ContentType, artifact.Content, artifact.URL, artifact.SavedPath, artifact.Entrypoint, artifact.Folder}, "\x00")
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		unique = append(unique, artifact)
	}
	return unique
}

func sameLogicalProjectPackage(left, right protocol.ChatArtifactRef) bool {
	if !strings.EqualFold(strings.TrimSpace(left.Type), "project_package") ||
		!strings.EqualFold(strings.TrimSpace(right.Type), "project_package") {
		return false
	}
	leftLocation := projectPackageLocation(left)
	rightLocation := projectPackageLocation(right)
	if leftLocation != "" && rightLocation != "" {
		return strings.EqualFold(leftLocation, rightLocation)
	}
	return strings.EqualFold(strings.TrimSpace(left.Title), strings.TrimSpace(right.Title)) && strings.TrimSpace(left.Title) != ""
}

func projectPackageLocation(artifact protocol.ChatArtifactRef) string {
	location := strings.TrimSpace(artifact.Folder)
	if location == "" {
		location = strings.TrimSpace(artifact.SavedPath)
	}
	if location == "" && strings.TrimSpace(artifact.Entrypoint) != "" {
		location = filepath.Dir(strings.ReplaceAll(strings.TrimSpace(artifact.Entrypoint), "\\", "/"))
	}
	return strings.Trim(strings.ReplaceAll(location, "\\", "/"), "/")
}

func projectPackageCompleteness(artifact protocol.ChatArtifactRef) int {
	score := 0
	for _, value := range []string{artifact.Folder, artifact.SavedPath, artifact.Entrypoint, artifact.Validation} {
		if strings.TrimSpace(value) != "" {
			score += 2
		}
	}
	if len(artifact.Files) > 0 {
		score += 2 + len(artifact.Files)
	}
	if strings.Contains(strings.TrimSpace(artifact.Content), "{") {
		score++
	}
	return score
}

func retainedArtifactCompletionSummary(artifacts []protocol.ChatArtifactRef) string {
	if len(artifacts) == 1 {
		title := strings.TrimSpace(artifacts[0].Title)
		if title != "" {
			return fmt.Sprintf("Created retained output: %s.", title)
		}
		return "Created one retained output."
	}
	return fmt.Sprintf("Created %d retained outputs.", len(artifacts))
}

func resultContractDegradedResponseText(requirement *teamResultRequirement) string {
	if requirement != nil && strings.EqualFold(strings.TrimSpace(requirement.Kind), "project_package") {
		return "Team output needs repair before it can be delivered."
	}
	return "Team work needs repair before it can be delivered."
}

func (a *Agent) buildInferRequest(input string, priorHistory []cognitive.ChatMessage) (cognitive.InferRequest, string) {
	sys := a.Manifest.SystemPrompt
	if sys == "" {
		sys = fmt.Sprintf("You are a %s in the %s team.", a.Manifest.Role, a.TeamID)
	}
	sys += agentProfileContextDirective(a.Manifest)
	sys += runtimeResponseDirective()
	if a.internalTools != nil {
		sys += a.internalTools.BuildContext(a.Manifest.ID, a.TeamID, a.Manifest.Role, a.TeamInputs, a.TeamDeliveries, input)
	}
	sys += a.buildToolsBlock(input)

	messages := []cognitive.ChatMessage{{Role: "system", Content: sys}}
	if len(priorHistory) > 0 {
		for _, message := range priorHistory {
			if message.Role == "system" {
				messages = append(messages, message)
			}
		}
		for _, message := range priorHistory {
			if message.Role != "system" {
				messages = append(messages, message)
			}
		}
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

func agentProfileContextDirective(manifest protocol.AgentManifest) string {
	if strings.TrimSpace(manifest.ProfileRef) == "" && len(manifest.Context) == 0 {
		return ""
	}
	var bindings []string
	for _, binding := range manifest.Context {
		label := strings.TrimSpace(binding.Kind)
		if binding.Ref != "" {
			label += ":" + strings.TrimSpace(binding.Ref)
		}
		if binding.Access != "" {
			label += " (" + strings.TrimSpace(binding.Access) + ")"
		}
		if label != "" {
			bindings = append(bindings, label)
		}
	}
	return fmt.Sprintf("\n\n## Worker Profile\nProfile: %s\nSelection: %s; scope: %s.\nApproved context bindings: %s. These bindings describe intended context only and never bypass tool, mount, MCP, capability, approval, or workspace policy.\n",
		firstNonEmptyString(manifest.ProfileRef, "custom"), firstNonEmptyString(manifest.Usage.Selection, "explicit"),
		firstNonEmptyString(manifest.Usage.Scope, "team"), firstNonEmptyString(strings.Join(bindings, ", "), "none"))
}

func runtimeResponseDirective() string {
	return "\n\n## Runtime Response Contract\n" +
		"- Be terse in agent-to-agent and operator-facing response text. Prefer compact status fragments over full prose.\n" +
		"- Say only target, action, evidence/result, blocker, and next step needed for execution.\n" +
		"- For configured online search, use web_search without asking for confirmation. Disclose path first: provider/tool, no confirmation, external results are leads to verify.\n" +
		"- Do not add tutorial text, pleasantries, or broad explanations unless explicitly requested.\n"
}

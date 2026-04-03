package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/pkg/protocol"
)

// mutationTools are tools that trigger proposal mode when used in chat.
// If an agent uses any of these, the response switches from answer → proposal.
var mutationTools = map[string]bool{
	"generate_blueprint": true,
	"delegate":           true,
	"write_file":         true,
	"publish_signal":     true,
	"broadcast":          true,
}

const emptyProviderOutputCode = "empty_provider_output"
const governedMutationRoutePrefix = "[GOVERNED MUTATION ROUTE]"
const directAnswerRoutePrefix = "[DIRECT ANSWER ROUTE]"
const directAnswerRetryRoutePrefix = "[DIRECT ANSWER RETRY]"

var (
	namedFilePattern     = regexp.MustCompile("(?i)(?:named|called|at path|path)\\s+[`'\"]?([^`'\"\\s]+)[`'\"]?")
	printsPattern        = regexp.MustCompile("(?i)prints?\\s+[`'\"]?([^`'\".]+(?:\\s+[^`'\".]+)*)[`'\"]?")
	quotedContentPattern = regexp.MustCompile("(?i)(?:with content|containing|that says)\\s+[`'\"]([^`'\"]+)[`'\"]")
)

type chatRequestMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatAgentResult struct {
	Text          string                           `json:"text"`
	ToolsUsed     []string                         `json:"tools_used,omitempty"`
	Artifacts     []protocol.ChatArtifactRef       `json:"artifacts,omitempty"`
	Availability  *cognitive.ExecutionAvailability `json:"availability,omitempty"`
	ProviderID    string                           `json:"provider_id,omitempty"`
	ModelUsed     string                           `json:"model_used,omitempty"`
	Consultations []protocol.ConsultationEntry     `json:"consultations,omitempty"`
}

// hasMutationTools checks if any tools in the list are mutation tools.
func hasMutationTools(tools []string) (bool, []string) {
	var mutations []string
	for _, t := range tools {
		if mutationTools[t] {
			mutations = append(mutations, t)
		}
	}
	return len(mutations) > 0, mutations
}

// chatToolRisk estimates risk level from mutation tools used.
func chatToolRisk(tools []string) string {
	for _, t := range tools {
		if t == "publish_signal" || t == "broadcast" {
			return "high"
		}
		if t == "generate_blueprint" || t == "delegate" || t == "write_file" {
			return "medium"
		}
	}
	return "low"
}

func uniqueOrderedTools(tools []string) []string {
	seen := make(map[string]struct{}, len(tools))
	out := make([]string, 0, len(tools))
	for _, tool := range tools {
		t := strings.TrimSpace(tool)
		if t == "" {
			continue
		}
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	return out
}

func latestUserMessageIndex(messages []chatRequestMessage) int {
	for i := len(messages) - 1; i >= 0; i-- {
		if strings.EqualFold(strings.TrimSpace(messages[i].Role), "user") {
			return i
		}
	}
	return -1
}

func latestUserMessageContent(messages []chatRequestMessage) string {
	idx := latestUserMessageIndex(messages)
	if idx < 0 {
		return ""
	}
	return strings.TrimSpace(messages[idx].Content)
}

func requestContainsAny(lower string, needles []string) bool {
	for _, needle := range needles {
		if strings.Contains(lower, needle) {
			return true
		}
	}
	return false
}

func inferMutationToolsFromText(text string) []string {
	lower := strings.ToLower(strings.TrimSpace(text))
	if lower == "" {
		return nil
	}

	var tools []string

	fileActions := []string{"create", "write", "update", "edit", "modify", "replace", "append", "save", "persist", "store", "generate", "draft"}
	fileTargets := []string{"file", "folder", "directory", "workspace", "path", "script", "config", "json", "yaml", "yml", "toml", "markdown", "document", "doc", "repo", "repository", "codebase"}
	if requestContainsAny(lower, fileActions) && requestContainsAny(lower, fileTargets) {
		tools = append(tools, "write_file")
	}

	blueprintActions := []string{"create", "generate", "draft", "build", "compose", "design"}
	blueprintTargets := []string{"blueprint", "architecture", "spec", "proposal", "plan", "workflow", "mission"}
	if requestContainsAny(lower, blueprintActions) && requestContainsAny(lower, blueprintTargets) {
		tools = append(tools, "generate_blueprint")
	}

	delegationActions := []string{"delegate", "assign", "route", "hand off", "handoff", "send to", "consult"}
	delegationTargets := []string{"team", "agent", "council", "member", "task"}
	if requestContainsAny(lower, delegationActions) && requestContainsAny(lower, delegationTargets) {
		tools = append(tools, "delegate")
	}

	signalActions := []string{"publish", "emit", "send", "post"}
	signalTargets := []string{"signal", "status", "result", "event"}
	if requestContainsAny(lower, signalActions) && requestContainsAny(lower, signalTargets) {
		tools = append(tools, "publish_signal")
	}

	if requestContainsAny(lower, []string{"broadcast", "announce to all", "fan out", "broadcast to"}) {
		tools = append(tools, "broadcast")
	}

	return uniqueOrderedTools(tools)
}

func normalizeChatRequestMessages(messages []chatRequestMessage) ([]chatRequestMessage, []string) {
	idx := latestUserMessageIndex(messages)
	if idx < 0 {
		return messages, nil
	}

	trimmed := strings.TrimSpace(messages[idx].Content)
	mutTools := inferMutationToolsFromText(messages[idx].Content)
	normalized := make([]chatRequestMessage, len(messages))
	copy(normalized, messages)
	if len(mutTools) == 0 {
		normalized[idx].Content = directAnswerRoutePrefix + "\n" +
			"Answer the latest request directly in readable text. Do not call mutating tools, do not emit tool_call JSON, and do not route work unless the user explicitly asked to change something.\n\n" +
			"Original request:\n" + trimmed
		return normalized, nil
	}

	normalized[idx].Content = governedMutationRoutePrefix + "\n" +
		"Treat this latest request as governed proposal-only work and emit tool_call JSON if a mutation is needed.\n\n" +
		"Original request:\n" + trimmed

	return normalized, mutTools
}

func applyDirectAnswerRetryInstruction(messages []chatRequestMessage, latestRequest string) []chatRequestMessage {
	idx := latestUserMessageIndex(messages)
	if idx < 0 {
		return messages
	}

	normalized := make([]chatRequestMessage, len(messages))
	copy(normalized, messages)
	normalized[idx].Content = directAnswerRetryRoutePrefix + "\n" +
		"Answer the latest request directly in readable text. Do not call tools. Do not emit tool_call JSON. If more context is needed, ask one concise clarifying question.\n\n" +
		"Original request:\n" + strings.TrimSpace(latestRequest)
	return normalized
}

func resolvePrimaryChatAskClass(isMutation bool) protocol.AskClass {
	if isMutation {
		return protocol.AskClassGovernedMutation
	}
	return protocol.AskClassDirectAnswer
}

func resolvePrimaryChatAskContract(isMutation bool) protocol.AskContract {
	class := resolvePrimaryChatAskClass(isMutation)
	contract, ok := protocol.AskContractForClass(class)
	if !ok {
		// The first slice only supports direct answers and governed mutation.
		return protocol.AskContract{
			AskClass:             protocol.AskClassDirectAnswer,
			DefaultAgentTarget:   "soma",
			DefaultExecutionMode: protocol.ModeAnswer,
			TemplateID:           protocol.TemplateChatToAnswer,
		}
	}
	return contract
}

func parsePlannedToolCall(text string) (protocol.PlannedToolCall, bool) {
	var envelope struct {
		ToolCall struct {
			Name      string         `json:"name"`
			Arguments map[string]any `json:"arguments"`
		} `json:"tool_call"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(text)), &envelope); err != nil {
		return protocol.PlannedToolCall{}, false
	}
	if strings.TrimSpace(envelope.ToolCall.Name) == "" {
		return protocol.PlannedToolCall{}, false
	}
	if envelope.ToolCall.Arguments == nil {
		envelope.ToolCall.Arguments = map[string]any{}
	}
	return protocol.PlannedToolCall{
		Name:      strings.TrimSpace(envelope.ToolCall.Name),
		Arguments: envelope.ToolCall.Arguments,
	}, true
}

func inferWriteFilePlanFromRequest(text string) (protocol.PlannedToolCall, bool) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return protocol.PlannedToolCall{}, false
	}

	pathMatch := namedFilePattern.FindStringSubmatch(trimmed)
	if len(pathMatch) < 2 {
		return protocol.PlannedToolCall{}, false
	}

	targetPath := strings.TrimSpace(pathMatch[1])
	if targetPath == "" {
		return protocol.PlannedToolCall{}, false
	}

	content := ""
	if quoted := quotedContentPattern.FindStringSubmatch(trimmed); len(quoted) >= 2 {
		content = strings.TrimSpace(quoted[1])
	} else if prints := printsPattern.FindStringSubmatch(trimmed); len(prints) >= 2 {
		value := strings.TrimSpace(prints[1])
		switch strings.ToLower(filepathExt(targetPath)) {
		case ".py":
			content = fmt.Sprintf("print(%q)\n", value)
		default:
			content = value + "\n"
		}
	} else if strings.EqualFold(filepathExt(targetPath), ".py") {
		content = "print(\"hello world\")\n"
	}

	if content == "" {
		return protocol.PlannedToolCall{}, false
	}

	return protocol.PlannedToolCall{
		Name: "write_file",
		Arguments: map[string]any{
			"path":    targetPath,
			"content": content,
		},
	}, true
}

func filepathExt(targetPath string) string {
	lastDot := strings.LastIndex(targetPath, ".")
	if lastDot < 0 {
		return ""
	}
	return strings.ToLower(targetPath[lastDot:])
}

func buildPlannedToolCalls(agentResult chatAgentResult, latestRequest string, mutTools []string) []protocol.PlannedToolCall {
	var planned []protocol.PlannedToolCall
	if call, ok := parsePlannedToolCall(agentResult.Text); ok {
		planned = append(planned, call)
	}
	if len(planned) == 0 {
		for _, tool := range mutTools {
			if tool == "write_file" {
				if call, ok := inferWriteFilePlanFromRequest(latestRequest); ok {
					planned = append(planned, call)
				}
			}
		}
	}
	return planned
}

func affectedResourcesForPlannedCalls(planned []protocol.PlannedToolCall) []string {
	var resources []string
	for _, call := range planned {
		switch strings.TrimSpace(call.Name) {
		case "write_file":
			if rawPath, ok := call.Arguments["path"].(string); ok && strings.TrimSpace(rawPath) != "" {
				resources = append(resources, strings.TrimSpace(rawPath))
				continue
			}
		case "publish_signal":
			if subject, ok := call.Arguments["subject"].(string); ok && strings.TrimSpace(subject) != "" {
				resources = append(resources, strings.TrimSpace(subject))
				continue
			}
		}
		resources = append(resources, "state")
	}
	if len(resources) == 0 {
		return []string{"state"}
	}
	return uniqueOrderedTools(resources)
}

func inferAdapterKindFromTool(tool string) string {
	t := strings.ToLower(strings.TrimSpace(tool))
	switch {
	case strings.HasPrefix(t, "mcp_"), strings.Contains(t, "mcp"):
		return "mcp"
	case strings.HasPrefix(t, "http_"), strings.Contains(t, "api"), strings.Contains(t, "webhook"):
		return "openapi"
	case strings.HasPrefix(t, "host_"), t == "local_command":
		return "host"
	default:
		return "internal"
	}
}

func buildTeamExpressionsFromTools(tools []string, teamID string, rolePlan []string) []protocol.ChatTeamExpression {
	deduped := uniqueOrderedTools(tools)
	if strings.TrimSpace(teamID) == "" {
		teamID = "admin-core"
	}
	if len(rolePlan) == 0 {
		rolePlan = []string{"admin"}
	}
	expressions := make([]protocol.ChatTeamExpression, 0, len(deduped))
	for i, tool := range deduped {
		idx := i + 1
		bindingID := fmt.Sprintf("binding-%d-%s", idx, strings.ReplaceAll(tool, "_", "-"))
		expressionID := fmt.Sprintf("expr-%d", idx)
		expressions = append(expressions, protocol.ChatTeamExpression{
			ExpressionID: expressionID,
			TeamID:       teamID,
			Objective:    fmt.Sprintf("Execute %s through governed module binding", tool),
			RolePlan:     rolePlan,
			ModuleBindings: []protocol.ChatModuleBinding{
				{
					BindingID:   bindingID,
					ModuleID:    tool,
					AdapterKind: inferAdapterKindFromTool(tool),
					Operation:   tool,
				},
			},
		})
	}
	return expressions
}

type proposalDisplayContract struct {
	OperatorSummary   string
	ExpectedResult    string
	AffectedResources []string
}

func firstStringArgument(arguments map[string]any, key string) string {
	if arguments == nil {
		return ""
	}
	raw, ok := arguments[key]
	if !ok {
		return ""
	}
	value, ok := raw.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(value)
}

func formatProposalResource(resource string) string {
	trimmed := strings.TrimSpace(resource)
	if trimmed == "" {
		return ""
	}
	if trimmed == "state" {
		return "governed state"
	}
	return trimmed
}

func buildProposalDisplayContract(planned []protocol.PlannedToolCall, latestRequest string, mutTools []string) proposalDisplayContract {
	display := proposalDisplayContract{
		OperatorSummary: "Carry out the requested governed action.",
		ExpectedResult:  "Soma will perform the approved action and return durable execution proof.",
	}

	for _, resource := range affectedResourcesForPlannedCalls(planned) {
		if formatted := formatProposalResource(resource); formatted != "" {
			display.AffectedResources = append(display.AffectedResources, formatted)
		}
	}

	if len(planned) > 0 {
		switch strings.TrimSpace(planned[0].Name) {
		case "write_file":
			path := firstStringArgument(planned[0].Arguments, "path")
			if path != "" {
				display.OperatorSummary = fmt.Sprintf("Create %q in the workspace.", path)
				display.ExpectedResult = fmt.Sprintf("One new workspace file will be created at %q after approval.", path)
				if len(display.AffectedResources) == 0 {
					display.AffectedResources = []string{path}
				}
				return display
			}
			display.OperatorSummary = "Create a new workspace file."
			display.ExpectedResult = "One new workspace file will be created after approval."
			return display
		case "publish_signal":
			subject := firstStringArgument(planned[0].Arguments, "subject")
			if subject != "" {
				display.OperatorSummary = fmt.Sprintf("Publish a governed signal to %q.", subject)
				display.ExpectedResult = fmt.Sprintf("A signal will be sent on %q after approval.", subject)
				if len(display.AffectedResources) == 0 {
					display.AffectedResources = []string{subject}
				}
				return display
			}
			display.OperatorSummary = "Publish a governed signal."
			display.ExpectedResult = "A governed signal will be sent after approval."
			return display
		case "generate_blueprint":
			display.OperatorSummary = "Prepare a reusable blueprint from this request."
			display.ExpectedResult = "A governed blueprint draft will be created for review."
			return display
		case "delegate", "delegate_task":
			display.OperatorSummary = "Hand the requested work to the right team."
			display.ExpectedResult = "The approved task will be routed to the selected team with execution proof."
			return display
		case "broadcast":
			display.OperatorSummary = "Broadcast the requested update to connected teams."
			display.ExpectedResult = "The approved broadcast will be sent and logged with execution proof."
			return display
		}
	}

	if len(mutTools) > 0 {
		switch strings.TrimSpace(mutTools[0]) {
		case "write_file":
			display.OperatorSummary = "Create a new workspace file."
			display.ExpectedResult = "One new workspace file will be created after approval."
		case "publish_signal":
			display.OperatorSummary = "Publish a governed signal."
			display.ExpectedResult = "A governed signal will be sent after approval."
		case "generate_blueprint":
			display.OperatorSummary = "Prepare a reusable blueprint from this request."
			display.ExpectedResult = "A governed blueprint draft will be created for review."
		case "delegate", "delegate_task":
			display.OperatorSummary = "Hand the requested work to the right team."
			display.ExpectedResult = "The approved task will be routed to the selected team with execution proof."
		case "broadcast":
			display.OperatorSummary = "Broadcast the requested update to connected teams."
			display.ExpectedResult = "The approved broadcast will be sent and logged with execution proof."
		}
	}

	if strings.TrimSpace(latestRequest) != "" && display.OperatorSummary == "Carry out the requested governed action." {
		display.ExpectedResult = "Soma will carry out the approved request and return durable execution proof."
	}

	return display
}

func buildMutationChatProposal(mutTools []string, proofID, confirmToken, teamID string, rolePlan []string, approval *protocol.ApprovalPolicy, profile *protocol.GovernanceProfileSnapshot, display proposalDisplayContract) *protocol.ChatProposal {
	deduped := uniqueOrderedTools(mutTools)
	return &protocol.ChatProposal{
		Intent:            "chat-action",
		OperatorSummary:   display.OperatorSummary,
		ExpectedResult:    display.ExpectedResult,
		AffectedResources: display.AffectedResources,
		Tools:             deduped,
		RiskLevel:         chatToolRisk(deduped),
		ConfirmToken:      confirmToken,
		IntentProofID:     proofID,
		TeamExpressions:   buildTeamExpressionsFromTools(deduped, teamID, rolePlan),
		Approval:          approval,
		GovernanceProfile: profile,
	}
}

func decodeChatAgentResult(data []byte) chatAgentResult {
	var result chatAgentResult
	if err := json.Unmarshal(data, &result); err == nil && result.hasStructuredState() {
		return result
	}
	return chatAgentResult{
		Text: string(data),
	}
}

func (r chatAgentResult) hasStructuredState() bool {
	return strings.TrimSpace(r.Text) != "" ||
		len(r.ToolsUsed) > 0 ||
		len(r.Artifacts) > 0 ||
		r.Availability != nil ||
		strings.TrimSpace(r.ProviderID) != "" ||
		strings.TrimSpace(r.ModelUsed) != "" ||
		len(r.Consultations) > 0
}

func buildChatBlocker(agentResult chatAgentResult, fallbackSummary string) cognitive.ExecutionAvailability {
	if agentResult.Availability != nil {
		blocker := *agentResult.Availability
		if blocker.Summary == "" {
			blocker.Summary = fallbackSummary
		}
		if blocker.Code == "" {
			blocker.Code = emptyProviderOutputCode
		}
		if blocker.RecommendedAction == "" {
			blocker.RecommendedAction = "Retry the request. If the issue persists, inspect the configured provider output or switch to another engine."
		}
		if blocker.ProviderID == "" {
			blocker.ProviderID = agentResult.ProviderID
		}
		if blocker.ModelID == "" {
			blocker.ModelID = agentResult.ModelUsed
		}
		blocker.Available = false
		return blocker
	}
	return cognitive.ExecutionAvailability{
		Available:         false,
		Code:              emptyProviderOutputCode,
		Summary:           fallbackSummary,
		RecommendedAction: "Retry the request. If the issue persists, inspect the configured provider output or switch to another engine.",
		ProviderID:        agentResult.ProviderID,
		ModelID:           agentResult.ModelUsed,
	}
}

func readableChatText(agentResult chatAgentResult, isMutation bool) string {
	if isMutation {
		if _, ok := parsePlannedToolCall(agentResult.Text); ok {
			return "Soma captured a governed mutation intent. Review the proposal details below."
		}
	}
	if !isMutation && containsToolCallJSON(agentResult.Text) {
		return ""
	}
	if strings.TrimSpace(agentResult.Text) != "" {
		return agentResult.Text
	}
	if isMutation {
		return "Soma captured a governed mutation intent. Review the proposal details below."
	}
	if len(agentResult.Artifacts) > 0 {
		return "Soma returned artifacts for this request."
	}
	return ""
}

func mergeMutationTools(agentTools, requestTools []string) (bool, []string) {
	if len(requestTools) == 0 {
		return false, nil
	}
	combined := uniqueOrderedTools(append(append([]string{}, requestTools...), agentTools...))
	isMutation, mutTools := hasMutationTools(combined)
	return isMutation, mutTools
}

func containsToolCallJSON(text string) bool {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return false
	}
	if _, ok := parsePlannedToolCall(trimmed); ok {
		return true
	}
	return strings.Contains(trimmed, `"tool_call"`)
}

func shouldRetryDirectAnswer(agentResult chatAgentResult, requestMutationTools []string) bool {
	if len(requestMutationTools) != 0 {
		return false
	}
	if isMutation, _ := hasMutationTools(agentResult.ToolsUsed); isMutation {
		return true
	}
	return containsToolCallJSON(agentResult.Text)
}

func directAnswerDriftBlocker(agentResult chatAgentResult) chatAgentResult {
	return chatAgentResult{
		Availability: &cognitive.ExecutionAvailability{
			Available:         false,
			Code:              emptyProviderOutputCode,
			Summary:           "Soma drifted into a governed action while answering a read-only request. Retry the request or restate it more directly.",
			RecommendedAction: "Retry the request. If this repeats, simplify the question or inspect the active cognitive provider output.",
			ProviderID:        agentResult.ProviderID,
			ModelID:           agentResult.ModelUsed,
		},
		ProviderID: agentResult.ProviderID,
		ModelUsed:  agentResult.ModelUsed,
	}
}

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

// GET /api/v1/cognitive/status
// Returns health and configuration of all cognitive engines (vLLM text + Diffusers media).
func (s *AdminServer) HandleCognitiveStatus(w http.ResponseWriter, r *http.Request) {
	if s.Cognitive == nil || s.Cognitive.Config == nil {
		respondJSON(w, map[string]any{"text": map[string]string{"status": "offline"}, "media": map[string]string{"status": "offline"}})
		return
	}

	type engineStatus struct {
		Status            string `json:"status"`
		Endpoint          string `json:"endpoint,omitempty"`
		Model             string `json:"model,omitempty"`
		Detail            string `json:"detail,omitempty"`
		RecommendedAction string `json:"recommended_action,omitempty"`
		SetupRequired     bool   `json:"setup_required,omitempty"`
	}

	result := map[string]*engineStatus{
		"text":  {Status: "offline"},
		"media": {Status: "offline"},
	}

	// Probe all openai_compatible text engines (vLLM, Ollama, LM Studio, etc.)
	cfg := s.Cognitive.Config
	textAvailability := s.Cognitive.ExecutionAvailability("chat", "")
	if !textAvailability.Available {
		result["text"] = &engineStatus{
			Status:            "offline",
			Model:             textAvailability.ModelID,
			Detail:            textAvailability.Summary,
			RecommendedAction: textAvailability.RecommendedAction,
			SetupRequired:     textAvailability.SetupRequired,
		}
	}
	for provID, prov := range cfg.Providers {
		if prov.Type != "openai_compatible" || prov.Endpoint == "" {
			continue
		}
		adapter, ok := s.Cognitive.Adapters[provID]
		if !ok {
			continue
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		alive, _ := adapter.Probe(ctx)
		cancel()
		if alive {
			result["text"] = &engineStatus{
				Status:   "online",
				Endpoint: prov.Endpoint,
				Model:    prov.ModelID,
			}
			break
		}
	}

	// Probe media engine
	if cfg.Media != nil && cfg.Media.Endpoint != "" {
		healthURL := cfg.Media.Endpoint[:len(cfg.Media.Endpoint)-3] + "/health" // strip /v1, add /health
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		httpReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
		resp, err := http.DefaultClient.Do(httpReq)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				result["media"] = &engineStatus{
					Status:   "online",
					Endpoint: cfg.Media.Endpoint,
					Model:    cfg.Media.ModelID,
				}
			}
		}
	}

	respondJSON(w, result)
}

// POST /api/v1/cognitive/infer
func (s *AdminServer) handleInfer(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.Cognitive == nil {
		http.Error(w, "Cognitive Matrix Offline", http.StatusServiceUnavailable)
		return
	}

	var req cognitive.InferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad JSON", http.StatusBadRequest)
		return
	}

	resp, err := s.Cognitive.Infer(req)
	if err != nil {
		log.Printf("Inference Failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, resp)
}

// GET /api/v1/cognitive/config
// Returns the current Cognitive Configuration (Profiles + Providers)
func (s *AdminServer) HandleCognitiveConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.Cognitive == nil || s.Cognitive.Config == nil {
		http.Error(w, "Cognitive Matrix Offline", http.StatusServiceUnavailable)
		return
	}

	// Return the raw config struct
	respondJSON(w, s.Cognitive.Config)
}

// POST /api/v1/chat
// Routes user messages exclusively through the Admin agent via NATS request-reply.
// The Admin agent has its full system prompt, tools, and council access.
// No raw LLM fallback — if the swarm is offline, the endpoint returns an error.
//
// The full conversation history is forwarded as JSON so the admin agent can
// maintain multi-turn context. The NATS payload is a JSON array of
// {role, content} objects; the agent's handleDirectRequest detects JSON arrays
// and reconstructs prior turns.
func (s *AdminServer) HandleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Messages []chatRequestMessage `json:"messages"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad JSON", http.StatusBadRequest)
		return
	}

	if len(req.Messages) == 0 {
		http.Error(w, "Empty conversation", http.StatusBadRequest)
		return
	}

	if availability := s.chatExecutionAvailability(); !availability.Available {
		respondAPIJSON(w, http.StatusServiceUnavailable, protocol.APIResponse{
			OK:    false,
			Error: availability.Summary,
			Data:  availability,
		})
		return
	}

	// 2. NATS must be available — the Admin agent is the ONLY path for chat.
	// No raw LLM fallback: agents must always operate within their context
	// (system prompt, tools, input/output rules).
	if s.NC == nil {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error":"Swarm offline — Admin agent unavailable. Start the organism first."}`, http.StatusServiceUnavailable)
		return
	}

	profile := userGovernanceProfileFromRequest(r)
	normalizedMessages, requestMutationTools := normalizeChatRequestMessages(req.Messages)
	normalizedMessages = applyGovernanceProfileToLatestMessage(normalizedMessages, profile)
	latestUserText := latestUserMessageContent(req.Messages)
	if len(normalizedMessages) > 0 {
		req.Messages = normalizedMessages
	}

	subject := fmt.Sprintf(protocol.TopicCouncilRequestFmt, "admin")
	agentResult, err := s.requestChatAgent(r.Context(), subject, req.Messages)
	if err != nil {
		log.Printf("Chat via Admin agent failed: %v", err)
		respondError(w, "Admin agent did not respond: "+err.Error(), http.StatusBadGateway)
		return
	}

	if shouldRetryDirectAnswer(agentResult, requestMutationTools) {
		retryMessages := applyDirectAnswerRetryInstruction(req.Messages, latestUserText)
		retryResult, retryErr := s.requestChatAgent(r.Context(), subject, retryMessages)
		if retryErr == nil {
			agentResult = retryResult
		}
	}

	if shouldRetryDirectAnswer(agentResult, requestMutationTools) {
		respondStructuredChatBlocker(w, directAnswerDriftBlocker(agentResult))
		return
	}

	isMutation, mutTools := mergeMutationTools(agentResult.ToolsUsed, requestMutationTools)
	if agentResult.Availability != nil && !agentResult.Availability.Available && (!isMutation || agentResult.Availability.Code != emptyProviderOutputCode) {
		respondStructuredChatBlocker(w, agentResult)
		return
	}
	if !isMutation && strings.TrimSpace(agentResult.Text) == "" && len(agentResult.Artifacts) == 0 {
		respondStructuredChatBlocker(w, agentResult)
		return
	}

	chatPayload := protocol.ChatResponsePayload{
		Text:          readableChatText(agentResult, isMutation),
		ToolsUsed:     mutTools,
		Artifacts:     agentResult.Artifacts,
		Consultations: agentResult.Consultations,
	}

	applyBrainProvenance(s, &chatPayload, agentResult)

	askContract := resolvePrimaryChatAskContract(isMutation)
	templateID := askContract.TemplateID
	mode := askContract.DefaultExecutionMode

	if isMutation {
		plannedToolCalls := buildPlannedToolCalls(agentResult, latestUserText, mutTools)
		approval := buildApprovalPolicy(profile, plannedToolCalls, mutTools)
		scope := &protocol.ScopeValidation{
			Tools:             mutTools,
			AffectedResources: affectedResourcesForPlannedCalls(plannedToolCalls),
			RiskLevel:         chatToolRisk(mutTools),
			PlannedToolCalls:  plannedToolCalls,
			Approval:          approval,
			GovernanceProfile: profile.snapshot(),
		}
		if approval != nil {
			scope.CapabilityIDs = approval.CapabilityIDs
			scope.ExternalDataUse = approval.ExternalDataUse
			scope.EstimatedCost = approval.EstimatedCost
		}

		auditEventID, _ := s.createAuditEvent(
			protocol.TemplateChatToProposal, "admin",
			"Chat mutation detected",
			map[string]any{
				"tools":           mutTools,
				"agent_tools":     agentResult.ToolsUsed,
				"requested_tools": requestMutationTools,
				"actor":           "Soma",
				"user":            auditUserLabelFromRequest(r),
				"ask_class":       string(askContract.AskClass),
				"action":          "proposal_generated",
				"result_status":   "pending",
				"approval_status": approvalStatusValue(approval),
				"approval_reason": approvalReasonValue(approval),
				"capability_used": strings.Join(scope.CapabilityIDs, ","),
			},
		)

		proof, _ := s.createIntentProof(protocol.TemplateChatToProposal, "chat-action", scope, auditEventID)
		var confirmToken *protocol.ConfirmToken
		if proof != nil {
			confirmToken, _ = s.generateConfirmToken(proof.ID, protocol.TemplateChatToProposal)
		}

		var proofID string
		var token string
		if proof != nil {
			proofID = proof.ID
		}
		if confirmToken != nil {
			token = confirmToken.Token
		}
		display := buildProposalDisplayContract(plannedToolCalls, latestUserText, mutTools)
		chatPayload.Proposal = buildMutationChatProposal(mutTools, proofID, token, "admin-core", []string{"admin"}, approval, profile.snapshot(), display)

		chatPayload.Provenance = &protocol.AnswerProvenance{
			ResolvedIntent:  "proposal",
			PermissionCheck: "pass",
			PolicyDecision:  policyDecisionForApproval(approval),
			AuditEventID:    auditEventID,
		}
	} else {
		auditEventID, _ := s.createAuditEvent(
			protocol.TemplateChatToAnswer, "admin",
			"Admin chat",
			map[string]any{
				"tools":         agentResult.ToolsUsed,
				"actor":         "Soma",
				"user":          auditUserLabelFromRequest(r),
				"ask_class":     string(askContract.AskClass),
				"action":        "answer_delivered",
				"result_status": "completed",
			},
		)
		chatPayload.Provenance = &protocol.AnswerProvenance{
			ResolvedIntent:  "answer",
			PermissionCheck: "pass",
			PolicyDecision:  "allow",
			AuditEventID:    auditEventID,
		}
		if len(agentResult.Artifacts) > 0 {
			for _, artifact := range agentResult.Artifacts {
				_, _ = s.createAuditEvent(
					protocol.TemplateChatToAnswer, "admin",
					"Chat artifact created",
					map[string]any{
						"actor":           "Soma",
						"user":            auditUserLabelFromRequest(r),
						"action":          "artifact_created",
						"result_status":   "completed",
						"capability_used": "artifact_output",
						"resource":        strings.TrimSpace(artifact.Title),
						"details":         map[string]any{"artifact_type": artifact.Type},
					},
				)
			}
		}
	}

	payloadBytes, _ := json.Marshal(chatPayload)

	envelope := protocol.CTSEnvelope{
		Meta: protocol.CTSMeta{
			SourceNode: "admin",
			Timestamp:  time.Now(),
		},
		SignalType: protocol.SignalChatResponse,
		TrustScore: protocol.TrustScoreCognitive,
		Payload:    payloadBytes,
		TemplateID: templateID,
		Mode:       mode,
	}

	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(envelope))
}

// ---------------------------------------------------------------------------
// Council Chat API — standardized, CTS-enveloped council interaction
// ---------------------------------------------------------------------------

// CouncilMemberInfo is returned by HandleListCouncilMembers.
type CouncilMemberInfo struct {
	ID   string `json:"id"`
	Role string `json:"role"`
	Team string `json:"team"`
}

// respondAPIJSON writes a protocol.APIResponse as JSON with an explicit status code.
func respondAPIJSON(w http.ResponseWriter, status int, resp protocol.APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

// respondAPIError writes a structured APIResponse error.
func respondAPIError(w http.ResponseWriter, msg string, status int) {
	respondAPIJSON(w, status, protocol.NewAPIError(msg))
}

// isCouncilMember checks whether memberID belongs to a standing council team
// (admin-core or council-core). Returns the team ID and role on match.
// Dynamic: add a new member to the YAML, restart, done.
func (s *AdminServer) isCouncilMember(memberID string) (teamID string, role string, ok bool) {
	if s.Soma == nil {
		return "", "", false
	}
	for _, tm := range s.Soma.ListTeams() {
		if tm.ID != "admin-core" && tm.ID != "council-core" {
			continue
		}
		for _, m := range tm.Members {
			if m.ID == memberID {
				return tm.ID, m.Role, true
			}
		}
	}
	return "", "", false
}

// GET /api/v1/council/members
// Returns all addressable council members from standing teams.
func (s *AdminServer) HandleListCouncilMembers(w http.ResponseWriter, r *http.Request) {
	if s.Soma == nil {
		respondAPIError(w, "Swarm offline", http.StatusServiceUnavailable)
		return
	}

	var members []CouncilMemberInfo
	for _, tm := range s.Soma.ListTeams() {
		if tm.ID != "admin-core" && tm.ID != "council-core" {
			continue
		}
		for _, m := range tm.Members {
			members = append(members, CouncilMemberInfo{
				ID:   m.ID,
				Role: m.Role,
				Team: tm.ID,
			})
		}
	}

	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(members))
}

// POST /api/v1/council/{member}/chat
// Routes user conversation to a specific council member via NATS request-reply.
// Returns a CTS envelope wrapped in APIResponse with trust score and provenance.
func (s *AdminServer) HandleCouncilChat(w http.ResponseWriter, r *http.Request) {
	memberID := r.PathValue("member")
	if memberID == "" {
		respondAPIError(w, "Missing council member ID", http.StatusBadRequest)
		return
	}

	// Validate member exists in standing council teams
	teamID, _, ok := s.isCouncilMember(memberID)
	if !ok {
		respondAPIError(w, fmt.Sprintf("Unknown council member: %s", memberID), http.StatusNotFound)
		return
	}

	var req struct {
		Messages []chatRequestMessage `json:"messages"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondAPIError(w, "Bad JSON", http.StatusBadRequest)
		return
	}
	if len(req.Messages) == 0 {
		respondAPIError(w, "Empty conversation", http.StatusBadRequest)
		return
	}

	if availability := s.chatExecutionAvailability(); !availability.Available {
		respondAPIJSON(w, http.StatusServiceUnavailable, protocol.APIResponse{
			OK:    false,
			Error: availability.Summary,
			Data:  availability,
		})
		return
	}

	// NATS must be available
	if s.NC == nil {
		respondAPIError(w, "Swarm offline — council agents unavailable. Start the organism first.", http.StatusServiceUnavailable)
		return
	}

	profile := userGovernanceProfileFromRequest(r)
	normalizedMessages, requestMutationTools := normalizeChatRequestMessages(req.Messages)
	normalizedMessages = applyGovernanceProfileToLatestMessage(normalizedMessages, profile)
	latestUserText := latestUserMessageContent(req.Messages)
	if len(normalizedMessages) > 0 {
		req.Messages = normalizedMessages
	}

	subject := fmt.Sprintf(protocol.TopicCouncilRequestFmt, memberID)
	agentResult, err := s.requestChatAgent(r.Context(), subject, req.Messages)
	if err != nil {
		log.Printf("Council chat with %s failed: %v", memberID, err)
		respondAPIError(w, fmt.Sprintf("Council member %s did not respond: %s", memberID, err.Error()), http.StatusBadGateway)
		return
	}

	if shouldRetryDirectAnswer(agentResult, requestMutationTools) {
		retryMessages := applyDirectAnswerRetryInstruction(req.Messages, latestUserText)
		retryResult, retryErr := s.requestChatAgent(r.Context(), subject, retryMessages)
		if retryErr == nil {
			agentResult = retryResult
		}
	}

	if shouldRetryDirectAnswer(agentResult, requestMutationTools) {
		respondStructuredChatBlocker(w, directAnswerDriftBlocker(agentResult))
		return
	}

	isMutation, mutTools := mergeMutationTools(agentResult.ToolsUsed, requestMutationTools)
	if agentResult.Availability != nil && !agentResult.Availability.Available && (!isMutation || agentResult.Availability.Code != emptyProviderOutputCode) {
		respondStructuredChatBlocker(w, agentResult)
		return
	}
	if !isMutation && strings.TrimSpace(agentResult.Text) == "" && len(agentResult.Artifacts) == 0 {
		respondStructuredChatBlocker(w, agentResult)
		return
	}

	// Wrap response in CTS envelope with trust score, provenance, and tool metadata
	chatPayload := protocol.ChatResponsePayload{
		Text:          readableChatText(agentResult, isMutation),
		ToolsUsed:     mutTools,
		Artifacts:     agentResult.Artifacts,
		Consultations: agentResult.Consultations,
	}

	applyBrainProvenance(s, &chatPayload, agentResult)

	askContract := resolvePrimaryChatAskContract(isMutation)
	templateID := askContract.TemplateID
	mode := askContract.DefaultExecutionMode

	if isMutation {
		plannedToolCalls := buildPlannedToolCalls(agentResult, latestUserText, mutTools)
		approval := buildApprovalPolicy(profile, plannedToolCalls, mutTools)
		scope := &protocol.ScopeValidation{
			Tools:             mutTools,
			AffectedResources: affectedResourcesForPlannedCalls(plannedToolCalls),
			RiskLevel:         chatToolRisk(mutTools),
			PlannedToolCalls:  plannedToolCalls,
			Approval:          approval,
			GovernanceProfile: profile.snapshot(),
		}
		if approval != nil {
			scope.CapabilityIDs = approval.CapabilityIDs
			scope.ExternalDataUse = approval.ExternalDataUse
			scope.EstimatedCost = approval.EstimatedCost
		}

		auditEventID, _ := s.createAuditEvent(
			protocol.TemplateChatToProposal, memberID,
			fmt.Sprintf("Council chat mutation detected from %s", memberID),
			map[string]any{
				"tools":           mutTools,
				"agent_tools":     agentResult.ToolsUsed,
				"requested_tools": requestMutationTools,
				"member":          memberID,
				"team":            teamID,
				"actor":           "Soma",
				"user":            auditUserLabelFromRequest(r),
				"ask_class":       string(askContract.AskClass),
				"action":          "proposal_generated",
				"result_status":   "pending",
				"approval_status": approvalStatusValue(approval),
				"approval_reason": approvalReasonValue(approval),
				"capability_used": strings.Join(scope.CapabilityIDs, ","),
			},
		)

		proof, _ := s.createIntentProof(protocol.TemplateChatToProposal, "chat-action", scope, auditEventID)
		var confirmToken *protocol.ConfirmToken
		if proof != nil {
			confirmToken, _ = s.generateConfirmToken(proof.ID, protocol.TemplateChatToProposal)
		}

		var proofID string
		var token string
		if proof != nil {
			proofID = proof.ID
		}
		if confirmToken != nil {
			token = confirmToken.Token
		}
		display := buildProposalDisplayContract(plannedToolCalls, latestUserText, mutTools)
		chatPayload.Proposal = buildMutationChatProposal(mutTools, proofID, token, teamID, []string{memberID}, approval, profile.snapshot(), display)

		chatPayload.Provenance = &protocol.AnswerProvenance{
			ResolvedIntent:  "proposal",
			PermissionCheck: "pass",
			PolicyDecision:  policyDecisionForApproval(approval),
			AuditEventID:    auditEventID,
		}
	} else {
		auditEventID, _ := s.createAuditEvent(
			protocol.TemplateChatToAnswer, memberID,
			fmt.Sprintf("Council chat with %s", memberID),
			map[string]any{
				"tools":         agentResult.ToolsUsed,
				"member":        memberID,
				"team":          teamID,
				"actor":         "Soma",
				"user":          auditUserLabelFromRequest(r),
				"ask_class":     string(askContract.AskClass),
				"action":        "answer_delivered",
				"result_status": "completed",
			},
		)
		chatPayload.Provenance = &protocol.AnswerProvenance{
			ResolvedIntent:  "answer",
			PermissionCheck: "pass",
			PolicyDecision:  "allow",
			AuditEventID:    auditEventID,
		}
		if len(agentResult.Artifacts) > 0 {
			for _, artifact := range agentResult.Artifacts {
				_, _ = s.createAuditEvent(
					protocol.TemplateChatToAnswer, memberID,
					fmt.Sprintf("Council artifact created by %s", memberID),
					map[string]any{
						"actor":           "Soma",
						"user":            auditUserLabelFromRequest(r),
						"action":          "artifact_created",
						"result_status":   "completed",
						"capability_used": "artifact_output",
						"resource":        strings.TrimSpace(artifact.Title),
						"details":         map[string]any{"artifact_type": artifact.Type, "member": memberID, "team": teamID},
					},
				)
			}
		}
	}

	payloadBytes, _ := json.Marshal(chatPayload)

	envelope := protocol.CTSEnvelope{
		Meta: protocol.CTSMeta{
			SourceNode: memberID,
			Timestamp:  time.Now(),
		},
		SignalType: protocol.SignalChatResponse,
		TrustScore: protocol.TrustScoreCognitive,
		Payload:    payloadBytes,
		TemplateID: templateID,
		Mode:       mode,
	}

	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(envelope))
	log.Printf("Council chat: member=%s team=%s trust=%.1f tools=%v template=%s", memberID, teamID, envelope.TrustScore, agentResult.ToolsUsed, envelope.TemplateID)
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

// PUT /api/v1/cognitive/profiles
// Updates which provider each cognitive profile uses.
// Persists to cognitive.yaml and updates the in-memory config.
func (s *AdminServer) HandleUpdateProfiles(w http.ResponseWriter, r *http.Request) {
	if s.Cognitive == nil || s.Cognitive.Config == nil {
		http.Error(w, "Cognitive Matrix Offline", http.StatusServiceUnavailable)
		return
	}

	var req struct {
		Profiles map[string]string `json:"profiles"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad JSON", http.StatusBadRequest)
		return
	}
	if len(req.Profiles) == 0 {
		http.Error(w, "No profiles provided", http.StatusBadRequest)
		return
	}

	// Validate: every profile value must reference an existing provider
	for profile, providerID := range req.Profiles {
		if _, ok := s.Cognitive.Config.Providers[providerID]; !ok {
			http.Error(w, fmt.Sprintf("Unknown provider '%s' for profile '%s'", providerID, profile), http.StatusBadRequest)
			return
		}
	}

	// Update in-memory config
	for profile, providerID := range req.Profiles {
		s.Cognitive.Config.Profiles[profile] = providerID
	}

	// Persist to YAML
	if err := s.Cognitive.SaveConfig(); err != nil {
		log.Printf("Failed to persist cognitive config: %v", err)
		http.Error(w, "Failed to save config", http.StatusInternalServerError)
		return
	}

	log.Printf("Cognitive profiles updated: %v", req.Profiles)
	respondJSON(w, s.Cognitive.Config)
}

// PUT /api/v1/cognitive/providers/{id}
// Updates a provider's configuration (endpoint, model_id, api_key_env).
// Reinitializes the adapter if the endpoint or type changes.
func (s *AdminServer) HandleUpdateProvider(w http.ResponseWriter, r *http.Request) {
	if s.Cognitive == nil || s.Cognitive.Config == nil {
		http.Error(w, "Cognitive Matrix Offline", http.StatusServiceUnavailable)
		return
	}

	providerID := r.PathValue("id")
	if providerID == "" {
		http.Error(w, "Missing provider ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Endpoint  string `json:"endpoint,omitempty"`
		ModelID   string `json:"model_id,omitempty"`
		APIKey    string `json:"api_key,omitempty"`     // Direct key (stored in-memory only, not persisted to YAML)
		APIKeyEnv string `json:"api_key_env,omitempty"` // Env var name (persisted to YAML)
		Type      string `json:"type,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad JSON", http.StatusBadRequest)
		return
	}

	existing, ok := s.Cognitive.Config.Providers[providerID]
	if !ok {
		// Create new provider
		existing = cognitive.ProviderConfig{}
	}

	if req.Type != "" {
		existing.Type = req.Type
	}
	if req.Endpoint != "" {
		existing.Endpoint = req.Endpoint
	}
	if req.ModelID != "" {
		existing.ModelID = req.ModelID
	}
	if req.APIKeyEnv != "" {
		existing.AuthKeyEnv = req.APIKeyEnv
	}
	if req.APIKey != "" {
		existing.AuthKey = req.APIKey
	}

	s.Cognitive.Config.Providers[providerID] = existing

	// Persist to YAML (AuthKey/AuthKeyEnv are json:"-" so won't leak)
	if err := s.Cognitive.SaveConfig(); err != nil {
		log.Printf("Failed to persist cognitive config: %v", err)
		http.Error(w, "Failed to save config", http.StatusInternalServerError)
		return
	}

	log.Printf("Provider '%s' updated: endpoint=%s model=%s", providerID, existing.Endpoint, existing.ModelID)

	// Return sanitized provider info (no secrets)
	respondJSON(w, map[string]any{
		"id":         providerID,
		"type":       existing.Type,
		"endpoint":   existing.Endpoint,
		"model_id":   existing.ModelID,
		"configured": existing.AuthKey != "" || existing.AuthKeyEnv != "",
	})
}

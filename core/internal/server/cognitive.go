package server

import (
	"regexp"
	"strings"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/pkg/protocol"
)

// mutationTools are tools that trigger proposal mode when used in chat.
// If an agent uses any of these, the response switches from answer → proposal.
var mutationTools = map[string]bool{
	"generate_blueprint":         true,
	"delegate":                   true,
	"write_file":                 true,
	"publish_signal":             true,
	"broadcast":                  true,
	"promote_deployment_context": true,
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
		if t == "promote_deployment_context" {
			return "high"
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

	teamCreationActions := []string{"create", "build", "launch", "instantiate", "manifest", "put together", "orchestrate"}
	teamCreationTargets := []string{"team", "teams", "specialist", "members", "lane", "lanes"}
	if requestContainsAny(lower, teamCreationActions) && requestContainsAny(lower, teamCreationTargets) {
		tools = append(tools, "generate_blueprint", "delegate")
	}

	mcpBindingActions := []string{"enable", "install", "connect", "associate", "configure", "assign"}
	mcpBindingTargets := []string{"mcp", "mcps", "tool", "tools", "web search", "github", "fetch", "browser", "host data", "shared-sources"}
	if requestContainsAny(lower, mcpBindingActions) && requestContainsAny(lower, mcpBindingTargets) {
		tools = append(tools, "delegate")
	}

	protectedBoundaryActions := []string{"use", "access", "connect", "enable", "store", "retain", "review", "apply", "configure"}
	protectedBoundaryTargets := []string{"private service", "internal service", "production service", "private api", "customer system", "client system", "api key", "token", "credential", "private data", "customer data", "deployment context", "company knowledge", "sensitive", "confidential"}
	if requestContainsAny(lower, protectedBoundaryActions) && requestContainsAny(lower, protectedBoundaryTargets) {
		tools = append(tools, "delegate")
	}

	recurringActions := []string{"store", "save", "persist", "make", "reuse", "apply"}
	recurringTargets := []string{"recurring", "standing behavior", "conversation template", "reusable template", "from now on", "every time"}
	if requestContainsAny(lower, recurringActions) && requestContainsAny(lower, recurringTargets) {
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

func isRuntimeStateQuestion(text string) bool {
	lower := strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(text)), " "))
	if lower == "" {
		return false
	}

	phrases := []string{
		"what is current state",
		"what is your current state",
		"what is the current state",
		"current org state",
		"organization state",
		"break down of org state",
		"breakdown of org state",
		"review current state",
		"summarize current state",
		"what teams currently exist",
		"what teams exist",
		"which teams exist",
		"list teams",
		"test connectivity",
		"connectivity status",
	}
	for _, phrase := range phrases {
		if strings.Contains(lower, phrase) {
			return true
		}
	}
	return false
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

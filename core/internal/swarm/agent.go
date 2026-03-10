package swarm

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/mycelis/core/pkg/pb/swarm"
)

// MCPToolExecutor resolves and invokes MCP tools by name.
// Implemented by composing mcp.Service (name lookup) and mcp.ClientPool (execution).
type MCPToolExecutor interface {
	FindToolByName(ctx context.Context, name string) (serverID uuid.UUID, toolName string, err error)
	CallTool(ctx context.Context, serverID uuid.UUID, toolName string, args map[string]any) (string, error)
}

// Agent represents a single node in a Swarm Team.
type Agent struct {
	Manifest       protocol.AgentManifest
	TeamID         string
	TeamInputs     []string // NATS subjects the team listens to
	TeamDeliveries []string // NATS subjects the team publishes to
	nc             *nats.Conn
	brain          *cognitive.Router
	toolExecutor   MCPToolExecutor
	toolDescs      map[string]string     // tool name → description for prompt injection
	internalTools  *InternalToolRegistry // live system state + context builder
	ctx            context.Context
	cancel         context.CancelFunc
	// V7 Event Spine: optional event emitter for tool audit trail.
	// Set by team.Start() propagating from Soma.ActivateBlueprint.
	// Nil = no emission (pre-V7 or degraded mode — silent, no panic).
	eventEmitter protocol.EventEmitter
	runID        string
	// V7 Conversation Log: optional full-fidelity turn logger.
	// Set by team.Start() propagating from Soma. Nil = silent mode.
	conversationLogger protocol.ConversationLogger
	sessionID          string // groups turns for a single request cycle
	turnIndex          int    // monotonic counter within session
	// V7 Interjection: buffered user redirect message.
	interjectionMu  sync.Mutex
	interjection    string
	interjectionSub *nats.Subscription
}

// NewAgent creates a new Agent instance with lifecycle context.
// toolExec may be nil if MCP tools are not available.
func NewAgent(ctx context.Context, manifest protocol.AgentManifest, teamID string, nc *nats.Conn, brain *cognitive.Router, toolExec MCPToolExecutor) *Agent {
	agentCtx, cancel := context.WithCancel(ctx)
	return &Agent{
		Manifest:     manifest,
		TeamID:       teamID,
		nc:           nc,
		brain:        brain,
		toolExecutor: toolExec,
		ctx:          agentCtx,
		cancel:       cancel,
	}
}

// SetToolDescriptions sets tool descriptions for system prompt injection.
// Called by Team after construction to provide the agent with tool metadata.
func (a *Agent) SetToolDescriptions(descs map[string]string) {
	a.toolDescs = descs
}

// SetInternalTools wires the internal tool registry for runtime context and tool execution.
func (a *Agent) SetInternalTools(tools *InternalToolRegistry) {
	a.internalTools = tools
}

// SetEventEmitter wires the V7 event emitter + run_id for tool audit trail emission.
// Called by team.Start() before the agent goroutine is launched.
func (a *Agent) SetEventEmitter(emitter protocol.EventEmitter, runID string) {
	a.eventEmitter = emitter
	a.runID = runID
}

// SetConversationLogger wires the V7 conversation logger for full-fidelity turn persistence.
// Called by team.Start() before the agent goroutine is launched.
func (a *Agent) SetConversationLogger(logger protocol.ConversationLogger) {
	a.conversationLogger = logger
}

// logTurn is a non-blocking helper that persists a conversation turn.
// Follows the same go-routine pattern as event emission.
func (a *Agent) logTurn(role, content, providerID, modelUsed, toolName string, toolArgs map[string]interface{}, parentTurnID, consultationOf string) {
	if a.conversationLogger == nil || a.sessionID == "" {
		return
	}
	idx := a.turnIndex
	a.turnIndex++
	go func() {
		_, err := a.conversationLogger.LogTurn(a.ctx, protocol.ConversationTurnData{
			RunID:          a.runID,
			SessionID:      a.sessionID,
			TenantID:       "default",
			AgentID:        a.Manifest.ID,
			TeamID:         a.TeamID,
			TurnIndex:      idx,
			Role:           role,
			Content:        content,
			ProviderID:     providerID,
			ModelUsed:      modelUsed,
			ToolName:       toolName,
			ToolArgs:       toolArgs,
			ParentTurnID:   parentTurnID,
			ConsultationOf: consultationOf,
		})
		if err != nil {
			log.Printf("[conversation] Agent [%s] log turn failed: %v", a.Manifest.ID, err)
		}
	}()
}

// checkInterjection returns and clears any buffered user interjection.
func (a *Agent) checkInterjection() string {
	a.interjectionMu.Lock()
	defer a.interjectionMu.Unlock()
	msg := a.interjection
	a.interjection = ""
	return msg
}

// subscribeInterjection sets up the NATS subscription for user interjections.
func (a *Agent) subscribeInterjection() {
	if a.nc == nil {
		return
	}
	subject := fmt.Sprintf(protocol.TopicAgentInterjectionFmt, a.Manifest.ID)
	sub, err := a.nc.Subscribe(subject, func(msg *nats.Msg) {
		a.interjectionMu.Lock()
		a.interjection = string(msg.Data)
		a.interjectionMu.Unlock()
		log.Printf("Agent [%s] received interjection: %s", a.Manifest.ID, truncateLog(string(msg.Data), 100))
	})
	if err != nil {
		log.Printf("Agent [%s] interjection subscribe failed: %v", a.Manifest.ID, err)
		return
	}
	a.interjectionSub = sub
}

// SetTeamTopology provides the agent with its team's NATS I/O contracts.
func (a *Agent) SetTeamTopology(inputs, deliveries []string) {
	a.TeamInputs = inputs
	a.TeamDeliveries = deliveries
}

// Start brings the Agent online to listen to its team's internal chatter.
func (a *Agent) Start() {
	// Listen to triggers on the internal team bus
	subject := fmt.Sprintf(protocol.TopicTeamInternalTrigger, a.TeamID)
	a.nc.Subscribe(subject, a.handleTrigger)
	log.Printf("Agent [%s] (%s) joined Team [%s]", a.Manifest.ID, a.Manifest.Role, a.TeamID)

	// Subscribe to personal request-reply topic (for council direct addressing)
	personalSubject := fmt.Sprintf(protocol.TopicCouncilRequestFmt, a.Manifest.ID)
	a.nc.Subscribe(personalSubject, a.handleDirectRequest)
	log.Printf("Agent [%s] listening for direct requests on [%s]", a.Manifest.ID, personalSubject)

	// V7: Subscribe to interjection topic for user redirects during active runs
	a.subscribeInterjection()

	// Start heartbeat goroutine
	go a.StartHeartbeat()
}

// StartHeartbeat publishes a periodic heartbeat on the global heartbeat topic.
func (a *Agent) StartHeartbeat() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-a.ctx.Done():
			log.Printf("Agent [%s] heartbeat stopped.", a.Manifest.ID)
			return
		case <-ticker.C:
			env := &pb.MsgEnvelope{
				Id:            uuid.New().String(),
				Timestamp:     timestamppb.Now(),
				SourceAgentId: a.Manifest.ID,
				Type:          pb.MessageType_MESSAGE_TYPE_EVENT,
				TeamId:        a.TeamID,
				Payload: &pb.MsgEnvelope_Event{
					Event: &pb.EventPayload{
						EventType: "agent.heartbeat",
					},
				},
			}

			data, err := proto.Marshal(env)
			if err != nil {
				log.Printf("Agent [%s] heartbeat marshal error: %v", a.Manifest.ID, err)
				continue
			}

			if err := a.nc.Publish(protocol.TopicGlobalHeartbeat, data); err != nil {
				log.Printf("Agent [%s] heartbeat publish error: %v", a.Manifest.ID, err)
			}
		}
	}
}

// Stop cancels the agent's context, stopping all goroutines.
func (a *Agent) Stop() {
	a.cancel()
}

// buildToolsBlock generates a tools description block for injection into the system prompt.
// Returns empty string if the agent has no tools.
func (a *Agent) buildToolsBlock() string {
	if len(a.Manifest.Tools) == 0 || len(a.toolDescs) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("\n\n## YOUR TOOLS (you MUST use these — never describe them to the user)\n")
	sb.WriteString("To call a tool, output ONLY this JSON (no markdown fences around it):\n")
	sb.WriteString(`{"tool_call": {"name": "TOOL_NAME", "arguments": {"key": "value"}}}`)
	sb.WriteString("\n\nThe system executes the tool and returns the result to you. ONE tool per response.\n")
	sb.WriteString("NEVER show tool_call JSON to the user as an example. NEVER say \"you can use\". Just call it.\n\n")
	seen := make(map[string]bool) // avoid duplicate tool entries
	for _, toolName := range a.Manifest.Tools {
		displayName := toolName
		// Handle mcp: prefixed tool names — strip prefix for display
		if strings.HasPrefix(toolName, "mcp:") {
			body := toolName[4:]
			if slash := strings.IndexByte(body, '/'); slash >= 0 {
				displayName = body[slash+1:]
			} else {
				displayName = body // "mcp:filesystem" → "filesystem"
			}
			// Wildcards: descriptions were already expanded into toolDescs by Team.Start()
			if displayName == "*" {
				continue
			}
		}
		// Skip toolset: references (expanded at construction time, not displayed)
		if strings.HasPrefix(toolName, "toolset:") {
			continue
		}
		if seen[displayName] {
			continue
		}
		if desc, ok := a.toolDescs[displayName]; ok {
			sb.WriteString(fmt.Sprintf("- **%s**: %s\n", displayName, desc))
			seen[displayName] = true
		} else if desc, ok := a.toolDescs[toolName]; ok {
			sb.WriteString(fmt.Sprintf("- **%s**: %s\n", toolName, desc))
			seen[toolName] = true
		}
	}
	return sb.String()
}

// ProcessResult holds the structured output of a processMessage call,
// including the final text and metadata about which tools were invoked.
type ProcessResult struct {
	Text      string                     `json:"text"`
	ToolsUsed []string                   `json:"tools_used,omitempty"`
	Artifacts []protocol.ChatArtifactRef `json:"artifacts,omitempty"`
	// Brain provenance: which provider/model executed this request.
	ProviderID string `json:"provider_id,omitempty"`
	ModelUsed  string `json:"model_used,omitempty"`
	// V7: Council consultations made during the ReAct loop (for frontend delegation trace)
	Consultations []protocol.ConsultationEntry `json:"consultations,omitempty"`
}

// processMessage handles LLM inference + ReAct tool loop, returning the response text.
// Shared by handleTrigger and handleDirectRequest.
// priorHistory is optional — if non-nil, earlier conversation turns are prepended
// after the system prompt so the agent has multi-turn context.
func (a *Agent) processMessage(input string, priorHistory []cognitive.ChatMessage) string {
	result := a.processMessageStructured(input, priorHistory)
	return result.Text
}

// processMessageStructured is the full-fidelity version that returns structured results.
func (a *Agent) processMessageStructured(input string, priorHistory []cognitive.ChatMessage) ProcessResult {
	if a.brain == nil {
		log.Printf("Agent [%s] has no brain. Skipping inference.", a.Manifest.ID)
		return ProcessResult{}
	}

	// V7 Conversation Log: generate a new session for this request cycle
	if a.conversationLogger != nil {
		a.sessionID = uuid.New().String()
		a.turnIndex = 0
	}

	// Build system prompt: static prompt + runtime context + tools block
	sys := a.Manifest.SystemPrompt
	if sys == "" {
		sys = fmt.Sprintf("You are a %s in the %s team.", a.Manifest.Role, a.TeamID)
	}

	// Inject live system state (active teams, MCP servers, NATS topology, cognitive config)
	if a.internalTools != nil {
		sys += a.internalTools.BuildContext(a.Manifest.ID, a.TeamID, a.Manifest.Role, a.TeamInputs, a.TeamDeliveries, input)
	}

	sys += a.buildToolsBlock()

	// Construct structured messages: system prompt + prior history + current input
	messages := []cognitive.ChatMessage{
		{Role: "system", Content: sys},
	}
	if len(priorHistory) > 0 {
		messages = append(messages, priorHistory...)
	}
	messages = append(messages, cognitive.ChatMessage{Role: "user", Content: input})

	// V7 Conversation Log: emit system prompt + user input
	a.logTurn("system", sys, "", "", "", nil, "", "")
	a.logTurn("user", input, "", "", "", nil, "", "")

	profile := "chat"
	if a.Manifest.Model != "" {
		profile = a.Manifest.Model
	}

	req := cognitive.InferRequest{
		Profile:  profile,
		Provider: a.Manifest.Provider,
		Messages: messages,
	}

	resp, err := a.brain.InferWithContract(a.ctx, req)
	if err != nil {
		log.Printf("Agent [%s] brain freeze: %v", a.Manifest.ID, err)
		return ProcessResult{}
	}

	// ReAct Tool Loop: if agent has tools bound and LLM returns a tool_call,
	// execute it and re-infer with the result. Track which tools were invoked.
	// Artifacts are captured from structured tool results and passed through
	// to the HTTP response (NOT fed back into the LLM context window).
	responseText := resp.Text
	var toolsUsed []string
	var artifacts []protocol.ChatArtifactRef
	var consultations []protocol.ConsultationEntry
	directAnswerPreferred := preferDirectDraftResponse(input)
	if a.toolExecutor != nil && len(a.Manifest.Tools) > 0 {
		reinferWithToolFeedback := func(toolName string, feedback string) bool {
			req.Messages = append(req.Messages, cognitive.ChatMessage{Role: "assistant", Content: responseText})
			req.Messages = append(req.Messages, cognitive.ChatMessage{
				Role:    "user",
				Content: fmt.Sprintf("Tool result from %s:\n%s\n\nContinue your response:", toolName, feedback),
			})
			updated, inferErr := a.brain.InferWithContract(a.ctx, req)
			if inferErr != nil {
				log.Printf("Agent [%s] re-inference after tool feedback failed: %v", a.Manifest.ID, inferErr)
				responseText = feedback
				return false
			}
			resp = updated
			responseText = updated.Text
			return true
		}
		preflightDone := map[string]bool{}
		failedToolCalls := map[string]int{}

		// Guardrail: if the model returns a "plan/step" style response instead of
		// a tool call for an actionable request, force one correction pass before
		// entering the normal tool loop.
		if parseToolCall(responseText) == nil && responseSuggestsUnexecutedAction(responseText) {
			correction := cognitive.ChatMessage{
				Role: "system",
				Content: "Policy correction: do not provide step-by-step plans when tools are available. " +
					"Emit exactly one tool_call JSON now for the user's actionable request, or return a concrete blocker.",
			}
			req.Messages = append(req.Messages, correction)
			req.Messages = append(req.Messages, cognitive.ChatMessage{
				Role:    "user",
				Content: "Re-answer the latest request now under the policy correction.",
			})
			repaired, repairErr := a.brain.InferWithContract(a.ctx, req)
			if repairErr != nil {
				log.Printf("Agent [%s] policy correction re-inference failed: %v", a.Manifest.ID, repairErr)
			} else if repaired != nil {
				resp = repaired
				responseText = repaired.Text
			}
		}

		maxIter := a.Manifest.EffectiveMaxIterations()
		for i := 0; i < maxIter; i++ {
			// V7 Interjection: check for user redirect between ReAct iterations
			if interjection := a.checkInterjection(); interjection != "" {
				messages = append(messages, cognitive.ChatMessage{
					Role:    "user",
					Content: "[OPERATOR INTERJECTION]: " + interjection,
				})
				req.Messages = messages
				a.logTurn("interjection", interjection, "", "", "", nil, "", "")
				log.Printf("Agent [%s] processing interjection: %s", a.Manifest.ID, truncateLog(interjection, 100))
				// Re-infer with interjection injected
				resp, err = a.brain.InferWithContract(a.ctx, req)
				if err != nil {
					log.Printf("Agent [%s] interjection re-inference failed: %v", a.Manifest.ID, err)
					break
				}
				responseText = resp.Text
			}

			toolCall := parseToolCall(responseText)
			if toolCall == nil {
				break
			}
			autofillToolArguments(toolCall, input)
			// Simple drafting requests should stay in chat instead of bouncing through tools.
			if directAnswerPreferred && shouldAvoidToolsForDirectDraft(toolCall.Name) {
				if !reinferWithToolFeedback(toolCall.Name,
					"Policy correction: the user asked for text content in this chat. Respond with the requested content directly. Do not call tools unless they explicitly asked to read or write files, save output, inspect runtime state, execute commands, or route work to other teams.") {
					break
				}
				continue
			}
			fingerprint := toolCallFingerprint(toolCall)
			if failedToolCalls[fingerprint] >= 2 {
				if !reinferWithToolFeedback(toolCall.Name,
					fmt.Sprintf("Policy correction: the exact tool call %s has already failed %d times in this turn. Do not retry it. Choose a different tool or answer directly without tools.",
						fingerprint, failedToolCalls[fingerprint])) {
					break
				}
				continue
			}
			if shouldCouncilPreflight(toolCall.Name) {
				member := councilPreflightMember(toolCall.Name)
				if member != "" && !preflightDone[member] {
					preflightDone[member] = true
					summary, err := a.runCouncilPreflight(member, input, toolCall)
					if err != nil {
						log.Printf("Agent [%s] council preflight failed for %s: %v", a.Manifest.ID, toolCall.Name, err)
					} else if strings.TrimSpace(summary) != "" {
						short := strings.TrimSpace(summary)
						if len(short) > 300 {
							short = short[:300] + "..."
						}
						consultations = append(consultations, protocol.ConsultationEntry{
							Member:  member,
							Summary: short,
						})
						// Feed the preflight back so the model can refine the next tool call.
						if !reinferWithToolFeedback("consult_council", fmt.Sprintf("Preflight (%s): %s", member, summary)) {
							break
						}
						continue
					}
				}
			}

			log.Printf("Agent [%s] tool_call [%d/%d]: %s", a.Manifest.ID, i+1, maxIter, toolCall.Name)
			toolsUsed = append(toolsUsed, toolCall.Name)

			// V7: emit tool.invoked before execution (best-effort, non-blocking)
			if a.eventEmitter != nil && a.runID != "" {
				go a.eventEmitter.Emit(a.ctx, a.runID, //nolint:errcheck
					protocol.EventToolInvoked, protocol.SeverityInfo,
					a.Manifest.ID, a.TeamID,
					map[string]interface{}{"tool": toolCall.Name, "iteration": i + 1})
			}

			// V7 Conversation Log: emit tool_call
			toolCallTurnID := ""
			if a.conversationLogger != nil {
				// Capture the turn index for this tool_call so tool_result can reference it
				toolCallTurnID = fmt.Sprintf("%s-%d", a.sessionID, a.turnIndex)
			}
			a.logTurn("tool_call", responseText, "", "", toolCall.Name, toolCall.Arguments, "", "")

			serverID, _, err := a.toolExecutor.FindToolByName(a.ctx, toolCall.Name)
			if err != nil {
				failedToolCalls[fingerprint]++
				log.Printf("Agent [%s] tool lookup failed: %v", a.Manifest.ID, err)
				// V7: emit tool.failed
				if a.eventEmitter != nil && a.runID != "" {
					go a.eventEmitter.Emit(a.ctx, a.runID, //nolint:errcheck
						protocol.EventToolFailed, protocol.SeverityError,
						a.Manifest.ID, a.TeamID,
						map[string]interface{}{"tool": toolCall.Name, "error": err.Error(), "phase": "lookup"})
				}
				if !reinferWithToolFeedback(toolCall.Name, fmt.Sprintf("Tool '%s' is not available: %v", toolCall.Name, err)) {
					break
				}
				continue
			}

			toolResult, err := a.toolExecutor.CallTool(a.ctx, serverID, toolCall.Name, toolCall.Arguments)
			if err != nil {
				failedToolCalls[fingerprint]++
				log.Printf("Agent [%s] tool call failed: %v", a.Manifest.ID, err)
				// V7: emit tool.failed
				if a.eventEmitter != nil && a.runID != "" {
					go a.eventEmitter.Emit(a.ctx, a.runID, //nolint:errcheck
						protocol.EventToolFailed, protocol.SeverityError,
						a.Manifest.ID, a.TeamID,
						map[string]interface{}{"tool": toolCall.Name, "error": err.Error(), "phase": "execute"})
				}
				if !reinferWithToolFeedback(toolCall.Name, fmt.Sprintf("Tool %s failed: %v", toolCall.Name, err)) {
					break
				}
				continue
			}
			// V7: emit tool.completed
			if a.eventEmitter != nil && a.runID != "" {
				go a.eventEmitter.Emit(a.ctx, a.runID, //nolint:errcheck
					protocol.EventToolCompleted, protocol.SeverityInfo,
					a.Manifest.ID, a.TeamID,
					map[string]interface{}{"tool": toolCall.Name, "iteration": i + 1})
			}

			// Capture consult_council calls for frontend delegation trace
			consultMember := ""
			if toolCall.Name == "consult_council" {
				member, _ := toolCall.Arguments["member"].(string)
				if member != "" {
					consultMember = member
					summary := strings.TrimSpace(toolResult)
					if len(summary) > 300 {
						summary = summary[:300] + "..."
					}
					consultations = append(consultations, protocol.ConsultationEntry{
						Member:  member,
						Summary: summary,
					})
				}
			}

			// V7 Conversation Log: emit tool_result (full content, not truncated)
			a.logTurn("tool_result", toolResult, "", "", toolCall.Name, nil, toolCallTurnID, consultMember)

			// Extract artifact from structured tool result (side-channel).
			// Only the text message goes back to the LLM — large payloads
			// like base64 images are captured here for the HTTP response.
			var toolOutput struct {
				Message  string                    `json:"message"`
				Artifact *protocol.ChatArtifactRef `json:"artifact"`
			}
			if json.Unmarshal([]byte(toolResult), &toolOutput) == nil && toolOutput.Artifact != nil {
				artifacts = append(artifacts, *toolOutput.Artifact)
				toolResult = toolOutput.Message
				if toolResult == "" {
					toolResult = fmt.Sprintf("Tool %s completed successfully.", toolCall.Name)
				}
			}

			// Append Assistant's tool call and User's tool result to history
			req.Messages = append(req.Messages, cognitive.ChatMessage{Role: "assistant", Content: responseText})
			req.Messages = append(req.Messages, cognitive.ChatMessage{Role: "user", Content: fmt.Sprintf("Tool result from %s:\n%s\n\nContinue your response:", toolCall.Name, toolResult)})

			resp, err = a.brain.InferWithContract(a.ctx, req)
			if err != nil {
				log.Printf("Agent [%s] re-inference failed: %v", a.Manifest.ID, err)
				break
			}
			responseText = resp.Text
		}
	}

	// Sanitize: strip any residual tool_call JSON from the response.
	// Small LLMs sometimes echo tool_call JSON as text even after execution,
	// or the loop may break early leaving raw JSON in the response.
	responseText = stripToolCallJSON(responseText)

	// Auto-summarize: every 15 messages, compress conversation into pgvector.
	// Non-blocking background goroutine — failures are logged, never fatal.
	if a.internalTools != nil && len(priorHistory) > 0 && len(priorHistory)%15 == 0 {
		histCopy := make([]cognitive.ChatMessage, len(priorHistory))
		copy(histCopy, priorHistory)
		go a.internalTools.AutoSummarize(a.ctx, a.Manifest.ID, histCopy)
	}

	// Capture brain provenance from the last inference response
	var providerID, modelUsed string
	if resp != nil {
		providerID = resp.Provider
		modelUsed = resp.ModelUsed
	}

	// V7 Conversation Log: emit final assistant response with brain provenance
	a.logTurn("assistant", responseText, providerID, modelUsed, "", nil, "", "")

	return ProcessResult{
		Text:          responseText,
		ToolsUsed:     toolsUsed,
		Artifacts:     artifacts,
		ProviderID:    providerID,
		ModelUsed:     modelUsed,
		Consultations: consultations,
	}
}

// stripToolCallJSON removes tool_call JSON blocks from response text.
// Returns the text before the JSON block, or a fallback if only JSON remains.
func stripToolCallJSON(text string) string {
	keyword := `"tool_call"`
	idx := strings.Index(text, keyword)
	if idx == -1 {
		return text
	}

	// Walk backwards from "tool_call" to find the opening brace
	start := -1
	for i := idx - 1; i >= 0; i-- {
		if text[i] == '{' {
			start = i
			break
		}
	}
	if start == -1 {
		return text
	}

	// Return the text before the JSON block, trimmed
	cleaned := strings.TrimSpace(text[:start])
	if cleaned != "" {
		return cleaned
	}
	return text // Don't return empty — keep original if nothing before the JSON
}

func (a *Agent) handleTrigger(msg *nats.Msg) {
	select {
	case <-a.ctx.Done():
		if msg.Reply != "" {
			msg.Respond([]byte("Agent shutting down."))
		}
		return
	default:
	}

	log.Printf("Agent [%s] thinking about: %s", a.Manifest.ID, string(msg.Data))

	responseText := a.processMessage(string(msg.Data), nil)
	if responseText == "" {
		// Still respond if caller is waiting (request-reply pattern)
		if msg.Reply != "" {
			msg.Respond([]byte(fmt.Sprintf("[%s] No response — LLM may be unavailable.", a.Manifest.ID)))
		}
		return
	}

	// If msg has a reply subject (request-reply pattern), respond directly
	if msg.Reply != "" {
		msg.Respond([]byte(responseText))
	}

	// Also publish to team respond bus (existing behavior)
	replySubject := fmt.Sprintf(protocol.TopicTeamInternalRespond, a.TeamID)
	a.nc.Publish(replySubject, []byte(responseText))
	log.Printf("Agent [%s] replied.", a.Manifest.ID)
}

// handleDirectRequest handles NATS request-reply for personal/council addressing.
// Always responds via msg.Respond() since this is always a request-reply.
//
// The payload may be either:
//   - Plain text (from consult_council, delegate_task, etc.)
//   - A JSON array of {role, content} objects (from HandleChat — full conversation history)
//
// When a JSON array is detected, prior turns are extracted and passed to processMessage
// so the agent maintains multi-turn conversational context.
//
// The response is JSON-encoded ProcessResult so the HTTP handler can extract
// tools_used metadata. Plain-text callers (consult_council) can still read .text.
func (a *Agent) handleDirectRequest(msg *nats.Msg) {
	select {
	case <-a.ctx.Done():
		return
	default:
	}

	data := msg.Data
	input, history := a.parseConversationPayload(data)

	log.Printf("Agent [%s] direct request (%d prior turns): %s", a.Manifest.ID, len(history), truncateLog(input, 200))

	result := a.processMessageStructured(input, history)
	if result.Text == "" {
		if msg.Reply != "" {
			msg.Respond([]byte("Agent unavailable — no cognitive engine."))
		}
		return
	}

	// Return structured JSON so callers can extract tools_used metadata
	if msg.Reply != "" {
		respBytes, err := json.Marshal(result)
		if err != nil {
			msg.Respond([]byte(result.Text))
		} else {
			msg.Respond(respBytes)
		}
	}
	log.Printf("Agent [%s] direct request replied (tools: %v).", a.Manifest.ID, result.ToolsUsed)
}

// parseConversationPayload detects whether the NATS payload is a JSON conversation
// array or plain text. Returns the last user message as input and any prior turns
// as ChatMessage history.
func (a *Agent) parseConversationPayload(data []byte) (string, []cognitive.ChatMessage) {
	// Quick check: does it look like a JSON array?
	trimmed := strings.TrimSpace(string(data))
	if len(trimmed) == 0 || trimmed[0] != '[' {
		return string(data), nil
	}

	// Try to parse as conversation array
	type chatTurn struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	var turns []chatTurn
	if err := json.Unmarshal(data, &turns); err != nil {
		// Not valid JSON array — treat as plain text
		return string(data), nil
	}

	if len(turns) == 0 {
		return "", nil
	}

	// Last turn is the current input; everything before is history
	last := turns[len(turns)-1]
	if len(turns) == 1 {
		return last.Content, nil
	}

	// Build prior history: map roles to LLM-compatible roles
	history := make([]cognitive.ChatMessage, 0, len(turns)-1)
	for _, t := range turns[:len(turns)-1] {
		role := t.Role
		switch role {
		case "admin", "architect", "assistant":
			role = "assistant"
		case "user":
			// keep as-is
		default:
			role = "user" // unknown roles treated as user context
		}
		history = append(history, cognitive.ChatMessage{Role: role, Content: t.Content})
	}

	return last.Content, history
}

// truncateLog shortens a string for log output.
func truncateLog(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// responseSuggestsUnexecutedAction detects common "I will/Step 1" patterns where
// the model narrates an action path instead of invoking tools.
func responseSuggestsUnexecutedAction(text string) bool {
	lower := strings.ToLower(strings.TrimSpace(text))
	if lower == "" {
		return false
	}
	patterns := []string{
		"we need to delegate",
		"let's proceed by",
		"step 1",
		"you can use",
		"would you like me to",
		"to have the",
		"example input",
		"this will route your request",
		"this will delegate",
		"this will consult",
		"i'll consult",
		"i will consult",
		"i'll delegate",
		"i will delegate",
		"route your request",
		"i've delegated",
		"i have delegated",
		"task has been delegated",
	}
	for _, p := range patterns {
		if strings.Contains(lower, p) {
			return true
		}
	}
	return false
}

// toolCallPayload represents a tool invocation extracted from LLM output.
type toolCallPayload struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

// parseToolCall extracts a tool_call JSON block from LLM response text.
// Handles both compact and pretty-printed JSON from LLMs:
//
//	Compact:  {"tool_call": {"name": "read_file", "arguments": {...}}}
//	Pretty:   {\n  "tool_call": {\n    "name": "read_file" ...
//	Fenced:   ```json\n{"tool_call": ...}\n```
//
// Returns nil if no tool call is found.
func parseToolCall(text string) *toolCallPayload {
	keyword := `"tool_call"`
	idx := strings.Index(text, keyword)
	if idx == -1 {
		return parseOperationCall(text)
	}

	// Walk backwards from "tool_call" to find the opening brace.
	// LLMs may emit whitespace/newlines between { and "tool_call".
	start := -1
	for i := idx - 1; i >= 0; i-- {
		ch := text[i]
		if ch == '{' {
			start = i
			break
		}
		if ch != ' ' && ch != '\t' && ch != '\n' && ch != '\r' {
			break // unexpected character before "tool_call"
		}
	}
	if start == -1 {
		return nil
	}

	// Find the matching closing brace, respecting JSON string escaping.
	depth := 0
	end := -1
	inStr := false
	esc := false
	for i := start; i < len(text); i++ {
		ch := text[i]
		if esc {
			esc = false
			continue
		}
		if ch == '\\' && inStr {
			esc = true
			continue
		}
		if ch == '"' {
			inStr = !inStr
			continue
		}
		if inStr {
			continue
		}
		switch ch {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				end = i + 1
			}
		}
		if end != -1 {
			break
		}
	}
	if end == -1 {
		if loose := parseLooseToolCall(text[start:]); loose != nil {
			return loose
		}
		return nil
	}

	var wrapper struct {
		ToolCall toolCallPayload `json:"tool_call"`
	}
	if err := json.Unmarshal([]byte(text[start:end]), &wrapper); err != nil {
		log.Printf("[parseToolCall] JSON unmarshal failed: %v (excerpt: %s)", err, truncateLog(text[start:end], 200))
		if loose := parseLooseToolCall(text[start:end]); loose != nil {
			return loose
		}
		return nil
	}
	if wrapper.ToolCall.Name == "" {
		return parseOperationCall(text)
	}
	return &wrapper.ToolCall
}

func parseLooseToolCall(text string) *toolCallPayload {
	nameRe := regexp.MustCompile(`"name"\s*:\s*"([^"]+)"`)
	m := nameRe.FindStringSubmatch(text)
	if len(m) < 2 {
		return nil
	}
	name := strings.TrimSpace(m[1])
	if name == "" {
		return nil
	}
	return &toolCallPayload{Name: name, Arguments: map[string]any{}}
}

// parseOperationCall extracts fallback operation-style payloads emitted by some models:
// {"operation":"consult_council","arguments":{...}}
// This lets the runtime execute the intended tool instead of returning instructions.
func parseOperationCall(text string) *toolCallPayload {
	keyword := `"operation"`
	idx := strings.Index(text, keyword)
	if idx == -1 {
		return nil
	}

	start := -1
	for i := idx - 1; i >= 0; i-- {
		ch := text[i]
		if ch == '{' {
			start = i
			break
		}
		if ch != ' ' && ch != '\t' && ch != '\n' && ch != '\r' {
			break
		}
	}
	if start == -1 {
		return nil
	}

	depth := 0
	end := -1
	inStr := false
	esc := false
	for i := start; i < len(text); i++ {
		ch := text[i]
		if esc {
			esc = false
			continue
		}
		if ch == '\\' && inStr {
			esc = true
			continue
		}
		if ch == '"' {
			inStr = !inStr
			continue
		}
		if inStr {
			continue
		}
		switch ch {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				end = i + 1
			}
		}
		if end != -1 {
			break
		}
	}
	if end == -1 {
		return nil
	}

	var payload struct {
		Operation string         `json:"operation"`
		Arguments map[string]any `json:"arguments"`
	}
	if err := json.Unmarshal([]byte(text[start:end]), &payload); err != nil {
		return nil
	}
	if strings.TrimSpace(payload.Operation) == "" {
		return nil
	}
	if payload.Arguments == nil {
		payload.Arguments = map[string]any{}
	}
	return &toolCallPayload{Name: payload.Operation, Arguments: payload.Arguments}
}

// autofillToolArguments patches common missing fields for tool calls that are
// clearly actionable but slightly malformed from smaller/local model outputs.
func autofillToolArguments(call *toolCallPayload, latestUserInput string) {
	if call == nil {
		return
	}
	if call.Arguments == nil {
		call.Arguments = map[string]any{}
	}

	switch call.Name {
	case "consult_council":
		member, _ := call.Arguments["member"].(string)
		if strings.TrimSpace(member) == "" {
			member = inferCouncilMemberFromInput(latestUserInput)
		}
		if normalized := normalizeCouncilMember(member); normalized != "" {
			call.Arguments["member"] = normalized
		}
		question, _ := call.Arguments["question"].(string)
		if strings.TrimSpace(question) == "" {
			if q := strings.TrimSpace(latestUserInput); q != "" {
				call.Arguments["question"] = q
			}
		}
	case "read_signals":
		subject, _ := call.Arguments["subject"].(string)
		if strings.TrimSpace(subject) == "" {
			if v, _ := call.Arguments["topic_pattern"].(string); strings.TrimSpace(v) != "" {
				call.Arguments["subject"] = strings.TrimSpace(v)
			} else if v, _ := call.Arguments["topic"].(string); strings.TrimSpace(v) != "" {
				call.Arguments["subject"] = strings.TrimSpace(v)
			} else if v, _ := call.Arguments["channel"].(string); strings.TrimSpace(v) != "" {
				call.Arguments["subject"] = strings.TrimSpace(v)
			} else if v := extractNATSSubject(latestUserInput); v != "" {
				call.Arguments["subject"] = v
			}
		}
	case "delegate_task":
		if _, hasTask := call.Arguments["task"]; !hasTask {
			if teamName, _ := call.Arguments["team_name"].(string); strings.TrimSpace(teamName) != "" {
				call.Name = "create_team"
				call.Arguments["team_id"] = strings.TrimSpace(teamName)
				if role, _ := call.Arguments["agent_type"].(string); strings.TrimSpace(role) != "" {
					call.Arguments["role"] = strings.TrimSpace(role)
				}
			}
		}
	}
}

func extractNATSSubject(input string) string {
	for _, tok := range strings.Fields(input) {
		clean := strings.Trim(tok, " \t\r\n,.;:()[]{}<>\"'")
		if clean == "" {
			continue
		}
		if strings.HasPrefix(clean, "swarm.") {
			return clean
		}
	}
	return ""
}

func normalizeCouncilMember(member string) string {
	m := strings.ToLower(strings.TrimSpace(member))
	if m == "" {
		return ""
	}
	switch m {
	case "architect", "council architect", "council-architect":
		return "council-architect"
	case "coder", "council coder", "council-coder":
		return "council-coder"
	case "creative", "council creative", "council-creative":
		return "council-creative"
	case "sentry", "council sentry", "council-sentry":
		return "council-sentry"
	default:
		if strings.HasPrefix(m, "council-") {
			return m
		}
		return member
	}
}

func inferCouncilMemberFromInput(input string) string {
	lower := strings.ToLower(input)
	switch {
	case strings.Contains(lower, "architect"):
		return "council-architect"
	case strings.Contains(lower, "coder"), strings.Contains(lower, "code"):
		return "council-coder"
	case strings.Contains(lower, "creative"), strings.Contains(lower, "design"), strings.Contains(lower, "image"):
		return "council-creative"
	case strings.Contains(lower, "sentry"), strings.Contains(lower, "security"), strings.Contains(lower, "risk"):
		return "council-sentry"
	default:
		return ""
	}
}

func shouldCouncilPreflight(toolName string) bool {
	switch strings.TrimSpace(toolName) {
	case "create_team", "delegate_task", "local_command":
		return true
	default:
		return false
	}
}

func councilPreflightMember(toolName string) string {
	switch strings.TrimSpace(toolName) {
	case "create_team", "delegate_task":
		return "council-architect"
	case "local_command":
		return "council-coder"
	default:
		return ""
	}
}

func formatCouncilPreflightQuestion(userInput string, call *toolCallPayload) string {
	if call == nil {
		return userInput
	}
	args := "{}"
	if call.Arguments != nil {
		if b, err := json.Marshal(call.Arguments); err == nil {
			args = string(b)
		}
	}
	return fmt.Sprintf("Preflight review before executing tool '%s' with args %s. User request: %s. Provide concise execution guidance and edge cases.", call.Name, args, strings.TrimSpace(userInput))
}

func (a *Agent) runCouncilPreflight(userMember string, userInput string, call *toolCallPayload) (string, error) {
	if a == nil || a.toolExecutor == nil {
		return "", fmt.Errorf("tool executor unavailable")
	}
	serverID, toolName, err := a.toolExecutor.FindToolByName(a.ctx, "consult_council")
	if err != nil {
		return "", err
	}
	preflightCtx, cancel := context.WithTimeout(a.ctx, 10*time.Second)
	defer cancel()
	args := map[string]any{
		"member":   userMember,
		"question": formatCouncilPreflightQuestion(userInput, call),
	}
	return a.toolExecutor.CallTool(preflightCtx, serverID, toolName, args)
}

func toolCallFingerprint(call *toolCallPayload) string {
	if call == nil {
		return ""
	}
	args := "{}"
	if call.Arguments != nil {
		if b, err := json.Marshal(call.Arguments); err == nil {
			args = string(b)
		}
	}
	return fmt.Sprintf("%s:%s", strings.TrimSpace(call.Name), args)
}

func preferDirectDraftResponse(input string) bool {
	lower := strings.ToLower(strings.TrimSpace(input))
	if lower == "" {
		return false
	}

	explicitAction := []string{
		"file", "path", "workspace", "folder", "directory", "save", "persist", "store",
		"read ", "open ", "inspect", "command", "run ", "execute", "shell", "terminal",
		"team", "agent", "council", "signal", "nats", "api", "http", "url", "image",
		"diagram", "code", "blueprint", "mission", "workflow",
	}
	for _, token := range explicitAction {
		if strings.Contains(lower, token) {
			return false
		}
	}

	requestMarkers := []string{
		"write ", "draft ", "compose ", "create ", "make ", "generate ",
	}
	hasRequestMarker := false
	for _, marker := range requestMarkers {
		if strings.Contains(lower, marker) {
			hasRequestMarker = true
			break
		}
	}
	if !hasRequestMarker {
		return false
	}

	textOutputs := []string{
		"letter", "email", "message", "note", "reply", "paragraph",
		"announcement", "bio", "summary", "caption", "introduction",
	}
	for _, target := range textOutputs {
		if strings.Contains(lower, target) {
			return true
		}
	}
	return false
}

func shouldAvoidToolsForDirectDraft(toolName string) bool {
	switch strings.TrimSpace(toolName) {
	case "write_file", "read_file", "local_command", "consult_council", "delegate_task",
		"research_for_blueprint", "generate_blueprint", "remember", "recall",
		"store_artifact", "list_teams", "list_missions", "get_system_status",
		"list_available_tools", "list_catalogue", "publish_signal", "read_signals",
		"broadcast", "create_team":
		return true
	default:
		return false
	}
}

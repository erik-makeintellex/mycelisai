package swarm

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
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
	Manifest     protocol.AgentManifest
	TeamID       string
	nc           *nats.Conn
	brain        *cognitive.Router
	toolExecutor MCPToolExecutor
	ctx          context.Context
	cancel       context.CancelFunc
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

// Start brings the Agent online to listen to its team's internal chatter.
func (a *Agent) Start() {
	// Listen to triggers on the internal team bus
	subject := fmt.Sprintf(protocol.TopicTeamInternalTrigger, a.TeamID)
	a.nc.Subscribe(subject, a.handleTrigger)
	log.Printf("Agent [%s] (%s) joined Team [%s]", a.Manifest.ID, a.Manifest.Role, a.TeamID)

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

func (a *Agent) handleTrigger(msg *nats.Msg) {
	// Respect context cancellation
	select {
	case <-a.ctx.Done():
		return
	default:
	}

	log.Printf("Agent [%s] thinking about: %s", a.Manifest.ID, string(msg.Data))

	if a.brain == nil {
		log.Printf("Agent [%s] has no brain. Skipping inference.", a.Manifest.ID)
		return
	}

	// 1. Construct Prompt from Role + Input
	// Use Custom System Prompt if available, else default to Role-based
	sys := a.Manifest.SystemPrompt
	if sys == "" {
		sys = fmt.Sprintf("You are a %s in the %s team.", a.Manifest.Role, a.TeamID)
	}
	prompt := fmt.Sprintf("%s Input: %s", sys, string(msg.Data))

	// 2. Infer
	// Use Model override if specified, else "chat" profile
	profile := "chat" // default
	if a.Manifest.Model != "" {
		profile = a.Manifest.Model
	}

	req := cognitive.InferRequest{
		Profile: profile,
		Prompt:  prompt,
	}

	resp, err := a.brain.InferWithContract(a.ctx, req)
	if err != nil {
		log.Printf("Agent [%s] brain freeze: %v", a.Manifest.ID, err)
		return
	}

	// 3. ReAct Tool Loop: if agent has tools bound and LLM returns a tool_call,
	// execute it and re-infer with the result (max 3 iterations).
	responseText := resp.Text
	if a.toolExecutor != nil && len(a.Manifest.Tools) > 0 {
		for i := 0; i < 3; i++ {
			toolCall := parseToolCall(responseText)
			if toolCall == nil {
				break
			}

			log.Printf("Agent [%s] tool_call: %s", a.Manifest.ID, toolCall.Name)

			serverID, _, err := a.toolExecutor.FindToolByName(a.ctx, toolCall.Name)
			if err != nil {
				log.Printf("Agent [%s] tool lookup failed: %v", a.Manifest.ID, err)
				break
			}

			toolResult, err := a.toolExecutor.CallTool(a.ctx, serverID, toolCall.Name, toolCall.Arguments)
			if err != nil {
				log.Printf("Agent [%s] tool call failed: %v", a.Manifest.ID, err)
				responseText = fmt.Sprintf("Tool %s failed: %v", toolCall.Name, err)
				break
			}

			// Re-infer with tool result appended to context
			prompt = fmt.Sprintf("%s\n\nTool result from %s:\n%s\n\nContinue your response:", prompt, toolCall.Name, toolResult)
			req.Prompt = prompt
			resp, err = a.brain.InferWithContract(a.ctx, req)
			if err != nil {
				log.Printf("Agent [%s] re-inference failed: %v", a.Manifest.ID, err)
				break
			}
			responseText = resp.Text
		}
	}

	// 4. Reply
	replySubject := fmt.Sprintf(protocol.TopicTeamInternalRespond, a.TeamID)
	a.nc.Publish(replySubject, []byte(responseText))
	log.Printf("Agent [%s] replied.", a.Manifest.ID)
}

// toolCallPayload represents a tool invocation extracted from LLM output.
type toolCallPayload struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

// parseToolCall extracts a tool_call JSON block from LLM response text.
// Expected format: {"tool_call": {"name": "...", "arguments": {...}}}
// Returns nil if no tool call is found.
func parseToolCall(text string) *toolCallPayload {
	// Find JSON-like block containing "tool_call"
	start := strings.Index(text, `{"tool_call"`)
	if start == -1 {
		return nil
	}

	// Find matching closing brace
	depth := 0
	end := -1
	for i := start; i < len(text); i++ {
		switch text[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				end = i + 1
				break
			}
		}
		if end != -1 {
			break
		}
	}
	if end == -1 {
		return nil
	}

	var wrapper struct {
		ToolCall toolCallPayload `json:"tool_call"`
	}
	if err := json.Unmarshal([]byte(text[start:end]), &wrapper); err != nil {
		return nil
	}
	if wrapper.ToolCall.Name == "" {
		return nil
	}
	return &wrapper.ToolCall
}

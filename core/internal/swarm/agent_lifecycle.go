package swarm

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/mycelis/core/pkg/pb/swarm"
)

// logTurn is a non-blocking helper that persists a conversation turn.
func (a *Agent) logTurn(role, content, providerID, modelUsed, toolName string, toolArgs map[string]interface{}, parentTurnID, consultationOf string) {
	if a.conversationLogger == nil || a.sessionID == "" {
		return
	}
	idx := a.turnIndex
	a.turnIndex++
	go func() {
		_, err := a.conversationLogger.LogTurn(a.ctx, protocol.ConversationTurnData{
			RunID: a.runID, SessionID: a.sessionID, TenantID: "default", AgentID: a.Manifest.ID, TeamID: a.TeamID,
			TurnIndex: idx, Role: role, Content: content, ProviderID: providerID, ModelUsed: modelUsed, ToolName: toolName,
			ToolArgs: toolArgs, ParentTurnID: parentTurnID, ConsultationOf: consultationOf,
		})
		if err != nil {
			log.Printf("[conversation] Agent [%s] log turn failed: %v", a.Manifest.ID, err)
		}
	}()
}

func (a *Agent) checkInterjection() string {
	a.interjectionMu.Lock()
	defer a.interjectionMu.Unlock()
	msg := a.interjection
	a.interjection = ""
	return msg
}

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

func (a *Agent) SetTeamTopology(inputs, deliveries []string) {
	a.TeamInputs = inputs
	a.TeamDeliveries = deliveries
}

// Start brings the Agent online to listen to its team's internal chatter.
func (a *Agent) Start() {
	subject := fmt.Sprintf(protocol.TopicTeamInternalTrigger, a.TeamID)
	a.nc.Subscribe(subject, a.handleTrigger)
	log.Printf("Agent [%s] (%s) joined Team [%s]", a.Manifest.ID, a.Manifest.Role, a.TeamID)

	personalSubject := fmt.Sprintf(protocol.TopicCouncilRequestFmt, a.Manifest.ID)
	a.nc.Subscribe(personalSubject, a.handleDirectRequest)
	log.Printf("Agent [%s] listening for direct requests on [%s]", a.Manifest.ID, personalSubject)

	a.subscribeInterjection()
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
				Payload:       &pb.MsgEnvelope_Event{Event: &pb.EventPayload{EventType: "agent.heartbeat"}},
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

func (a *Agent) Stop() { a.cancel() }

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
	seen := make(map[string]bool)
	for _, toolName := range a.Manifest.Tools {
		displayName := toolName
		if strings.HasPrefix(toolName, "mcp:") {
			body := toolName[4:]
			if slash := strings.IndexByte(body, '/'); slash >= 0 {
				displayName = body[slash+1:]
			} else {
				displayName = body
			}
			if displayName == "*" {
				continue
			}
		}
		if strings.HasPrefix(toolName, "toolset:") || seen[displayName] {
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

type toolOutputEnvelope struct {
	Message   string                     `json:"message"`
	Artifact  *protocol.ChatArtifactRef  `json:"artifact,omitempty"`
	Artifacts []protocol.ChatArtifactRef `json:"artifacts,omitempty"`
}

func extractToolOutputArtifacts(toolResult string) (string, []protocol.ChatArtifactRef, bool) {
	var toolOutput toolOutputEnvelope
	if err := json.Unmarshal([]byte(toolResult), &toolOutput); err != nil {
		return toolResult, nil, false
	}
	artifacts := make([]protocol.ChatArtifactRef, 0, len(toolOutput.Artifacts)+1)
	if toolOutput.Artifact != nil {
		artifacts = append(artifacts, *toolOutput.Artifact)
	}
	if len(toolOutput.Artifacts) > 0 {
		artifacts = append(artifacts, toolOutput.Artifacts...)
	}
	if len(artifacts) == 0 {
		return toolResult, nil, false
	}
	return toolOutput.Message, artifacts, true
}

package swarm

import (
	"fmt"
	"log"
	"strings"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/pkg/protocol"
)

func (a *Agent) prepareToolCall(input string, toolCall *toolCallPayload, failedToolCalls map[string]int, preflightDone map[string]bool, reinfer func(string, string) bool, result *agentToolLoopResult) bool {
	fingerprint := toolCallFingerprint(toolCall)
	if failedToolCalls[fingerprint] >= 2 {
		reinfer(toolCall.Name, fmt.Sprintf("Policy correction: the exact tool call %s has already failed %d times in this turn. Do not retry it. Choose a different tool or answer directly without tools.", fingerprint, failedToolCalls[fingerprint]))
		return false
	}
	if !shouldCouncilPreflight(toolCall.Name) {
		return true
	}
	member := councilPreflightMember(toolCall.Name)
	if member == "" || preflightDone[member] {
		return true
	}
	preflightDone[member] = true
	summary, err := a.runCouncilPreflight(member, input, toolCall)
	if err != nil {
		log.Printf("Agent [%s] council preflight failed for %s: %v", a.Manifest.ID, toolCall.Name, err)
		return true
	}
	if strings.TrimSpace(summary) == "" {
		return true
	}
	short := strings.TrimSpace(summary)
	if len(short) > 300 {
		short = short[:300] + "..."
	}
	result.consultations = append(result.consultations, protocol.ConsultationEntry{Member: member, Summary: short})
	reinfer("consult_council", fmt.Sprintf("Preflight (%s): %s", member, summary))
	return false
}

func (a *Agent) executeToolIteration(i int, req *cognitive.InferRequest, toolCall *toolCallPayload, failedToolCalls map[string]int, reinfer func(string, string) bool, result *agentToolLoopResult) bool {
	fingerprint := toolCallFingerprint(toolCall)
	log.Printf("Agent [%s] tool_call [%d/%d]: %s", a.Manifest.ID, i+1, a.Manifest.EffectiveMaxIterations(), toolCall.Name)
	result.toolsUsed = append(result.toolsUsed, toolCall.Name)
	if a.eventEmitter != nil && a.runID != "" {
		go a.eventEmitter.Emit(a.ctx, a.runID, protocol.EventToolInvoked, protocol.SeverityInfo, a.Manifest.ID, a.TeamID, map[string]interface{}{"tool": toolCall.Name, "iteration": i + 1}) //nolint:errcheck
	}

	toolCallTurnID := ""
	if a.conversationLogger != nil {
		toolCallTurnID = fmt.Sprintf("%s-%d", a.sessionID, a.turnIndex)
	}
	a.logTurn("tool_call", result.responseText, "", "", toolCall.Name, toolCall.Arguments, "", "")

	toolCtx := WithToolInvocationContext(a.ctx, ToolInvocationContext{
		RunID: a.runID, TeamID: a.TeamID, AgentID: a.Manifest.ID, SourceKind: protocol.SourceKindSystem,
		SourceChannel: fmt.Sprintf(protocol.TopicTeamInternalTrigger, a.TeamID), PayloadKind: protocol.PayloadKindCommand, PlanningOnly: true,
	})
	serverID, _, err := a.toolExecutor.FindToolByName(toolCtx, toolCall.Name)
	if err != nil {
		failedToolCalls[fingerprint]++
		log.Printf("Agent [%s] tool lookup failed: %v", a.Manifest.ID, err)
		if a.eventEmitter != nil && a.runID != "" {
			go a.eventEmitter.Emit(a.ctx, a.runID, protocol.EventToolFailed, protocol.SeverityError, a.Manifest.ID, a.TeamID, map[string]interface{}{"tool": toolCall.Name, "error": err.Error(), "phase": "lookup"}) //nolint:errcheck
		}
		reinfer(toolCall.Name, fmt.Sprintf("Tool '%s' is not available: %v", toolCall.Name, err))
		return false
	}

	isMCPTool := serverID != InternalServerID
	if isMCPTool {
		a.publishToolBusSignal(protocol.PayloadKindStatus, protocol.SourceKindMCP, map[string]any{"state": "invoked", "tool": toolCall.Name, "server_id": serverID.String(), "iteration": i + 1, "arguments": toolCall.Arguments, "team_input": fmt.Sprintf(protocol.TopicTeamInternalTrigger, a.TeamID)})
	}
	toolResult, err := a.toolExecutor.CallTool(toolCtx, serverID, toolCall.Name, toolCall.Arguments)
	if err != nil {
		failedToolCalls[fingerprint]++
		log.Printf("Agent [%s] tool call failed: %v", a.Manifest.ID, err)
		if a.eventEmitter != nil && a.runID != "" {
			go a.eventEmitter.Emit(a.ctx, a.runID, protocol.EventToolFailed, protocol.SeverityError, a.Manifest.ID, a.TeamID, map[string]interface{}{"tool": toolCall.Name, "error": err.Error(), "phase": "execute"}) //nolint:errcheck
		}
		reinfer(toolCall.Name, fmt.Sprintf("Tool %s failed: %v", toolCall.Name, err))
		if isMCPTool {
			a.publishToolBusSignal(protocol.PayloadKindResult, protocol.SourceKindMCP, map[string]any{"state": "failed", "tool": toolCall.Name, "server_id": serverID.String(), "iteration": i + 1, "error": err.Error(), "team_input": fmt.Sprintf(protocol.TopicTeamInternalTrigger, a.TeamID)})
		}
		return false
	}
	if a.eventEmitter != nil && a.runID != "" {
		go a.eventEmitter.Emit(a.ctx, a.runID, protocol.EventToolCompleted, protocol.SeverityInfo, a.Manifest.ID, a.TeamID, map[string]interface{}{"tool": toolCall.Name, "iteration": i + 1}) //nolint:errcheck
	}

	consultMember := ""
	if toolCall.Name == "consult_council" {
		if member, _ := toolCall.Arguments["member"].(string); member != "" {
			consultMember = member
			summary := strings.TrimSpace(toolResult)
			if len(summary) > 300 {
				summary = summary[:300] + "..."
			}
			result.consultations = append(result.consultations, protocol.ConsultationEntry{Member: member, Summary: summary})
		}
	}
	a.logTurn("tool_result", toolResult, "", "", toolCall.Name, nil, toolCallTurnID, consultMember)
	if toolMessage, toolArtifacts, ok := extractToolOutputArtifacts(toolResult); ok {
		result.artifacts = append(result.artifacts, toolArtifacts...)
		toolResult = toolMessage
		if toolResult == "" {
			toolResult = fmt.Sprintf("Tool %s completed successfully.", toolCall.Name)
		}
	}
	if isMCPTool {
		a.publishToolBusSignal(protocol.PayloadKindResult, protocol.SourceKindMCP, map[string]any{"state": "completed", "tool": toolCall.Name, "server_id": serverID.String(), "iteration": i + 1, "result_preview": truncateLog(toolResult, 500), "team_input": fmt.Sprintf(protocol.TopicTeamInternalTrigger, a.TeamID)})
	}
	req.Messages = append(req.Messages,
		cognitive.ChatMessage{Role: "assistant", Content: result.responseText},
		cognitive.ChatMessage{Role: "user", Content: fmt.Sprintf("Tool result from %s:\n%s\n\nContinue your response:", toolCall.Name, toolResult)},
	)
	updated, err := a.brain.InferWithContract(a.ctx, *req)
	if err != nil {
		log.Printf("Agent [%s] re-inference failed: %v", a.Manifest.ID, err)
		return false
	}
	result.resp = updated
	result.responseText = updated.Text
	return true
}

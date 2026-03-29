package swarm

import (
	"fmt"
	"log"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/pkg/protocol"
)

type agentToolLoopResult struct {
	resp          *cognitive.InferResponse
	responseText  string
	toolsUsed     []string
	artifacts     []protocol.ChatArtifactRef
	consultations []protocol.ConsultationEntry
}

func (a *Agent) runToolLoop(input string, priorHistory []cognitive.ChatMessage, req *cognitive.InferRequest, resp *cognitive.InferResponse, profile string) agentToolLoopResult {
	result := agentToolLoopResult{resp: resp, responseText: resp.Text}
	if a.toolExecutor == nil || len(a.Manifest.Tools) == 0 {
		return result
	}

	directAnswerPreferred := preferDirectDraftResponse(input)
	reinferWithToolFeedback := func(toolName string, feedback string) bool {
		req.Messages = append(req.Messages, cognitive.ChatMessage{Role: "assistant", Content: result.responseText})
		req.Messages = append(req.Messages, cognitive.ChatMessage{Role: "user", Content: fmt.Sprintf("Tool result from %s:\n%s\n\nContinue your response:", toolName, feedback)})
		updated, inferErr := a.brain.InferWithContract(a.ctx, *req)
		if inferErr != nil {
			log.Printf("Agent [%s] re-inference after tool feedback failed: %v", a.Manifest.ID, inferErr)
			result.responseText = feedback
			return false
		}
		result.resp = updated
		result.responseText = updated.Text
		return true
	}

	preflightDone := map[string]bool{}
	failedToolCalls := map[string]int{}
	if parseToolCall(result.responseText) == nil && responseSuggestsUnexecutedAction(result.responseText) {
		req.Messages = append(req.Messages,
			cognitive.ChatMessage{Role: "system", Content: "Policy correction: do not provide step-by-step plans when tools are available. Emit exactly one tool_call JSON now for the user's actionable request, or return a concrete blocker."},
			cognitive.ChatMessage{Role: "user", Content: "Re-answer the latest request now under the policy correction."},
		)
		if repaired, repairErr := a.brain.InferWithContract(a.ctx, *req); repairErr == nil && repaired != nil {
			result.resp = repaired
			result.responseText = repaired.Text
		}
	}

	for i := 0; i < a.Manifest.EffectiveMaxIterations(); i++ {
		if interjection := a.checkInterjection(); interjection != "" {
			req.Messages = append(req.Messages, cognitive.ChatMessage{Role: "user", Content: "[OPERATOR INTERJECTION]: " + interjection})
			a.logTurn("interjection", interjection, "", "", "", nil, "", "")
			log.Printf("Agent [%s] processing interjection: %s", a.Manifest.ID, truncateLog(interjection, 100))
			updated, err := a.brain.InferWithContract(a.ctx, *req)
			if err != nil {
				log.Printf("Agent [%s] interjection re-inference failed: %v", a.Manifest.ID, err)
				break
			}
			result.resp = updated
			result.responseText = updated.Text
		}

		toolCall := parseToolCall(result.responseText)
		if toolCall == nil {
			break
		}
		autofillToolArguments(toolCall, input)
		if blocksProposalPlanningTool(toolCall.Name) {
			log.Printf("Agent [%s] proposal-planning tool captured without execution: %s", a.Manifest.ID, toolCall.Name)
			result.toolsUsed = append(result.toolsUsed, toolCall.Name)
			a.logTurn("tool_call", result.responseText, "", "", toolCall.Name, toolCall.Arguments, "", "")
			break
		}
		if directAnswerPreferred && shouldAvoidToolsForDirectDraft(toolCall.Name) {
			if !reinferWithToolFeedback(toolCall.Name, "Policy correction: the user asked for text content in this chat. Respond with the requested content directly. Do not call tools unless they explicitly asked to read or write files, save output, inspect runtime state, execute commands, or route work to other teams.") {
				break
			}
			continue
		}
		if !a.prepareToolCall(input, toolCall, failedToolCalls, preflightDone, reinferWithToolFeedback, &result) {
			continue
		}
		if !a.executeToolIteration(i, req, toolCall, failedToolCalls, reinferWithToolFeedback, &result) {
			continue
		}
	}

	return result
}

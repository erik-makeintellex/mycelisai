package swarm

import (
	"fmt"
	"log"
	"strings"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/pkg/protocol"
)

func appendAssistantHistory(messages *[]cognitive.ChatMessage, content string) {
	if strings.TrimSpace(content) == "" {
		return
	}
	*messages = append(*messages, cognitive.ChatMessage{Role: "assistant", Content: content})
}

type agentToolLoopResult struct {
	resp          *cognitive.InferResponse
	responseText  string
	toolsUsed     []string
	plannedCalls  []protocol.PlannedToolCall
	artifacts     []protocol.ChatArtifactRef
	consultations []protocol.ConsultationEntry
	toolEvidence  []successfulToolEvidence
	// runtimeRecoveryAllowed means the agent made concrete retained-output
	// progress, then the cognitive loop failed before satisfying proof.
	runtimeRecoveryAllowed bool
}

func (a *Agent) runToolLoop(input string, priorHistory []cognitive.ChatMessage, req *cognitive.InferRequest, resp *cognitive.InferResponse, profile string, planningOnly bool, requirement *teamResultRequirement) agentToolLoopResult {
	result := agentToolLoopResult{resp: resp, responseText: resp.Text}
	if a.toolExecutor == nil || len(a.Manifest.Tools) == 0 {
		return result
	}

	directAnswerPreferred := preferDirectDraftResponse(input)
	directAnswerRoute := isDirectAnswerRoute(input)
	projectPackageBase := projectPackageInferenceBase(req.Messages, requirement)
	inferenceAttempt := 2
	infer := func(reason string) (*cognitive.InferResponse, error) {
		attempt := inferenceAttempt
		inferenceAttempt++
		return a.inferWithExecutionBounds(*req, reason, attempt)
	}
	reinferWithToolFeedback := func(toolName string, feedback string) bool {
		latest := cognitive.ChatMessage{Role: "user", Content: fmt.Sprintf("Tool result from %s:\n%s\n\nContinue your response:", toolName, feedback)}
		if projectPackageHistoryEnabled(requirement) {
			req.Messages = compactProjectPackageInferenceHistory(projectPackageBase, result.toolEvidence, latest)
		} else {
			appendAssistantHistory(&req.Messages, result.responseText)
			req.Messages = append(req.Messages, latest)
		}
		updated, inferErr := infer("tool_feedback")
		if inferErr != nil || updated == nil {
			log.Printf("Agent [%s] re-inference after tool feedback failed: %v", a.Manifest.ID, inferErr)
			result.responseText = feedback
			if len(result.toolEvidence) > 0 {
				result.runtimeRecoveryAllowed = true
			}
			return false
		}
		result.resp = updated
		result.responseText = updated.Text
		return true
	}

	preflightDone := map[string]bool{}
	failedToolCalls := map[string]int{}
	completedToolCalls := map[string]bool{}
	contractCorrections := 0
	unsafeCorrectionUsed := map[string]bool{}
	if parseToolCall(result.responseText) == nil && responseSuggestsUnexecutedAction(result.responseText) {
		policy := "Policy correction: do not provide step-by-step plans when tools are available. Emit exactly one tool_call JSON now for the user's actionable request, or return a concrete blocker. Re-answer the latest request now."
		if projectPackageHistoryEnabled(requirement) {
			req.Messages = compactProjectPackageInferenceHistory(projectPackageBase, result.toolEvidence, cognitive.ChatMessage{Role: "user", Content: policy})
		} else {
			req.Messages = append(req.Messages, cognitive.ChatMessage{Role: "system", Content: policy})
		}
		if repaired, repairErr := infer("policy_correction"); repairErr == nil && repaired != nil {
			result.resp = repaired
			result.responseText = repaired.Text
		}
	}

	loopLimit := resultContractLoopLimit(a.Manifest.EffectiveMaxIterations(), requirement)
	for i := 0; i < loopLimit; i++ {
		if interjection := a.checkInterjection(); interjection != "" {
			latest := cognitive.ChatMessage{Role: "user", Content: "[OPERATOR INTERJECTION]: " + interjection}
			if projectPackageHistoryEnabled(requirement) {
				req.Messages = compactProjectPackageInferenceHistory(projectPackageBase, result.toolEvidence, latest)
			} else {
				req.Messages = append(req.Messages, latest)
			}
			a.logTurn("interjection", interjection, "", "", "", nil, "", "")
			log.Printf("Agent [%s] processing interjection: %s", a.Manifest.ID, truncateLog(interjection, 100))
			updated, err := infer("interjection")
			if err != nil {
				log.Printf("Agent [%s] interjection re-inference failed: %v", a.Manifest.ID, err)
				break
			}
			result.resp = updated
			result.responseText = updated.Text
		}

		toolCall, parseFailure := parseToolCallForExecution(result.responseText)
		if parseFailure != nil {
			target := "unsafe:" + strings.TrimSpace(parseFailure.ToolName)
			if unsafeCorrectionUsed[target] {
				log.Printf("Agent [%s] inference correction suppressed reason=malformed_tool target=%s", a.Manifest.ID, parseFailure.ToolName)
				break
			}
			unsafeCorrectionUsed[target] = true
			if !reinferWithToolFeedback(parseFailure.ToolName, "Tool call correction: "+parseFailure.Error()+". Return one complete, valid tool_call JSON object with every required argument. The malformed call was not executed.") {
				break
			}
			continue
		}
		if toolCall == nil {
			issues := resultContractIssues(requirement, result.artifacts, result.toolEvidence)
			if len(issues) > 0 && contractCorrections < maxResultContractCorrections {
				contractCorrections++
				requirementPrompt := resultContractCorrectionPrompt(requirement, issues, result.artifacts, result.toolEvidence)
				if projectPackageHistoryEnabled(requirement) {
					req.Messages = compactProjectPackageInferenceHistory(projectPackageBase, result.toolEvidence, cognitive.ChatMessage{
						Role: "user", Content: requirementPrompt + " Continue the approved delivery now and satisfy the missing result-contract evidence.",
					})
				} else {
					appendAssistantHistory(&req.Messages, result.responseText)
					req.Messages = append(req.Messages,
						cognitive.ChatMessage{Role: "system", Content: requirementPrompt},
						cognitive.ChatMessage{Role: "user", Content: "Continue the approved delivery now and satisfy the missing result-contract evidence."},
					)
				}
				updated, err := infer("result_contract")
				if err != nil || updated == nil {
					log.Printf("Agent [%s] result-contract correction failed: %v", a.Manifest.ID, err)
					if len(result.toolEvidence) > 0 {
						result.runtimeRecoveryAllowed = true
					}
					break
				}
				result.resp = updated
				result.responseText = updated.Text
				continue
			}
			break
		}
		normalizeAgentToolCallArguments(toolCall, a.TeamID, input)
		if validationFailure := validateMutationToolCall(toolCall); validationFailure != nil {
			target := "unsafe:" + strings.TrimSpace(validationFailure.ToolName)
			if unsafeCorrectionUsed[target] {
				log.Printf("Agent [%s] inference correction suppressed reason=invalid_tool target=%s", a.Manifest.ID, validationFailure.ToolName)
				break
			}
			unsafeCorrectionUsed[target] = true
			if !reinferWithToolFeedback(validationFailure.ToolName, "Tool call correction: "+validationFailure.Error()+". Return one complete, valid tool_call JSON object with every required argument. The invalid call was not executed.") {
				break
			}
			continue
		}
		if requirement.active() && strings.EqualFold(requirement.Kind, "project_package") && toolCall.Name == "store_artifact" {
			if !reinferWithToolFeedback(toolCall.Name, "Project-package contracts require physical files. Do not call store_artifact. Use write_file for the next missing required file, or read_file on the written entrypoint when only structural readback remains.") {
				break
			}
			continue
		}
		if issues := resultContractIssues(requirement, result.artifacts, result.toolEvidence); len(issues) > 0 && !resultContractEvidenceToolAllowed(requirement, toolCall.Name, result.artifacts, result.toolEvidence) {
			feedback := resultContractCorrectionPrompt(requirement, issues, result.artifacts, result.toolEvidence) + " Do not call " + toolCall.Name + " while required package evidence is incomplete."
			if !reinferWithToolFeedback(toolCall.Name, feedback) {
				break
			}
			continue
		}
		fingerprint := toolCallFingerprint(toolCall)
		if completedToolCalls[fingerprint] {
			if !reinferWithToolFeedback(toolCall.Name, "That exact tool call already completed successfully in this turn. Do not repeat it. Return the concise final result or choose a different tool required to finish the ask.") {
				break
			}
			continue
		}
		if directAnswerRoute && blocksProposalPlanningTool(toolCall.Name) {
			if !reinferWithToolFeedback(toolCall.Name, "Authority correction: direct-answer mode cannot run mutation-capable tools. Use the requested read-only tool when one is available, or answer without a tool. Do not delegate, create, write, store, activate, publish, or execute commands.") {
				break
			}
			continue
		}
		if planningOnly && blocksProposalPlanningTool(toolCall.Name) {
			log.Printf("Agent [%s] proposal-planning tool captured without execution: %s", a.Manifest.ID, toolCall.Name)
			result.toolsUsed = append(result.toolsUsed, toolCall.Name)
			result.plannedCalls = append(result.plannedCalls, protocol.PlannedToolCall{
				Name:      strings.TrimSpace(toolCall.Name),
				Arguments: toolCall.Arguments,
			})
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
		evidenceCount := len(result.toolEvidence)
		if !a.executeToolIteration(i, loopLimit, input, req, toolCall, failedToolCalls, reinferWithToolFeedback, &result, planningOnly, requirement, projectPackageBase) {
			if len(result.toolEvidence) > evidenceCount {
				completedToolCalls[fingerprint] = true
				contractCorrections = 0
				delete(unsafeCorrectionUsed, "unsafe:"+strings.TrimSpace(toolCall.Name))
			}
			if result.runtimeRecoveryAllowed {
				break
			}
			continue
		}
		completedToolCalls[fingerprint] = true
		delete(unsafeCorrectionUsed, "unsafe:"+strings.TrimSpace(toolCall.Name))
		if len(result.toolEvidence) > evidenceCount {
			contractCorrections = 0
		}
		if entrypoint := currentProjectPackageEntrypointReadback(requirement, result.artifacts, result.toolEvidence); entrypoint != "" {
			completedToolCalls[toolCallFingerprint(&toolCallPayload{Name: "read_file", Arguments: map[string]any{"path": entrypoint}})] = true
		}
	}

	if a.completeProjectPackageRuntimeFallback(input, requirement, &result, planningOnly) {
		result.artifacts = reconcileToolBackedArtifacts(result.artifacts, result.toolEvidence, input)
	}
	return result
}

package swarm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
)

func stripToolCallJSON(text string) string {
	keyword := `"tool_call"`
	idx := strings.Index(text, keyword)
	if idx == -1 {
		return text
	}
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
	if cleaned := strings.TrimSpace(text[:start]); cleaned != "" {
		return cleaned
	}
	return ""
}

func (a *Agent) publishToolBusSignal(payloadKind protocol.SignalPayloadKind, sourceKind protocol.SignalSourceKind, payload map[string]any) {
	if a.nc == nil || strings.TrimSpace(a.TeamID) == "" {
		return
	}
	subject := ""
	switch payloadKind {
	case protocol.PayloadKindStatus:
		subject = fmt.Sprintf(protocol.TopicTeamSignalStatus, a.TeamID)
	case protocol.PayloadKindResult:
		subject = fmt.Sprintf(protocol.TopicTeamSignalResult, a.TeamID)
	default:
		return
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Agent [%s] signal payload marshal failed: %v", a.Manifest.ID, err)
		return
	}
	sourceChannel := fmt.Sprintf(protocol.TopicTeamInternalTrigger, a.TeamID)
	wrapped, err := protocol.WrapSignalPayloadWithMeta(sourceKind, sourceChannel, payloadKind, a.runID, a.TeamID, a.Manifest.ID, raw)
	if err != nil {
		log.Printf("Agent [%s] signal envelope wrap failed: %v", a.Manifest.ID, err)
		return
	}
	if err := a.nc.Publish(subject, wrapped); err != nil {
		log.Printf("Agent [%s] publish signal failed on [%s]: %v", a.Manifest.ID, subject, err)
		return
	}
	if a.internalTools != nil {
		channelKey := resolveSignalCheckpointChannelKey(subject, nil)
		metadata := map[string]any{"subject": subject, "source_kind": string(sourceKind), "payload_kind": string(payloadKind), "team_id": a.TeamID, "agent_id": a.Manifest.ID}
		if strings.TrimSpace(a.runID) != "" {
			metadata["run_id"] = strings.TrimSpace(a.runID)
		}
		if _, err := a.internalTools.upsertSignalCheckpoint(a.ctx, channelKey, a.Manifest.ID, string(wrapped), metadata); err != nil {
			log.Printf("Agent [%s] checkpoint update failed on [%s]: %v", a.Manifest.ID, channelKey, err)
		}
	}
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
	input := normalizeTeamTriggerInput(msg.Data)
	planningOnly := teamTriggerPlanningOnly(msg.Data)
	requirement := teamResultRequirementFromTrigger(msg.Data, planningOnly)
	log.Printf("Agent [%s] thinking about: %s", a.Manifest.ID, input)
	result := a.processMessageStructuredWithRequirement(input, nil, planningOnly, requirement)
	if result.Text == "" && len(result.Artifacts) == 0 && result.Availability == nil {
		if msg.Reply != "" {
			msg.Respond([]byte(fmt.Sprintf("[%s] No response — LLM may be unavailable.", a.Manifest.ID)))
		}
		return
	}
	if msg.Reply != "" {
		msg.Respond([]byte(teamAgentRequestReply(result)))
	}
	a.nc.Publish(fmt.Sprintf(protocol.TopicTeamInternalRespond, a.TeamID), teamAgentResponsePayloadForTrigger(result, msg.Data, a.TeamID))
	log.Printf("Agent [%s] replied.", a.Manifest.ID)
}

func teamAgentResponsePayloadForTrigger(result ProcessResult, trigger []byte, teamID string) []byte {
	payload := teamAgentResponsePayload(result)
	correlation := correlationFromPayload(trigger)
	if correlation == nil {
		return payload
	}
	correlation.TeamID = firstNonEmptySignalString(correlation.TeamID, teamID)
	return correlatedTeamResponsePayload(payload, correlation)
}

func teamTriggerPlanningOnly(data []byte) bool {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return true
	}

	var ask protocol.TeamAsk
	if err := json.Unmarshal(trimmed, &ask); err != nil || ask.IsZero() {
		return true
	}
	for _, key := range []string{"run_id", "contract_id", "intent_proof_id"} {
		value, ok := ask.Context[key].(string)
		if !ok || strings.TrimSpace(value) == "" {
			return true
		}
	}
	return false
}

func teamAgentResponsePayload(result ProcessResult) []byte {
	if result.Availability != nil && !result.Availability.Available {
		degradationState := strings.TrimSpace(result.Availability.Code)
		if degradationState == "" {
			degradationState = "cognitive_engine_unavailable"
		}
		summary := strings.TrimSpace(result.Availability.Summary)
		if summary == "" {
			summary = "The team could not use its configured cognitive engine."
		}
		recoveryOptions := []string{}
		if action := strings.TrimSpace(result.Availability.RecommendedAction); action != "" {
			recoveryOptions = append(recoveryOptions, action)
		}
		responsePayload, err := json.Marshal(map[string]any{
			"text":               result.Text,
			"tools_used":         result.ToolsUsed,
			"planned_tool_calls": result.PlannedToolCalls,
			"artifacts":          result.Artifacts,
			"availability":       result.Availability,
			"provider_id":        result.ProviderID,
			"model_used":         result.ModelUsed,
			"consultations":      result.Consultations,
			"state":              "degraded",
			"headline":           "Team work needs attention",
			"summary":            summary,
			"details":            summary,
			"degradation_state":  degradationState,
			"recovery_options":   recoveryOptions,
		})
		if err == nil {
			return responsePayload
		}
	}
	responsePayload, err := json.Marshal(result)
	if err != nil {
		return []byte(result.Text)
	}
	return responsePayload
}

func normalizeTeamTriggerInput(data []byte) string {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return ""
	}

	var ask protocol.TeamAsk
	if err := json.Unmarshal(trimmed, &ask); err == nil && !ask.IsZero() {
		return renderTeamAskPrompt(ask.Normalize())
	}
	return string(data)
}

func renderTeamAskPrompt(ask protocol.TeamAsk) string {
	var sb strings.Builder
	sb.WriteString("You have received a structured team ask.\n")
	sb.WriteString("Use the ask to stay aligned on mission, scope, and proof needs.\n")
	sb.WriteString("Do not force your response into a rigid template unless the ask explicitly requires one.\n")
	sb.WriteString("Deliver the best output for the job while making sure it satisfies the ask goal, constraints, and required evidence.\n")
	sb.WriteString(fmt.Sprintf("Ask kind: %s\n", ask.AskKind))
	sb.WriteString(fmt.Sprintf("Lane role: %s\n", ask.LaneRole))
	sb.WriteString(fmt.Sprintf("Goal: %s\n", ask.Goal))
	if len(ask.OwnedScope) > 0 {
		sb.WriteString("Owned scope:\n")
		for _, item := range ask.OwnedScope {
			sb.WriteString(fmt.Sprintf("- %s\n", item))
		}
	}
	if len(ask.Constraints) > 0 {
		sb.WriteString("Constraints:\n")
		for _, item := range ask.Constraints {
			sb.WriteString(fmt.Sprintf("- %s\n", item))
		}
	}
	if len(ask.RequiredCapabilities) > 0 {
		sb.WriteString("Required capabilities:\n")
		for _, item := range ask.RequiredCapabilities {
			sb.WriteString(fmt.Sprintf("- %s\n", item))
		}
	}
	if len(ask.ExitCriteria) > 0 {
		sb.WriteString("Exit criteria:\n")
		for _, item := range ask.ExitCriteria {
			sb.WriteString(fmt.Sprintf("- %s\n", item))
		}
	}
	if len(ask.EvidenceRequired) > 0 {
		sb.WriteString("Evidence required:\n")
		for _, item := range ask.EvidenceRequired {
			sb.WriteString(fmt.Sprintf("- %s\n", item))
		}
	}
	renderTeamAskResultContract(&sb, ask.Context)
	if ask.Context != nil {
		if raw, err := json.Marshal(compactTeamAskContext(ask.Context)); err == nil {
			sb.WriteString(fmt.Sprintf("Context: %s\n", string(raw)))
		}
	}
	sb.WriteString("Complete the ask within scope, match the mission, and report the outcome clearly.")
	return sb.String()
}

func renderTeamAskResultContract(sb *strings.Builder, context map[string]any) {
	if len(context) == 0 {
		return
	}
	contract, ok := context["result_contract"].(map[string]any)
	if !ok || len(contract) == 0 {
		return
	}
	kind := strings.TrimSpace(stringValue(contract["kind"]))
	files := stringSlice(contract["files_required"])
	expectedOutputs := stringSlice(contract["expected_outputs"])
	acceptanceCriteria := stringSlice(contract["acceptance_criteria"])
	proofRequirements := stringSlice(contract["proof_required"])
	entrypointRequired := boolValue(contract["entrypoint_required"])
	folderRequired := boolValue(contract["folder_required"])
	validationRequired := boolValue(contract["validation_required"])
	proofRequired := boolValue(contract["proof_ref_required"])
	repairChannel := strings.TrimSpace(stringValue(contract["repair_channel"]))
	if kind == "" && len(files) == 0 && len(expectedOutputs) == 0 && len(acceptanceCriteria) == 0 && len(proofRequirements) == 0 &&
		!entrypointRequired && !folderRequired && !validationRequired && !proofRequired && repairChannel == "" {
		return
	}

	sb.WriteString("Output contract:\n")
	if kind != "" {
		sb.WriteString(fmt.Sprintf("- Kind: %s\n", kind))
	}
	if len(files) > 0 {
		sb.WriteString(fmt.Sprintf("- Required files: %s\n", strings.Join(files, ", ")))
	}
	renderTeamAskContractList(sb, "Expected outputs", expectedOutputs)
	renderTeamAskContractList(sb, "Acceptance criteria", acceptanceCriteria)
	renderTeamAskContractList(sb, "Proof requirements", proofRequirements)
	if entrypointRequired {
		sb.WriteString("- Return a direct entrypoint.\n")
	}
	if folderRequired {
		sb.WriteString("- Return the retained output folder.\n")
	}
	if validationRequired {
		sb.WriteString("- Read the retained output back to establish structural evidence. Readback alone does not prove semantic acceptance.\n")
	}
	if proofRequired {
		sb.WriteString("- Provide tool-backed evidence for the server/live validation layer to attach the authoritative proof reference.\n")
	}
	if repairChannel != "" {
		sb.WriteString(fmt.Sprintf("- If validation fails, report the blocker through %s instead of claiming completion.\n", repairChannel))
	}
	sb.WriteString("- Worker completion requires successful retained-output tool writes and successful structural readback when requested. Semantic acceptance and final proof remain authoritative in server/live validation. Prose and declared metadata are not evidence.\n")
}

func compactTeamAskContext(context map[string]any) map[string]any {
	if len(context) == 0 {
		return nil
	}
	compact := make(map[string]any)
	for _, key := range []string{
		"run_id",
		"contract_id",
		"intent_proof_id",
		"work_item_id",
		"team_id",
		"team_evocation_brief",
		"research_council_handoff",
		"delivery_team_responsibility",
	} {
		if value, ok := context[key]; ok {
			compact[key] = value
		}
	}
	return compact
}

func (a *Agent) handleDirectRequest(msg *nats.Msg) {
	select {
	case <-a.ctx.Done():
		return
	default:
	}
	input, history := a.parseConversationPayload(msg.Data)
	log.Printf("Agent [%s] direct request (%d prior turns): %s", a.Manifest.ID, len(history), truncateLog(input, 200))
	result := a.processMessageStructured(input, history)
	if msg.Reply != "" {
		if respBytes, err := json.Marshal(result); err == nil {
			msg.Respond(respBytes)
		} else {
			fallback := result.Text
			if fallback == "" && result.Availability != nil {
				fallback = result.Availability.Summary
			}
			msg.Respond([]byte(fallback))
		}
	}
	log.Printf("Agent [%s] direct request replied (tools: %v readable=%t).", a.Manifest.ID, result.ToolsUsed, strings.TrimSpace(result.Text) != "")
}

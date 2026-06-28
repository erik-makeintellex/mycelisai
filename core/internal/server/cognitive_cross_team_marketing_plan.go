package server

import (
	"fmt"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

func contentMarketingCrossTeamMutationTools(text, lower string) []string {
	if requestAsksForContentMarketingCrossTeam(lower) {
		return []string{"create_team", "write_file", "delegate_task"}
	}
	return nil
}

func inferContentMarketingCrossTeamPlanFromRequest(text string) ([]protocol.PlannedToolCall, bool) {
	trimmed := strings.TrimSpace(text)
	lower := strings.ToLower(trimmed)
	if trimmed == "" || !requestAsksForContentMarketingCrossTeam(lower) {
		return nil, false
	}

	sourceTeamID := sourceTeamIDForCrossTeamMarketingRequest(trimmed)
	marketingTeamID := marketingTeamIDForCrossTeamRequest(trimmed)
	marketingContract := marketingContentContractForSourceRequest(trimmed)
	marketingTeam := protocol.PlannedToolCall{
		Name: "create_team",
		Arguments: map[string]any{
			"team_id":                  marketingTeamID,
			"name":                     "Outcome Marketing Team",
			"role":                     "marketing lead",
			"goal":                     "Turn verified output changes and evidence into operator-reviewable marketing material.",
			"staffing_mode":            "lead_only_start",
			"initial_member_count":     1,
			"recommended_member_limit": 3,
			"expansion_policy":         "Add copy, visual, or launch specialists only after the lead names a concrete deliverable and proof need.",
			"content_contract":         marketingContract,
			"required_capabilities":    []string{"team_orchestration", "write_file", "store_artifact"},
		},
	}

	evidencePath := sourceEvidencePathForMarketingRequest(trimmed, sourceTeamID)
	marketingHandoffPath := groupWorkspaceFolderForTeamID(marketingTeamID) + "/marketing/MARKETING_HANDOFF.md"
	sourceImprovementCall := protocol.PlannedToolCall{
		Name: "delegate_task",
		Arguments: map[string]any{
			"team_id": sourceTeamID,
			"task": fmt.Sprintf(
				"Improve the source deliverable for this Outcome, then retain evidence examples at %s showing the changes before marketing uses them.",
				evidencePath,
			),
			"ask": map[string]any{
				"ask_kind":          string(protocol.TeamAskKindImplementation),
				"lane_role":         string(protocol.TeamLaneRoleImplementer),
				"goal":              "Improve the deliverable and provide evidence showing the change.",
				"operation":         "source_improvement_for_marketing_handoff",
				"approval_posture":  string(protocol.ApprovalPostureRequired),
				"owned_scope":       []string{groupWorkspaceFolderForTeamID(sourceTeamID)},
				"exit_criteria":     sourceImprovementExitCriteria(trimmed),
				"evidence_required": sourceEvidenceRequirements(trimmed),
				"context": map[string]any{
					"operator_request":       trimmed,
					"marketing_team_id":      marketingTeamID,
					"source_evidence_path":   evidencePath,
					"marketing_handoff_path": marketingHandoffPath,
				},
			},
			"expected_outputs": sourceExpectedOutputs(trimmed),
			"expected_proof":   sourceEvidenceRequirements(trimmed),
		},
	}
	evidenceExamplesCall := protocol.PlannedToolCall{
		Name: "write_file",
		Arguments: map[string]any{
			"path":             evidencePath,
			"content":          sourceChangeEvidenceMarkdown(trimmed, sourceTeamID, marketingTeamID),
			"validation":       "Evidence examples must name the changed behavior or output, how to see it, and the proof the marketing team can rely on.",
			"source_team":      sourceTeamID,
			"target_team":      marketingTeamID,
			"content_contract": contentContractForTeamRequest("source proof examples for marketing handoff " + trimmed),
		},
	}
	marketingHandoffCall := protocol.PlannedToolCall{
		Name: "write_file",
		Arguments: map[string]any{
			"path":             marketingHandoffPath,
			"content":          contentMarketingHandoffMarkdown(trimmed, sourceTeamID, marketingTeamID, evidencePath),
			"validation":       "Marketing handoff must be grounded in retained evidence examples and clearly separate confirmed facts from proposed messaging.",
			"source_team":      sourceTeamID,
			"target_team":      marketingTeamID,
			"content_contract": marketingContract,
		},
	}
	marketingDelegateCall := protocol.PlannedToolCall{
		Name: "delegate_task",
		Arguments: map[string]any{
			"team_id": marketingTeamID,
			"task": fmt.Sprintf(
				"Use %s and %s to generate marketing about the improved deliverable without inventing unproven claims.",
				marketingHandoffPath,
				evidencePath,
			),
			"ask": map[string]any{
				"ask_kind":         string(protocol.TeamAskKindImplementation),
				"lane_role":        string(protocol.TeamLaneRoleImplementer),
				"goal":             "Generate marketing output grounded in the updated deliverable and retained evidence examples.",
				"operation":        "marketing_from_source_evidence_handoff",
				"approval_posture": string(protocol.ApprovalPostureRequired),
				"owned_scope":      []string{groupWorkspaceFolderForTeamID(marketingTeamID)},
				"exit_criteria": []string{
					"marketing names the deliverable change it is based on",
					"claims trace back to retained evidence examples",
					"output includes copy direction and next review step",
				},
				"evidence_required": []string{"marketing output path", "source evidence reference", "assumptions list"},
				"context": map[string]any{
					"operator_request":       trimmed,
					"source_team_id":         sourceTeamID,
					"source_evidence_path":   evidencePath,
					"marketing_handoff_path": marketingHandoffPath,
				},
			},
			"expected_outputs": []string{"marketing brief or campaign copy grounded in retained evidence"},
			"expected_proof":   []string{"retained marketing output", "linked source evidence"},
		},
	}

	return []protocol.PlannedToolCall{
		marketingTeam,
		sourceImprovementCall,
		evidenceExamplesCall,
		marketingHandoffCall,
		marketingDelegateCall,
	}, true
}

func requestAsksForContentMarketingCrossTeam(lower string) bool {
	if !strings.Contains(lower, "marketing") {
		return false
	}
	if !requestContainsAny(lower, []string{"game", "gameplay", "playable", "app", "application", "media", "content", "document", "report", "product", "deliverable", "output"}) {
		return false
	}
	return requestContainsAny(lower, []string{
		"improve", "change", "update", "let", "notify", "inform", "tell", "handoff", "hand off", "generate marketing", "marketing about",
	})
}

func sourceTeamIDForCrossTeamMarketingRequest(text string) string {
	for _, candidate := range groupIDsFromText(text) {
		if !strings.Contains(candidate, "marketing") {
			return candidate
		}
	}
	lower := strings.ToLower(text)
	if strings.Contains(lower, "game delivery team") {
		return "game-delivery-team"
	}
	if strings.Contains(lower, "media") {
		return "media-delivery-team"
	}
	if strings.Contains(lower, "document") || strings.Contains(lower, "report") {
		return "content-delivery-team"
	}
	return "source-delivery-team"
}

func marketingTeamIDForCrossTeamRequest(text string) string {
	lower := strings.ToLower(text)
	if id := extractLabeledTokenAfterMarker(text, "marketing", "team_id"); id != "" {
		return id
	}
	if id := extractLabeledTokenAfterMarker(text, "marketing", "team id"); id != "" {
		return id
	}
	for _, candidate := range groupIDsFromText(text) {
		if strings.Contains(candidate, "marketing") {
			return candidate
		}
	}
	return generatedTeamIDForRequest("Marketing Delivery Team", lower)
}

func marketingContentContractForSourceRequest(request string) map[string]any {
	contract := contentContractForTeamRequest("marketing text report media brief " + request)
	contract["content_types"] = uniqueOrderedTools(append(confirmedActionStringSlice(contract["content_types"]), "marketing"))
	contract["expected_outputs"] = uniqueOrderedTools(append(confirmedActionStringSlice(contract["expected_outputs"]),
		"marketing brief grounded in retained evidence",
		"campaign copy or positioning notes",
	))
	contract["acceptance_criteria"] = uniqueOrderedTools(append(confirmedActionStringSlice(contract["acceptance_criteria"]),
		"marketing claims trace to retained source examples",
		"confirmed source facts are separated from creative positioning",
		"deliverable change, audience, message, and next review step are clear",
	))
	contract["proof_required"] = uniqueOrderedTools(append(confirmedActionStringSlice(contract["proof_required"]),
		"linked source evidence examples",
		"marketing output readback proof",
	))
	return contract
}

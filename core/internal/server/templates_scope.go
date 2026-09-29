package server

import "github.com/mycelis/core/pkg/protocol"

// buildScopeFromBlueprintFor extracts scope validation metadata from a
// blueprint, including the server-side approval classification (A2b C1). It
// classifies a blueprint with the same rules and
// constants as chat and council proposals: every agent tool is a planned
// action for buildApprovalPolicy (tool risk classes, cost, external data),
// agent/team count raises capability risk, and applyApproverTier mirrors the
// confirm-time tier. Only the decoded blueprint is read, never client risk.
func buildScopeFromBlueprintFor(bp *protocol.MissionBlueprint, profile userGovernanceProfile) *protocol.ScopeValidation {
	toolSet := make(map[string]bool)
	tools := []string{}
	planned := []protocol.PlannedToolCall{}
	totalAgents := 0
	for _, team := range bp.Teams {
		totalAgents += len(team.Agents)
		for _, agent := range team.Agents {
			for _, tool := range agent.Tools {
				planned = append(planned, protocol.PlannedToolCall{Name: tool})
				if !toolSet[tool] {
					toolSet[tool] = true
					tools = append(tools, tool)
				}
			}
		}
	}

	risk := "low"
	if totalAgents > 5 || len(bp.Teams) > 2 {
		risk = "medium"
	}
	if totalAgents > 10 || len(bp.Teams) > 4 {
		risk = "high"
	}
	sizeRisk := risk // RiskLevel keeps its size-based meaning
	for _, tool := range tools {
		risk = maxRisk(risk, blueprintToolRisk(tool))
	}

	approval := buildApprovalPolicy(profile, planned, nil)
	if approval == nil && risk != "low" {
		approval = &protocol.ApprovalPolicy{ApprovalMode: "auto_allowed", ApprovalReason: "auto_approve", CapabilityRisk: "low"}
	}
	if approval != nil && approvalRank(risk) > approvalRank(approval.CapabilityRisk) {
		approval.CapabilityRisk = risk
		if approvalRank(risk) >= 3 {
			approval.ApprovalRequired, approval.ApprovalMode, approval.ApprovalReason = true, "required", "capability_risk"
		}
	}
	approval = applyApproverTier(approval)

	scope := &protocol.ScopeValidation{
		Tools:             tools,
		AffectedResources: []string{"missions", "teams", "service_manifests"},
		RiskLevel:         sizeRisk,
		Approval:          approval,
	}
	if approval != nil {
		scope.CapabilityIDs = approval.CapabilityIDs
		scope.ExternalDataUse = approval.ExternalDataUse
		scope.EstimatedCost = approval.EstimatedCost
	}
	return scope
}

// blueprintToolRisk uses the shared fail-closed classifier (A2b C1b).
func blueprintToolRisk(tool string) string {
	return capabilityRiskForTool(tool, nil)
}

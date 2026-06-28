package server

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

var (
	teamEvocationBriefPathPattern = regexp.MustCompile(`(?i)(groups/[a-z0-9][a-z0-9._-]*/planning/TEAM_EVOCATION\.md)`)
	groupScopedPathPattern        = regexp.MustCompile(`(?i)groups/([a-z0-9][a-z0-9._-]*)/`)
	explicitTeamIDPattern         = regexp.MustCompile(`(?i)(?:team_id|team id)\s+([a-z0-9][a-z0-9._-]{2,})`)
)

func teamEvocationContinuationMutationTools(text, lower string) []string {
	if requestAsksToContinueTeamEvocation(lower) && teamIDFromTeamEvocationContinuationRequest(text) != "" {
		return []string{"write_file", "delegate_task"}
	}
	return nil
}

func inferTeamEvocationContinuationPlanFromRequest(text string) ([]protocol.PlannedToolCall, bool) {
	trimmed := strings.TrimSpace(text)
	lower := strings.ToLower(trimmed)
	if trimmed == "" || !requestAsksToContinueTeamEvocation(lower) {
		return nil, false
	}

	teamID := teamIDFromTeamEvocationContinuationRequest(trimmed)
	if teamID == "" {
		return nil, false
	}

	briefPath := firstTeamEvocationBriefPath(trimmed)
	if briefPath == "" {
		briefPath = groupWorkspaceFolderForTeamID(teamID) + "/planning/TEAM_EVOCATION.md"
	}
	contract := contentContractForTeamRequest(trimmed)
	researchPath := groupWorkspaceFolderForTeamID(teamID) + "/planning/RESEARCH_COUNCIL_HANDOFF.md"
	researchCall := protocol.PlannedToolCall{
		Name: "write_file",
		Arguments: map[string]any{
			"path":                researchPath,
			"content":             teamResearchHandoffMarkdown(trimmed, teamID, briefPath, contract),
			"validation":          "Retained research/council handoff must make the implementation strategy, team responsibilities, output contract, proof gates, and unknowns clear before delegated build work starts.",
			"content_contract":    contract,
			"acceptance_criteria": confirmedActionStringSlice(contract["acceptance_criteria"]),
			"proof_required":      confirmedActionStringSlice(contract["proof_required"]),
			"evocation_brief":     briefPath,
		},
	}
	delegateCall := protocol.PlannedToolCall{
		Name: "delegate_task",
		Arguments: map[string]any{
			"team_id": teamID,
			"task":    teamEvocationDelegationGoal(trimmed, contract),
			"ask": map[string]any{
				"ask_kind":              string(protocol.TeamAskKindImplementation),
				"lane_role":             string(protocol.TeamLaneRoleImplementer),
				"goal":                  teamEvocationDelegationGoal(trimmed, contract),
				"operation":             "continue_from_research_handoff",
				"approval_posture":      string(protocol.ApprovalPostureRequired),
				"owned_scope":           []string{groupWorkspaceFolderForTeamID(teamID)},
				"constraints":           teamEvocationDelegationConstraints(),
				"required_capabilities": requiredCapabilitiesForContentContract(contract),
				"exit_criteria":         confirmedActionStringSlice(contract["acceptance_criteria"]),
				"evidence_required":     confirmedActionStringSlice(contract["proof_required"]),
				"context": map[string]any{
					"operator_request":             trimmed,
					"team_evocation_brief":         briefPath,
					"research_council_handoff":     researchPath,
					"research_team_responsibility": "Prepare domain/stack/options review, unknowns, and specialist recommendations before implementation.",
					"delivery_team_responsibility": "Use the handoff to produce the retained user-facing output package and proof.",
				},
			},
			"expected_outputs":      confirmedActionStringSlice(contract["expected_outputs"]),
			"expected_proof":        confirmedActionStringSlice(contract["proof_required"]),
			"required_capabilities": requiredCapabilitiesForContentContract(contract),
			"evocation_brief":       briefPath,
			"research_handoff":      researchPath,
		},
	}
	return []protocol.PlannedToolCall{researchCall, delegateCall}, true
}

func requestAsksToContinueTeamEvocation(lower string) bool {
	if !requestContainsAny(lower, []string{"brief", "evocation", "handoff", "retained", "research", "council"}) {
		return false
	}
	return requestContainsAny(lower, []string{
		"use", "continue", "now", "build", "make", "work happen", "actual", "produce", "generate", "delegate", "hand off", "handoff",
	})
}

func teamIDFromTeamEvocationContinuationRequest(text string) string {
	if path := firstTeamEvocationBriefPath(text); path != "" {
		if matches := groupScopedPathPattern.FindStringSubmatch(path); len(matches) == 2 {
			return strings.TrimSpace(matches[1])
		}
	}
	if matches := explicitTeamIDPattern.FindStringSubmatch(text); len(matches) == 2 {
		return strings.Trim(strings.TrimSpace(matches[1]), ".,;:)")
	}
	return ""
}

func firstTeamEvocationBriefPath(text string) string {
	if matches := teamEvocationBriefPathPattern.FindStringSubmatch(text); len(matches) == 2 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

func teamResearchHandoffMarkdown(request, teamID, briefPath string, contract map[string]any) string {
	var b strings.Builder
	b.WriteString("# Research And Council Handoff\n\n")
	b.WriteString("Soma retained this handoff so the research lane can prepare the evoked delivery team before implementation starts.\n\n")
	b.WriteString("## Operator request\n\n")
	b.WriteString(request)
	b.WriteString("\n\n")
	b.WriteString("## Linked team\n\n")
	b.WriteString("- Team ID: ")
	b.WriteString(teamID)
	b.WriteString("\n")
	b.WriteString("- Evocation brief: ")
	b.WriteString(briefPath)
	b.WriteString("\n\n")
	writeMarkdownList(&b, "Research lane responsibilities", []string{
		"Review available local/current sources and expose any unavailable external research boundary.",
		"Identify the output format, implementation stack, workstreams, and risks before build delegation.",
		"Name any specialist addition only when the needed capability, owned task, and removal point are clear.",
	})
	writeMarkdownList(&b, "Delivery lane responsibilities", []string{
		"Use this handoff and the evocation brief as the execution contract context.",
		"Produce the retained user-facing output package rather than another planning-only artifact.",
		"Return direct launch/view references, proof notes, and repair requests through Soma.",
	})
	writeMarkdownList(&b, "Expected outputs", confirmedActionStringSlice(contract["expected_outputs"]))
	writeMarkdownList(&b, "Acceptance criteria", confirmedActionStringSlice(contract["acceptance_criteria"]))
	writeMarkdownList(&b, "Proof required", confirmedActionStringSlice(contract["proof_required"]))
	b.WriteString("## Unknowns to expose\n\n")
	b.WriteString("- External web research availability must be stated before relying on current public information.\n")
	b.WriteString("- Any unsupported media, executable, host, or browser capability must become a recovery note instead of a silent omission.\n")
	return b.String()
}

func teamEvocationDelegationGoal(request string, contract map[string]any) string {
	outputs := confirmedActionStringSlice(contract["expected_outputs"])
	if len(outputs) == 0 {
		return "Use the retained research/council handoff to produce the requested deliverable with proof."
	}
	return fmt.Sprintf("Use the retained research/council handoff to produce %s for the operator request: %s", strings.Join(outputs, ", "), request)
}

func teamEvocationDelegationConstraints() []string {
	return []string{
		"Do not return another planning-only response as the final deliverable.",
		"Keep team-generated internal scratch separate from user-facing retained outputs.",
		"Provide a direct launch, view, or open path for every user-facing deliverable.",
		"Report validation defects back through Soma as repair work instead of silently editing outside the approved scope.",
	}
}

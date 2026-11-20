package server

import (
	"fmt"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

func ensureTeamMembersMapHasReviewer(raw any) map[string]any {
	resolved := map[string]any{}
	hasReviewer := false
	if source, ok := raw.(map[string]any); ok && len(source) > 0 {
		for role, agent := range source {
			resolved[role] = agent
			if hasIdealReviewerToken(role) || hasIdealReviewerToken(firstNonEmptyString(agent)) {
				hasReviewer = true
			}
		}
	} else {
		resolved["architect"] = "council-architect"
		resolved["developer"] = "prime-development-agent"
		resolved["designer"] = "agui-design-architect-agent"
	}
	if !hasReviewer {
		resolved["delivery-reviewer"] = "council-sentry"
	}
	return resolved
}

func isComplexProjectReviewTeamNeeded(latestRequest string, primaryCall protocol.PlannedToolCall) bool {
	teamID := strings.TrimSpace(firstNonEmptyString(primaryCall.Arguments["team_id"]))
	if teamID == "" || strings.HasSuffix(teamID, "-review-team") {
		return false
	}
	lower := strings.ToLower(strings.TrimSpace(latestRequest))
	if lower == "" {
		return false
	}
	if isTeamExecutionBundleRequest(lower) || isTetrisGameBundleRequest(lower) {
		return true
	}
	return requestContainsAny(lower, []string{
		"complex", "review team", "dedicated review", "architecture", "multi-team",
		"full-stack", "production", "game", "pipeline", "blueprint", "deliverable",
	})
}

func buildCompanionReviewTeamCall(primaryCall protocol.PlannedToolCall) protocol.PlannedToolCall {
	primaryTeamID := firstNonEmptyString(primaryCall.Arguments["team_id"], "mission-delivery-team")
	primaryTeamName := firstNonEmptyString(primaryCall.Arguments["name"], titleCaseTeamID(primaryTeamID))
	reviewTeamID := primaryTeamID + "-review-team"
	reviewTeamName := primaryTeamName + " — Delivery Review Team"
	primaryResultSubject := "swarm.team." + primaryTeamID + ".signal.result"

	return protocol.PlannedToolCall{
		Name: "create_team",
		Arguments: map[string]any{
			"team_id":     reviewTeamID,
			"name":        reviewTeamName,
			"description": fmt.Sprintf("Dedicated review gate team for %s. Listens on %s, audits OutputProofEnvelope SHA-256 readback, and certifies expected team delivery.", primaryTeamName, primaryResultSubject),
			"inputs":      []string{primaryResultSubject, "swarm.team." + reviewTeamID + ".internal.command"},
			"deliveries":  []string{"swarm.team." + reviewTeamID + ".signal.result", "swarm.team." + reviewTeamID + ".signal.status"},
			"members": map[string]any{
				"lead-reviewer":        "council-sentry",
				"architecture-auditor": "council-architect",
				"qa-verifier":          "agui-design-architect-agent",
			},
			"tools": []string{"read_file", "publish_signal"},
		},
	}
}

func teamExecutionBundleCalls(text string) []protocol.PlannedToolCall {
	teamID, teamName := extractTargetTeamFromRequest(text)
	reviewTeamID := teamID + "-review-team"
	cmdSub := "swarm.team." + teamID + ".internal.command"
	resSub := "swarm.team." + teamID + ".signal.result"
	revSub := "swarm.team." + reviewTeamID + ".signal.result"
	specs := []deliverableBundleFileSpec{
		{
			name: "team_execution_plan.md", title: teamName + " — Team Execution Plan", contentType: "text/markdown",
			content: "# " + teamName + " — Team Execution Plan\n\n## Mission Goal\n" + text + "\n\n## Specialist Role Assignments\n- **Lead Architect (`council-architect`):** Decomposes scope and validates write-isolation partitions.\n- **Delivery Specialist (`prime-development-agent`):** Executes next steps and produces verifiable workspace artifacts.\n- **Ideal Delivery Reviewer (`council-sentry`):** In-team verification gate auditing SHA-256 readback and OutputProofEnvelope gates.\n- **Dedicated Companion Review Team (`" + reviewTeamID + "`):** Listens on `" + resSub + "` and publishes final review certification on `" + revSub + "`.\n\n## Governed NATS Channels\n- **Command Input:** `" + cmdSub + "`\n- **Verified Result:** `" + resSub + "`\n- **Review Team Gate:** `" + revSub + "`\n",
		},
		{
			name: "next_steps_action_matrix.md", title: teamName + " — Next Steps & Action Matrix", contentType: "text/markdown",
			content: "# " + teamName + " — Next Steps & Action Matrix\n\n| Step | Owner Role | Verification Gate | Expected Output |\n| --- | --- | --- | --- |\n| 1. Scope & Channel Alignment | `council-architect` | NATS subject contract verified | `team_execution_plan.md` |\n| 2. Deliverable Execution | `prime-development-agent` | Disjoint write-isolation pass | `next_steps_action_matrix.md` |\n| 3. In-Team Proof & Handoff Release | `council-sentry` (`delivery_reviewer`) | SHA-256 `OutputProofEnvelope` verified | `team_handoff_summary.json`, `proof.md` |\n| 4. Dedicated Review Team Sign-Off | `" + reviewTeamID + "` | Cross-team delivery certification on `" + revSub + "` | Certified `DeliveryPackage` |\n",
		},
		{
			name: "team_handoff_summary.json", title: teamName + " — Channel & Handoff Summary (JSON)", contentType: "application/json",
			content: "{\n  \"team_id\": " + strconvQuoteSafe(teamID) + ",\n  \"team_name\": " + strconvQuoteSafe(teamName) + ",\n  \"reviewer_role\": \"delivery_reviewer\",\n  \"reviewer_agent\": \"council-sentry\",\n  \"review_team_id\": " + strconvQuoteSafe(reviewTeamID) + ",\n  \"command_channel\": " + strconvQuoteSafe(cmdSub) + ",\n  \"result_channel\": " + strconvQuoteSafe(resSub) + ",\n  \"review_gate_channel\": " + strconvQuoteSafe(revSub) + ",\n  \"status\": \"verified\",\n  \"next_steps\": [\n    \"Execute scoped team deliverables under disjoint file ownership\",\n    \"Publish verified OutputProofEnvelope result on " + resSub + "\",\n    \"Certify final package via dedicated review team " + reviewTeamID + " on " + revSub + "\"\n  ]\n}\n",
		},
	}
	calls := buildDeliverableBundleCalls("deliverables/teams/"+teamID, "team_execution_bundle", teamID, specs)
	for i := range calls {
		calls[i].Arguments["team_id"] = teamID
		calls[i].Arguments["team_name"] = teamName
	}
	return calls
}

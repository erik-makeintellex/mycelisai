package server

import "strings"

func teamEvocationForRequest(request string, contract map[string]any) map[string]any {
	lower := strings.ToLower(strings.TrimSpace(request))
	if !requestNeedsTeamEvocationReview(lower, contract) {
		return map[string]any{
			"mode":            "simple_lead_start",
			"staffing_policy": "Start with one lead and add specialists only after the lead names a concrete missing capability.",
		}
	}
	return map[string]any{
		"mode":                    "research_council_then_staff",
		"research_required":       true,
		"council_review_required": true,
		"research_instruction":    "Use configured search/local-source research when available; if unavailable, expose the boundary and continue from retained/local context without inventing facts.",
		"council_instruction":     "Before implementation, review the requested domain, delivery format, stack/package options, risks, proof needs, and the smallest useful team composition.",
		"staffing_policy":         "Begin with a lead. Add or request temporary specialists only after research/council review identifies a named workstream, owned deliverable, proof expectation, and removal point.",
		"role_questions": []string{
			"What domain expertise is required for this output?",
			"What implementation stack or output format best fits the operator's deployment target?",
			"What proof would convince the operator that the output works?",
			"What specialist roles are necessary now versus later?",
		},
		"suggested_workstreams": suggestedWorkstreamsForContentContract(contract),
	}
}

func requestNeedsTeamEvocationReview(lower string, contract map[string]any) bool {
	if len(confirmedActionStringSlice(contract["content_types"])) == 1 &&
		containsToolName(confirmedActionStringSlice(contract["content_types"]), "work_product") &&
		!contentContractNeedsResearch(contract) {
		return requestContainsAny(lower, []string{
			"complex", "advanced", "production", "deploy", "deployable", "executable", "application", "app", "package",
			"detailed", "substantial", "hard", "internet", "look up", "latest", "research on", "research the", "research available", "team leads", "specialists", "architecture", "multi-team",
		})
	}
	if requestContainsAny(lower, []string{
		"complex", "advanced", "production", "deploy", "deployable", "executable", "application", "app", "package",
		"detailed", "substantial", "hard", "internet", "look up", "latest", "research on", "research the", "research available", "team leads", "specialists", "architecture", "multi-team",
	}) {
		return true
	}
	return len(confirmedActionStringSlice(contract["content_types"])) > 1
}

func suggestedWorkstreamsForContentContract(contract map[string]any) []string {
	workstreams := []string{"delivery lead", "research/review lead", "proof/QA lead"}
	for _, kind := range confirmedActionStringSlice(contract["content_types"]) {
		switch kind {
		case "game":
			workstreams = append(workstreams, "interaction/design lead", "implementation lead", "audio/media lead")
		case "media":
			workstreams = append(workstreams, "visual/media lead", "prompt/constraint reviewer")
		case "text":
			workstreams = append(workstreams, "editorial lead", "source/claims reviewer")
		}
	}
	return uniqueOrderedTools(workstreams)
}

func contentContractNeedsResearch(contract map[string]any) bool {
	for _, item := range confirmedActionStringSlice(contract["team_preparation"]) {
		if strings.Contains(strings.ToLower(item), "research") {
			return true
		}
	}
	return false
}

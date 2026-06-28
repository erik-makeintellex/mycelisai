package server

import (
	"fmt"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

func inferTeamPreparationBriefPlanFromRequest(text string, teamCall protocol.PlannedToolCall) (protocol.PlannedToolCall, bool) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return protocol.PlannedToolCall{}, false
	}
	contract := mapArgument(teamCall.Arguments["content_contract"])
	if len(contract) == 0 {
		contract = contentContractForTeamRequest(trimmed)
	}
	evocation := mapArgument(teamCall.Arguments["team_evocation"])
	if len(evocation) == 0 {
		evocation = teamEvocationForRequest(trimmed, contract)
	}
	if !teamPreparationBriefIsUseful(trimmed, contract, evocation) {
		return protocol.PlannedToolCall{}, false
	}

	teamID := firstNonEmptyString(teamCall.Arguments["team_id"], teamCall.Arguments["id"], teamCall.Arguments["team_name"])
	teamName := firstNonEmptyString(teamCall.Arguments["name"], teamID, "Soma Delivery Team")
	slug := slugID(firstNonEmptyString(teamID, teamName, "soma-delivery-team"))
	if slug == "" {
		slug = "soma-delivery-team"
	}
	folder := "workspace/generated/" + slug + "-team-evocation"
	if teamID != "" {
		if groupFolder := groupWorkspaceFolderForTeamID(teamID); groupFolder != "" {
			folder = groupFolder + "/planning"
		}
	}
	path := folder + "/TEAM_EVOCATION.md"

	return protocol.PlannedToolCall{
		Name: "write_file",
		Arguments: map[string]any{
			"path":                path,
			"content":             teamPreparationBriefMarkdown(trimmed, teamName, teamID, contract, evocation),
			"validation":          "Retained team-evocation brief must identify research/council needs, role boundaries, output contract, proof gates, and the next Soma-mediated action before implementation.",
			"content_contract":    contract,
			"team_evocation":      evocation,
			"acceptance_criteria": confirmedActionStringSlice(contract["acceptance_criteria"]),
			"proof_required":      confirmedActionStringSlice(contract["proof_required"]),
		},
	}, true
}

func teamPreparationBriefIsUseful(request string, contract, evocation map[string]any) bool {
	mode := strings.TrimSpace(fmt.Sprint(evocation["mode"]))
	if mode == "research_council_then_staff" {
		return true
	}
	lower := strings.ToLower(strings.TrimSpace(request))
	if len(confirmedActionStringSlice(contract["content_types"])) > 1 {
		return true
	}
	return requestContainsAny(lower, []string{"complex", "advanced", "production", "deploy", "deployable", "executable", "application", "app", "package", "detailed", "substantial", "hard", "latest", "look up", "research on", "research the", "multi-team"})
}

func teamPreparationBriefMarkdown(request, teamName, teamID string, contract, evocation map[string]any) string {
	var b strings.Builder
	b.WriteString("# Team Evocation Brief\n\n")
	b.WriteString("Soma retained this preparation brief so complex work starts with research, council review, role clarity, and proof instead of a hardcoded domain template.\n\n")
	b.WriteString("## Operator request\n\n")
	b.WriteString(request)
	b.WriteString("\n\n")
	b.WriteString("## Team\n\n")
	b.WriteString("- Name: ")
	b.WriteString(firstNonEmptyString(teamName, "Soma Delivery Team"))
	b.WriteString("\n")
	if strings.TrimSpace(teamID) != "" {
		b.WriteString("- ID: ")
		b.WriteString(strings.TrimSpace(teamID))
		b.WriteString("\n")
	}
	b.WriteString("- Mode: ")
	b.WriteString(firstNonEmptyString(fmt.Sprint(evocation["mode"]), "simple_lead_start"))
	b.WriteString("\n\n")
	writeMarkdownList(&b, "Content types", confirmedActionStringSlice(contract["content_types"]))
	writeMarkdownList(&b, "Expected outputs", confirmedActionStringSlice(contract["expected_outputs"]))
	writeMarkdownList(&b, "Research and council preparation", confirmedActionStringSlice(contract["team_preparation"]))
	writeMarkdownList(&b, "Suggested workstreams", confirmedActionStringSlice(evocation["suggested_workstreams"]))
	writeMarkdownList(&b, "Acceptance criteria", confirmedActionStringSlice(contract["acceptance_criteria"]))
	writeMarkdownList(&b, "Proof required", confirmedActionStringSlice(contract["proof_required"]))
	b.WriteString("## Next Soma-mediated action\n\n")
	b.WriteString("Have Soma use this brief to research available context, confirm the implementation strategy, staff only the needed specialists, and then ask for approval before producing the final deliverable.\n")
	return b.String()
}

func writeMarkdownList(b *strings.Builder, title string, items []string) {
	b.WriteString("## ")
	b.WriteString(title)
	b.WriteString("\n\n")
	if len(items) == 0 {
		b.WriteString("- Not required for this request.\n\n")
		return
	}
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		b.WriteString("- ")
		b.WriteString(trimmed)
		b.WriteString("\n")
	}
	b.WriteString("\n")
}

func mapArgument(value any) map[string]any {
	if value == nil {
		return nil
	}
	if mapped, ok := value.(map[string]any); ok {
		return mapped
	}
	return nil
}

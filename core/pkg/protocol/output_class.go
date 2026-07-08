package protocol

import "strings"

type OutputClass string

const (
	OutputClassUserDeliverable OutputClass = "user_deliverable"
	OutputClassPlanning        OutputClass = "planning"
	OutputClassProof           OutputClass = "proof"
	OutputClassInternalHandoff OutputClass = "internal_handoff"
	OutputClassSourceMaterial  OutputClass = "source_material"
)

func NormalizeOutputClass(value string) OutputClass {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "user_deliverable", "deliverable", "output", "final":
		return OutputClassUserDeliverable
	case "planning", "plan", "draft":
		return OutputClassPlanning
	case "proof", "evidence", "trust":
		return OutputClassProof
	case "internal", "internal_handoff", "handoff":
		return OutputClassInternalHandoff
	case "source", "support", "source_material", "supporting_material":
		return OutputClassSourceMaterial
	default:
		return ""
	}
}

func InferOutputClass(kind string, refs ...string) OutputClass {
	for _, ref := range refs {
		class := inferOutputClassFromPath(ref)
		if class != "" {
			return class
		}
	}
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "team", "text_reply", "tool_result", "mcp_tool_result":
		return OutputClassInternalHandoff
	default:
		return OutputClassUserDeliverable
	}
}

func IsUserDeliverableOutputClass(value string) bool {
	return NormalizeOutputClass(value) == OutputClassUserDeliverable
}

func inferOutputClassFromPath(value string) OutputClass {
	path := strings.ToLower(strings.TrimSpace(value))
	path = strings.ReplaceAll(path, "\\", "/")
	path = strings.Trim(path, "/")
	if path == "" {
		return ""
	}
	parts := strings.Split(path, "/")
	base := parts[len(parts)-1]
	if strings.Contains(path, "/planning/") ||
		base == "team_evocation.md" ||
		base == "research_council_handoff.md" {
		return OutputClassPlanning
	}
	if strings.Contains(path, "/proof/") || base == "proof.md" {
		return OutputClassProof
	}
	if strings.Contains(path, "/source/") ||
		strings.Contains(path, "/support/") ||
		strings.Contains(path, "/watch/") {
		return OutputClassSourceMaterial
	}
	return ""
}

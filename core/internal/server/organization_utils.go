package server

import (
	"fmt"
	"strings"
)

func compactOutputSlice(targetOutputs []string, index int) []string {
	if len(targetOutputs) == 0 {
		return []string{}
	}
	if index < 0 {
		index = 0
	}
	if index >= len(targetOutputs) {
		index = len(targetOutputs) - 1
	}
	return []string{targetOutputs[index]}
}

func humanJoin(items []string) string {
	items = normalizeExecutionCapabilityList(items)
	switch len(items) {
	case 0:
		return "planned outputs"
	case 1:
		return items[0]
	case 2:
		return items[0] + " and " + items[1]
	default:
		return strings.Join(items[:len(items)-1], ", ") + ", and " + items[len(items)-1]
	}
}

func maxInt(left, right int) int {
	if left > right {
		return left
	}
	return right
}

func normalizeExecutionCapabilityList(items []string) []string {
	if len(items) == 0 {
		return []string{}
	}
	out := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}

func referencesImageTeamOutput(normalized string) bool {
	return strings.Contains(normalized, "image") ||
		strings.Contains(normalized, "hero art") ||
		strings.Contains(normalized, "hero image") ||
		strings.Contains(normalized, "visual") ||
		strings.Contains(normalized, "illustration") ||
		strings.Contains(normalized, "poster") ||
		strings.Contains(normalized, "moodboard")
}

func referencesExternalWorkflowContract(normalized string) bool {
	return strings.Contains(normalized, "n8n") ||
		strings.Contains(normalized, "workflow contract") ||
		strings.Contains(normalized, "external workflow") ||
		strings.Contains(normalized, "comfyui")
}

func externalWorkflowTarget(normalized string) string {
	switch {
	case strings.Contains(normalized, "n8n"):
		return "n8n workflow contract"
	case strings.Contains(normalized, "comfyui"):
		return "ComfyUI workflow contract"
	default:
		return "External workflow contract"
	}
}

func templateSpecificSuggestion(home OrganizationHomePayload) string {
	if home.StartMode == OrganizationStartModeTemplate && strings.TrimSpace(home.TemplateName) != "" {
		return fmt.Sprintf("Use the %s starter as the first operating guide.", home.TemplateName)
	}
	return "Review how the Team Lead should shape the first Department and Specialist setup."
}

func formatConfiguredCountForGuidance(count int, label string) string {
	if count == 0 {
		return "not configured yet"
	}
	return fmt.Sprintf("%d %s%s ready", count, label, pluralSuffix(count))
}

func spreadSpecialists(total, departmentCount, index int) int {
	if departmentCount <= 0 || total <= 0 {
		return 0
	}
	base := total / departmentCount
	remainder := total % departmentCount
	if index < remainder {
		return base + 1
	}
	return base
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func safeOrganizationName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "this AI Organization"
	}
	return name
}

func safeTeamLeadLabel(label string) string {
	label = strings.TrimSpace(label)
	if label == "" {
		return "Team Lead"
	}
	return label
}

func safePurposeText(purpose string) string {
	purpose = strings.TrimSpace(purpose)
	if purpose == "" {
		return "the current AI Organization priorities"
	}
	return purpose
}

func pluralSuffix(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}

func (h OrganizationHomePayload) startingPointLabel() string {
	if h.StartMode == OrganizationStartModeTemplate && strings.TrimSpace(h.TemplateName) != "" {
		return h.TemplateName
	}
	return "starting organization shape"
}

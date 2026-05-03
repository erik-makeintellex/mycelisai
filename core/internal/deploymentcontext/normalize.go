package deploymentcontext

import "strings"

func normalizeSourceKind(raw string) string {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "workspace_file":
		return "workspace_file"
	case "web_research":
		return "web_research"
	case "user_note":
		return "user_note"
	case "user_record":
		return "user_record"
	case "diary_entry":
		return "diary_entry"
	case "finance_record":
		return "finance_record"
	case "lesson":
		return "lesson"
	case "inferred_pattern":
		return "inferred_pattern"
	case "contradiction":
		return "contradiction"
	case "trajectory_shift":
		return "trajectory_shift"
	case "meta_observation":
		return "meta_observation"
	case "synthesis_note":
		return "synthesis_note"
	default:
		return "user_document"
	}
}

func normalizeKnowledgeClass(raw string) string {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case KnowledgeClassCompanyKnowledge:
		return KnowledgeClassCompanyKnowledge
	case KnowledgeClassSomaOperating:
		return KnowledgeClassSomaOperating
	case KnowledgeClassUserPrivate:
		return KnowledgeClassUserPrivate
	case KnowledgeClassReflection:
		return KnowledgeClassReflection
	default:
		return KnowledgeClassCustomerContext
	}
}

func normalizeContentDomain(raw string) string {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "private_records":
		return "private_records"
	case "diary":
		return "diary"
	case "finance":
		return "finance"
	case "health":
		return "health"
	case "legal":
		return "legal"
	case "creative":
		return "creative"
	case "operations":
		return "operations"
	case "reflection":
		return "reflection"
	default:
		return ""
	}
}

func normalizeSomaContextKind(raw string) string {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "output_specificity":
		return "output_specificity"
	case "operating_stance":
		return "operating_stance"
	case "identity":
		return "identity"
	case "policy":
		return "policy"
	default:
		return ""
	}
}

func normalizeOutputSpecificity(raw string) string {
	value := strings.TrimSpace(strings.ToLower(raw))
	switch value {
	case "concise", "balanced", "detailed", "executive":
		return value
	default:
		return ""
	}
}

func normalizeSourceLabel(raw string) string {
	label := strings.TrimSpace(raw)
	if label == "" {
		return "operator provided"
	}
	return label
}

func normalizeVisibility(raw string) string {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "private":
		return "private"
	case "team":
		return "team"
	default:
		return "global"
	}
}

func normalizeSensitivityClass(raw string) string {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "restricted":
		return "restricted"
	case "team_scoped":
		return "team_scoped"
	default:
		return "role_scoped"
	}
}

func normalizeTrustClass(raw string) string {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "validated_external":
		return "validated_external"
	case "bounded_external":
		return "bounded_external"
	case "trusted_internal":
		return "trusted_internal"
	default:
		return "user_provided"
	}
}

func normalizeContentType(raw string) string {
	value := strings.TrimSpace(strings.ToLower(raw))
	if value == "" {
		return "text/markdown"
	}
	return value
}

func normalizeTags(tags []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(tags))
	for _, raw := range tags {
		tag := strings.TrimSpace(strings.ToLower(raw))
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		out = append(out, tag)
	}
	return out
}

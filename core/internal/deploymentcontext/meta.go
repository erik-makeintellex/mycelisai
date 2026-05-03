package deploymentcontext

import "strings"

func stringMeta(meta map[string]any, key, fallback string) string {
	if meta == nil {
		return fallback
	}
	if value, ok := meta[key].(string); ok && strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	return fallback
}

func intMeta(meta map[string]any, key string, fallback int) int {
	if meta == nil {
		return fallback
	}
	switch value := meta[key].(type) {
	case float64:
		return int(value)
	case int:
		return value
	default:
		return fallback
	}
}

func stringSliceMeta(meta map[string]any, key string) []string {
	if meta == nil {
		return nil
	}
	raw, ok := meta[key]
	if !ok {
		return nil
	}
	switch value := raw.(type) {
	case []any:
		out := make([]string, 0, len(value))
		for _, item := range value {
			if text, ok := item.(string); ok && strings.TrimSpace(text) != "" {
				out = append(out, strings.TrimSpace(text))
			}
		}
		return out
	case []string:
		return normalizeTags(value)
	default:
		return nil
	}
}

package searchcap

import "strings"

func slugSourceID(raw string) string {
	var b strings.Builder
	lastUnderscore := false
	for _, r := range strings.ToLower(strings.TrimSpace(raw)) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastUnderscore = false
		case r == '_' || r == '-' || r == ' ':
			if !lastUnderscore && b.Len() > 0 {
				b.WriteByte('_')
				lastUnderscore = true
			}
		}
	}
	return strings.Trim(b.String(), "_")
}

func cloneSources(sources []Source) []Source {
	if len(sources) == 0 {
		return []Source{}
	}
	out := make([]Source, len(sources))
	copy(out, sources)
	return out
}

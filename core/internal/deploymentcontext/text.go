package deploymentcontext

import (
	"fmt"
	"strings"
)

func buildEmbeddingText(knowledgeClass, title, sourceLabel, sourceKind, chunk string, chunkIndex, chunkCount int) string {
	return fmt.Sprintf(
		"[%s] title=%s source=%s source_kind=%s chunk=%d/%d\n%s",
		knowledgeClass,
		title,
		sourceLabel,
		sourceKind,
		chunkIndex,
		chunkCount,
		chunk,
	)
}

func chunkText(content string, maxRunes, overlapRunes int) []string {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil
	}
	if maxRunes <= 0 {
		maxRunes = defaultChunkSize
	}
	if overlapRunes < 0 {
		overlapRunes = 0
	}

	runes := []rune(content)
	if len(runes) <= maxRunes {
		return []string{content}
	}

	var chunks []string
	for start := 0; start < len(runes); {
		end := start + maxRunes
		if end > len(runes) {
			end = len(runes)
		}

		if end < len(runes) {
			window := string(runes[start:end])
			if split := strings.LastIndex(window, "\n\n"); split >= maxRunes/2 {
				end = start + split
			} else if split := strings.LastIndex(window, "\n"); split >= maxRunes/2 {
				end = start + split
			} else if split := strings.LastIndex(window, ". "); split >= maxRunes/2 {
				end = start + split + 1
			}
		}

		chunk := strings.TrimSpace(string(runes[start:end]))
		if chunk != "" {
			chunks = append(chunks, chunk)
		}

		if end >= len(runes) {
			break
		}

		nextStart := end - overlapRunes
		if nextStart <= start {
			nextStart = end
		}
		start = nextStart
	}

	return chunks
}

func previewContent(content string, maxRunes int) string {
	content = strings.TrimSpace(content)
	if content == "" {
		return ""
	}
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\n", " ")
	content = strings.Join(strings.Fields(content), " ")
	runes := []rune(content)
	if len(runes) <= maxRunes {
		return content
	}
	return strings.TrimSpace(string(runes[:maxRunes])) + "..."
}

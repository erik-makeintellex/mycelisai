package server

import (
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

func inferResponseDepthFromRequest(text string, isMutation bool) protocol.ResponseDepth {
	if isMutation {
		return protocol.ResponseDepthExecutionProposal
	}
	lower := strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(text)), " "))
	if lower == "" {
		return protocol.ResponseDepthStructuredSummary
	}
	if asksForDecisionBrief(lower) {
		return protocol.ResponseDepthDecisionBrief
	}
	if asksForStructuredSummary(lower) {
		return protocol.ResponseDepthStructuredSummary
	}
	if asksForQuickBox(lower) {
		return protocol.ResponseDepthQuickBox
	}
	return protocol.ResponseDepthStructuredSummary
}

func asksForQuickBox(lower string) bool {
	if requestContainsAny(lower, []string{
		"top 5", "top five", "just show", "just list", "only links", "show links", "list the sources",
		"source list", "quick list", "quick answer", "short answer", "briefly", "in brief", "give me a table",
	}) {
		return true
	}
	for _, word := range []string{"links", "sources", "list", "table", "options"} {
		if hasExactWord(lower, word) {
			return true
		}
	}
	return false
}

func asksForStructuredSummary(lower string) bool {
	return requestContainsAny(lower, []string{
		"summarize", "summary", "compare", "comparison", "main patterns", "patterns", "what's notable",
		"what is notable", "breakdown", "explain", "overview", "details", "detailed", "walk me through",
	})
}

func asksForDecisionBrief(lower string) bool {
	return requestContainsAny(lower, []string{
		"what should we do", "what should i do", "recommend", "recommendation", "which one should",
		"which should", "how does this affect", "how should this change", "what's your take",
		"what is your take", "decision brief", "pros and cons", "tradeoffs", "risks",
	})
}

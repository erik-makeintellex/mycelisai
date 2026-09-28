package server

import (
	"strings"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

// mrkUnwrapAmbientRecallQuery mirrors swarm.recallQuery's unwrap loop
// (core/internal/swarm/internal_tools_memory_context.go) using the shared
// protocol.ChatOriginalRequestMarker constant. The two implementations stay
// independent (server must not import swarm's internal package for this),
// but both must agree on where the operator's original request starts
// inside a wrapped chat turn. If this constant's wording or the wrap shape
// ever drifts between the two packages, this test is the one that catches
// it from the server side.
func mrkUnwrapAmbientRecallQuery(input string) string {
	query := strings.TrimSpace(input)
	for strings.HasPrefix(query, "[") {
		idx := strings.Index(query, protocol.ChatOriginalRequestMarker)
		if idx < 0 {
			break
		}
		rest := query[idx+len(protocol.ChatOriginalRequestMarker):]
		if rest != "" && rest[0] != '\n' {
			break
		}
		query = strings.TrimSpace(rest)
	}
	return query
}

// TestMrkAmbientRecallRecoversOriginalRequestFromWrappedChatTurn builds the
// latest user turn through the real wrapper functions, stacked exactly as
// POST /api/v1/chat does it (cognitive_chat_handler.go: normalizeChatRequestMessages
// then applyGovernanceProfileToLatestMessage), and proves ambient recall's
// unwrap logic recovers the operator's exact original request from the real
// doubly-wrapped text -- not a hand-typed stand-in for it.
func TestMrkAmbientRecallRecoversOriginalRequestFromWrappedChatTurn(t *testing.T) {
	ask := "Write a two-line promo for our cardamom knot weekend special."
	messages := []chatRequestMessage{{Role: "user", Content: ask}}

	normalized, mutationTools := normalizeChatRequestMessages(messages)
	if len(mutationTools) != 0 {
		t.Fatalf("expected a direct-answer route with no mutation tools, got %v", mutationTools)
	}

	profile := defaultUserGovernanceProfile("owner")
	wrapped := applyGovernanceProfileToLatestMessage(normalized, profile)
	wrappedContent := wrapped[latestUserMessageIndex(wrapped)].Content

	// This is the exact invariant ambient recall depends on: the wrapped
	// turn must end in the shared marker immediately followed by the
	// untouched original ask.
	if !strings.HasSuffix(wrappedContent, protocol.ChatOriginalRequestMarker+"\n"+ask) {
		t.Fatalf("wrapped turn does not end with marker + ask:\n%s", wrappedContent)
	}

	if got := mrkUnwrapAmbientRecallQuery(wrappedContent); got != ask {
		t.Fatalf("ambient recall did not recover the operator's exact request: got %q want %q", got, ask)
	}
}

// TestMrkAmbientRecallRecoversOriginalRequestFromGovernedMutationRoute covers
// the governed-mutation route header (the other branch of
// normalizeChatRequestMessages), stacked with the governance profile header
// the same way the chat handler does.
func TestMrkAmbientRecallRecoversOriginalRequestFromGovernedMutationRoute(t *testing.T) {
	ask := "Create a team named Bakery Ops and delegate the promo task to it."
	messages := []chatRequestMessage{{Role: "user", Content: ask}}

	normalized, mutationTools := normalizeChatRequestMessages(messages)
	if len(mutationTools) == 0 {
		t.Fatalf("expected inferred mutation tools for a team-creation/delegation request")
	}

	profile := defaultUserGovernanceProfile("operator")
	wrapped := applyGovernanceProfileToLatestMessage(normalized, profile)
	wrappedContent := wrapped[latestUserMessageIndex(wrapped)].Content

	if got := mrkUnwrapAmbientRecallQuery(wrappedContent); got != ask {
		t.Fatalf("ambient recall did not recover the operator's exact request: got %q want %q", got, ask)
	}
}

// TestMrkAmbientRecallRecoversOriginalRequestFromRetryRoute covers the
// direct-answer retry header (applyDirectAnswerRetryInstruction), the third
// site that writes the shared marker.
func TestMrkAmbientRecallRecoversOriginalRequestFromRetryRoute(t *testing.T) {
	ask := "Summarize last week's deployment incidents."
	messages := []chatRequestMessage{{Role: "user", Content: "try again"}}

	retried := applyDirectAnswerRetryInstruction(messages, ask)
	profile := defaultUserGovernanceProfile("reviewer")
	wrapped := applyGovernanceProfileToLatestMessage(retried, profile)
	wrappedContent := wrapped[latestUserMessageIndex(wrapped)].Content

	if got := mrkUnwrapAmbientRecallQuery(wrappedContent); got != ask {
		t.Fatalf("ambient recall did not recover the operator's exact request: got %q want %q", got, ask)
	}
}

// TestMrkAmbientRecallDoesNotOverstripBracketedRequestText is the adversarial
// case: an operator request that itself starts with a bracketed tag must not
// be stripped further once the wrapper headers are gone, because no
// remaining marker exists past that point.
func TestMrkAmbientRecallDoesNotOverstripBracketedRequestText(t *testing.T) {
	ask := "[urgent] Restart the ingest worker and confirm the queue drains."
	messages := []chatRequestMessage{{Role: "user", Content: ask}}

	normalized, _ := normalizeChatRequestMessages(messages)
	profile := defaultUserGovernanceProfile("owner")
	wrapped := applyGovernanceProfileToLatestMessage(normalized, profile)
	wrappedContent := wrapped[latestUserMessageIndex(wrapped)].Content

	if got := mrkUnwrapAmbientRecallQuery(wrappedContent); got != ask {
		t.Fatalf("ambient recall over-stripped a request that itself starts with a bracket: got %q want %q", got, ask)
	}
}

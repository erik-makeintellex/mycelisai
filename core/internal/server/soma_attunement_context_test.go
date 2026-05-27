package server

import (
	"strings"
	"testing"

	"github.com/mycelis/core/internal/searchcap"
)

func TestBuildSomaAttunementContextForMediaTeamNamesOutputAndProof(t *testing.T) {
	context := buildSomaAttunementContext(
		"Have Soma create a comic page team with artist, character designer, dialogue writer, and generate the image.",
		searchcap.Status{Provider: searchcap.ProviderLocalSources, Enabled: true, SupportsLocalSources: true},
	)
	if !strings.Contains(context, somaAttunementContextHeader) {
		t.Fatalf("context missing header: %q", context)
	}
	if !strings.Contains(context, "retained visual/media deliverable") {
		t.Fatalf("context did not infer media output: %q", context)
	}
	if !strings.Contains(context, "output path, and proof") {
		t.Fatalf("context did not preserve proof/output expectation: %q", context)
	}
	if !strings.Contains(context, "prefer local Mycelis context") {
		t.Fatalf("context did not prefer local knowledge: %q", context)
	}
}

func TestBuildSomaAttunementContextForFreshResearchDisclosesSearchBoundary(t *testing.T) {
	context := buildSomaAttunementContext(
		"Research the latest local media generation tools and write a sourced report.",
		searchcap.Status{Provider: searchcap.ProviderSearXNG, Enabled: true, SupportsPublicWeb: true},
	)
	if !strings.Contains(context, "sourced report or investigation package") {
		t.Fatalf("context did not infer report output: %q", context)
	}
	if !strings.Contains(context, "disclose the source boundary as searxng") {
		t.Fatalf("context did not name search source boundary: %q", context)
	}
	if !strings.Contains(context, "external results as leads") {
		t.Fatalf("context did not guard external interpretation: %q", context)
	}
}

func TestPrependSomaAttunementContextKeepsLatestUserMessageStable(t *testing.T) {
	messages := []chatRequestMessage{
		{Role: "user", Content: "Build a playable browser game."},
	}
	got := prependSomaAttunementContext(messages, messages[0].Content, searchcap.Status{})
	if len(got) != 2 {
		t.Fatalf("message count = %d, want 2", len(got))
	}
	if got[0].Role != "system" || !strings.Contains(got[0].Content, somaAttunementContextHeader) {
		t.Fatalf("first message = %+v, want system attunement context", got[0])
	}
	if latest := latestUserMessageContent(got); latest != "Build a playable browser game." {
		t.Fatalf("latest user message = %q", latest)
	}
}

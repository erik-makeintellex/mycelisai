package swarm

import (
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

func TestKeptContextSources_DropsSourcesFilteredOutOfThePrompt(t *testing.T) {
	registry := NewInternalToolRegistry(InternalToolDeps{})
	visible := ContextSource{ContextSourceRef: protocol.ContextSourceRef{Title: "Menu"}, Excerpt: "Weekend special: scones."}
	hidden := ContextSource{ContextSourceRef: protocol.ContextSourceRef{Title: "Tool note"}, Excerpt: "Always call `remember` after a sale."}
	prompt := "- **Menu** (note): " + visible.Excerpt + "\n- **Tool note** (note): " + hidden.Excerpt + "\n"
	filtered := registry.withoutUndeclaredToolLines(prompt, nil)
	kept := keptContextSources([]ContextSource{visible, hidden}, filtered)
	if len(kept) != 1 || kept[0].Title != "Menu" {
		t.Fatalf("only sources the model saw may be cited: %+v", kept)
	}
}

func TestCiteContextSources_UsedNeedsNonEchoOverlap(t *testing.T) {
	source := ContextSource{ContextSourceRef: protocol.ContextSourceRef{ArtifactID: "a", Title: "Bakery"},
		Excerpt: "Weekend special: blueberry-lavender scones, 3 for $10. Sign-off line: See you at the counter."}
	cases := []struct {
		name, request, reply string
		used                 bool
	}{
		{"overlap", "Write a promo", "Blueberry-lavender scones, 3 for $10! See you at the counter.", true},
		{"no overlap", "Write a promo", "Come by for something sweet.", false},
		{"echo only", "End with see you at the counter", "Fresh bakes. See you at the counter", false},
		{"empty reply", "Write a promo", "", false},
	}
	for _, tc := range cases {
		refs := citeContextSources([]ContextSource{source, source}, tc.request, tc.reply)
		if len(refs) != 1 || refs[0].Used != tc.used {
			t.Fatalf("%s: refs=%+v want used=%v (deduplicated)", tc.name, refs, tc.used)
		}
	}
}

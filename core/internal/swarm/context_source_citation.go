package swarm

import (
	"strings"
	"unicode"

	"github.com/mycelis/core/pkg/protocol"
)

const citationShingleWords = 4

// citeContextSources marks a source used only when the reply shares at least
// one normalized 4-word shingle with the excerpt the model saw that the user
// request did not already contain. Everything else stays "consulted".
func citeContextSources(sources []ContextSource, request, reply string) []protocol.ContextSourceRef {
	if len(sources) == 0 {
		return nil
	}
	requestShingles := shingles(request)
	replyShingles := shingles(reply)
	refs := make([]protocol.ContextSourceRef, 0, len(sources))
	seen := map[string]int{}
	for _, source := range sources {
		ref := source.ContextSourceRef
		ref.Used = false
		for shingle := range shingles(source.Excerpt) {
			if replyShingles[shingle] && !requestShingles[shingle] {
				ref.Used = true
				break
			}
		}
		key := ref.ArtifactID + "\x00" + ref.Title
		if index, ok := seen[key]; ok {
			refs[index].Used = refs[index].Used || ref.Used
			continue
		}
		seen[key] = len(refs)
		refs = append(refs, ref)
	}
	return refs
}

// keptContextSources drops sources whose excerpt did not survive prompt
// filtering, so a ref is only reported for text the model actually saw.
func keptContextSources(sources []ContextSource, prompt string) []ContextSource {
	kept := make([]ContextSource, 0, len(sources))
	for _, source := range sources {
		if strings.Contains(prompt, source.Excerpt) {
			kept = append(kept, source)
		}
	}
	return kept
}

func shingles(text string) map[string]bool {
	words := strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	out := map[string]bool{}
	for i := 0; i+citationShingleWords <= len(words); i++ {
		out[strings.Join(words[i:i+citationShingleWords], " ")] = true
	}
	return out
}

package swarm

import (
	"regexp"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

// minDistinctiveTokenMatches is how many shared ordinary distinctive tokens
// (not stopwords, not echoed from the request) are needed to call a source
// used. A single rare/compound unit (a hyphenated term, a price, or a
// "<n> for" quantity pattern) is specific enough to count on its own.
const minDistinctiveTokenMatches = 2

var (
	hyphenTokenRe     = regexp.MustCompile(`[a-z]+(?:-[a-z]+)+`)
	priceTokenRe      = regexp.MustCompile(`\$\d+(?:\.\d+)?`)
	quantityForRe     = regexp.MustCompile(`\b(\d+)\s+for\b`)
	forQuantityRe     = regexp.MustCompile(`\bfor\s+\$?(\d+)\b`)
	distinctiveWordRe = regexp.MustCompile(`[a-z]+`)
)

// commonWords are stopwords and very frequent English words that are too
// generic to count as evidence a reply actually drew on a source: shared
// only because both texts are ordinary English, not because they share a
// fact.
var commonWords = map[string]bool{
	"the": true, "a": true, "an": true, "and": true, "or": true, "to": true, "of": true,
	"in": true, "on": true, "for": true, "with": true, "is": true, "are": true, "was": true,
	"were": true, "be": true, "been": true, "being": true, "this": true, "that": true,
	"these": true, "those": true, "it": true, "its": true, "at": true, "as": true, "by": true,
	"from": true, "we": true, "our": true, "you": true, "your": true, "i": true, "he": true,
	"she": true, "they": true, "them": true, "his": true, "her": true, "their": true,
	"not": true, "no": true, "do": true, "does": true, "did": true, "have": true, "has": true,
	"had": true, "will": true, "would": true, "can": true, "could": true, "should": true,
	"may": true, "might": true, "must": true, "up": true, "down": true, "out": true,
	"over": true, "under": true, "again": true, "further": true, "then": true, "once": true,
	"here": true, "there": true, "when": true, "where": true, "why": true, "how": true,
	"all": true, "any": true, "both": true, "each": true, "few": true, "more": true,
	"most": true, "other": true, "some": true, "such": true, "only": true, "own": true,
	"same": true, "so": true, "than": true, "too": true, "very": true, "just": true,
	"but": true, "if": true, "because": true, "while": true, "about": true, "against": true,
	"between": true, "into": true, "through": true, "during": true, "before": true,
	"after": true, "above": true, "below": true, "now": true, "come": true, "try": true,
	"get": true, "got": true, "one": true, "two": true, "three": true, "day": true,
	"time": true, "way": true, "make": true, "made": true, "see": true,
	"end": true, "line": true, "off": true, "counter": true, "sign": true, "us": true,
	"me": true, "my": true, "mine": true, "also": true, "well": true, "good": true,
	"great": true, "new": true, "old": true, "want": true, "need": true, "like": true,
	"back": true, "away": true, "delicious": true, "sweet": true, "something": true,
}

// citeContextSources marks a source used when the reply shares distinctive
// evidence with the excerpt the model saw: at least two ordinary distinctive
// words, or one rare/compound unit (a hyphenated term, a price, or a
// quantity pattern such as "3 for"). Tokens the user's own request already
// contained are excluded so a source is never credited for an echo, and
// common stopwords never count as evidence. Everything else stays
// "consulted".
func citeContextSources(sources []ContextSource, request, reply string) []protocol.ContextSourceRef {
	if len(sources) == 0 {
		return nil
	}
	requestRare, requestWords := distinctiveUnits(request)
	replyRare, replyWords := distinctiveUnits(reply)
	refs := make([]protocol.ContextSourceRef, 0, len(sources))
	seen := map[string]int{}
	for _, source := range sources {
		ref := source.ContextSourceRef
		sourceRare, sourceWords := distinctiveUnits(source.Excerpt)
		ref.Used = sharesDistinctiveEvidence(sourceRare, sourceWords, replyRare, replyWords, requestRare, requestWords)
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

// sharesDistinctiveEvidence reports whether the source's distinctive units
// reappear in the reply without having simply been echoed from the request.
func sharesDistinctiveEvidence(sourceRare, sourceWords, replyRare, replyWords, requestRare, requestWords map[string]bool) bool {
	for unit := range sourceRare {
		if replyRare[unit] && !requestRare[unit] {
			return true
		}
	}
	matches := 0
	for word := range sourceWords {
		if replyWords[word] && !requestWords[word] {
			matches++
			if matches >= minDistinctiveTokenMatches {
				return true
			}
		}
	}
	return false
}

// distinctiveUnits splits text into two sets: "rare" compound units
// (hyphenated terms, prices, and "<n> for"/"for <n>" quantity patterns,
// keyed by the number so word order does not matter) that are specific
// enough to count alone, and ordinary distinctive words (length >= 3,
// letters only, not a common/stop word) that need corroboration.
func distinctiveUnits(text string) (rare map[string]bool, words map[string]bool) {
	lower := strings.ToLower(text)
	rare = map[string]bool{}
	for _, m := range hyphenTokenRe.FindAllString(lower, -1) {
		rare["hyphen:"+m] = true
	}
	for _, m := range priceTokenRe.FindAllString(lower, -1) {
		rare["price:"+m] = true
	}
	for _, m := range quantityForRe.FindAllStringSubmatch(lower, -1) {
		rare["qty:"+m[1]] = true
	}
	for _, m := range forQuantityRe.FindAllStringSubmatch(lower, -1) {
		rare["qty:"+m[1]] = true
	}
	words = map[string]bool{}
	for _, w := range distinctiveWordRe.FindAllString(lower, -1) {
		if len(w) < 3 || commonWords[w] {
			continue
		}
		words[w] = true
	}
	return rare, words
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

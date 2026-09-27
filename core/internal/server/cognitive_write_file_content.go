package server

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/mycelis/core/pkg/protocol"
)

const (
	draftPreviewMaxLines = 20
	draftPreviewMaxBytes = 1536
	// echoResidueMaxWords is how many words a draft may add around a verbatim
	// copy of the request and still count as an echo (for example a heading).
	echoResidueMaxWords = 4
)

// legacyPlaceholderSignature marks the removed request-echo template. Proposals
// minted before the fix can still carry it, so execution rejects it.
const legacyPlaceholderSignature = "soma created this file from your request"

// writeFileContentIsRequestEcho reports whether content is empty, a fragment of
// the request, the request plus a few framing words, or the legacy template.
func writeFileContentIsRequestEcho(content, request string) bool {
	body := echoWords(content)
	if body == "" {
		return true
	}
	if strings.Contains(body, legacyPlaceholderSignature) {
		return true
	}
	req := echoWords(request)
	if req == "" {
		return false
	}
	if strings.Contains(" "+req+" ", " "+body+" ") {
		return true
	}
	if !strings.Contains(body, req) {
		return false
	}
	residue := strings.Fields(strings.Replace(body, req, " ", -1))
	return len(residue) <= echoResidueMaxWords
}

// echoWords lowercases text and keeps only letters and digits as single-space
// separated words, so Markdown markers and punctuation do not hide an echo.
func echoWords(text string) string {
	fields := strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	return strings.Join(fields, " ")
}

// stripWrappingCodeFence removes one Markdown code fence that wraps the whole
// reply, which models add despite instructions. Inner fences are kept.
func stripWrappingCodeFence(text string) string {
	trimmed := strings.TrimSpace(text)
	if !strings.HasPrefix(trimmed, "```") || !strings.HasSuffix(trimmed, "```") || len(trimmed) < 6 {
		return trimmed
	}
	firstNewline := strings.Index(trimmed, "\n")
	if firstNewline < 0 {
		return trimmed
	}
	inner := strings.TrimSpace(trimmed[firstNewline+1 : len(trimmed)-3])
	if strings.Contains(inner, "```") {
		return trimmed
	}
	return inner
}

// buildProposalDraftPreview bounds the preview to the first 20 lines and 1.5KB,
// cut on a UTF-8 boundary. FullDraft is true only when nothing was cut.
func buildProposalDraftPreview(path, content string) protocol.ProposalDraftPreview {
	lines := strings.Split(strings.TrimRight(content, "\n"), "\n")
	preview := strings.Join(lines[:min(len(lines), draftPreviewMaxLines)], "\n")
	if len(preview) > draftPreviewMaxBytes {
		cut := draftPreviewMaxBytes
		for cut > 0 && !utf8.RuneStart(preview[cut]) {
			cut--
		}
		preview = preview[:cut]
	}
	return protocol.ProposalDraftPreview{
		Path:      path,
		Preview:   preview,
		FullDraft: preview == strings.TrimRight(content, "\n"),
		Lines:     len(lines),
		Bytes:     len(content),
	}
}

// writeFileExecutionContentError is the execution-time guard for approved
// write_file calls: empty content or the legacy request-echo template fails the
// run with an honest code instead of being written and reported as verified.
func writeFileExecutionContentError(call protocol.PlannedToolCall) error {
	if !strings.EqualFold(strings.TrimSpace(call.Name), "write_file") {
		return nil
	}
	content := firstNonEmptyString(call.Arguments["content"])
	if echoWords(content) != "" && !strings.Contains(echoWords(content), legacyPlaceholderSignature) {
		return nil
	}
	return fmt.Errorf("%s: approved content for %s is empty or only repeats the request; nothing was written",
		emptyProviderOutputCode, displayFileName(firstNonEmptyString(call.Arguments["path"])))
}

// validateApprovedPlanBeforeExecution checks every approved step before any
// step runs, so an invalid later step cannot leave earlier side effects
// (a created team, a written file) behind.
func validateApprovedPlanBeforeExecution(calls []protocol.PlannedToolCall) error {
	for i, call := range calls {
		call = normalizePlannedToolCall(call)
		if strings.TrimSpace(call.Name) == "" {
			return fmt.Errorf("approved execution plan contained an empty tool name at step %d", i+1)
		}
		if err := writeFileExecutionContentError(call); err != nil {
			return err
		}
	}
	return nil
}

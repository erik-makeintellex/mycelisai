package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/pkg/protocol"
)

// writeFileDraftTimeout bounds the single proposal-time drafting inference.
// It is a variable only so tests can shorten it.
var writeFileDraftTimeout = 90 * time.Second

// writeFileDraftMaxPerTurn and writeFileDraftTurnDeadline bound all drafting in
// one chat turn. Variables only so tests can shorten them.
var (
	writeFileDraftMaxPerTurn   = 3
	writeFileDraftTurnDeadline = 120 * time.Second
)

const writeFileDraftTimeoutCode = "provider_timeout"

const writeFileDraftSystemPrompt = "You write the complete contents of exactly one file. " +
	"Return only the file contents: no preamble, no explanation, no code fences, no tool calls. " +
	"Do not repeat the request back; write the finished content it asks for."

// writeFileDraftBlocker is the honest outcome when drafting cannot produce real
// content. No proposal, intent proof, or confirm token is minted for it.
type writeFileDraftBlocker struct {
	Status       int
	Availability cognitive.ExecutionAvailability
}

// draftMissingWriteFileContent fills every planned write_file call whose content
// is missing or only echoes the request with one bounded model inference (no
// tools). It returns previews for the proposal, or a blocker when any draft
// fails. It never substitutes template content.
func (s *AdminServer) draftMissingWriteFileContent(ctx context.Context, planned []protocol.PlannedToolCall, request string) ([]protocol.ProposalDraftPreview, *writeFileDraftBlocker) {
	explicit := requestHasExplicitWriteFileContent(request)
	var targets []int
	for i, call := range planned {
		if !strings.EqualFold(strings.TrimSpace(call.Name), "write_file") || strings.TrimSpace(call.ToolRef) != "" {
			continue
		}
		existing := firstNonEmptyString(call.Arguments["content"])
		if firstNonEmptyString(call.Arguments["path"]) == "" || (existing != "" && (explicit || !writeFileContentIsRequestEcho(existing, request))) {
			continue
		}
		targets = append(targets, i)
	}
	if len(targets) > writeFileDraftMaxPerTurn {
		return nil, &writeFileDraftBlocker{Status: http.StatusUnprocessableEntity, Availability: cognitive.ExecutionAvailability{
			Code:              emptyProviderOutputCode,
			Summary:           fmt.Sprintf("Soma can draft up to %d files per request, and this request needs %d, so nothing was proposed.", writeFileDraftMaxPerTurn, len(targets)),
			RecommendedAction: fmt.Sprintf("Ask for %d or fewer files at a time, or include the text you want in your request.", writeFileDraftMaxPerTurn),
		}}
	}
	turnCtx, cancel := context.WithTimeout(ctx, writeFileDraftTurnDeadline)
	defer cancel()
	// Drafts are applied only after every draft succeeds, so a failure never
	// leaves a partially drafted plan behind.
	contents := make([]string, len(targets))
	for n, i := range targets {
		content, blocker := s.draftWriteFileContent(turnCtx, firstNonEmptyString(planned[i].Arguments["path"]), request)
		if blocker != nil {
			return nil, blocker
		}
		contents[n] = content
	}
	previews := make([]protocol.ProposalDraftPreview, 0, len(targets))
	for n, i := range targets {
		planned[i].Arguments["content"] = contents[n]
		previews = append(previews, buildProposalDraftPreview(firstNonEmptyString(planned[i].Arguments["path"]), contents[n]))
	}
	return previews, nil
}

func (s *AdminServer) draftWriteFileContent(ctx context.Context, path, request string) (string, *writeFileDraftBlocker) {
	name := displayFileName(path)
	if s == nil || s.Cognitive == nil {
		return "", draftBlocker(http.StatusServiceUnavailable, cognitive.ExecutionRouterUnavailable, name,
			"Soma could not draft %s because no AI engine is available, so nothing was proposed.",
			"Ask an admin to enable an AI engine for Soma, then try again.")
	}
	profile := writeFileDraftProfile(s.Cognitive, path)
	if availability := s.Cognitive.ExecutionAvailability(profile, ""); !availability.Available {
		availability.Summary = fmt.Sprintf("Soma could not draft %s because its AI engine is unavailable, so nothing was proposed.", name)
		return "", &writeFileDraftBlocker{Status: http.StatusServiceUnavailable, Availability: availability}
	}
	draftCtx, cancel := context.WithTimeout(ctx, writeFileDraftTimeout)
	defer cancel()
	prompt := fmt.Sprintf("File to write: %s (%s).\nRequest: %s\nWrite the full contents of %s now.", path, writeFileKindLabel(path), strings.TrimSpace(request), name)
	resp, err := s.Cognitive.InferWithContract(draftCtx, cognitive.InferRequest{
		Profile: profile,
		Prompt:  prompt,
		Messages: []cognitive.ChatMessage{
			{Role: "system", Content: writeFileDraftSystemPrompt},
			{Role: "user", Content: prompt},
		},
	})
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(draftCtx.Err(), context.DeadlineExceeded) {
			return "", draftBlocker(http.StatusGatewayTimeout, writeFileDraftTimeoutCode, name,
				"Soma ran out of time drafting %s, so nothing was proposed.",
				"Try again, or include the text you want in your request.")
		}
		return "", draftBlocker(http.StatusBadGateway, emptyProviderOutputCode, name,
			"Soma could not draft %s, so nothing was proposed.",
			"Try again, or include the text you want in your request.")
	}
	content := ""
	if resp != nil {
		content = stripWrappingCodeFence(resp.Text)
	}
	if writeFileContentIsRequestEcho(content, request) {
		return "", draftBlocker(http.StatusBadGateway, emptyProviderOutputCode, name,
			"Soma's draft for %s was empty or only repeated your request, so nothing was proposed.",
			"Try again with more detail, or include the text you want in your request.")
	}
	return strings.TrimRight(content, "\r\n") + "\n", nil
}

func draftBlocker(status int, code, name, summaryFormat, action string) *writeFileDraftBlocker {
	return &writeFileDraftBlocker{Status: status, Availability: cognitive.ExecutionAvailability{
		Available:         false,
		Code:              code,
		Summary:           fmt.Sprintf(summaryFormat, name),
		RecommendedAction: action,
	}}
}

// withDraftPreviews attaches drafted-content previews so the operator approves
// the real content, and says so in the proposal's expected result.
func withDraftPreviews(proposal *protocol.ChatProposal, previews []protocol.ProposalDraftPreview) *protocol.ChatProposal {
	if proposal == nil || len(previews) == 0 {
		return proposal
	}
	proposal.DraftPreviews = previews
	proposal.ExpectedResult = strings.TrimSpace(proposal.ExpectedResult + " Soma drafted the content; review the draft, then approve to save it.")
	return proposal
}

func respondWriteFileDraftBlocker(w http.ResponseWriter, blocker *writeFileDraftBlocker) {
	respondAPIJSON(w, blocker.Status, protocol.APIResponse{
		OK:    false,
		Error: blocker.Availability.Summary,
		Data:  blocker.Availability,
	})
}

// writeFileDraftProfile uses the coder profile for code files when it is bound
// and available, and otherwise the chat profile that serves Soma's chat path.
func writeFileDraftProfile(router *cognitive.Router, path string) string {
	if writeFileIsCode(path) && router.ExecutionAvailability("coder", "").Available {
		return "coder"
	}
	return cognitive.DefaultExecutionProfileName
}

func writeFileIsCode(path string) bool {
	switch filepathExt(path) {
	case ".py", ".js", ".mjs", ".ts", ".tsx", ".jsx", ".go", ".rs", ".java", ".sh", ".html", ".css", ".sql":
		return true
	}
	return false
}

func writeFileKindLabel(path string) string {
	switch filepathExt(path) {
	case ".md", ".markdown":
		return "Markdown document with headings and real content"
	case ".txt", ".log":
		return "plain text"
	case ".json":
		return "valid JSON"
	case ".yaml", ".yml":
		return "valid YAML"
	case ".csv":
		return "CSV with a header row"
	case ".html":
		return "a complete, self-contained HTML page"
	}
	if writeFileIsCode(path) {
		return "complete, runnable source code"
	}
	return "file contents matching the extension"
}

func displayFileName(path string) string {
	trimmed := strings.TrimRight(strings.ReplaceAll(strings.TrimSpace(path), `\`, "/"), "/")
	if idx := strings.LastIndex(trimmed, "/"); idx >= 0 {
		return trimmed[idx+1:]
	}
	return trimmed
}

// agentReplyIsEmpty reports a mutation-turn agent reply with no text, tool
// choice, planned call, or artifact. Such a reply never becomes a proposal.
func agentReplyIsEmpty(agentResult chatAgentResult) bool {
	return strings.TrimSpace(agentResult.Text) == "" && len(agentResult.ToolsUsed) == 0 &&
		len(agentResult.PlannedToolCalls) == 0 && len(agentResult.Artifacts) == 0
}

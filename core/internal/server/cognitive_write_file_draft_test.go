package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
)

const teamBriefRequest = "Create a short markdown file named team-brief.md that lists three things Mycelis can do for a small business."

const teamBriefDraft = "# Team brief\n\nMycelis can help a small business:\n\n1. Draft customer replies.\n2. Summarize weekly sales notes.\n3. Keep a reviewable record of approved work.\n"

// draftTestProvider is a fake router adapter: it records prompts and returns a
// configured reply, error, or delay.
type draftTestProvider struct {
	mu      sync.Mutex
	text    string
	err     error
	delay   time.Duration
	prompts []cognitive.InferOptions
}

func (p *draftTestProvider) Infer(ctx context.Context, _ string, opts cognitive.InferOptions) (*cognitive.InferResponse, error) {
	p.mu.Lock()
	p.prompts = append(p.prompts, opts)
	p.mu.Unlock()
	if p.delay > 0 {
		select {
		case <-time.After(p.delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	if p.err != nil {
		return nil, p.err
	}
	return &cognitive.InferResponse{Text: p.text, Provider: "mock", ModelUsed: "test-model"}, nil
}

func (p *draftTestProvider) Probe(context.Context) (bool, error) { return true, nil }

func (p *draftTestProvider) calls() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.prompts)
}

func draftTestRouter(provider cognitive.LLMProvider) *cognitive.Router {
	return &cognitive.Router{
		Config: &cognitive.BrainConfig{
			Profiles:  map[string]string{"chat": "mock"},
			Providers: map[string]cognitive.ProviderConfig{"mock": {Type: "mock", Enabled: true, ModelID: "test-model"}},
		},
		Adapters: map[string]cognitive.LLMProvider{"mock": provider},
	}
}

// respondAsAdminAgentForTest stands in for Soma's admin agent on NATS so chat
// tests exercise the real agent path instead of a deterministic bypass.
func respondAsAdminAgentForTest(t *testing.T, s *AdminServer, reply map[string]any) {
	t.Helper()
	payload, _ := json.Marshal(reply)
	if _, err := s.NC.Subscribe("swarm.council.admin.request", func(msg *nats.Msg) { _ = msg.Respond(payload) }); err != nil {
		t.Fatalf("subscribe admin agent: %v", err)
	}
	if err := s.NC.Flush(); err != nil {
		t.Fatalf("flush: %v", err)
	}
}

// requestPlanResultForTest plans a request the way the agent path does and
// reports the primary output target in Text for assertions.
func requestPlanResultForTest(request string, tools []string) chatAgentResult {
	planned := buildPlannedToolCalls(chatAgentResult{}, request, tools)
	return chatAgentResult{Text: "target: " + firstPlannedOutputTarget(planned), ToolsUsed: toolsForPlannedCalls(planned, tools)}
}

func postDraftChat(t *testing.T, s *AdminServer, request string) *httptest.ResponseRecorder {
	t.Helper()
	respondAsAdminAgentForTest(t, s, map[string]any{
		"text":       `{"tool_call":{"name":"write_file","arguments":{"path":"` + extractRequestedFilePath(request) + `"}}}`,
		"tools_used": []string{"write_file"},
	})
	body, _ := json.Marshal(map[string]any{"messages": []chatRequestMessage{{Role: "user", Content: request}}})
	rr := httptest.NewRecorder()
	http.HandlerFunc(s.HandleChat).ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/api/v1/chat", bytes.NewBuffer(body)))
	return rr
}

func assertNoPlaceholderTemplate(t *testing.T, body string) {
	t.Helper()
	for _, marker := range []string{"Generated note", "Soma created this file from your request", "What you asked for", "Welcome to Mycelis"} {
		if strings.Contains(body, marker) {
			t.Fatalf("response carries placeholder template text %q: %s", marker, body)
		}
	}
}

func TestHandleChat_DraftsMissingWriteFileContentWithModel(t *testing.T) {
	provider := &draftTestProvider{text: "```markdown\n" + teamBriefDraft + "```"}
	s := newTestServer(withNATS(t))
	s.Cognitive = draftTestRouter(provider)

	rr := postDraftChat(t, s, teamBriefRequest)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rr.Code, rr.Body.String())
	}
	assertNoPlaceholderTemplate(t, rr.Body.String())
	payload := decodeChatPayloadForTest(t, rr.Body.Bytes())
	if payload.Proposal == nil || len(payload.Proposal.DraftPreviews) != 1 {
		t.Fatalf("proposal draft previews = %+v", payload.Proposal)
	}
	preview := payload.Proposal.DraftPreviews[0]
	if preview.Path != "team-brief.md" || !preview.FullDraft || preview.Preview != strings.TrimRight(teamBriefDraft, "\n") {
		t.Fatalf("preview = %+v, want full model draft", preview)
	}
	if !strings.Contains(payload.Proposal.ExpectedResult, "review the draft") {
		t.Fatalf("expected_result = %q, want draft review copy", payload.Proposal.ExpectedResult)
	}
	if provider.calls() != 1 {
		t.Fatalf("draft inferences = %d, want exactly one", provider.calls())
	}
	opts := provider.prompts[0]
	if len(opts.Messages) != 2 || opts.Messages[0].Role != "system" || !strings.Contains(opts.Messages[1].Content, "team-brief.md") || !strings.Contains(opts.Messages[1].Content, "Markdown") {
		t.Fatalf("draft messages = %+v, want system + file-typed user prompt", opts.Messages)
	}
}

func TestHandleChat_DraftFailureReturnsBlockerWithoutProposal(t *testing.T) {
	cases := []struct {
		name     string
		provider *draftTestProvider
		wantCode string
		wantHTTP int
	}{
		{"engine error", &draftTestProvider{err: errors.New("dial tcp: refused")}, emptyProviderOutputCode, http.StatusBadGateway},
		{"empty draft", &draftTestProvider{text: "   \n"}, emptyProviderOutputCode, http.StatusBadGateway},
		{"echo-only draft", &draftTestProvider{text: "# Request\n\n" + teamBriefRequest}, emptyProviderOutputCode, http.StatusBadGateway},
		{"legacy template draft", &draftTestProvider{text: "# Generated note\n\nSoma created this file from your request.\n\n## What you asked for\n\n" + teamBriefRequest}, emptyProviderOutputCode, http.StatusBadGateway},
		{"timeout", &draftTestProvider{text: teamBriefDraft, delay: time.Second}, writeFileDraftTimeoutCode, http.StatusGatewayTimeout},
	}
	previous := writeFileDraftTimeout
	writeFileDraftTimeout = 50 * time.Millisecond
	t.Cleanup(func() { writeFileDraftTimeout = previous })
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := newTestServer(withNATS(t))
			s.Cognitive = draftTestRouter(tc.provider)
			rr := postDraftChat(t, s, teamBriefRequest)
			if rr.Code != tc.wantHTTP {
				t.Fatalf("status = %d, want %d body=%s", rr.Code, tc.wantHTTP, rr.Body.String())
			}
			body := rr.Body.String()
			for _, forbidden := range []string{"confirm_token", "intent_proof_id", "proposal"} {
				if strings.Contains(body, forbidden) {
					t.Fatalf("blocker body carries %q: %s", forbidden, body)
				}
			}
			var resp struct {
				OK   bool                            `json:"ok"`
				Data cognitive.ExecutionAvailability `json:"data"`
			}
			if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if resp.OK || resp.Data.Code != tc.wantCode || !strings.Contains(resp.Data.Summary, "team-brief.md") || !strings.Contains(resp.Data.Summary, "nothing was proposed") {
				t.Fatalf("blocker = %+v, want code %s", resp.Data, tc.wantCode)
			}
		})
	}
}

func TestHandleChat_ExplicitQuotedContentIsNotDrafted(t *testing.T) {
	provider := &draftTestProvider{text: "should not be used"}
	s := newTestServer(withNATS(t))
	s.Cognitive = draftTestRouter(provider)
	rr := postDraftChat(t, s, "Create a file named workspace/notes/hello.txt with content \"hello from the operator\".")
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rr.Code, rr.Body.String())
	}
	payload := decodeChatPayloadForTest(t, rr.Body.Bytes())
	// The content is not model-drafted (the provider is never called), but the
	// operator must still see what will be written: explicit quoted content
	// gets its own preview, matching the drafted-content path.
	if payload.Proposal == nil || len(payload.Proposal.DraftPreviews) != 1 {
		t.Fatalf("explicit content must still carry a preview: %+v", payload.Proposal)
	}
	preview := payload.Proposal.DraftPreviews[0]
	if preview.Path != "workspace/notes/hello.txt" || preview.Preview != "hello from the operator" || !preview.FullDraft {
		t.Fatalf("preview = %+v, want preview of the explicit content", preview)
	}
	if provider.calls() != 0 {
		t.Fatalf("draft inferences = %d, want none for explicit content", provider.calls())
	}
	calls := buildPlannedToolCalls(chatAgentResult{}, "Create a file named workspace/notes/hello.txt with content \"hello from the operator\".", []string{"write_file"})
	if len(calls) != 1 || calls[0].Arguments["content"] != "hello from the operator" {
		t.Fatalf("planned = %#v, want explicit quoted content unchanged", calls)
	}
}

func TestDraftedWriteFile_ExecutesApprovedDraftExactly(t *testing.T) {
	workspace := t.TempDir()
	t.Setenv("MYCELIS_WORKSPACE", workspace)
	s := newTestServer()
	s.Cognitive = draftTestRouter(&draftTestProvider{text: teamBriefDraft})

	planned := buildPlannedToolCalls(chatAgentResult{}, teamBriefRequest, []string{"write_file"})
	if len(planned) != 1 || firstNonEmptyString(planned[0].Arguments["content"]) != "" {
		t.Fatalf("plan before drafting = %#v, want path-only write_file", planned)
	}
	previews, blocker := s.draftMissingWriteFileContent(t.Context(), planned, teamBriefRequest)
	if blocker != nil || len(previews) != 1 {
		t.Fatalf("draft blocker=%+v previews=%+v", blocker, previews)
	}
	scope := &protocol.ScopeValidation{Tools: []string{"write_file"}, PlannedToolCalls: planned}
	if _, err := s.executePlannedToolCalls(t.Context(), scope, "test-user", "", "", "", "", false); err != nil {
		t.Fatalf("execute: %v", err)
	}
	written, err := os.ReadFile(filepath.Join(workspace, "team-brief.md"))
	if err != nil {
		t.Fatalf("read written file: %v", err)
	}
	if string(written) != teamBriefDraft {
		t.Fatalf("written = %q, want approved draft %q", written, teamBriefDraft)
	}
}

func TestExecutePlannedToolCalls_RejectsEmptyOrTemplateWriteFileContent(t *testing.T) {
	workspace := t.TempDir()
	t.Setenv("MYCELIS_WORKSPACE", workspace)
	s := newTestServer()
	for name, content := range map[string]any{
		"missing":  nil,
		"blank":    "  \n",
		"template": "# Generated note\n\nSoma created this file from your request.\n\n## What you asked for\n\nx\n",
	} {
		args := map[string]any{"path": "rejected-" + name + ".md"}
		if content != nil {
			args["content"] = content
		}
		scope := &protocol.ScopeValidation{PlannedToolCalls: []protocol.PlannedToolCall{{Name: "write_file", Arguments: args}}}
		_, err := s.executePlannedToolCalls(t.Context(), scope, "test-user", "", "", "", "", false)
		if err == nil || !strings.HasPrefix(err.Error(), emptyProviderOutputCode+":") {
			t.Fatalf("%s: err = %v, want %s failure", name, err, emptyProviderOutputCode)
		}
		if _, statErr := os.Stat(filepath.Join(workspace, "rejected-"+name+".md")); !os.IsNotExist(statErr) {
			t.Fatalf("%s: file was written despite rejection", name)
		}
	}
}

func TestWriteFileContentIsRequestEcho(t *testing.T) {
	cases := []struct {
		content string
		want    bool
	}{
		{"", true},
		{"team-brief.md", true},
		{teamBriefRequest, true},
		{"# Your request\n\n> " + teamBriefRequest, true},
		{teamBriefDraft, false},
		{"ok", false},
	}
	for _, tc := range cases {
		if got := writeFileContentIsRequestEcho(tc.content, teamBriefRequest); got != tc.want {
			t.Fatalf("echo(%q) = %v, want %v", tc.content, got, tc.want)
		}
	}
}

func TestBuildProposalDraftPreview_BoundsLinesAndBytes(t *testing.T) {
	long := strings.Repeat("line\n", 30)
	preview := buildProposalDraftPreview("a.md", long)
	if preview.FullDraft || strings.Count(preview.Preview, "\n") != draftPreviewMaxLines-1 || preview.Lines != 30 {
		t.Fatalf("line-bounded preview = %+v", preview)
	}
	wide := strings.Repeat("é", 2000)
	preview = buildProposalDraftPreview("b.md", wide)
	if preview.FullDraft || len(preview.Preview) > draftPreviewMaxBytes || !strings.HasPrefix(wide, preview.Preview) {
		t.Fatalf("byte-bounded preview len=%d full=%v", len(preview.Preview), preview.FullDraft)
	}
}

func TestDraftMissingWriteFileContent_TurnDeadlineBlocksWithoutPartialPlan(t *testing.T) {
	previous := writeFileDraftTurnDeadline
	writeFileDraftTurnDeadline = 50 * time.Millisecond
	t.Cleanup(func() { writeFileDraftTurnDeadline = previous })
	s := newTestServer()
	s.Cognitive = draftTestRouter(&draftTestProvider{text: teamBriefDraft, delay: 30 * time.Millisecond})
	planned := []protocol.PlannedToolCall{
		{Name: "write_file", Arguments: map[string]any{"path": "a.md"}},
		{Name: "write_file", Arguments: map[string]any{"path": "b.md"}},
	}
	previews, blocker := s.draftMissingWriteFileContent(t.Context(), planned, teamBriefRequest)
	if blocker == nil || blocker.Availability.Code != writeFileDraftTimeoutCode || previews != nil {
		t.Fatalf("blocker=%+v previews=%+v, want turn-deadline timeout", blocker, previews)
	}
	for _, call := range planned {
		if _, has := call.Arguments["content"]; has {
			t.Fatalf("partial draft applied to %v", call.Arguments["path"])
		}
	}
}

func TestExecutePlannedToolCalls_ValidatesWholePlanBeforeAnyStep(t *testing.T) {
	workspace := t.TempDir()
	t.Setenv("MYCELIS_WORKSPACE", workspace)
	s := newTestServer()
	scope := &protocol.ScopeValidation{PlannedToolCalls: []protocol.PlannedToolCall{
		{Name: "write_file", Arguments: map[string]any{"path": "step-one.md", "content": "# Real content\n"}},
		{Name: "write_file", Arguments: map[string]any{"path": "step-two.md", "content": ""}},
	}}
	if _, err := s.executePlannedToolCalls(t.Context(), scope, "test-user", "", "", "", "", false); err == nil {
		t.Fatal("want plan rejected")
	}
	if _, err := os.Stat(filepath.Join(workspace, "step-one.md")); !os.IsNotExist(err) {
		t.Fatalf("step one executed before step two was validated: %v", err)
	}
}

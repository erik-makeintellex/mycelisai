package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

// TestHandleChat_AgentOutputContentGetsPreview reproduces the live-probe defect:
// when the agent writes the file content itself (no explicit quoted content, no
// model drafting call), the resulting write_file proposal must still carry a
// preview of exactly that content. Before this fix, agent-authored content
// skipped draftMissingWriteFileContent entirely because it was neither empty
// nor a request echo, so no draft_previews entry was attached and the operator
// approved content they never saw.
func TestHandleChat_AgentOutputContentGetsPreview(t *testing.T) {
	agentContent := "# Three ways a florist can use AI\n\n" +
		"1. Draft personalized bouquet recommendations for customers.\n" +
		"2. Forecast seasonal demand from past orders.\n" +
		"3. Auto-generate social media captions for new arrivals.\n"
	provider := &draftTestProvider{text: "should not be called for agent-authored content"}
	s := newTestServer(withNATS(t))
	s.Cognitive = draftTestRouter(provider)
	respondAsAdminAgentForTest(t, s, map[string]any{
		"text":       agentContent,
		"tools_used": []string{"write_file"},
	})
	request := "Create a markdown file named florist-ai.md listing three ways a florist can use AI."
	body, _ := json.Marshal(map[string]any{"messages": []chatRequestMessage{{Role: "user", Content: request}}})
	rr := httptest.NewRecorder()
	http.HandlerFunc(s.HandleChat).ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/api/v1/chat", bytes.NewBuffer(body)))
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rr.Code, rr.Body.String())
	}
	assertNoPlaceholderTemplate(t, rr.Body.String())
	payload := decodeChatPayloadForTest(t, rr.Body.Bytes())
	if payload.Proposal == nil || len(payload.Proposal.DraftPreviews) != 1 {
		t.Fatalf("agent-authored content must carry a preview: %+v", payload.Proposal)
	}
	preview := payload.Proposal.DraftPreviews[0]
	if preview.Path != "florist-ai.md" || !preview.FullDraft || preview.Preview != strings.TrimRight(agentContent, "\n") {
		t.Fatalf("preview = %+v, want preview of the agent's own output", preview)
	}
	if provider.calls() != 0 {
		t.Fatalf("draft inferences = %d, want none: agent already wrote the content", provider.calls())
	}
}

// TestAgentOutputWriteFile_ExecutesPreviewedContentExactly proves the executed
// file content equals what was previewed, for the agent-authored-content path
// (not just the model-drafted path already covered elsewhere).
func TestAgentOutputWriteFile_ExecutesPreviewedContentExactly(t *testing.T) {
	workspace := t.TempDir()
	t.Setenv("MYCELIS_WORKSPACE", workspace)
	s := newTestServer()
	s.Cognitive = draftTestRouter(&draftTestProvider{text: "should not be called"})

	agentContent := "# Notes\n\nSome real agent-authored content.\n"
	request := "Create a markdown file named notes/agent-notes.md with the plan."
	planned := buildPlannedToolCalls(chatAgentResult{Text: agentContent, ToolsUsed: []string{"write_file"}}, request, []string{"write_file"})
	if len(planned) != 1 || firstNonEmptyString(planned[0].Arguments["content"]) != strings.TrimSpace(agentContent) {
		t.Fatalf("plan = %#v, want agent output as content before preview", planned)
	}
	previews, blocker := s.draftMissingWriteFileContent(t.Context(), planned, request)
	if blocker != nil || len(previews) != 1 {
		t.Fatalf("draft blocker=%+v previews=%+v", blocker, previews)
	}
	previewed := previews[0].Preview
	scope := &protocol.ScopeValidation{Tools: []string{"write_file"}, PlannedToolCalls: planned}
	if _, err := s.executePlannedToolCalls(t.Context(), scope, "test-user", "", "", "", "", false); err != nil {
		t.Fatalf("execute: %v", err)
	}
	written, err := os.ReadFile(filepath.Join(workspace, "notes/agent-notes.md"))
	if err != nil {
		t.Fatalf("read written file: %v", err)
	}
	if strings.TrimRight(string(written), "\n") != previewed {
		t.Fatalf("written = %q, want it to equal the previewed content %q", written, previewed)
	}
}

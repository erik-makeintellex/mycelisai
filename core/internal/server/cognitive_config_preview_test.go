package server

import (
	"context"
	"strings"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

const validOutcomeTemplatePreview = `apiVersion: mycelis.ai/v1
kind: OutcomeTemplate
metadata:
  id: quarterly-launch-review
  name: Quarterly launch review
  version: "1"
  owner_id: soma
  scope: {kind: workspace, ref: primary}
  enabled: true
  source: {kind: soma, ref: conversation:test}
  governance: {risk_level: low, approval_posture: optional}
spec:
  defaults:
    target_outcome: Review a quarterly launch
    delivery_form: Decision-ready launch review
    acceptance_evidence: [Owner and target date are present]
`

func TestExecuteRequestedConfigPreviewValidatesWithoutProposal(t *testing.T) {
	server := &AdminServer{}
	result, ok := server.executeRequestedConfigPreview(
		context.Background(),
		"Preview this Outcome Template and do not save it:\n\n"+validOutcomeTemplatePreview,
		chatAgentResult{},
	)
	if !ok {
		t.Fatal("expected deterministic preview")
	}
	if len(result.ToolsUsed) != 1 || result.ToolsUsed[0] != "preview_config_document" {
		t.Fatalf("tools_used = %#v", result.ToolsUsed)
	}
	for _, want := range []string{"preview is valid", "Nothing was saved or activated"} {
		if !strings.Contains(result.Text, want) {
			t.Fatalf("result %q missing %q", result.Text, want)
		}
	}
	if len(result.PlannedToolCalls) != 0 || len(result.Artifacts) != 0 {
		t.Fatalf("preview created durable state: calls=%#v artifacts=%#v", result.PlannedToolCalls, result.Artifacts)
	}
}

func TestExecuteRequestedConfigPreviewUsesModelAuthoredDocumentFallback(t *testing.T) {
	server := &AdminServer{}
	result, ok := server.executeRequestedConfigPreview(
		context.Background(),
		"Draft and preview an Outcome Template for a quarterly launch.",
		chatAgentResult{Text: "```yaml\n" + validOutcomeTemplatePreview + "```", ProviderID: "mock"},
	)
	if !ok || result.ProviderID != "mock" || !strings.Contains(result.Text, "preview is valid") {
		t.Fatalf("fallback preview = %#v, ok=%v", result, ok)
	}
}

func TestBuildPlannedToolCallsPreservesAgentPlannedConfigArguments(t *testing.T) {
	agentResult := chatAgentResult{PlannedToolCalls: []protocol.PlannedToolCall{{
		Name: "activate_config_document",
		Arguments: map[string]any{
			"record_id": "11111111-1111-1111-1111-111111111111",
			"action":    "activate",
		},
	}}}

	planned := buildPlannedToolCalls(agentResult, "Use this outcome template.", []string{"activate_config_document"})
	if len(planned) != 1 {
		t.Fatalf("planned calls = %#v, want one", planned)
	}
	if planned[0].Name != "activate_config_document" || planned[0].Arguments["record_id"] != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("planned call = %#v, want preserved activation arguments", planned[0])
	}
}

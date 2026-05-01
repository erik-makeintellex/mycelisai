package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

func TestBuildSomaReferentialReviewRequiresTeamMCPConfirmation(t *testing.T) {
	s := newTestServer()
	messages := []chatRequestMessage{{
		Role:    "user",
		Content: "Build a team with the right specialists and target MCP tools for web search.",
	}}

	review := s.buildSomaReferentialReview(t.Context(), messages)

	if !review.NeedsConfirmation {
		t.Fatal("expected team + MCP action to require confirmation")
	}
	if !strings.Contains(review.InferredAction, "specialist roles") {
		t.Fatalf("inferred action = %q", review.InferredAction)
	}
	if len(review.MutationTools) != 2 || review.MutationTools[0] != "generate_blueprint" || review.MutationTools[1] != "delegate" {
		t.Fatalf("mutation tools = %#v, want generate_blueprint + delegate", review.MutationTools)
	}
	if len(review.ConfigurationAdvice) == 0 {
		t.Fatal("expected MCP configuration advice")
	}
	if review.TemplateID != "team-with-target-tools" {
		t.Fatalf("template id = %q, want team-with-target-tools", review.TemplateID)
	}
}

func TestBuildSomaReferentialReviewAppliesConfirmationToPriorAction(t *testing.T) {
	s := newTestServer()
	messages := []chatRequestMessage{
		{Role: "user", Content: "Create a compact team for release review and associate web search MCP tools."},
		{Role: "assistant", Content: "Please confirm whether I should proceed once."},
		{Role: "user", Content: "yes proceed once"},
	}

	review := s.buildSomaReferentialReview(t.Context(), messages)

	if !review.Confirmed {
		t.Fatal("expected latest confirmation to bind to prior action")
	}
	if !strings.Contains(review.EffectiveRequest, "Create a compact team") {
		t.Fatalf("effective request = %q", review.EffectiveRequest)
	}
	if !strings.Contains(review.contextBlock(), "User confirmation applies to prior request") {
		t.Fatalf("context block missing confirmation linkage: %s", review.contextBlock())
	}
}

func TestBuildSomaReferentialReviewProtectsPrivateServiceActions(t *testing.T) {
	s := newTestServer()
	messages := []chatRequestMessage{{
		Role:    "user",
		Content: "Use the private service through its API token and connect the GitHub MCP for deployment notes.",
	}}

	review := s.buildSomaReferentialReview(t.Context(), messages)

	if !review.NeedsConfirmation {
		t.Fatal("expected private service action to require confirmation")
	}
	if review.TemplateID != "protected-service-action" {
		t.Fatalf("template id = %q, want protected-service-action", review.TemplateID)
	}
	if !strings.Contains(review.ProtectionReason, "private service") {
		t.Fatalf("protection reason = %q", review.ProtectionReason)
	}
	if !containsString(review.Concepts, "credential_boundary") {
		t.Fatalf("concepts = %#v, want credential_boundary", review.Concepts)
	}
}

func TestBuildSomaReferentialReviewAppliesPrivateServiceConfirmation(t *testing.T) {
	s := newTestServer()
	messages := []chatRequestMessage{
		{Role: "user", Content: "Use the private service through its API token for deployment notes."},
		{Role: "assistant", Content: "Confirm the target service, allowed action, and whether this is one-time or reusable."},
		{Role: "user", Content: "confirm one time"},
	}

	review := s.buildSomaReferentialReview(t.Context(), messages)

	if !review.Confirmed {
		t.Fatal("expected private service confirmation to bind to prior protected action")
	}
	if review.TemplateID != "protected-service-action" {
		t.Fatalf("template id = %q, want protected-service-action", review.TemplateID)
	}
	confirmed := applyConfirmedReferentialAction(messages, review)
	normalized, tools := normalizeChatRequestMessages(confirmed)
	last := normalized[len(normalized)-1].Content
	if !strings.Contains(last, governedMutationRoutePrefix) {
		t.Fatalf("last normalized content missing mutation route: %q", last)
	}
	if len(tools) != 1 || tools[0] != "delegate" {
		t.Fatalf("tools = %#v, want delegate", tools)
	}
}

func TestBuildSomaReferentialReviewAllowsReadOnlyMCPReview(t *testing.T) {
	s := newTestServer()
	messages := []chatRequestMessage{{
		Role:    "user",
		Content: "Review the current MCP tool structure and recommend what should be connected next.",
	}}

	review := s.buildSomaReferentialReview(t.Context(), messages)

	if review.NeedsConfirmation {
		t.Fatalf("read-only MCP review should not require confirmation: %#v", review)
	}
	if review.InferredAction != "review current MCP/tool posture and recommend next connected tools" {
		t.Fatalf("inferred action = %q", review.InferredAction)
	}
	if !containsString(review.ThemeIDs, "mcp_tool_posture") {
		t.Fatalf("theme ids = %#v, want mcp_tool_posture", review.ThemeIDs)
	}
}

func TestBuildSomaReferentialReviewProtectsRecurringTemplateBehavior(t *testing.T) {
	s := newTestServer()
	messages := []chatRequestMessage{{
		Role:    "user",
		Content: "Always use this as a reusable template for customer data reviews from deployment context.",
	}}

	review := s.buildSomaReferentialReview(t.Context(), messages)

	if !review.NeedsConfirmation {
		t.Fatal("expected recurring private-data behavior to require confirmation")
	}
	if review.TemplateID != "private-data-review" {
		t.Fatalf("template id = %q, want private-data-review", review.TemplateID)
	}
	if !containsString(review.Concepts, "conversation_template") {
		t.Fatalf("concepts = %#v, want conversation_template", review.Concepts)
	}
}

func TestMatchSomaInteractionTemplateCombinesRecurringPrivateDataConcepts(t *testing.T) {
	match := matchSomaInteractionTemplate("Always review customer data from deployment context as a reusable template before answering.")

	if !match.Protected {
		t.Fatal("expected recurring private-data template to be protected")
	}
	if match.Template.ID != "private-data-review" {
		t.Fatalf("template id = %q, want private-data-review as strongest concrete data action", match.Template.ID)
	}
	if !containsString(match.Concepts, "conversation_template") {
		t.Fatalf("concepts = %#v, want conversation_template", match.Concepts)
	}
	if !containsString(match.Concepts, "private_data") {
		t.Fatalf("concepts = %#v, want private_data", match.Concepts)
	}
}

func TestHandleChatRequestsReferentialConfirmationBeforeTeamMutation(t *testing.T) {
	s := newTestServer()
	reqBody := bytes.NewBufferString(`{"messages":[{"role":"user","content":"Create a team with specialists and target MCP tools for research."}]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat", reqBody)
	rr := httptest.NewRecorder()

	http.HandlerFunc(s.HandleChat).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d body=%s", rr.Code, http.StatusOK, rr.Body.String())
	}
	payload := decodeChatPayloadFromTestResponse(t, rr.Body.Bytes())
	if !strings.Contains(payload.Text, "Confirm whether Soma should create the team plan once") {
		t.Fatalf("payload text = %q", payload.Text)
	}
	if payload.Provenance == nil || payload.Provenance.ResolvedIntent != "confirmation_required" {
		t.Fatalf("provenance = %+v, want confirmation_required", payload.Provenance)
	}
}

func TestConfirmedReferentialActionNormalizesToGovernedMutation(t *testing.T) {
	s := newTestServer()
	messages := []chatRequestMessage{
		{Role: "user", Content: "Create a compact team for release review and associate web search MCP tools."},
		{Role: "assistant", Content: "Please confirm whether I should proceed once."},
		{Role: "user", Content: "confirm"},
	}

	review := s.buildSomaReferentialReview(t.Context(), messages)
	confirmed := applyConfirmedReferentialAction(messages, review)
	normalized, tools := normalizeChatRequestMessages(confirmed)
	last := normalized[len(normalized)-1].Content

	if !strings.Contains(last, governedMutationRoutePrefix) {
		t.Fatalf("last normalized content missing mutation route: %q", last)
	}
	if !strings.Contains(last, "Confirmed action from prior user request") {
		t.Fatalf("last normalized content missing confirmed prior action: %q", last)
	}
	if len(tools) != 2 || tools[0] != "generate_blueprint" || tools[1] != "delegate" {
		t.Fatalf("tools = %#v, want generate_blueprint + delegate", tools)
	}
}

func decodeChatPayloadFromTestResponse(t *testing.T, body []byte) protocol.ChatResponsePayload {
	t.Helper()
	var resp protocol.APIResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("decode api response: %v", err)
	}
	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		t.Fatalf("marshal data: %v", err)
	}
	var envelope protocol.CTSEnvelope
	if err := json.Unmarshal(dataBytes, &envelope); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	var payload protocol.ChatResponsePayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		t.Fatalf("decode chat payload: %v", err)
	}
	return payload
}

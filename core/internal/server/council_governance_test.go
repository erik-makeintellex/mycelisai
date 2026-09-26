package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/internal/configdocuments"
	"github.com/mycelis/core/internal/governance"
	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
)

// councilMutation runs one council turn against a stub member that answers
// with text, and returns the recorder.
func councilMutation(t *testing.T, s *AdminServer, messages []chatRequestMessage) *httptest.ResponseRecorder {
	t.Helper()
	s.Cognitive = &cognitive.Router{
		Config: &cognitive.BrainConfig{Profiles: map[string]string{"chat": "mock"},
			Providers: map[string]cognitive.ProviderConfig{"mock": {Type: "mock", Enabled: true, ModelID: "m"}}},
		Adapters: map[string]cognitive.LLMProvider{"mock": cognitiveTestProvider{}},
	}
	if _, err := s.NC.Subscribe("swarm.council.council-architect.request", func(msg *nats.Msg) {
		resp, _ := json.Marshal(map[string]any{"text": "Here is the plan."})
		msg.Respond(resp)
	}); err != nil {
		t.Fatal(err)
	}
	_ = s.NC.Flush()
	body, _ := json.Marshal(map[string]any{"messages": messages})
	mux := setupMux(t, "POST /api/v1/council/{member}/chat", s.HandleCouncilChat)
	return doAuthenticatedRequest(t, mux, "POST", "/api/v1/council/council-architect/chat", string(body))
}

func councilProposalApproval(t *testing.T, rr *httptest.ResponseRecorder) *protocol.ApprovalPolicy {
	t.Helper()
	assertStatus(t, rr, http.StatusOK)
	var resp struct {
		Data protocol.CTSEnvelope `json:"data"`
	}
	assertJSON(t, rr, &resp)
	var payload protocol.ChatResponsePayload
	if err := json.Unmarshal(resp.Data.Payload, &payload); err != nil || payload.Proposal == nil {
		t.Fatalf("expected a council proposal: %v %s", err, rr.Body.String())
	}
	return payload.Proposal.Approval
}

// councilTemplateThread is a workspace template scoped to the council team.
func councilTemplateYAML(id, scopeRef string) string {
	return strings.NewReplacer("id: retained-browser-app", "id: "+id, "ref: workspace-1", "ref: "+scopeRef).
		Replace(retainedOutcomeTemplateYAML)
}

func councilTemplateThread(id, scopeRef string) []chatRequestMessage {
	return []chatRequestMessage{
		{Role: "user", Content: "Preview and save this Outcome Template:\n```yaml\n" + councilTemplateYAML(id, scopeRef) + "```"},
		{Role: "assistant", Content: "The Outcome Template revision was saved after approval."},
		{Role: "user", Content: "Use the active Outcome Template to write the file output/index.html."},
	}
}

// A2b item 2: template work in council cannot resolve outside its workspace,
// so council refuses it with 409 before anything is audited or minted.
func TestCouncilChatUnresolvedTemplateFailsClosed(t *testing.T) {
	dbOpt, mock := withDB(t)
	mock.ExpectExec("INSERT INTO log_entries").WillReturnResult(sqlmock.NewResult(1, 1))
	s := newTestServer(withCouncilSoma(), withNATS(t), dbOpt)
	rr := councilMutation(t, s, councilTemplateThread("retained-browser-app", "workspace-1"))
	assertStatus(t, rr, http.StatusConflict)
	if !strings.Contains(rr.Body.String(), codeCouncilTemplateUnresolved) ||
		!strings.Contains(rr.Body.String(), "through Soma in its organization") {
		t.Fatalf("expected normalized blocker, got %s", rr.Body.String())
	}
	if mock.ExpectationsWereMet() == nil {
		t.Fatal("nothing may be audited, proven or minted for an unresolved template")
	}
}

// A2b item 2: resolved template work in council gets the posture floor; a
// degraded guard still raises it to required with the role gate.
func TestCouncilChatTemplateWorkGetsPostureFloor(t *testing.T) {
	const templateID = "delivery-posture-governed-enterprise"
	thread := councilTemplateThread(templateID, "council-core")
	document, err := configdocuments.ParseDocument([]byte(councilTemplateYAML(templateID, "council-core")), "yaml")
	if err != nil {
		t.Fatal(err)
	}
	digest, _ := protocol.CanonicalConfigDocumentDigest(document)
	dbOpt, mock := withDB(t)
	mock.ExpectQuery("FROM config_document_activations activation.*JOIN config_documents document").
		WillReturnRows(serverConfigRevisionRows("44444444-4444-4444-4444-444444444444", document, digest))
	s := newTestServer(withCouncilSoma(), withNATS(t), dbOpt)
	s.Guard = governance.NewDegradedGuard(errors.New("missing policy"))
	approval := councilProposalApproval(t, councilMutation(t, s, thread))
	if approval == nil || !approval.ApprovalRequired || !requiresApprover(&protocol.ScopeValidation{Approval: approval}) {
		t.Fatalf("council template work must be raised to approver-required: %+v", approval)
	}
}

// Plain council mutations (no template) keep their capability-only policy.
func TestCouncilChatPlainMutationUnchanged(t *testing.T) {
	s := newTestServer(withCouncilSoma(), withNATS(t))
	s.Guard = governance.NewDegradedGuard(errors.New("missing policy"))
	approval := councilProposalApproval(t, councilMutation(t, s, []chatRequestMessage{
		{Role: "user", Content: "Create a simple python file named workspace/logs/council_plain.py that prints hello world."},
	}))
	if approval != nil && requiresApprover(&protocol.ScopeValidation{Approval: approval}) {
		t.Fatalf("plain council work must not get the posture floor: %+v", approval)
	}
}

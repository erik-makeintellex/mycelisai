package server

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/pkg/protocol"
)

// ── CE-1: Template Registry ────────────────────────────────────────

func TestHandleListTemplates(t *testing.T) {
	s := newTestServer()

	rr := doRequest(t, http.HandlerFunc(s.handleListTemplatesAPI), "GET", "/api/v1/templates", "")
	assertStatus(t, rr, http.StatusOK)

	var resp protocol.APIResponse
	assertJSON(t, rr, &resp)
	if !resp.OK {
		t.Fatal("Expected ok=true")
	}

	templates, ok := resp.Data.([]any)
	if !ok {
		t.Fatalf("Expected data to be array, got %T", resp.Data)
	}
	if len(templates) != len(protocol.TemplateRegistry) {
		t.Errorf("Expected %d templates, got %d", len(protocol.TemplateRegistry), len(templates))
	}
}

// ── CE-1: Intent Proof Lifecycle ───────────────────────────────────

func TestCreateIntentProof(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)

	mock.ExpectExec("INSERT INTO intent_proofs").
		WillReturnResult(sqlmock.NewResult(0, 1))

	scope := &protocol.ScopeValidation{
		Tools:             []string{"research_for_blueprint"},
		AffectedResources: []string{"missions"},
		RiskLevel:         "low",
	}

	proof, err := s.createIntentProof(protocol.TemplateChatToProposal, "Build a dashboard", scope, "")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if proof == nil {
		t.Fatal("Expected proof, got nil")
	}
	if proof.TemplateID != protocol.TemplateChatToProposal {
		t.Errorf("Expected template_id %q, got %q", protocol.TemplateChatToProposal, proof.TemplateID)
	}
	if proof.Status != "pending" {
		t.Errorf("Expected status 'pending', got %q", proof.Status)
	}
	if proof.ResolvedIntent != "Build a dashboard" {
		t.Errorf("Expected intent 'Build a dashboard', got %q", proof.ResolvedIntent)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet DB expectations: %v", err)
	}
}

func TestCreateIntentProof_NilDB(t *testing.T) {
	s := newTestServer()

	proof, err := s.createIntentProof(protocol.TemplateChatToProposal, "test", nil, "")
	if err != nil {
		t.Fatalf("Expected nil error for nil DB, got: %v", err)
	}
	if proof != nil {
		t.Error("Expected nil proof when DB is nil")
	}
}

// ── CE-1: Confirm Token Lifecycle ──────────────────────────────────

func TestGenerateConfirmToken(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)

	mock.ExpectExec("INSERT INTO confirm_tokens").
		WillReturnResult(sqlmock.NewResult(0, 1))

	token, err := s.generateConfirmToken("proof-id-123", protocol.TemplateChatToProposal)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if token == nil {
		t.Fatal("Expected token, got nil")
	}
	if token.IntentProofID != "proof-id-123" {
		t.Errorf("Expected proof ID 'proof-id-123', got %q", token.IntentProofID)
	}
	if token.TemplateID != protocol.TemplateChatToProposal {
		t.Errorf("Expected template_id %q, got %q", protocol.TemplateChatToProposal, token.TemplateID)
	}
	if token.Token == "" {
		t.Error("Expected non-empty token string")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet DB expectations: %v", err)
	}
}

func TestGenerateConfirmToken_NilDB(t *testing.T) {
	s := newTestServer()

	token, err := s.generateConfirmToken("proof-id", protocol.TemplateChatToProposal)
	if err != nil {
		t.Fatalf("Expected nil error for nil DB, got: %v", err)
	}
	if token != nil {
		t.Error("Expected nil token when DB is nil")
	}
}

// ── CE-1: Token Validation ─────────────────────────────────────────

func TestValidateConfirmToken_InvalidFormat(t *testing.T) {
	dbOpt, _ := withDB(t)
	s := newTestServer(dbOpt)

	_, err := s.validateConfirmToken("not-a-uuid")
	if err != errInvalidToken {
		t.Errorf("Expected errInvalidToken, got: %v", err)
	}
}

func TestValidateConfirmToken_NilDB(t *testing.T) {
	s := newTestServer()

	_, err := s.validateConfirmToken("11111111-1111-1111-1111-111111111111")
	if err != errDBUnavailable {
		t.Errorf("Expected errDBUnavailable, got: %v", err)
	}
}

func TestValidateConfirmToken_NotFound(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)

	mock.ExpectQuery("SELECT .+ FROM confirm_tokens").
		WillReturnRows(sqlmock.NewRows([]string{"intent_proof_id", "consumed", "expires_at"}))

	_, err := s.validateConfirmToken("11111111-1111-1111-1111-111111111111")
	if err != errTokenNotFound {
		t.Errorf("Expected errTokenNotFound, got: %v", err)
	}
}

// ── CE-1: Audit Event ──────────────────────────────────────────────

func TestCreateAuditEvent(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)

	mock.ExpectExec("INSERT INTO log_entries").
		WillReturnResult(sqlmock.NewResult(0, 1))

	eventID, err := s.createAuditEvent(protocol.TemplateChatToAnswer, "test", "Test event", map[string]any{"key": "value"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if eventID == "" {
		t.Error("Expected non-empty audit event ID")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet DB expectations: %v", err)
	}
}

func TestCreateAuditEvent_NilDB(t *testing.T) {
	s := newTestServer()

	eventID, err := s.createAuditEvent(protocol.TemplateChatToAnswer, "test", "Test", nil)
	if err != nil {
		t.Fatalf("Expected nil error for nil DB, got: %v", err)
	}
	if eventID != "" {
		t.Error("Expected empty event ID when DB is nil")
	}
}

// ── CE-1: Scope Validation Builder ─────────────────────────────────

func TestBuildScopeFromBlueprint(t *testing.T) {
	bp := &protocol.MissionBlueprint{
		Teams: []protocol.BlueprintTeam{
			{
				Name: "research-team",
				Agents: []protocol.AgentManifest{
					{ID: "researcher", Tools: []string{"search_memory", "recall"}},
					{ID: "writer", Tools: []string{"write_file"}},
				},
			},
			{
				Name: "analysis-team",
				Agents: []protocol.AgentManifest{
					{ID: "analyst", Tools: []string{"search_memory"}},
				},
			},
		},
	}

	scope := buildScopeFromBlueprint(bp)
	if scope == nil {
		t.Fatal("Expected scope, got nil")
	}

	// 3 unique tools: search_memory, recall, write_file
	if len(scope.Tools) != 3 {
		t.Errorf("Expected 3 unique tools, got %d: %v", len(scope.Tools), scope.Tools)
	}
	if scope.RiskLevel != "low" {
		t.Errorf("Expected risk 'low' for 3 agents, got %q", scope.RiskLevel)
	}
}

func TestBuildScopeFromBlueprint_HighRisk(t *testing.T) {
	bp := &protocol.MissionBlueprint{
		Teams: make([]protocol.BlueprintTeam, 5),
	}
	for i := range bp.Teams {
		bp.Teams[i].Agents = []protocol.AgentManifest{
			{ID: "a1"}, {ID: "a2"}, {ID: "a3"},
		}
	}

	scope := buildScopeFromBlueprint(bp)
	if scope.RiskLevel != "high" {
		t.Errorf("Expected risk 'high' for 15 agents / 5 teams, got %q", scope.RiskLevel)
	}
}

// ── CE-1: GET /api/v1/intent/proof/{id} ────────────────────────────

func TestHandleGetIntentProof_InvalidID(t *testing.T) {
	dbOpt, _ := withDB(t)
	s := newTestServer(dbOpt)

	mux := setupMux(t, "GET /api/v1/intent/proof/{id}", s.handleGetIntentProof)
	rr := doRequest(t, mux, "GET", "/api/v1/intent/proof/not-a-uuid", "")
	assertStatus(t, rr, http.StatusBadRequest)
}

func TestHandleGetIntentProof_NilDB(t *testing.T) {
	s := newTestServer()
	mux := setupMux(t, "GET /api/v1/intent/proof/{id}", s.handleGetIntentProof)
	rr := doRequest(t, mux, "GET", "/api/v1/intent/proof/11111111-1111-1111-1111-111111111111", "")
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

func TestHandleGetIntentProof_NotFound(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)

	mock.ExpectQuery("SELECT .+ FROM intent_proofs").
		WillReturnRows(sqlmock.NewRows([]string{"id", "template_id", "resolved_intent", "user_confirmation_token", "permission_check", "policy_decision", "scope_validation", "audit_event_id", "mission_id", "status", "created_at", "confirmed_at"}))

	mux := setupMux(t, "GET /api/v1/intent/proof/{id}", s.handleGetIntentProof)
	rr := doRequest(t, mux, "GET", "/api/v1/intent/proof/11111111-1111-1111-1111-111111111111", "")
	assertStatus(t, rr, http.StatusNotFound)
}

// ── CE-1: Template Protocol Types ──────────────────────────────────

func TestTemplateRegistryContainsBothTemplates(t *testing.T) {
	if _, ok := protocol.TemplateRegistry[protocol.TemplateChatToAnswer]; !ok {
		t.Error("Template registry missing chat-to-answer")
	}
	if _, ok := protocol.TemplateRegistry[protocol.TemplateChatToProposal]; !ok {
		t.Error("Template registry missing chat-to-proposal")
	}
}

func TestChatToAnswerTemplateSpec(t *testing.T) {
	spec := protocol.TemplateRegistry[protocol.TemplateChatToAnswer]
	if spec.RequiresConfirm {
		t.Error("Chat-to-Answer should not require confirmation")
	}
	if spec.MutatesState {
		t.Error("Chat-to-Answer should not mutate state")
	}
	if spec.Mode != protocol.ModeAnswer {
		t.Errorf("Expected mode 'answer', got %q", spec.Mode)
	}
}

func TestChatToProposalTemplateSpec(t *testing.T) {
	spec := protocol.TemplateRegistry[protocol.TemplateChatToProposal]
	if !spec.RequiresConfirm {
		t.Error("Chat-to-Proposal must require confirmation")
	}
	if !spec.MutatesState {
		t.Error("Chat-to-Proposal must mutate state")
	}
	if spec.Mode != protocol.ModeProposal {
		t.Errorf("Expected mode 'proposal', got %q", spec.Mode)
	}
}

// ── CE-1: CommitRequest serialization ──────────────────────────────

func TestCommitRequest_ConfirmTokenField(t *testing.T) {
	body := `{"intent":"Build scraper","confirm_token":"abc-123","teams":[]}`
	var req protocol.CommitRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}
	if req.ConfirmToken != "abc-123" {
		t.Errorf("Expected confirm_token 'abc-123', got %q", req.ConfirmToken)
	}
	if req.Intent != "Build scraper" {
		t.Errorf("Expected intent 'Build scraper', got %q", req.Intent)
	}
}

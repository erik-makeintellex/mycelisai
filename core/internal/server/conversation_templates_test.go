package server

import (
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/pkg/protocol"
)

func TestHandleCreateConversationTemplate_PersistsUserOwnedTemplate(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)
	now := time.Now().UTC()
	mock.ExpectQuery("INSERT INTO conversation_templates").
		WithArgs("default", "Marketing package", "", "temporary_group", "test-user-001", "user", "active", "Create {{output}} for {{product}}.", `{"output":"launch plan"}`, `{}`, `{"name":"Marketing Launch Team"}`, `{}`, `["audit"]`).
		WillReturnRows(conversationTemplateTestRows().
			AddRow("template-1", "default", "Marketing package", "", "temporary_group", "test-user-001", "user", "active",
				"Create {{output}} for {{product}}.", `{"output":"launch plan"}`, `{}`, `{"name":"Marketing Launch Team"}`, `{}`, `["audit"]`, now, now, nil))

	mux := setupMux(t, "POST /api/v1/conversation-templates", s.HandleCreateConversationTemplate)
	rr := doAuthenticatedRequest(t, mux, "POST", "/api/v1/conversation-templates", `{
		"name":"Marketing package",
		"scope":"temporary_group",
		"template_body":"Create {{output}} for {{product}}.",
		"variables":{"output":"launch plan"},
		"recommended_team_shape":{"name":"Marketing Launch Team"},
		"governance_tags":["audit"]
	}`)
	assertStatus(t, rr, http.StatusCreated)

	var resp protocol.APIResponse
	assertJSON(t, rr, &resp)
	data, ok := resp.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected object payload, got %T", resp.Data)
	}
	if data["id"] != "template-1" || data["scope"] != "temporary_group" {
		t.Fatalf("unexpected response: %+v", data)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestHandleInstantiateConversationTemplate_ReturnsNonExecutingGroupDraft(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)
	now := time.Now().UTC()
	mock.ExpectQuery("SELECT id, tenant_id, name").
		WithArgs("template-2").
		WillReturnRows(conversationTemplateTestRows().
			AddRow("template-2", "default", "Launch package", "", "temporary_group", "admin", "soma", "active",
				"Create a launch package for {{product}}.", `{}`, `{}`, `{"name":"Launch Team","coordinator_profile":"Launch lead"}`, `{}`, `[]`, now, now, nil))
	mock.ExpectExec("UPDATE conversation_templates SET last_used_at").
		WithArgs("template-2").
		WillReturnResult(sqlmock.NewResult(0, 1))

	mux := setupMux(t, "POST /api/v1/conversation-templates/{id}/instantiate", s.HandleInstantiateConversationTemplate)
	rr := doAuthenticatedRequest(t, mux, "POST", "/api/v1/conversation-templates/template-2/instantiate", `{"variables":{"product":"Mycelis"}}`)
	assertStatus(t, rr, http.StatusOK)

	var resp protocol.APIResponse
	assertJSON(t, rr, &resp)
	data, ok := resp.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected object payload, got %T", resp.Data)
	}
	if data["rendered_prompt"] != "Create a launch package for Mycelis." {
		t.Fatalf("unexpected rendered prompt: %+v", data)
	}
	if _, ok := data["workflow_group"].(map[string]any); !ok {
		t.Fatalf("expected workflow group draft, got %+v", data)
	}
	if _, ok := data["team_ask"].(map[string]any); !ok {
		t.Fatalf("expected team ask, got %+v", data)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func conversationTemplateTestRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id",
		"tenant_id",
		"name",
		"description",
		"scope",
		"created_by",
		"creator_kind",
		"status",
		"template_body",
		"variables",
		"output_contract",
		"recommended_team_shape",
		"model_routing_hint",
		"governance_tags",
		"created_at",
		"updated_at",
		"last_used_at",
	})
}

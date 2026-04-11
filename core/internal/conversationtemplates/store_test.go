package conversationtemplates

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/pkg/protocol"
)

func TestStoreCreateConversationTemplate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	now := time.Now().UTC()
	rows := conversationTemplateRows().
		AddRow("template-1", "default", "Marketing package", "", "temporary_group", "admin", "user", "active",
			"Create {{output}}", `{"output":"brief"}`, `{"artifact":"document"}`, `{"name":"Marketing Team"}`, `{"model":"qwen3:8b"}`, `["audit"]`, now, now, nil)
	mock.ExpectQuery("INSERT INTO conversation_templates").
		WithArgs("default", "Marketing package", "", "temporary_group", "admin", "user", "active", "Create {{output}}", `{"output":"brief"}`, `{"artifact":"document"}`, `{"name":"Marketing Team"}`, `{"model":"qwen3:8b"}`, `["audit"]`).
		WillReturnRows(rows)

	created, err := NewStore(db).Create(context.Background(), protocol.ConversationTemplate{
		Name:                 "Marketing package",
		Scope:                protocol.ConversationTemplateScopeTemporaryGroup,
		CreatedBy:            "admin",
		TemplateBody:         "Create {{output}}",
		Variables:            map[string]any{"output": "brief"},
		OutputContract:       map[string]any{"artifact": "document"},
		RecommendedTeamShape: map[string]any{"name": "Marketing Team"},
		ModelRoutingHint:     map[string]any{"model": "qwen3:8b"},
		GovernanceTags:       []string{"audit"},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.ID != "template-1" || created.Scope != protocol.ConversationTemplateScopeTemporaryGroup {
		t.Fatalf("unexpected created template: %#v", created)
	}
	if created.Variables["output"] != "brief" {
		t.Fatalf("variables = %#v", created.Variables)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestStoreInstantiateUpdatesLastUsedAndReturnsTeamPackage(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	now := time.Now().UTC()
	rows := conversationTemplateRows().
		AddRow("template-2", "default", "Launch package", "", "temporary_group", "admin", "soma", "active",
			"Create a launch package for {{product}}.", `{}`, `{}`, `{"name":"Launch Team"}`, `{}`, `[]`, now, now, nil)
	mock.ExpectQuery("SELECT id, tenant_id, name").
		WithArgs("template-2").
		WillReturnRows(rows)
	mock.ExpectExec("UPDATE conversation_templates SET last_used_at").
		WithArgs("template-2").
		WillReturnResult(sqlmock.NewResult(0, 1))

	instantiation, err := NewStore(db).Instantiate(context.Background(), "template-2", map[string]any{"product": "Mycelis"})
	if err != nil {
		t.Fatalf("instantiate: %v", err)
	}
	if instantiation.RenderedPrompt != "Create a launch package for Mycelis." {
		t.Fatalf("rendered prompt = %q", instantiation.RenderedPrompt)
	}
	if instantiation.TeamAsk == nil || instantiation.WorkflowGroup == nil {
		t.Fatalf("expected team ask and workflow group, got %#v", instantiation)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func conversationTemplateRows() *sqlmock.Rows {
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

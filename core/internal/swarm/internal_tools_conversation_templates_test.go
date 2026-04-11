package swarm

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestHandleStoreConversationTemplate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	now := time.Now().UTC()
	mock.ExpectQuery("INSERT INTO conversation_templates").
		WithArgs("default", "Reusable launch ask", "", "temporary_group", "admin", "soma", "active", "Create {{output}} for {{product}}.", `{"output":"brief"}`, `{}`, `{"name":"Launch Team"}`, `{}`, `["audit"]`).
		WillReturnRows(conversationTemplateToolRows().
			AddRow("template-1", "default", "Reusable launch ask", "", "temporary_group", "admin", "soma", "active",
				"Create {{output}} for {{product}}.", `{"output":"brief"}`, `{}`, `{"name":"Launch Team"}`, `{}`, `["audit"]`, now, now, nil))

	registry := NewInternalToolRegistry(InternalToolDeps{DB: db})
	out, err := registry.handleStoreConversationTemplate(context.Background(), map[string]any{
		"name":                   "Reusable launch ask",
		"scope":                  "temporary_group",
		"template_body":          "Create {{output}} for {{product}}.",
		"variables":              map[string]any{"output": "brief"},
		"recommended_team_shape": map[string]any{"name": "Launch Team"},
		"governance_tags":        []any{"audit"},
	})
	if err != nil {
		t.Fatalf("store tool: %v", err)
	}
	if !strings.Contains(out, "Conversation template stored.") || !strings.Contains(out, "template-1") {
		t.Fatalf("unexpected output: %s", out)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestHandleInstantiateConversationTemplate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	now := time.Now().UTC()
	mock.ExpectQuery("SELECT id, tenant_id, name").
		WithArgs("template-2").
		WillReturnRows(conversationTemplateToolRows().
			AddRow("template-2", "default", "Launch package", "", "temporary_group", "admin", "soma", "active",
				"Create a launch package for {{product}}.", `{}`, `{}`, `{"name":"Launch Team"}`, `{}`, `[]`, now, now, nil))
	mock.ExpectExec("UPDATE conversation_templates SET last_used_at").
		WithArgs("template-2").
		WillReturnResult(sqlmock.NewResult(0, 1))

	registry := NewInternalToolRegistry(InternalToolDeps{DB: db})
	out, err := registry.handleInstantiateConversationTemplate(context.Background(), map[string]any{
		"template_id": "template-2",
		"variables":   map[string]any{"product": "Mycelis"},
	})
	if err != nil {
		t.Fatalf("instantiate tool: %v", err)
	}
	if !strings.Contains(out, "Conversation template instantiated without execution.") || !strings.Contains(out, "Create a launch package for Mycelis.") {
		t.Fatalf("unexpected output: %s", out)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func conversationTemplateToolRows() *sqlmock.Rows {
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

package server

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/configdocuments"
	"github.com/mycelis/core/internal/conversations"
	"github.com/mycelis/core/pkg/protocol"
)

func TestConfirmedConfigMutationPlanRejectsMixedExecution(t *testing.T) {
	configOnly, err := confirmedConfigMutationPlan(&protocol.ScopeValidation{PlannedToolCalls: []protocol.PlannedToolCall{
		{Name: "store_config_document"}, {Name: "delegate_task"},
	}})
	if err == nil || configOnly {
		t.Fatalf("mixed plan = %v, %v, want rejected", configOnly, err)
	}
	configOnly, err = confirmedConfigMutationPlan(&protocol.ScopeValidation{PlannedToolCalls: []protocol.PlannedToolCall{
		{Name: "store_config_document"}, {Name: "activate_config_document"},
	}})
	if err != nil || !configOnly {
		t.Fatalf("config plan = %v, %v, want accepted", configOnly, err)
	}
}

func TestConfigDocumentFromConfirmedArgumentsEnforcesRequestBoundary(t *testing.T) {
	base := retainedOutcomeTemplateDocument(t)
	boundary := &protocol.ConfigDocumentRequestBoundary{
		OrganizationID: "org-1",
		WorkspaceID:    "workspace-1",
		TeamID:         "team-1",
		OperatorID:     "operator-1",
	}
	tests := []struct {
		name       string
		scope      protocol.ConfigDocumentScope
		noBoundary bool
		invalidate bool
		wantErr    bool
	}{
		{name: "current organization", scope: protocol.ConfigDocumentScope{Kind: protocol.ConfigDocumentScopeOrganization, Ref: "org-1"}},
		{name: "current workspace", scope: protocol.ConfigDocumentScope{Kind: protocol.ConfigDocumentScopeWorkspace, Ref: "workspace-1"}},
		{name: "current team", scope: protocol.ConfigDocumentScope{Kind: protocol.ConfigDocumentScopeWorkspace, Ref: "team-1"}},
		{name: "current operator", scope: protocol.ConfigDocumentScope{Kind: protocol.ConfigDocumentScopeOperator, Ref: "operator-1"}},
		{name: "foreign workspace", scope: protocol.ConfigDocumentScope{Kind: protocol.ConfigDocumentScopeWorkspace, Ref: "workspace-2"}, wantErr: true},
		{name: "foreign operator", scope: protocol.ConfigDocumentScope{Kind: protocol.ConfigDocumentScopeOperator, Ref: "operator-2"}, wantErr: true},
		{name: "built in", scope: protocol.ConfigDocumentScope{Kind: protocol.ConfigDocumentScopeBuiltIn}, wantErr: true},
		{name: "missing approved boundary", scope: protocol.ConfigDocumentScope{Kind: protocol.ConfigDocumentScopeWorkspace, Ref: "workspace-1"}, noBoundary: true, wantErr: true},
		{name: "invalid parsed document", scope: protocol.ConfigDocumentScope{Kind: protocol.ConfigDocumentScopeWorkspace, Ref: "workspace-1"}, invalidate: true, wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			document := base
			document.Metadata.Scope = test.scope
			if test.invalidate {
				document.APIVersion = "unsupported/v2"
			}
			content, err := json.Marshal(document)
			if err != nil {
				t.Fatal(err)
			}
			approved := boundary
			if test.noBoundary {
				approved = nil
			}
			_, err = configDocumentFromConfirmedArguments(map[string]any{
				"format": "json", "content": string(content),
			}, approved)
			if (err != nil) != test.wantErr {
				t.Fatalf("validation error = %v, wantErr=%v", err, test.wantErr)
			}
		})
	}
}

func TestExecuteConfigDocumentMutationRejectsForeignWorkspaceBeforeStore(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectRollback()
	tx, err := db.BeginTx(t.Context(), nil)
	if err != nil {
		t.Fatal(err)
	}
	document := retainedOutcomeTemplateDocument(t)
	document.Metadata.Scope.Ref = "workspace-2"
	content, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	s := &AdminServer{DB: db}
	_, err = s.executeConfigDocumentMutationTx(t.Context(), tx, "store_config_document", map[string]any{
		"format": "json", "content": string(content),
	}, &protocol.ScopeValidation{ConfigRequestBoundary: &protocol.ConfigDocumentRequestBoundary{
		WorkspaceID: "workspace-1",
	}}, "operator-1")
	if err == nil || !strings.Contains(err.Error(), "outside the approved request boundary") {
		t.Fatalf("store error = %v, want request-boundary rejection", err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestConfigDocumentThreadReceiptFailureRemainsInCallerTransaction(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectExec("SELECT pg_advisory_xact_lock\\(hashtextextended\\(\\$1, 0\\)\\)").
		WithArgs("11111111-1111-1111-1111-111111111111").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("SELECT COALESCE\\(MAX\\(turn_index\\), -1\\) \\+ 1").
		WithArgs("11111111-1111-1111-1111-111111111111").
		WillReturnRows(sqlmock.NewRows([]string{"turn_index"}).AddRow(2))
	mock.ExpectExec("INSERT INTO conversation_turns").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO conversation_turns").WillReturnError(errors.New("receipt write failed"))
	mock.ExpectRollback()

	revision := configdocuments.RevisionRecord{
		RecordID: "22222222-2222-2222-2222-222222222222", TenantID: "default",
		Document: protocol.ConfigDocument{Metadata: protocol.ConfigDocumentMetadata{
			Name: "Launch brief", Version: "1.0.0",
			Scope: protocol.ConfigDocumentScope{Kind: protocol.ConfigDocumentScopeOrganization, Ref: "org-1"},
		}},
		Digest: "digest", CreatedAt: time.Now(),
	}
	output, _ := json.Marshal(map[string]any{"revision": revision})
	s := &AdminServer{DB: db, Conversations: conversations.NewStore(db)}
	tx, err := db.BeginTx(t.Context(), nil)
	if err != nil {
		t.Fatal(err)
	}
	err = s.logConfigDocumentThreadReceiptsTx(t.Context(), tx, &protocol.ScopeValidation{
		ConversationSessionID: "11111111-1111-1111-1111-111111111111",
	}, []plannedToolExecutionResult{{Name: "store_config_document", Output: string(output)}})
	if err == nil {
		t.Fatal("expected receipt failure")
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

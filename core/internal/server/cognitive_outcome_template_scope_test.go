package server

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/configdocuments"
	"github.com/mycelis/core/internal/conversations"
	"github.com/mycelis/core/pkg/protocol"
)

func TestValidateOutcomeTemplateRequestScopeUsesCurrentBoundary(t *testing.T) {
	document := retainedOutcomeTemplateDocument(t)
	tests := []struct {
		name         string
		scopeKind    protocol.ConfigDocumentScopeKind
		scopeRef     string
		organization string
		team         string
		actor        string
		wantErr      bool
	}{
		{name: "workspace organization", scopeKind: protocol.ConfigDocumentScopeWorkspace, scopeRef: "org-1", organization: "org-1", team: "team-2"},
		{name: "team fallback", scopeKind: protocol.ConfigDocumentScopeWorkspace, scopeRef: "team-1", team: "team-1"},
		{name: "team cannot cross organization", scopeKind: protocol.ConfigDocumentScopeWorkspace, scopeRef: "org-1", organization: "org-2", team: "org-1", wantErr: true},
		{name: "organization", scopeKind: protocol.ConfigDocumentScopeOrganization, scopeRef: "org-1", organization: "org-1"},
		{name: "cross organization", scopeKind: protocol.ConfigDocumentScopeOrganization, scopeRef: "org-1", organization: "org-2", wantErr: true},
		{name: "operator", scopeKind: protocol.ConfigDocumentScopeOperator, scopeRef: "user-1", actor: "user-1"},
		{name: "cross operator", scopeKind: protocol.ConfigDocumentScopeOperator, scopeRef: "user-1", actor: "user-2", wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			document.Metadata.Scope = protocol.ConfigDocumentScope{Kind: test.scopeKind, Ref: test.scopeRef}
			err := validateOutcomeTemplateRequestScope(document, test.organization, test.team, test.actor)
			if (err != nil) != test.wantErr {
				t.Fatalf("scope validation error = %v, wantErr=%v", err, test.wantErr)
			}
		})
	}
}

func TestConfigDocumentReceiptsKeepIdentityInInspectAndUseHumanSummary(t *testing.T) {
	document := retainedOutcomeTemplateDocument(t)
	digest, _ := protocol.CanonicalConfigDocumentDigest(document)
	recordID := "11111111-1111-1111-1111-111111111111"
	output, _ := json.Marshal(map[string]any{
		"revision": configdocuments.RevisionRecord{
			RecordID: recordID, Document: document, Digest: digest, ValidationState: "valid",
		},
	})
	result := plannedToolExecutionResult{Name: "store_config_document", Output: string(output)}
	receipt, ok := configDocumentReceiptFromResult(result)
	if !ok || receipt.RecordID != recordID || receipt.Digest != digest ||
		receipt.Name != document.Metadata.Name || receipt.Version != document.Metadata.Version {
		t.Fatalf("receipt did not preserve exact retained identity: %#v", receipt)
	}
	if summary := configDocumentResultSummary(result); summary != "Retained browser app v1.0.0 saved but not active." {
		t.Fatalf("visible summary = %q", summary)
	}
	activationOutput, _ := json.Marshal(map[string]any{
		"activation": configdocuments.ActivationResult{
			Revision: configdocuments.RevisionRecord{
				RecordID: recordID, Document: document, Digest: digest, ValidationState: "valid",
			},
		},
	})
	activeResult := plannedToolExecutionResult{Name: "activate_config_document", Output: string(activationOutput)}
	if summary := configDocumentResultSummary(activeResult); summary != "Retained browser app v1.0.0 is active for this workspace." {
		t.Fatalf("active summary = %q", summary)
	}
	if !isSynchronousConfigAction([]plannedToolExecutionResult{activeResult}) {
		t.Fatal("config activation must suppress generic work-started thread events")
	}
	if isSynchronousConfigAction([]plannedToolExecutionResult{
		activeResult,
		{Name: "delegate_task", Output: `{"status":"queued"}`},
	}) {
		t.Fatal("mixed config and delegated work must retain generic work-started thread events")
	}
	retained, ok := latestConfigDocumentReceiptFromTurns([]conversations.ConversationTurn{
		{Role: "assistant", Content: "Retained browser app v1.0.0 saved but not active."},
		{Role: "tool_result", ToolName: "store_config_document", Content: string(output)},
	})
	if !ok || retained.RecordID != recordID || retained.Digest != digest {
		t.Fatalf("reload receipt = %#v, want exact retained revision", retained)
	}
}

func TestResolveThreadOutcomeTemplateActivationRequiresSelectedRevisionIdentity(t *testing.T) {
	selected := retainedOutcomeTemplateDocument(t)
	selectedDigest, _ := protocol.CanonicalConfigDocumentDigest(selected)
	wrongDocument := selected
	wrongDocument.Metadata.ID = "same-scope-substitute"
	wrongDocumentDigest, _ := protocol.CanonicalConfigDocumentDigest(wrongDocument)
	wrongVersion := selected
	wrongVersion.Metadata.Version = "2.0.0"
	wrongVersionDigest, _ := protocol.CanonicalConfigDocumentDigest(wrongVersion)

	tests := []struct {
		name     string
		document protocol.ConfigDocument
		digest   string
		wantErr  bool
	}{
		{name: "matching revision", document: selected, digest: selectedDigest},
		{name: "same scope wrong document", document: wrongDocument, digest: wrongDocumentDigest, wantErr: true},
		{name: "same scope forged selected digest", document: wrongDocument, digest: selectedDigest, wantErr: true},
		{name: "same document wrong digest", document: selected, digest: "sha256:substituted", wantErr: true},
		{name: "same document wrong version", document: wrongVersion, digest: wrongVersionDigest, wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			withDatabase, mock := withDB(t)
			s := newTestServer(withDatabase)
			recordID := "44444444-4444-4444-4444-444444444444"
			mock.ExpectQuery("SELECT .*FROM config_documents.*WHERE tenant_id = \\$1 AND record_id = \\$2::uuid").
				WithArgs("default", recordID).
				WillReturnRows(serverConfigRevisionRows(recordID, test.document, test.digest))

			planned, err := s.resolveThreadOutcomeTemplateActivation(
				t.Context(), "", retainedOutcomeTemplateThread(), "workspace-1", "", "operator-1",
				[]protocol.PlannedToolCall{{Name: "activate_config_document", Arguments: map[string]any{
					"record_id": recordID, "action": "activate",
				}}},
			)
			if (err != nil) != test.wantErr {
				t.Fatalf("activation = (%#v, %v), wantErr=%v", planned, err, test.wantErr)
			}
			if !test.wantErr && firstNonEmptyString(planned[0].Arguments["record_id"]) != recordID {
				t.Fatalf("matching activation changed record_id: %#v", planned)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("unmet expectations: %v", err)
			}
		})
	}
}

func TestResolveThreadOutcomeTemplateActivationAcceptsMatchingRetainedReceipt(t *testing.T) {
	withDatabase, mock := withDB(t)
	s := newTestServer(withDatabase)
	s.Conversations = conversations.NewStore(s.getDB())
	document := retainedOutcomeTemplateDocument(t)
	digest, _ := protocol.CanonicalConfigDocumentDigest(document)
	recordID := "55555555-5555-5555-5555-555555555555"
	sessionID := "66666666-6666-6666-6666-666666666666"
	output, _ := json.Marshal(map[string]any{"revision": configdocuments.RevisionRecord{
		RecordID: recordID, Document: document, Digest: digest, ValidationState: "valid",
	}})
	turns := sqlmock.NewRows([]string{
		"id", "run_id", "session_id", "tenant_id", "agent_id", "team_id", "turn_index", "role", "content",
		"provider_id", "model_used", "tool_name", "tool_args", "parent_turn_id", "consultation_of", "created_at",
	}).AddRow(
		"turn-1", "", sessionID, "default", "admin", "admin-core", 0, "tool_result", string(output),
		"", "", "store_config_document", "", "", "", time.Now().UTC(),
	)
	mock.ExpectQuery("SELECT .+ FROM conversation_turns WHERE session_id = \\$1 ORDER BY").
		WithArgs(sessionID).WillReturnRows(turns)
	for range 2 {
		mock.ExpectQuery("SELECT .*FROM config_documents.*WHERE tenant_id = \\$1 AND record_id = \\$2::uuid").
			WithArgs("default", recordID).
			WillReturnRows(serverConfigRevisionRows(recordID, document, digest))
	}

	planned, err := s.resolveThreadOutcomeTemplateActivation(
		t.Context(), sessionID, nil, "workspace-1", "", "operator-1",
		[]protocol.PlannedToolCall{{Name: "activate_config_document", Arguments: map[string]any{
			"record_id": recordID, "action": "activate",
		}}},
	)
	if err != nil || firstNonEmptyString(planned[0].Arguments["record_id"]) != recordID {
		t.Fatalf("matching retained activation = (%#v, %v)", planned, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func retainedOutcomeTemplateDocument(t *testing.T) protocol.ConfigDocument {
	t.Helper()
	document, err := configdocuments.ParseDocument([]byte(retainedOutcomeTemplateYAML), "yaml")
	if err != nil {
		t.Fatalf("ParseDocument: %v", err)
	}
	return document
}

func retainedOutcomeTemplateThread() []chatRequestMessage {
	return []chatRequestMessage{
		{Role: "user", Content: "Preview and save this Outcome Template:\n```yaml\n" + retainedOutcomeTemplateYAML + "```"},
		{Role: "assistant", Content: "The Outcome Template revision was saved after approval."},
	}
}

func serverConfigRevisionRows(recordID string, document protocol.ConfigDocument, digest string) *sqlmock.Rows {
	secretRefs, _ := json.Marshal(document.Metadata.SecretRefs)
	governance, _ := json.Marshal(document.Metadata.Governance)
	return sqlmock.NewRows([]string{
		"record_id", "tenant_id", "document_id", "api_version", "kind", "name", "version",
		"owner_id", "scope_kind", "scope_ref", "enabled", "source_kind", "source_ref",
		"secret_refs", "governance", "spec", "digest", "validation_state", "created_by", "created_at",
	}).AddRow(
		recordID, "default", document.Metadata.ID, document.APIVersion, string(document.Kind), document.Metadata.Name,
		document.Metadata.Version, document.Metadata.OwnerID, string(document.Metadata.Scope.Kind), document.Metadata.Scope.Ref,
		document.Metadata.Enabled, string(document.Metadata.Source.Kind), document.Metadata.Source.Ref,
		string(secretRefs), string(governance), string(document.Spec), digest, "valid", "operator-1", time.Now().UTC(),
	)
}

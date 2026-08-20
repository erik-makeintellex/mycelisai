package server

import (
	"encoding/json"
	"testing"

	"github.com/mycelis/core/internal/configdocuments"
	"github.com/mycelis/core/pkg/protocol"
)

const retainedWorkerProfileYAML = `apiVersion: mycelis.ai/v1
kind: WorkerProfile
metadata:
  id: qa-builder
  name: Workspace builder
  version: alpha
  owner_id: operator-1
  scope: {kind: workspace, ref: workspace-1}
  enabled: true
  source: {kind: soma, ref: session:test}
  governance: {risk_level: medium, approval_posture: required}
spec:
  description: Builds bounded workspace deliverables.
  role: builder
  system_prompt: Build the approved work and return retained proof.
  capability_refs: [write_file, store_artifact]
  usage_policy: {selection: soma_or_manual, scope: workspace}
  inputs: [approved_work]
  outputs: [retained_output]
  verification_strategy: semantic
  verification_rubric: [The requested output is retained]
`

func TestResolveThreadConfigActivationUsesSavedWorkerProfile(t *testing.T) {
	withDatabase, mock := withDB(t)
	s := newTestServer(withDatabase)
	document := retainedWorkerProfileDocument(t, "alpha")
	digest, _ := protocol.CanonicalConfigDocumentDigest(document)
	recordID := "44444444-4444-4444-4444-444444444444"
	mock.ExpectQuery("SELECT .*FROM config_documents.*ORDER BY created_at DESC").
		WithArgs(
			"default", string(document.Kind), string(document.Metadata.Scope.Kind),
			document.Metadata.Scope.Ref, document.Metadata.ID, 20,
		).
		WillReturnRows(serverConfigRevisionRows(recordID, document, digest))

	planned, err := s.resolveThreadOutcomeTemplateActivation(
		t.Context(), "", retainedWorkerProfileThread(document, "Activate this Worker Profile."),
		"workspace-1", "", "operator-1",
		[]protocol.PlannedToolCall{{Name: "activate_config_document", Arguments: map[string]any{}}},
	)
	if err != nil {
		t.Fatalf("resolve Worker Profile activation: %v", err)
	}
	if got := firstNonEmptyString(planned[0].Arguments["record_id"]); got != recordID {
		t.Fatalf("record_id = %q, want %q", got, recordID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestResolveThreadConfigRollbackSelectsNamedWorkerProfileVersion(t *testing.T) {
	withDatabase, mock := withDB(t)
	s := newTestServer(withDatabase)
	threadDocument := retainedWorkerProfileDocument(t, "beta")
	targetDocument := retainedWorkerProfileDocument(t, "alpha")
	targetDigest, _ := protocol.CanonicalConfigDocumentDigest(targetDocument)
	targetRecordID := "55555555-5555-5555-5555-555555555555"
	mock.ExpectQuery("SELECT .*FROM config_documents.*ORDER BY created_at DESC").
		WithArgs(
			"default", string(threadDocument.Kind), string(threadDocument.Metadata.Scope.Kind),
			threadDocument.Metadata.Scope.Ref, threadDocument.Metadata.ID, 20,
		).
		WillReturnRows(serverConfigRevisionRows(targetRecordID, targetDocument, targetDigest))

	planned, err := s.resolveThreadOutcomeTemplateActivation(
		t.Context(), "", retainedWorkerProfileThread(threadDocument, "Roll back this Worker Profile to version alpha."),
		"workspace-1", "", "operator-1",
		[]protocol.PlannedToolCall{{Name: "activate_config_document", Arguments: map[string]any{}}},
	)
	if err != nil {
		t.Fatalf("resolve Worker Profile rollback: %v", err)
	}
	if got := firstNonEmptyString(planned[0].Arguments["record_id"]); got != targetRecordID {
		t.Fatalf("record_id = %q, want %q", got, targetRecordID)
	}
	if got := firstNonEmptyString(planned[0].Arguments["action"]); got != "rollback" {
		t.Fatalf("action = %q, want rollback", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestResolveThreadConfigRollbackRequiresTargetVersion(t *testing.T) {
	withDatabase, _ := withDB(t)
	s := newTestServer(withDatabase)
	document := retainedWorkerProfileDocument(t, "beta")
	planned, err := s.resolveThreadOutcomeTemplateActivation(
		t.Context(), "", retainedWorkerProfileThread(document, "Roll back this Worker Profile."),
		"workspace-1", "", "operator-1",
		[]protocol.PlannedToolCall{{Name: "activate_config_document", Arguments: map[string]any{"action": "rollback"}}},
	)
	if err == nil || planned != nil {
		t.Fatalf("ambiguous rollback = (%#v, %v), want rejection", planned, err)
	}
}

func retainedWorkerProfileDocument(t *testing.T, version string) protocol.ConfigDocument {
	t.Helper()
	document, err := configdocuments.ParseDocument([]byte(retainedWorkerProfileYAML), "yaml")
	if err != nil {
		t.Fatalf("ParseDocument: %v", err)
	}
	document.Metadata.Version = version
	return document
}

func retainedWorkerProfileThread(document protocol.ConfigDocument, request string) []chatRequestMessage {
	content, _ := json.Marshal(document)
	return []chatRequestMessage{
		{Role: "user", Content: "Save this Worker Profile:\n```json\n" + string(content) + "```"},
		{Role: "assistant", Content: "The Worker Profile revision was saved after approval."},
		{Role: "user", Content: request},
	}
}

package server

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/mycelis/core/internal/runs"
	"github.com/mycelis/core/pkg/protocol"
)

func TestHandleConfirmAction_CreatesCodexTaskBundleWithProof(t *testing.T) {
	workspace := t.TempDir()
	t.Setenv("MYCELIS_WORKSPACE", workspace)

	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)
	mock.MatchExpectationsInOrder(false)

	token := "41111111-1111-1111-1111-111111111111"
	proofID := "42222222-2222-2222-2222-222222222222"
	tokenUUID, err := uuid.Parse(token)
	if err != nil {
		t.Fatalf("parse token uuid: %v", err)
	}
	proofUUID, err := uuid.Parse(proofID)
	if err != nil {
		t.Fatalf("parse proof uuid: %v", err)
	}
	expiresAt := time.Now().Add(time.Hour)
	planned, ok := inferCodexTaskBundlePlanFromRequest("Generate a Codex task package for MYC-RC-001.")
	if !ok {
		t.Fatal("expected codex package plan")
	}
	scope := protocol.ScopeValidation{
		Tools:             []string{"write_file"},
		AffectedResources: affectedResourcesForPlannedCalls(planned),
		PlannedToolCalls:  planned,
		CapabilityIDs:     []string{"write_file"},
	}
	scopeJSON, err := json.Marshal(scope)
	if err != nil {
		t.Fatalf("marshal scope: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT intent_proof_id, consumed, expires_at FROM confirm_tokens WHERE token = \\$1").
		WithArgs(tokenUUID).
		WillReturnRows(sqlmock.NewRows([]string{"intent_proof_id", "consumed", "expires_at"}).
			AddRow(proofID, false, expiresAt))
	mock.ExpectExec("UPDATE confirm_tokens SET consumed = TRUE, consumed_at = \\$1 WHERE token = \\$2 AND consumed = FALSE").
		WithArgs(sqlmock.AnyArg(), tokenUUID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("SELECT scope_validation FROM intent_proofs WHERE id = \\$1").
		WithArgs(proofUUID).
		WillReturnRows(sqlmock.NewRows([]string{"scope_validation"}).AddRow(scopeJSON))
	mock.ExpectExec("INSERT INTO mission_runs \\(id, mission_id, tenant_id, status, run_depth, started_at\\)").
		WithArgs(sqlmock.AnyArg(), proofID, "default", runs.StatusRunning, 0, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	for i := 0; i < 10; i++ {
		mock.ExpectExec("INSERT INTO log_entries \\(id, trace_id, timestamp, level, source, intent, message, context\\)").
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "audit", "confirm-action", string(protocol.TemplateChatToProposal), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 1))
	}
	mock.ExpectExec("INSERT INTO mission_events").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "default", string(protocol.EventArtifactCreated), string(protocol.SeverityInfo), "admin", "governance", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE mission_runs SET status = \\$1, completed_at = NOW\\(\\) WHERE id = \\$2").
		WithArgs(runs.StatusCompleted, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO mission_events").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "default", string(protocol.EventMissionCompleted), string(protocol.SeverityInfo), "admin", "governance", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE intent_proofs SET status = 'confirmed', confirmed_at = \\$1 WHERE id = \\$2").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	reqBody := `{"confirm_token":"` + token + `"}`
	rr := doRequest(t, http.HandlerFunc(s.HandleConfirmAction), http.MethodPost, "/api/v1/intent/confirm-action", reqBody)
	assertStatus(t, rr, http.StatusOK)

	var resp protocol.APIResponse
	assertJSON(t, rr, &resp)
	if !resp.OK {
		t.Fatalf("expected ok=true, got error=%q", resp.Error)
	}
	data, ok := resp.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected response data map, got %T", resp.Data)
	}
	if data["execution_state"] != "verified" {
		t.Fatalf("execution_state = %v, want verified", data["execution_state"])
	}
	pkg, ok := data["package"].(map[string]any)
	if !ok {
		t.Fatalf("expected package metadata, got %#v", data["package"])
	}
	if pkg["package_kind"] != "codex_task_bundle" {
		t.Fatalf("package_kind = %v", pkg["package_kind"])
	}
	if pkg["proof_path"] != "codex-task-packages/myc-rc-001/proof.md" {
		t.Fatalf("proof_path = %v", pkg["proof_path"])
	}
	if pkg["package_id"] != "codex_task_bundle:myc-rc-001" {
		t.Fatalf("package_id = %v", pkg["package_id"])
	}
	outputRefs, ok := data["output_refs"].([]any)
	if !ok || len(outputRefs) != 5 {
		t.Fatalf("output_refs = %#v, want 5 normalized refs", data["output_refs"])
	}
	firstRef, ok := outputRefs[0].(map[string]any)
	if !ok {
		t.Fatalf("first output ref = %#v, want map", outputRefs[0])
	}
	if firstRef["role"] != "primary" || firstRef["source_run_id"] == "" || firstRef["source_contract_id"] != proofID {
		t.Fatalf("first output ref metadata = %#v", firstRef)
	}
	proofRef, ok := outputRefs[4].(map[string]any)
	if !ok {
		t.Fatalf("proof output ref = %#v, want map", outputRefs[4])
	}
	if proofRef["role"] != "proof" || proofRef["storage_ref"] != "codex-task-packages/myc-rc-001/proof.md" {
		t.Fatalf("proof output ref metadata = %#v", proofRef)
	}
	deliveryPackage, ok := data["delivery_package"].(map[string]any)
	if !ok {
		t.Fatalf("delivery_package = %#v, want map", data["delivery_package"])
	}
	if deliveryPackage["package_id"] != "codex_task_bundle:myc-rc-001" || deliveryPackage["status"] != "verified" {
		t.Fatalf("delivery package metadata = %#v", deliveryPackage)
	}
	executionSummary, ok := data["execution_summary"].(map[string]any)
	if !ok {
		t.Fatalf("execution_summary = %#v, want map", data["execution_summary"])
	}
	if executionSummary["status"] != "verified" {
		t.Fatalf("execution summary = %#v", executionSummary)
	}
	artifacts, ok := data["artifacts"].([]any)
	if !ok || len(artifacts) != 5 {
		t.Fatalf("artifacts = %#v, want 5 package artifacts", data["artifacts"])
	}

	for _, name := range []string{
		"implementation_task.md",
		"validation_checklist.md",
		"affected_files_summary.md",
		"next_recommended_slices.md",
		"proof.md",
	} {
		if _, err := os.Stat(filepath.Join(workspace, "codex-task-packages", "myc-rc-001", name)); err != nil {
			t.Fatalf("expected package file %s: %v", name, err)
		}
	}

	proof, err := os.ReadFile(filepath.Join(workspace, "codex-task-packages", "myc-rc-001", "proof.md"))
	if err != nil {
		t.Fatalf("read proof: %v", err)
	}
	if !strings.Contains(string(proof), "Intent proof ID: `"+proofID+"`") {
		t.Fatalf("proof missing intent proof id: %s", string(proof))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet db expectations: %v", err)
	}
}

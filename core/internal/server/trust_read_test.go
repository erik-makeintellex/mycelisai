package server

import (
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/pkg/protocol"
)

func TestHandleListProofArtifacts_ReturnsProofArtifacts(t *testing.T) {
	opt, mock := withDB(t)
	s := newTestServer(opt)
	now := time.Now().UTC()
	contractID := "11111111-1111-1111-1111-111111111111"
	mock.ExpectQuery("SELECT id::text, COALESCE\\(contract_id::text").
		WithArgs(contractID, "", "", "", 3).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "contract_id", "intent_proof_id", "run_id", "artifact_kind", "status", "proof_class",
			"validation_source", "evidence_strength", "proof_quality", "output_refs", "audit_refs",
			"review_lineage", "degradation", "recovery", "payload", "created_at",
		}).AddRow(
			"44444444-4444-4444-4444-444444444444",
			contractID,
			"",
			"",
			"confirm_action",
			string(protocol.ProofArtifactStatusSuccess),
			string(protocol.ExecutionProofClassRunAudit),
			string(protocol.TrustValidationSourceConfirmAction),
			string(protocol.TrustEvidenceStrengthRunAudit),
			string(protocol.TrustProofQualityVerified),
			[]byte(`[]`),
			[]byte(`[]`),
			[]byte(`[]`),
			[]byte(`{}`),
			[]byte(`{"recovery_state":"verified"}`),
			[]byte(`{"summary":"done"}`),
			now,
		))

	mux := setupMux(t, "GET /api/v1/trust/proof-artifacts", s.HandleListProofArtifacts)
	rr := doRequest(t, mux, http.MethodGet, "/api/v1/trust/proof-artifacts?contract_id="+contractID+"&limit=3", "")

	assertStatus(t, rr, http.StatusOK)
	var resp map[string]any
	assertJSON(t, rr, &resp)
	items := resp["data"].([]any)
	if len(items) != 1 {
		t.Fatalf("expected 1 proof artifact, got %d", len(items))
	}
	first := items[0].(map[string]any)
	if first["id"] != "44444444-4444-4444-4444-444444444444" {
		t.Fatalf("id = %v", first["id"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestHandleGetExecutionContract_RejectsInvalidID(t *testing.T) {
	opt, mock := withDB(t)
	s := newTestServer(opt)
	mux := setupMux(t, "GET /api/v1/trust/execution-contracts/{id}", s.HandleGetExecutionContract)
	rr := doRequest(t, mux, http.MethodGet, "/api/v1/trust/execution-contracts/not-a-uuid", "")

	assertStatus(t, rr, http.StatusBadRequest)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

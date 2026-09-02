package server

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/dispatchoutbox"
	"github.com/mycelis/core/internal/workerauthority"
	"github.com/mycelis/core/internal/workers"
)

type frameworkRunCreateIntent struct {
	Request workers.WorkerRunRequest `json:"request"`
}

// stageFrameworkRunCreateTx is the Slice A commit boundary. It records Core's
// authority and the exact external-create intent without performing I/O.
// Slice C may activate and dispatch the committed outbox row.
func (s *AdminServer) stageFrameworkRunCreateTx(ctx context.Context, tx *sql.Tx, request workers.WorkerRunRequest) (bool, error) {
	if s == nil || s.WorkerAuthority == nil || s.DispatchOutbox == nil {
		return false, fmt.Errorf("framework run authority requires binding and outbox stores")
	}
	if err := validateFrameworkCreateRequest(request); err != nil {
		return false, err
	}
	payload, err := json.Marshal(frameworkRunCreateIntent{Request: request})
	if err != nil {
		return false, err
	}
	digestBytes := sha256.Sum256(payload)
	digest := hex.EncodeToString(digestBytes[:])
	c := request.Correlation
	bindingCreated, err := s.WorkerAuthority.StageBindingTx(ctx, tx, workerauthority.Binding{
		RunID: c.RunID, IntentProofID: c.IntentProofID,
		ExecutionContractID: c.ExecutionContractID, TeamID: c.TeamID,
		WorkItemID: c.WorkItemID, OutcomeID: c.OutcomeID,
		IdempotencyKey: c.IdempotencyKey, GraphRevision: c.GraphRevision,
		SourceKind: c.SourceKind, SourceChannel: c.SourceChannel,
		PayloadKind: c.PayloadKind, RequestDigest: digest,
	})
	if err != nil {
		return false, err
	}
	outboxCreated, err := s.DispatchOutbox.EnqueueFrameworkCreateTx(ctx, tx, dispatchoutbox.Item{
		ID: uuid.NewString(), IdempotencyKey: "framework-run-create:" + c.IdempotencyKey,
		RunID: c.RunID, IntentProofID: c.IntentProofID,
		ContractID: c.ExecutionContractID, TeamID: c.TeamID, WorkItemID: c.WorkItemID,
		SourceKind: c.SourceKind, SourceChannel: c.SourceChannel,
		PayloadKind: c.PayloadKind, Payload: payload,
		Recovery: json.RawMessage(`{"action":"activate_framework_run_create_after_slice_c","operator_required":false}`),
	})
	if err != nil {
		return false, err
	}
	if bindingCreated != outboxCreated {
		return false, workerauthority.ErrConflict
	}
	return bindingCreated, nil
}

func validateFrameworkCreateRequest(request workers.WorkerRunRequest) error {
	c := request.Correlation
	for label, value := range map[string]string{
		"run_id": request.RunID, "correlation.run_id": c.RunID,
		"intent_proof_id": c.IntentProofID, "execution_contract_id": c.ExecutionContractID,
		"work_item_id":    c.WorkItemID,
		"idempotency_key": c.IdempotencyKey, "graph_revision": c.GraphRevision,
		"source_kind": c.SourceKind, "source_channel": c.SourceChannel,
		"payload_kind": c.PayloadKind, "intent": request.Intent,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("framework run create %s is required", label)
		}
	}
	if request.RunID != c.RunID {
		return fmt.Errorf("framework run create authority identity mismatch")
	}
	if c.PayloadKind != "command" || c.SourceKind != "web_api" {
		return fmt.Errorf("framework run create vocabulary is not canonical")
	}
	return nil
}

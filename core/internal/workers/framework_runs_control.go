package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// StopRun cannot invent a Core command identity. External callers must use
// StopRunCommand; central keeps the legacy interface for local finalization.
func (b *FrameworkRunsBackend) StopRun(context.Context, string) error {
	return fmt.Errorf("durable stop requires Core command_id and expected_version")
}

func (b *FrameworkRunsBackend) SubmitApproval(ctx context.Context, runID string, approval WorkerApprovalDecision) error {
	_, err := b.SubmitApprovalCommand(ctx, runID, approval)
	return err
}

func (b *FrameworkRunsBackend) StopRunCommand(ctx context.Context, runID string, command WorkerStopCommand) (WorkerControlReceipt, error) {
	if err := validateControl(runID, command.CommandID, command.ActorID, command.ExpectedVersion); err != nil {
		return WorkerControlReceipt{}, err
	}
	receipt, err := b.postControl(ctx, "/v1/runs/"+url.PathEscape(runID)+"/stop", command, runID, command.CommandID)
	if err == nil && receipt.Kind != "stop" {
		return WorkerControlReceipt{}, WorkerBackendError("invalid_backend_response", "External worker returned the wrong control kind.", false)
	}
	return receipt, err
}

func (b *FrameworkRunsBackend) SubmitApprovalCommand(ctx context.Context, runID string, approval WorkerApprovalDecision) (WorkerControlReceipt, error) {
	if strings.TrimSpace(approval.ApprovalID) == "" {
		return WorkerControlReceipt{}, fmt.Errorf("worker approval_id is required")
	}
	if approval.Decision != DecisionApprove && approval.Decision != DecisionDeny {
		return WorkerControlReceipt{}, fmt.Errorf("unsupported approval decision %q", approval.Decision)
	}
	if err := validateControl(runID, approval.CommandID, approval.ActorID, approval.ExpectedVersion); err != nil {
		return WorkerControlReceipt{}, err
	}
	path := "/v1/runs/" + url.PathEscape(runID) + "/approvals/" + url.PathEscape(approval.ApprovalID)
	receipt, err := b.postControl(ctx, path, approval, runID, approval.CommandID)
	if err == nil && receipt.Kind != string(approval.Decision) {
		return WorkerControlReceipt{}, WorkerBackendError("invalid_backend_response", "External worker returned the wrong control kind.", false)
	}
	return receipt, err
}

func validateControl(runID, commandID, actorID string, expectedVersion int64) error {
	for label, value := range map[string]string{"run_id": runID, "command_id": commandID, "actor_id": actorID} {
		if strings.TrimSpace(value) == "" || strings.TrimSpace(value) != value {
			return fmt.Errorf("worker control %s is required and canonical", label)
		}
	}
	if expectedVersion < 1 {
		return fmt.Errorf("worker control expected_version must be positive")
	}
	return nil
}

func (b *FrameworkRunsBackend) postControl(ctx context.Context, path string, payload any, runID, commandID string) (WorkerControlReceipt, error) {
	req, err := b.newRequest(ctx, http.MethodPost, path, payload)
	if err != nil {
		return WorkerControlReceipt{}, err
	}
	res, err := b.Client.Do(req)
	if err != nil {
		return WorkerControlReceipt{}, fmt.Errorf("framework runs control request failed: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusAccepted && res.StatusCode != http.StatusOK {
		return WorkerControlReceipt{}, statusError("framework runs control", res)
	}
	var raw map[string]any
	if json.NewDecoder(res.Body).Decode(&raw) != nil {
		return WorkerControlReceipt{}, WorkerBackendError("invalid_backend_response", "External worker returned invalid control JSON.", false)
	}
	return controlReceiptFromMap(raw, runID, commandID, res.StatusCode == http.StatusOK)
}

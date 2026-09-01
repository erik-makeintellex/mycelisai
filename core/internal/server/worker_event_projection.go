package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/runs"
	"github.com/mycelis/core/internal/workers"
	"github.com/mycelis/core/pkg/protocol"
)

// projectWorkerEvent is the durable, idempotent boundary from a normalized
// WorkerEvent into Mycelis-owned mission state. It is intentionally not wired
// to framework_runs selection until the external runtime passes its remaining
// production gates.
func (s *AdminServer) projectWorkerEvent(ctx context.Context, correlation workers.WorkerCorrelation, backendRunID string, event workers.WorkerEvent) (bool, error) {
	if s == nil || s.getDB() == nil {
		return false, fmt.Errorf("worker event projection requires database")
	}
	if err := validateWorkerProjection(correlation, backendRunID, event); err != nil {
		return false, err
	}
	eventType, severity := projectedWorkerState(event)
	payloadJSON, err := json.Marshal(workerEventPayload(correlation, backendRunID, event))
	if err != nil {
		return false, err
	}
	projectionID := uuid.NewSHA1(uuid.NameSpaceURL, []byte(strings.Join([]string{
		"mycelis", "worker-event", correlation.RunID, string(event.Backend), event.EventID,
	}, ":"))).String()
	emittedAt := event.Timestamp
	if emittedAt.IsZero() {
		emittedAt = time.Now().UTC()
	}
	tx, err := s.getDB().BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	var inserted string
	err = tx.QueryRowContext(ctx, `
INSERT INTO mission_events
    (id, run_id, tenant_id, event_type, severity, source_agent, source_team, payload, emitted_at)
VALUES ($1,$2,'default',$3,$4,'external-worker',NULLIF($5,''),$6,$7)
ON CONFLICT (id) DO NOTHING
RETURNING id::text`, projectionID, correlation.RunID, string(eventType), string(severity), correlation.TeamID, payloadJSON, emittedAt).Scan(&inserted)
	if err == sql.ErrNoRows {
		return false, tx.Commit()
	}
	if err != nil {
		return false, fmt.Errorf("persist worker event projection: %w", err)
	}
	if event.Kind != workers.EventFailed && event.Kind != workers.EventCancelled {
		_, err = tx.ExecContext(ctx, `
UPDATE mission_runs SET status=$2
WHERE id=$1 AND status NOT IN ($3,$4)`, correlation.RunID, runs.StatusRunning, runs.StatusCompleted, runs.StatusFailed)
		if err != nil {
			return false, fmt.Errorf("update projected mission state: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return false, err
	}
	return true, nil
}

func validateWorkerProjection(correlation workers.WorkerCorrelation, backendRunID string, event workers.WorkerEvent) error {
	for _, value := range []string{
		correlation.RunID, correlation.IntentProofID, correlation.ExecutionContractID,
		correlation.WorkItemID, correlation.IdempotencyKey, correlation.SourceKind,
		correlation.SourceChannel, correlation.PayloadKind, correlation.GraphRevision,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("worker event is missing required Mycelis correlation")
		}
	}
	if strings.TrimSpace(event.EventID) == "" {
		return fmt.Errorf("worker event_id is required for durable projection")
	}
	if event.Backend != workers.BackendFrameworkRuns {
		return fmt.Errorf("worker event projection only accepts framework_runs evidence")
	}
	if event.BackendRunID != "" && strings.TrimSpace(event.BackendRunID) != strings.TrimSpace(backendRunID) {
		return fmt.Errorf("worker event backend_run_id does not match durable binding")
	}
	eventRunID := strings.TrimSpace(event.RunID)
	if eventRunID != "" && eventRunID != correlation.RunID && eventRunID != strings.TrimSpace(backendRunID) {
		return fmt.Errorf("worker event run_id does not match correlated run")
	}
	switch event.Kind {
	case workers.EventAccepted, workers.EventProgress, workers.EventApprovalNeeded, workers.EventCompleted, workers.EventFailed, workers.EventCancelled:
		return nil
	default:
		return fmt.Errorf("unsupported worker event kind %q", event.Kind)
	}
}

func projectedWorkerState(event workers.WorkerEvent) (protocol.EventType, protocol.EventSeverity) {
	switch event.Kind {
	case workers.EventCompleted:
		return protocol.EventTeamWorkStatus, protocol.SeverityInfo
	case workers.EventFailed:
		return protocol.EventTeamWorkStatus, protocol.SeverityWarn
	case workers.EventCancelled:
		return protocol.EventTeamWorkStatus, protocol.SeverityWarn
	case workers.EventApprovalNeeded:
		return protocol.EventTeamWorkStatus, protocol.SeverityWarn
	case workers.EventProgress:
		return protocol.EventTeamWorkStatus, protocol.SeverityInfo
	default:
		return protocol.EventMissionStarted, protocol.SeverityInfo
	}
}

func workerEventPayload(correlation workers.WorkerCorrelation, backendRunID string, event workers.WorkerEvent) map[string]any {
	payload := map[string]any{
		"worker_event_id":       event.EventID,
		"worker_backend":        string(event.Backend),
		"backend_run_id":        strings.TrimSpace(backendRunID),
		"worker_event_kind":     string(event.Kind),
		"worker_status":         string(event.Status),
		"run_id":                correlation.RunID,
		"intent_proof_id":       correlation.IntentProofID,
		"execution_contract_id": correlation.ExecutionContractID,
		"work_item_id":          correlation.WorkItemID,
		"idempotency_key":       correlation.IdempotencyKey,
		"source_kind":           correlation.SourceKind,
		"source_channel":        correlation.SourceChannel,
		"payload_kind":          correlation.PayloadKind,
		"graph_revision":        correlation.GraphRevision,
	}
	if correlation.TeamID != "" {
		payload["team_id"] = correlation.TeamID
	}
	if correlation.OutcomeID != "" {
		payload["outcome_id"] = correlation.OutcomeID
	}
	if event.Result != nil {
		payload["output_count"] = len(event.Result.Outputs)
	}
	if event.Kind == workers.EventCompleted {
		payload["completion_authority"] = "candidate"
		payload["requires_core_validation"] = true
		payload["verified"] = false
	}
	if event.Error != nil {
		payload["error_code"] = strings.TrimSpace(event.Error.Code)
		payload["recoverable"] = event.Error.Recoverable
	}
	if event.Approval != nil {
		payload["approval_id"] = strings.TrimSpace(event.Approval.ID)
		payload["approval_kind"] = strings.TrimSpace(event.Approval.Kind)
	}
	if event.Usage != nil {
		payload["usage"] = map[string]any{
			"input_tokens":  event.Usage.InputTokens,
			"output_tokens": event.Usage.OutputTokens,
			"duration_ms":   event.Usage.DurationMS,
		}
	}
	return payload
}

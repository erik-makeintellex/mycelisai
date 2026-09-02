package journal

import (
	"fmt"
	"time"

	"github.com/mycelis/framework-runs/internal/protocol"
)

func ApplyOutcome(run *protocol.Run, command Command, outcome protocol.ExecutorOutcome, now time.Time) (protocol.Event, error) {
	if err := protocol.ValidateOutcome(&outcome, run.RunID); err != nil {
		return protocol.Event{}, err
	}
	if err := legalTransition(run.Status, outcome.Status, command.Kind); err != nil {
		return protocol.Event{}, err
	}
	run.Status = outcome.Status
	run.Version++
	run.UpdatedAt = now.UTC()
	run.Approval = protocol.Clone(outcome.Approval)
	run.Result = protocol.Clone(outcome.Result)
	run.Error = protocol.Clone(outcome.Error)
	run.Usage = protocol.Clone(outcome.Usage)
	run.Metadata = mergeMetadata(run.Metadata, outcome.Metadata)
	run.Metadata["execution_authority"] = "mycelis_core"
	run.Metadata["storage"] = "durable_journal"

	event := protocol.Event{
		EventID:     eventID(run.RunID, run.Version),
		Version:     run.Version,
		RunID:       run.RunID,
		Correlation: run.Correlation,
		Kind:        eventKind(outcome.Status),
		Status:      outcome.Status,
		Timestamp:   now.UTC(),
		Message:     outcome.Message,
		Approval:    protocol.Clone(outcome.Approval),
		Result:      protocol.Clone(outcome.Result),
		Error:       protocol.Clone(outcome.Error),
		Usage:       protocol.Clone(outcome.Usage),
		Metadata:    mergeMetadata(outcome.Metadata, map[string]any{"execution_authority": "mycelis_core"}),
	}
	return event, nil
}

func mergeMetadata(base, additions map[string]any) map[string]any {
	merged := protocol.Clone(base)
	if merged == nil {
		merged = map[string]any{}
	}
	for key, value := range additions {
		merged[key] = protocol.Clone(value)
	}
	return merged
}

func AcceptedEvent(run protocol.Run) protocol.Event {
	return protocol.Event{
		EventID:     eventID(run.RunID, 1),
		Sequence:    1,
		Version:     1,
		RunID:       run.RunID,
		Correlation: run.Correlation,
		Kind:        protocol.EventAccepted,
		Status:      protocol.StatusAccepted,
		Timestamp:   run.CreatedAt,
		Message:     "Run accepted by framework Runs service.",
	}
}

func eventID(runID string, version uint64) string {
	digest, _ := protocol.Digest(map[string]any{"run_id": runID, "version": version})
	return "evt:" + digest[:32]
}

func eventKind(status protocol.Status) protocol.EventKind {
	switch status {
	case protocol.StatusRunning:
		return protocol.EventProgress
	case protocol.StatusApprovalNeeded:
		return protocol.EventApprovalNeeded
	case protocol.StatusCompleted:
		return protocol.EventCompleted
	case protocol.StatusFailed:
		return protocol.EventFailed
	case protocol.StatusCancelled:
		return protocol.EventCancelled
	default:
		return ""
	}
}

func legalTransition(from, to protocol.Status, commandKind string) error {
	if from == protocol.StatusCompleted || from == protocol.StatusFailed || from == protocol.StatusCancelled {
		return fmt.Errorf("%w: terminal run cannot transition", ErrConflict)
	}
	if commandKind == "stop" && to != protocol.StatusCancelled {
		return fmt.Errorf("%w: stop must produce cancelled", ErrConflict)
	}
	if commandKind == "deny" && to != protocol.StatusFailed {
		return fmt.Errorf("%w: denial must produce failed", ErrConflict)
	}
	allowed := map[protocol.Status]map[protocol.Status]bool{
		protocol.StatusAccepted: {
			protocol.StatusRunning: true, protocol.StatusApprovalNeeded: true,
			protocol.StatusCompleted: true, protocol.StatusFailed: true, protocol.StatusCancelled: true,
		},
		protocol.StatusRunning: {
			protocol.StatusRunning: true, protocol.StatusApprovalNeeded: true,
			protocol.StatusCompleted: true, protocol.StatusFailed: true, protocol.StatusCancelled: true,
		},
		protocol.StatusApprovalNeeded: {
			protocol.StatusRunning: true, protocol.StatusFailed: true,
			protocol.StatusCompleted: true, protocol.StatusCancelled: true,
		},
	}
	if !allowed[from][to] {
		return fmt.Errorf("%w: illegal status transition %s to %s", ErrConflict, from, to)
	}
	return nil
}

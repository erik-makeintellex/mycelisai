package workers

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type CentralBackend struct {
	mu   sync.RWMutex
	runs map[string]WorkerRunHandle
}

func NewCentralBackend() *CentralBackend {
	return &CentralBackend{runs: map[string]WorkerRunHandle{}}
}

func (b *CentralBackend) CreateRun(_ context.Context, req WorkerRunRequest) (WorkerRunHandle, error) {
	if req.Intent == "" {
		return WorkerRunHandle{}, fmt.Errorf("worker run intent is required")
	}
	now := time.Now().UTC()
	handle := WorkerRunHandle{
		RunID:     uuid.NewString(),
		Backend:   BackendCentral,
		Status:    StatusAccepted,
		Protocol:  ProtocolRunsAPI,
		CreatedAt: now,
		UpdatedAt: now,
		AuditRecord: &WorkerAuditRecord{
			Backend:      BackendCentral,
			ActorID:      req.UserID,
			DecisionPath: []string{"policy.accepted", "backend.central"},
			CreatedAt:    now,
		},
		Metadata: map[string]any{"intent": req.Intent},
	}
	handle.AuditRecord.RunID = handle.RunID

	b.mu.Lock()
	b.runs[handle.RunID] = handle
	b.mu.Unlock()
	return handle, nil
}

func (b *CentralBackend) StreamRunEvents(ctx context.Context, runID string) (<-chan WorkerEvent, error) {
	handle, err := b.GetRun(ctx, runID)
	if err != nil {
		return nil, err
	}
	events := make(chan WorkerEvent, 3)
	go func() {
		defer close(events)
		now := time.Now().UTC()
		send := func(event WorkerEvent) bool {
			select {
			case <-ctx.Done():
				return false
			case events <- event:
				return true
			}
		}
		if !send(WorkerEvent{RunID: runID, Backend: BackendCentral, Kind: EventAccepted, Status: handle.Status, Message: "Run accepted.", Timestamp: now}) {
			return
		}
		if handle.Status == StatusAccepted {
			_ = b.setStatus(runID, StatusCompleted)
			result := WorkerResult{Summary: "Central worker accepted the run.", FinishedAt: time.Now().UTC()}
			_ = b.setResult(runID, result)
			send(WorkerEvent{RunID: runID, Backend: BackendCentral, Kind: EventCompleted, Status: StatusCompleted, Message: result.Summary, Result: &result, Timestamp: time.Now().UTC()})
		}
	}()
	return events, nil
}

func (b *CentralBackend) GetRun(_ context.Context, runID string) (WorkerRunHandle, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	run, ok := b.runs[runID]
	if !ok {
		return WorkerRunHandle{}, fmt.Errorf("worker run not found: %s", runID)
	}
	return run, nil
}

func (b *CentralBackend) StopRun(_ context.Context, runID string) error {
	return b.setStatus(runID, StatusCancelled)
}

func (b *CentralBackend) SubmitApproval(_ context.Context, runID string, decision WorkerApprovalDecision) error {
	if decision.Decision != DecisionApprove && decision.Decision != DecisionDeny {
		return fmt.Errorf("unsupported approval decision %q", decision.Decision)
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	run, ok := b.runs[runID]
	if !ok {
		return fmt.Errorf("worker run not found: %s", runID)
	}
	if run.Approval == nil || run.Approval.ID != decision.ApprovalID {
		return fmt.Errorf("approval request not found for run %s", runID)
	}
	if decision.Decision == DecisionDeny {
		run.Status = StatusFailed
		run.Error = &WorkerError{Code: "approval_denied", Message: "Operator denied worker approval.", Recoverable: true}
	} else {
		run.Status = StatusRunning
		run.Approval = nil
	}
	run.UpdatedAt = time.Now().UTC()
	b.runs[runID] = run
	return nil
}

func (b *CentralBackend) GetCapabilities(context.Context) (WorkerCapabilities, error) {
	return WorkerCapabilities{
		Backend:              BackendCentral,
		Healthy:              true,
		SupportedProtocols:   []Protocol{ProtocolRunsAPI},
		SupportsEvents:       true,
		SupportsCancellation: true,
		SupportsApprovals:    true,
		SupportsUsage:        true,
		Features:             []string{"managed_workers", "central_policy", "central_audit"},
	}, nil
}

func (b *CentralBackend) HealthCheck(context.Context) (WorkerHealth, error) {
	return WorkerHealth{Backend: BackendCentral, Healthy: true, Message: "central worker backend ready", CheckedAt: time.Now().UTC()}, nil
}

func (b *CentralBackend) setStatus(runID string, status RunStatus) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	run, ok := b.runs[runID]
	if !ok {
		return fmt.Errorf("worker run not found: %s", runID)
	}
	run.Status = status
	run.UpdatedAt = time.Now().UTC()
	b.runs[runID] = run
	return nil
}

func (b *CentralBackend) setResult(runID string, result WorkerResult) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	run, ok := b.runs[runID]
	if !ok {
		return fmt.Errorf("worker run not found: %s", runID)
	}
	run.Result = &result
	run.Status = StatusCompleted
	run.UpdatedAt = time.Now().UTC()
	b.runs[runID] = run
	return nil
}

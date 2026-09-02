package controller

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mycelis/framework-runs/internal/executor"
	"github.com/mycelis/framework-runs/internal/journal"
	"github.com/mycelis/framework-runs/internal/protocol"
)

var ErrExecutorUnavailable = errors.New("no production executor is configured")

type FaultPoint string

const (
	FaultAfterExecutorEffect FaultPoint = "after_executor_effect_before_commit"
)

type FaultHook func(FaultPoint, journal.Command) error

type Service struct {
	Journal       journal.Repository
	Executor      executor.Executor
	MaxRuns       int
	LeaseDuration time.Duration
	Now           func() time.Time
	Fault         FaultHook
	MaxAttempts   int
}

func New(repository journal.Repository, worker executor.Executor) *Service {
	return &Service{
		Journal: repository, Executor: worker, MaxRuns: 10_000,
		LeaseDuration: 30 * time.Second, Now: time.Now, MaxAttempts: 5,
	}
}

func (service *Service) Health(ctx context.Context) error {
	if service == nil || service.Journal == nil {
		return errors.New("journal is unavailable")
	}
	return service.Journal.Health(ctx)
}

func (service *Service) Create(ctx context.Context, request protocol.CreateRequest) (protocol.Run, bool, error) {
	if service.Executor == nil {
		return protocol.Run{}, false, ErrExecutorUnavailable
	}
	protocol.NormalizeCreate(&request)
	if err := protocol.ValidateCreate(request); err != nil {
		return protocol.Run{}, false, fmt.Errorf("invalid create request: %w", err)
	}
	digest, err := protocol.Digest(request)
	if err != nil {
		return protocol.Run{}, false, err
	}
	now := service.Now().UTC()
	start := journal.Command{
		CommandID: startCommandID(request.RunID), RunID: request.RunID,
		Kind: "start", Digest: digest, CreateRequest: &request,
		State: journal.CommandPending, AvailableAt: now,
		Receipt: protocol.ControlReceipt{
			CommandID: startCommandID(request.RunID), RunID: request.RunID,
			Kind: "start", State: journal.CommandPending, Version: 1,
			CreatedAt: now, UpdatedAt: now,
		},
	}
	return service.Journal.Create(ctx, request, digest, start, service.MaxRuns)
}

func (service *Service) Get(ctx context.Context, runID string) (protocol.Run, error) {
	return service.Journal.Get(ctx, runID)
}

func (service *Service) Events(ctx context.Context, runID string, after uint64) ([]protocol.Event, error) {
	return service.Journal.Events(ctx, runID, after)
}

func (service *Service) Stop(ctx context.Context, runID string, request protocol.StopRequest, principalID string) (protocol.ControlReceipt, bool, error) {
	if service.Executor == nil {
		return protocol.ControlReceipt{}, false, ErrExecutorUnavailable
	}
	if err := protocol.ValidateStop(request); err != nil {
		return protocol.ControlReceipt{}, false, fmt.Errorf("invalid stop request: %w", err)
	}
	now := service.Now().UTC()
	command := journal.Command{
		CommandID: request.CommandID, RunID: runID, Kind: "stop",
		ExpectedVersion: request.ExpectedVersion, ActorID: request.ActorID,
		Reason: request.Reason, Metadata: request.Metadata, AvailableAt: now,
		Receipt: protocol.ControlReceipt{
			CommandID: request.CommandID, RunID: runID, Kind: "stop",
			State: journal.CommandPending, Version: request.ExpectedVersion,
			CreatedAt: now, UpdatedAt: now,
		},
	}
	command.Metadata = withServicePrincipal(command.Metadata, principalID)
	digest, err := protocol.Digest(map[string]any{"request": request, "service_principal": principalID})
	if err != nil {
		return protocol.ControlReceipt{}, false, err
	}
	command.Digest = digest
	return service.Journal.SubmitControl(ctx, command)
}

func (service *Service) Decide(ctx context.Context, runID, approvalID string, request protocol.ApprovalDecisionRequest, principalID string) (protocol.ControlReceipt, bool, error) {
	if service.Executor == nil {
		return protocol.ControlReceipt{}, false, ErrExecutorUnavailable
	}
	if request.ApprovalID != approvalID {
		return protocol.ControlReceipt{}, false, fmt.Errorf("invalid approval request: route and body approval_id differ")
	}
	if err := protocol.ValidateApprovalDecision(request); err != nil {
		return protocol.ControlReceipt{}, false, fmt.Errorf("invalid approval request: %w", err)
	}
	now := service.Now().UTC()
	command := journal.Command{
		CommandID: request.CommandID, RunID: runID, Kind: request.Decision,
		ExpectedVersion: request.ExpectedVersion, ApprovalID: approvalID,
		Decision: request.Decision, ActorID: request.ActorID,
		Reason: request.Reason, Metadata: request.Metadata, AvailableAt: now,
		Receipt: protocol.ControlReceipt{
			CommandID: request.CommandID, RunID: runID, Kind: request.Decision,
			State: journal.CommandPending, Version: request.ExpectedVersion,
			CreatedAt: now, UpdatedAt: now,
		},
	}
	command.Metadata = withServicePrincipal(command.Metadata, principalID)
	digest, err := protocol.Digest(map[string]any{"request": request, "service_principal": principalID})
	if err != nil {
		return protocol.ControlReceipt{}, false, err
	}
	command.Digest = digest
	return service.Journal.SubmitControl(ctx, command)
}

func (service *Service) DispatchOnce(ctx context.Context, owner string) (bool, error) {
	if service.Executor == nil {
		return false, ErrExecutorUnavailable
	}
	now := service.Now().UTC()
	lease, err := service.Journal.Claim(ctx, owner, now, service.LeaseDuration)
	if err != nil || lease == nil {
		return false, err
	}
	outcome, err := service.Executor.Apply(ctx, lease.Command)
	if err != nil {
		if lease.Command.Attempts >= service.MaxAttempts {
			failed := protocol.Error{
				Code: "executor_failed", Message: "Executor command exhausted bounded retries.", Recoverable: true,
			}
			return true, service.Journal.Fail(ctx, *lease, failed, service.Now().UTC())
		}
		retryAt := service.Now().UTC().Add(service.LeaseDuration)
		return true, service.Journal.Retry(ctx, *lease, retryAt, "Executor command failed.")
	}
	if service.Fault != nil {
		if err := service.Fault(FaultAfterExecutorEffect, lease.Command); err != nil {
			return true, err
		}
	}
	_, err = service.Journal.Complete(ctx, *lease, outcome, service.Now().UTC())
	return true, err
}

func (service *Service) Capabilities() protocol.Capabilities {
	return protocol.Capabilities{
		Healthy: true, Backend: "framework_runs", SupportedProtocols: []string{"runs_api"},
		SupportsEvents: true, SupportsCancellation: true, SupportsApprovals: true,
		SupportsUsage: true, ProductionReady: false,
		Features: []string{
			"durable_journal", "replayable_sse", "core_supplied_identity",
			"durable_control_cas", "immutable_candidate_manifests",
		},
		Metadata: map[string]any{"executor_configured": service.Executor != nil},
	}
}

func startCommandID(runID string) string {
	digest, _ := protocol.Digest(map[string]string{"run_id": runID, "kind": "start"})
	return "start:" + digest[:32]
}

func withServicePrincipal(metadata map[string]any, principalID string) map[string]any {
	cloned := protocol.Clone(metadata)
	if cloned == nil {
		cloned = map[string]any{}
	}
	cloned["service_principal"] = principalID
	return cloned
}

package journal

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/mycelis/framework-runs/internal/protocol"
)

type memoryRun struct {
	request          protocol.CreateRequest
	requestDigest    string
	run              protocol.Run
	events           []protocol.Event
	nextSequence     uint64
	pendingCommandID string
}

type MemoryRepository struct {
	mu               sync.Mutex
	runs             map[string]*memoryRun
	idempotencyToRun map[string]string
	commands         map[string]*Command
	commandOrder     []string
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		runs: map[string]*memoryRun{}, idempotencyToRun: map[string]string{},
		commands: map[string]*Command{},
	}
}

func (repository *MemoryRepository) Health(context.Context) error { return nil }

func (repository *MemoryRepository) Create(
	_ context.Context,
	request protocol.CreateRequest,
	digest string,
	start Command,
	maxRuns int,
) (protocol.Run, bool, error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if existing := repository.runs[request.RunID]; existing != nil {
		if existing.requestDigest == digest && existing.request.Correlation.IdempotencyKey == request.Correlation.IdempotencyKey {
			return protocol.Clone(existing.run), true, nil
		}
		return protocol.Run{}, false, ErrRunConflict
	}
	if runID := repository.idempotencyToRun[request.Correlation.IdempotencyKey]; runID != "" {
		return protocol.Run{}, false, ErrRunConflict
	}
	if maxRuns > 0 && len(repository.runs) >= maxRuns {
		return protocol.Run{}, false, ErrCapacity
	}
	now := start.Receipt.CreatedAt.UTC()
	run := protocol.Run{
		RunID: request.RunID, Correlation: request.Correlation,
		Status: protocol.StatusAccepted, Version: 1,
		CreatedAt: now, UpdatedAt: now,
		Metadata: map[string]any{
			"execution_authority": "mycelis_core",
			"storage":             "durable_journal",
		},
	}
	event := AcceptedEvent(run)
	repository.runs[run.RunID] = &memoryRun{
		request: protocol.Clone(request), requestDigest: digest, run: run,
		events: []protocol.Event{event}, nextSequence: 2,
		pendingCommandID: start.CommandID,
	}
	repository.idempotencyToRun[request.Correlation.IdempotencyKey] = request.RunID
	start.State = CommandPending
	repository.commands[start.CommandID] = cloneCommand(start)
	repository.commandOrder = append(repository.commandOrder, start.CommandID)
	return protocol.Clone(run), false, nil
}

func (repository *MemoryRepository) Get(_ context.Context, runID string) (protocol.Run, error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	stored := repository.runs[runID]
	if stored == nil {
		return protocol.Run{}, ErrNotFound
	}
	return protocol.Clone(stored.run), nil
}

func (repository *MemoryRepository) Events(_ context.Context, runID string, after uint64) ([]protocol.Event, error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	stored := repository.runs[runID]
	if stored == nil {
		return nil, ErrNotFound
	}
	last := stored.nextSequence - 1
	if after > last {
		return nil, ErrCursorGap
	}
	events := make([]protocol.Event, 0, last-after)
	for _, event := range stored.events {
		if event.Sequence > after {
			events = append(events, protocol.Clone(event))
		}
	}
	return events, nil
}

func (repository *MemoryRepository) SubmitControl(_ context.Context, command Command) (protocol.ControlReceipt, bool, error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if existing := repository.commands[command.CommandID]; existing != nil {
		if existing.Digest != command.Digest || existing.RunID != command.RunID || existing.Kind != command.Kind {
			return protocol.ControlReceipt{}, false, ErrCommandConflict
		}
		return protocol.Clone(existing.Receipt), true, nil
	}
	stored := repository.runs[command.RunID]
	if stored == nil {
		return protocol.ControlReceipt{}, false, ErrNotFound
	}
	if stored.run.Version != command.ExpectedVersion {
		return protocol.ControlReceipt{}, false, ErrVersionConflict
	}
	if stored.pendingCommandID != "" {
		return protocol.ControlReceipt{}, false, ErrCommandConflict
	}
	if err := validateControlAgainstRun(stored.run, command); err != nil {
		return protocol.ControlReceipt{}, false, err
	}
	command.State = CommandPending
	command.Receipt.State = CommandPending
	repository.commands[command.CommandID] = cloneCommand(command)
	repository.commandOrder = append(repository.commandOrder, command.CommandID)
	stored.pendingCommandID = command.CommandID
	return protocol.Clone(command.Receipt), false, nil
}

func (repository *MemoryRepository) Claim(_ context.Context, owner string, now time.Time, leaseDuration time.Duration) (*Lease, error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	for _, commandID := range repository.commandOrder {
		command := repository.commands[commandID]
		if command.State == CommandPending && command.AvailableAt.After(now) {
			continue
		}
		if command.State != CommandPending && !(command.State == CommandLeased && command.LeaseUntil.Before(now)) {
			continue
		}
		command.State = CommandLeased
		command.Attempts++
		command.LeaseGeneration++
		command.LeaseOwner = owner
		command.LeaseToken = leaseToken(command.CommandID, command.LeaseGeneration)
		command.LeaseUntil = now.Add(leaseDuration)
		return &Lease{
			Command: cloneCommandValue(*command), Owner: owner,
			Token: command.LeaseToken, Epoch: command.LeaseGeneration,
		}, nil
	}
	return nil, nil
}

func (repository *MemoryRepository) Complete(_ context.Context, lease Lease, outcome protocol.ExecutorOutcome, now time.Time) (protocol.Run, error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	command := repository.commands[lease.Command.CommandID]
	if command == nil || command.State != CommandLeased || command.LeaseOwner != lease.Owner || command.LeaseToken != lease.Token || command.LeaseGeneration != lease.Epoch {
		return protocol.Run{}, ErrLeaseLost
	}
	stored := repository.runs[command.RunID]
	if stored == nil || stored.pendingCommandID != command.CommandID {
		return protocol.Run{}, ErrConflict
	}
	event, err := ApplyOutcome(&stored.run, *command, outcome, now)
	if err != nil {
		return protocol.Run{}, err
	}
	event.Sequence = stored.nextSequence
	if err := protocol.ValidateEvent(event); err != nil {
		return protocol.Run{}, err
	}
	stored.nextSequence++
	stored.events = append(stored.events, protocol.Clone(event))
	stored.pendingCommandID = ""
	command.State = CommandApplied
	command.Receipt.State = CommandApplied
	command.Receipt.Version = stored.run.Version
	command.Receipt.UpdatedAt = now.UTC()
	return protocol.Clone(stored.run), nil
}

func (repository *MemoryRepository) Fail(_ context.Context, lease Lease, failure protocol.Error, now time.Time) error {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	command := repository.commands[lease.Command.CommandID]
	if command == nil || command.State != CommandLeased || command.LeaseOwner != lease.Owner ||
		command.LeaseToken != lease.Token || command.LeaseGeneration != lease.Epoch {
		return ErrLeaseLost
	}
	stored := repository.runs[command.RunID]
	if stored == nil || stored.pendingCommandID != command.CommandID {
		return ErrConflict
	}
	stored.pendingCommandID = ""
	stored.run.UpdatedAt = now.UTC()
	command.State = CommandFailed
	command.Receipt.State = CommandFailed
	command.Receipt.Error = protocol.Clone(&failure)
	command.Receipt.UpdatedAt = now.UTC()
	command.LeaseOwner, command.LeaseToken = "", ""
	command.LeaseUntil = time.Time{}
	return nil
}

func (repository *MemoryRepository) Retry(_ context.Context, lease Lease, availableAt time.Time, message string) error {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	command := repository.commands[lease.Command.CommandID]
	if command == nil || command.State != CommandLeased || command.LeaseToken != lease.Token || command.LeaseGeneration != lease.Epoch {
		return ErrLeaseLost
	}
	command.State = CommandPending
	command.AvailableAt = availableAt
	command.LeaseUntil = time.Time{}
	command.LeaseOwner = ""
	command.LeaseToken = ""
	command.Receipt.Error = &protocol.Error{Code: "executor_retry", Message: message, Recoverable: true}
	command.Receipt.UpdatedAt = availableAt.UTC()
	return nil
}

func (repository *MemoryRepository) Receipts() []protocol.ControlReceipt {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	receipts := make([]protocol.ControlReceipt, 0, len(repository.commands))
	for _, command := range repository.commands {
		if command.Kind != "start" {
			receipts = append(receipts, protocol.Clone(command.Receipt))
		}
	}
	sort.Slice(receipts, func(i, j int) bool { return receipts[i].CommandID < receipts[j].CommandID })
	return receipts
}

func validateControlAgainstRun(run protocol.Run, command Command) error {
	if run.Status == protocol.StatusCompleted || run.Status == protocol.StatusFailed || run.Status == protocol.StatusCancelled {
		return ErrInvalidRunState
	}
	if command.Kind == "stop" {
		return nil
	}
	if run.Status != protocol.StatusApprovalNeeded || run.Approval == nil {
		return ErrInvalidRunState
	}
	if run.Approval.ID != command.ApprovalID {
		return ErrApprovalNotFound
	}
	if command.Kind != "approve" && command.Kind != "deny" {
		return ErrCommandConflict
	}
	return nil
}

func leaseToken(commandID string, generation uint64) string {
	digest, _ := protocol.Digest(map[string]any{"command_id": commandID, "generation": generation, "at": time.Now().UnixNano()})
	return "lease:" + digest[:32]
}

func cloneCommand(command Command) *Command {
	copy := cloneCommandValue(command)
	return &copy
}

func cloneCommandValue(command Command) Command {
	command.Metadata = protocol.Clone(command.Metadata)
	command.CreateRequest = protocol.Clone(command.CreateRequest)
	command.Receipt = protocol.Clone(command.Receipt)
	return command
}

func (repository *MemoryRepository) String() string {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	return fmt.Sprintf("memory journal: %d runs", len(repository.runs))
}

package journal

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mycelis/framework-runs/internal/protocol"
)

var (
	ErrNotFound         = errors.New("not found")
	ErrConflict         = errors.New("conflict")
	ErrRunConflict      = fmt.Errorf("%w: run identity", ErrConflict)
	ErrVersionConflict  = fmt.Errorf("%w: run version", ErrConflict)
	ErrCommandConflict  = fmt.Errorf("%w: command identity", ErrConflict)
	ErrInvalidRunState  = fmt.Errorf("%w: invalid run state", ErrConflict)
	ErrApprovalNotFound = fmt.Errorf("%w: approval not found", ErrNotFound)
	ErrCapacity         = errors.New("capacity unavailable")
	ErrCursorGap        = errors.New("cursor gap")
	ErrLeaseLost        = errors.New("command lease lost")
	ErrSchemaMismatch   = errors.New("service schema is partial or incompatible")
)

const (
	CommandPending = "pending"
	CommandLeased  = "leased"
	CommandApplied = "applied"
	CommandFailed  = "failed"
)

type Command struct {
	CommandID       string
	RunID           string
	Kind            string
	Digest          string
	ExpectedVersion uint64
	ApprovalID      string
	Decision        string
	ActorID         string
	Reason          string
	Metadata        map[string]any
	CreateRequest   *protocol.CreateRequest
	State           string
	Attempts        int
	LeaseOwner      string
	LeaseToken      string
	LeaseGeneration uint64
	AvailableAt     time.Time
	LeaseUntil      time.Time
	Receipt         protocol.ControlReceipt
}

type Lease struct {
	Command Command
	Owner   string
	Token   string
	Epoch   uint64
}

type Repository interface {
	Health(context.Context) error
	Create(context.Context, protocol.CreateRequest, string, Command, int) (protocol.Run, bool, error)
	Get(context.Context, string) (protocol.Run, error)
	Events(context.Context, string, uint64) ([]protocol.Event, error)
	SubmitControl(context.Context, Command) (protocol.ControlReceipt, bool, error)
	Claim(context.Context, string, time.Time, time.Duration) (*Lease, error)
	Complete(context.Context, Lease, protocol.ExecutorOutcome, time.Time) (protocol.Run, error)
	Fail(context.Context, Lease, protocol.Error, time.Time) error
	Retry(context.Context, Lease, time.Time, string) error
}

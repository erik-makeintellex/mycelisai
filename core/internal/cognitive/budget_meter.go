package cognitive

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Execution kinds recorded in token_usage_ledger.execution_kind.
const (
	ExecutionKindSomaTurn       = "soma_turn"
	ExecutionKindAgentTurn      = "agent_turn"
	ExecutionKindCouncilConsult = "council_consult"
	ExecutionKindDraft          = "draft"
	ExecutionKindAgentry        = "agentry"
	ExecutionKindSystem         = "system"
)

// ErrTokenBudgetExhausted matches every *TokenBudgetExhaustedError via errors.Is.
var ErrTokenBudgetExhausted = errors.New(TokenBudgetExhaustedCode)

// TokenBudgetExhaustedError is the honest stop: the named scope has less than
// MinBudgetHeadroom tokens left, so no provider call was made.
type TokenBudgetExhaustedError struct {
	Scope    string
	Ref      string
	Used     int
	Limit    int
	ResetsAt time.Time // zero for execution and run scopes
}

func (e *TokenBudgetExhaustedError) Error() string {
	return fmt.Sprintf("%s: %s scope used %d of %d tokens", TokenBudgetExhaustedCode, e.Scope, e.Used, e.Limit)
}

func (e *TokenBudgetExhaustedError) Is(target error) bool { return target == ErrTokenBudgetExhausted }

func (e *TokenBudgetExhaustedError) Code() string { return TokenBudgetExhaustedCode }

// AsTokenBudgetExhausted unwraps a budget stop from err, if any.
func AsTokenBudgetExhausted(err error) *TokenBudgetExhaustedError {
	var stop *TokenBudgetExhaustedError
	if errors.As(err, &stop) {
		return stop
	}
	return nil
}

// ExecutionMeter is the in-memory, per-execution token account. It is created
// at unit entry (swarm message, D2 drafting pass, agentry Run) and travels in
// the InferRequest or the context. Only Router.InferWithContract charges it.
type ExecutionMeter struct {
	mu          sync.Mutex
	id          string
	kind        string
	correlation InferenceCorrelation
	used        int
	warned      bool
	stop        *TokenBudgetExhaustedError
}

func NewExecutionMeter(kind string, correlation InferenceCorrelation) *ExecutionMeter {
	if strings.TrimSpace(kind) == "" {
		kind = ExecutionKindSystem
	}
	return &ExecutionMeter{id: uuid.NewString(), kind: kind, correlation: correlation}
}

func (m *ExecutionMeter) ID() string   { return m.id }
func (m *ExecutionMeter) Kind() string { return m.kind }

func (m *ExecutionMeter) Used() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.used
}

// Stop returns the budget stop that ended this execution, if any.
func (m *ExecutionMeter) Stop() *TokenBudgetExhaustedError {
	if m == nil {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.stop
}

type executionMeterKey struct{}

func WithExecutionMeter(ctx context.Context, meter *ExecutionMeter) context.Context {
	if meter == nil {
		return ctx
	}
	return context.WithValue(ctx, executionMeterKey{}, meter)
}

func ExecutionMeterFrom(ctx context.Context) *ExecutionMeter {
	if ctx == nil {
		return nil
	}
	meter, _ := ctx.Value(executionMeterKey{}).(*ExecutionMeter)
	return meter
}

// meterFor picks the request meter, then the context meter. Without either,
// a correlated call is its own agent_turn and an uncorrelated one a system
// execution (per_execution only, excluded from team/agent day caps).
func meterFor(ctx context.Context, req InferRequest) *ExecutionMeter {
	if req.Meter != nil {
		return req.Meter
	}
	if meter := ExecutionMeterFrom(ctx); meter != nil {
		return meter
	}
	kind := ExecutionKindSystem
	if req.Correlation.TeamID != "" || req.Correlation.AgentID != "" || req.Correlation.RunID != "" {
		kind = ExecutionKindAgentTurn
	}
	return NewExecutionMeter(kind, req.Correlation)
}

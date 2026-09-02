package executor

import (
	"context"

	"github.com/mycelis/framework-runs/internal/journal"
	"github.com/mycelis/framework-runs/internal/protocol"
)

// Executor applies one durable command. Implementations must make CommandID
// idempotent and return the prior outcome when a response or commit is lost.
// A lease fences service commits; it cannot by itself make an external effect
// exactly once.
type Executor interface {
	Apply(context.Context, journal.Command) (protocol.ExecutorOutcome, error)
}

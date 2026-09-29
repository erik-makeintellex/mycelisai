package server

import (
	"context"
	"log"
	"net/http"

	"github.com/mycelis/core/internal/memory"
	"github.com/mycelis/core/internal/swarm"
	"github.com/nats-io/nats.go"
)

// Soma recall user scope (SRU). A Soma chat turn, a council chat and a
// confirmed tool call recall saved memory as the requesting user, with the
// MEM-LIST read rule (memoryReader). No identity reads org-wide entries only;
// a failed membership lookup withholds the governed lane (Unavailable) and
// never falls back to the unscoped set.

// memoryReaderFromContext applies memoryReader to a context that carries the
// request identity; the confirmed tool path has no *http.Request.
func (s *AdminServer) memoryReaderFromContext(ctx context.Context) (memory.GovernedReader, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/", nil)
	if err != nil {
		return memory.GovernedReader{}, err
	}
	return s.memoryReader(req)
}

// recallAccessFor is the saved-memory read scope of the identity in ctx.
func (s *AdminServer) recallAccessFor(ctx context.Context) swarm.RecallAccess {
	if IdentityFromContext(ctx) == nil {
		return swarm.RecallAccess{}
	}
	reader, err := s.memoryReaderFromContext(ctx)
	if err != nil {
		log.Printf("soma recall scope: memory access check failed, governed recall withheld: %v", err)
		return swarm.RecallAccess{Unavailable: true}
	}
	return swarm.RecallAccess{Reader: reader, User: true}
}

// recallTurnMsg builds the direct request to an agent. A turn with a user
// (or a failed lookup) carries a single-use, in-process token for its read
// scope; release drops the token if the agent never claimed it.
func (s *AdminServer) recallTurnMsg(ctx context.Context, subject string, payload []byte) (*nats.Msg, func()) {
	msg := nats.NewMsg(subject)
	msg.Data = payload
	access := s.recallAccessFor(ctx)
	if !access.User && !access.Unavailable {
		return msg, func() {}
	}
	token, release := swarm.RegisterRecallTurn(access)
	msg.Header.Set(swarm.RecallTurnHeader, token)
	return msg, release
}

// withConfirmedRecallAccess scopes a confirmed plan's memory tools to the
// confirming user. The async dispatch path has no request identity, so its
// memory tools read org-wide only and promote is refused.
func (s *AdminServer) withConfirmedRecallAccess(ctx context.Context) context.Context {
	inv, ok := swarm.ToolInvocationContextFromContext(ctx)
	if !ok {
		return ctx
	}
	inv.Recall = s.recallAccessFor(ctx)
	return swarm.WithToolInvocationContext(ctx, inv)
}

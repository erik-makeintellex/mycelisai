package swarm

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/mycelis/core/internal/memory"
)

// RecallAccess is the saved-memory read scope of one Soma turn or tool call
// (SRU). Core builds it from the requesting user's identity with the
// MEM-LIST rule (server memoryReader); models, tool arguments and NATS
// payloads never supply it.
//
//   - User: a verified request identity backs Reader, so governed recall
//     admits org-wide entries plus what that user may read.
//   - No user (the zero value: scheduler, team-internal agents, background
//     work, any caller without Core's turn token): org-wide entries only.
//   - Unavailable: the identity or membership lookup failed, so the governed
//     lane contributes nothing and the reply says memory was unavailable. It
//     never falls back to the unscoped set.
type RecallAccess struct {
	Reader      memory.GovernedReader
	User        bool
	Unavailable bool
}

// RecallTurnHeader carries the Core-minted turn token on a direct request to
// an agent. The token is random, single use and resolvable only inside this
// process, so a NATS publisher cannot claim another user's read scope: an
// unknown token resolves to Unavailable, a missing one to no user.
const RecallTurnHeader = "Mycelis-Recall-Turn"

// errRecallUnavailable is returned by memory tools when the turn's read
// scope could not be verified.
var errRecallUnavailable = errors.New("saved memory is unavailable for this turn: the requesting user's memory access could not be verified")

var recallTurns = struct {
	sync.Mutex
	m map[string]RecallAccess
}{m: map[string]RecallAccess{}}

// RegisterRecallTurn stores access under a fresh token for one direct
// request. release removes it if the agent never claimed it.
func RegisterRecallTurn(access RecallAccess) (string, func()) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		// No token: the agent treats the turn as unavailable, never unscoped.
		return "unavailable", func() {}
	}
	token := hex.EncodeToString(raw)
	recallTurns.Lock()
	recallTurns.m[token] = access
	recallTurns.Unlock()
	return token, func() {
		recallTurns.Lock()
		delete(recallTurns.m, token)
		recallTurns.Unlock()
	}
}

// claimRecallTurn resolves and consumes a turn token. No token means no
// user; a token this process did not mint (or already consumed) is a failed
// lookup.
func claimRecallTurn(token string) RecallAccess {
	token = strings.TrimSpace(token)
	if token == "" {
		return RecallAccess{}
	}
	recallTurns.Lock()
	access, ok := recallTurns.m[token]
	delete(recallTurns.m, token)
	recallTurns.Unlock()
	if !ok {
		return RecallAccess{Unavailable: true}
	}
	return access
}

// governedReader is the reader every governed recall of this scope applies.
// Recall never runs with a nil reader, which would be unscoped.
func (a RecallAccess) governedReader() (*memory.GovernedReader, error) {
	if a.Unavailable {
		return nil, errRecallUnavailable
	}
	if !a.User {
		return &memory.GovernedReader{}, nil
	}
	reader := a.Reader
	return &reader, nil
}

// recallAccessFromContext is the read scope of a tool call; a call without
// Core's invocation context has no user.
func recallAccessFromContext(ctx context.Context) RecallAccess {
	if inv, ok := ToolInvocationContextFromContext(ctx); ok {
		return inv.Recall
	}
	return RecallAccess{}
}

// promoteSourceRefused answers an unknown and an unreadable source alike, so
// promote is no existence oracle for other users' entries.
const promoteSourceRefused = "promote_deployment_context refused: the source entry was not found or the requesting user cannot read it. Nothing was promoted."

// requirePromotableSource lets a promote read its source only for a
// verified requesting user who may read that governed entry (SRU). Saving
// into company_knowledge stays gated by requireConfirmedOrgWideWrite.
func (r *InternalToolRegistry) requirePromotableSource(ctx context.Context, sourceID string) error {
	access := recallAccessFromContext(ctx)
	switch {
	case access.Unavailable:
		return fmt.Errorf("promote_deployment_context refused: %v. Nothing was promoted.", errRecallUnavailable)
	case !access.User:
		return errors.New("promote_deployment_context refused: promoting saved memory needs a signed-in user who can read the source entry. Nothing was promoted.")
	case r.db == nil:
		return errors.New("promote_deployment_context refused: the memory store is unavailable. Nothing was promoted.")
	}
	readable, err := memory.GovernedEntryReadable(ctx, r.db, sourceID, access.Reader)
	if err != nil {
		log.Printf("promote_deployment_context: read check failed: %v", err)
		return errors.New("promote_deployment_context refused: read access to the source entry could not be checked. Nothing was promoted.")
	}
	if !readable {
		return errors.New(promoteSourceRefused)
	}
	return nil
}

// recallUnavailableNote is appended to a reply whose governed lane was
// withheld, so the answer says memory context was unavailable.
const recallUnavailableNote = "Saved memory context was unavailable for this reply, so it may be missing details you saved."

func withRecallUnavailableNote(text string) string {
	if strings.TrimSpace(text) == "" {
		return text
	}
	return strings.TrimRight(text, "\n") + "\n\n" + recallUnavailableNote
}

package swarm

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

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
	// OrgWideWrite is the M2 authority to share memory org-wide: a verified
	// root admin with memory:write. Core sets it from the request identity
	// (MEM-LANES-2); a model's visibility argument never does.
	OrgWideWrite bool
	// atCapacity marks an Unavailable turn whose token Core refused because
	// the outstanding-turn cap was reached (claimRecallTurn sets it).
	atCapacity bool
}

// errRecallUnavailable is returned by memory tools when the turn's read
// scope could not be verified.
var errRecallUnavailable = errors.New("saved memory is unavailable for this turn: the requesting user's memory access could not be verified")

// errRecallAtCapacity is returned instead when Core refused the turn token
// for capacity (MEM-LANES-2), so the refusal names the real cause.
var errRecallAtCapacity = errors.New("saved memory is unavailable for this turn: Core is at capacity for concurrent memory turns; try again in a moment")

// unavailableErr is the honest reason an Unavailable turn has no memory.
func (a RecallAccess) unavailableErr() error {
	if a.atCapacity {
		return errRecallAtCapacity
	}
	return errRecallUnavailable
}

// errNoOwnerID refuses a memory write for a signed-in identity without a user
// id: it could neither be attributed nor read back.
var errNoOwnerID = errors.New("saved memory needs a signed-in user id for this turn")

// ownerUserID is whom a memory write of this scope belongs to (MEM-LANES):
// the verified requesting user, or "" for a turn with no user. A write whose
// user could not be verified is refused, never saved unowned.
func (a RecallAccess) ownerUserID() (string, error) {
	if a.Unavailable {
		return "", a.unavailableErr()
	}
	if !a.User {
		return "", nil
	}
	if owner := strings.TrimSpace(a.Reader.UserID); owner != "" {
		return owner, nil
	}
	return "", errNoOwnerID
}

// laneReader is the reader of the owner lanes (temp channels, conversation
// summaries): the requesting user, or the empty reader for no user. It
// refuses exactly when a write of the same scope would be refused.
func (a RecallAccess) laneReader() (memory.GovernedReader, error) {
	if _, err := a.ownerUserID(); err != nil {
		return memory.GovernedReader{}, err
	}
	if !a.User {
		return memory.GovernedReader{}, nil
	}
	return a.Reader, nil
}

// governedReader is the reader every governed recall of this scope applies.
// Recall never runs with a nil reader, which would be unscoped.
func (a RecallAccess) governedReader() (*memory.GovernedReader, error) {
	if a.Unavailable {
		return nil, a.unavailableErr()
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
		return fmt.Errorf("promote_deployment_context refused: %v. Nothing was promoted.", access.unavailableErr())
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

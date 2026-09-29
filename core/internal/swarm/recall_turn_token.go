package swarm

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"sync"
	"time"
)

// RecallTurnHeader carries the Core-minted turn token on a direct request to
// an agent (SRU). NATS is unauthenticated, so a subscriber can read the
// token; it is therefore bound (MEM-LANES) to:
//
//   - the agent subject Core sent it to, and the reply inbox Core waits on, so
//     a token spent on another agent, or from another inbox, resolves to
//     Unavailable (and is not consumed, so Core's own request still works);
//   - a short lifetime (the chat request timeout, at most recallTurnMaxTTL);
//   - a single use.
//
// At most recallTurnCap tokens are outstanding. An unknown, expired or
// mismatched token resolves to Unavailable; a missing one to no user.
const RecallTurnHeader = "Mycelis-Recall-Turn"

const (
	recallTurnMaxTTL = 5 * time.Minute
	recallTurnCap    = 256
	// recallTurnRefused is sent when no token could be registered; it never
	// resolves, so the agent treats the turn as unavailable, never unscoped.
	recallTurnRefused = "unavailable"
)

type recallTurn struct {
	access         RecallAccess
	subject, reply string
	expires        time.Time
}

var recallTurns = struct {
	sync.Mutex
	m map[string]recallTurn
}{m: map[string]recallTurn{}}

// RegisterRecallTurn stores access for one direct request published to
// subject whose reply Core awaits on reply, valid for ttl (capped at
// recallTurnMaxTTL). release removes it if the agent never claimed it. An
// unbound request, a full registry or a failed random read gets a token
// that never resolves.
func RegisterRecallTurn(access RecallAccess, subject, reply string, ttl time.Duration) (string, func()) {
	subject, reply = strings.TrimSpace(subject), strings.TrimSpace(reply)
	if subject == "" || reply == "" {
		return recallTurnRefused, func() {}
	}
	if ttl <= 0 || ttl > recallTurnMaxTTL {
		ttl = recallTurnMaxTTL
	}
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return recallTurnRefused, func() {}
	}
	token := hex.EncodeToString(raw)
	now := time.Now()
	recallTurns.Lock()
	for key, turn := range recallTurns.m {
		if !now.Before(turn.expires) {
			delete(recallTurns.m, key)
		}
	}
	if len(recallTurns.m) >= recallTurnCap {
		recallTurns.Unlock()
		return recallTurnRefused, func() {}
	}
	recallTurns.m[token] = recallTurn{access: access, subject: subject, reply: reply, expires: now.Add(ttl)}
	recallTurns.Unlock()
	return token, func() {
		recallTurns.Lock()
		delete(recallTurns.m, token)
		recallTurns.Unlock()
	}
}

// claimRecallTurn resolves a turn token for a request that arrived on subject
// with reply inbox reply. It is consumed only by a matching, unexpired claim.
func claimRecallTurn(token, subject, reply string) RecallAccess {
	token = strings.TrimSpace(token)
	if token == "" {
		return RecallAccess{}
	}
	recallTurns.Lock()
	defer recallTurns.Unlock()
	turn, ok := recallTurns.m[token]
	if !ok {
		return RecallAccess{Unavailable: true}
	}
	if !time.Now().Before(turn.expires) {
		delete(recallTurns.m, token)
		return RecallAccess{Unavailable: true}
	}
	if turn.subject != strings.TrimSpace(subject) || turn.reply != strings.TrimSpace(reply) {
		return RecallAccess{Unavailable: true}
	}
	delete(recallTurns.m, token)
	return turn.access
}

package swarm

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
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
// At most recallTurnCap tokens are outstanding, and at most recallTurnUserCap
// for one user (MEM-LANES-2), so one user cannot hold them all. An unknown,
// expired or mismatched token resolves to Unavailable; a missing one to no
// user.
const RecallTurnHeader = "Mycelis-Recall-Turn"

const (
	recallTurnMaxTTL  = 5 * time.Minute
	recallTurnCap     = 256
	recallTurnUserCap = 32
	// recallTurnRefused is sent when no token could be registered; it never
	// resolves, so the agent treats the turn as unavailable, never unscoped.
	recallTurnRefused = "unavailable"
	// recallTurnAtCapacity is sent when a cap refused the token. It resolves
	// to Unavailable too, with the capacity reason (it grants nothing, so a
	// forged copy only changes the wording of a refusal).
	recallTurnAtCapacity = "at-capacity"
	// recallAtCapacityNote follows the unavailable note on a reply whose turn
	// was refused for capacity.
	recallAtCapacityNote = "Core was at capacity for concurrent memory turns; try again in a moment."
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

// recallTurnHolder is whose share of the per-user cap a token counts
// against: the verified user, or one shared bucket for unverified turns.
func recallTurnHolder(access RecallAccess) string {
	if access.User && !access.Unavailable {
		return "user:" + strings.TrimSpace(access.Reader.UserID)
	}
	return "unverified"
}

// RegisterRecallTurn stores access for one direct request published to
// subject whose reply Core awaits on reply, valid for ttl (capped at
// recallTurnMaxTTL). release removes it if the agent never claimed it. An
// unbound request, a full registry, a user at their cap or a failed random
// read gets a token that never resolves; every refusal is logged.
func RegisterRecallTurn(access RecallAccess, subject, reply string, ttl time.Duration) (string, func()) {
	subject, reply = strings.TrimSpace(subject), strings.TrimSpace(reply)
	if subject == "" || reply == "" {
		log.Printf("recall turn refused: the request to %q has no reply inbox to bind", subject)
		return recallTurnRefused, func() {}
	}
	if ttl <= 0 || ttl > recallTurnMaxTTL {
		ttl = recallTurnMaxTTL
	}
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		log.Printf("recall turn refused: no token could be generated for %q: %v", subject, err)
		return recallTurnRefused, func() {}
	}
	token := hex.EncodeToString(raw)
	now := time.Now()
	holder := recallTurnHolder(access)
	recallTurns.Lock()
	held := 0
	for key, turn := range recallTurns.m {
		if !now.Before(turn.expires) {
			delete(recallTurns.m, key)
		} else if recallTurnHolder(turn.access) == holder {
			held++
		}
	}
	total := len(recallTurns.m)
	if total >= recallTurnCap || held >= recallTurnUserCap {
		recallTurns.Unlock()
		log.Printf("recall turn refused: at capacity for %q (%s holds %d of %d, %d of %d outstanding in all)",
			subject, holder, held, recallTurnUserCap, total, recallTurnCap)
		return recallTurnAtCapacity, func() {}
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
	if token == recallTurnAtCapacity {
		return RecallAccess{Unavailable: true, atCapacity: true}
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

// requestWithRecallTurn is a request-reply to an agent subject that carries
// the calling turn's access (MEM-LANES-2): a turn with a user, or one whose
// user could not be verified, registers a token bound to subject and a fresh
// reply inbox, exactly as Core's chat path does, so the agent's reads and
// writes stay that user's. A turn with no user sends an ordinary request.
func requestWithRecallTurn(ctx context.Context, nc *nats.Conn, access RecallAccess, subject string, data []byte, ttl time.Duration) (*nats.Msg, error) {
	if !access.User && !access.Unavailable {
		return nc.RequestWithContext(ctx, subject, data)
	}
	msg := nats.NewMsg(subject)
	msg.Data = data
	msg.Reply = nats.NewInbox()
	token, release := RegisterRecallTurn(access, subject, msg.Reply, ttl)
	defer release()
	msg.Header.Set(RecallTurnHeader, token)
	sub, err := nc.SubscribeSync(msg.Reply)
	if err != nil {
		return nil, err
	}
	defer sub.Unsubscribe() //nolint:errcheck
	if err := nc.PublishMsg(msg); err != nil {
		return nil, err
	}
	reply, err := sub.NextMsgWithContext(ctx)
	if err != nil {
		return nil, err
	}
	if len(reply.Data) == 0 && reply.Header.Get("Status") == "503" {
		return nil, nats.ErrNoResponders
	}
	return reply, nil
}

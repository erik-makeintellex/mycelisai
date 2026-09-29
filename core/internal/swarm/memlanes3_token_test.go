package swarm

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/memory"
)

// Item 6: eight users at the per-user cap filled the shared cap and refused
// everyone else. A user's first outstanding turn draws from a reserve the
// others cannot fill.
func TestMemLanes3Token_EightUsersCannotStarveAFirstTurn(t *testing.T) {
	var releases []func()
	t.Cleanup(func() {
		for _, release := range releases {
			release()
		}
	})
	register := func(uid string) (string, string, string) {
		subject, reply := "swarm.council.memlanes3.request", "_INBOX.memlanes3."+uuid.NewString()
		token, release := RegisterRecallTurn(RecallAccess{User: true, Reader: memory.GovernedReader{UserID: uid}}, subject, reply, time.Minute)
		releases = append(releases, release)
		return token, subject, reply
	}
	for u := 0; u < 8; u++ {
		for i := 0; i < recallTurnCap; i++ {
			if token, _, _ := register(fmt.Sprintf("memlanes3-sybil-%d", u)); token == recallTurnAtCapacity {
				break
			}
		}
	}
	for _, who := range []string{"memlanes3-carol", "memlanes3-dave"} {
		token, subject, reply := register(who)
		if got := claimRecallTurn(token, subject, reply); !got.User || got.Reader.UserID != who {
			t.Errorf("%s's first turn was refused after eight users filled their caps: %q", who, token)
		}
	}
	recallTurns.Lock()
	outstanding := len(recallTurns.m)
	recallTurns.Unlock()
	if outstanding > recallTurnCap {
		t.Errorf("%d tokens outstanding, over the cap of %d", outstanding, recallTurnCap)
	}
}

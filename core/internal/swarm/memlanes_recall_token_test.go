package swarm

import (
	"testing"
	"time"
)

// MEM-LANES token binding: a turn token is spendable once, only on the
// subject and reply inbox it was issued for, and only before it expires.

func TestMemLanesRecallToken_BoundToSubjectAndReply(t *testing.T) {
	access := RecallAccess{User: true}
	token, release := RegisterRecallTurn(access, "swarm.council.admin.request", "_INBOX.core", time.Minute)
	defer release()
	for name, got := range map[string]RecallAccess{
		"another agent":  claimRecallTurn(token, "swarm.council.council-x.request", "_INBOX.core"),
		"another inbox":  claimRecallTurn(token, "swarm.council.admin.request", "_INBOX.attacker"),
		"no reply inbox": claimRecallTurn(token, "swarm.council.admin.request", ""),
	} {
		if !got.Unavailable || got.User {
			t.Errorf("%s: a mismatched claim must fail closed: %+v", name, got)
		}
	}
	if got := claimRecallTurn(token, "swarm.council.admin.request", "_INBOX.core"); !got.User || got.Unavailable {
		t.Fatalf("mismatched claims must not burn the token: %+v", got)
	}
	if got := claimRecallTurn(token, "swarm.council.admin.request", "_INBOX.core"); !got.Unavailable {
		t.Fatalf("a token is single use: %+v", got)
	}
}

func TestMemLanesRecallToken_ExpiredTokenIsRefused(t *testing.T) {
	token, release := RegisterRecallTurn(RecallAccess{User: true}, "swarm.council.admin.request", "_INBOX.core", 20*time.Millisecond)
	defer release()
	time.Sleep(40 * time.Millisecond)
	if got := claimRecallTurn(token, "swarm.council.admin.request", "_INBOX.core"); !got.Unavailable || got.User {
		t.Fatalf("an expired token must be refused: %+v", got)
	}
}

func TestMemLanesRecallToken_UnboundAndOverCapRegistrationsNeverResolve(t *testing.T) {
	if token, _ := RegisterRecallTurn(RecallAccess{User: true}, "swarm.council.admin.request", "", time.Minute); !claimRecallTurn(token, "swarm.council.admin.request", "").Unavailable {
		t.Fatal("a token without a reply inbox must never resolve")
	}
	releases := []func(){}
	defer func() {
		for _, release := range releases {
			release()
		}
	}()
	for i := 0; i < recallTurnCap; i++ {
		_, release := RegisterRecallTurn(RecallAccess{User: true}, "swarm.council.admin.request", "_INBOX.fill", time.Minute)
		releases = append(releases, release)
	}
	token, release := RegisterRecallTurn(RecallAccess{User: true}, "swarm.council.admin.request", "_INBOX.over", time.Minute)
	releases = append(releases, release)
	if got := claimRecallTurn(token, "swarm.council.admin.request", "_INBOX.over"); !got.Unavailable || got.User {
		t.Fatalf("a registration over the cap must fail closed: %+v", got)
	}
}

package swarm

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/memory"
	"github.com/mycelis/core/pkg/protocol"
)

// LOW (MEMLANES-QA): one user could hold every outstanding recall turn token,
// and the refusal said "could not be verified" with nothing logged. A user's
// outstanding tokens are capped below the shared cap, a refusal names
// capacity, and each refusal is logged.
func TestMemLanes2Token_OneUserCanFillTheCap(t *testing.T) {
	var logs bytes.Buffer
	log.SetOutput(&logs)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })
	var releases []func()
	t.Cleanup(func() {
		for _, release := range releases {
			release()
		}
	})
	subject := func(who string, i int) (string, string) {
		return fmt.Sprintf("swarm.council.memlanes2-%s-%d.request", who, i), fmt.Sprintf("_INBOX.memlanes2.%s.%d", who, i)
	}
	a := RecallAccess{User: true, Reader: memory.GovernedReader{UserID: "memlanes2-user-a"}}
	var refused string
	var refusedSubject, refusedReply string
	for i := 0; i < recallTurnCap; i++ {
		s, r := subject("a", i)
		token, release := RegisterRecallTurn(a, s, r, time.Minute)
		releases = append(releases, release)
		recallTurns.Lock()
		_, live := recallTurns.m[token]
		recallTurns.Unlock()
		if !live {
			refused, refusedSubject, refusedReply = token, s, r
			break
		}
	}
	if refused == "" {
		t.Fatal("one user was allowed to hold every outstanding recall turn token")
	}

	b := RecallAccess{User: true, Reader: memory.GovernedReader{UserID: "memlanes2-user-b"}}
	s, r := subject("b", 0)
	token, release := RegisterRecallTurn(b, s, r, time.Minute)
	releases = append(releases, release)
	if got := claimRecallTurn(token, s, r); !got.User || got.Reader.UserID != "memlanes2-user-b" {
		t.Errorf("another user's turn was refused while one user held tokens: %+v", got)
	}

	got := claimRecallTurn(refused, refusedSubject, refusedReply)
	if !got.Unavailable {
		t.Fatalf("a refused token must resolve to unavailable: %+v", got)
	}
	if _, err := got.ownerUserID(); err == nil || !strings.Contains(err.Error(), "capacity") {
		t.Errorf("the refusal must name capacity: %v", err)
	}
	if !strings.Contains(logs.String(), "recall turn refused") {
		t.Errorf("the refusal must be logged: %q", logs.String())
	}
}

// memlanes2RecallExecutor records the recall scope each tool call ran under.
type memlanes2RecallExecutor struct {
	mu   sync.Mutex
	seen []RecallAccess
}

func (e *memlanes2RecallExecutor) FindToolByName(_ context.Context, name string) (uuid.UUID, string, error) {
	return InternalServerID, name, nil
}

func (e *memlanes2RecallExecutor) CallTool(ctx context.Context, _ uuid.UUID, _ string, _ map[string]any) (string, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.seen = append(e.seen, recallAccessFromContext(ctx))
	return "preflight ok", nil
}

// C2: the Core-owned council preflight inside a user's turn consults the
// council as that user, not as a no-user caller.
func TestMemLanes2_CouncilPreflightCarriesTheTurnUser(t *testing.T) {
	exec := &memlanes2RecallExecutor{}
	agent := NewAgent(context.Background(), protocol.AgentManifest{ID: "admin", Role: "admin"}, "admin-core", nil, nil, exec)
	a := RecallAccess{User: true, Reader: memory.GovernedReader{UserID: "memlanes2-preflight-user"}}
	if _, err := agent.runCouncilPreflight(a, "council-architect", "plan the launch", &toolCallPayload{Name: "write_file"}); err != nil {
		t.Fatalf("preflight: %v", err)
	}
	if len(exec.seen) != 1 || !exec.seen[0].User || exec.seen[0].Reader.UserID != "memlanes2-preflight-user" {
		t.Fatalf("the preflight consult lost the turn's user: %+v", exec.seen)
	}
}

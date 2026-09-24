package invocation

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"testing"

	_ "github.com/lib/pq"
)

const crashPhaseEnv = "MYCELIS_INVOCATION_CRASH_PHASE"
const crashIDEnv = "MYCELIS_INVOCATION_CRASH_ID"

// This child uses the real store and adapter, then exits without deferred
// cleanup or a receipt update. Only the test process itself is terminated.
func TestPGCrashProcessChild(t *testing.T) {
	phase := os.Getenv(crashPhaseEnv)
	if phase == "" {
		return
	}
	id := os.Getenv(crashIDEnv)
	db, err := sql.Open("postgres", os.Getenv("MYCELIS_INVOCATION_TEST_DSN"))
	if err != nil {
		crashChildFail(err)
	}
	store := NewStore(db)
	owner, generation, err := store.Claim(context.Background(), id)
	if err != nil {
		crashChildFail(fmt.Errorf("claim: %w", err))
	}
	if phase == "before_call" {
		os.Exit(23)
	}
	if phase != "after_call" {
		crashChildFail(fmt.Errorf("unknown crash phase %q", phase))
	}
	if err := store.Start(context.Background(), id, owner, generation); err != nil {
		crashChildFail(fmt.Errorf("start: %w", err))
	}
	var bindingJSON, inputJSON []byte
	err = db.QueryRowContext(context.Background(), `
		SELECT g.binding_snapshot::text,g.input_snapshot::text
		FROM execution_invocations i JOIN execution_effect_grants g ON g.id=i.grant_id
		WHERE i.id=$1`, id).Scan(&bindingJSON, &inputJSON)
	if err != nil {
		crashChildFail(fmt.Errorf("read immutable pin: %w", err))
	}
	var b binding
	var input Input
	if err := json.Unmarshal(bindingJSON, &b); err != nil {
		crashChildFail(err)
	}
	if err := json.Unmarshal(inputJSON, &input); err != nil {
		crashChildFail(err)
	}
	evidence, err := invokeCounting(context.Background(), b, id, input)
	if err != nil || evidence.ObservedCount != 1 {
		crashChildFail(fmt.Errorf("effect call: evidence=%+v err=%v", evidence, err))
	}
	os.Exit(23)
}

func crashChildFail(err error) {
	_, _ = fmt.Fprintln(os.Stderr, err)
	os.Exit(24)
}

func TestPGWorkerProcessDeathPreservesEffectBoundary(t *testing.T) {
	for _, phase := range []string{"before_call", "after_call"} {
		t.Run(phase, func(t *testing.T) {
			f := newPGFixture(t)
			grant := f.grant(t, 1)
			inv := f.admit(t, grant, "process-"+phase)
			executable, err := os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			cmd := exec.CommandContext(t.Context(), executable, "-test.run=^TestPGCrashProcessChild$")
			cmd.Env = append(os.Environ(), crashPhaseEnv+"="+phase, crashIDEnv+"="+inv.ID)
			output, err := cmd.CombinedOutput()
			var exit *exec.ExitError
			if !errors.As(err, &exit) || exit.ExitCode() != 23 {
				t.Fatalf("child exit=%v output=%s", err, output)
			}
			count := f.counter.count(inv.ID)
			if phase == "before_call" && count != 0 || phase == "after_call" && count != 1 {
				t.Fatalf("phase=%s effect count=%d", phase, count)
			}
			f.setState(t, `UPDATE execution_invocations SET lease_until=clock_timestamp()-INTERVAL '1 second' WHERE id=$1`, inv.ID)
			restarted := NewStore(f.db)
			if phase == "before_call" {
				got, err := restarted.Execute(t.Context(), inv.ID)
				if err != nil || got.State != "verified" || f.counter.count(inv.ID) != 1 {
					t.Fatalf("pre-call recovery state=%q count=%d err=%v", got.State, f.counter.count(inv.ID), err)
				}
				return
			}
			changed, err := restarted.RecoverExpired(t.Context())
			if err != nil || changed < 1 {
				t.Fatalf("post-call recovery changed=%d err=%v", changed, err)
			}
			got, err := restarted.Execute(t.Context(), inv.ID)
			if err != nil || got.State != "unknown_effect" || f.counter.count(inv.ID) != 1 {
				t.Fatalf("post-call recovery state=%q count=%d err=%v", got.State, f.counter.count(inv.ID), err)
			}
		})
	}
}

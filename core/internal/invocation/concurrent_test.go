package invocation

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func TestPGDuplicateExecuteWorkersCallOnce(t *testing.T) {
	f := newPGFixture(t)
	g := f.grant(t, 1)
	inv := f.admit(t, g, "workers-once")
	start := make(chan struct{})
	var wg sync.WaitGroup
	results := make([]Invocation, 2)
	errorsSeen := make([]error, 2)
	for i := range results {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			<-start
			results[index], errorsSeen[index] = f.store.Execute(t.Context(), inv.ID)
		}(i)
	}
	close(start)
	wg.Wait()
	for i, err := range errorsSeen {
		if err != nil {
			t.Fatalf("worker %d: %v", i, err)
		}
	}
	if f.counter.count(inv.ID) != 1 {
		t.Fatalf("duplicate workers called %d times", f.counter.count(inv.ID))
	}
	final, err := f.store.Get(t.Context(), f.userID, inv.ID)
	if err != nil || final.State != "verified" {
		t.Fatalf("final state=%q err=%v", final.State, err)
	}
}

func TestPGRevokeRacesExecutionStartAtBoundary(t *testing.T) {
	f := newPGFixture(t)
	g := f.grant(t, 1)
	inv := f.admit(t, g, "race-revoke")
	owner, generation, err := f.store.Claim(t.Context(), inv.ID)
	if err != nil {
		t.Fatal(err)
	}
	start := make(chan struct{})
	var wg sync.WaitGroup
	var startErr, revokeErr error
	wg.Add(2)
	go func() { defer wg.Done(); <-start; startErr = f.store.Start(t.Context(), inv.ID, owner, generation) }()
	go func() { defer wg.Done(); <-start; revokeErr = f.store.Revoke(t.Context(), f.userID, g.ID, "race") }()
	close(start)
	wg.Wait()
	if revokeErr != nil {
		t.Fatalf("revocation: %v", revokeErr)
	}
	got, err := f.store.Get(t.Context(), f.userID, inv.ID)
	if err != nil {
		t.Fatal(err)
	}
	if startErr == nil {
		if got.State != "executing" {
			t.Fatalf("start won but state=%q", got.State)
		}
	} else if errors.Is(startErr, ErrDenied) {
		if got.State != "failed_before_effect" {
			t.Fatalf("revoke won but state=%q", got.State)
		}
	} else {
		t.Fatalf("unexpected start error: %v", startErr)
	}
	if f.counter.count(inv.ID) != 0 {
		t.Fatal("authority race itself caused effect")
	}
}

func TestPGRevocationAfterExecutionStartPreservesInFlightState(t *testing.T) {
	f := newPGFixture(t)
	g := f.grant(t, 1)
	inv := f.admit(t, g, "start-wins")
	owner, generation, err := f.store.Claim(t.Context(), inv.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.store.Start(t.Context(), inv.ID, owner, generation); err != nil {
		t.Fatal(err)
	}
	if err := f.store.Revoke(t.Context(), f.userID, g.ID, "after start"); err != nil {
		t.Fatal(err)
	}
	got, err := f.store.Get(t.Context(), f.userID, inv.ID)
	if err != nil || got.State != "executing" {
		t.Fatalf("revocation rewrote in-flight receipt: %q %v", got.State, err)
	}
	if f.counter.count(inv.ID) != 0 {
		t.Fatal("start or revocation caused effect")
	}
}

func TestPGInvocationIdentitySnapshotCannotRetarget(t *testing.T) {
	f := newPGFixture(t)
	g := f.grant(t, 1)
	inv := f.admit(t, g, "immutable-snapshot")
	other := f.grant(t, 1)
	for _, query := range []string{
		`UPDATE execution_invocations SET binding_snapshot='{"endpoint":"http://evil.invalid"}'::jsonb WHERE id=$1`,
		`UPDATE execution_invocations SET input_snapshot='{"counter":"changed"}'::jsonb WHERE id=$1`,
		`UPDATE execution_invocations SET grant_id=$2 WHERE id=$1`,
	} {
		var err error
		if query == `UPDATE execution_invocations SET grant_id=$2 WHERE id=$1` {
			_, err = f.db.ExecContext(t.Context(), query, inv.ID, other.ID)
		} else {
			_, err = f.db.ExecContext(t.Context(), query, inv.ID)
		}
		if err == nil {
			t.Fatalf("identity mutation succeeded: %s", query)
		}
	}
}

func TestPGLeaseExpiresWhileInvocationLockHeld(t *testing.T) {
	for _, finish := range []bool{false, true} {
		name := "start"
		if finish {
			name = "finish"
		}
		t.Run(name, func(t *testing.T) {
			f := newPGFixture(t)
			g := f.grant(t, 1)
			inv := f.admit(t, g, "blocked-"+name)
			owner, generation, err := f.store.Claim(t.Context(), inv.ID)
			if err != nil {
				t.Fatal(err)
			}
			if finish {
				if err := f.store.Start(t.Context(), inv.ID, owner, generation); err != nil {
					t.Fatal(err)
				}
			}
			f.setState(t, `UPDATE execution_invocations SET lease_until=clock_timestamp()+INTERVAL '100 milliseconds' WHERE id=$1`, inv.ID)
			tx, err := f.db.BeginTx(t.Context(), nil)
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback()
			var held string
			if err := tx.QueryRowContext(t.Context(), `SELECT id::text FROM execution_invocations WHERE id=$1 FOR UPDATE`, inv.ID).Scan(&held); err != nil {
				t.Fatal(err)
			}
			done := make(chan error, 1)
			go func() {
				if finish {
					done <- f.store.finish(t.Context(), inv.ID, owner, generation, "verified", map[string]any{"observed_count": 1})
				} else {
					done <- f.store.Start(t.Context(), inv.ID, owner, generation)
				}
			}()
			time.Sleep(250 * time.Millisecond)
			select {
			case err := <-done:
				t.Fatalf("operation crossed held invocation lock: %v", err)
			default:
			}
			if err := tx.Commit(); err != nil {
				t.Fatal(err)
			}
			err = <-done
			if finish {
				if err != nil {
					t.Fatal(err)
				}
				got, err := f.store.Get(t.Context(), f.userID, inv.ID)
				if err != nil || got.State != "unknown_effect" {
					t.Fatalf("late finish state=%q err=%v", got.State, err)
				}
			} else if !errors.Is(err, ErrConflict) {
				t.Fatalf("expired start: %v", err)
			}
			if count := f.counter.count(inv.ID); count != 0 {
				t.Fatalf("lock-delayed path called adapter %d times", count)
			}
		})
	}
}

func TestPGMembershipExpiresWhileInvocationLockHeld(t *testing.T) {
	f := newPGFixture(t)
	f.setState(t, `UPDATE org_memberships SET expires_at=clock_timestamp()+INTERVAL '3 seconds' WHERE id=$1`, f.membershipID)
	g := f.grant(t, 1)
	inv := f.admit(t, g, "membership-lock-delay")
	owner, generation, err := f.store.Claim(t.Context(), inv.ID)
	if err != nil {
		t.Fatal(err)
	}
	tx, err := f.db.BeginTx(t.Context(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	var held string
	if err := tx.QueryRowContext(t.Context(), `SELECT id::text FROM execution_invocations WHERE id=$1 FOR UPDATE`, inv.ID).Scan(&held); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- f.store.Start(t.Context(), inv.ID, owner, generation) }()
	blocked := false
	for range 100 {
		var waiting int
		err := f.db.QueryRowContext(t.Context(), `
			SELECT count(*) FROM pg_stat_activity
			WHERE state='active' AND wait_event_type='Lock'
			AND query LIKE '%FROM execution_invocations WHERE id=$1 FOR UPDATE%'`).Scan(&waiting)
		if err != nil {
			t.Fatal(err)
		}
		if waiting > 0 {
			blocked = true
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if !blocked {
		t.Fatal("Start did not reach held invocation row before membership expiry")
	}
	for range 180 {
		var expired bool
		if err := f.db.QueryRowContext(t.Context(), `SELECT expires_at<=clock_timestamp() FROM org_memberships WHERE id=$1`, f.membershipID).Scan(&expired); err != nil {
			t.Fatal(err)
		}
		if expired {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	if err := <-done; !errors.Is(err, ErrDenied) {
		t.Fatalf("expired membership admitted at Start: %v", err)
	}
	got, err := f.store.Get(t.Context(), f.userID, inv.ID)
	if err != nil || got.State != "failed_before_effect" || f.counter.count(inv.ID) != 0 {
		t.Fatalf("state=%q count=%d err=%v", got.State, f.counter.count(inv.ID), err)
	}
}

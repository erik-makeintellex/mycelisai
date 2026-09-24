package invocation

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestPGAdmitDuplicateAndBudgetRace(t *testing.T) {
	f := newPGFixture(t)
	g := f.grant(t, 1)
	var wg sync.WaitGroup
	var mu sync.Mutex
	success, denied := 0, 0
	for _, key := range []string{"budget-a", "budget-b"} {
		wg.Add(1)
		go func(key string) {
			defer wg.Done()
			_, reused, err := f.store.Admit(t.Context(), f.userID, g.ID, g.Digest, key, Input{Counter: "ledger"})
			mu.Lock()
			defer mu.Unlock()
			if err == nil && !reused {
				success++
			} else if errors.Is(err, ErrDenied) {
				denied++
			} else {
				t.Errorf("admit race key=%s reused=%v err=%v", key, reused, err)
			}
		}(key)
	}
	wg.Wait()
	if success != 1 || denied != 1 {
		t.Fatalf("budget race success=%d denied=%d", success, denied)
	}
	var key string
	if err := f.db.QueryRowContext(t.Context(), `SELECT idempotency_key FROM execution_invocations WHERE grant_id=$1`, g.ID).Scan(&key); err != nil {
		t.Fatal(err)
	}
	first, reused, err := f.store.Admit(t.Context(), f.userID, g.ID, g.Digest, key, Input{Counter: "ledger"})
	if err != nil || !reused {
		t.Fatalf("exact duplicate: %#v %v %v", first, reused, err)
	}
	if _, _, err := f.store.Admit(t.Context(), f.userID, g.ID, g.Digest, key, Input{Counter: "changed"}); !errors.Is(err, ErrConflict) {
		t.Fatalf("changed input duplicate: %v", err)
	}
	var reserved int
	if err := f.db.QueryRowContext(t.Context(), `SELECT reserved_units FROM execution_effect_grant_state WHERE grant_id=$1`, g.ID).Scan(&reserved); err != nil || reserved != 1 {
		t.Fatalf("reserved=%d err=%v", reserved, err)
	}
}

func TestPGAdmitRevocationAndPinDrift(t *testing.T) {
	f := newPGFixture(t)
	g := f.grant(t, 2)
	if _, _, err := f.store.Admit(t.Context(), uuid.NewString(), g.ID, g.Digest, "forged", Input{Counter: "ledger"}); !errors.Is(err, ErrDenied) {
		t.Fatalf("forged subject: %v", err)
	}
	if _, _, err := f.store.Admit(t.Context(), f.userID, g.ID, "wrong-digest", "wrong", Input{Counter: "ledger"}); !errors.Is(err, ErrDenied) {
		t.Fatalf("wrong digest: %v", err)
	}
	f.setBinding(t, f.counter.server.URL+"/retargeted")
	if _, _, err := f.store.Admit(t.Context(), f.userID, g.ID, g.Digest, "drift", Input{Counter: "ledger"}); !errors.Is(err, ErrDenied) {
		t.Fatalf("binding drift: %v", err)
	}
	f.setBinding(t, f.counter.server.URL+"/increment")
	if err := f.store.Revoke(t.Context(), f.userID, g.ID, "operator stopped grant"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := f.store.Admit(t.Context(), f.userID, g.ID, g.Digest, "revoked", Input{Counter: "ledger"}); !errors.Is(err, ErrDenied) {
		t.Fatalf("revoked grant: %v", err)
	}
}

func TestPGStartRechecksAuthorityAndBindingBeforeEffect(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(*testing.T, *pgFixture, Grant)
	}{
		{"grant revoked", func(t *testing.T, f *pgFixture, g Grant) {
			if err := f.store.Revoke(t.Context(), f.userID, g.ID, "stop"); err != nil {
				t.Fatal(err)
			}
		}},
		{"membership disabled", func(t *testing.T, f *pgFixture, _ Grant) {
			f.setState(t, `UPDATE org_memberships SET status='disabled' WHERE id=$1`, f.membershipID)
		}},
		{"binding changed", func(t *testing.T, f *pgFixture, _ Grant) { f.setBinding(t, f.counter.server.URL+"/changed") }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newPGFixture(t)
			g := f.grant(t, 1)
			inv := f.admit(t, g, "start-check")
			owner, generation, err := f.store.Claim(t.Context(), inv.ID)
			if err != nil {
				t.Fatal(err)
			}
			tc.change(t, f, g)
			if err := f.store.Start(t.Context(), inv.ID, owner, generation); !errors.Is(err, ErrDenied) {
				t.Fatalf("start accepted changed authority: %v", err)
			}
			got, err := f.store.Get(t.Context(), f.userID, inv.ID)
			if err != nil || got.State != "failed_before_effect" || f.counter.count(inv.ID) != 0 {
				t.Fatalf("state=%q count=%d err=%v", got.State, f.counter.count(inv.ID), err)
			}
		})
	}
}

func TestPGExecuteVerifiedAndNoReplay(t *testing.T) {
	f := newPGFixture(t)
	g := f.grant(t, 1)
	inv := f.admit(t, g, "once")
	got, err := f.store.Execute(t.Context(), inv.ID)
	if err != nil || got.State != "verified" {
		t.Fatalf("execute state=%q err=%v result=%s", got.State, err, got.Result)
	}
	if f.counter.count(inv.ID) != 1 {
		t.Fatalf("effect count=%d", f.counter.count(inv.ID))
	}
	again, err := f.store.Execute(t.Context(), inv.ID)
	if err != nil || again.State != "verified" || f.counter.count(inv.ID) != 1 {
		t.Fatalf("replayed effect: state=%q count=%d err=%v", again.State, f.counter.count(inv.ID), err)
	}
	duplicate, reused, err := f.store.Admit(t.Context(), f.userID, g.ID, g.Digest, "once", Input{Counter: "ledger"})
	if err != nil || !reused || duplicate.ID != inv.ID {
		t.Fatalf("duplicate receipt: %#v reused=%v err=%v", duplicate, reused, err)
	}
}

func TestPGClaimFencingAndCrashBoundaries(t *testing.T) {
	f := newPGFixture(t)
	g := f.grant(t, 2)
	first := f.admit(t, g, "crash-before")
	oldOwner, oldGeneration, err := f.store.Claim(t.Context(), first.ID)
	if err != nil {
		t.Fatal(err)
	}
	if f.counter.count(first.ID) != 0 {
		t.Fatal("claim caused effect")
	}
	f.setState(t, `UPDATE execution_invocations SET lease_until=NOW()-INTERVAL '1 second' WHERE id=$1`, first.ID)
	newOwner, newGeneration, err := f.store.Claim(t.Context(), first.ID)
	if err != nil || newGeneration <= oldGeneration {
		t.Fatalf("reclaim generation=%d old=%d err=%v", newGeneration, oldGeneration, err)
	}
	if err := f.store.Start(t.Context(), first.ID, oldOwner, oldGeneration); !errors.Is(err, ErrConflict) {
		t.Fatalf("stale owner started: %v", err)
	}
	if err := f.store.Start(t.Context(), first.ID, newOwner, newGeneration); err != nil {
		t.Fatal(err)
	}
	if f.counter.count(first.ID) != 0 {
		t.Fatal("start itself caused effect")
	}
	f.setState(t, `UPDATE execution_invocations SET lease_until=NOW()-INTERVAL '1 second' WHERE id=$1`, first.ID)
	changed, err := f.store.RecoverExpired(t.Context())
	if err != nil || changed < 1 {
		t.Fatalf("recover changed=%d err=%v", changed, err)
	}
	after, err := f.store.Execute(t.Context(), first.ID)
	if err != nil || after.State != "unknown_effect" || f.counter.count(first.ID) != 0 {
		t.Fatalf("crash-after-start replay: state=%q count=%d err=%v", after.State, f.counter.count(first.ID), err)
	}
	second := f.admit(t, g, "crash-after")
	owner, generation, err := f.store.Claim(t.Context(), second.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.store.Start(t.Context(), second.ID, owner, generation); err != nil {
		t.Fatal(err)
	}
	if err := f.store.finish(t.Context(), second.ID, owner, generation, "verified", countEvidence{InvocationID: second.ID, Counter: "ledger", ObservedCount: 1}); err != nil {
		t.Fatal(err)
	}
	late, err := f.store.Get(t.Context(), f.userID, second.ID)
	if err != nil || late.State != "verified" {
		t.Fatalf("on-time observation: %q %v", late.State, err)
	}
}

func TestPGResponseLossUnknownAndReconciliation(t *testing.T) {
	f := newPGFixture(t)
	g := f.grant(t, 1)
	inv := f.admit(t, g, "lost-response")
	f.counter.mu.Lock()
	f.counter.dropResponse = true
	f.counter.mu.Unlock()
	got, err := f.store.Execute(t.Context(), inv.ID)
	if err != nil || got.State != "unknown_effect" || f.counter.count(inv.ID) != 1 {
		t.Fatalf("response loss state=%q count=%d err=%v", got.State, f.counter.count(inv.ID), err)
	}
	again, err := f.store.Execute(t.Context(), inv.ID)
	if err != nil || again.State != "unknown_effect" || f.counter.count(inv.ID) != 1 {
		t.Fatalf("unknown replay state=%q count=%d err=%v", again.State, f.counter.count(inv.ID), err)
	}
	if _, err := f.store.Reconcile(t.Context(), f.userID, inv.ID, "committed", Evidence{Source: "fixture-readback", Summary: "one committed increment", InvocationID: inv.ID}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("evidence-free reconcile: %v", err)
	}
	count := f.counter.count(inv.ID)
	reconciled, err := f.store.Reconcile(t.Context(), f.userID, inv.ID, "committed", Evidence{Source: "fixture-readback", Summary: "one committed increment", InvocationID: inv.ID, ObservedCount: &count})
	if err != nil || reconciled.State != "reconciled" {
		t.Fatalf("reconcile state=%q err=%v", reconciled.State, err)
	}
	if f.counter.count(inv.ID) != 1 {
		t.Fatal("reconciliation replayed effect")
	}
}

func TestPGLateObservationDoesNotReplaceUnknown(t *testing.T) {
	f := newPGFixture(t)
	g := f.grant(t, 1)
	inv := f.admit(t, g, "late")
	owner, generation, err := f.store.Claim(t.Context(), inv.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.store.Start(t.Context(), inv.ID, owner, generation); err != nil {
		t.Fatal(err)
	}
	f.setState(t, `UPDATE execution_invocations SET lease_until=$2 WHERE id=$1`, inv.ID, time.Now().Add(-time.Second))
	if err := f.store.finish(t.Context(), inv.ID, owner, generation, "verified", countEvidence{InvocationID: inv.ID, Counter: "ledger", ObservedCount: 1}); err != nil {
		t.Fatal(err)
	}
	got, err := f.store.Get(t.Context(), f.userID, inv.ID)
	if err != nil || got.State != "unknown_effect" {
		t.Fatalf("late evidence state=%q err=%v", got.State, err)
	}
	if f.counter.count(inv.ID) != 0 {
		t.Fatal("late evidence caused effect")
	}
}

func TestPGUnknownVerificationNeverReexecutesOrRefunds(t *testing.T) {
	f := newPGFixture(t)
	g := f.grant(t, 1)
	inv := f.admit(t, g, "verify-zero")
	owner, generation, err := f.store.Claim(t.Context(), inv.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.store.Start(t.Context(), inv.ID, owner, generation); err != nil {
		t.Fatal(err)
	}
	f.setState(t, `UPDATE execution_invocations SET lease_until=NOW()-INTERVAL '1 second' WHERE id=$1`, inv.ID)
	if _, err := f.store.RecoverExpired(t.Context()); err != nil {
		t.Fatal(err)
	}
	unknown, err := f.store.Reconcile(t.Context(), f.userID, inv.ID, "still_unknown", Evidence{Source: "operator-review", Summary: "no conclusive receipt", InvocationID: inv.ID})
	if err != nil || unknown.State != "unknown_effect" {
		t.Fatalf("still unknown state=%q err=%v", unknown.State, err)
	}
	zero := f.counter.count(inv.ID)
	closed, err := f.store.Reconcile(t.Context(), f.userID, inv.ID, "not_committed", Evidence{Source: "fixture-readback", Summary: "no increment found", InvocationID: inv.ID, ObservedCount: &zero})
	if err != nil || closed.State != "reconciled" {
		t.Fatalf("not committed state=%q err=%v", closed.State, err)
	}
	if _, err := f.store.Execute(t.Context(), inv.ID); err != nil {
		t.Fatal(err)
	}
	if f.counter.count(inv.ID) != 0 {
		t.Fatal("verification replayed effect")
	}
	if _, _, err := f.store.Admit(t.Context(), f.userID, g.ID, g.Digest, "new-attempt", Input{Counter: "ledger"}); !errors.Is(err, ErrDenied) {
		t.Fatalf("budget refunded after non-effect: %v", err)
	}
}

func TestPGObservedResponseNeedsIndependentEvidence(t *testing.T) {
	f := newPGFixture(t)
	g := f.grant(t, 1)
	inv := f.admit(t, g, "observed-only")
	f.counter.mu.Lock()
	f.counter.unavailableRead = true
	f.counter.mu.Unlock()
	got, err := f.store.Execute(t.Context(), inv.ID)
	if err != nil || got.State != "observed" || f.counter.count(inv.ID) != 1 {
		t.Fatalf("unverified observation state=%q count=%d err=%v", got.State, f.counter.count(inv.ID), err)
	}
	count := f.counter.count(inv.ID)
	closed, err := f.store.Reconcile(t.Context(), f.userID, inv.ID, "committed", Evidence{Source: "operator-inspection", Summary: "count service audit contains invocation", InvocationID: inv.ID, ObservedCount: &count})
	if err != nil || closed.State != "reconciled" {
		t.Fatalf("observation reconciliation state=%q err=%v", closed.State, err)
	}
	if f.counter.count(inv.ID) != 1 {
		t.Fatal("observed reconciliation replayed effect")
	}
}

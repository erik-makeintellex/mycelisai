package invocation

import (
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestPGProposalConfirmationAndImmutableGrant(t *testing.T) {
	f := newPGFixture(t)
	proposal, err := f.store.Propose(t.Context(), f.userID, Proposal{GroupID: f.groupID, Counter: "ledger", Budget: 2})
	if err != nil {
		t.Fatal(err)
	}
	if proposal.ProofID == "" || proposal.ContractID == "" || proposal.ConfirmToken == "" {
		t.Fatalf("missing proposal refs: %#v", proposal)
	}
	handles, err := f.store.HandlesToken(t.Context(), proposal.ConfirmToken)
	if err != nil || !handles {
		t.Fatalf("counting token not intercepted: %v %v", handles, err)
	}
	if _, err := f.store.Confirm(t.Context(), uuid.NewString(), proposal.ConfirmToken); !errors.Is(err, ErrDenied) {
		t.Fatalf("forged subject: %v", err)
	}
	grant, err := f.store.Confirm(t.Context(), f.userID, proposal.ConfirmToken)
	if err != nil {
		t.Fatal(err)
	}
	if grant.ProofID != proposal.ProofID || grant.ContractID != proposal.ContractID || grant.Budget != 2 || grant.Digest == "" {
		t.Fatalf("grant mismatch: %#v", grant)
	}
	if _, err := f.store.Confirm(t.Context(), f.userID, proposal.ConfirmToken); !errors.Is(err, ErrDenied) {
		t.Fatalf("double confirm: %v", err)
	}
	if _, err := f.db.ExecContext(t.Context(), `UPDATE execution_effect_grants SET unit_budget=99 WHERE id=$1`, grant.ID); err == nil {
		t.Fatal("immutable grant UPDATE succeeded")
	}
	if _, err := f.db.ExecContext(t.Context(), `DELETE FROM execution_effect_grants WHERE id=$1`, grant.ID); err == nil {
		t.Fatal("immutable grant DELETE succeeded")
	}
}

func TestPGXProposalConfirmAdmit(t *testing.T) {
	f := newPGFixture(t)
	pgxDB, err := sql.Open("pgx", os.Getenv("MYCELIS_INVOCATION_TEST_DSN"))
	if err != nil {
		t.Fatal(err)
	}
	defer pgxDB.Close()
	pgxStore := NewStore(pgxDB)
	proposal, err := pgxStore.Propose(t.Context(), f.userID, Proposal{GroupID: f.groupID, Counter: "ledger", Budget: 1})
	if err != nil {
		t.Fatalf("pgx propose: %v", err)
	}
	g, err := pgxStore.Confirm(t.Context(), f.userID, proposal.ConfirmToken)
	if err != nil {
		t.Fatalf("pgx confirm: %v", err)
	}
	_, _, err = pgxStore.Admit(t.Context(), f.userID, g.ID, g.Digest, "pgx-throughout", Input{Counter: "ledger"})
	if err != nil {
		t.Fatalf("pgx admit after pgx proposal and confirm: %v", err)
	}
}

func TestPGRestoredAuthorityRequiresFreshGrant(t *testing.T) {
	f := newPGFixture(t)
	old := f.grant(t, 1)
	f.setState(t, `UPDATE org_memberships SET status='disabled' WHERE id=$1`, f.membershipID)
	f.setState(t, `UPDATE org_memberships SET status='active' WHERE id=$1`, f.membershipID)
	if _, _, err := f.store.Admit(t.Context(), f.userID, old.ID, old.Digest, "restored-old", Input{Counter: "ledger"}); !errors.Is(err, ErrDenied) {
		t.Fatalf("restored status revived old grant: %v", err)
	}
	fresh := f.grant(t, 1)
	if _, _, err := f.store.Admit(t.Context(), f.userID, fresh.ID, fresh.Digest, "restored-fresh", Input{Counter: "ledger"}); err != nil {
		t.Fatalf("fresh grant after restoration: %v", err)
	}
}

func TestPGProposalRejectsAbsentOrForeignAuthority(t *testing.T) {
	f := newPGFixture(t)
	request := Proposal{GroupID: f.groupID, Counter: "ledger", Budget: 1}
	if _, err := f.store.Propose(t.Context(), uuid.NewString(), request); !errors.Is(err, ErrDenied) {
		t.Fatalf("unprovisioned user: %v", err)
	}
	request.GroupID = uuid.NewString()
	if _, err := f.store.Propose(t.Context(), f.userID, request); !errors.Is(err, ErrDenied) {
		t.Fatalf("foreign group: %v", err)
	}
	request.GroupID = f.groupID
	f.setState(t, `UPDATE org_memberships SET status='disabled' WHERE id=$1`, f.membershipID)
	if _, err := f.store.Propose(t.Context(), f.userID, request); !errors.Is(err, ErrDenied) {
		t.Fatalf("disabled membership: %v", err)
	}
	f.setState(t, `UPDATE org_memberships SET status='active',expires_at=NOW()-INTERVAL '1 second' WHERE id=$1`, f.membershipID)
	if _, err := f.store.Propose(t.Context(), f.userID, request); !errors.Is(err, ErrDenied) {
		t.Fatalf("expired membership: %v", err)
	}
	f.setState(t, `UPDATE org_memberships SET expires_at=NULL WHERE id=$1`, f.membershipID)
	f.setState(t, `DELETE FROM role_permissions WHERE role_id=$1 AND permission_key=$2`, f.roleID, Permission)
	if _, err := f.store.Propose(t.Context(), f.userID, request); !errors.Is(err, ErrDenied) {
		t.Fatalf("missing permission: %v", err)
	}
}

func TestPGAmbiguousMembershipFailsClosed(t *testing.T) {
	f := newPGFixture(t)
	secondRole, secondMembership := uuid.NewString(), uuid.NewString()
	f.setState(t, `INSERT INTO roles(id,account_id,key,name,scope) VALUES($1,$2,$3,'Second','group')`, secondRole, f.accountID, "second-"+secondRole)
	f.setState(t, `INSERT INTO role_permissions(role_id,permission_key) VALUES($1,$2)`, secondRole, Permission)
	f.setState(t, `INSERT INTO org_memberships(id,account_id,user_id,group_id,role_id,status) VALUES($1,$2,$3,$4,$5,'active')`, secondMembership, f.accountID, f.userID, f.groupID, secondRole)
	_, err := f.store.Propose(t.Context(), f.userID, Proposal{GroupID: f.groupID, Counter: "ledger", Budget: 1})
	if !errors.Is(err, ErrDenied) {
		t.Fatalf("ambiguous authority: %v", err)
	}
}

func TestPGConfirmationRejectsMissingOrChangedBoundary(t *testing.T) {
	cases := []struct{ name, mutate string }{
		{"missing boundary", `UPDATE intent_proofs SET scope_validation=scope_validation-'invocation_boundary' WHERE id=$1`},
		{"wrong action", `UPDATE intent_proofs SET resolved_intent='other.action' WHERE id=$1`},
		{"contract mismatch", `UPDATE execution_contracts SET resolved_intent='other.action' WHERE intent_proof_id=$1`},
		{"expired proof", `UPDATE intent_proofs SET expires_at=NOW()-INTERVAL '1 second' WHERE id=$1`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newPGFixture(t)
			proposal, err := f.store.Propose(t.Context(), f.userID, Proposal{GroupID: f.groupID, Counter: "ledger", Budget: 1})
			if err != nil {
				t.Fatal(err)
			}
			f.setState(t, tc.mutate, proposal.ProofID)
			handles, err := f.store.HandlesToken(t.Context(), proposal.ConfirmToken)
			if err != nil || !handles {
				t.Fatalf("counting proof bypassed: %v %v", handles, err)
			}
			if _, err := f.store.Confirm(t.Context(), f.userID, proposal.ConfirmToken); !errors.Is(err, ErrDenied) {
				t.Fatalf("changed proof accepted: %v", err)
			}
		})
	}
}

func TestPGConfirmationRejectsTokenAndBindingDrift(t *testing.T) {
	t.Run("expired token", func(t *testing.T) {
		f := newPGFixture(t)
		p, err := f.store.Propose(t.Context(), f.userID, Proposal{GroupID: f.groupID, Counter: "ledger", Budget: 1})
		if err != nil {
			t.Fatal(err)
		}
		f.setState(t, `UPDATE confirm_tokens SET expires_at=NOW()-INTERVAL '1 second' WHERE token=$1`, p.ConfirmToken)
		if _, err := f.store.Confirm(t.Context(), f.userID, p.ConfirmToken); !errors.Is(err, ErrDenied) {
			t.Fatalf("expired token: %v", err)
		}
	})
	t.Run("binding drift", func(t *testing.T) {
		f := newPGFixture(t)
		p, err := f.store.Propose(t.Context(), f.userID, Proposal{GroupID: f.groupID, Counter: "ledger", Budget: 1})
		if err != nil {
			t.Fatal(err)
		}
		f.setBinding(t, f.counter.server.URL+"/changed")
		if _, err := f.store.Confirm(t.Context(), f.userID, p.ConfirmToken); !errors.Is(err, ErrDenied) {
			t.Fatalf("changed binding: %v", err)
		}
	})
	t.Run("authority expiry before confirmation", func(t *testing.T) {
		f := newPGFixture(t)
		p, err := f.store.Propose(t.Context(), f.userID, Proposal{GroupID: f.groupID, Counter: "ledger", Budget: 1})
		if err != nil {
			t.Fatal(err)
		}
		f.setState(t, `UPDATE org_memberships SET expires_at=$2 WHERE id=$1`, f.membershipID, time.Now().Add(-time.Second))
		if _, err := f.store.Confirm(t.Context(), f.userID, p.ConfirmToken); !errors.Is(err, ErrDenied) {
			t.Fatalf("expired authority: %v", err)
		}
	})
}

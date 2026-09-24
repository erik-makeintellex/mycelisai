package identity

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type teamFixture struct {
	db                                        *sql.DB
	store                                     *TeamOwnershipStore
	account, actor, group, role, member, team string
}

func newTeamFixture(t *testing.T) *teamFixture {
	t.Helper()
	dsn := os.Getenv("MYCELIS_TEAM_OWNERSHIP_TEST_DSN")
	if dsn == "" {
		t.Skip("requires isolated MYCELIS_TEAM_OWNERSHIP_TEST_DSN")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.PingContext(t.Context()); err != nil {
		t.Fatal(err)
	}
	f := &teamFixture{db: db, store: NewTeamOwnershipStore(db), account: uuid.NewString(), actor: uuid.NewString(), group: uuid.NewString(), role: uuid.NewString(), member: uuid.NewString(), team: "team-" + uuid.NewString()}
	f.exec(t, `INSERT INTO accounts(id,slug,name) VALUES($1,$2,'Team ownership fixture')`, f.account, f.account)
	f.exec(t, `INSERT INTO users(id,username,account_id) VALUES($1,$2,$3)`, f.actor, f.actor, f.account)
	f.exec(t, `INSERT INTO groups(id,account_id,key,name) VALUES($1,$2,$3,'Team group')`, f.group, f.account, f.group)
	f.exec(t, `INSERT INTO roles(id,key,name,scope) VALUES($1,$2,'Deployment owner','system')`, f.role, f.role)
	f.exec(t, `INSERT INTO role_permissions(role_id,permission_key) VALUES($1,$2)`, f.role, TeamProvisionPermission)
	f.exec(t, `INSERT INTO org_memberships(id,account_id,user_id,role_id) VALUES($1,$2,$3,$4)`, f.member, f.account, f.actor, f.role)
	f.exec(t, `INSERT INTO runtime_team_manifests(tenant_id,team_id,manifest_digest,manifest) VALUES('default',$1,'sha256:fixture',jsonb_build_object('id',$1::text))`, f.team)
	return f
}

func (f *teamFixture) exec(t *testing.T, query string, args ...any) {
	t.Helper()
	if _, err := f.db.ExecContext(t.Context(), query, args...); err != nil {
		t.Fatalf("%s: %v", query, err)
	}
}

func (f *teamFixture) inspect(t *testing.T) TeamOwnership {
	t.Helper()
	row, err := f.store.Inspect(t.Context(), f.actor, f.team, f.group)
	if err != nil {
		t.Fatal(err)
	}
	return row
}

func TestTeamOwnershipPostgresBindRevokeAudit(t *testing.T) {
	f := newTeamFixture(t)
	initial := f.inspect(t)
	if initial.OwnerGroupID != "" || initial.Revision == "" {
		t.Fatalf("unexpected initial ownership: %+v", initial)
	}
	bound, err := f.store.Bind(t.Context(), f.actor, f.team, f.group, initial.ManifestDigest, initial.Revision, "Reviewed named deployment team")
	if err != nil {
		t.Fatal(err)
	}
	if bound.OwnerGroupID != f.group || bound.OwnerAccountID != f.account || bound.ProvisionedBy != f.actor || bound.ProvisionedAt == nil {
		t.Fatalf("bind did not retain owner evidence: %+v", bound)
	}
	if _, err := f.store.Bind(t.Context(), f.actor, f.team, f.group, bound.ManifestDigest, bound.Revision, "overwrite active owner"); !errors.Is(err, ErrTeamOwnershipConflict) {
		t.Fatalf("active owner overwrite: %v", err)
	}
	if _, err := f.store.Bind(t.Context(), f.actor, f.team, f.group, initial.ManifestDigest, initial.Revision, "duplicate"); !errors.Is(err, ErrTeamOwnershipConflict) {
		t.Fatalf("duplicate bind: %v", err)
	}
	if _, err := f.store.Revoke(t.Context(), f.actor, f.team, f.group, bound.ManifestDigest, initial.Revision, "stale"); !errors.Is(err, ErrTeamOwnershipConflict) {
		t.Fatalf("stale revision: %v", err)
	}
	revoked, err := f.store.Revoke(t.Context(), f.actor, f.team, f.group, bound.ManifestDigest, bound.Revision, "Operator retirement")
	if err != nil {
		t.Fatal(err)
	}
	if revoked.RevokedAt == nil || revoked.RevokedBy != f.actor || revoked.OwnerGroupID != f.group || revoked.ProvisionedBy != f.actor {
		t.Fatalf("revocation lost evidence: %+v", revoked)
	}
	if _, err := f.store.Bind(t.Context(), f.actor, f.team, f.group, revoked.ManifestDigest, bound.Revision, "stale rebind"); !errors.Is(err, ErrTeamOwnershipConflict) {
		t.Fatalf("stale rebind: %v", err)
	}
	rebound, err := f.store.Bind(t.Context(), f.actor, f.team, f.group, revoked.ManifestDigest, revoked.Revision, "Reviewed renewed ownership")
	if err != nil || rebound.RevokedAt != nil {
		t.Fatalf("rebind: %+v %v", rebound, err)
	}
	var audits int
	if err := f.db.QueryRow(`SELECT count(*) FROM identity_audit_events WHERE target_id=$1 AND event_type LIKE 'framework_runs.team_ownership_%'`, f.team).Scan(&audits); err != nil || audits != 3 {
		t.Fatalf("audits=%d err=%v", audits, err)
	}
}

func TestTeamOwnershipPostgresDeniedAuthority(t *testing.T) {
	f := newTeamFixture(t)
	for name, actor := range map[string]string{"forged": uuid.NewString(), "malformed": "web-subject"} {
		t.Run(name, func(t *testing.T) {
			_, err := f.store.Inspect(t.Context(), actor, f.team, f.group)
			if !errors.Is(err, ErrTeamOwnershipDenied) {
				t.Fatalf("expected denial, got %v", err)
			}
		})
	}
	otherAccount, otherGroup := uuid.NewString(), uuid.NewString()
	f.exec(t, `INSERT INTO accounts(id,slug,name) VALUES($1,$2,'Other account')`, otherAccount, otherAccount)
	f.exec(t, `INSERT INTO groups(id,account_id,key,name) VALUES($1,$2,$3,'Other group')`, otherGroup, otherAccount, otherGroup)
	if _, err := f.store.Inspect(t.Context(), f.actor, f.team, otherGroup); !errors.Is(err, ErrTeamOwnershipDenied) {
		t.Fatalf("cross-account group: %v", err)
	}
	f.exec(t, `UPDATE org_memberships SET expires_at=clock_timestamp()-interval '1 second' WHERE id=$1`, f.member)
	if _, err := f.store.Inspect(t.Context(), f.actor, f.team, f.group); !errors.Is(err, ErrTeamOwnershipDenied) {
		t.Fatalf("expired membership: %v", err)
	}
	f.exec(t, `UPDATE org_memberships SET expires_at=NULL WHERE id=$1`, f.member)
	f.exec(t, `UPDATE users SET status='disabled' WHERE id=$1`, f.actor)
	if _, err := f.store.Inspect(t.Context(), f.actor, f.team, f.group); !errors.Is(err, ErrTeamOwnershipDenied) {
		t.Fatalf("inactive actor: %v", err)
	}
	f.exec(t, `UPDATE users SET status='active' WHERE id=$1`, f.actor)
	f.exec(t, `UPDATE accounts SET status='paused' WHERE id=$1`, f.account)
	if _, err := f.store.Inspect(t.Context(), f.actor, f.team, f.group); !errors.Is(err, ErrTeamOwnershipDenied) {
		t.Fatalf("inactive account: %v", err)
	}
	f.exec(t, `UPDATE accounts SET status='active' WHERE id=$1`, f.account)
	f.exec(t, `DELETE FROM role_permissions WHERE role_id=$1`, f.role)
	if _, err := f.store.Inspect(t.Context(), f.actor, f.team, f.group); !errors.Is(err, ErrTeamOwnershipDenied) {
		t.Fatalf("missing permission: %v", err)
	}
	f.exec(t, `INSERT INTO role_permissions(role_id,permission_key) VALUES($1,$2)`, f.role, TeamProvisionPermission)
	f.exec(t, `UPDATE org_memberships SET group_id=$2 WHERE id=$1`, f.member, f.group)
	if _, err := f.store.Inspect(t.Context(), f.actor, f.team, f.group); !errors.Is(err, ErrTeamOwnershipDenied) {
		t.Fatalf("group-local permission: %v", err)
	}
	f.exec(t, `UPDATE org_memberships SET group_id=NULL WHERE id=$1`, f.member)
	f.exec(t, `UPDATE roles SET scope='account',account_id=$2 WHERE id=$1`, f.role, f.account)
	if _, err := f.store.Inspect(t.Context(), f.actor, f.team, f.group); !errors.Is(err, ErrTeamOwnershipDenied) {
		t.Fatalf("account role without system scope: %v", err)
	}
	f.exec(t, `UPDATE roles SET scope='system',account_id=NULL WHERE id=$1`, f.role)
	f.exec(t, `UPDATE accounts SET tenant_id='other-tenant' WHERE id=$1`, f.account)
	if _, err := f.store.Inspect(t.Context(), f.actor, f.team, f.group); !errors.Is(err, ErrTeamOwnershipMissing) {
		t.Fatalf("cross-tenant manifest should not be addressable: %v", err)
	}
}

func TestTeamOwnershipPostgresAmbiguousAndRacingBind(t *testing.T) {
	f := newTeamFixture(t)
	secondRole, secondMember := uuid.NewString(), uuid.NewString()
	f.exec(t, `INSERT INTO roles(id,key,name,scope) VALUES($1,$2,'Second owner','system')`, secondRole, secondRole)
	f.exec(t, `INSERT INTO role_permissions(role_id,permission_key) VALUES($1,$2)`, secondRole, TeamProvisionPermission)
	f.exec(t, `INSERT INTO org_memberships(id,account_id,user_id,role_id) VALUES($1,$2,$3,$4)`, secondMember, f.account, f.actor, secondRole)
	if _, err := f.store.Inspect(t.Context(), f.actor, f.team, f.group); !errors.Is(err, ErrTeamOwnershipDenied) {
		t.Fatalf("ambiguous authority: %v", err)
	}
	f.exec(t, `UPDATE org_memberships SET status='disabled' WHERE id=$1`, secondMember)
	initial := f.inspect(t)
	if _, err := f.store.Bind(t.Context(), f.actor, f.team, f.group, "sha256:stale", initial.Revision, "Bad manifest digest"); !errors.Is(err, ErrTeamOwnershipConflict) {
		t.Fatalf("stale manifest digest: %v", err)
	}
	var beforeRaceAudits int
	if err := f.db.QueryRowContext(t.Context(), `SELECT count(*) FROM identity_audit_events WHERE target_id=$1`, f.team).Scan(&beforeRaceAudits); err != nil || beforeRaceAudits != 0 {
		t.Fatalf("denied bind emitted audit: count=%d err=%v", beforeRaceAudits, err)
	}
	start := make(chan struct{})
	results := make(chan error, 2)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, err := f.store.Bind(context.Background(), f.actor, f.team, f.group, initial.ManifestDigest, initial.Revision, "Concurrent reviewed assignment")
			results <- err
		}()
	}
	close(start)
	wg.Wait()
	close(results)
	success, conflict := 0, 0
	for err := range results {
		switch {
		case err == nil:
			success++
		case errors.Is(err, ErrTeamOwnershipConflict):
			conflict++
		default:
			t.Fatalf("unexpected race error: %v", err)
		}
	}
	if success != 1 || conflict != 1 {
		t.Fatalf("race success=%d conflict=%d", success, conflict)
	}
	bound := f.inspect(t)
	start = make(chan struct{})
	results = make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, err := f.store.Revoke(context.Background(), f.actor, f.team, f.group, bound.ManifestDigest, bound.Revision, "Concurrent retirement")
			results <- err
		}()
	}
	close(start)
	wg.Wait()
	close(results)
	success, conflict = 0, 0
	for err := range results {
		switch {
		case err == nil:
			success++
		case errors.Is(err, ErrTeamOwnershipConflict):
			conflict++
		default:
			t.Fatalf("unexpected revoke race error: %v", err)
		}
	}
	if success != 1 || conflict != 1 {
		t.Fatalf("revoke race success=%d conflict=%d", success, conflict)
	}
}

func TestTeamOwnershipPostgresAuditFailureRollsBack(t *testing.T) {
	f := newTeamFixture(t)
	initial := f.inspect(t)
	name := "reject_ownership_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	f.exec(t, fmt.Sprintf(`CREATE FUNCTION %s() RETURNS trigger LANGUAGE plpgsql AS $$
		BEGIN IF NEW.target_id='%s' THEN RAISE EXCEPTION 'fixture audit rejection'; END IF; RETURN NEW; END $$`, name, f.team))
	f.exec(t, fmt.Sprintf(`CREATE TRIGGER %s BEFORE INSERT ON identity_audit_events FOR EACH ROW EXECUTE FUNCTION %s()`, name, name))
	t.Cleanup(func() {
		_, _ = f.db.Exec(`DROP TRIGGER IF EXISTS ` + name + ` ON identity_audit_events`)
		_, _ = f.db.Exec(`DROP FUNCTION IF EXISTS ` + name + `() `)
	})
	if _, err := f.store.Bind(t.Context(), f.actor, f.team, f.group, initial.ManifestDigest, initial.Revision, "Reviewed assignment"); err == nil {
		t.Fatal("expected audit failure")
	}
	after := f.inspect(t)
	if after.OwnerGroupID != "" || after.Revision != initial.Revision {
		t.Fatalf("audit failure committed ownership: %+v", after)
	}
}

func TestTeamOwnershipPostgresGroupMembershipBecomesAccountLevel(t *testing.T) {
	f := newTeamFixture(t)
	groupScoped := uuid.NewString()
	secondRole := uuid.NewString()
	f.exec(t, `INSERT INTO roles(id,key,name,scope) VALUES($1,$2,'Promotion role','system')`, secondRole, secondRole)
	f.exec(t, `INSERT INTO role_permissions(role_id,permission_key) VALUES($1,$2)`, secondRole, TeamProvisionPermission)
	f.exec(t, `INSERT INTO org_memberships(id,account_id,user_id,group_id,role_id)
		VALUES($1,$2,$3,$4,$5)`, groupScoped, f.account, f.actor, f.group, secondRole)
	blocker, err := f.db.BeginTx(t.Context(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer blocker.Rollback()
	var locked string
	if err := blocker.QueryRowContext(t.Context(), `SELECT id::text FROM org_memberships WHERE id=$1 FOR UPDATE`, groupScoped).Scan(&locked); err != nil {
		t.Fatal(err)
	}
	result := make(chan error, 1)
	go func() {
		_, err := f.store.Inspect(context.Background(), f.actor, f.team, f.group)
		result <- err
	}()
	deadline := time.Now().Add(3 * time.Second)
	blocked := false
	for time.Now().Before(deadline) {
		var waiting int
		err := f.db.QueryRowContext(t.Context(), `SELECT count(*) FROM pg_stat_activity
			WHERE datname=current_database() AND wait_event_type='Lock'
			AND query LIKE '%FROM org_memberships%'`).Scan(&waiting)
		if err != nil {
			t.Fatal(err)
		}
		if waiting > 0 {
			blocked = true
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if !blocked {
		t.Fatal("inspect did not lock existing group-scoped membership")
	}
	if _, err := blocker.ExecContext(t.Context(), `UPDATE org_memberships SET group_id=NULL WHERE id=$1`, groupScoped); err != nil {
		t.Fatal(err)
	}
	if err := blocker.Commit(); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-result:
		if !errors.Is(err, ErrTeamOwnershipDenied) {
			t.Fatalf("membership promoted while inspect blocked: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("inspect did not resume")
	}
}

func TestTeamOwnershipPostgresExpiryWhileWaitingForManifest(t *testing.T) {
	f := newTeamFixture(t)
	initial := f.inspect(t)
	f.exec(t, `UPDATE org_memberships SET expires_at=clock_timestamp()+interval '3 seconds' WHERE id=$1`, f.member)
	blocker, err := f.db.BeginTx(t.Context(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer blocker.Rollback()
	var locked string
	if err := blocker.QueryRowContext(t.Context(), `SELECT team_id FROM runtime_team_manifests WHERE team_id=$1 FOR UPDATE`, f.team).Scan(&locked); err != nil {
		t.Fatal(err)
	}
	result := make(chan error, 1)
	go func() {
		_, err := f.store.Bind(context.Background(), f.actor, f.team, f.group, initial.ManifestDigest, initial.Revision, "Timed operator assignment")
		result <- err
	}()
	deadline := time.Now().Add(6 * time.Second)
	blocked := false
	for time.Now().Before(deadline) {
		var waiting int
		if err := f.db.QueryRowContext(t.Context(), `SELECT count(*) FROM pg_stat_activity
			WHERE datname=current_database() AND wait_event_type='Lock'
			AND query LIKE '%FROM runtime_team_manifests%'`).Scan(&waiting); err != nil {
			t.Fatal(err)
		}
		if waiting > 0 {
			blocked = true
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if !blocked {
		t.Fatal("bind did not wait for manifest row")
	}
	for time.Now().Before(deadline) {
		var expired bool
		if err := f.db.QueryRowContext(t.Context(), `SELECT expires_at<=clock_timestamp() FROM org_memberships WHERE id=$1`, f.member).Scan(&expired); err != nil {
			t.Fatal(err)
		}
		if expired {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if err := blocker.Commit(); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-result:
		if !errors.Is(err, ErrTeamOwnershipDenied) {
			t.Fatalf("expired while waiting for manifest: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("bind did not resume")
	}
	var owner sql.NullString
	if err := f.db.QueryRowContext(t.Context(), `SELECT owner_group_id::text FROM runtime_team_manifests WHERE team_id=$1`, f.team).Scan(&owner); err != nil {
		t.Fatal(err)
	}
	if owner.Valid {
		t.Fatalf("expired authority bound group %s", owner.String)
	}
	var audits int
	if err := f.db.QueryRowContext(t.Context(), `SELECT count(*) FROM identity_audit_events WHERE target_id=$1`, f.team).Scan(&audits); err != nil || audits != 0 {
		t.Fatalf("expired bind audits=%d err=%v", audits, err)
	}
}

package swarm

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/mycelis/core/internal/governance"
)

func TestDeleteRuntimeTeamRefusesActiveOwnership(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT owner_group_id::text, ownership_revoked_at").
		WithArgs(durableRuntimeTenant, "owned-team").
		WillReturnRows(sqlmock.NewRows([]string{"owner_group_id", "ownership_revoked_at"}).
			AddRow("group-1", nil))
	mock.ExpectRollback()
	err = NewPostgresDurableTeamLoader(db).DeleteRuntimeTeam(context.Background(), "owned-team")
	if !errors.Is(err, ErrRuntimeTeamOwnershipActive) {
		t.Fatalf("delete active owner = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteRuntimeTeamAllowsRevokedAndAbsentRows(t *testing.T) {
	for _, tc := range []struct {
		name, owner string
		revokedAt   any
		found       bool
	}{
		{name: "legacy", found: true},
		{name: "revoked", owner: "group-1", revokedAt: time.Now().UTC(), found: true},
		{name: "absent"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = db.Close() })
			mock.ExpectBegin()
			query := mock.ExpectQuery("SELECT owner_group_id::text, ownership_revoked_at").
				WithArgs(durableRuntimeTenant, "team")
			if tc.found {
				var owner any
				if tc.owner != "" {
					owner = tc.owner
				}
				query.WillReturnRows(sqlmock.NewRows([]string{"owner_group_id", "ownership_revoked_at"}).
					AddRow(owner, tc.revokedAt))
				mock.ExpectExec("DELETE FROM runtime_team_manifests").
					WithArgs(durableRuntimeTenant, "team").
					WillReturnResult(sqlmock.NewResult(0, 1))
			} else {
				query.WillReturnError(sql.ErrNoRows)
			}
			mock.ExpectCommit()
			if err := NewPostgresDurableTeamLoader(db).DeleteRuntimeTeam(context.Background(), "team"); err != nil {
				t.Fatalf("delete %s: %v", tc.name, err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

type refusingDurableTeamStore struct{ memoryDurableTeamStore }

func (*refusingDurableTeamStore) DeleteRuntimeTeam(context.Context, string) error {
	return ErrRuntimeTeamOwnershipActive
}

func TestStopTeamDurablyPreservesRuntimeOnOwnershipRefusal(t *testing.T) {
	_, nc := startTestNATS(t)
	store := &refusingDurableTeamStore{}
	soma := NewSoma(nc, &governance.Guard{}, NewRegistryFromRuntimeOrganization(&RuntimeOrganization{}), nil, nil, nil, nil)
	soma.SetDurableTeamStore(store)
	manifest := completeDurableManifest()
	if err := soma.SpawnTeam(manifest); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(soma.Shutdown)
	found, err := soma.StopTeamDurably(manifest.ID)
	if !found || !errors.Is(err, ErrRuntimeTeamOwnershipActive) {
		t.Fatalf("stop owned runtime = (%v, %v)", found, err)
	}
	if got := soma.ListTeams(); len(got) != 1 || got[0].ID != manifest.ID {
		t.Fatalf("runtime team lost after refusal: %#v", got)
	}
	if !reflect.DeepEqual(store.saved, []*TeamManifest{manifest}) {
		t.Fatalf("persisted team changed after refusal: %#v", store.saved)
	}
}

func TestDurableTeamOwnershipDeleteBindRaceAndUnownedRecreationPostgres(t *testing.T) {
	dsn := strings.TrimSpace(os.Getenv("MYCELIS_TEAM_OWNERSHIP_TEST_DSN"))
	if dsn == "" {
		t.Skip("MYCELIS_TEAM_OWNERSHIP_TEST_DSN not set; real PostgreSQL proof requires a disposable schema")
	}
	parsedDSN, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	applicationName := "c2a-swarm-" + uuid.NewString()
	query := parsedDSN.Query()
	query.Set("application_name", applicationName)
	parsedDSN.RawQuery = query.Encode()
	db, err := sql.Open("pgx", parsedDSN.String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.PingContext(t.Context()); err != nil {
		t.Fatal(err)
	}
	var ownershipColumn bool
	if err := db.QueryRowContext(t.Context(), `SELECT EXISTS (
		SELECT 1 FROM information_schema.columns
		WHERE table_name='runtime_team_manifests' AND column_name='owner_group_id')`).Scan(&ownershipColumn); err != nil || !ownershipColumn {
		t.Fatalf("C2a ownership schema unavailable: %v", err)
	}
	accountID, groupID, userID := uuid.NewString(), uuid.NewString(), uuid.NewString()
	manifest := completeDurableManifest()
	manifest.ID = "team-" + uuid.NewString()
	ctx := t.Context()
	if _, err := db.ExecContext(ctx, `INSERT INTO accounts(id,tenant_id,slug,name,status)
		VALUES($1,'default',$2,'Ownership Test','active')`, accountID, "ownership-"+accountID); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := db.ExecContext(context.Background(), `DELETE FROM accounts WHERE id=$1`, accountID); err != nil {
			t.Errorf("cleanup ownership account: %v", err)
		}
	})
	if _, err := db.ExecContext(ctx, `INSERT INTO users(id,username,role,account_id,status)
		VALUES($1,$2,'operator',$3,'active')`, userID, "ownership-"+userID, accountID); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := db.ExecContext(context.Background(), `DELETE FROM users WHERE id=$1`, userID); err != nil {
			t.Errorf("cleanup ownership user: %v", err)
		}
	})
	if _, err := db.ExecContext(ctx, `INSERT INTO groups(id,account_id,key,name)
		VALUES($1,$2,$3,'Ownership Test')`, groupID, accountID, "ownership-"+groupID); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := db.ExecContext(context.Background(), `DELETE FROM groups WHERE id=$1`, groupID); err != nil {
			t.Errorf("cleanup ownership group: %v", err)
		}
	})
	store := NewPostgresDurableTeamLoader(db)
	if err := store.SaveRuntimeTeam(ctx, manifest); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := db.ExecContext(context.Background(), `DELETE FROM runtime_team_manifests WHERE tenant_id=$1 AND team_id=$2`, durableRuntimeTenant, manifest.ID); err != nil {
			t.Errorf("cleanup ownership manifest: %v", err)
		}
	})
	bindTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer bindTx.Rollback()
	var previousOwner sql.NullString
	if err := bindTx.QueryRowContext(ctx, `SELECT owner_group_id::text FROM runtime_team_manifests
		WHERE tenant_id=$1 AND team_id=$2 FOR UPDATE`, durableRuntimeTenant, manifest.ID).Scan(&previousOwner); err != nil || previousOwner.Valid {
		t.Fatalf("unowned row lock: %v, %#v", err, previousOwner)
	}
	deleteDone := make(chan error, 1)
	go func() { deleteDone <- store.DeleteRuntimeTeam(ctx, manifest.ID) }()
	deadline := time.Now().Add(5 * time.Second)
	for {
		var blocked bool
		if err := db.QueryRowContext(ctx, `SELECT EXISTS (
			SELECT 1 FROM pg_stat_activity
			WHERE application_name=$1 AND state='active' AND wait_event_type='Lock'
			AND query LIKE '%SELECT owner_group_id::text, ownership_revoked_at%'
		)`, applicationName).Scan(&blocked); err != nil {
			t.Fatalf("observe delete lock wait: %v", err)
		}
		if blocked {
			break
		}
		select {
		case err := <-deleteDone:
			t.Fatalf("delete completed before bind released row lock: %v", err)
		default:
		}
		if time.Now().After(deadline) {
			t.Fatal("delete did not block on the manifest row lock")
		}
		time.Sleep(10 * time.Millisecond)
	}
	if _, err := bindTx.ExecContext(ctx, `UPDATE runtime_team_manifests SET
		owner_account_id=$3,owner_group_id=$4,ownership_provisioned_by=$5,ownership_provisioned_at=NOW()
		WHERE tenant_id=$1 AND team_id=$2`, durableRuntimeTenant, manifest.ID, accountID, groupID, userID); err != nil {
		t.Fatal(err)
	}
	if err := bindTx.Commit(); err != nil {
		t.Fatal(err)
	}
	if err := <-deleteDone; !errors.Is(err, ErrRuntimeTeamOwnershipActive) {
		t.Fatalf("delete racing bind = %v", err)
	}
	if err := store.SaveRuntimeTeam(ctx, manifest); err != nil {
		t.Fatalf("idempotent save must preserve owner: %v", err)
	}
	var owner string
	var oldRevision string
	if err := db.QueryRowContext(ctx, `SELECT owner_group_id::text,xmin::text FROM runtime_team_manifests
		WHERE tenant_id=$1 AND team_id=$2`, durableRuntimeTenant, manifest.ID).Scan(&owner, &oldRevision); err != nil || owner != groupID {
		t.Fatalf("active owner after save = %q, %v", owner, err)
	}
	if _, err := db.ExecContext(ctx, `UPDATE runtime_team_manifests SET
		ownership_revoked_by=$3,ownership_revoked_at=NOW()
		WHERE tenant_id=$1 AND team_id=$2`, durableRuntimeTenant, manifest.ID, userID); err != nil {
		t.Fatal(err)
	}
	var revokedRevision string
	if err := db.QueryRowContext(ctx, `SELECT xmin::text FROM runtime_team_manifests
		WHERE tenant_id=$1 AND team_id=$2`, durableRuntimeTenant, manifest.ID).Scan(&revokedRevision); err != nil {
		t.Fatal(err)
	}
	if err := store.DeleteRuntimeTeam(ctx, manifest.ID); err != nil {
		t.Fatalf("delete revoked team: %v", err)
	}
	if err := store.SaveRuntimeTeam(ctx, manifest); err != nil {
		t.Fatalf("recreate unowned team: %v", err)
	}
	var newOwner sql.NullString
	var newRevision string
	if err := db.QueryRowContext(ctx, `SELECT owner_group_id::text,xmin::text FROM runtime_team_manifests
		WHERE tenant_id=$1 AND team_id=$2`, durableRuntimeTenant, manifest.ID).Scan(&newOwner, &newRevision); err != nil || newOwner.Valid || newRevision == oldRevision || newRevision == revokedRevision {
		t.Fatalf("stale owner or revision recurred after recreation: owner=%#v revision=%q old=%q revoked=%q error=%v", newOwner, newRevision, oldRevision, revokedRevision, err)
	}
}

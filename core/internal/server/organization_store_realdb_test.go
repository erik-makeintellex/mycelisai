package server

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"os"
	"testing"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Real PostgreSQL proof. MYCELIS_ORGANIZATION_STORE_TEST_DSN must point at a
// disposable database with 001_current_schema.sql installed. Skip is not proof.
func openOrganizationTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("MYCELIS_ORGANIZATION_STORE_TEST_DSN")
	if dsn == "" {
		t.Skip("requires disposable MYCELIS_ORGANIZATION_STORE_TEST_DSN")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	var table sql.NullString
	if err := db.QueryRow(`SELECT to_regclass('public.organizations')::text`).Scan(&table); err != nil || !table.Valid {
		t.Fatalf("DSN must have 001_current_schema.sql installed (organizations table): %v", err)
	}
	return db
}

func realDBOrganization(t *testing.T, db *sql.DB, scopeID string) OrganizationHomePayload {
	t.Helper()
	home := OrganizationHomePayload{OrganizationSummary: OrganizationSummary{
		ID: uuid.NewString(), Name: "Restart Proof " + uuid.NewString()[:8], Purpose: "Survive restart",
		StartMode: OrganizationStartModeEmpty,
	}, QAFixtureScopeID: scopeID}
	home = normalizeOrganizationHome(home)
	if _, err := NewOrganizationStore(db).Save(context.Background(), home); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = db.Exec(`DELETE FROM organizations WHERE id=$1::uuid`, home.ID) })
	return home
}

func TestOrganizationStoreRealDB_SurvivesNewAdminServer(t *testing.T) {
	db := openOrganizationTestDB(t)
	scopeID := uuid.NewString()
	home := realDBOrganization(t, db, scopeID)
	serverA := newTestServer(func(s *AdminServer) { s.DB = db; s.Organizations = NewOrganizationStore(db) })
	mux := setupMux(t, "PATCH /api/v1/organizations/{id}/ai-engine", serverA.handleUpdateOrganizationAIEngine)
	assertStatus(t, doAuthenticatedRequest(t, mux, http.MethodPatch, "/api/v1/organizations/"+home.ID+"/ai-engine", `{"profile_id":"high_reasoning"}`), http.StatusOK)
	routing := setupMux(t, "PATCH /api/v1/organizations/{id}/output-model-routing", serverA.handleUpdateOrganizationOutputModelRouting)
	assertStatus(t, doAuthenticatedRequest(t, routing, http.MethodPatch, "/api/v1/organizations/"+home.ID+"/output-model-routing", `{"routing_mode":"single_model","default_model_id":"qwen3:8b"}`), http.StatusOK)
	homeA := doAuthenticatedRequest(t, setupMux(t, "GET /api/v1/organizations/{id}/home", serverA.handleGetOrganizationHome), http.MethodGet, "/api/v1/organizations/"+home.ID+"/home", "")

	serverB := &AdminServer{DB: db, Organizations: NewOrganizationStore(db)} // restart: fresh process state
	homeB := doAuthenticatedRequest(t, setupMux(t, "GET /api/v1/organizations/{id}/home", serverB.handleGetOrganizationHome), http.MethodGet, "/api/v1/organizations/"+home.ID+"/home", "")
	assertStatus(t, homeB, http.StatusOK)
	if homeA.Body.String() != homeB.Body.String() {
		t.Fatalf("home changed across restart:\nA %s\nB %s", homeA.Body.String(), homeB.Body.String())
	}
	if scope, ok, err := serverB.organizationStore().QAFixtureScope(context.Background(), home.ID); err != nil || !ok || scope != scopeID {
		t.Fatalf("fixture scope after restart = %q %v %v", scope, ok, err)
	}
	var tenant, column string
	if err := db.QueryRow(`SELECT tenant_id, qa_fixture_scope_id FROM organizations WHERE id=$1::uuid`, home.ID).Scan(&tenant, &column); err != nil || tenant != organizationTenantID || column != scopeID {
		t.Fatalf("row tenant/scope = %q/%q err=%v", tenant, column, err)
	}
}

func TestOrganizationStoreRealDB_DuplicateAndConcurrentUpdates(t *testing.T) {
	db := openOrganizationTestDB(t)
	home := realDBOrganization(t, db, "")
	if _, err := NewOrganizationStore(db).Save(context.Background(), home); err != ErrOrganizationConflict {
		t.Fatalf("duplicate insert err = %v, want conflict", err)
	}
	runConcurrentOrganizationIncrements(t, NewOrganizationStore(db), home.ID, 20)
}

func TestOrganizationStoreRealDB_PurgeRollbackKeepsRowAndCommitRemovesExactID(t *testing.T) {
	db := openOrganizationTestDB(t)
	scopeID := uuid.NewString()
	claimed := realDBOrganization(t, db, scopeID)
	bystander := realDBOrganization(t, db, "")
	resources := []qaFixtureResource{{Kind: "organization", Ref: claimed.ID}}
	purge := func(commit bool) []string {
		tx, err := db.BeginTx(context.Background(), nil)
		if err != nil {
			t.Fatal(err)
		}
		removed, err := deleteQAFixtureDatabaseResources(context.Background(), tx, qaFixtureTenantID, scopeID, resources, map[string]int64{})
		if err != nil {
			t.Fatal(err)
		}
		if commit {
			if err := tx.Commit(); err != nil {
				t.Fatal(err)
			}
		} else if err := tx.Rollback(); err != nil {
			t.Fatal(err)
		}
		return removed
	}
	store := NewOrganizationStore(db)
	purge(false)
	if _, err := store.Get(context.Background(), claimed.ID); err != nil {
		t.Fatalf("rolled-back purge removed the row: %v", err)
	}
	if removed := purge(true); len(removed) != 1 || removed[0] != claimed.ID {
		t.Fatalf("removed = %v, want [%s]", removed, claimed.ID)
	}
	if _, err := store.Get(context.Background(), claimed.ID); err != ErrOrganizationNotFound {
		t.Fatalf("claimed organization still present: %v", err)
	}
	if _, err := store.Get(context.Background(), bystander.ID); err != nil {
		t.Fatalf("purge touched an unclaimed organization: %v", err)
	}
}

func TestOrganizationStoreRealDB_PurgeRefusesOrganizationOutsideScope(t *testing.T) {
	db := openOrganizationTestDB(t)
	scopeID := uuid.NewString()
	nonFixture := realDBOrganization(t, db, "")
	otherScope := realDBOrganization(t, db, uuid.NewString())
	store := NewOrganizationStore(db)
	for _, target := range []OrganizationHomePayload{nonFixture, otherScope} {
		tx, err := db.BeginTx(context.Background(), nil)
		if err != nil {
			t.Fatal(err)
		}
		deleted := map[string]int64{}
		removed, err := deleteQAFixtureDatabaseResources(context.Background(), tx, qaFixtureTenantID, scopeID,
			[]qaFixtureResource{{Kind: "organization", Ref: target.ID}}, deleted)
		_ = tx.Rollback()
		if !errors.Is(err, errQAFixtureResourceUnowned) || len(removed) != 0 || deleted["organizations"] != 0 {
			t.Fatalf("out-of-scope purge err=%v removed=%v deleted=%v", err, removed, deleted)
		}
		if _, err := store.Get(context.Background(), target.ID); err != nil {
			t.Fatalf("out-of-scope organization removed: %v", err)
		}
	}
}

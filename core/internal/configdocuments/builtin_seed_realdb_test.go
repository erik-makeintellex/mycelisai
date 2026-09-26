package configdocuments

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/mycelis/core/pkg/protocol"
)

const (
	seedRealDBNewID       = "aaa-seed-realdb-new"
	seedRealDBOperatorRef = "seed-realdb-operator"
)

type seedTableCounts struct{ documents, activations, history int }

func openSeedRealDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := strings.TrimSpace(os.Getenv("MYCELIS_CONFIG_DOCUMENT_SEED_TEST_DSN"))
	if dsn == "" {
		t.Skip("MYCELIS_CONFIG_DOCUMENT_SEED_TEST_DSN not set; real PostgreSQL seeding proof runs in the disposable fixture lane")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.PingContext(t.Context()); err != nil {
		t.Fatal(err)
	}
	var found sql.NullString
	if err := db.QueryRowContext(t.Context(), `SELECT to_regclass('config_document_activation_history')::text`).Scan(&found); err != nil || !found.Valid {
		t.Fatalf("canonical schema unavailable: %v %q", err, found.String)
	}
	return db
}

func seedRealDBIDs() []string {
	ids := []string{seedRealDBNewID}
	for _, name := range deliveryPostureFiles {
		ids = append(ids, strings.TrimSuffix(name, ".yaml"))
	}
	return ids
}

func cleanSeedRealDB(t *testing.T, db *sql.DB) {
	t.Helper()
	for _, statement := range []string{
		`DELETE FROM config_document_activation_history WHERE tenant_id = $1 AND document_id = ANY($2)`,
		`DELETE FROM config_document_activations WHERE tenant_id = $1 AND document_id = ANY($2)`,
		`DELETE FROM config_documents WHERE tenant_id = $1 AND document_id = ANY($2)`,
	} {
		if _, err := db.Exec(statement, BuiltInSeedTenant, pqArray(seedRealDBIDs())); err != nil {
			t.Fatalf("clean: %v", err)
		}
	}
}

func pqArray(values []string) string { return "{" + strings.Join(values, ",") + "}" }

func countSeedRows(t *testing.T, db *sql.DB) seedTableCounts {
	t.Helper()
	var counts seedTableCounts
	for table, dest := range map[string]*int{
		"config_documents": &counts.documents, "config_document_activations": &counts.activations,
		"config_document_activation_history": &counts.history,
	} {
		if err := db.QueryRow(`SELECT count(*) FROM `+table+` WHERE tenant_id = $1 AND document_id = ANY($2)`,
			BuiltInSeedTenant, pqArray(seedRealDBIDs())).Scan(dest); err != nil {
			t.Fatalf("count %s: %v", table, err)
		}
	}
	return counts
}

type operatorRowSnapshot struct {
	recordID, digest, createdBy, activeRecordID, activatedBy string
	activatedAt                                              time.Time
}

func snapshotOperatorRow(t *testing.T, db *sql.DB, documentID string) operatorRowSnapshot {
	t.Helper()
	var snap operatorRowSnapshot
	err := db.QueryRow(`
		SELECT d.record_id::text, d.digest, d.created_by, a.config_document_record_id::text, a.activated_by, a.activated_at
		FROM config_documents d
		JOIN config_document_activations a
		  ON a.tenant_id = d.tenant_id AND a.document_id = d.document_id AND a.scope_kind = d.scope_kind AND a.scope_ref = d.scope_ref
		WHERE d.tenant_id = $1 AND d.document_id = $2 AND d.scope_kind = 'operator' AND d.scope_ref = $3
	`, BuiltInSeedTenant, documentID, seedRealDBOperatorRef).Scan(
		&snap.recordID, &snap.digest, &snap.createdBy, &snap.activeRecordID, &snap.activatedBy, &snap.activatedAt)
	if err != nil {
		t.Fatalf("operator row snapshot: %v", err)
	}
	return snap
}

func TestSeedBuiltInRevisionsRealDB(t *testing.T) {
	db := openSeedRealDB(t)
	cleanSeedRealDB(t, db)
	t.Cleanup(func() { cleanSeedRealDB(t, db) })
	var foreign int
	if err := db.QueryRow(`SELECT count(*) FROM config_documents WHERE (scope_kind = 'built_in' OR source_kind = 'built_in') AND created_by <> $1`,
		BuiltInSeedActor).Scan(&foreign); err != nil || foreign != 0 {
		t.Fatalf("fixture DB has %d non-bootstrap built-in rows (%v); use a disposable database", foreign, err)
	}
	store := NewStore(db)
	ctx := t.Context()

	// (4) setup: an operator-scope copy of a posture with the same id, stored
	// and activated through the public path before any seeding.
	operatorID := strings.TrimSuffix(deliveryPostureFiles[3], ".yaml")
	operatorCopy, err := ParseDocument(readDeliveryPostureFile(t, deliveryPostureFiles[3]), "yaml")
	if err != nil {
		t.Fatal(err)
	}
	operatorCopy.Metadata.Scope = protocol.ConfigDocumentScope{Kind: protocol.ConfigDocumentScopeOperator, Ref: seedRealDBOperatorRef}
	operatorCopy.Metadata.Source = protocol.ConfigDocumentSource{Kind: protocol.ConfigDocumentSourceFile, Ref: "copies/" + deliveryPostureFiles[3]}
	operatorRecord, err := store.StoreRevision(ctx, BuiltInSeedTenant, "operator-seed-test", operatorCopy)
	if err != nil {
		t.Fatalf("store operator copy: %v", err)
	}
	if _, err := store.ActivateRevision(ctx, BuiltInSeedTenant, operatorRecord.RecordID, "operator-seed-test", "", ActivationActionActivate); err != nil {
		t.Fatalf("activate operator copy: %v", err)
	}
	operatorBefore := snapshotOperatorRow(t, db, operatorID)
	baseline := countSeedRows(t, db)

	// (1) first seed: exactly 4 revisions + 4 activations at built_in scope.
	result, err := store.SeedBuiltInRevisions(ctx, deliveryPostureDir)
	if err != nil {
		t.Fatalf("first seed: %v", err)
	}
	if result.Inserted != 4 || result.Activated != 4 {
		t.Fatalf("first seed result = %#v", result)
	}
	for _, seed := range loadSeeds(t, deliveryPostureDir) {
		var revisions int
		var createdBy, digest, activeRecord, activatedBy, recordID string
		err := db.QueryRow(`
			SELECT count(*) OVER (), d.record_id::text, d.created_by, d.digest, a.config_document_record_id::text, a.activated_by
			FROM config_documents d
			JOIN config_document_activations a
			  ON a.tenant_id = d.tenant_id AND a.document_id = d.document_id AND a.scope_kind = 'built_in' AND a.scope_ref = ''
			WHERE d.tenant_id = $1 AND d.document_id = $2 AND d.scope_kind = 'built_in' AND d.scope_ref = ''
		`, BuiltInSeedTenant, seed.Document.Metadata.ID).Scan(&revisions, &recordID, &createdBy, &digest, &activeRecord, &activatedBy)
		if err != nil {
			t.Fatalf("%s: seeded row: %v", seed.File, err)
		}
		if revisions != 1 || createdBy != BuiltInSeedActor || activatedBy != BuiltInSeedActor || digest != seed.Digest || activeRecord != recordID {
			t.Fatalf("%s: revisions=%d created_by=%q activated_by=%q digest ok=%v active=%v",
				seed.File, revisions, createdBy, activatedBy, digest == seed.Digest, activeRecord == recordID)
		}
		if _, err := store.GetActiveRevision(ctx, BuiltInSeedTenant, seed.Document.Kind, seed.Document.Metadata.ID, seed.Document.Metadata.Scope); err != nil {
			t.Fatalf("%s: active built-in revision not loadable: %v", seed.File, err)
		}
	}
	afterFirst := countSeedRows(t, db)
	if want := (seedTableCounts{baseline.documents + 4, baseline.activations + 4, baseline.history + 4}); afterFirst != want {
		t.Fatalf("after first seed counts = %+v, want %+v", afterFirst, want)
	}

	// (2) reseed writes nothing.
	result, err = store.SeedBuiltInRevisions(ctx, deliveryPostureDir)
	if err != nil || result.Inserted != 0 || result.Activated != 0 || result.Unchanged != 4 {
		t.Fatalf("reseed = %#v, %v", result, err)
	}
	if got := countSeedRows(t, db); got != afterFirst {
		t.Fatalf("reseed counts = %+v, want %+v", got, afterFirst)
	}

	// (3) a tampered batch fails closed with no partial writes: a new valid
	// file sorts first and would insert, then the last file changes content
	// without a version bump.
	dir := seedDir(t, deliveryPostureFiles...)
	newName := seedRealDBNewID + ".yaml"
	newRaw := strings.NewReplacer(
		"id: "+operatorID, "id: "+seedRealDBNewID,
		"ref: "+BuiltInSeedSourcePrefix+deliveryPostureFiles[3], "ref: "+BuiltInSeedSourcePrefix+newName,
	).Replace(string(readDeliveryPostureFile(t, deliveryPostureFiles[3])))
	if err := os.WriteFile(filepath.Join(dir, newName), []byte(newRaw), 0o600); err != nil {
		t.Fatal(err)
	}
	tampered := "delivery-posture-product-delivery-team.yaml"
	rewriteSeed(t, dir, tampered, "primary_deliverable: ", "primary_deliverable: Tampered after release. ")
	if seeds := loadSeeds(t, dir); len(seeds) != 5 {
		t.Fatalf("tampered batch must still validate so it reaches the database, got %d files", len(seeds))
	}
	if _, err := store.SeedBuiltInRevisions(ctx, dir); err == nil || !strings.Contains(err.Error(), "bump metadata.version") {
		t.Fatalf("tampered batch error = %v, want fail-closed version conflict", err)
	}
	if got := countSeedRows(t, db); got != afterFirst {
		t.Fatalf("tampered batch wrote rows: %+v, want %+v", got, afterFirst)
	}

	// (4) the operator-scope copy is untouched by every seed.
	if after := snapshotOperatorRow(t, db, operatorID); after != operatorBefore {
		t.Fatalf("operator row changed: before %+v after %+v", operatorBefore, after)
	}
}

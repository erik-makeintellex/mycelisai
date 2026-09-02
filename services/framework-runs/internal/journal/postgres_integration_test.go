//go:build postgres_integration

package journal_test

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/mycelis/framework-runs/internal/controller"
	"github.com/mycelis/framework-runs/internal/journal"
	"github.com/mycelis/framework-runs/internal/protocol"
)

var disposableDatabase = regexp.MustCompile(`^mycelis_framework_runs_test_[a-z0-9_]+$`)

type postgresExecutor struct{}

func (postgresExecutor) Apply(_ context.Context, command journal.Command) (protocol.ExecutorOutcome, error) {
	return protocol.ExecutorOutcome{Status: protocol.StatusCompleted, Result: &protocol.Result{
		Summary: "Candidate ready.", FinishedAt: time.Now().UTC(),
		Outputs: []protocol.Output{{ID: "output-1", Kind: "document",
			URI: "candidate://" + command.RunID + "/output-1", ContentType: "application/json",
			SizeBytes: 2, SHA256: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}},
	}}, nil
}

func TestPostgresCurrentSchemaDurabilityAndImmutability(t *testing.T) {
	ctx := context.Background()
	databaseURL, connect := newDisposableDatabase(t, ctx)
	repository, err := journal.OpenPostgres(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	service := controller.New(repository, postgresExecutor{})
	request := postgresRequest("run-pg-1")
	run, replay, err := service.Create(ctx, request)
	if err != nil || replay || run.Version != 1 {
		t.Fatalf("create: %#v %v %v", run, replay, err)
	}
	if _, err := service.DispatchOnce(ctx, "pg-test-worker"); err != nil {
		t.Fatal(err)
	}
	repository.Close()

	// A fully compatible nonempty schema is an idempotent no-op and state survives restart.
	repository, err = journal.OpenPostgres(ctx, databaseURL)
	if err != nil {
		t.Fatalf("compatible reopen: %v", err)
	}
	defer repository.Close()
	snapshot, err := repository.Get(ctx, run.RunID)
	if err != nil || snapshot.Status != protocol.StatusCompleted || snapshot.Version != 2 {
		t.Fatalf("restart snapshot: %#v %v", snapshot, err)
	}
	events, err := repository.Events(ctx, run.RunID, 0)
	if err != nil || len(events) != 2 || events[1].Sequence != 2 {
		t.Fatalf("events: %#v %v", events, err)
	}
	connection, err := connect(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close(ctx)
	if _, err := connection.Exec(ctx, `DELETE FROM run_events WHERE run_id=$1`, run.RunID); err == nil {
		t.Fatal("immutable journal event was deleted")
	}
	if _, err := connection.Exec(ctx, `UPDATE candidate_manifests SET size_bytes=3 WHERE run_id=$1`, run.RunID); err == nil {
		t.Fatal("immutable candidate manifest was updated")
	}
}

func TestPostgresRejectsPartialSchemaBeforeBaseline(t *testing.T) {
	ctx := context.Background()
	databaseURL, connect := newDisposableDatabase(t, ctx)
	connection, err := connect(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := connection.Exec(ctx, `CREATE TABLE unrelated_partial(id integer primary key)`); err != nil {
		t.Fatal(err)
	}
	connection.Close(ctx)
	if repository, err := journal.OpenPostgres(ctx, databaseURL); !errors.Is(err, journal.ErrSchemaMismatch) {
		if repository != nil {
			repository.Close()
		}
		t.Fatalf("partial schema error = %v", err)
	}
	connection, err = connect(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close(ctx)
	var baselineTables int
	if err := connection.QueryRow(ctx, `SELECT count(*) FROM pg_tables WHERE schemaname='public' AND tablename='runs'`).Scan(&baselineTables); err != nil {
		t.Fatal(err)
	}
	if baselineTables != 0 {
		t.Fatal("baseline executed against partial schema")
	}
}

func TestPostgresRejectsDamagedCompatibleSchema(t *testing.T) {
	ctx := context.Background()
	databaseURL, connect := newDisposableDatabase(t, ctx)
	repository, err := journal.OpenPostgres(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	repository.Close()
	connection, err := connect(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := connection.Exec(ctx, `DROP INDEX run_commands_claim_order`); err != nil {
		connection.Close(ctx)
		t.Fatal(err)
	}
	connection.Close(ctx)
	if repository, err := journal.OpenPostgres(ctx, databaseURL); !errors.Is(err, journal.ErrSchemaMismatch) {
		if repository != nil {
			repository.Close()
		}
		t.Fatalf("damaged schema error = %v", err)
	}
}

func newDisposableDatabase(t *testing.T, ctx context.Context) (string, func(context.Context) (*pgx.Conn, error)) {
	t.Helper()
	adminURL := os.Getenv("FRAMEWORK_RUNS_TEST_ADMIN_DSN")
	if adminURL == "" {
		t.Skip("FRAMEWORK_RUNS_TEST_ADMIN_DSN is required for tagged Postgres proof")
	}
	adminConfig, err := pgx.ParseConfig(adminURL)
	if err != nil {
		t.Fatal(err)
	}
	adminConfig.Database = "postgres"
	admin, err := pgx.ConnectConfig(ctx, adminConfig)
	if err != nil {
		t.Fatal(err)
	}
	database := fmt.Sprintf("mycelis_framework_runs_test_%d_%d", os.Getpid(), time.Now().UnixNano())
	if !disposableDatabase.MatchString(database) {
		t.Fatalf("refusing unsafe disposable database name %q", database)
	}
	t.Logf("disposable database: %s", database)
	identifier := pgx.Identifier{database}.Sanitize()
	if _, err := admin.Exec(ctx, "CREATE DATABASE "+identifier); err != nil {
		admin.Close(ctx)
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if !disposableDatabase.MatchString(database) {
			t.Errorf("refusing unsafe cleanup database %q", database)
			return
		}
		_, dropErr := admin.Exec(context.Background(), "DROP DATABASE "+identifier+" WITH (FORCE)")
		if dropErr != nil {
			t.Errorf("drop disposable database: %v", dropErr)
		}
		var remaining int
		if countErr := admin.QueryRow(context.Background(), `SELECT count(*) FROM pg_database WHERE datname=$1`, database).Scan(&remaining); countErr != nil {
			t.Errorf("verify disposable database cleanup: %v", countErr)
		} else {
			t.Logf("disposable database %s cleanup after-count: %d", database, remaining)
		}
		admin.Close(context.Background())
	})
	targetConfig := adminConfig.Copy()
	targetConfig.Database = database
	connect := func(ctx context.Context) (*pgx.Conn, error) { return pgx.ConnectConfig(ctx, targetConfig.Copy()) }
	proof, err := connect(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var connectedDatabase string
	if err := proof.QueryRow(ctx, `SELECT current_database()`).Scan(&connectedDatabase); err != nil {
		proof.Close(ctx)
		t.Fatal(err)
	}
	proof.Close(ctx)
	if connectedDatabase != database || !disposableDatabase.MatchString(connectedDatabase) {
		t.Fatalf("refusing target database %q; expected %q", connectedDatabase, database)
	}
	targetURL, err := url.Parse(adminURL)
	if err != nil {
		t.Fatal(err)
	}
	targetURL.Path = "/" + database
	targetURL.RawPath = ""
	return targetURL.String(), connect
}

func postgresRequest(runID string) protocol.CreateRequest {
	return protocol.CreateRequest{RunID: runID, Intent: "Produce candidate", Correlation: protocol.Correlation{
		RunID: runID, IntentProofID: "proof-pg", ExecutionContractID: "contract-pg",
		WorkItemID: "work-pg", IdempotencyKey: "idem:" + runID, SourceKind: "system",
		SourceChannel: "test.pg", PayloadKind: "command", GraphRevision: "graph-pg",
	}}
}

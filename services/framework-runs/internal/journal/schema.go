package journal

import (
	"context"
	"fmt"
	"sort"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mycelis/framework-runs/migrations"
)

var expectedTables = []string{
	"candidate_manifests", "framework_runs_schema_metadata", "run_approvals",
	"run_commands", "run_events", "runs",
}

var expectedConstraints = []string{
	"candidate_manifests_candidate_uri_key", "candidate_manifests_immutable",
	"candidate_manifests_pkey", "candidate_manifests_run_id_fkey",
	"candidate_manifests_sha256_check", "candidate_manifests_size_bytes_check", "candidate_uri_scoped",
	"framework_runs_schema_metadata_pkey", "framework_runs_schema_metadata_schema_contract_check",
	"framework_runs_schema_metadata_schema_version_check", "framework_runs_schema_metadata_singleton_check",
	"run_approvals_pkey", "run_approvals_run_id_fkey", "run_approvals_state_check",
	"run_approvals_decision_command_id_fkey", "run_approvals_decision_digest_check",
	"run_approvals_request_identity",
	"run_commands_attempts_check", "run_commands_expected_version_check", "run_commands_kind_check",
	"run_commands_lease_generation_check", "run_commands_payload_digest_check", "run_commands_pkey",
	"run_commands_run_id_fkey", "run_commands_state_check", "run_events_digest_hex",
	"run_events_event_id_key", "run_events_identity", "run_events_immutable", "run_events_kind_status",
	"run_events_pkey", "run_events_run_id_fkey", "run_events_sequence_check", "run_events_sequence_identity",
	"runs_idempotency_key_key", "runs_pending_command_id_fkey",
	"runs_next_sequence_check", "runs_pkey", "runs_request_digest_hex", "runs_request_identity",
	"runs_snapshot_identity", "runs_snapshot_status", "runs_snapshot_version", "runs_status_exact", "runs_version_check",
}

var expectedIndexes = []string{
	"run_approvals_one_pending_per_run", "run_commands_claim_order", "run_commands_one_pending_per_run",
}

var expectedColumns = map[string][]string{
	"framework_runs_schema_metadata": {"installed_at", "schema_contract", "schema_version", "singleton"},
	"runs": {"created_at", "idempotency_key", "next_sequence", "pending_command_id", "request_digest",
		"request_json", "run_id", "snapshot_json", "status", "updated_at", "version"},
	"run_events": {"created_at", "event_id", "kind", "payload_digest", "payload_json", "run_id", "sequence", "status"},
	"run_commands": {"approval_id", "attempts", "available_at", "command_id", "created_at", "expected_version",
		"kind", "last_error", "lease_generation", "lease_owner", "lease_token", "lease_until", "payload_digest",
		"payload_json", "receipt_json", "run_id", "state", "updated_at"},
	"run_approvals": {"actor_id", "approval_id", "created_at", "decided_at", "decision_command_id",
		"decision_digest", "request_json", "run_id", "state"},
	"candidate_manifests": {"candidate_uri", "content_type", "created_at", "kind", "metadata", "name",
		"output_id", "run_id", "sha256", "size_bytes"},
}

func EnsureCurrentSchema(ctx context.Context, pool *pgxpool.Pool) error {
	if pool == nil {
		return fmt.Errorf("%w: database pool is nil", ErrSchemaMismatch)
	}
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended('mycelis-framework-runs-schema', 0))`); err != nil {
		return err
	}
	tables, err := publicTables(ctx, tx)
	if err != nil {
		return err
	}
	if len(tables) == 0 {
		if _, err := tx.Exec(ctx, migrations.CurrentSchema); err != nil {
			return fmt.Errorf("install framework Runs schema: %w", err)
		}
	} else if !equalStrings(tables, expectedTables) {
		return fmt.Errorf("%w: unexpected public tables %v", ErrSchemaMismatch, tables)
	}
	if err := verifySchema(ctx, tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func publicTables(ctx context.Context, tx pgx.Tx) ([]string, error) {
	rows, err := tx.Query(ctx, `
SELECT tablename FROM pg_catalog.pg_tables
WHERE schemaname='public' ORDER BY tablename`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}
	return tables, rows.Err()
}

func verifySchema(ctx context.Context, tx pgx.Tx) error {
	var version int
	var contract string
	if err := tx.QueryRow(ctx, `SELECT schema_version, schema_contract FROM framework_runs_schema_metadata WHERE singleton`).Scan(&version, &contract); err != nil {
		return fmt.Errorf("%w: metadata: %v", ErrSchemaMismatch, err)
	}
	if version != 1 || contract != "framework-runs-v1" {
		return fmt.Errorf("%w: unsupported contract", ErrSchemaMismatch)
	}
	if err := verifyColumns(ctx, tx); err != nil {
		return err
	}
	rows, err := tx.Query(ctx, `
SELECT conname FROM pg_constraint c
JOIN pg_class t ON t.oid=c.conrelid
JOIN pg_namespace n ON n.oid=t.relnamespace
WHERE n.nspname='public' AND t.relname IN
('framework_runs_schema_metadata','runs','run_events','run_commands','run_approvals','candidate_manifests')
UNION ALL
SELECT tgname FROM pg_trigger g
JOIN pg_class t ON t.oid=g.tgrelid
JOIN pg_namespace n ON n.oid=t.relnamespace
WHERE n.nspname='public' AND NOT g.tgisinternal
ORDER BY 1`)
	if err != nil {
		return err
	}
	defer rows.Close()
	available := map[string]bool{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return err
		}
		available[name] = true
	}
	for _, name := range expectedConstraints {
		if !available[name] {
			return fmt.Errorf("%w: missing constraint or trigger %s", ErrSchemaMismatch, name)
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return verifyIndexesAndFunction(ctx, tx)
}

func verifyColumns(ctx context.Context, tx pgx.Tx) error {
	rows, err := tx.Query(ctx, `
SELECT table_name,column_name FROM information_schema.columns
WHERE table_schema='public' ORDER BY table_name,column_name`)
	if err != nil {
		return err
	}
	defer rows.Close()
	actual := map[string][]string{}
	for rows.Next() {
		var table, column string
		if err := rows.Scan(&table, &column); err != nil {
			return err
		}
		actual[table] = append(actual[table], column)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for table, columns := range expectedColumns {
		if !equalStrings(actual[table], columns) {
			return fmt.Errorf("%w: unexpected columns for %s", ErrSchemaMismatch, table)
		}
	}
	return nil
}

func verifyIndexesAndFunction(ctx context.Context, tx pgx.Tx) error {
	rows, err := tx.Query(ctx, `
SELECT indexname FROM pg_catalog.pg_indexes
WHERE schemaname='public' AND indexname = ANY($1) ORDER BY indexname`, expectedIndexes)
	if err != nil {
		return err
	}
	var indexes []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			rows.Close()
			return err
		}
		indexes = append(indexes, name)
	}
	rows.Close()
	if !equalStrings(indexes, expectedIndexes) {
		return fmt.Errorf("%w: missing required index", ErrSchemaMismatch)
	}
	var functions int
	if err := tx.QueryRow(ctx, `
SELECT count(*) FROM pg_proc p JOIN pg_namespace n ON n.oid=p.pronamespace
WHERE n.nspname='public' AND p.proname='reject_framework_runs_immutable_change'`).Scan(&functions); err != nil {
		return err
	}
	if functions != 1 {
		return fmt.Errorf("%w: immutable trigger function missing", ErrSchemaMismatch)
	}
	return nil
}

func equalStrings(left, right []string) bool {
	left = append([]string(nil), left...)
	right = append([]string(nil), right...)
	sort.Strings(left)
	sort.Strings(right)
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

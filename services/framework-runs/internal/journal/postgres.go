package journal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mycelis/framework-runs/internal/protocol"
)

type PostgresRepository struct{ pool *pgxpool.Pool }

func OpenPostgres(ctx context.Context, databaseURL string) (*PostgresRepository, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse framework Runs database configuration: %w", err)
	}
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("open framework Runs database: %w", err)
	}
	repository := &PostgresRepository{pool: pool}
	if err := EnsureCurrentSchema(ctx, pool); err != nil {
		pool.Close()
		return nil, err
	}
	return repository, nil
}

func (repository *PostgresRepository) Close() {
	if repository != nil && repository.pool != nil {
		repository.pool.Close()
	}
}

func (repository *PostgresRepository) Health(ctx context.Context) error {
	if repository == nil || repository.pool == nil {
		return errors.New("database pool is unavailable")
	}
	return repository.pool.Ping(ctx)
}

func (repository *PostgresRepository) Create(
	ctx context.Context,
	request protocol.CreateRequest,
	digest string,
	start Command,
	maxRuns int,
) (protocol.Run, bool, error) {
	tx, err := repository.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return protocol.Run{}, false, err
	}
	defer tx.Rollback(ctx)
	if run, found, replay, err := existingCreate(ctx, tx, request.RunID, request.Correlation.IdempotencyKey, digest); err != nil || found {
		return run, replay, err
	}
	if maxRuns > 0 {
		var count int
		if err := tx.QueryRow(ctx, `SELECT COUNT(*) FROM runs`).Scan(&count); err != nil {
			return protocol.Run{}, false, err
		}
		if count >= maxRuns {
			return protocol.Run{}, false, ErrCapacity
		}
	}
	now := start.Receipt.CreatedAt.UTC()
	run := protocol.Run{
		RunID: request.RunID, Correlation: request.Correlation,
		Status: protocol.StatusAccepted, Version: 1,
		CreatedAt: now, UpdatedAt: now,
		Metadata: map[string]any{"execution_authority": "mycelis_core", "storage": "durable_journal"},
	}
	requestJSON, _ := json.Marshal(request)
	snapshotJSON, _ := json.Marshal(run)
	_, err = tx.Exec(ctx, `
INSERT INTO runs (
 run_id,idempotency_key,request_digest,request_json,snapshot_json,status,
 version,next_sequence,pending_command_id,created_at,updated_at
) VALUES ($1,$2,$3,$4,$5,$6,1,2,$7,$8,$8)`,
		run.RunID, request.Correlation.IdempotencyKey, digest, requestJSON, snapshotJSON,
		run.Status, start.CommandID, now)
	if err != nil {
		if isConstraintConflict(err) {
			return protocol.Run{}, false, ErrRunConflict
		}
		return protocol.Run{}, false, err
	}
	accepted := AcceptedEvent(run)
	if err := insertEvent(ctx, tx, accepted); err != nil {
		return protocol.Run{}, false, err
	}
	if err := insertCommand(ctx, tx, start); err != nil {
		return protocol.Run{}, false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return protocol.Run{}, false, err
	}
	return run, false, nil
}

func existingCreate(ctx context.Context, tx pgx.Tx, runID, idempotencyKey, digest string) (protocol.Run, bool, bool, error) {
	var storedRunID, storedKey, storedDigest string
	var snapshotJSON []byte
	err := tx.QueryRow(ctx, `
SELECT run_id,idempotency_key,request_digest,snapshot_json
FROM runs WHERE run_id=$1 OR idempotency_key=$2
ORDER BY CASE WHEN run_id=$1 THEN 0 ELSE 1 END LIMIT 1`, runID, idempotencyKey).
		Scan(&storedRunID, &storedKey, &storedDigest, &snapshotJSON)
	if errors.Is(err, pgx.ErrNoRows) {
		return protocol.Run{}, false, false, nil
	}
	if err != nil {
		return protocol.Run{}, false, false, err
	}
	if storedRunID != runID || storedKey != idempotencyKey || storedDigest != digest {
		return protocol.Run{}, true, false, ErrRunConflict
	}
	var run protocol.Run
	if err := json.Unmarshal(snapshotJSON, &run); err != nil {
		return protocol.Run{}, true, false, err
	}
	return run, true, true, nil
}

func (repository *PostgresRepository) Get(ctx context.Context, runID string) (protocol.Run, error) {
	var snapshotJSON []byte
	if err := repository.pool.QueryRow(ctx, `SELECT snapshot_json FROM runs WHERE run_id=$1`, runID).Scan(&snapshotJSON); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return protocol.Run{}, ErrNotFound
		}
		return protocol.Run{}, err
	}
	var run protocol.Run
	if err := json.Unmarshal(snapshotJSON, &run); err != nil {
		return protocol.Run{}, err
	}
	return run, nil
}

func (repository *PostgresRepository) Events(ctx context.Context, runID string, after uint64) ([]protocol.Event, error) {
	var next uint64
	if err := repository.pool.QueryRow(ctx, `SELECT next_sequence FROM runs WHERE run_id=$1`, runID).Scan(&next); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if after >= next && after != next-1 {
		return nil, ErrCursorGap
	}
	rows, err := repository.pool.Query(ctx, `
SELECT payload_json FROM run_events WHERE run_id=$1 AND sequence>$2 ORDER BY sequence`, runID, after)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var events []protocol.Event
	for rows.Next() {
		var raw []byte
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}
		var event protocol.Event
		if err := json.Unmarshal(raw, &event); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, rows.Err()
}

func insertEvent(ctx context.Context, tx pgx.Tx, event protocol.Event) error {
	if err := protocol.ValidateEvent(event); err != nil {
		return err
	}
	raw, _ := json.Marshal(event)
	digest, _ := protocol.Digest(event)
	_, err := tx.Exec(ctx, `
INSERT INTO run_events(run_id,sequence,event_id,payload_digest,payload_json,kind,status,created_at)
VALUES($1,$2,$3,$4,$5,$6,$7,$8)`, event.RunID, event.Sequence, event.EventID,
		digest, raw, event.Kind, event.Status, event.Timestamp)
	return err
}

func insertCommand(ctx context.Context, tx pgx.Tx, command Command) error {
	payload, _ := json.Marshal(command)
	receipt, _ := json.Marshal(command.Receipt)
	_, err := tx.Exec(ctx, `
INSERT INTO run_commands(
 command_id,run_id,kind,payload_digest,payload_json,expected_version,approval_id,
 state,attempts,available_at,receipt_json,created_at,updated_at
) VALUES($1,$2,$3,$4,$5,$6,NULLIF($7,''),$8,0,$9,$10,$9,$9)`,
		command.CommandID, command.RunID, command.Kind, command.Digest, payload,
		command.ExpectedVersion, command.ApprovalID, CommandPending,
		command.AvailableAt, receipt)
	return err
}

func isConstraintConflict(err error) bool {
	var postgresError *pgconn.PgError
	return errors.As(err, &postgresError) && postgresError.Code == "23505"
}

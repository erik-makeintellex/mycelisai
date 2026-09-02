package journal

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/mycelis/framework-runs/internal/protocol"
)

func (repository *PostgresRepository) SubmitControl(ctx context.Context, command Command) (protocol.ControlReceipt, bool, error) {
	tx, err := repository.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return protocol.ControlReceipt{}, false, err
	}
	defer tx.Rollback(ctx)
	var existingDigest, existingRunID, existingKind string
	var existingReceipt []byte
	err = tx.QueryRow(ctx, `SELECT payload_digest,run_id,kind,receipt_json FROM run_commands WHERE command_id=$1`, command.CommandID).
		Scan(&existingDigest, &existingRunID, &existingKind, &existingReceipt)
	if err == nil {
		if existingDigest != command.Digest || existingRunID != command.RunID || existingKind != command.Kind {
			return protocol.ControlReceipt{}, false, ErrCommandConflict
		}
		var receipt protocol.ControlReceipt
		if err := json.Unmarshal(existingReceipt, &receipt); err != nil {
			return protocol.ControlReceipt{}, false, err
		}
		return receipt, true, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return protocol.ControlReceipt{}, false, err
	}
	var snapshotJSON []byte
	var pending *string
	if err := tx.QueryRow(ctx, `SELECT snapshot_json,pending_command_id FROM runs WHERE run_id=$1 FOR UPDATE`, command.RunID).
		Scan(&snapshotJSON, &pending); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return protocol.ControlReceipt{}, false, ErrNotFound
		}
		return protocol.ControlReceipt{}, false, err
	}
	var run protocol.Run
	if err := json.Unmarshal(snapshotJSON, &run); err != nil {
		return protocol.ControlReceipt{}, false, err
	}
	if run.Version != command.ExpectedVersion {
		return protocol.ControlReceipt{}, false, ErrVersionConflict
	}
	if pending != nil {
		return protocol.ControlReceipt{}, false, ErrCommandConflict
	}
	if err := validateControlAgainstRun(run, command); err != nil {
		return protocol.ControlReceipt{}, false, err
	}
	if err := insertCommand(ctx, tx, command); err != nil {
		if isConstraintConflict(err) {
			return protocol.ControlReceipt{}, false, ErrCommandConflict
		}
		return protocol.ControlReceipt{}, false, err
	}
	if _, err := tx.Exec(ctx, `UPDATE runs SET pending_command_id=$2,updated_at=$3 WHERE run_id=$1`, command.RunID, command.CommandID, command.Receipt.UpdatedAt); err != nil {
		return protocol.ControlReceipt{}, false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return protocol.ControlReceipt{}, false, err
	}
	return command.Receipt, false, nil
}

func (repository *PostgresRepository) Claim(ctx context.Context, owner string, now time.Time, duration time.Duration) (*Lease, error) {
	token, err := randomToken("lease:")
	if err != nil {
		return nil, err
	}
	var payload []byte
	var epoch uint64
	var attempts int
	err = repository.pool.QueryRow(ctx, `
WITH candidate AS (
 SELECT command_id FROM run_commands
 WHERE (state='pending' AND available_at<=$1) OR (state='leased' AND lease_until<$1)
 ORDER BY created_at FOR UPDATE SKIP LOCKED LIMIT 1
)
UPDATE run_commands c SET state='leased',attempts=attempts+1,lease_owner=$2,
 lease_token=$3,lease_generation=lease_generation+1,lease_until=$4,updated_at=$1
FROM candidate WHERE c.command_id=candidate.command_id
RETURNING c.payload_json,c.lease_generation,c.attempts`, now, owner, token, now.Add(duration)).Scan(&payload, &epoch, &attempts)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var command Command
	if err := json.Unmarshal(payload, &command); err != nil {
		return nil, err
	}
	command.State = CommandLeased
	command.LeaseOwner = owner
	command.LeaseToken = token
	command.LeaseGeneration = epoch
	command.Attempts = attempts
	command.LeaseUntil = now.Add(duration)
	return &Lease{Command: command, Owner: owner, Token: token, Epoch: epoch}, nil
}

func (repository *PostgresRepository) Complete(ctx context.Context, lease Lease, outcome protocol.ExecutorOutcome, now time.Time) (protocol.Run, error) {
	tx, err := repository.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return protocol.Run{}, err
	}
	defer tx.Rollback(ctx)
	command, err := lockedLeasedCommand(ctx, tx, lease)
	if err != nil {
		return protocol.Run{}, err
	}
	var snapshotJSON []byte
	var nextSequence uint64
	var pending *string
	if err := tx.QueryRow(ctx, `SELECT snapshot_json,next_sequence,pending_command_id FROM runs WHERE run_id=$1 FOR UPDATE`, command.RunID).
		Scan(&snapshotJSON, &nextSequence, &pending); err != nil {
		return protocol.Run{}, err
	}
	if pending == nil || *pending != command.CommandID {
		return protocol.Run{}, ErrConflict
	}
	var run protocol.Run
	if err := json.Unmarshal(snapshotJSON, &run); err != nil {
		return protocol.Run{}, err
	}
	event, err := ApplyOutcome(&run, command, outcome, now)
	if err != nil {
		return protocol.Run{}, err
	}
	event.Sequence = nextSequence
	if err := insertEvent(ctx, tx, event); err != nil {
		return protocol.Run{}, err
	}
	if err := persistApprovalAndCandidates(ctx, tx, command, run, now); err != nil {
		return protocol.Run{}, err
	}
	snapshotJSON, _ = json.Marshal(run)
	if _, err := tx.Exec(ctx, `
UPDATE runs SET snapshot_json=$2,status=$3,version=$4,next_sequence=$5,
 pending_command_id=NULL,updated_at=$6 WHERE run_id=$1`, run.RunID, snapshotJSON,
		run.Status, run.Version, nextSequence+1, now); err != nil {
		return protocol.Run{}, err
	}
	command.Receipt.State = CommandApplied
	command.Receipt.Version = run.Version
	command.Receipt.UpdatedAt = now.UTC()
	receiptJSON, _ := json.Marshal(command.Receipt)
	if _, err := tx.Exec(ctx, `
UPDATE run_commands SET state='applied',receipt_json=$2,lease_owner=NULL,lease_token=NULL,
 lease_until=NULL,updated_at=$3 WHERE command_id=$1`, command.CommandID, receiptJSON, now); err != nil {
		return protocol.Run{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return protocol.Run{}, err
	}
	return run, nil
}

func (repository *PostgresRepository) Fail(ctx context.Context, lease Lease, failure protocol.Error, now time.Time) error {
	tx, err := repository.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	command, err := lockedLeasedCommand(ctx, tx, lease)
	if err != nil {
		return err
	}
	command.Receipt.State = CommandFailed
	command.Receipt.Error = protocol.Clone(&failure)
	command.Receipt.UpdatedAt = now.UTC()
	receiptJSON, _ := json.Marshal(command.Receipt)
	tag, err := tx.Exec(ctx, `
UPDATE runs SET pending_command_id=NULL,updated_at=$3
WHERE run_id=$1 AND pending_command_id=$2`, command.RunID, command.CommandID, now)
	if err != nil || tag.RowsAffected() != 1 {
		if err != nil {
			return err
		}
		return ErrConflict
	}
	_, err = tx.Exec(ctx, `
UPDATE run_commands SET state='failed',receipt_json=$2,last_error=$3,
 lease_owner=NULL,lease_token=NULL,lease_until=NULL,updated_at=$4
WHERE command_id=$1`, command.CommandID, receiptJSON, failure.Message, now)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (repository *PostgresRepository) Retry(ctx context.Context, lease Lease, availableAt time.Time, message string) error {
	tag, err := repository.pool.Exec(ctx, `
UPDATE run_commands SET state='pending',available_at=$5,lease_owner=NULL,lease_token=NULL,
 lease_until=NULL,last_error=$6,updated_at=$5
WHERE command_id=$1 AND state='leased' AND lease_owner=$2 AND lease_token=$3 AND lease_generation=$4`,
		lease.Command.CommandID, lease.Owner, lease.Token, lease.Epoch, availableAt, message)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return ErrLeaseLost
	}
	return nil
}

func lockedLeasedCommand(ctx context.Context, tx pgx.Tx, lease Lease) (Command, error) {
	var payload []byte
	var owner, token string
	var epoch uint64
	var state string
	if err := tx.QueryRow(ctx, `
SELECT payload_json,state,COALESCE(lease_owner,''),COALESCE(lease_token,''),lease_generation
FROM run_commands WHERE command_id=$1 FOR UPDATE`, lease.Command.CommandID).
		Scan(&payload, &state, &owner, &token, &epoch); err != nil {
		return Command{}, err
	}
	if state != CommandLeased || owner != lease.Owner || token != lease.Token || epoch != lease.Epoch {
		return Command{}, ErrLeaseLost
	}
	var command Command
	if err := json.Unmarshal(payload, &command); err != nil {
		return Command{}, err
	}
	return command, nil
}

func persistApprovalAndCandidates(ctx context.Context, tx pgx.Tx, command Command, run protocol.Run, now time.Time) error {
	if command.Kind == "approve" || command.Kind == "deny" {
		state := "approved"
		if command.Kind == "deny" {
			state = "denied"
		}
		tag, err := tx.Exec(ctx, `
UPDATE run_approvals SET state=$2,decision_command_id=$3,decision_digest=$4,
 actor_id=$5,decided_at=$6 WHERE approval_id=$1 AND state='pending'`,
			command.ApprovalID, state, command.CommandID, command.Digest, command.ActorID, now)
		if err != nil {
			return err
		}
		if tag.RowsAffected() != 1 {
			return ErrApprovalNotFound
		}
	}
	if run.Approval != nil {
		raw, _ := json.Marshal(run.Approval)
		if _, err := tx.Exec(ctx, `
INSERT INTO run_approvals(approval_id,run_id,request_json,state,created_at)
VALUES($1,$2,$3,'pending',$4)`, run.Approval.ID, run.RunID, raw, now); err != nil {
			return err
		}
	}
	if run.Result == nil {
		return nil
	}
	for _, output := range run.Result.Outputs {
		if output.URI == "" {
			continue
		}
		metadata, _ := json.Marshal(output.Metadata)
		if _, err := tx.Exec(ctx, `
INSERT INTO candidate_manifests(
 run_id,output_id,candidate_uri,kind,name,content_type,size_bytes,sha256,metadata,created_at
) VALUES($1,$2,$3,$4,NULLIF($5,''),$6,$7,$8,$9,$10)`, run.RunID, output.ID,
			output.URI, output.Kind, output.Name, output.ContentType, output.SizeBytes,
			output.SHA256, metadata, now); err != nil {
			return err
		}
	}
	return nil
}

func randomToken(prefix string) (string, error) {
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return prefix + hex.EncodeToString(buffer), nil
}

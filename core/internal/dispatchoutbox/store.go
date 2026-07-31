package dispatchoutbox

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	StatusStaged    = "staged"
	StatusPending   = "pending"
	StatusExecuting = "executing"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
)

var ErrUnavailable = errors.New("dispatch outbox unavailable")

type Item struct {
	ID             string
	IdempotencyKey string
	DispatchKind   string
	Status         string
	RunID          string
	IntentProofID  string
	ContractID     string
	TeamID         string
	WorkItemID     string
	SourceKind     string
	SourceChannel  string
	PayloadKind    string
	Payload        json.RawMessage
	AttemptCount   int
	AvailableAt    time.Time
	LeaseUntil     *time.Time
	LastError      string
	Recovery       json.RawMessage
}

type Store struct{ db *sql.DB }

func NewStore(db *sql.DB) *Store { return &Store{db: db} }

func (s *Store) EnqueueTx(ctx context.Context, tx *sql.Tx, item Item) (Item, error) {
	if s == nil || s.db == nil || tx == nil {
		return Item{}, ErrUnavailable
	}
	if err := validateItem(item); err != nil {
		return Item{}, err
	}
	if item.Status == "" {
		item.Status = StatusStaged
	}
	if len(item.Recovery) == 0 {
		item.Recovery = json.RawMessage(`{}`)
	}
	if item.AvailableAt.IsZero() {
		item.AvailableAt = time.Now().UTC()
	}
	_, err := tx.ExecContext(ctx, `
INSERT INTO execution_dispatch_outbox (
    id, idempotency_key, dispatch_kind, status, run_id, intent_proof_id,
    contract_id, team_id, work_item_id, source_kind, source_channel,
    payload_kind, payload, available_at, recovery
) VALUES ($1,$2,$3,$4,$5,$6,NULLIF($7,'')::uuid,NULLIF($8,''),NULLIF($9,''),$10,$11,$12,$13::jsonb,$14,$15::jsonb)
ON CONFLICT (idempotency_key) DO NOTHING`,
		item.ID, item.IdempotencyKey, item.DispatchKind, item.Status, item.RunID,
		item.IntentProofID, item.ContractID, item.TeamID, item.WorkItemID,
		item.SourceKind, item.SourceChannel, item.PayloadKind, string(item.Payload),
		item.AvailableAt, string(item.Recovery),
	)
	if err != nil {
		return Item{}, err
	}
	return item, nil
}

func (s *Store) Activate(ctx context.Context, idempotencyKey string) error {
	if s == nil || s.db == nil {
		return ErrUnavailable
	}
	_, err := s.db.ExecContext(ctx, `
UPDATE execution_dispatch_outbox
SET status=$2, available_at=NOW(), updated_at=NOW()
WHERE idempotency_key=$1 AND status=$3`, idempotencyKey, StatusPending, StatusStaged)
	return err
}

func (s *Store) UpdatePayloadAndActivate(ctx context.Context, idempotencyKey string, payload json.RawMessage) error {
	if s == nil || s.db == nil {
		return ErrUnavailable
	}
	if !json.Valid(payload) || containsRawSecret(payload) {
		return errors.New("dispatch payload must be valid JSON without raw secrets")
	}
	_, err := s.db.ExecContext(ctx, `
UPDATE execution_dispatch_outbox
SET payload=$2::jsonb, status=$3, available_at=NOW(), updated_at=NOW()
WHERE idempotency_key=$1 AND status=$4`, idempotencyKey, string(payload), StatusPending, StatusStaged)
	return err
}

func (s *Store) ClaimNext(ctx context.Context, lease time.Duration) (*Item, error) {
	if s == nil || s.db == nil {
		return nil, ErrUnavailable
	}
	if lease <= 0 {
		lease = 30 * time.Second
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	row := tx.QueryRowContext(ctx, `
WITH candidate AS (
    SELECT id FROM execution_dispatch_outbox
    WHERE available_at <= NOW()
      AND (status=$1 OR (status=$2 AND lease_until < NOW()) OR (status=$3 AND created_at < NOW() - INTERVAL '5 seconds'))
    ORDER BY created_at ASC
    FOR UPDATE SKIP LOCKED
    LIMIT 1
)
UPDATE execution_dispatch_outbox d
SET status=$2, attempt_count=attempt_count+1, lease_until=NOW()+($4 * INTERVAL '1 millisecond'), updated_at=NOW()
FROM candidate
WHERE d.id=candidate.id
RETURNING d.id::text,d.idempotency_key,d.dispatch_kind,d.status,d.run_id,d.intent_proof_id::text,
          COALESCE(d.contract_id::text,''),COALESCE(d.team_id,''),COALESCE(d.work_item_id,''),
          d.source_kind,d.source_channel,d.payload_kind,d.payload,d.attempt_count,d.available_at,
          d.lease_until,COALESCE(d.last_error,''),d.recovery`,
		StatusPending, StatusExecuting, StatusStaged, lease.Milliseconds())
	item, err := scanItem(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, tx.Commit()
	}
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *Store) MarkCompleted(ctx context.Context, id string) error {
	return s.finish(ctx, id, StatusCompleted, "", 0)
}

func (s *Store) MarkRetry(ctx context.Context, id string, cause error, delay time.Duration) error {
	return s.finish(ctx, id, StatusPending, errorText(cause), delay)
}

func (s *Store) MarkFailed(ctx context.Context, id string, cause error) error {
	return s.finish(ctx, id, StatusFailed, errorText(cause), 0)
}

func (s *Store) finish(ctx context.Context, id, status, lastError string, delay time.Duration) error {
	if s == nil || s.db == nil {
		return ErrUnavailable
	}
	completed := status == StatusCompleted || status == StatusFailed
	_, err := s.db.ExecContext(ctx, `
UPDATE execution_dispatch_outbox
SET status=$2, last_error=NULLIF($3,''), lease_until=NULL,
    available_at=NOW()+($4 * INTERVAL '1 millisecond'), updated_at=NOW(),
    published_at=CASE WHEN $2=$5 THEN COALESCE(published_at,NOW()) ELSE published_at END,
    completed_at=CASE WHEN $6 THEN NOW() ELSE NULL END
WHERE id=$1`, id, status, lastError, delay.Milliseconds(), StatusCompleted, completed)
	return err
}

type scanner interface{ Scan(...any) error }

func scanItem(row scanner) (Item, error) {
	var item Item
	err := row.Scan(&item.ID, &item.IdempotencyKey, &item.DispatchKind, &item.Status,
		&item.RunID, &item.IntentProofID, &item.ContractID, &item.TeamID, &item.WorkItemID,
		&item.SourceKind, &item.SourceChannel, &item.PayloadKind, &item.Payload,
		&item.AttemptCount, &item.AvailableAt, &item.LeaseUntil, &item.LastError, &item.Recovery)
	return item, err
}

func validateItem(item Item) error {
	for label, value := range map[string]string{
		"id": item.ID, "idempotency_key": item.IdempotencyKey, "dispatch_kind": item.DispatchKind,
		"run_id": item.RunID, "intent_proof_id": item.IntentProofID, "source_kind": item.SourceKind,
		"source_channel": item.SourceChannel, "payload_kind": item.PayloadKind,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", label)
		}
	}
	if !json.Valid(item.Payload) {
		return errors.New("payload must be valid JSON")
	}
	if containsRawSecret(item.Payload) {
		return errors.New("dispatch payload contains a raw secret; use a secret reference")
	}
	return nil
}

func containsRawSecret(raw json.RawMessage) bool {
	var value any
	if json.Unmarshal(raw, &value) != nil {
		return false
	}
	return walkSecret(value)
}

func walkSecret(value any) bool {
	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			name := strings.ToLower(strings.TrimSpace(key))
			if !strings.HasSuffix(name, "_ref") && (name == "password" || name == "token" || name == "api_key" || name == "secret") && strings.TrimSpace(fmt.Sprint(child)) != "" {
				return true
			}
			if walkSecret(child) {
				return true
			}
		}
	case []any:
		for _, child := range typed {
			if walkSecret(child) {
				return true
			}
		}
	}
	return false
}

func errorText(err error) string {
	if err == nil {
		return ""
	}
	return strings.TrimSpace(err.Error())
}

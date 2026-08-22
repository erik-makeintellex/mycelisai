package signal

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"
)

// StreamEvent is one durable operator-facing SSE payload.
type StreamEvent struct {
	Sequence  int64
	ID        string
	Payload   string
	CreatedAt time.Time
}

// ReplayGap tells clients that bounded replay omitted older events.
type ReplayGap struct {
	Reason         string
	RequestedAfter int64
	FirstReplayed  int64
	Omitted        int64
	Replayable     bool
}

type ReplayBatch struct {
	Events []StreamEvent
	Gap    *ReplayGap
}

type eventPersistence interface {
	Append(context.Context, string) (StreamEvent, error)
	Replay(context.Context, int64, int) (ReplayBatch, error)
}

type EventStore struct {
	db *sql.DB
}

func NewEventStore(db *sql.DB) *EventStore {
	return &EventStore{db: db}
}

func (s *EventStore) Append(ctx context.Context, payload string) (StreamEvent, error) {
	if s == nil || s.db == nil {
		return StreamEvent{}, fmt.Errorf("SSE event store unavailable")
	}
	payload = sanitizeOperatorPayload(payload)
	event := StreamEvent{Payload: payload}
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO operator_sse_events (payload)
		VALUES ($1)
		RETURNING sequence, created_at`, payload,
	).Scan(&event.Sequence, &event.CreatedAt)
	if err != nil {
		return StreamEvent{}, fmt.Errorf("persist SSE event: %w", err)
	}
	event.ID = strconv.FormatInt(event.Sequence, 10)
	return event, nil
}

func (s *EventStore) Replay(ctx context.Context, after int64, limit int) (ReplayBatch, error) {
	if s == nil || s.db == nil {
		return ReplayBatch{}, fmt.Errorf("SSE event store unavailable")
	}
	if limit < 1 {
		limit = defaultReplayLimit
	}
	rows, err := s.db.QueryContext(ctx, `
		WITH eligible AS (
			SELECT sequence, payload, created_at, COUNT(*) OVER () AS total_count
			FROM operator_sse_events
			WHERE sequence > $1
		), recent AS (
			SELECT * FROM eligible ORDER BY sequence DESC LIMIT $2
		)
		SELECT sequence, payload, created_at, total_count
		FROM recent ORDER BY sequence ASC`, after, limit)
	if err != nil {
		return ReplayBatch{}, fmt.Errorf("query SSE replay: %w", err)
	}
	defer rows.Close()

	batch := ReplayBatch{Events: make([]StreamEvent, 0, limit)}
	var total int64
	for rows.Next() {
		var event StreamEvent
		if err := rows.Scan(&event.Sequence, &event.Payload, &event.CreatedAt, &total); err != nil {
			return ReplayBatch{}, fmt.Errorf("scan SSE replay: %w", err)
		}
		event.ID = strconv.FormatInt(event.Sequence, 10)
		batch.Events = append(batch.Events, event)
	}
	if err := rows.Err(); err != nil {
		return ReplayBatch{}, fmt.Errorf("iterate SSE replay: %w", err)
	}
	if total > int64(len(batch.Events)) && len(batch.Events) > 0 {
		batch.Gap = &ReplayGap{
			Reason: "replay_limit", RequestedAfter: after,
			FirstReplayed: batch.Events[0].Sequence,
			Omitted:       total - int64(len(batch.Events)), Replayable: true,
		}
	}
	return batch, nil
}

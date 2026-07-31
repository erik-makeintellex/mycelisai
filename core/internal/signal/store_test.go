package signal

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestEventStoreAppendReturnsStableSequenceID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Now().UTC()
	mock.ExpectQuery("INSERT INTO operator_sse_events").
		WithArgs(`{"type":"status"}`).
		WillReturnRows(sqlmock.NewRows([]string{"sequence", "created_at"}).AddRow(int64(42), now))

	event, err := NewEventStore(db).Append(context.Background(), `{"type":"status"}`)
	if err != nil {
		t.Fatalf("append: %v", err)
	}
	if event.Sequence != 42 || event.ID != "42" || event.Payload != `{"type":"status"}` {
		t.Fatalf("event = %#v", event)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestEventStoreReplayReturnsLatestWindowInSequence(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Now().UTC()
	mock.ExpectQuery("WITH eligible AS").WithArgs(int64(1), 2).WillReturnRows(
		sqlmock.NewRows([]string{"sequence", "payload", "created_at", "total_count"}).
			AddRow(int64(5), `{"value":5}`, now, int64(4)).
			AddRow(int64(6), `{"value":6}`, now, int64(4)),
	)

	batch, err := NewEventStore(db).Replay(context.Background(), 1, 2)
	if err != nil {
		t.Fatalf("replay: %v", err)
	}
	if len(batch.Events) != 2 || batch.Events[0].ID != "5" || batch.Events[1].ID != "6" {
		t.Fatalf("events = %#v", batch.Events)
	}
	if batch.Gap == nil || batch.Gap.Omitted != 2 || batch.Gap.FirstReplayed != 5 {
		t.Fatalf("gap = %#v", batch.Gap)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestEventStoreExcludesLastEventID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectQuery("WITH eligible AS").WithArgs(int64(9), 10).
		WillReturnRows(sqlmock.NewRows([]string{"sequence", "payload", "created_at", "total_count"}))

	batch, err := NewEventStore(db).Replay(context.Background(), 9, 10)
	if err != nil {
		t.Fatalf("replay: %v", err)
	}
	if len(batch.Events) != 0 || batch.Gap != nil {
		t.Fatalf("batch = %#v", batch)
	}
}

func TestEventStoreDatabaseFailuresAreReturned(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectQuery("INSERT INTO operator_sse_events").
		WithArgs("payload").WillReturnError(errors.New("insert unavailable"))
	if _, err := NewEventStore(db).Append(context.Background(), "payload"); err == nil {
		t.Fatal("expected append failure")
	}
	mock.ExpectQuery("WITH eligible AS").WithArgs(int64(2), 5).
		WillReturnError(errors.New("query unavailable"))
	if _, err := NewEventStore(db).Replay(context.Background(), 2, 5); err == nil {
		t.Fatal("expected replay failure")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSanitizeOperatorPayloadRedactsCredentialFields(t *testing.T) {
	payload := sanitizeOperatorPayload(`{"type":"status","authorization":"Bearer private","nested":{"api_key":"private"}}`)
	if strings.Contains(payload, "Bearer private") || strings.Contains(payload, `"api_key":"private"`) {
		t.Fatalf("credential values remained in payload: %s", payload)
	}
	if !strings.Contains(payload, `"authorization":"[REDACTED]"`) {
		t.Fatalf("redacted authorization missing: %s", payload)
	}
}

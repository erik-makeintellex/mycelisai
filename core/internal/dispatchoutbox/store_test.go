package dispatchoutbox

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestEnqueueTxStoresOneIdempotentApprovedDispatch(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectBegin()
	tx, err := db.BeginTx(t.Context(), nil)
	if err != nil {
		t.Fatal(err)
	}
	item := testItem()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO execution_dispatch_outbox")).
		WithArgs(item.ID, item.IdempotencyKey, item.DispatchKind, StatusStaged, item.RunID,
			item.IntentProofID, item.ContractID, item.TeamID, item.WorkItemID,
			item.SourceKind, item.SourceChannel, item.PayloadKind, string(item.Payload),
			sqlmock.AnyArg(), "{}").
		WillReturnResult(sqlmock.NewResult(1, 1))
	if _, err := NewStore(db).EnqueueTx(t.Context(), tx, item); err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	mock.ExpectRollback()
	_ = tx.Rollback()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestEnqueueTxRejectsRawSecrets(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectBegin()
	tx, err := db.BeginTx(t.Context(), nil)
	if err != nil {
		t.Fatal(err)
	}
	item := testItem()
	item.Payload = []byte(`{"password":"do-not-store"}`)
	if _, err := NewStore(db).EnqueueTx(context.Background(), tx, item); err == nil {
		t.Fatal("expected raw secret rejection")
	}
	mock.ExpectRollback()
	_ = tx.Rollback()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestFrameworkCreateIntentIsExactAndNonClaimable(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	item := testItem()
	item.IdempotencyKey = "framework-run-create:" + item.IdempotencyKey
	mock.ExpectBegin()
	tx, _ := db.BeginTx(t.Context(), nil)
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO execution_dispatch_outbox")).
		WithArgs(item.ID, item.IdempotencyKey, DispatchKindFrameworkRunCreate,
			StatusAwaitingHandler, item.RunID, item.IntentProofID, item.ContractID,
			item.TeamID, item.WorkItemID, item.SourceKind, item.SourceChannel,
			item.PayloadKind, string(item.Payload), sqlmock.AnyArg(), "{}").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(item.ID))
	created, err := NewStore(db).EnqueueFrameworkCreateTx(t.Context(), tx, item)
	if err != nil || !created {
		t.Fatalf("framework create intent = %v, %v", created, err)
	}
	mock.ExpectRollback()
	_ = tx.Rollback()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestClaimNextReclaimsExpiredLease(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	item := testItem()
	now := time.Now().UTC()
	mock.ExpectBegin()
	rows := sqlmock.NewRows([]string{
		"id", "idempotency_key", "dispatch_kind", "status", "run_id", "intent_proof_id",
		"contract_id", "team_id", "work_item_id", "source_kind", "source_channel",
		"payload_kind", "payload", "attempt_count", "available_at", "lease_until", "last_error", "recovery",
	}).AddRow(item.ID, item.IdempotencyKey, item.DispatchKind, StatusExecuting, item.RunID,
		item.IntentProofID, item.ContractID, item.TeamID, item.WorkItemID, item.SourceKind,
		item.SourceChannel, item.PayloadKind, item.Payload, 2, now, now.Add(time.Minute), "", []byte(`{}`))
	mock.ExpectQuery("WITH candidate AS").
		WithArgs(StatusPending, StatusExecuting, StatusStaged, int64(30000)).
		WillReturnRows(rows)
	mock.ExpectCommit()
	claimed, err := NewStore(db).ClaimNext(t.Context(), 30*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if claimed == nil || claimed.AttemptCount != 2 || claimed.Status != StatusExecuting {
		t.Fatalf("unexpected claim: %#v", claimed)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestClaimNextReturnsNilWhenQueueEmpty(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectQuery("WITH candidate AS").
		WithArgs(StatusPending, StatusExecuting, StatusStaged, int64(30000)).
		WillReturnError(errors.New("query failed"))
	mock.ExpectRollback()
	if _, err := NewStore(db).ClaimNext(t.Context(), 30*time.Second); err == nil {
		t.Fatal("expected query failure")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func testItem() Item {
	return Item{
		ID:             "11111111-1111-1111-1111-111111111111",
		IdempotencyKey: "confirm:proof-1",
		DispatchKind:   "confirmed_action",
		RunID:          "22222222-2222-2222-2222-222222222222",
		IntentProofID:  "33333333-3333-3333-3333-333333333333",
		ContractID:     "44444444-4444-4444-4444-444444444444",
		TeamID:         "delivery-team",
		WorkItemID:     "55555555-5555-5555-5555-555555555555",
		SourceKind:     "web_api",
		SourceChannel:  "api.intent.confirm-action",
		PayloadKind:    "command",
		Payload:        []byte(`{"scope":{"planned_tool_calls":[]},"secret_ref":"env:KEY"}`),
	}
}

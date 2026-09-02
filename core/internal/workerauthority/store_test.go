package workerauthority

import (
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

const testDigest = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func TestStageBindingExactReplayAndConflict(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	binding := testBinding()
	store := NewStore(db)

	mock.ExpectBegin()
	tx, _ := db.BeginTx(t.Context(), nil)
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO worker_run_bindings")).
		WithArgs(binding.RunID, binding.IntentProofID, binding.ExecutionContractID,
			binding.TeamID, binding.WorkItemID, binding.OutcomeID, binding.IdempotencyKey,
			binding.GraphRevision, binding.SourceKind, binding.SourceChannel,
			binding.PayloadKind, binding.RequestDigest).
		WillReturnRows(sqlmock.NewRows([]string{"run_id"}).AddRow(binding.RunID))
	if created, err := store.StageBindingTx(t.Context(), tx, binding); err != nil || !created {
		t.Fatalf("create binding = %v, %v", created, err)
	}
	mock.ExpectRollback()
	_ = tx.Rollback()

	for _, conflict := range []bool{false, true} {
		mock.ExpectBegin()
		tx, _ = db.BeginTx(t.Context(), nil)
		mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO worker_run_bindings")).
			WithArgs(binding.RunID, binding.IntentProofID, binding.ExecutionContractID,
				binding.TeamID, binding.WorkItemID, binding.OutcomeID, binding.IdempotencyKey,
				binding.GraphRevision, binding.SourceKind, binding.SourceChannel,
				binding.PayloadKind, binding.RequestDigest).
			WillReturnRows(sqlmock.NewRows([]string{"run_id"}))
		digest := binding.RequestDigest
		if conflict {
			digest = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
		}
		mock.ExpectQuery("SELECT run_id::text").WithArgs(binding.RunID, binding.IdempotencyKey).
			WillReturnRows(bindingRow(binding, digest))
		created, err := store.StageBindingTx(t.Context(), tx, binding)
		if conflict && (!errors.Is(err, ErrConflict) || created) {
			t.Fatalf("conflicting replay = %v, %v", created, err)
		}
		if !conflict && (err != nil || created) {
			t.Fatalf("exact replay = %v, %v", created, err)
		}
		mock.ExpectRollback()
		_ = tx.Rollback()
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestProjectEventClaimsProjectsAndAdvancesCursorAtomically(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	store := NewStore(db)
	receipt := testReceipt()
	binding := testBinding()
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT state").WithArgs(receipt.RunID).
		WillReturnRows(authorityRow(binding, BindingBound, 4, 0, 0))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO worker_event_receipts")).
		WithArgs(receipt.ID, receipt.RunID, receipt.EventID, receipt.Sequence,
			receipt.EventKind, receipt.ServiceVersion, receipt.PayloadDigest, string(receipt.NormalizedPayload)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO mission_events").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE worker_event_receipts").
		WithArgs(receipt.RunID, receipt.Sequence, "11111111-1111-1111-1111-111111111115").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE worker_run_bindings").
		WithArgs(receipt.RunID, receipt.Sequence, receipt.ServiceVersion,
			receipt.RunStatus, receipt.ExpectedCursorVersion, int64(0)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	projected, err := store.ProjectEvent(t.Context(), receipt, func(tx *sql.Tx) (string, error) {
		_, err := tx.ExecContext(t.Context(), "INSERT INTO mission_events VALUES ()")
		return "11111111-1111-1111-1111-111111111115", err
	})
	if err != nil || !projected {
		t.Fatalf("project event = %v, %v", projected, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestProjectEventExactDuplicateNoOpAndConflictFailClosed(t *testing.T) {
	for _, conflict := range []bool{false, true} {
		db, mock, _ := sqlmock.New()
		store := NewStore(db)
		receipt := testReceipt()
		binding := testBinding()
		mock.ExpectBegin()
		mock.ExpectQuery("SELECT state").WithArgs(receipt.RunID).
			WillReturnRows(authorityRow(binding, BindingBound, 4, 1, 1))
		digest := receipt.PayloadDigest
		if conflict {
			digest = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
		}
		mock.ExpectQuery("SELECT event_id,sequence,payload_digest").
			WithArgs(receipt.RunID, receipt.EventID, receipt.Sequence).
			WillReturnRows(sqlmock.NewRows([]string{"event_id", "sequence", "payload_digest"}).
				AddRow(receipt.EventID, receipt.Sequence, digest))
		if conflict {
			mock.ExpectRollback()
		} else {
			mock.ExpectCommit()
		}
		projected, err := store.ProjectEvent(t.Context(), receipt, func(*sql.Tx) (string, error) {
			t.Fatal("duplicate reached projection")
			return "", nil
		})
		if conflict && !errors.Is(err, ErrConflict) {
			t.Fatalf("conflict error = %v", err)
		}
		if !conflict && (err != nil || projected) {
			t.Fatalf("exact duplicate = %v, %v", projected, err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
		db.Close()
	}
}

func TestProjectEventGapAndCursorCASFailClosed(t *testing.T) {
	for name, testCase := range map[string]struct {
		sequence, cursor int64
		want             error
	}{
		"gap":       {3, 0, ErrEventGap},
		"stale CAS": {1, 9, ErrStaleVersion},
	} {
		t.Run(name, func(t *testing.T) {
			db, mock, _ := sqlmock.New()
			defer db.Close()
			receipt := testReceipt()
			receipt.Sequence, receipt.ExpectedCursorVersion = testCase.sequence, testCase.cursor
			mock.ExpectBegin()
			mock.ExpectQuery("SELECT state").WithArgs(receipt.RunID).
				WillReturnRows(authorityRow(testBinding(), BindingBound, 4, 0, 0))
			mock.ExpectRollback()
			_, err := NewStore(db).ProjectEvent(t.Context(), receipt, func(*sql.Tx) (string, error) {
				return "", nil
			})
			if !errors.Is(err, testCase.want) {
				t.Fatalf("error = %v, want %v", err, testCase.want)
			}
		})
	}
}

func TestProjectEventRejectsInconsistentKindStatusBeforeDatabaseUse(t *testing.T) {
	receipt := testReceipt()
	receipt.RunStatus = "running"
	if _, err := NewStore(&sql.DB{}).ProjectEvent(t.Context(), receipt, func(*sql.Tx) (string, error) {
		return "", nil
	}); err == nil || err.Error() != "worker event kind and status are inconsistent" {
		t.Fatalf("kind/status validation error = %v", err)
	}
}

func TestProjectEventRejectsDigestThatDoesNotAttestPayload(t *testing.T) {
	receipt := testReceipt()
	receipt.PayloadDigest = testDigest
	if _, err := NewStore(&sql.DB{}).ProjectEvent(t.Context(), receipt, func(*sql.Tx) (string, error) {
		return "", nil
	}); err == nil || err.Error() != "worker event digest and normalized payload are invalid" {
		t.Fatalf("payload digest validation error = %v", err)
	}
}

func TestProjectEventAcceptsOrderedReplayOlderThanObservedSnapshot(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	receipt := testReceipt()
	receipt.ServiceVersion = 1
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT state").WithArgs(receipt.RunID).
		WillReturnRows(authorityRow(testBinding(), BindingBound, 2, 0, 0))
	mock.ExpectExec("INSERT INTO worker_event_receipts").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE worker_event_receipts").
		WithArgs(receipt.RunID, receipt.Sequence, "11111111-1111-1111-1111-111111111115").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE worker_run_bindings").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	projected, err := NewStore(db).ProjectEvent(t.Context(), receipt, func(*sql.Tx) (string, error) {
		return "11111111-1111-1111-1111-111111111115", nil
	})
	if err != nil || !projected {
		t.Fatalf("ordered historical replay = %v, %v", projected, err)
	}
}

func TestProjectEventRejectsMissingProjectionIdentityAndRollsBack(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	receipt := testReceipt()
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT state").WithArgs(receipt.RunID).
		WillReturnRows(authorityRow(testBinding(), BindingBound, 4, 0, 0))
	mock.ExpectExec("INSERT INTO worker_event_receipts").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectRollback()
	projected, err := NewStore(db).ProjectEvent(t.Context(), receipt, func(*sql.Tx) (string, error) {
		return "", nil
	})
	if err == nil || projected {
		t.Fatalf("missing projection id = %v, %v", projected, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestTransitionCommandRejectsIllegalStateChangeBeforeDatabaseUse(t *testing.T) {
	err := NewStore(&sql.DB{}).TransitionCommand(
		t.Context(), "command-1", CommandAcknowledged, CommandPending, "", 0,
	)
	if err == nil || err.Error() != "invalid worker command transition" {
		t.Fatalf("illegal transition error = %v", err)
	}
}

func TestBindingApprovalAndCommandCASRejectStaleVersions(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	store := NewStore(db)
	for _, call := range []func() error{
		func() error { return store.BindRun(t.Context(), testBinding().RunID, 1, 7) },
		func() error { return store.DecideApproval(t.Context(), "approval-1", "approve", "operator", "ok", 3) },
		func() error {
			return store.TransitionCommand(t.Context(), "command-1", CommandStaged, CommandPending, "", 2)
		},
	} {
		mock.ExpectExec("UPDATE worker_").WillReturnResult(sqlmock.NewResult(0, 0))
		if err := call(); !errors.Is(err, ErrStaleVersion) {
			t.Fatalf("CAS error = %v", err)
		}
	}
}

func testBinding() Binding {
	return Binding{
		RunID: "11111111-1111-1111-1111-111111111111", IntentProofID: "11111111-1111-1111-1111-111111111112",
		ExecutionContractID: "11111111-1111-1111-1111-111111111113", TeamID: "delivery-team",
		WorkItemID: "11111111-1111-1111-1111-111111111114", OutcomeID: "outcome-1",
		IdempotencyKey: "confirm-action:proof-1", GraphRevision: "graph-v1", SourceKind: "web_api",
		SourceChannel: "api.intent.confirm-action", PayloadKind: "command", RequestDigest: testDigest,
	}
}

func testReceipt() EventReceipt {
	b := testBinding()
	payload := []byte(`{"kind":"accepted"}`)
	digest := sha256.Sum256(payload)
	return EventReceipt{
		ID: "21111111-1111-1111-1111-111111111111", RunID: b.RunID, EventID: "event-1",
		Sequence: 1, ServiceVersion: 5, ExpectedCursorVersion: 0, EventKind: "accepted",
		RunStatus: "accepted", PayloadDigest: fmt.Sprintf("%x", digest), NormalizedPayload: payload,
		Correlation: Correlation{RunID: b.RunID, IntentProofID: b.IntentProofID,
			ExecutionContractID: b.ExecutionContractID, TeamID: b.TeamID, WorkItemID: b.WorkItemID,
			OutcomeID: b.OutcomeID, IdempotencyKey: b.IdempotencyKey, GraphRevision: b.GraphRevision,
			SourceKind: b.SourceKind, SourceChannel: b.SourceChannel, PayloadKind: b.PayloadKind},
	}
}

func bindingRow(b Binding, digest string) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"run_id", "proof", "contract", "team", "work", "outcome", "key", "graph", "source_kind", "source_channel", "payload_kind", "digest"}).
		AddRow(b.RunID, b.IntentProofID, b.ExecutionContractID, b.TeamID, b.WorkItemID,
			b.OutcomeID, b.IdempotencyKey, b.GraphRevision, b.SourceKind, b.SourceChannel, b.PayloadKind, digest)
}

func authorityRow(b Binding, state string, serviceVersion, lastSequence, cursorVersion int64) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"state", "proof", "contract", "team", "work", "outcome", "key", "graph", "source_kind", "source_channel", "payload_kind", "digest", "service_version", "last_sequence", "cursor_version"}).
		AddRow(state, b.IntentProofID, b.ExecutionContractID, b.TeamID, b.WorkItemID,
			b.OutcomeID, b.IdempotencyKey, b.GraphRevision, b.SourceKind, b.SourceChannel,
			b.PayloadKind, b.RequestDigest, serviceVersion, lastSequence, cursorVersion)
}

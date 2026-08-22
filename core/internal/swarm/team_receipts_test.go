package swarm

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestPostgresCommandReceiptStoreAcceptsOnce(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sql mock: %v", err)
	}
	defer db.Close()

	store := NewPostgresCommandReceiptStore(db)
	correlation := teamCommandCorrelation{
		TeamID:         "delivery-team",
		WorkItemID:     "11111111-1111-1111-1111-111111111111",
		IdempotencyKey: "delivery-command-1",
	}
	query := regexp.QuoteMeta("INSERT INTO team_signal_receipts")
	for _, rowsAffected := range []int64{1, 0} {
		mock.ExpectExec(query).
			WithArgs(sqlmock.AnyArg(), correlation.TeamID, correlation.WorkItemID, correlation.IdempotencyKey, "swarm.team.delivery-team.internal.command").
			WillReturnResult(sqlmock.NewResult(0, rowsAffected))
	}

	accepted, err := store.AcceptCommand(context.Background(), correlation, "swarm.team.delivery-team.internal.command")
	if err != nil {
		t.Fatalf("accept command: %v", err)
	}
	if !accepted {
		t.Fatal("expected first receipt claim to be accepted")
	}

	accepted, err = store.AcceptCommand(context.Background(), correlation, "swarm.team.delivery-team.internal.command")
	if err != nil {
		t.Fatalf("replay command: %v", err)
	}
	if accepted {
		t.Fatal("expected duplicate receipt claim to be rejected")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("receipt expectations: %v", err)
	}
}

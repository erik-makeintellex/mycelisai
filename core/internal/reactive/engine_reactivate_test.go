package reactive

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestReactivateFromDB_NilDB(t *testing.T) {
	e := New(nil, nil)
	err := e.ReactivateFromDB(context.Background())
	if err != nil {
		t.Errorf("expected nil error when db is nil, got: %v", err)
	}
}

func TestReactivateFromDB_NilNC(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	e := New(nil, nil)
	e.SetDB(db)
	err = e.ReactivateFromDB(context.Background())
	if err != nil {
		t.Errorf("expected nil error when nc is nil, got: %v", err)
	}
}

func TestReactivateFromDB_WithActiveProfiles(t *testing.T) {
	srv, nc := startTestNATS(t)
	defer srv.Shutdown()
	defer nc.Close()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	handler := func(profileID, topic string, msg []byte) {}
	e := New(nc, handler)
	e.SetDB(db)
	defer e.Close()

	subs1, _ := json.Marshal([]ProfileSubscription{
		{Topic: "swarm.team.research.*"},
	})
	subs2, _ := json.Marshal([]ProfileSubscription{
		{Topic: "swarm.team.ops.>"},
		{Topic: "swarm.events.mission.*"},
	})

	mock.ExpectQuery("SELECT .+ FROM mission_profiles").
		WillReturnRows(sqlmock.NewRows([]string{"id", "subscriptions"}).
			AddRow("prof-aaa", subs1).
			AddRow("prof-bbb", subs2))

	err = e.ReactivateFromDB(context.Background())
	if err != nil {
		t.Fatalf("ReactivateFromDB error: %v", err)
	}
	if e.ActiveSubscriptionCount() != 3 {
		t.Errorf("expected 3 active subs after reactivation, got %d", e.ActiveSubscriptionCount())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet DB expectations: %v", err)
	}
}

func TestReactivateFromDB_EmptySubscriptions(t *testing.T) {
	srv, nc := startTestNATS(t)
	defer srv.Shutdown()
	defer nc.Close()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	e := New(nc, nil)
	e.SetDB(db)
	defer e.Close()

	emptySubs, _ := json.Marshal([]ProfileSubscription{})

	mock.ExpectQuery("SELECT .+ FROM mission_profiles").
		WillReturnRows(sqlmock.NewRows([]string{"id", "subscriptions"}).
			AddRow("prof-empty", emptySubs))

	err = e.ReactivateFromDB(context.Background())
	if err != nil {
		t.Fatalf("ReactivateFromDB error: %v", err)
	}
	if e.ActiveSubscriptionCount() != 0 {
		t.Errorf("expected 0 subs for empty subscriptions array, got %d", e.ActiveSubscriptionCount())
	}
}

func TestReactivateFromDB_InvalidJSON(t *testing.T) {
	srv, nc := startTestNATS(t)
	defer srv.Shutdown()
	defer nc.Close()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	e := New(nc, nil)
	e.SetDB(db)
	defer e.Close()

	mock.ExpectQuery("SELECT .+ FROM mission_profiles").
		WillReturnRows(sqlmock.NewRows([]string{"id", "subscriptions"}).
			AddRow("prof-bad", []byte(`not valid json`)))

	err = e.ReactivateFromDB(context.Background())
	if err != nil {
		t.Fatalf("ReactivateFromDB should not error on bad JSON (skips row), got: %v", err)
	}
	if e.ActiveSubscriptionCount() != 0 {
		t.Errorf("expected 0 subs for invalid JSON, got %d", e.ActiveSubscriptionCount())
	}
}

func TestReactivateFromDB_QueryError(t *testing.T) {
	srv, nc := startTestNATS(t)
	defer srv.Shutdown()
	defer nc.Close()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	e := New(nc, nil)
	e.SetDB(db)
	defer e.Close()

	mock.ExpectQuery("SELECT .+ FROM mission_profiles").
		WillReturnError(sql.ErrConnDone)

	err = e.ReactivateFromDB(context.Background())
	if err == nil {
		t.Error("expected error from query failure")
	}
}

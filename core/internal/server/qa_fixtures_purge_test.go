package server

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestDeleteQAFixtureGroupRequiresTenantMatch(t *testing.T) {
	withDatabase, mock := withDB(t)
	s := newTestServer(withDatabase)
	groupID := "22222222-2222-2222-2222-222222222222"
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM collaboration_groups WHERE id=\\$1::uuid AND tenant_id=\\$2").
		WithArgs(groupID, qaFixtureTenantID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	tx, err := s.getDB().BeginTx(t.Context(), nil)
	if err != nil {
		t.Fatal(err)
	}
	deleted := make(map[string]int64)
	err = deleteQAFixtureDatabaseResource(t.Context(), tx, qaFixtureTenantID, qaFixtureResource{
		Kind: "group",
		Ref:  groupID,
	}, deleted)
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	if deleted["collaboration_groups"] != 1 {
		t.Fatalf("deleted rows = %v", deleted)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

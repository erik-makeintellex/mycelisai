package memory

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

// ownerMetaArg checks the metadata stored by PutTempMemory carries the
// expected owner stamp and never a caller-supplied one (MEM-LANES).
type ownerMetaArg struct{ owner string }

func (a ownerMetaArg) Match(v driver.Value) bool {
	raw, ok := v.([]byte)
	if !ok {
		return false
	}
	var meta map[string]any
	if json.Unmarshal(raw, &meta) != nil {
		return false
	}
	if a.owner == "" {
		_, hasOwner := meta[OwnerUserIDKey]
		return meta[OwnerClassKey] == OwnerClassSystem && !hasOwner
	}
	_, hasClass := meta[OwnerClassKey]
	return meta[OwnerUserIDKey] == a.owner && !hasClass
}

func TestPutTempMemory_HappyPath(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	svc := NewServiceWithDB(db)
	mock.ExpectQuery("INSERT INTO temp_memory_channels").
		WithArgs("default", "lead.shared", "admin", "checkpoint", ownerMetaArg{owner: "user-a"}, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("mem-1"))
	mock.ExpectQuery("INSERT INTO temp_memory_channels").
		WithArgs("default", "lead.shared", "admin", "checkpoint", ownerMetaArg{}, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("mem-2"))

	forged := map[string]any{"k": "v", OwnerUserIDKey: "user-b", OwnerClassKey: "x"}
	id, err := svc.PutTempMemory(context.Background(), "default", "lead.shared", "admin", "checkpoint", forged, 30, "user-a")
	if err != nil {
		t.Fatalf("PutTempMemory: %v", err)
	}
	if id != "mem-1" {
		t.Fatalf("id = %q, want mem-1", id)
	}
	if _, err := svc.PutTempMemory(context.Background(), "default", "lead.shared", "admin", "checkpoint", forged, 0, ""); err != nil {
		t.Fatalf("PutTempMemory without a user: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestGetTempMemory_HappyPath(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	svc := NewServiceWithDB(db)
	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "tenant_id", "channel_key", "owner_agent_id", "content", "metadata",
		"expires_at", "created_at", "updated_at",
	}).AddRow("mem-1", "default", "lead.shared", "admin", "checkpoint", `{"phase":"research"}`, nil, now, now)

	mock.ExpectQuery("SELECT id::text, tenant_id, channel_key, owner_agent_id, content, metadata").
		WithArgs("default", "lead.shared", "user-a", SignalCheckpointChannelPrefix, 10).
		WillReturnRows(rows)

	entries, err := svc.GetTempMemory(context.Background(), "default", "lead.shared", 10, GovernedReader{UserID: "user-a"})
	if err != nil {
		t.Fatalf("GetTempMemory: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("entries len = %d, want 1", len(entries))
	}
	if entries[0].ChannelKey != "lead.shared" {
		t.Fatalf("channel = %q", entries[0].ChannelKey)
	}
	if entries[0].Metadata["phase"] != "research" {
		t.Fatalf("metadata.phase = %v", entries[0].Metadata["phase"])
	}
}

func TestClearTempMemory_HappyPath(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	svc := NewServiceWithDB(db)
	mock.ExpectExec("DELETE FROM temp_memory_channels").
		WithArgs("default", "lead.shared", "user-a").
		WillReturnResult(sqlmock.NewResult(0, 2))

	deleted, err := svc.ClearTempMemory(context.Background(), "default", "lead.shared", GovernedReader{UserID: "user-a"})
	if err != nil {
		t.Fatalf("ClearTempMemory: %v", err)
	}
	if deleted != 2 {
		t.Fatalf("deleted = %d, want 2", deleted)
	}
}

func TestPutTempMemory_Validation(t *testing.T) {
	svc := NewServiceWithDB(nil)
	if _, err := svc.PutTempMemory(context.Background(), "default", "lead.shared", "admin", "x", nil, 0, ""); err == nil {
		t.Fatal("expected error when db is nil")
	}
}

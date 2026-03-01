package server

import (
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/memory"
)

func withMemoryDB(t *testing.T) (func(*AdminServer), sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	memSvc := memory.NewServiceWithDB(db)
	return func(s *AdminServer) {
		s.Mem = memSvc
	}, mock
}

func TestHandleTempMemory_Get_HappyPath(t *testing.T) {
	opt, mock := withMemoryDB(t)
	s := newTestServer(opt)
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "tenant_id", "channel_key", "owner_agent_id", "content", "metadata", "expires_at", "created_at", "updated_at",
	}).AddRow("mem-1", "default", "lead.shared", "admin", "checkpoint", `{"phase":"draft"}`, nil, now, now)
	mock.ExpectQuery("SELECT id::text, tenant_id, channel_key, owner_agent_id, content, metadata").
		WithArgs("default", "lead.shared", 10).
		WillReturnRows(rows)

	rr := doRequest(t, http.HandlerFunc(s.HandleTempMemory), "GET", "/api/v1/memory/temp?channel=lead.shared", "")
	assertStatus(t, rr, http.StatusOK)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	if resp["ok"] != true {
		t.Fatalf("expected ok=true, got %v", resp["ok"])
	}
}

func TestHandleTempMemory_Post_HappyPath(t *testing.T) {
	opt, mock := withMemoryDB(t)
	s := newTestServer(opt)

	mock.ExpectQuery("INSERT INTO temp_memory_channels").
		WithArgs("default", "lead.shared", "admin", "checkpoint", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("mem-2"))

	body := `{"channel":"lead.shared","content":"checkpoint","owner_agent_id":"admin","ttl_minutes":30,"metadata":{"phase":"draft"}}`
	rr := doRequest(t, http.HandlerFunc(s.HandleTempMemory), "POST", "/api/v1/memory/temp", body)
	assertStatus(t, rr, http.StatusOK)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	if resp["ok"] != true {
		t.Fatalf("expected ok=true, got %v", resp["ok"])
	}
}

func TestHandleTempMemory_Delete_HappyPath(t *testing.T) {
	opt, mock := withMemoryDB(t)
	s := newTestServer(opt)

	mock.ExpectExec("DELETE FROM temp_memory_channels").
		WithArgs("default", "lead.shared").
		WillReturnResult(sqlmock.NewResult(0, 1))

	rr := doRequest(t, http.HandlerFunc(s.HandleTempMemory), "DELETE", "/api/v1/memory/temp?channel=lead.shared", "")
	assertStatus(t, rr, http.StatusOK)
}

func TestHandleTempMemory_ValidationAndNilMem(t *testing.T) {
	s := newTestServer()

	rr := doRequest(t, http.HandlerFunc(s.HandleTempMemory), "GET", "/api/v1/memory/temp?channel=lead.shared", "")
	assertStatus(t, rr, http.StatusServiceUnavailable)

	opt, _ := withMemoryDB(t)
	s = newTestServer(opt)
	rr = doRequest(t, http.HandlerFunc(s.HandleTempMemory), "GET", "/api/v1/memory/temp", "")
	assertStatus(t, rr, http.StatusBadRequest)

	rr = doRequest(t, http.HandlerFunc(s.HandleTempMemory), "POST", "/api/v1/memory/temp", `{"channel":"x"}`)
	assertStatus(t, rr, http.StatusBadRequest)

	rr = doRequest(t, http.HandlerFunc(s.HandleTempMemory), "PATCH", "/api/v1/memory/temp", "")
	assertStatus(t, rr, http.StatusMethodNotAllowed)
}


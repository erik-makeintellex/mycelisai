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

// tempRootAdmin is the only caller the temp route admits (MEM-LANES).
var tempRootAdmin = memoryUser("user-root-temp", "root", "admin", "memory:write")

func tempRequest(t *testing.T, s *AdminServer, method, path, body string) int {
	t.Helper()
	return doAuthenticatedRequestAs(t, http.HandlerFunc(s.HandleTempMemory), method, path, body, tempRootAdmin).Code
}

func TestHandleTempMemory_Get_HappyPath(t *testing.T) {
	opt, mock := withMemoryDB(t)
	s := newTestServer(opt)
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "tenant_id", "channel_key", "owner_agent_id", "content", "metadata", "expires_at", "created_at", "updated_at",
	}).AddRow("mem-1", "default", "lead.shared", "admin", "checkpoint", `{"phase":"draft"}`, nil, now, now)
	mock.ExpectQuery("SELECT id::text, tenant_id, channel_key, owner_agent_id, content, metadata").
		WithArgs("default", "lead.shared", tempRootAdmin.UserID, memory.SignalCheckpointChannelPrefix, 10).
		WillReturnRows(rows)

	rr := doAuthenticatedRequestAs(t, http.HandlerFunc(s.HandleTempMemory), "GET", "/api/v1/memory/temp?channel=lead.shared", "", tempRootAdmin)
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
	if code := tempRequest(t, s, "POST", "/api/v1/memory/temp", body); code != http.StatusOK {
		t.Fatalf("status %d", code)
	}
}

func TestHandleTempMemory_Delete_HappyPath(t *testing.T) {
	opt, mock := withMemoryDB(t)
	s := newTestServer(opt)

	mock.ExpectExec("DELETE FROM temp_memory_channels").
		WithArgs("default", "lead.shared", tempRootAdmin.UserID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if code := tempRequest(t, s, "DELETE", "/api/v1/memory/temp?channel=lead.shared", ""); code != http.StatusOK {
		t.Fatalf("status %d", code)
	}
}

func TestHandleTempMemory_ValidationAndNilMem(t *testing.T) {
	s := newTestServer()
	if code := tempRequest(t, s, "GET", "/api/v1/memory/temp?channel=lead.shared", ""); code != http.StatusServiceUnavailable {
		t.Fatalf("nil mem: status %d", code)
	}

	opt, _ := withMemoryDB(t)
	s = newTestServer(opt)
	for _, tc := range []struct {
		method, path, body string
		want               int
	}{
		{"GET", "/api/v1/memory/temp", "", http.StatusBadRequest},
		{"POST", "/api/v1/memory/temp", `{"channel":"x"}`, http.StatusBadRequest},
		{"PATCH", "/api/v1/memory/temp", "", http.StatusMethodNotAllowed},
	} {
		if code := tempRequest(t, s, tc.method, tc.path, tc.body); code != tc.want {
			t.Errorf("%s %s: status %d, want %d", tc.method, tc.path, code, tc.want)
		}
	}
}

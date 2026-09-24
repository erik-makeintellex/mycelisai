package invocation

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type countFixture struct {
	server          *httptest.Server
	mu              sync.Mutex
	counts          map[string]int
	dropResponse    bool
	unavailableRead bool
}

func newCountFixture(t *testing.T) *countFixture {
	t.Helper()
	f := &countFixture{counts: map[string]int{}}
	f.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/increment" {
			http.NotFound(w, r)
			return
		}
		if r.Method == http.MethodPost {
			var request struct {
				InvocationID string `json:"invocation_id"`
				Counter      string `json:"counter"`
			}
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
			if request.InvocationID == "" || request.Counter == "" {
				http.Error(w, "missing fields", http.StatusBadRequest)
				return
			}
			f.mu.Lock()
			f.counts[request.InvocationID]++
			drop := f.dropResponse
			f.mu.Unlock()
			if drop {
				conn, _, err := w.(http.Hijacker).Hijack()
				if err == nil {
					_ = conn.Close()
				}
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"accepted":true}`))
			return
		}
		if r.Method == http.MethodGet {
			f.mu.Lock()
			unavailable := f.unavailableRead
			f.mu.Unlock()
			if unavailable {
				http.Error(w, "readback unavailable", http.StatusServiceUnavailable)
				return
			}
			id := r.URL.Query().Get("invocation_id")
			f.mu.Lock()
			count := f.counts[id]
			f.mu.Unlock()
			_ = json.NewEncoder(w).Encode(countEvidence{InvocationID: id, Counter: "ledger", ObservedCount: count})
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	}))
	t.Cleanup(f.server.Close)
	return f
}

func (f *countFixture) count(id string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.counts[id]
}

type pgFixture struct {
	db                                               *sql.DB
	store                                            *Store
	userID, accountID, groupID, membershipID, roleID string
	counter                                          *countFixture
}

func newPGFixture(t *testing.T) *pgFixture {
	t.Helper()
	dsn := strings.TrimSpace(os.Getenv("MYCELIS_INVOCATION_TEST_DSN"))
	if dsn == "" {
		t.Skip("MYCELIS_INVOCATION_TEST_DSN not set; real PostgreSQL proof is run in the disposable fixture lane")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.PingContext(t.Context()); err != nil {
		t.Fatal(err)
	}
	var found string
	if err := db.QueryRowContext(t.Context(), `SELECT to_regclass('execution_effect_grants')::text`).Scan(&found); err != nil || found == "" {
		t.Fatalf("canonical G4/E10 schema unavailable: %v %q", err, found)
	}
	f := &pgFixture{db: db, store: NewStore(db), counter: newCountFixture(t), userID: uuid.NewString(), accountID: uuid.NewString(), groupID: uuid.NewString(), membershipID: uuid.NewString(), roleID: uuid.NewString()}
	_, err = db.ExecContext(t.Context(), `INSERT INTO accounts(id,tenant_id,slug,name,status) VALUES($1,'default',$2,'Invocation Test','active')`, f.accountID, "invocation-"+f.accountID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.ExecContext(t.Context(), `INSERT INTO users(id,username,role,account_id,status) VALUES($1,$2,'operator',$3,'active')`, f.userID, "invocation-"+f.userID, f.accountID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.ExecContext(t.Context(), `INSERT INTO groups(id,account_id,key,name) VALUES($1,$2,$3,'Counting')`, f.groupID, f.accountID, "counting-"+f.groupID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.ExecContext(t.Context(), `INSERT INTO roles(id,account_id,key,name,scope) VALUES($1,$2,$3,'Counter','group')`, f.roleID, f.accountID, "counter-"+f.roleID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.ExecContext(t.Context(), `INSERT INTO role_permissions(role_id,permission_key) VALUES($1,$2)`, f.roleID, Permission)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.ExecContext(t.Context(), `INSERT INTO org_memberships(id,account_id,user_id,group_id,role_id,status) VALUES($1,$2,$3,$4,$5,'active')`, f.membershipID, f.accountID, f.userID, f.groupID, f.roleID)
	if err != nil {
		t.Fatal(err)
	}
	f.setBinding(t, f.counter.server.URL+"/increment")
	return f
}

func (f *pgFixture) setBinding(t *testing.T, endpoint string) {
	t.Helper()
	metadata := map[string]any{"invocation_binding": binding{Endpoint: endpoint, Adapter: "counting-http", Version: "1", Method: "POST", Schema: "counting.v1", UnitCost: 1}}
	_, err := f.db.ExecContext(t.Context(), `
		INSERT INTO capability_manifests(id,capability_id,display_name,kind,source,status,risk_class,metadata)
		VALUES($1,$1,'Counting','effect','invocation','available','low',$2::jsonb)
		ON CONFLICT(id) DO UPDATE SET status='available',metadata=EXCLUDED.metadata,updated_at=NOW()`, CapabilityID, string(jsonBytes(metadata)))
	if err != nil {
		t.Fatal(err)
	}
}

func (f *pgFixture) grant(t *testing.T, budget int) Grant {
	t.Helper()
	proposal, err := f.store.Propose(t.Context(), f.userID, Proposal{GroupID: f.groupID, Counter: "ledger", Budget: budget})
	if err != nil {
		t.Fatal(err)
	}
	granted, err := f.store.Confirm(t.Context(), f.userID, proposal.ConfirmToken)
	if err != nil {
		t.Fatal(err)
	}
	return granted
}

func (f *pgFixture) admit(t *testing.T, grant Grant, key string) Invocation {
	t.Helper()
	inv, reused, err := f.store.Admit(t.Context(), f.userID, grant.ID, grant.Digest, key, Input{Counter: "ledger"})
	if err != nil || reused {
		t.Fatalf("admit %s: reused=%v err=%v", key, reused, err)
	}
	return inv
}

func (f *pgFixture) setState(t *testing.T, statement string, args ...any) {
	t.Helper()
	if _, err := f.db.ExecContext(t.Context(), statement, args...); err != nil {
		t.Fatal(fmt.Errorf("fixture state: %w", err))
	}
}

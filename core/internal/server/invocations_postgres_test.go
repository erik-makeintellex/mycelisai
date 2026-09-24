package server

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/mycelis/core/internal/invocation"
)

// This traverses real HTTP authentication, proposal parsing, the existing
// confirmation endpoint, PostgreSQL admission and an external counting server.
func TestInvocationPostgresHTTPBoundary(t *testing.T) {
	dsn := os.Getenv("MYCELIS_INVOCATION_TEST_DSN")
	if dsn == "" {
		t.Skip("requires isolated MYCELIS_INVOCATION_TEST_DSN")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var mu sync.Mutex
	counts := map[string]int{}
	counter := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		if r.Method == http.MethodPost {
			var body struct {
				InvocationID string `json:"invocation_id"`
				Counter      string `json:"counter"`
			}
			if json.NewDecoder(r.Body).Decode(&body) != nil {
				w.WriteHeader(400)
				return
			}
			counts[body.InvocationID]++
			w.WriteHeader(200)
			return
		}
		id := r.URL.Query().Get("invocation_id")
		json.NewEncoder(w).Encode(map[string]any{"invocation_id": id, "counter": "http-proof", "observed_count": counts[id]})
	}))
	defer counter.Close()
	account, user, group, role, member := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()
	exec := func(query string, args ...any) {
		t.Helper()
		if _, err := db.Exec(query, args...); err != nil {
			t.Fatal(err)
		}
	}
	exec(`INSERT INTO accounts(id,slug,name) VALUES($1::uuid,$1::text,'HTTP proof')`, account)
	exec(`INSERT INTO users(id,username,account_id) VALUES($1::uuid,$1::text,$2)`, user, account)
	exec(`INSERT INTO groups(id,account_id,key,name) VALUES($1::uuid,$2,$1::text,'HTTP proof')`, group, account)
	exec(`INSERT INTO roles(id,account_id,key,name,scope) VALUES($1::uuid,$2,$1::text,'HTTP proof','group')`, role, account)
	exec(`INSERT INTO role_permissions(role_id,permission_key) VALUES($1,'counting.increment')`, role)
	exec(`INSERT INTO org_memberships(id,account_id,user_id,group_id,role_id) VALUES($1,$2,$3,$4,$5)`, member, account, user, group, role)
	metadata, _ := json.Marshal(map[string]any{"invocation_binding": map[string]any{
		"endpoint": counter.URL, "adapter": "counting-http", "version": "1", "method": "POST", "schema": "counting.v1", "unit_cost": 1,
	}})
	exec(`INSERT INTO capability_manifests(id,capability_id,display_name,kind,source,status,risk_class,metadata)
	 VALUES('counting.increment','counting.increment','Counting','governed_effect','core','available','low-risk',$1)
	 ON CONFLICT(id) DO UPDATE SET capability_id=EXCLUDED.capability_id,status='available',metadata=EXCLUDED.metadata`, metadata)
	t.Setenv("MYCELIS_LOCAL_ADMIN_USER_ID", user)
	s := &AdminServer{DB: db, Invocations: invocation.NewStore(db)}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/invocations/tool-proposals", s.HandleInvocationToolProposal)
	mux.HandleFunc("POST /api/v1/intent/confirm-action", s.HandleConfirmAction)
	mux.HandleFunc("POST /api/v1/invocations", s.HandleInvocation)
	mux.HandleFunc("GET /api/v1/invocations/{id}", s.HandleInvocationRead)
	api := httptest.NewServer(AuthMiddleware("non-secret-fixture-key", mux))
	defer api.Close()
	call := func(method, path string, body any, target any, want int) {
		t.Helper()
		encoded, _ := json.Marshal(body)
		req, _ := http.NewRequest(method, api.URL+path, bytes.NewReader(encoded))
		req.Header.Set("Authorization", "Bearer non-secret-fixture-key")
		req.Header.Set("Content-Type", "application/json")
		res, err := api.Client().Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		var envelope struct {
			Data  json.RawMessage `json:"data"`
			Error string          `json:"error"`
		}
		if err := json.NewDecoder(res.Body).Decode(&envelope); err != nil {
			t.Fatal(err)
		}
		if res.StatusCode != want {
			t.Fatalf("%s: status %d expected %d: %s", path, res.StatusCode, want, envelope.Error)
		}
		if target != nil {
			if err := json.Unmarshal(envelope.Data, target); err != nil {
				t.Fatal(err)
			}
		}
	}
	var proposed invocation.Proposed
	call("POST", "/api/v1/invocations/tool-proposals", map[string]any{"name": "counting.increment", "arguments": invocation.Proposal{GroupID: group, Counter: "http-proof", Budget: 1}}, &proposed, 200)
	mu.Lock()
	effects := len(counts)
	mu.Unlock()
	if effects != 0 {
		t.Fatal("model proposal caused effect")
	}
	s.InvocationAdmissionUnavailable = true
	call("POST", "/api/v1/intent/confirm-action", map[string]string{"confirm_token": proposed.ConfirmToken}, nil, 503)
	s.InvocationAdmissionUnavailable = false
	var grant invocation.Grant
	call("POST", "/api/v1/intent/confirm-action", map[string]string{"confirm_token": proposed.ConfirmToken}, &grant, 200)
	mu.Lock()
	effects = len(counts)
	mu.Unlock()
	if effects != 0 {
		t.Fatal("confirmation caused effect")
	}
	request := map[string]any{"grant_id": grant.ID, "grant_digest": grant.Digest, "idempotency_key": "http-once", "input": invocation.Input{Counter: "http-proof"}}
	var first, duplicate invocation.Invocation
	call("POST", "/api/v1/invocations", request, &first, 200)
	call("POST", "/api/v1/invocations", request, &duplicate, 200)
	if first.ID != duplicate.ID || first.State != "verified" {
		t.Fatalf("receipts: %+v %+v", first, duplicate)
	}
	mu.Lock()
	effects = counts[first.ID]
	mu.Unlock()
	if effects != 1 {
		t.Fatalf("duplicate effects: %d", effects)
	}
	call("GET", fmt.Sprintf("/api/v1/invocations/%s", first.ID), nil, &duplicate, 200)
	request["subject"] = "forged"
	call("POST", "/api/v1/invocations", request, nil, 400)
}

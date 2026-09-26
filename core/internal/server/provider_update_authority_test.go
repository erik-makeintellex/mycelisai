package server

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

// S6e F1: PUT /api/v1/cognitive/providers/{id} is gated, audited first,
// serialized on the shared routing-write mutex, and refuses to strand a
// bound provider.

const providerPutPath = "/api/v1/cognitive/providers/"

func newProviderUpdateFixture(t *testing.T) *routingFixture {
	t.Helper()
	f := newRoutingFixture(t)
	f.mux.HandleFunc("PUT /api/v1/cognitive/providers/{id}", f.s.HandleUpdateProvider)
	return f
}

func TestProviderUpdateAuthority_RejectsCallersWithoutScope(t *testing.T) {
	callers := []struct {
		name     string
		identity *RequestIdentity
		want     int
	}{
		{"anonymous", nil, http.StatusUnauthorized},
		{"standard user", &RequestIdentity{UserID: "u-1", Role: "user", Scopes: []string{"soma:work"}}, http.StatusForbidden},
		{"admin without cognitive:write", &RequestIdentity{UserID: "u-2", Role: "admin", Scopes: []string{"cognitive:read"}}, http.StatusForbidden},
		{"non-admin with cognitive:write", &RequestIdentity{UserID: "u-3", Role: "user", Scopes: []string{"cognitive:write"}}, http.StatusForbidden},
		{"non-admin with wildcard", &RequestIdentity{UserID: "u-4", Role: "user", Scopes: []string{"*"}}, http.StatusForbidden},
	}
	body := `{"endpoint":"http://attacker.invalid/v1","model_id":"evil"}`
	for _, caller := range callers {
		for _, id := range []string{"ollama", "evil"} {
			t.Run(caller.name+"/"+id, func(t *testing.T) {
				f := newProviderUpdateFixture(t)
				var rr *httptest.ResponseRecorder
				if caller.identity == nil {
					rr = doRequest(t, f.mux, http.MethodPut, providerPutPath+id, body)
				} else {
					rr = doAuthenticatedRequestAs(t, f.mux, http.MethodPut, providerPutPath+id, body, caller.identity)
				}
				assertStatus(t, rr, caller.want)
				f.assertRoutingUnchanged(t)
			})
		}
	}
}

func TestProviderUpdateAuthority_AuditFailureChangesNothing(t *testing.T) {
	t.Run("audit insert fails", func(t *testing.T) {
		f := newProviderUpdateFixture(t)
		f.mock.ExpectExec("INSERT INTO log_entries").WillReturnError(errors.New("audit down"))
		rr := doAuthenticatedRequest(t, f.mux, http.MethodPut, providerPutPath+"ollama", `{"endpoint":"http://127.0.0.1:7/v1"}`)
		assertStatus(t, rr, http.StatusServiceUnavailable)
		if !strings.Contains(rr.Body.String(), "Audit unavailable") {
			t.Fatalf("body = %s", rr.Body.String())
		}
		f.assertRoutingUnchanged(t)
	})
	t.Run("no audit store", func(t *testing.T) {
		f := newProviderUpdateFixture(t)
		f.s.DB = nil
		rr := doAuthenticatedRequest(t, f.mux, http.MethodPut, providerPutPath+"brand-new", `{"type":"openai","model_id":"x"}`)
		assertStatus(t, rr, http.StatusServiceUnavailable)
		f.assertRoutingUnchanged(t)
	})
}

// vllm is root for every execution profile. An update that leaves it
// non-executable (a whitespace-only model) is 409 provider_bound, before
// any audit or change. The route has no enabled field, so it cannot disable.
func TestProviderUpdateAuthority_BoundProviderRejected(t *testing.T) {
	f := newProviderUpdateFixture(t)
	rr := doAuthenticatedRequest(t, f.mux, http.MethodPut, providerPutPath+"vllm", `{"model_id":"   "}`)
	assertStatus(t, rr, http.StatusConflict)
	if !strings.Contains(rr.Body.String(), `"code":"provider_bound"`) || !strings.Contains(rr.Body.String(), "chat") {
		t.Fatalf("body = %s", rr.Body.String())
	}
	f.assertRoutingUnchanged(t)
}

func TestProviderUpdateAuthority_AdminWithScopeIsAuditedFirst(t *testing.T) {
	admin := &RequestIdentity{UserID: "root-1", Role: "admin", Scopes: []string{"cognitive:write"}}
	t.Run("bound provider stays executable", func(t *testing.T) {
		f := newProviderUpdateFixture(t)
		f.mock.ExpectExec("INSERT INTO log_entries").
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "audit", routingMutationAuditSource, sqlmock.AnyArg(), "Cognitive provider config updated", sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))
		rr := doAuthenticatedRequestAs(t, f.mux, http.MethodPut, providerPutPath+"vllm", `{"endpoint":"http://127.0.0.1:7/v1","model_id":"m2"}`, admin)
		assertStatus(t, rr, http.StatusOK)
		if p, _ := f.s.Cognitive.ProviderSnapshot("vllm"); p.Endpoint != "http://127.0.0.1:7/v1" || p.ModelID != "m2" || !p.Enabled {
			t.Fatalf("vllm = %+v", p)
		}
		if err := f.mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("sql: %v", err)
		}
	})
	t.Run("unbound provider may go non-executable", func(t *testing.T) {
		f := newProviderUpdateFixture(t)
		expectAudit(f.mock)
		assertStatus(t, doAuthenticatedRequestAs(t, f.mux, http.MethodPut, providerPutPath+"ollama", `{"model_id":"   "}`, admin), http.StatusOK)
	})
}

// The handler must block on the shared routing-write mutex.
func TestRoutingMutationMutex_ProviderUpdateWaitsForLock(t *testing.T) {
	f := newProviderUpdateFixture(t)
	expectAudit(f.mock)
	unlock := lockRoutingWrite()
	done := make(chan int, 1)
	go func() {
		done <- doAuthenticatedRequest(t, f.mux, http.MethodPut, providerPutPath+"ollama", `{"endpoint":"http://127.0.0.1:7/v1"}`).Code
	}()
	select {
	case code := <-done:
		unlock()
		t.Fatalf("provider PUT finished (%d) while the routing-write mutex was held", code)
	case <-time.After(150 * time.Millisecond):
	}
	if p, _ := f.s.Cognitive.ProviderSnapshot("ollama"); p.Endpoint != "http://127.0.0.1:9/v1" {
		unlock()
		t.Fatalf("provider changed while the mutex was held: %+v", p)
	}
	unlock()
	if code := <-done; code != http.StatusOK {
		t.Fatalf("status = %d", code)
	}
}

// A stale provider PUT must never undo a concurrent brains toggle: the end
// state equals one serial order (toggle and endpoint both applied).
func TestProviderUpdateRace_ConcurrentToggleEqualsSerialOrder(t *testing.T) {
	for i := 0; i < 30; i++ {
		f := newProviderUpdateFixture(t)
		f.mock.MatchExpectationsInOrder(false)
		expectAudit(f.mock)
		expectAudit(f.mock)
		var wg sync.WaitGroup
		codes := make([]int, 2)
		wg.Add(2)
		go func() {
			defer wg.Done()
			codes[0] = doAuthenticatedRequest(t, f.mux, http.MethodPut, "/api/v1/brains/ollama/toggle", `{"enabled":false}`).Code
		}()
		go func() {
			defer wg.Done()
			codes[1] = doAuthenticatedRequest(t, f.mux, http.MethodPut, providerPutPath+"ollama", `{"endpoint":"http://127.0.0.1:7/v1"}`).Code
		}()
		wg.Wait()
		if codes[0] != http.StatusOK || codes[1] != http.StatusOK {
			t.Fatalf("round %d codes = %v", i, codes)
		}
		p, _ := f.s.Cognitive.ProviderSnapshot("ollama")
		if p.Enabled || p.Endpoint != "http://127.0.0.1:7/v1" || p.ModelID != "m-ollama" {
			t.Fatalf("round %d: state is not a serial order: %+v", i, p)
		}
		if err := f.mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("round %d sql: %v", i, err)
		}
	}
}

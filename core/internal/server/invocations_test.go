package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mycelis/core/internal/invocation"
)

func invocationRequest(body string, authenticated bool) *http.Request {
	r := httptest.NewRequest(http.MethodPost, "/api/v1/invocations/tool-proposals", strings.NewReader(body))
	if authenticated {
		r = r.WithContext(context.WithValue(r.Context(), ctxKeyIdentity, &RequestIdentity{UserID: "00000000-0000-0000-0000-000000000001"}))
	}
	return r
}

func TestInvocationToolProposalRejectsForgedAuthorityBeforePersistence(t *testing.T) {
	s := &AdminServer{Invocations: invocation.NewStore(nil)}
	for _, body := range []string{
		`{"name":"counting.increment","arguments":{"group_id":"g","counter":"x","budget":1,"user_id":"forged"}}`,
		`{"name":"counting.increment","arguments":{},"subject":"forged"}`,
		`{"name":"counting.increment","arguments":{},"grant_id":"forged"}`,
		`{"name":"counting.increment","arguments":{"endpoint":"http://arbitrary"}}`,
		`{"name":"counting.increment","arguments":{"agent_id":"model"}}`,
		`{"name":"local_command","arguments":{}}`,
		`{"name":"counting.increment","arguments":{}} {"name":"second"}`,
	} {
		w := httptest.NewRecorder()
		s.HandleInvocationToolProposal(w, invocationRequest(body, true))
		if w.Code != http.StatusBadRequest {
			t.Fatalf("request %s: status=%d body=%s", body, w.Code, w.Body.String())
		}
	}
}

func TestInvocationRoutesRequireAuthenticationAndPersistence(t *testing.T) {
	s := &AdminServer{}
	for _, handler := range []http.HandlerFunc{s.HandleInvocationProposal, s.HandleInvocationToolProposal, s.HandleInvocation, s.HandleInvocationRead, s.HandleInvocationReconcile} {
		w := httptest.NewRecorder()
		handler(w, invocationRequest(`{}`, false))
		if w.Code != http.StatusUnauthorized {
			t.Fatal(w.Code)
		}
		w = httptest.NewRecorder()
		handler(w, invocationRequest(`{}`, true))
		if w.Code != http.StatusServiceUnavailable {
			t.Fatal(w.Code)
		}
	}
}

func TestInvocationReconcileCannotCarryExecutionAuthority(t *testing.T) {
	s := &AdminServer{Invocations: invocation.NewStore(nil)}
	w := httptest.NewRecorder()
	s.HandleInvocationReconcile(w, invocationRequest(`{"observation":"committed","evidence":{"source":"operator"},"retry":true}`, true))
	if w.Code != http.StatusBadRequest {
		t.Fatal(w.Code)
	}
}

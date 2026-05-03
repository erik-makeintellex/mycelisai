package server

import (
	"net/http"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestHandleIntentCommit_MissingToken(t *testing.T) {
	dbOpt, _ := withDB(t)
	s := newTestServer(dbOpt)

	body := `{
		"intent": "Build a scraper",
		"teams": [{"name": "t", "role": "r", "agents": []}]
	}`
	rr := doRequest(t, http.HandlerFunc(s.handleIntentCommit), "POST", "/api/v1/intent/commit", body)
	assertStatus(t, rr, http.StatusForbidden)
}

func TestHandleIntentCommit_InvalidToken(t *testing.T) {
	dbOpt, _ := withDB(t)
	s := newTestServer(dbOpt)

	body := `{
		"intent": "Build a scraper",
		"confirm_token": "not-a-uuid",
		"teams": [{"name": "t", "role": "r", "agents": []}]
	}`
	rr := doRequest(t, http.HandlerFunc(s.handleIntentCommit), "POST", "/api/v1/intent/commit", body)
	assertStatus(t, rr, http.StatusForbidden)
}

func TestHandleIntentCommit_TokenNotFound(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)

	mock.ExpectQuery("SELECT .+ FROM confirm_tokens").
		WillReturnRows(sqlmock.NewRows([]string{"intent_proof_id", "consumed", "expires_at"}))

	body := `{
		"intent": "Build a scraper",
		"confirm_token": "11111111-1111-1111-1111-111111111111",
		"teams": [{"name": "t", "role": "r", "agents": []}]
	}`
	rr := doRequest(t, http.HandlerFunc(s.handleIntentCommit), "POST", "/api/v1/intent/commit", body)
	assertStatus(t, rr, http.StatusForbidden)
}

func TestHandleIntentCommit_MissingIntent(t *testing.T) {
	dbOpt, _ := withDB(t)
	s := newTestServer(dbOpt)

	body := `{"teams":[{"name":"t","role":"r","agents":[]}]}`
	rr := doRequest(t, http.HandlerFunc(s.handleIntentCommit), "POST", "/api/v1/intent/commit", body)
	assertStatus(t, rr, http.StatusBadRequest)
}

func TestHandleIntentCommit_InvalidJSON(t *testing.T) {
	dbOpt, _ := withDB(t)
	s := newTestServer(dbOpt)

	rr := doRequest(t, http.HandlerFunc(s.handleIntentCommit), "POST", "/api/v1/intent/commit", "not-json")
	assertStatus(t, rr, http.StatusBadRequest)
}

func TestHandleIntentNegotiate_NilArchitect(t *testing.T) {
	s := newTestServer()
	rr := doRequest(t, http.HandlerFunc(s.handleIntentNegotiate), "POST", "/api/v1/intent/negotiate", `{"intent":"Build something"}`)
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

func TestHandleIntentNegotiate_MissingIntent(t *testing.T) {
	s := newTestServer()
	rr := doRequest(t, http.HandlerFunc(s.handleIntentNegotiate), "POST", "/api/v1/intent/negotiate", `{"intent":""}`)
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

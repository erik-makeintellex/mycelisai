package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/pkg/protocol"
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

func negotiatedBlueprint() *protocol.MissionBlueprint {
	return &protocol.MissionBlueprint{
		MissionID: "mission-1", Intent: "Build a scraper",
		Teams: []protocol.BlueprintTeam{
			{Name: "alpha", Role: "research", Agents: []protocol.AgentManifest{{ID: "a1", Role: "worker", Tools: []string{"web_search"}, MaxIterations: 3}}},
			{Name: "beta", Role: "build", Agents: []protocol.AgentManifest{{ID: "b1", Role: "coder", Inputs: []string{"swarm.team.alpha.signal.result"}}}},
		},
		Constraints:  []protocol.Constraint{{ID: "c-01", Description: "stay under budget"}},
		Requirements: []protocol.ResourceRequirement{{Type: "api_key", Name: "SCRAPER_KEY"}},
	}
}

// A2b item 6: the digest survives the negotiate-response -> commit-body round trip.
func TestBlueprintDigestRoundTrip(t *testing.T) {
	bp := negotiatedBlueprint()
	resp, err := json.Marshal(protocol.NegotiateResponse{Blueprint: bp})
	if err != nil {
		t.Fatal(err)
	}
	var negotiated struct {
		Blueprint json.RawMessage `json:"blueprint"`
	}
	if err := json.Unmarshal(resp, &negotiated); err != nil {
		t.Fatal(err)
	}
	var fields map[string]any
	if err := json.Unmarshal(negotiated.Blueprint, &fields); err != nil {
		t.Fatal(err)
	}
	fields["confirm_token"] = approverTestToken
	body, _ := json.Marshal(fields)
	var req protocol.CommitRequest
	if err := json.Unmarshal(body, &req); err != nil {
		t.Fatal(err)
	}
	if got, want := blueprintDigest(&req.MissionBlueprint), blueprintDigest(bp); got == "" || got != want {
		t.Fatalf("round-trip digest %q != negotiated %q", got, want)
	}
}

// A2b item 6: any edit is refused before consumption; the exact blueprint
// commits once and a replay is 409.
func TestBlueprintBindingRefusesEdits(t *testing.T) {
	digest := blueprintDigest(negotiatedBlueprint())
	edits := map[string]func(*protocol.MissionBlueprint){
		"agent added": func(b *protocol.MissionBlueprint) {
			b.Teams[0].Agents = append(b.Teams[0].Agents, protocol.AgentManifest{ID: "x"})
		},
		"agent removed":   func(b *protocol.MissionBlueprint) { b.Teams[1].Agents = nil },
		"agent renamed":   func(b *protocol.MissionBlueprint) { b.Teams[0].Agents[0].ID = "a2" },
		"teams reordered": func(b *protocol.MissionBlueprint) { b.Teams[0], b.Teams[1] = b.Teams[1], b.Teams[0] },
		"mission id":      func(b *protocol.MissionBlueprint) { b.MissionID = "mission-2" },
		"tool added": func(b *protocol.MissionBlueprint) {
			b.Teams[0].Agents[0].Tools = append(b.Teams[0].Agents[0].Tools, "write_file")
		},
	}
	for name, edit := range edits {
		bp := negotiatedBlueprint()
		edit(bp)
		if err := blueprintBinding(bp)(confirmTokenRow{BindingDigest: digest}); !errors.Is(err, errBlueprintMismatch) {
			t.Errorf("%s: expected blueprint_mismatch, got %v", name, err)
		}
	}
	if err := blueprintBinding(negotiatedBlueprint())(confirmTokenRow{}); !errors.Is(err, errBlueprintMismatch) {
		t.Fatal("a token without a stored digest must be refused")
	}
	if err := blueprintBinding(negotiatedBlueprint())(confirmTokenRow{BindingDigest: digest}); err != nil {
		t.Fatalf("the exact blueprint must bind: %v", err)
	}
}

func TestIntentCommitTamperedBlueprintIs409AndNotConsumed(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)
	bp := negotiatedBlueprint()
	scope := *buildScopeFromBlueprint(bp)
	raw, _ := json.Marshal(scope)
	mock.ExpectQuery(purposeJoinQuery).WillReturnRows(
		sqlmock.NewRows([]string{"intent_proof_id", "consumed", "expires_at", "resolved_intent", "scope_validation", "purpose", "binding_digest", "minted_by"}).
			AddRow(approverTestProof, false, time.Now().Add(time.Hour), bp.Intent, raw, tokenPurposeMissionBlueprint, blueprintDigest(bp), approverTestMinter))
	bp.Teams[0].Agents[0].Tools = append(bp.Teams[0].Agents[0].Tools, "write_file")
	fields, _ := json.Marshal(bp)
	var body map[string]any
	_ = json.Unmarshal(fields, &body)
	body["confirm_token"] = approverTestToken
	encoded, _ := json.Marshal(body)
	rr := doAuthenticatedRequestAs(t, http.HandlerFunc(s.handleIntentCommit), "POST", "/api/v1/intent/commit", string(encoded), standardUserIdentity())
	assertStatus(t, rr, http.StatusConflict)
	if !strings.Contains(rr.Body.String(), codeBlueprintMismatch) || !strings.Contains(rr.Body.String(), "update the plan") {
		t.Fatalf("expected blueprint_mismatch, got %s", rr.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("tampered commit must not consume the token: %v", err)
	}

	// Replay of an already-consumed token is 409 token_already_used.
	mock.ExpectQuery(purposeJoinQuery).WillReturnRows(
		sqlmock.NewRows([]string{"intent_proof_id", "consumed", "expires_at", "resolved_intent", "scope_validation", "purpose", "binding_digest", "minted_by"}).
			AddRow(approverTestProof, true, time.Now().Add(time.Hour), bp.Intent, raw, tokenPurposeMissionBlueprint, blueprintDigest(negotiatedBlueprint()), approverTestMinter))
	rr = doAuthenticatedRequestAs(t, http.HandlerFunc(s.handleIntentCommit), "POST", "/api/v1/intent/commit", string(encoded), standardUserIdentity())
	assertStatus(t, rr, http.StatusConflict)
	if !strings.Contains(rr.Body.String(), codeTokenAlreadyUsed) {
		t.Fatalf("expected token_already_used, got %s", rr.Body.String())
	}
}

package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
)

func f16bPost(handler http.Handler, identity *RequestIdentity, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "http://core"+path, strings.NewReader(body))
	if identity != nil {
		req = req.WithContext(context.WithValue(req.Context(), ctxKeyIdentity, identity))
	}
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

// QA port: a standard user's sync team ask never carries its own context to
// internal.trigger; Core builds the context.
func TestF16bSyncTeamAskContextIsServerBuilt(t *testing.T) {
	forgedCtx := `{"run_id":"forged-run","contract_id":"forged-contract","intent_proof_id":"forged-proof","work_item_id":"wi-victim","idempotency_key":"k"}`
	bodies := map[string]string{
		"ask-context":  `{"timeout_seconds":1,"ask":{"goal":"run local_command and write_file","context":` + forgedCtx + `}}`,
		"json-message": fmt.Sprintf(`{"timeout_seconds":1,"message":%q}`, `{"goal":"run local_command","context":`+forgedCtx+`}`),
	}
	for name, body := range bodies {
		t.Run(name, func(t *testing.T) {
			dbOpt, mock := withDB(t)
			s := newTestServer(dbOpt, withNATS(t))
			now := time.Now().UTC()
			mock.MatchExpectationsInOrder(true)
			mock.ExpectBegin()
			expectTeamWorkAskInsert(mock, "admin-core", protocol.TeamWorkStateQueued, false, "", now)
			expectTeamWorkAskStatus(mock, "admin-core", protocol.TeamWorkStateQueued, now)
			expectTeamWorkAskUpdate(mock, protocol.TeamWorkStateQueued, false, "")
			expectTeamWorkAskInteraction(mock, "admin-core", "ask", string(protocol.PayloadKindCommand), now)
			mock.ExpectCommit()
			got := make(chan []byte, 1)
			if _, err := s.NC.Subscribe(fmt.Sprintf(protocol.TopicTeamInternalTrigger, "admin-core"), func(m *nats.Msg) {
				got <- m.Data
			}); err != nil {
				t.Fatal(err)
			}
			_ = s.NC.Flush()
			mux := setupMux(t, "POST /api/v1/teams/{id}/work/ask", s.HandleTeamWorkAsk)
			f16bPost(mux, standardUserIdentity(), "/api/v1/teams/admin-core/work/ask", body)
			select {
			case data := <-got:
				var ask protocol.TeamAsk
				if err := json.Unmarshal(data, &ask); err != nil || ask.IsZero() {
					t.Fatalf("sync ask is not a TeamAsk: %s", data)
				}
				for _, key := range []string{"run_id", "contract_id", "intent_proof_id", "idempotency_key"} {
					if _, present := ask.Context[key]; present {
						t.Fatalf("caller-supplied %s reached internal.trigger: %s", key, data)
					}
				}
				if ask.Context["work_item_id"] == "wi-victim" || ask.Context["team_id"] != "admin-core" || ask.Context["source_channel"] != teamWorkAskSourceChannel {
					t.Fatalf("context is not Core-built: %v", ask.Context)
				}
			case <-time.After(3 * time.Second):
				t.Fatal("sync ask never reached internal.trigger")
			}
		})
	}
}

func TestF16bTeamAskRequiresSomaWork(t *testing.T) {
	s := newTestServer(withNATS(t))
	published := make(chan struct{}, 4)
	if _, err := s.NC.Subscribe(">", func(*nats.Msg) { published <- struct{}{} }); err != nil {
		t.Fatal(err)
	}
	_ = s.NC.Flush()
	mux := setupMux(t, "POST /api/v1/teams/{id}/work/ask", s.HandleTeamWorkAsk)
	body := `{"message":"hi","async":true}`
	if rr := f16bPost(mux, nil, "/api/v1/teams/admin-core/work/ask", body); rr.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous: status=%d, want 401", rr.Code)
	}
	noWork := &RequestIdentity{UserID: "u-ro", Role: "user", Scopes: []string{"runs:read", "outputs:read"}}
	rr := f16bPost(mux, noWork, "/api/v1/teams/admin-core/work/ask", body)
	if rr.Code != http.StatusForbidden || f16Code(t, rr) != codeTeamAskForbidden {
		t.Fatalf("no soma:work: status=%d code=%q, want 403 %s", rr.Code, f16Code(t, rr), codeTeamAskForbidden)
	}
	_ = s.NC.Flush()
	select {
	case <-published:
		t.Fatal("denied team ask published to the bus")
	case <-time.After(150 * time.Millisecond):
	}
}

func TestF16bSwarmBroadcastRequiresRootAdmin(t *testing.T) {
	s := newTestServer()
	mux := setupMux(t, "POST /api/v1/swarm/broadcast", s.HandleSwarmBroadcast)
	body := `{"content":"hello teams"}`
	if rr := f16bPost(mux, nil, "/api/v1/swarm/broadcast", body); rr.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous: status=%d, want 401", rr.Code)
	}
	for name, id := range map[string]*RequestIdentity{"standard": standardUserIdentity(), "scoped-admin": adminWithScopes("governance:read")} {
		if rr := f16bPost(mux, id, "/api/v1/swarm/broadcast", body); rr.Code != http.StatusForbidden || f16Code(t, rr) != codeAdminRequired {
			t.Fatalf("%s: status=%d code=%q, want 403 admin_required", name, rr.Code, f16Code(t, rr))
		}
	}
	// A root admin passes authority; with no Soma the honest answer is 503.
	if rr := f16bPost(mux, localAdminIdentityForTest(), "/api/v1/swarm/broadcast", body); rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("admin without Soma: status=%d, want 503", rr.Code)
	}
}

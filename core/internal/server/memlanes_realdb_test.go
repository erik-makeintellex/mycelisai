package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/swarm"
	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
)

// MEM-LANES. The recall turn token travels on unauthenticated NATS, so a
// passive subscriber can read it. It must be spendable only on the agent
// subject and reply inbox Core issued it for, and only briefly.

func memlanesDecode(t *testing.T, label string, reply *nats.Msg) swarm.ProcessResult {
	t.Helper()
	var result swarm.ProcessResult
	if err := json.Unmarshal(reply.Data, &result); err != nil {
		t.Fatalf("%s: decode %s", label, reply.Data)
	}
	return result
}

// memlanesSpend sends msg as Core would: to its own reply inbox when it has
// one, otherwise as a plain request.
func memlanesSpend(t *testing.T, nc *nats.Conn, msg *nats.Msg) *nats.Msg {
	t.Helper()
	if msg.Reply == "" {
		reply, err := nc.RequestMsg(msg, 20*time.Second)
		if err != nil {
			t.Fatalf("request: %v", err)
		}
		return reply
	}
	sub, err := nc.SubscribeSync(msg.Reply)
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	defer sub.Unsubscribe()
	if err := nc.PublishMsg(msg); err != nil {
		t.Fatalf("publish: %v", err)
	}
	reply, err := sub.NextMsg(20 * time.Second)
	if err != nil {
		t.Fatalf("reply: %v", err)
	}
	return reply
}

func TestMemLanesTokenRealDB_HarvestedTokenCannotBeSpentElsewhere(t *testing.T) {
	w := sruChatSeed(t)
	ctx := context.WithValue(context.Background(), ctxKeyIdentity, w.ada)
	payload := []byte(`[{"role":"user","content":"What is our quokka weekend plan?"}]`)
	soma := fmt.Sprintf(protocol.TopicCouncilRequestFmt, "admin")
	attack := func(label, token string) {
		t.Helper()
		msg := nats.NewMsg(soma)
		msg.Data = payload
		msg.Header.Set(swarm.RecallTurnHeader, token)
		w.provider.take()
		result := memlanesDecode(t, label, memlanesSpend(t, w.s.NC, msg))
		sruChatCheck(t, label, w.provider.take(), nil, []string{"ada-private", "ada-reflection", "team"})
		if !strings.Contains(result.Text, "Saved memory context was unavailable") {
			t.Errorf("%s: a harvested token must withhold memory and say so: %q", label, result.Text)
		}
	}
	// A token issued for another agent, spent on Soma from the attacker's inbox.
	other, releaseOther := w.s.recallTurnMsg(ctx, fmt.Sprintf(protocol.TopicCouncilRequestFmt, "council-memlanes"), payload)
	defer releaseOther()
	attack("token for another agent", other.Header.Get(swarm.RecallTurnHeader))

	// A token issued for Soma, raced from the attacker's own reply inbox.
	legit, release := w.s.recallTurnMsg(ctx, soma, payload)
	defer release()
	attack("token raced from another inbox", legit.Header.Get(swarm.RecallTurnHeader))

	// The refused races did not burn it: Core's own request still recalls as ada.
	w.provider.take()
	memlanesDecode(t, "legit", memlanesSpend(t, w.s.NC, legit))
	sruChatCheck(t, "legit turn", w.provider.take(), []string{"ada-private", "org"}, []string{"bob-private"})
}

func memlanesTempDB(t *testing.T) (*sql.DB, *AdminServer) {
	t.Helper()
	db := openMemoryServerTestDB(t)
	clean := func() { _, _ = db.Exec(`DELETE FROM temp_memory_channels WHERE content LIKE 'MemLanes%'`) }
	clean()
	t.Cleanup(clean)
	return db, memoryTestServer(db, nil)
}

func TestMemLanesTempRouteRealDB_RootAdminOnlyAndOwnRows(t *testing.T) {
	db, s := memlanesTempDB(t)
	h := http.HandlerFunc(s.HandleTempMemory)
	ada := memoryUser(uuid.NewString(), "ada-lanes", "operator", "memory:read", "memory:write")
	root := memoryUser(uuid.NewString(), "root-lanes", "admin", "memory:write")
	if _, err := db.Exec(`INSERT INTO temp_memory_channels (tenant_id, channel_key, owner_agent_id, content, metadata)
		VALUES ('default', 'lead.shared', 'admin', 'MemLanes ada kiwi 3131.', jsonb_build_object('owner_user_id', $1::text))`, ada.UserID); err != nil {
		t.Fatalf("seed ada row: %v", err)
	}
	for _, method := range []string{http.MethodGet, http.MethodPost, http.MethodDelete} {
		body := ""
		if method == http.MethodPost {
			body = `{"channel":"lead.shared","content":"MemLanes forged 9999."}`
		}
		if rr := doRequest(t, h, method, "/api/v1/memory/temp?channel=lead.shared", body); rr.Code != http.StatusUnauthorized {
			t.Errorf("%s without identity: status %d, want 401", method, rr.Code)
		}
		if rr := doAuthenticatedRequestAs(t, h, method, "/api/v1/memory/temp?channel=lead.shared", body, ada); rr.Code != http.StatusForbidden {
			t.Errorf("%s as an operator: status %d, want 403", method, rr.Code)
		}
	}
	rr := doAuthenticatedRequestAs(t, h, http.MethodPost, "/api/v1/memory/temp", `{"channel":"lead.shared","content":"MemLanes root kiwi 4141.","metadata":{"owner_user_id":"`+ada.UserID+`"}}`, root)
	assertStatus(t, rr, http.StatusOK)
	rr = doAuthenticatedRequestAs(t, h, http.MethodGet, "/api/v1/memory/temp?channel=lead.shared", "", root)
	assertStatus(t, rr, http.StatusOK)
	if body := rr.Body.String(); !strings.Contains(body, "4141") || strings.Contains(body, "3131") {
		t.Fatalf("root reads only its own rows: %s", body)
	}
	assertStatus(t, doAuthenticatedRequestAs(t, h, http.MethodDelete, "/api/v1/memory/temp?channel=lead.shared", "", root), http.StatusOK)
	var adaRows, forged int
	_ = db.QueryRow(`SELECT count(*) FROM temp_memory_channels WHERE content LIKE 'MemLanes ada%'`).Scan(&adaRows)
	_ = db.QueryRow(`SELECT count(*) FROM temp_memory_channels WHERE content LIKE 'MemLanes forged%' OR content LIKE 'MemLanes root%'`).Scan(&forged)
	if adaRows != 1 || forged != 0 {
		t.Fatalf("root's clear must remove only its own rows: ada=%d root/forged=%d", adaRows, forged)
	}
}

// A failed memory-access check answers with the blocker envelope.
func TestMemLanesSearchRealDB_AccessFailureUsesBlockerEnvelope(t *testing.T) {
	db := openMemoryServerTestDB(t)
	s := memoryTestServer(db, nil)
	broken, err := sql.Open("pgx", "postgres://nobody:nothing@127.0.0.1:1/none?sslmode=disable&connect_timeout=1")
	if err != nil {
		t.Fatal(err)
	}
	defer broken.Close()
	s.Artifacts.DB = broken
	rr := doAuthenticatedRequestAs(t, http.HandlerFunc(s.HandleSearch), http.MethodPost, "/api/v1/search", `{"query":"quokka","source_scope":"local_sources"}`,
		memoryUser(uuid.NewString(), "ada-lanes", "operator", "memory:read"))
	assertStatus(t, rr, http.StatusServiceUnavailable)
	var resp struct {
		OK   bool              `json:"ok"`
		Data map[string]string `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil || resp.OK || resp.Data["code"] != codeServiceUnavailable {
		t.Fatalf("want the service_unavailable blocker envelope: %s", rr.Body.String())
	}
}

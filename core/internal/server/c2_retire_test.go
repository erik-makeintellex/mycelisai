package server

import (
	"bytes"
	"database/sql/driver"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"google.golang.org/protobuf/proto"

	"github.com/mycelis/core/internal/governance"
	"github.com/mycelis/core/internal/router"
	pb "github.com/mycelis/core/pkg/pb/swarm"
	"github.com/mycelis/core/pkg/protocol"
)

// C2-RETIRE: the in-memory approval queue routes are gone, with no alias.
// A full admin reaches the real mux and still gets 404 on every one of them.
func TestC2RetireApprovalQueueRoutesReturn404(t *testing.T) {
	s := newTestServer(withGuard(defaultTestPolicyConfig()))
	mux := http.NewServeMux()
	s.RegisterRoutes(mux)
	for _, c := range []struct{ method, path, body string }{
		{"GET", "/api/v1/governance/pending", ""},
		{"POST", "/api/v1/governance/resolve/req-1", `{"action":"APPROVE"}`},
		{"GET", "/admin/approvals", ""},
		{"POST", "/admin/approvals/req-1", `{"action":"APPROVE"}`},
		{"GET", "/admin/approvals/req-1", ""},
	} {
		rr := doAuthenticatedRequest(t, mux, c.method, c.path, c.body)
		if rr.Code != http.StatusNotFound {
			t.Errorf("%s %s: expected 404, got %d %s", c.method, c.path, rr.Code, rr.Body.String())
		}
	}
	// The durable governance routes stay registered.
	if rr := doAuthenticatedRequest(t, mux, "GET", "/api/v1/governance/policy", ""); rr.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/governance/policy: expected 200, got %d", rr.Code)
	}
}

// c2rContextCapture records the context JSON argument of log_entries inserts;
// it is locked because the Router writes from a NATS goroutine.
type c2rContextCapture struct {
	mu   sync.Mutex
	ctxs []string
}

func (c *c2rContextCapture) Match(v driver.Value) bool {
	if b, ok := v.([]byte); ok {
		c.mu.Lock()
		c.ctxs = append(c.ctxs, string(b))
		c.mu.Unlock()
	}
	return true
}

func (c *c2rContextCapture) contexts() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]string(nil), c.ctxs...)
}

// c2rLockedBuffer is a log sink safe to read while the Router writes to it.
type c2rLockedBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *c2rLockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *c2rLockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func c2rExpectAuditRow(mock sqlmock.Sqlmock, capture *c2rContextCapture) {
	a := sqlmock.AnyArg()
	mock.ExpectExec("INSERT INTO log_entries").
		WithArgs(a, a, a, a, a, a, a, capture).
		WillReturnResult(sqlmock.NewResult(1, 1))
}

func c2rObservation() router.ApprovalObservation {
	return router.ApprovalObservation{Subject: "swarm.team.alpha.signal.result", MessageID: "m-1",
		TeamID: "alpha", SourceAgentID: "w", Intent: "k8s.delete.cluster"}
}

// The server's auditor writes one audit row that names the observation and
// says Core did not act; it never claims an approval or a queue entry.
func TestC2RetireObservationAuditWritesOneRow(t *testing.T) {
	dbOpt, mock := withDB(t)
	capture := &c2rContextCapture{}
	c2rExpectAuditRow(mock, capture)
	s := newTestServer(dbOpt)
	id, err := s.auditPolicyApprovalObserved(c2rObservation())
	if err != nil || id == "" {
		t.Fatalf("expected an audit id, got %q err=%v", id, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
	contexts := capture.contexts()
	if len(contexts) != 1 {
		t.Fatalf("expected one audit row, got %d", len(contexts))
	}
	for _, want := range []string{`"action":"` + router.PolicyApprovalObservedAction + `"`, `"team_id":"alpha"`,
		`"source_agent":"w"`, `"intent":"k8s.delete.cluster"`, `"result_status":"not_acted_on"`, `"approval_path":"proposals"`} {
		if !strings.Contains(contexts[0], want) {
			t.Fatalf("audit context missing %s: %s", want, contexts[0])
		}
	}
	for _, forbidden := range []string{"request_id", "resolved", "pending"} {
		if strings.Contains(contexts[0], forbidden) {
			t.Fatalf("audit context must not claim a queue entry (%s): %s", forbidden, contexts[0])
		}
	}
}

// Audit down: no DB gives an empty id, an insert error gives the error; the
// Router treats both as unavailable and logs instead.
func TestC2RetireObservationAuditUnavailable(t *testing.T) {
	if id, _ := newTestServer().auditPolicyApprovalObserved(c2rObservation()); id != "" {
		t.Fatalf("no audit DB must return an empty id, got %q", id)
	}
	dbOpt, mock := withDB(t)
	mock.ExpectExec("INSERT INTO log_entries").WillReturnError(errors.New("audit down"))
	if id, err := newTestServer(dbOpt).auditPolicyApprovalObserved(c2rObservation()); id != "" || err == nil {
		t.Fatalf("insert failure must return an error and no id, got %q %v", id, err)
	}
}

// End to end: NewAdminServer wires its audit store into the Router, so a
// REQUIRE_APPROVAL event on the bus becomes exactly one audit row.
func TestC2RetireNewAdminServerWiresObservationAudit(t *testing.T) {
	logs := &c2rLockedBuffer{}
	log.SetOutput(logs)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })
	carrier := newTestServer(withNATS(t))
	nc := carrier.NC
	dbOpt, mock := withDB(t)
	dbHolder := newTestServer(dbOpt)
	capture := &c2rContextCapture{}
	c2rExpectAuditRow(mock, capture)

	guard := governance.NewDegradedGuard(errors.New("start degraded"))
	guard.UpdatePolicyConfig(&governance.PolicyConfig{
		Groups: []governance.PolicyGroup{{Name: "g", Targets: []string{"*"},
			Rules: []governance.PolicyRule{{Intent: "k8s.delete.cluster", Action: governance.ActionRequireApproval}}}},
		Defaults: governance.DefaultConfig{DefaultAction: governance.ActionAllow},
	})
	rt := router.NewRouter(nc, guard)
	if err := rt.Start(); err != nil {
		t.Fatal(err)
	}
	s := NewAdminServer(rt, guard, nil, dbHolder.DB, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	if s.Router != rt {
		t.Fatal("NewAdminServer must keep the Router")
	}

	data, err := proto.Marshal(&pb.MsgEnvelope{Id: "m-9", SourceAgentId: "w", TeamId: "alpha",
		Payload: &pb.MsgEnvelope_Event{Event: &pb.EventPayload{EventType: "k8s.delete.cluster"}}})
	if err != nil {
		t.Fatal(err)
	}
	if err := nc.Publish(fmt.Sprintf(protocol.TopicTeamTelemetryFmt, "alpha"), data); err != nil {
		t.Fatal(err)
	}
	_ = nc.Flush()
	// The Router logs after the audit call returns; wait for that line (not
	// for sqlmock, which is not safe to poll while the insert runs).
	deadline := time.Now().Add(3 * time.Second)
	for !strings.Contains(logs.String(), router.PolicyApprovalObservedAction) && time.Now().Before(deadline) {
		time.Sleep(20 * time.Millisecond)
	}
	if !strings.Contains(logs.String(), router.PolicyApprovalObservedAction+" audit_id=") {
		t.Fatalf("expected an audited observation, log: %s", logs.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("REQUIRE_APPROVAL on the bus must write one audit row through the wired server: %v", err)
	}
	if contexts := capture.contexts(); len(contexts) == 0 || !strings.Contains(contexts[0], `"message_id":"m-9"`) {
		t.Fatalf("audit row must carry the observed envelope: %v", contexts)
	}
}

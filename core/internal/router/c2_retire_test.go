package router

import (
	"bytes"
	"errors"
	"log"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/mycelis/core/internal/governance"
	"github.com/mycelis/core/pkg/protocol"
)

// c2rApprovalGuard is a loaded guard that requires approval for
// k8s.delete.cluster and allows everything else.
func c2rApprovalGuard(extra ...governance.PolicyRule) *governance.Guard {
	g := governance.NewDegradedGuard(errors.New("start degraded"))
	rules := append([]governance.PolicyRule{{Intent: "k8s.delete.cluster", Action: governance.ActionRequireApproval}}, extra...)
	g.UpdatePolicyConfig(&governance.PolicyConfig{
		Groups:   []governance.PolicyGroup{{Name: "g", Targets: []string{"*"}, Rules: rules}},
		Defaults: governance.DefaultConfig{DefaultAction: governance.ActionAllow},
	})
	return g
}

// c2rBusRecorder records every subject Core publishes on the product bus.
func c2rBusRecorder(t *testing.T, nc *nats.Conn) func() []string {
	t.Helper()
	var mu sync.Mutex
	var subjects []string
	if _, err := nc.Subscribe(protocol.TopicSwarmWild, func(m *nats.Msg) {
		mu.Lock()
		subjects = append(subjects, m.Subject)
		mu.Unlock()
	}); err != nil {
		t.Fatal(err)
	}
	if err := nc.Flush(); err != nil {
		t.Fatal(err)
	}
	return func() []string {
		time.Sleep(150 * time.Millisecond) // audit trace publishes asynchronously
		_ = nc.Flush()
		mu.Lock()
		defer mu.Unlock()
		return append([]string(nil), subjects...)
	}
}

// C2-RETIRE: a REQUIRE_APPROVAL bus event is observed, not parked. Core
// publishes nothing for it: no swarm.governance.needed, no re-publish to an
// *.output subject, no audit trace.
func TestC2RetireRequireApprovalPublishesNothing(t *testing.T) {
	nc := startTestNATS(t)
	published := c2rBusRecorder(t, nc)
	r := NewRouter(nc, c2rApprovalGuard())
	deliver(t, r, "swarm.team.alpha.signal.result", "w", "k8s.delete.cluster")
	if got := published(); len(got) != 0 {
		t.Fatalf("REQUIRE_APPROVAL must publish nothing, got %v", got)
	}
}

// c2rAuditSink records every observation handed to the Router's auditor.
type c2rAuditSink struct {
	mu   sync.Mutex
	seen []ApprovalObservation
	id   string
	err  error
}

func (s *c2rAuditSink) record(obs ApprovalObservation) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seen = append(s.seen, obs)
	return s.id, s.err
}

func (s *c2rAuditSink) observations() []ApprovalObservation {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]ApprovalObservation(nil), s.seen...)
}

// One REQUIRE_APPROVAL event writes exactly one audit record carrying the
// envelope facts; ALLOW and DENY events write none; nothing is published for
// the REQUIRE_APPROVAL event (the one publish is the ALLOW event's trace).
func TestC2RetireRequireApprovalWritesOneAuditRecord(t *testing.T) {
	nc := startTestNATS(t)
	published := c2rBusRecorder(t, nc)
	r := NewRouter(nc, c2rApprovalGuard(governance.PolicyRule{Intent: "system.shutdown", Action: governance.ActionDeny}))
	sink := &c2rAuditSink{id: "audit-42"}
	r.SetApprovalAuditor(sink.record)

	deliver(t, r, "swarm.team.alpha.signal.result", "w", "task.completed")
	deliver(t, r, "swarm.team.alpha.signal.result", "w", "system.shutdown")
	deliver(t, r, "swarm.team.alpha.signal.result", "w", "k8s.delete.cluster")

	got := sink.observations()
	if len(got) != 1 {
		t.Fatalf("expected exactly one approval observation, got %d: %+v", len(got), got)
	}
	want := ApprovalObservation{Subject: "swarm.team.alpha.signal.result", TeamID: "alpha", SourceAgentID: "w", Intent: "k8s.delete.cluster"}
	if got[0] != want {
		t.Fatalf("observation = %+v, want %+v", got[0], want)
	}
	if subjects := published(); len(subjects) != 1 || subjects[0] != protocol.TopicAuditTrace {
		t.Fatalf("only the allowed event's audit trace may be published, got %v", subjects)
	}
}

// Audit down (error, empty id, or no sink yet): one structured log line, still
// nothing published and nothing parked.
func TestC2RetireAuditUnavailableLogsAndPublishesNothing(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })
	for name, sink := range map[string]*c2rAuditSink{
		"insert-error": {err: errors.New("audit down")},
		"empty-id":     {},
		"no-sink":      nil,
	} {
		t.Run(name, func(t *testing.T) {
			buf.Reset()
			nc := startTestNATS(t)
			published := c2rBusRecorder(t, nc)
			r := NewRouter(nc, c2rApprovalGuard())
			if sink != nil {
				r.SetApprovalAuditor(sink.record)
			}
			deliver(t, r, "swarm.team.alpha.signal.result", "w", "k8s.delete.cluster")
			if got := published(); len(got) != 0 {
				t.Fatalf("audit down must still publish nothing, got %v", got)
			}
			line := buf.String()
			if strings.Count(line, PolicyApprovalObservedAction+" audit_unavailable=true") != 1 {
				t.Fatalf("expected one audit-unavailable log line, got %q", line)
			}
		})
	}
}

// A degraded guard denies before policy evaluation: no observation is audited.
func TestC2RetireDegradedGuardAuditsNothing(t *testing.T) {
	nc := startTestNATS(t)
	r := NewRouter(nc, governance.NewDegradedGuard(errors.New("missing policy")))
	sink := &c2rAuditSink{id: "audit-1"}
	r.SetApprovalAuditor(sink.record)
	deliver(t, r, "swarm.team.alpha.signal.result", "w", "k8s.delete.cluster")
	if got := sink.observations(); len(got) != 0 {
		t.Fatalf("degraded guard must not audit observations, got %+v", got)
	}
}

// The audit-trace prefix the Router ignores comes from the protocol constant.
func TestC2RetireAuditSubjectPrefixFromProtocol(t *testing.T) {
	if auditSubjectPrefix != "swarm.audit." {
		t.Fatalf("auditSubjectPrefix = %q, want swarm.audit.", auditSubjectPrefix)
	}
}

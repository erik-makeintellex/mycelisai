package governance

import (
	"errors"
	"reflect"
	"testing"
)

// c2rApprovalGuard returns a loaded guard whose only non-ALLOW rule is a
// REQUIRE_APPROVAL on k8s.delete.cluster.
func c2rApprovalGuard() *Guard {
	g := NewDegradedGuard(errors.New("start degraded"))
	g.UpdatePolicyConfig(&PolicyConfig{
		Groups: []PolicyGroup{{
			Name: "g", Targets: []string{"*"},
			Rules: []PolicyRule{{Intent: "k8s.delete.cluster", Action: ActionRequireApproval}},
		}},
		Defaults: DefaultConfig{DefaultAction: ActionAllow},
	})
	return g
}

// C2-RETIRE: the in-memory approval queue is gone. The Guard has no parked
// state and no API to list, fetch or resolve a parked request. Reflection keeps
// this test compiling against the old Guard so it can be shown to fail there.
func TestC2RetireGuardHasNoParkedApprovalState(t *testing.T) {
	gt := reflect.TypeOf(Guard{})
	if _, ok := gt.FieldByName("PendingBuffer"); ok {
		t.Error("Guard must not carry a PendingBuffer")
	}
	pt := reflect.TypeOf(&Guard{})
	for _, name := range []string{"ListPending", "PendingRequest", "Resolve", "createApprovalRequest"} {
		if _, ok := pt.MethodByName(name); ok {
			t.Errorf("Guard must not expose %s", name)
		}
	}
	for i := 0; i < gt.NumField(); i++ {
		if f := gt.Field(i); f.Type.Kind() == reflect.Map && f.Type.Elem().String() == "*swarm.ApprovalRequest" {
			t.Errorf("Guard field %s parks approval requests", f.Name)
		}
	}
}

// REQUIRE_APPROVAL means "do not proceed" and nothing else: no request id.
func TestC2RetireInterceptRequireApprovalReturnsNoRequestID(t *testing.T) {
	g := c2rApprovalGuard()
	out := reflect.ValueOf(g).MethodByName("Intercept").Call(
		[]reflect.Value{reflect.ValueOf(eventEnvelope("alpha", "a", "k8s.delete.cluster"))})
	if len(out) != 2 {
		t.Fatalf("Intercept must return (proceed, action) only, got %d values", len(out))
	}
	if out[0].Bool() || out[1].String() != ActionRequireApproval {
		t.Fatalf("Intercept = %v %q, want false %q", out[0].Bool(), out[1].String(), ActionRequireApproval)
	}
	allowed := reflect.ValueOf(g).MethodByName("Intercept").Call(
		[]reflect.Value{reflect.ValueOf(eventEnvelope("alpha", "a", "task.completed"))})
	if !allowed[0].Bool() || allowed[1].String() != ActionAllow {
		t.Fatalf("an allowed intent must still proceed, got %v %q", allowed[0].Bool(), allowed[1].String())
	}
}

package governance

import (
	"errors"
	"sync"
	"testing"

	pb "github.com/mycelis/core/pkg/pb/swarm"
)

func eventEnvelope(team, agent, intent string) *pb.MsgEnvelope {
	return &pb.MsgEnvelope{
		TeamId:        team,
		SourceAgentId: agent,
		Payload:       &pb.MsgEnvelope_Event{Event: &pb.EventPayload{EventType: intent}},
	}
}

func allowAllConfig() *PolicyConfig {
	return &PolicyConfig{Defaults: DefaultConfig{DefaultAction: ActionAllow}}
}

func TestDegradedGuardFailsClosed(t *testing.T) {
	g := NewDegradedGuard(errors.New("open config/policy.yaml: no such file"))
	if !g.Degraded() || g.LoadError() != PolicyUnavailableCode {
		t.Fatalf("expected degraded guard with fixed code, got degraded=%v code=%q", g.Degraded(), g.LoadError())
	}
	for _, intent := range []string{"", "task.completed", "agent.heartbeat", "system.shutdown"} {
		proceed, action, reqID := g.Intercept(eventEnvelope("alpha", "agent-1", intent))
		if proceed || action != ActionDeny || reqID != "" {
			t.Fatalf("degraded Intercept(%q) = %v %q %q, want DENY", intent, proceed, action, reqID)
		}
	}
	if len(g.ListPending()) != 0 {
		t.Fatal("degraded guard must not park approval requests")
	}
	if required, _ := g.PostureRequiresApproval("delivery-posture-lean", nil, nil); !required {
		t.Fatal("degraded guard must require approval for posture work")
	}
	if g.GetPolicyConfig() != nil {
		t.Fatal("degraded guard must not expose a policy config")
	}
}

func TestNilGuardIsDegraded(t *testing.T) {
	var g *Guard
	if !g.Degraded() || g.LoadError() != PolicyUnavailableCode {
		t.Fatal("nil guard must report degraded")
	}
	if required, _ := g.PostureRequiresApproval("p", nil, nil); !required {
		t.Fatal("nil guard must require approval for posture work")
	}
}

func TestNilEngineEvaluateDenies(t *testing.T) {
	if got := (&Engine{}).Evaluate("t", "a", "x", nil); got != ActionDeny {
		t.Fatalf("Evaluate with nil config = %q, want DENY", got)
	}
}

func TestUpdatePolicyConfigClearsDegraded(t *testing.T) {
	g := NewDegradedGuard(errors.New("boom"))
	g.UpdatePolicyConfig(allowAllConfig())
	if g.Degraded() || g.LoadError() != "" {
		t.Fatal("valid policy must clear degraded state")
	}
	if proceed, action, _ := g.Intercept(eventEnvelope("alpha", "a", "task.completed")); !proceed || action != ActionAllow {
		t.Fatalf("recovered guard should allow by policy, got %v %q", proceed, action)
	}
}

func TestReplacePolicyPersistFailureLeavesMemoryUnchanged(t *testing.T) {
	g := NewDegradedGuard(errors.New("boom"))
	if err := g.ReplacePolicy(allowAllConfig(), func() error { return errors.New("disk full") }); err == nil {
		t.Fatal("expected persist error")
	}
	if !g.Degraded() {
		t.Fatal("failed persist must not swap memory")
	}
	if err := g.ReplacePolicy(nil, nil); err == nil {
		t.Fatal("nil config must be rejected")
	}
	if err := g.ReplacePolicy(allowAllConfig(), func() error { return nil }); err != nil || g.Degraded() {
		t.Fatalf("successful replace should clear degraded: err=%v", err)
	}
}

// Run with -race: policy swaps must not race Intercept or posture reads.
func TestGuardConcurrentPolicySwapIsRaceFree(t *testing.T) {
	g := NewDegradedGuard(errors.New("boom"))
	posture := &PolicyConfig{
		Groups: []PolicyGroup{{
			Name:    "posture",
			Targets: []string{"posture:p"},
			Rules:   []PolicyRule{{Intent: "^.*$", Action: ActionRequireApproval}},
		}},
		Defaults: DefaultConfig{DefaultAction: ActionAllow},
	}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(3)
		go func(i int) {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				if (i+j)%2 == 0 {
					g.UpdatePolicyConfig(posture)
				} else {
					_ = g.ReplacePolicy(allowAllConfig(), nil)
				}
			}
		}(i)
		go func() {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				g.Intercept(eventEnvelope("alpha", "a", "task.completed"))
				_ = g.Degraded()
			}
		}()
		go func() {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				g.PostureRequiresApproval("p", []string{"write_file"}, nil)
				_ = g.GetPolicyConfig()
			}
		}()
	}
	wg.Wait()
}

func TestValidatePolicyConfigRejectsFailOpenPolicies(t *testing.T) {
	deny := PolicyRule{Intent: "x", Action: ActionDeny}
	cases := map[string]*PolicyConfig{
		"empty":             {},
		"typo-default":      {Groups: []PolicyGroup{{Name: "g", Targets: []string{"*"}, Rules: []PolicyRule{deny}}}, Defaults: DefaultConfig{DefaultAction: "DENNY"}},
		"lowercase-default": {Groups: []PolicyGroup{{Name: "g", Targets: []string{"*"}, Rules: []PolicyRule{deny}}}, Defaults: DefaultConfig{DefaultAction: "deny"}},
		"typo-rule":         {Groups: []PolicyGroup{{Name: "g", Targets: []string{"*"}, Rules: []PolicyRule{{Intent: "x", Action: "deny"}}}}, Defaults: DefaultConfig{DefaultAction: ActionAllow}},
		"allow-only":        {Defaults: DefaultConfig{DefaultAction: ActionAllow}},
		"allow-rules-only":  {Groups: []PolicyGroup{{Name: "g", Targets: []string{"*"}, Rules: []PolicyRule{{Intent: "x", Action: ActionAllow}}}}, Defaults: DefaultConfig{DefaultAction: ActionAllow}},
	}
	for name, cfg := range cases {
		if err := ValidatePolicyConfig(cfg); err == nil {
			t.Errorf("%s: fail-open policy accepted", name)
		}
	}
	for name, cfg := range map[string]*PolicyConfig{
		"deny-default-no-rules": {Defaults: DefaultConfig{DefaultAction: ActionDeny}},
		"allow-with-deny-rule":  {Groups: []PolicyGroup{{Name: "g", Targets: []string{"*"}, Rules: []PolicyRule{deny}}}, Defaults: DefaultConfig{DefaultAction: ActionAllow}},
	} {
		if err := ValidatePolicyConfig(cfg); err != nil {
			t.Errorf("%s: explicit policy rejected: %v", name, err)
		}
	}
}

func TestInterceptDeniesUnknownAction(t *testing.T) {
	g := NewDegradedGuard(errors.New("start"))
	g.UpdatePolicyConfig(&PolicyConfig{Defaults: DefaultConfig{DefaultAction: "DENNY"}}) // bypasses validation on purpose
	if proceed, action, _ := g.Intercept(eventEnvelope("alpha", "a", "x")); proceed || action != ActionDeny {
		t.Fatalf("unknown action must fail closed, got %v %q", proceed, action)
	}
}

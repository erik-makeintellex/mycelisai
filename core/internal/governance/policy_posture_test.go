package governance

import (
	"os"
	"path/filepath"
	"testing"
)

func postureGroup(targets []string, rules ...PolicyRule) PolicyGroup {
	return PolicyGroup{Name: "posture", Targets: targets, Rules: rules}
}

func TestValidatePolicyConfigRejectsNonStricterPostureGroups(t *testing.T) {
	target := []string{"posture:delivery-posture-governed-enterprise"}
	cases := map[string]PolicyGroup{
		"allow action":      postureGroup(target, PolicyRule{Intent: "^.*$", Action: ActionAllow}),
		"deny action":       postureGroup(target, PolicyRule{Intent: "^.*$", Action: ActionDeny}),
		"condition":         postureGroup(target, PolicyRule{Intent: "^.*$", Condition: "amount > 5", Action: ActionRequireApproval}),
		"mixed targets":     postureGroup([]string{"posture:x-posture", "team:admin-core"}, PolicyRule{Intent: "^.*$", Action: ActionRequireApproval}),
		"wildcard target":   postureGroup([]string{"posture:x-posture", "*"}, PolicyRule{Intent: "^.*$", Action: ActionRequireApproval}),
		"empty posture id":  postureGroup([]string{"posture:"}, PolicyRule{Intent: "^.*$", Action: ActionRequireApproval}),
		"unanchored intent": postureGroup(target, PolicyRule{Intent: "write_file", Action: ActionRequireApproval}),
		"bare wildcard":     postureGroup(target, PolicyRule{Intent: "*", Action: ActionRequireApproval}),
		"invalid regex":     postureGroup(target, PolicyRule{Intent: "^(write$", Action: ActionRequireApproval}),
		"no rules":          postureGroup(target),
	}
	for name, group := range cases {
		t.Run(name, func(t *testing.T) {
			cfg := &PolicyConfig{Groups: []PolicyGroup{group}, Defaults: DefaultConfig{DefaultAction: ActionDeny}}
			if err := ValidatePolicyConfig(cfg); err == nil {
				t.Fatalf("posture group %q accepted", name)
			}
		})
	}
}

func TestNewEngineRejectsInvalidPostureGroupAndLoadsShippedPolicy(t *testing.T) {
	path := filepath.Join(t.TempDir(), "policy.yaml")
	bad := "groups:\n  - name: lower\n    targets: [\"posture:delivery-posture-governed-enterprise\"]\n    rules:\n      - intent: \"^.*$\"\n        action: \"ALLOW\"\ndefaults:\n  default_action: ALLOW\n"
	if err := os.WriteFile(path, []byte(bad), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := NewEngine(path); err == nil {
		t.Fatal("NewEngine accepted an ALLOW posture group")
	}
	engine, err := NewEngine("../../config/policy.yaml")
	if err != nil {
		t.Fatalf("shipped policy: %v", err)
	}
	cases := []struct {
		posture string
		tools   []string
		caps    []string
		want    bool
	}{
		{"delivery-posture-governed-enterprise", []string{"remember"}, []string{"learning"}, true},
		{"delivery-posture-governed-enterprise", nil, nil, true},
		{"delivery-posture-client-delivery-studio", []string{"write_file"}, []string{"file_output"}, true},
		{"delivery-posture-client-delivery-studio", []string{"delegate_task"}, []string{"team_orchestration"}, false},
		{"delivery-posture-product-delivery-team", []string{"delegate_task"}, nil, true},
		{"delivery-posture-product-delivery-team", []string{"write_file_extra"}, nil, false},
		{"delivery-posture-operations-desk", []string{"write_file"}, []string{"file_output"}, false},
		{"unknown-posture", []string{"write_file"}, nil, false},
		{"", []string{"write_file"}, nil, false},
	}
	for _, tc := range cases {
		if got, _ := engine.PostureRequiresApproval(tc.posture, tc.tools, tc.caps); got != tc.want {
			t.Errorf("PostureRequiresApproval(%q, %v, %v) = %v, want %v", tc.posture, tc.tools, tc.caps, got, tc.want)
		}
	}
}

func TestPostureGroupsDoNotChangeNATSEvaluate(t *testing.T) {
	base := &PolicyConfig{
		Groups: []PolicyGroup{{
			Name: "finance", Targets: []string{"team:finance"},
			Rules: []PolicyRule{{Intent: "payment.create", Condition: "amount > 50", Action: ActionRequireApproval}},
		}},
		Defaults: DefaultConfig{DefaultAction: ActionAllow},
	}
	withPosture := &PolicyConfig{
		Groups: append(append([]PolicyGroup(nil), base.Groups...), postureGroup(
			[]string{"posture:delivery-posture-governed-enterprise"},
			PolicyRule{Intent: "^.*$", Action: ActionRequireApproval},
		)),
		Defaults: base.Defaults,
	}
	if err := ValidatePolicyConfig(withPosture); err != nil {
		t.Fatal(err)
	}
	inputs := []struct {
		team, agent, intent string
		ctx                 map[string]interface{}
	}{
		{"finance", "bot", "payment.create", map[string]interface{}{"amount": 100.0}},
		{"finance", "bot", "payment.create", map[string]interface{}{"amount": 10.0}},
		{"posture", "delivery-posture-governed-enterprise", "anything", nil},
		{"delivery-posture-governed-enterprise", "x", "write_file", nil},
		{"admin-core", "admin", "k8s.delete.pod", nil},
	}
	for _, in := range inputs {
		want := (&Engine{Config: base}).Evaluate(in.team, in.agent, in.intent, in.ctx)
		got := (&Engine{Config: withPosture}).Evaluate(in.team, in.agent, in.intent, in.ctx)
		if got != want {
			t.Errorf("Evaluate(%q,%q,%q) = %q with posture group, want %q", in.team, in.agent, in.intent, got, want)
		}
	}
}

func TestGuardPostureRequiresApprovalFailsClosedWithoutPolicy(t *testing.T) {
	var guard *Guard
	if got, _ := guard.PostureRequiresApproval("delivery-posture-operations-desk", nil, nil); !got {
		t.Fatal("nil guard did not require approval")
	}
	if got, _ := (&Guard{Engine: &Engine{}}).PostureRequiresApproval("delivery-posture-operations-desk", nil, nil); !got {
		t.Fatal("guard without config did not require approval")
	}
}

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadGovernanceGuardStartsDegradedOnBadPolicy(t *testing.T) {
	dir := t.TempDir()
	malformed := filepath.Join(dir, "malformed.yaml")
	if err := os.WriteFile(malformed, []byte("groups: [unterminated"), 0o600); err != nil {
		t.Fatal(err)
	}
	invalid := filepath.Join(dir, "invalid.yaml")
	if err := os.WriteFile(invalid, []byte("groups:\n  - name: p\n    targets: [\"posture:x\"]\n    rules:\n      - intent: \"^.*$\"\n        action: ALLOW\ndefaults:\n  default_action: ALLOW\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	cases := map[string]string{
		"missing":   filepath.Join(dir, "absent.yaml"),
		"malformed": malformed,
		"invalid":   invalid,
	}
	for name, body := range map[string]string{
		"empty":        "",
		"typo-default": "groups:\n  - name: g\n    targets: [\"*\"]\n    rules:\n      - intent: x\n        action: DENY\ndefaults:\n  default_action: DENNY\n",
		"typo-rule":    "groups:\n  - name: g\n    targets: [\"*\"]\n    rules:\n      - intent: x\n        action: deny\ndefaults:\n  default_action: ALLOW\n",
		"allow-only":   "defaults:\n  default_action: ALLOW\n",
	} {
		path := filepath.Join(dir, name+".yaml")
		if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
		cases[name] = path
	}
	for name, path := range cases {
		t.Run(name, func(t *testing.T) {
			guard := loadGovernanceGuardFrom(path)
			if guard == nil {
				t.Fatal("loadGovernanceGuard must never return nil (nil used to mean allow-all)")
			}
			if !guard.Degraded() {
				t.Fatal("expected degraded guard")
			}
		})
	}
}

func TestLoadGovernanceGuardLoadsShippedPolicy(t *testing.T) {
	guard := loadGovernanceGuardFrom(filepath.Join("..", "..", "config", "policy.yaml"))
	if guard == nil || guard.Degraded() {
		t.Fatal("shipped policy must load as an active guard")
	}
}

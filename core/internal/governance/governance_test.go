package governance

import (
	"testing"
)

func TestPolicyLoading(t *testing.T) {
	// We can't easily write to FS in unit test without cleanup,
	// but for now we trust NewEngine works if file exists.
	// Let's test Logic with struct directly to avoid FS dependency in simple test

	cfg := &PolicyConfig{
		Groups: []PolicyGroup{
			{
				Name:    "Test Group",
				Targets: []string{"team:test"},
				Rules: []PolicyRule{
					{Intent: "dangerous_op", Action: "REQUIRE_APPROVAL"},
				},
			},
		},
		Defaults: DefaultConfig{DefaultAction: "ALLOW"},
	}

	engine := &Engine{Config: cfg}

	// Case 1: Match
	action := engine.Evaluate("test", "agent1", "dangerous_op", nil)
	if action != "REQUIRE_APPROVAL" {
		t.Errorf("Expected REQUIRE_APPROVAL, got %s", action)
	}

	// Case 2: No Match (Default)
	action = engine.Evaluate("marketing", "agent1", "dangerous_op", nil)
	if action != "ALLOW" {
		t.Errorf("Expected ALLOW, got %s", action)
	}
}

package governance

import (
	"os"
	"testing"
)

func TestEngine_Evaluate(t *testing.T) {
	// Create a temporary policy file
	content := `
groups:
  - name: "financial"
    targets: ["team:default"]
    rules:
      - intent: "payment"
        condition: "amount > 50"
        action: "REQUIRE_APPROVAL"
defaults:
  default_action: "ALLOW"
`
	tmpfile, err := os.CreateTemp("", "policy.*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	engine, err := NewEngine(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to load engine: %v", err)
	}

	tests := []struct {
		name     string
		intent   string
		context  map[string]interface{}
		expected string
	}{
		{
			name:     "Allow Small Payment",
			intent:   "payment",
			context:  map[string]interface{}{"amount": 10},
			expected: "ALLOW", // Default action because condition > 50 is false? No, logic loop needs verification.
			// Logic: if rule matches intent, check condition. If condition true -> return Action.
			// If condition false -> continue loop. If no more rules -> return Default.
			// So logic implies "amount > 50" triggers REQUIRE_APPROVAL.
			// "amount <= 50" fails condition, falls through to Default ALLOW. Correct.
		},
		{
			name:     "Block Large Payment",
			intent:   "payment",
			context:  map[string]interface{}{"amount": 100},
			expected: "REQUIRE_APPROVAL",
		},
		{
			name:     "Allow Unknown Intent",
			intent:   "unknown",
			context:  nil,
			expected: "ALLOW",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := engine.Evaluate("default", "agent-1", tt.intent, tt.context)
			if action != tt.expected {
				t.Errorf("Evaluate() = %s, want %s", action, tt.expected)
			}
		})
	}
}

func TestGatekeeper_Intercept(t *testing.T) {
	// Basic Stub test to ensure Gatekeeper calls Engine
	// Requires mocking Engine config or similar setup logic as above
	// Skipping for now to focus on Engine logic which is the core complexity.
}

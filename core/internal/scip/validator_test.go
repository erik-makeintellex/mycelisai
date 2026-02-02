package scip

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mycelis/core/pkg/scip"
)

func TestValidator(t *testing.T) {
	// 1. Setup Mock Contracts
	mockYaml := `
contracts:
  - id: "test.chat"
    allowed_types: ["TEXT_UTF8"]
    max_bytes: 10
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "contracts.yaml")
	if err := os.WriteFile(configPath, []byte(mockYaml), 0644); err != nil {
		t.Fatalf("failed to write mock config: %v", err)
	}

	// 2. Initialize Validator
	v, err := NewValidator(configPath)
	if err != nil {
		t.Fatalf("NewValidator failed: %v", err)
	}

	// 3. Test Cases
	tests := []struct {
		name      string
		envelope  *scip.SignalEnvelope
		wantError bool
	}{
		{
			name: "Valid Chat",
			envelope: &scip.SignalEnvelope{
				TraceId:  "123",
				Intent:   "test.chat",
				DataType: scip.DataType_TEXT_UTF8,
				Payload:  []byte("hello"),
			},
			wantError: false,
		},
		{
			name: "Invalid Intent (Unknown)",
			envelope: &scip.SignalEnvelope{
				TraceId:  "123",
				Intent:   "unknown.intent",
				DataType: scip.DataType_TEXT_UTF8,
				Payload:  []byte("hello"),
			},
			wantError: true,
		},
		{
			name: "Invalid Type (Image forbidden)",
			envelope: &scip.SignalEnvelope{
				TraceId:  "123",
				Intent:   "test.chat",
				DataType: scip.DataType_IMAGE_BINARY,
				Payload:  []byte("fake_image"),
			},
			wantError: true,
		},
		{
			name: "Oversized Payload",
			envelope: &scip.SignalEnvelope{
				TraceId:  "123",
				Intent:   "test.chat",
				DataType: scip.DataType_TEXT_UTF8,
				Payload:  []byte("0123456789A"), // 11 bytes > 10
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Validate(tt.envelope)
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

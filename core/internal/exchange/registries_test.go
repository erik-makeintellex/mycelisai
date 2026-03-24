package exchange

import "testing"

func TestSeedSchemasReferenceKnownFields(t *testing.T) {
	for _, schema := range SeedSchemas {
		for _, fieldName := range append(append([]string{}, schema.RequiredFields...), schema.OptionalFields...) {
			if _, ok := FieldByName(fieldName); !ok {
				t.Fatalf("schema %s references unknown field %s", schema.ID, fieldName)
			}
		}
		for _, capabilityID := range schema.RequiredCapabilities {
			if _, ok := CapabilityByID(capabilityID); !ok {
				t.Fatalf("schema %s references unknown capability %s", schema.ID, capabilityID)
			}
		}
	}
}

func TestValidatePayload(t *testing.T) {
	valid := map[string]any{
		"summary":     "Research completed.",
		"status":      "completed",
		"source_role": "mcp:fetch",
		"target_role": "soma",
		"created_at":  "2026-03-24T07:00:00Z",
		"review_required": true,
	}
	if err := validatePayload("ToolResult", valid); err != nil {
		t.Fatalf("validatePayload(valid) error = %v", err)
	}

	invalid := map[string]any{
		"summary":     "Research completed.",
		"status":      "completed",
		"source_role": "mcp:fetch",
	}
	if err := validatePayload("ToolResult", invalid); err == nil {
		t.Fatal("validatePayload(invalid) expected error")
	}
}

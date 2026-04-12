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
		"summary":         "Research completed.",
		"status":          "completed",
		"source_role":     "mcp:fetch",
		"target_role":     "soma",
		"created_at":      "2026-03-24T07:00:00Z",
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

func TestLearningCandidateRequiresClassificationAndReviewPosture(t *testing.T) {
	valid := map[string]any{
		"summary":         "The operator's priority shifted toward investor demo readiness.",
		"status":          "candidate",
		"classification":  "trajectory_shift",
		"memory_layer":    "REFLECTION_MEMORY",
		"confidence":      0.82,
		"review_required": true,
		"tags":            []string{"reflection", "trajectory_shift"},
		"continuity_key":  "operator-investor-demo",
		"created_at":      "2026-04-11T18:45:00Z",
	}
	if err := validatePayload("LearningCandidate", valid); err != nil {
		t.Fatalf("validatePayload(valid learning candidate) error = %v", err)
	}

	invalid := map[string]any{
		"summary":        "The operator's priority shifted toward investor demo readiness.",
		"status":         "candidate",
		"confidence":     0.82,
		"tags":           []string{"reflection", "trajectory_shift"},
		"continuity_key": "operator-investor-demo",
		"created_at":     "2026-04-11T18:45:00Z",
	}
	if err := validatePayload("LearningCandidate", invalid); err == nil {
		t.Fatal("validatePayload(invalid learning candidate) expected error")
	}
}

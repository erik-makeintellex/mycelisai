package protocol

import (
	"encoding/json"
	"testing"
)

func TestChatResponsePayload_ExecutionSummaryIsAdditive(t *testing.T) {
	verified := true
	payload := ChatResponsePayload{
		Text:      "Readable answer",
		Artifacts: []ChatArtifactRef{{ID: "artifact-1", Type: "document", Title: "Brief"}},
		Provenance: &AnswerProvenance{
			ResolvedIntent:  "answer",
			PermissionCheck: "pass",
			PolicyDecision:  "allow",
			AuditEventID:    "audit-1",
		},
		ExecutionSummary: &ExecutionSummary{
			Intent: ExecutionIntent{
				Original: "Explain the brief",
				Resolved: "answer",
			},
			Understanding: ExecutionUnderstanding{Summary: "Readable answer"},
			Execution: ExecutionState{
				Shape:   ExecutionShapeDirectSoma,
				Status:  ExecutionStatusCompleted,
				Summary: "Soma completed a direct response.",
			},
			Outputs: []ExecutionOutput{{
				ID:             "artifact-1",
				Kind:           "document",
				Title:          "Brief",
				Retained:       &verified,
				RetentionClass: ExecutionRetentionClassRetained,
			}},
			Proof: ExecutionProof{
				RunClass:     ExecutionRunClassNoRun,
				NoRunReason:  "direct answer",
				ProofClass:   ExecutionProofClassAuditOnly,
				AuditEventID: "audit-1",
				Verified:     &verified,
			},
		},
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	for _, key := range []string{"text", "artifacts", "provenance", "execution_summary"} {
		if _, ok := decoded[key]; !ok {
			t.Fatalf("missing additive payload key %q in %s", key, string(raw))
		}
	}

	summary := decoded["execution_summary"].(map[string]any)
	execution := summary["execution"].(map[string]any)
	if execution["shape"] != string(ExecutionShapeDirectSoma) {
		t.Fatalf("execution.shape = %v", execution["shape"])
	}
	if execution["status"] != string(ExecutionStatusCompleted) {
		t.Fatalf("execution.status = %v", execution["status"])
	}
	proof := summary["proof"].(map[string]any)
	if proof["run_class"] != string(ExecutionRunClassNoRun) || proof["proof_class"] != string(ExecutionProofClassAuditOnly) {
		t.Fatalf("proof classification = %+v", proof)
	}
	outputs := summary["outputs"].([]any)
	output := outputs[0].(map[string]any)
	if output["retention_class"] != string(ExecutionRetentionClassRetained) {
		t.Fatalf("output retention_class = %v", output["retention_class"])
	}
}

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
			ContractID: "contract-1",
			ProofID:    "proof-artifact-1",
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
				ContractID:   "contract-1",
				ProofID:      "proof-artifact-1",
				RunClass:     ExecutionRunClassNoRun,
				NoRunReason:  "direct answer",
				ProofClass:   ExecutionProofClassAuditOnly,
				AuditEventID: "audit-1",
				Verified:     &verified,
			},
			AuditRecovery: AuditRecovery{
				RecoveryState: "blocked",
				Degradation: &ExecutionDegradation{
					Code:              "search_provider_disabled",
					WhatFailed:        "Search provider disabled.",
					TrustedState:      "The audit record remains trusted.",
					InvalidatedProof:  "No search proof is available.",
					SafeContinuation:  "Configure search and retry.",
					RequiresAttention: true,
				},
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
	if summary["contract_id"] != "contract-1" || summary["proof_id"] != "proof-artifact-1" || proof["contract_id"] != "contract-1" || proof["proof_id"] != "proof-artifact-1" {
		t.Fatalf("trust object ids missing from summary/proof: summary=%+v proof=%+v", summary, proof)
	}
	outputs := summary["outputs"].([]any)
	output := outputs[0].(map[string]any)
	if output["retention_class"] != string(ExecutionRetentionClassRetained) {
		t.Fatalf("output retention_class = %v", output["retention_class"])
	}
	auditRecovery := summary["audit_recovery"].(map[string]any)
	degradation := auditRecovery["degradation"].(map[string]any)
	if degradation["code"] != "search_provider_disabled" || degradation["requires_attention"] != true {
		t.Fatalf("degradation = %+v", degradation)
	}
}

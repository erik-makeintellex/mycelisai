package memory

import (
	"testing"
)

func TestArbitrateVectorResults_PrecedenceAndEvidenceAnchors(t *testing.T) {
	origMeta := map[string]any{
		"knowledge_class": "company_knowledge",
		"artifact_id":     "art-doc-1",
		"source_kind":     "governance_policy",
		"trust_class":     "trusted_internal",
	}

	input := []VectorResult{
		{
			ID:      "vec-reflection",
			Content: "Inferred pattern from recent chat",
			Score:   0.99,
			Metadata: map[string]any{
				"knowledge_class": "reflection_synthesis",
				"reflection_kind": "lesson",
				"source_kind":     "synthesis_note",
			},
		},
		{
			ID:      "vec-soma",
			Content: "User prefers concise bullets",
			Score:   0.95,
			Metadata: map[string]any{
				"knowledge_class": "user_private_context",
				"artifact_id":     "art-soma-1",
				"source_kind":     "user_record",
				"trust_class":     "user_provided",
			},
		},
		{
			ID:      "vec-agent",
			Content: "Team alpha deployment checklist",
			Score:   0.91,
			Metadata: map[string]any{
				"type":       "sitrep",
				"team_id":    "team-alpha",
				"run_id":     "run-101",
				"visibility": "team",
			},
		},
		{
			ID:      "vec-project",
			Content: "Customer scope brief",
			Score:   0.88,
			Metadata: map[string]any{
				"knowledge_class": "customer_context",
				"artifact_id":     "art-proj-1",
				"source_kind":     "user_document",
				"trust_class":     "customer_provided",
			},
		},
		{
			ID:       "vec-doctrine",
			Content:  "Approved architecture boundary rule",
			Score:    0.82,
			Metadata: origMeta,
		},
	}

	got, summary := ArbitrateVectorResults(input, 10, 3)
	if len(got) != 5 {
		t.Fatalf("expected 5 results, got %d", len(got))
	}

	wantOrder := []struct {
		id     string
		layer  MemoryLayer
		status MemoryVerificationStatus
	}{
		{"vec-doctrine", LayerGovernedDoctrine, VerificationVerified},
		{"vec-project", LayerProjectMemory, VerificationVerified},
		{"vec-agent", LayerAgentMemory, VerificationVerified},
		{"vec-soma", LayerSomaMemory, VerificationVerified},
		{"vec-reflection", LayerReflectionMemory, VerificationCandidate},
	}

	for i, want := range wantOrder {
		if got[i].ID != want.id {
			t.Fatalf("pos %d: expected ID %s, got %s", i, want.id, got[i].ID)
		}
		if got[i].Metadata["memory_layer"] != string(want.layer) {
			t.Fatalf("pos %d: expected layer %s, got %v", i, want.layer, got[i].Metadata["memory_layer"])
		}
		if got[i].Metadata["verification_status"] != string(want.status) {
			t.Fatalf("pos %d: expected status %s, got %v", i, want.status, got[i].Metadata["verification_status"])
		}
		anchor, ok := got[i].Metadata["evidence_anchor"].(EvidenceAnchor)
		if !ok {
			t.Fatalf("pos %d: expected EvidenceAnchor in metadata, got %T", i, got[i].Metadata["evidence_anchor"])
		}
		if anchor.Verified != (want.status == VerificationVerified) {
			t.Fatalf("pos %d: anchor.Verified = %v, want %v", i, anchor.Verified, want.status == VerificationVerified)
		}
	}

	if _, mutated := origMeta["memory_layer"]; mutated {
		t.Fatalf("expected original input metadata map to remain unmutated")
	}
	if summary.HaltRecommended || summary.SupersededFiltered != 0 || summary.OverflowTrimmed != 0 {
		t.Fatalf("unexpected summary flags: %+v", summary)
	}
}

func TestArbitrateVectorResults_SupersessionAndContradictionHalts(t *testing.T) {
	input := []VectorResult{
		{
			ID:      "vec-old",
			Content: "Outdated release guidance",
			Score:   0.96,
			Metadata: map[string]any{
				"knowledge_class": "company_knowledge",
				"superseded_by":   "vec-current",
			},
		},
		{
			ID:      "vec-status-superseded",
			Content: "Legacy team note",
			Score:   0.92,
			Metadata: map[string]any{
				"type":                "agent_memory",
				"verification_status": "superseded",
			},
		},
		{
			ID:      "vec-contradicted",
			Content: "Unverified claim that port 9999 is canonical",
			Score:   0.97,
			Metadata: map[string]any{
				"knowledge_class":      "reflection_synthesis",
				"contradicts_id":       "vec-current",
				"contradiction_reason": "conflicts with governed port policy",
			},
		},
		{
			ID:      "vec-current",
			Content: "Core API runs on governed port 8081",
			Score:   0.89,
			Metadata: map[string]any{
				"knowledge_class": "company_knowledge",
				"artifact_id":     "art-gov-2",
			},
		},
	}

	got, summary := ArbitrateVectorResults(input, 5, 3)
	if len(got) != 1 || got[0].ID != "vec-current" {
		t.Fatalf("expected only vec-current to survive arbitration, got %+v", got)
	}
	if summary.SupersededFiltered != 2 {
		t.Fatalf("expected 2 superseded vectors filtered, got %d", summary.SupersededFiltered)
	}
	if !summary.HaltRecommended || len(summary.ContradictionHalts) != 1 {
		t.Fatalf("expected 1 contradiction halt and HaltRecommended=true, got %+v", summary)
	}
	halt := summary.ContradictionHalts[0]
	if halt.VectorID != "vec-contradicted" || halt.ContradictsID != "vec-current" || halt.Layer != LayerReflectionMemory {
		t.Fatalf("unexpected contradiction halt payload: %+v", halt)
	}
}

func TestArbitrateVectorResults_BoundedGrowthAndScoreTieBreak(t *testing.T) {
	input := []VectorResult{
		{ID: "doc-low", Score: 0.70, Metadata: map[string]any{"knowledge_class": "company_knowledge"}},
		{ID: "doc-high", Score: 0.95, Metadata: map[string]any{"knowledge_class": "company_knowledge"}},
		{ID: "doc-mid", Score: 0.85, Metadata: map[string]any{"knowledge_class": "company_knowledge"}},
		{ID: "doc-extra", Score: 0.60, Metadata: map[string]any{"knowledge_class": "company_knowledge"}},
		{ID: "proj-1", Score: 0.90, Metadata: map[string]any{"knowledge_class": "customer_context", "artifact_id": "a1"}},
		{ID: "proj-2", Score: 0.80, Metadata: map[string]any{"knowledge_class": "customer_context", "artifact_id": "a2"}},
	}

	// Default maxPerLayer (<=0 -> 3), maxTotal = 4
	got, summary := ArbitrateVectorResults(input, 4, 0)
	if len(got) != 4 {
		t.Fatalf("expected 4 results after caps, got %d", len(got))
	}
	wantIDs := []string{"doc-high", "doc-mid", "doc-low", "proj-1"}
	for i, wantID := range wantIDs {
		if got[i].ID != wantID {
			t.Fatalf("pos %d: expected %s, got %s", i, wantID, got[i].ID)
		}
	}
	if summary.LayerCounts[string(LayerGovernedDoctrine)] != 3 {
		t.Fatalf("expected 3 GOVERNED_DOCTRINE results, got %d", summary.LayerCounts[string(LayerGovernedDoctrine)])
	}
	if summary.LayerCounts[string(LayerProjectMemory)] != 1 {
		t.Fatalf("expected 1 PROJECT_MEMORY result, got %d", summary.LayerCounts[string(LayerProjectMemory)])
	}
	if summary.OverflowTrimmed != 2 {
		t.Fatalf("expected 2 overflow-trimmed results (doc-extra + proj-2), got %d", summary.OverflowTrimmed)
	}
}

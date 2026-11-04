package memory

import (
	"fmt"
	"sort"
	"strings"
)

// MemoryLayer identifies the canonical trusted-memory taxonomy tier.
type MemoryLayer string

const (
	LayerGovernedDoctrine MemoryLayer = "GOVERNED_DOCTRINE"
	LayerProjectMemory    MemoryLayer = "PROJECT_MEMORY"
	LayerAgentMemory      MemoryLayer = "AGENT_MEMORY"
	LayerSomaMemory       MemoryLayer = "SOMA_MEMORY"
	LayerReflectionMemory MemoryLayer = "REFLECTION_MEMORY"
)

// MemoryVerificationStatus tracks retrieval trust and lifecycle state for a memory vector.
type MemoryVerificationStatus string

const (
	VerificationVerified     MemoryVerificationStatus = "verified"
	VerificationCandidate    MemoryVerificationStatus = "candidate"
	VerificationContradicted MemoryVerificationStatus = "contradicted"
	VerificationSuperseded   MemoryVerificationStatus = "superseded"
)

// EvidenceAnchor links a recalled vector to deterministic execution or artifact lineage.
type EvidenceAnchor struct {
	ArtifactID string `json:"artifact_id,omitempty"`
	RunID      string `json:"run_id,omitempty"`
	SourceKind string `json:"source_kind,omitempty"`
	TrustClass string `json:"trust_class,omitempty"`
	Verified   bool   `json:"verified"`
}

// ContradictionHalt records a vector excluded from active recall due to contradiction metadata.
type ContradictionHalt struct {
	VectorID      string      `json:"vector_id"`
	ContradictsID string      `json:"contradicts_id,omitempty"`
	Layer         MemoryLayer `json:"layer"`
	Reason        string      `json:"reason"`
}

// ArbitrationSummary reports the retrieval-time control plane decisions applied to a search set.
type ArbitrationSummary struct {
	PrecedenceOrder    []MemoryLayer       `json:"precedence_order"`
	LayerCounts        map[string]int      `json:"layer_counts"`
	SupersededFiltered int                 `json:"superseded_filtered"`
	OverflowTrimmed    int                 `json:"overflow_trimmed"`
	ContradictionHalts []ContradictionHalt `json:"contradiction_halts,omitempty"`
	HaltRecommended    bool                `json:"halt_recommended"`
}

// CanonicalPrecedenceOrder returns the deterministic layer precedence order.
func CanonicalPrecedenceOrder() []MemoryLayer {
	return []MemoryLayer{
		LayerGovernedDoctrine,
		LayerProjectMemory,
		LayerAgentMemory,
		LayerSomaMemory,
		LayerReflectionMemory,
	}
}

// ArbitrateVectorResults classifies, filters, orders, and bounds recalled vectors.
func ArbitrateVectorResults(results []VectorResult, maxTotal int, maxPerLayer int) ([]VectorResult, ArbitrationSummary) {
	if maxPerLayer <= 0 {
		maxPerLayer = 3
	}
	if maxTotal <= 0 {
		maxTotal = len(results)
	}

	summary := ArbitrationSummary{
		PrecedenceOrder: CanonicalPrecedenceOrder(),
		LayerCounts:     make(map[string]int),
	}
	if len(results) == 0 {
		return []VectorResult{}, summary
	}

	type classifiedItem struct {
		result VectorResult
		layer  MemoryLayer
		status MemoryVerificationStatus
		index  int
	}

	candidates := make([]classifiedItem, 0, len(results))
	for idx, item := range results {
		clonedMeta := cloneMetadata(item.Metadata)
		layer := classifyMemoryLayer(clonedMeta)
		status := classifyVerificationStatus(clonedMeta, layer)
		anchor := buildEvidenceAnchor(clonedMeta, status)

		clonedMeta["memory_layer"] = string(layer)
		clonedMeta["verification_status"] = string(status)
		clonedMeta["evidence_anchor"] = anchor
		item.Metadata = clonedMeta

		if status == VerificationSuperseded || metaString(clonedMeta, "superseded_by") != "" {
			summary.SupersededFiltered++
			continue
		}

		contradictsID := metaString(clonedMeta, "contradicts_id")
		if status == VerificationContradicted || contradictsID != "" {
			reason := metaString(clonedMeta, "contradiction_reason")
			if reason == "" {
				reason = metaString(clonedMeta, "reason")
			}
			if reason == "" {
				if contradictsID != "" {
					reason = fmt.Sprintf("contradicts vector %s", contradictsID)
				} else {
					reason = "contradicted memory claim excluded from active recall"
				}
			}
			summary.ContradictionHalts = append(summary.ContradictionHalts, ContradictionHalt{
				VectorID:      item.ID,
				ContradictsID: contradictsID,
				Layer:         layer,
				Reason:        reason,
			})
			summary.HaltRecommended = true
			continue
		}

		candidates = append(candidates, classifiedItem{
			result: item,
			layer:  layer,
			status: status,
			index:  idx,
		})
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		a, b := candidates[i], candidates[j]
		rankA := precedenceRank(a.layer, a.status)
		rankB := precedenceRank(b.layer, b.status)
		if rankA != rankB {
			return rankA < rankB
		}
		if a.result.Score != b.result.Score {
			return a.result.Score > b.result.Score
		}
		return a.index < b.index
	})

	accepted := make([]VectorResult, 0, len(candidates))
	for _, cand := range candidates {
		layerKey := string(cand.layer)
		if summary.LayerCounts[layerKey] >= maxPerLayer || len(accepted) >= maxTotal {
			summary.OverflowTrimmed++
			continue
		}
		accepted = append(accepted, cand.result)
		summary.LayerCounts[layerKey]++
	}

	return accepted, summary
}

func classifyMemoryLayer(meta map[string]any) MemoryLayer {
	switch strings.ToUpper(metaString(meta, "memory_layer")) {
	case string(LayerGovernedDoctrine):
		return LayerGovernedDoctrine
	case string(LayerProjectMemory):
		return LayerProjectMemory
	case string(LayerAgentMemory):
		return LayerAgentMemory
	case string(LayerSomaMemory):
		return LayerSomaMemory
	case string(LayerReflectionMemory):
		return LayerReflectionMemory
	}

	kc := strings.ToLower(metaString(meta, "knowledge_class"))
	vt := strings.ToLower(metaString(meta, "type"))
	switch {
	case metaString(meta, "reflection_kind") != "" || kc == "reflection_synthesis" || vt == "reflection_synthesis" || vt == "reflection_memory" || vt == "learning_candidate":
		return LayerReflectionMemory
	case kc == "company_knowledge" || kc == "soma_operating_context" || kc == "governed_doctrine" || vt == "company_knowledge" || vt == "soma_operating_context" || vt == "governed_doctrine" || metaString(meta, "soma_context_kind") != "":
		return LayerGovernedDoctrine
	case kc == "customer_context" || kc == "project_memory" || vt == "customer_context" || vt == "project_memory" || vt == "project_context":
		return LayerProjectMemory
	case kc == "agent_memory" || vt == "agent_memory" || vt == "sitrep" || vt == "team_context" || strings.EqualFold(metaString(meta, "visibility"), "team") || metaString(meta, "team_id") != "":
		return LayerAgentMemory
	default:
		return LayerSomaMemory
	}
}

func classifyVerificationStatus(meta map[string]any, layer MemoryLayer) MemoryVerificationStatus {
	if metaString(meta, "superseded_by") != "" || metaBool(meta, "superseded") {
		return VerificationSuperseded
	}
	if metaString(meta, "contradicts_id") != "" || metaBool(meta, "contradicted") {
		return VerificationContradicted
	}

	switch strings.ToLower(metaString(meta, "verification_status")) {
	case string(VerificationSuperseded):
		return VerificationSuperseded
	case string(VerificationContradicted):
		return VerificationContradicted
	case string(VerificationVerified):
		return VerificationVerified
	case string(VerificationCandidate):
		return VerificationCandidate
	}

	if v, ok := meta["verified"].(bool); ok {
		if v {
			return VerificationVerified
		}
		return VerificationCandidate
	}
	if layer == LayerReflectionMemory {
		return VerificationCandidate
	}
	if layer == LayerGovernedDoctrine || metaString(meta, "artifact_id") != "" || metaString(meta, "run_id") != "" || metaString(meta, "trust_class") == "trusted_internal" {
		return VerificationVerified
	}
	return VerificationCandidate
}

func buildEvidenceAnchor(meta map[string]any, status MemoryVerificationStatus) EvidenceAnchor {
	sourceKind := metaString(meta, "source_kind")
	if sourceKind == "" {
		sourceKind = metaString(meta, "source")
	}
	return EvidenceAnchor{
		ArtifactID: metaString(meta, "artifact_id"),
		RunID:      metaString(meta, "run_id"),
		SourceKind: sourceKind,
		TrustClass: metaString(meta, "trust_class"),
		Verified:   status == VerificationVerified,
	}
}

func precedenceRank(layer MemoryLayer, status MemoryVerificationStatus) int {
	base := 40
	switch layer {
	case LayerGovernedDoctrine:
		base = 10
	case LayerProjectMemory:
		base = 20
	case LayerAgentMemory:
		base = 30
	case LayerSomaMemory:
		base = 40
	case LayerReflectionMemory:
		base = 50
	}
	if status != VerificationVerified {
		base += 5
	}
	return base
}

func cloneMetadata(src map[string]any) map[string]any {
	out := make(map[string]any, len(src)+3)
	for k, v := range src {
		out[k] = v
	}
	return out
}

func metaString(meta map[string]any, key string) string {
	if meta == nil {
		return ""
	}
	raw, ok := meta[key]
	if !ok || raw == nil {
		return ""
	}
	if s, ok := raw.(string); ok {
		return strings.TrimSpace(s)
	}
	return strings.TrimSpace(fmt.Sprintf("%v", raw))
}

func metaBool(meta map[string]any, key string) bool {
	if meta == nil {
		return false
	}
	b, ok := meta[key].(bool)
	return ok && b
}

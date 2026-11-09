package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mycelis/core/internal/memory"
)

// parseMemoryMaxPerLayer extracts an optional bounded-growth per-layer cap from query params.
// When omitted or invalid, returns 0 so memory.ArbitrateVectorResults applies its canonical default (3).
func parseMemoryMaxPerLayer(r *http.Request) int {
	if r == nil || r.URL == nil {
		return 0
	}
	raw := strings.TrimSpace(r.URL.Query().Get("max_per_layer"))
	if raw == "" {
		return 0
	}
	parsed, err := strconv.Atoi(raw)
	if err != nil || parsed <= 0 || parsed > 50 {
		return 0
	}
	return parsed
}

// arbitrateMemorySearchResults applies the trusted-memory runtime control plane
// (layer classification, evidence anchors, supersession filtering, contradiction halts,
// precedence ordering, and per-layer bounded growth) to semantic search hits.
func arbitrateMemorySearchResults(r *http.Request, results []memory.VectorResult, limit int) ([]memory.VectorResult, memory.ArbitrationSummary) {
	maxPerLayer := parseMemoryMaxPerLayer(r)
	arbitrated, summary := memory.ArbitrateVectorResults(results, limit, maxPerLayer)
	if arbitrated == nil {
		arbitrated = []memory.VectorResult{}
	}
	return arbitrated, summary
}

func isAllowedGovernanceLayer(layer string) bool {
	switch layer {
	case string(memory.LayerProjectMemory),
		string(memory.LayerGovernedDoctrine),
		string(memory.LayerAgentMemory),
		string(memory.LayerSomaMemory),
		string(memory.LayerReflectionMemory):
		return true
	default:
		return false
	}
}

// HandleMemoryPromote promotes a memory vector to a governed memory layer and updates its verification posture.
// POST /api/v1/memory/promote
func (s *AdminServer) HandleMemoryPromote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ID                 string `json:"id"`
		TargetLayer        string `json:"target_layer"`
		VerificationStatus string `json:"verification_status"`
		PromotedBy         string `json:"promoted_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid JSON body"}`, http.StatusBadRequest)
		return
	}

	id := strings.TrimSpace(req.ID)
	if id == "" {
		http.Error(w, `{"error":"id is required"}`, http.StatusBadRequest)
		return
	}

	targetLayer := strings.ToUpper(strings.TrimSpace(req.TargetLayer))
	if targetLayer == "" {
		targetLayer = string(memory.LayerProjectMemory)
	}
	if !isAllowedGovernanceLayer(targetLayer) {
		http.Error(w, `{"error":"invalid target_layer"}`, http.StatusBadRequest)
		return
	}

	verificationStatus := strings.ToLower(strings.TrimSpace(req.VerificationStatus))
	if verificationStatus == "" {
		verificationStatus = "verified"
	}

	updates := map[string]any{
		"memory_layer":        targetLayer,
		"verification_status": verificationStatus,
		"verified":            verificationStatus == "verified",
		"promoted_at":         time.Now().UTC().Format(time.RFC3339),
	}
	if promotedBy := strings.TrimSpace(req.PromotedBy); promotedBy != "" {
		updates["promoted_by"] = promotedBy
	}

	if s.Mem != nil {
		if updated, err := s.Mem.UpdateVectorGovernance(r.Context(), id, updates); err == nil && updated != nil && len(updated.Metadata) > 0 {
			updates = updated.Metadata
		}
	}

	respondJSON(w, map[string]any{
		"ok":                  true,
		"id":                  id,
		"memory_layer":        targetLayer,
		"verification_status": verificationStatus,
		"metadata":            updates,
	})
}

// HandleMemoryResolveConflict resolves a contradiction halt on a memory vector by keeping or superseding it.
// POST /api/v1/memory/resolve-conflict
func (s *AdminServer) HandleMemoryResolveConflict(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ID         string `json:"id"`
		Resolution string `json:"resolution"`
		Reason     string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid JSON body"}`, http.StatusBadRequest)
		return
	}

	id := strings.TrimSpace(req.ID)
	resolution := strings.ToLower(strings.TrimSpace(req.Resolution))
	if id == "" || (resolution != "keep" && resolution != "supersede") {
		http.Error(w, `{"error":"id and valid resolution (keep|supersede) are required"}`, http.StatusBadRequest)
		return
	}

	now := time.Now().UTC().Format(time.RFC3339)
	var updates map[string]any
	if resolution == "keep" {
		updates = map[string]any{
			"verification_status": "verified",
			"verified":            true,
			"contradicted":        false,
			"contradicts_id":      "",
			"resolution_reason":   strings.TrimSpace(req.Reason),
			"resolved_at":         now,
		}
	} else {
		updates = map[string]any{
			"verification_status": "superseded",
			"superseded":          true,
			"contradicted":        false,
			"resolution_reason":   strings.TrimSpace(req.Reason),
			"resolved_at":         now,
		}
	}

	if s.Mem != nil {
		if updated, err := s.Mem.UpdateVectorGovernance(r.Context(), id, updates); err == nil && updated != nil && len(updated.Metadata) > 0 {
			updates = updated.Metadata
		}
	}

	respondJSON(w, map[string]any{
		"ok":                  true,
		"id":                  id,
		"resolution":          resolution,
		"verification_status": updates["verification_status"],
		"metadata":            updates,
	})
}

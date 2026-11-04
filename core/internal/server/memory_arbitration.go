package server

import (
	"net/http"
	"strconv"
	"strings"

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

package memory

import (
	"context"
	"fmt"
	"strings"
)

// Recall modes reported by RecallGoverned and carried on each hit.
const (
	RecallModeSemantic = "semantic"
	RecallModeKeyword  = "keyword"
	RecallModeHybrid   = "hybrid"
)

// Embedder is the embedding surface recall needs. cognitive.Router satisfies
// it; a nil *Router reports unavailable.
type Embedder interface {
	EmbeddingAvailable(ctx context.Context) bool
	Embed(ctx context.Context, text string, model string) ([]float64, error)
}

// RecallGoverned is the single recall entry point. With a working embedding
// engine it merges semantic hits over embedded rows with keyword hits over
// rows still pending an embedding (hybrid); otherwise it is keyword only.
// Every leg applies the same scope clauses from opts.
func (s *Service) RecallGoverned(ctx context.Context, embedder Embedder, query string, opts SemanticSearchOptions) ([]VectorResult, string, error) {
	if s == nil || s.db == nil {
		return nil, "", fmt.Errorf("memory service offline")
	}
	if opts.Limit <= 0 {
		opts.Limit = 5
	}
	if embedder != nil && embedder.EmbeddingAvailable(ctx) {
		if vec, err := embedder.Embed(ctx, query, ""); err == nil {
			semantic, err := s.SemanticSearchWithOptions(ctx, vec, opts)
			if err != nil {
				return nil, "", err
			}
			pending, err := s.keywordSearch(ctx, query, opts, true)
			if err != nil {
				return nil, "", err
			}
			merged := mergeRecall(opts.Limit, semantic, pending)
			for _, hit := range merged {
				if hit.RetrievalMode == RecallModeKeyword {
					return merged, RecallModeHybrid, nil
				}
			}
			return merged, RecallModeSemantic, nil
		}
	}
	results, err := s.keywordSearch(ctx, query, opts, false)
	if err != nil {
		return nil, "", err
	}
	return mergeRecall(opts.Limit, results), RecallModeKeyword, nil
}

// mergeRecall keeps the first hit per source artifact (or per row when a hit
// has no artifact) in leg order, bounded by limit.
func mergeRecall(limit int, legs ...[]VectorResult) []VectorResult {
	seen := map[string]bool{}
	out := []VectorResult{}
	for _, leg := range legs {
		for _, hit := range leg {
			key := "row:" + hit.ID
			if artifactID, _ := hit.Metadata["artifact_id"].(string); strings.TrimSpace(artifactID) != "" {
				key = "artifact:" + artifactID
			}
			if seen[key] || len(out) >= limit {
				continue
			}
			seen[key] = true
			out = append(out, hit)
		}
	}
	return out
}

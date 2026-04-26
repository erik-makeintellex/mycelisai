package swarm

import (
	"context"
	"fmt"

	"github.com/mycelis/core/internal/searchcap"
)

func (r *InternalToolRegistry) handleWebSearch(ctx context.Context, args map[string]any) (string, error) {
	query := stringValue(args["query"])
	if query == "" {
		return "", fmt.Errorf("web_search requires 'query'")
	}
	if r.search == nil {
		resp, _ := searchcap.NewService(searchcap.Config{Provider: searchcap.ProviderDisabled}, nil, nil).Search(ctx, searchcap.Request{Query: query})
		return mustJSON(resp), nil
	}
	req := searchcap.Request{
		Query:       query,
		SourceScope: stringValue(args["source_scope"]),
		MaxResults:  intValue(args["max_results"]),
		TimeRange:   stringValue(args["time_range"]),
		TeamID:      stringValue(args["team_id"]),
		AgentID:     stringValue(args["agent_id"]),
		Visibility:  stringValue(args["visibility"]),
		Types:       stringSlice(args["types"]),
	}
	resp, err := r.search.Search(ctx, req)
	if err != nil {
		return fmt.Sprintf("Search failed: %v", err), nil
	}
	return mustJSON(resp), nil
}

func intValue(v any) int {
	switch raw := v.(type) {
	case int:
		return raw
	case float64:
		return int(raw)
	default:
		return 0
	}
}

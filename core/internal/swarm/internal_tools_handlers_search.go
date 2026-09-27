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
		return webSearchResult(resp)
	}
	req := searchcap.Request{
		Query:       query,
		SourceID:    stringValue(args["source_id"]),
		SourceScope: stringValue(args["source_scope"]),
		MaxResults:  intValue(args["max_results"]),
		TimeRange:   stringValue(args["time_range"]),
		TeamID:      stringValue(args["team_id"]),
		HostID:      stringValue(args["host_id"]),
		AgentID:     stringValue(args["agent_id"]),
		Visibility:  stringValue(args["visibility"]),
		Types:       stringSlice(args["types"]),
	}
	resp, err := r.search.Search(ctx, req)
	if err != nil {
		return "", fmt.Errorf("web_search failed: %w", err)
	}
	return webSearchResult(resp)
}

// WebSearchBlockedError is a search the provider could not run (disabled,
// not allowed, unreachable, unknown source). It is a tool failure, so the
// model hears it as one and tool.failed is emitted.
type WebSearchBlockedError struct {
	Code, Message, NextAction string
}

func (e *WebSearchBlockedError) Error() string {
	return fmt.Sprintf("web_search blocked (%s): %s Next action: %s", e.Code, e.Message, e.NextAction)
}

func webSearchResult(resp searchcap.Response) (string, error) {
	if resp.Status != "blocked" {
		return mustJSON(resp), nil
	}
	blocked := &WebSearchBlockedError{Code: "search_blocked", Message: "Search could not run.", NextAction: "Check the Mycelis Search configuration."}
	if resp.Blocker != nil {
		blocked.Code, blocked.Message, blocked.NextAction = resp.Blocker.Code, resp.Blocker.Message, resp.Blocker.NextAction
	}
	return "", blocked
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

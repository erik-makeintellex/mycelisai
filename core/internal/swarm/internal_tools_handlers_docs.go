package swarm

import (
	"context"
	"fmt"

	"github.com/mycelis/core/internal/helpdocs"
)

func (r *InternalToolRegistry) docsStore() (*helpdocs.Store, error) {
	return helpdocs.NewStore("")
}

func (r *InternalToolRegistry) handleListDocs(_ context.Context, _ map[string]any) (string, error) {
	store, err := r.docsStore()
	if err != nil {
		return "", err
	}
	return mustJSON(map[string]any{"sections": store.Sections()}), nil
}

func (r *InternalToolRegistry) handleReadDoc(_ context.Context, args map[string]any) (string, error) {
	slug := stringValue(args["slug"])
	if slug == "" {
		return "", fmt.Errorf("read_doc requires 'slug'")
	}
	store, err := r.docsStore()
	if err != nil {
		return "", err
	}
	doc, err := store.Read(slug)
	if err != nil {
		return "", err
	}
	return mustJSON(doc), nil
}

func (r *InternalToolRegistry) handleSearchDocs(_ context.Context, args map[string]any) (string, error) {
	query := stringValue(args["query"])
	if query == "" {
		return "", fmt.Errorf("search_docs requires 'query'")
	}
	limit := 8
	if l, ok := args["limit"].(float64); ok && l > 0 {
		limit = int(l)
	}
	store, err := r.docsStore()
	if err != nil {
		return "", err
	}
	results, err := store.Search(query, limit)
	if err != nil {
		return "", err
	}
	return mustJSON(map[string]any{"results": results}), nil
}

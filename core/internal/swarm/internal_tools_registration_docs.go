package swarm

func (r *InternalToolRegistry) registerDocsTools() {
	r.tools["list_docs"] = &InternalTool{Name: "list_docs", Description: "List curated Mycelis help and architecture docs that Soma can cite directly. This is read-only and separate from memory/context promotion.", InputSchema: map[string]any{"type": "object", "properties": map[string]any{}}, Handler: r.handleListDocs}
	r.tools["read_doc"] = &InternalTool{Name: "read_doc", Description: "Read one curated Mycelis documentation page by slug and return content with slug/path citation metadata.", InputSchema: map[string]any{"type": "object", "properties": map[string]any{"slug": map[string]any{"type": "string", "description": "Documentation slug from list_docs"}}, "required": []string{"slug"}}, Handler: r.handleReadDoc}
	r.tools["search_docs"] = &InternalTool{Name: "search_docs", Description: "Search curated Mycelis documentation by meaning/keywords and return citable excerpts. Does not mutate memory.", InputSchema: map[string]any{"type": "object", "properties": map[string]any{"query": map[string]any{"type": "string", "description": "Documentation search query"}, "limit": map[string]any{"type": "integer", "description": "Max results, default 8"}}, "required": []string{"query"}}, Handler: r.handleSearchDocs}
}

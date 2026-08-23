package swarm

func (r *InternalToolRegistry) registerCodeContextTools() {
	r.tools["code_context.query"] = &InternalTool{
		Name:        "code_context.query",
		Description: "Query registered native code context sources and return bounded source refs without exposing raw graph internals.",
		InputSchema: map[string]any{"type": "object", "properties": map[string]any{
			"query":     map[string]any{"type": "string", "description": "Symbol, file path fragment, package name, or source phrase to search."},
			"source_id": map[string]any{"type": "string", "description": "Optional registered code source id."},
			"path":      map[string]any{"type": "string", "description": "Optional source-relative folder or file boundary."},
			"limit":     map[string]any{"type": "integer", "description": "Maximum refs to return."},
		}, "required": []string{"query"}},
		Handler: r.handleCodeContextQuery,
	}
	r.tools["code_context.impact"] = &InternalTool{
		Name:        "code_context.impact",
		Description: "Review likely impact for a file, symbol, package, or phrase using extracted refs plus clearly labeled inferred relationships.",
		InputSchema: map[string]any{"type": "object", "properties": map[string]any{
			"target":    map[string]any{"type": "string", "description": "File, symbol, package, or phrase to review for impact."},
			"symbol":    map[string]any{"type": "string", "description": "Optional symbol target."},
			"path":      map[string]any{"type": "string", "description": "Optional source-relative file or folder target."},
			"query":     map[string]any{"type": "string", "description": "Optional fallback target query."},
			"source_id": map[string]any{"type": "string", "description": "Optional registered code source id."},
			"limit":     map[string]any{"type": "integer", "description": "Maximum refs to return."},
		}},
		Handler: r.handleCodeContextImpact,
	}
	r.tools["code_context.explain"] = &InternalTool{
		Name:        "code_context.explain",
		Description: "Explain a file or symbol from registered native code context using source-derived facts and citable refs.",
		InputSchema: map[string]any{"type": "object", "properties": map[string]any{
			"path":      map[string]any{"type": "string", "description": "Source-relative file to explain."},
			"symbol":    map[string]any{"type": "string", "description": "Optional symbol to explain."},
			"target":    map[string]any{"type": "string", "description": "Optional file, symbol, package, or phrase target."},
			"query":     map[string]any{"type": "string", "description": "Optional fallback target query."},
			"source_id": map[string]any{"type": "string", "description": "Optional registered code source id."},
			"limit":     map[string]any{"type": "integer", "description": "Maximum refs to return."},
		}},
		Handler: r.handleCodeContextExplain,
	}
}

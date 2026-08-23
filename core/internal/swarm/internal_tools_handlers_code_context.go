package swarm

import (
	"context"

	"github.com/mycelis/core/internal/codecontext"
)

func (r *InternalToolRegistry) handleCodeContextQuery(ctx context.Context, args map[string]any) (string, error) {
	resp, err := r.codeContextService().Query(ctx, codeContextRequest(args))
	if err != nil {
		return "", err
	}
	return mustJSON(resp), nil
}

func (r *InternalToolRegistry) handleCodeContextImpact(ctx context.Context, args map[string]any) (string, error) {
	resp, err := r.codeContextService().Impact(ctx, codeContextRequest(args))
	if err != nil {
		return "", err
	}
	return mustJSON(resp), nil
}

func (r *InternalToolRegistry) handleCodeContextExplain(ctx context.Context, args map[string]any) (string, error) {
	resp, err := r.codeContextService().Explain(ctx, codeContextRequest(args))
	if err != nil {
		return "", err
	}
	return mustJSON(resp), nil
}

func (r *InternalToolRegistry) codeContextService() *codecontext.Service {
	if r.codeContext != nil {
		return r.codeContext
	}
	return codecontext.NewService(codecontext.Config{})
}

func codeContextRequest(args map[string]any) codecontext.Request {
	return codecontext.Request{
		Query:    stringValue(args["query"]),
		SourceID: stringValue(args["source_id"]),
		Path:     stringValue(args["path"]),
		Target:   stringValue(args["target"]),
		Symbol:   stringValue(args["symbol"]),
		Limit:    intValue(args["limit"]),
	}
}

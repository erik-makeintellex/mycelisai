package swarm

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mycelis/core/internal/searchcap"
)

func TestInternalToolRegistryWebSearchRegistered(t *testing.T) {
	r := NewInternalToolRegistry(InternalToolDeps{})
	if !r.Has("web_search") {
		t.Fatal("expected web_search internal tool to be registered")
	}
}

func TestInternalToolRegistryWebSearchDisabledBlocker(t *testing.T) {
	r := NewInternalToolRegistry(InternalToolDeps{
		Search: searchcap.NewService(searchcap.Config{Provider: searchcap.ProviderDisabled}, nil, nil),
	})
	out, err := r.handleWebSearch(context.Background(), map[string]any{"query": "can you search the web?", "source_scope": "web"})
	if err != nil {
		t.Fatalf("handleWebSearch: %v", err)
	}
	var resp searchcap.Response
	if err := json.Unmarshal([]byte(out), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Blocker == nil || resp.Blocker.Code != "search_provider_disabled" {
		t.Fatalf("resp = %+v", resp)
	}
}

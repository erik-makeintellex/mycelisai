package swarm

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/mycelis/core/internal/searchcap"
)

func TestInternalToolRegistryWebSearchRegistered(t *testing.T) {
	r := NewInternalToolRegistry(InternalToolDeps{})
	if !r.Has("web_search") {
		t.Fatal("expected web_search internal tool to be registered")
	}
}

func assertWebSearchBlocked(t *testing.T, out string, err error, code string) {
	t.Helper()
	var blocked *WebSearchBlockedError
	if !errors.As(err, &blocked) {
		t.Fatalf("web_search = %q, %v; want a blocked failure", out, err)
	}
	if out != "" || blocked.Code != code || blocked.NextAction == "" || !strings.Contains(err.Error(), "Next action:") {
		t.Fatalf("blocked = %+v, out = %q; want code %s with a next action", blocked, out, code)
	}
}

func TestInternalToolRegistryWebSearchDisabledIsFailure(t *testing.T) {
	r := NewInternalToolRegistry(InternalToolDeps{
		Search: searchcap.NewService(searchcap.Config{Provider: searchcap.ProviderDisabled}, nil, nil),
	})
	out, err := r.handleWebSearch(context.Background(), map[string]any{"query": "can you search the web?", "source_scope": "web"})
	assertWebSearchBlocked(t, out, err, "search_provider_disabled")

	out, err = NewInternalToolRegistry(InternalToolDeps{}).handleWebSearch(context.Background(), map[string]any{"query": "x"})
	assertWebSearchBlocked(t, out, err, "search_provider_disabled")
}

func TestInternalToolRegistryWebSearchForwardsSourceID(t *testing.T) {
	r := NewInternalToolRegistry(InternalToolDeps{
		Search: searchcap.NewService(searchcap.Config{Provider: searchcap.ProviderDisabled}, nil, nil),
	})
	out, err := r.handleWebSearch(context.Background(), map[string]any{
		"query":     "release notes",
		"source_id": "missing-source",
	})
	assertWebSearchBlocked(t, out, err, "search_source_not_found")
}

func TestInternalToolRegistryWebSearchUnreachableProviderIsFailure(t *testing.T) {
	server := newSwarmLocalHTTPTestServer(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	endpoint := server.URL + "/search"
	server.Close()
	r := NewInternalToolRegistry(InternalToolDeps{
		Search: searchcap.NewService(searchcap.Config{Provider: searchcap.ProviderLocalAPI, LocalAPIEndpoint: endpoint}, nil, nil),
	})
	out, err := r.handleWebSearch(context.Background(), map[string]any{"query": "status"})
	assertWebSearchBlocked(t, out, err, "local_api_unreachable")
}

func TestInternalToolRegistryWebSearchSuccessUnchanged(t *testing.T) {
	server := newSwarmLocalHTTPTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":[{"title":"Mycelis","url":"https://example.test/m","snippet":"hit"}]}`))
	}))
	r := NewInternalToolRegistry(InternalToolDeps{
		Search: searchcap.NewService(searchcap.Config{Provider: searchcap.ProviderLocalAPI, LocalAPIEndpoint: server.URL + "/search"}, nil, nil),
	})
	out, err := r.handleWebSearch(context.Background(), map[string]any{"query": "mycelis"})
	if err != nil {
		t.Fatalf("handleWebSearch: %v", err)
	}
	var resp searchcap.Response
	if err := json.Unmarshal([]byte(out), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Status != "ok" || resp.Count != 1 || resp.Blocker != nil {
		t.Fatalf("resp = %+v", resp)
	}
}

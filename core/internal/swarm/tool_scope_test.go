package swarm

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"
)

var _ ToolSetResolver = (*fakeToolSetMCP)(nil)

// scopeTestRegistry registers counting internal tools with the given names.
func scopeTestRegistry(t *testing.T, names ...string) (*InternalToolRegistry, map[string]*atomic.Int32) {
	t.Helper()
	reg := &InternalToolRegistry{tools: map[string]*InternalTool{}}
	counts := map[string]*atomic.Int32{}
	for _, name := range names {
		count := &atomic.Int32{}
		counts[name] = count
		reg.tools[name] = &InternalTool{Name: name, Handler: func(context.Context, map[string]any) (string, error) {
			count.Add(1)
			return "ok:" + name, nil
		}}
	}
	return reg, counts
}

func TestScopedToolExecutor_DeclaredInternalToolRuns(t *testing.T) {
	reg, counts := scopeTestRegistry(t, "remember", "broadcast")
	scoped := NewScopedToolExecutor(NewCompositeToolExecutor(reg, nil), []string{" remember "}, nil)

	serverID, _, err := scoped.FindToolByName(context.Background(), "remember")
	if err != nil || serverID != InternalServerID {
		t.Fatalf("declared remember: (%v, %v)", serverID, err)
	}
	if out, err := scoped.CallTool(context.Background(), serverID, "remember", nil); err != nil || out != "ok:remember" {
		t.Fatalf("CallTool remember = (%q, %v)", out, err)
	}
	if counts["remember"].Load() != 1 {
		t.Fatalf("remember executed %d times, want 1", counts["remember"].Load())
	}
}

func TestScopedToolExecutor_UndeclaredInternalToolDeniedNotExecuted(t *testing.T) {
	reg, counts := scopeTestRegistry(t, "remember", "broadcast", "local_command")
	scoped := NewScopedToolExecutor(NewCompositeToolExecutor(reg, nil), []string{"remember"}, nil)

	for _, name := range []string{"broadcast", "local_command", " broadcast ", "Remember", "", "mcp:filesystem/*"} {
		if _, _, err := scoped.FindToolByName(context.Background(), name); !IsToolNotPermitted(err) {
			t.Fatalf("FindToolByName(%q) err = %v, want not permitted", name, err)
		}
	}
	// A caller that skips lookup and calls with the internal sentinel is still denied.
	_, err := scoped.CallTool(context.Background(), InternalServerID, "broadcast", nil)
	if !IsToolNotPermitted(err) {
		t.Fatalf("direct CallTool(broadcast) err = %v, want not permitted", err)
	}
	if got := err.Error(); got != `tool "broadcast" is not permitted for this agent` {
		t.Fatalf("denial message = %q", got)
	}
	if counts["broadcast"].Load() != 0 || counts["local_command"].Load() != 0 {
		t.Fatal("undeclared internal tool executed")
	}
}

func TestScopedToolExecutor_EmptyDeclaredListGrantsNothing(t *testing.T) {
	reg, _ := scopeTestRegistry(t, "remember", "recall")
	for _, declared := range [][]string{nil, {}, {"  "}, {"toolset:"}, {"mcp:"}} {
		scoped := NewScopedToolExecutor(NewCompositeToolExecutor(reg, &mockMCPExecutor{tools: map[string]uuid.UUID{"fetch": uuid.New()}}), declared, nil)
		for _, name := range []string{"remember", "recall", "fetch"} {
			if _, _, err := scoped.FindToolByName(context.Background(), name); !IsToolNotPermitted(err) {
				t.Fatalf("declared=%q %s: err = %v, want not permitted", declared, name, err)
			}
		}
	}
}

func TestScopedToolExecutor_RuntimeOwnedBaseToolset(t *testing.T) {
	reg, counts := scopeTestRegistry(t, "consult_council", "read_file", "write_file")
	scoped := NewScopedToolExecutor(NewCompositeToolExecutor(reg, nil), []string{"store_artifact"}, nil)
	runtimeCtx := WithToolInvocationContext(context.Background(), ToolInvocationContext{RuntimeOwned: true})

	for _, name := range []string{"consult_council", "read_file"} {
		serverID, _, err := scoped.FindToolByName(runtimeCtx, name)
		if err != nil {
			t.Fatalf("runtime-owned base %s: %v", name, err)
		}
		if _, err := scoped.CallTool(runtimeCtx, serverID, name, nil); err != nil {
			t.Fatalf("runtime-owned base call %s: %v", name, err)
		}
		// The same tool chosen by the model is not implicitly granted.
		if _, _, err := scoped.FindToolByName(context.Background(), name); !IsToolNotPermitted(err) {
			t.Fatalf("model-selected %s: err = %v, want not permitted", name, err)
		}
	}
	if _, _, err := scoped.FindToolByName(runtimeCtx, "write_file"); !IsToolNotPermitted(err) {
		t.Fatalf("runtime-owned write_file without declaration: err = %v, want not permitted", err)
	}
	if counts["consult_council"].Load() != 1 || counts["read_file"].Load() != 1 || counts["write_file"].Load() != 0 {
		t.Fatal("unexpected base toolset execution counts")
	}
}

// fakeToolSetMCP is an MCP executor that also resolves toolsets.
type fakeToolSetMCP struct {
	mockMCPExecutor
	sets     map[string][]string
	failures atomic.Int32
	resolves atomic.Int32
}

func (f *fakeToolSetMCP) ResolveRefs(_ context.Context, tools []string) ([]string, error) {
	f.resolves.Add(1)
	if f.failures.Load() > 0 {
		f.failures.Add(-1)
		return nil, errors.New("toolset store unavailable")
	}
	var out []string
	for _, entry := range tools {
		out = append(out, f.sets[strings.TrimPrefix(entry, "toolset:")]...)
	}
	return out, nil
}

func TestScopedToolExecutor_ToolsetExpansion(t *testing.T) {
	fsID, ghID := uuid.New(), uuid.New()
	mcpExec := &fakeToolSetMCP{
		mockMCPExecutor: mockMCPExecutor{tools: map[string]uuid.UUID{"list_dir": fsID, "create_issue": ghID}},
		sets:            map[string][]string{"workspace": {"mcp:filesystem/*", "recall"}},
	}
	reg, _ := scopeTestRegistry(t, "recall", "broadcast")
	names := map[uuid.UUID]string{fsID: "filesystem", ghID: "github"}
	mcpExec.failures.Store(1)
	scoped := NewScopedToolExecutor(NewCompositeToolExecutor(reg, mcpExec), []string{"toolset:workspace", "toolset:missing"}, names)

	// Resolution failure grants nothing, then retries.
	if _, _, err := scoped.FindToolByName(context.Background(), "list_dir"); !IsToolNotPermitted(err) {
		t.Fatalf("list_dir during resolver outage: err = %v, want not permitted", err)
	}
	for _, name := range []string{"list_dir", "recall"} {
		if _, _, err := scoped.FindToolByName(context.Background(), name); err != nil {
			t.Fatalf("toolset-granted %s: %v", name, err)
		}
	}
	for _, name := range []string{"create_issue", "broadcast"} {
		if _, _, err := scoped.FindToolByName(context.Background(), name); !IsToolNotPermitted(err) {
			t.Fatalf("%s outside toolset: err = %v, want not permitted", name, err)
		}
	}
	if got := mcpExec.resolves.Load(); got != 2 {
		t.Fatalf("toolset resolved %d times, want 2 (one failure, one cached success)", got)
	}
}

func TestScopedToolExecutor_ConcurrentAgentsKeepSeparateScopes(t *testing.T) {
	reg, counts := scopeTestRegistry(t, "remember", "broadcast")
	composite := NewCompositeToolExecutor(reg, &fakeToolSetMCP{sets: map[string][]string{"ops": {"broadcast"}}})
	var wg sync.WaitGroup
	var leaks atomic.Int32
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			// Each goroutine builds a fresh agent scope, as a team restart would.
			declared := []string{"remember"}
			if i%2 == 0 {
				declared = []string{"toolset:ops"}
			}
			agent := NewAgent(context.Background(), agentManifestWithTools("a", declared), "team", nil, nil, composite)
			for j := 0; j < 50; j++ {
				_, _, errB := agent.toolExecutor.FindToolByName(context.Background(), "broadcast")
				_, errC := agent.toolExecutor.CallTool(context.Background(), InternalServerID, "remember", nil)
				if (i%2 == 0) != (errB == nil) || (i%2 == 1) != (errC == nil) {
					leaks.Add(1)
				}
			}
		}(i)
	}
	wg.Wait()
	if leaks.Load() != 0 {
		t.Fatalf("%d cross-agent scope leaks", leaks.Load())
	}
	if counts["remember"].Load() != 8*50 || counts["broadcast"].Load() != 0 {
		t.Fatalf("executions remember=%d broadcast=%d", counts["remember"].Load(), counts["broadcast"].Load())
	}
}

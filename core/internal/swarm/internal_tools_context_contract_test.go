package swarm

import (
	"strings"
	"testing"
)

func TestBuildContext_IncludesMCPTranslationProcedure(t *testing.T) {
	r := NewInternalToolRegistry(InternalToolDeps{})
	ctx := r.BuildContext("admin", "admin-core", "admin", []string{"swarm.global.input.user"}, []string{"swarm.team.admin-core.signal.status"}, "find latest docs")
	if !strings.Contains(ctx, "MCP Translation Procedure") {
		t.Fatalf("expected MCP translation procedure in runtime context, got:\n%s", ctx)
	}
	if !strings.Contains(ctx, "web access tasks") {
		t.Fatalf("expected web-access execution rule in runtime context, got:\n%s", ctx)
	}
	if !strings.Contains(ctx, "Team command bus") {
		t.Fatalf("expected standardized team command bus in runtime context, got:\n%s", ctx)
	}
	if !strings.Contains(ctx, "Team result bus") {
		t.Fatalf("expected standardized team result bus in runtime context, got:\n%s", ctx)
	}
}

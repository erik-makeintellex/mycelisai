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

func TestBuildContext_IncludesGovernedMemoryBoundaries(t *testing.T) {
	r := NewInternalToolRegistry(InternalToolDeps{})
	ctx := r.BuildContext("council-architect", "design-core", "architect", nil, nil, "review the customer deployment brief")

	for _, expected := range []string{
		"### Memory Boundaries",
		"SOMA_MEMORY",
		"AGENT_MEMORY",
		"PROJECT_MEMORY",
		"REFLECTION_MEMORY",
		"LearningCandidate",
		"knowledge_class=customer_context",
		"knowledge_class=company_knowledge",
		"knowledge_class=soma_operating_context",
		"knowledge_class=user_private_context",
		"reflection_synthesis",
	} {
		if !strings.Contains(ctx, expected) {
			t.Fatalf("expected %q in runtime context, got:\n%s", expected, ctx)
		}
	}
}

func TestBuildContext_ScopesWorkerAwayFromOrganizationInventory(t *testing.T) {
	r := NewInternalToolRegistry(InternalToolDeps{})
	ctx := r.BuildContext(
		"delivery-worker",
		"customer-delivery",
		"worker",
		[]string{"swarm.team.customer-delivery.internal.command"},
		[]string{"swarm.team.customer-delivery.signal.result"},
		"Create and validate the retained package.",
	)

	for _, expected := range []string{
		"### Your Identity & NATS Topology",
		"### Scoped Execution Protocol",
		"create, read back, and validate",
		"retained artifact references",
	} {
		if !strings.Contains(ctx, expected) {
			t.Fatalf("expected %q in worker runtime context, got:\n%s", expected, ctx)
		}
	}
	for _, excluded := range []string{
		"### Active Teams",
		"### Installed MCP Servers & Tools",
		"### Memory Boundaries",
		"### Interaction Protocol",
	} {
		if strings.Contains(ctx, excluded) {
			t.Fatalf("did not expect %q in scoped worker runtime context, got:\n%s", excluded, ctx)
		}
	}
	if len(ctx) > 4_000 {
		t.Fatalf("worker runtime context length = %d, want <= 4000", len(ctx))
	}
}

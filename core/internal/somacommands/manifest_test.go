package somacommands_test

import (
	"testing"

	"github.com/mycelis/core/internal/somacommands"
	"github.com/mycelis/core/internal/swarm"
)

func TestDefaultCommandManifestMatchesInternalToolHandlers(t *testing.T) {
	registry, err := somacommands.LoadDefault()
	if err != nil {
		t.Fatalf("LoadDefault: %v", err)
	}
	tools := swarm.NewInternalToolRegistry(swarm.InternalToolDeps{})
	if err := registry.ValidateHandlers(tools.ListNames()); err != nil {
		t.Fatalf("ValidateHandlers: %v", err)
	}
	if got, want := len(registry.Commands), len(tools.ListNames()); got != want {
		t.Fatalf("commands = %d, want %d", got, want)
	}
}

func TestDefaultCommandManifestExposesSearchPosture(t *testing.T) {
	registry, err := somacommands.LoadDefault()
	if err != nil {
		t.Fatalf("LoadDefault: %v", err)
	}
	command, ok := registry.ByHandler()["web_search"]
	if !ok {
		t.Fatal("web_search manifest missing")
	}
	if command.UserQuote == "" || command.CapabilityID != "web_search" {
		t.Fatalf("unexpected web_search manifest: %+v", command)
	}
	if command.Scope.Default != "workspace" || len(command.Scope.Groups) == 0 {
		t.Fatalf("web_search scope should expose configured sources: %+v", command.Scope)
	}
	if !command.Delivery.ProofRequired {
		t.Fatalf("web_search should require proof: %+v", command.Delivery)
	}
}

func TestDefaultCommandManifestExposesCodeContextPosture(t *testing.T) {
	registry, err := somacommands.LoadDefault()
	if err != nil {
		t.Fatalf("LoadDefault: %v", err)
	}
	command, ok := registry.ByHandler()["code_context.impact"]
	if !ok {
		t.Fatal("code_context.impact manifest missing")
	}
	if command.CapabilityID != "code_context" {
		t.Fatalf("capability_id = %q", command.CapabilityID)
	}
	if !command.Delivery.ProofRequired || command.Metadata["graph_internals"] != "hidden" {
		t.Fatalf("unexpected code context manifest: %+v", command)
	}
}

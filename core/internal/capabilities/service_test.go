package capabilities

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/exchange"
	"github.com/mycelis/core/internal/mcp"
	"github.com/mycelis/core/internal/searchcap"
)

type fakeMCPRegistry struct {
	servers []mcp.ServerConfig
	tools   []mcp.ToolDef
}

func (f fakeMCPRegistry) List(context.Context) ([]mcp.ServerConfig, error) {
	return f.servers, nil
}

func (f fakeMCPRegistry) ListAllTools(context.Context) ([]mcp.ToolDef, error) {
	return f.tools, nil
}

type fakeToolLister map[string]string

func (f fakeToolLister) ListDescriptions() map[string]string {
	return map[string]string(f)
}

type fakeSearchStatusProvider struct {
	status searchcap.Status
}

func (f fakeSearchStatusProvider) Status() searchcap.Status {
	return f.status
}

func TestServiceDerivesCapabilityManifestSnapshot(t *testing.T) {
	now := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)
	serverID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	toolID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	inputSchema := json.RawMessage(`{"type":"object","properties":{"path":{"type":"string"}}}`)

	svc := NewService(Dependencies{
		ExchangeCapabilities: []exchange.CapabilityDefinition{{
			ID:                  "planning",
			Label:               "Planning",
			Source:              "internal_tool",
			RiskClass:           "low-risk",
			DefaultAllowedRoles: []string{"soma"},
			Description:         "Planning outputs.",
		}},
		MCP: fakeMCPRegistry{
			servers: []mcp.ServerConfig{{
				ID:        serverID,
				Name:      "filesystem",
				Transport: "stdio",
				Command:   "npx",
				Status:    "installed",
			}},
			tools: []mcp.ToolDef{{
				ID:          toolID,
				ServerID:    serverID,
				ServerName:  "filesystem",
				Name:        "read_file",
				Description: "Read a workspace file.",
				InputSchema: inputSchema,
			}},
		},
		InternalTools: fakeToolLister{"delegate_task": "Delegate work."},
		Search: fakeSearchStatusProvider{status: searchcap.Status{
			Provider:              searchcap.ProviderSearXNG,
			Enabled:               true,
			Configured:            true,
			SupportsPublicWeb:     true,
			SomaToolName:          "web_search",
			DirectSomaInteraction: true,
			MaxResults:            6,
		}},
		HostCommands: func() []string { return []string{"hostname"} },
		Now:          func() time.Time { return now },
	})

	snap, err := svc.Refresh(context.Background())
	if err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if snap.Count != 6 {
		t.Fatalf("Count = %d, want 6", snap.Count)
	}

	assertManifest(t, snap, "planning", "exchange_capability", "available")
	assertManifest(t, snap, "search:web_search", "search_capability", "enabled")
	assertManifest(t, snap, "internal_tool:delegate_task", "internal_tool", "available")
	assertManifest(t, snap, "hostcmd:hostname", "host_command", "allowlisted")
	assertManifest(t, snap, "mcp_server:"+serverID.String(), "mcp_server", "installed")
	mcpTool := assertManifest(t, snap, "mcp_tool:"+toolID.String(), "mcp_tool", "installed")
	if got := mcpTool.ToolRefs[0]; got != "mcp:filesystem/read_file" {
		t.Fatalf("MCP tool ref = %q", got)
	}
	if mcpTool.Metadata["input_schema"] == nil {
		t.Fatal("MCP tool input_schema metadata missing")
	}
}

func TestServiceGetUsesCachedSnapshot(t *testing.T) {
	now := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)
	svc := NewService(Dependencies{
		ExchangeCapabilities: []exchange.CapabilityDefinition{{
			ID:        "review",
			Label:     "Review",
			Source:    "internal_tool",
			RiskClass: "medium-risk",
		}},
		HostCommands: func() []string { return nil },
		Now:          func() time.Time { return now },
	})

	manifest, err := svc.Get(context.Background(), "review")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if manifest == nil {
		t.Fatal("Get(review) returned nil")
	}
	if manifest.DerivedAt != now {
		t.Fatalf("DerivedAt = %s, want %s", manifest.DerivedAt, now)
	}
}

func assertManifest(t *testing.T, snap Snapshot, id, kind, status string) Manifest {
	t.Helper()
	for _, manifest := range snap.Manifests {
		if manifest.ID == id {
			if manifest.Kind != kind {
				t.Fatalf("%s kind = %q, want %q", id, manifest.Kind, kind)
			}
			if manifest.Status != status {
				t.Fatalf("%s status = %q, want %q", id, manifest.Status, status)
			}
			return manifest
		}
	}
	t.Fatalf("manifest %q not found in %#v", id, snap.Manifests)
	return Manifest{}
}

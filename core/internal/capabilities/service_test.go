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
	"github.com/mycelis/core/internal/somacommands"
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

type fakeManifestToolLister struct {
	descriptions map[string]string
	commands     []somacommands.Command
}

func (f fakeManifestToolLister) ListDescriptions() map[string]string {
	return f.descriptions
}

func (f fakeManifestToolLister) ListCommandManifests() []somacommands.Command {
	return f.commands
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
	if mcpTool.CapabilityID != mcpTool.ID {
		t.Fatalf("capability_id = %q, want %q", mcpTool.CapabilityID, mcpTool.ID)
	}
	if mcpTool.ManifestVersion != ManifestVersion {
		t.Fatalf("manifest_version = %q, want %q", mcpTool.ManifestVersion, ManifestVersion)
	}
	if mcpTool.Health != "healthy" {
		t.Fatalf("health = %q, want healthy", mcpTool.Health)
	}
	if mcpTool.ApprovalPosture != "not_required" {
		t.Fatalf("approval_posture = %q, want not_required", mcpTool.ApprovalPosture)
	}
	if mcpTool.OutputSchemaRef == "" {
		t.Fatal("output_schema_ref missing")
	}
	if mcpTool.FailurePosture == "" || mcpTool.RecoveryPosture == "" {
		t.Fatalf("failure/recovery posture missing: %+v", mcpTool)
	}
}

func TestServiceProjectsInternalToolCommandManifest(t *testing.T) {
	now := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)
	svc := NewService(Dependencies{
		ExchangeCapabilities: []exchange.CapabilityDefinition{},
		InternalTools: fakeManifestToolLister{
			descriptions: map[string]string{"web_search": "Fallback description."},
			commands: []somacommands.Command{{
				ID:              "search.web",
				Handler:         "web_search",
				Title:           "Web and source search",
				Summary:         "Search configured sources.",
				UserQuote:       "Research this for me.",
				Category:        "Search",
				CapabilityID:    "web_search",
				InputSchemaRef:  "internal_tool.web_search.input",
				OutputSchemaRef: "search_results.v1",
				Scope: somacommands.Scope{
					Default: "workspace",
					Roles:   []string{"soma", "team_lead"},
					Groups:  []string{"configured_sources"},
				},
				Governance: somacommands.Governance{
					RiskClass:       "medium-risk",
					ApprovalPosture: "not_required",
					AuditRequired:   true,
				},
				Delivery: somacommands.Delivery{
					OutputKinds:     []string{"search_results"},
					ProofRequired:   true,
					RecoveryPosture: "configure_source_or_retry_local",
				},
			}},
		},
		HostCommands: func() []string { return nil },
		Now:          func() time.Time { return now },
	})

	snap, err := svc.Refresh(context.Background())
	if err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	manifest := assertManifest(t, snap, "internal_tool:web_search", "internal_tool", "available")
	if manifest.DisplayName != "Web and source search" {
		t.Fatalf("display_name = %q", manifest.DisplayName)
	}
	if manifest.CapabilityID != "web_search" {
		t.Fatalf("capability_id = %q", manifest.CapabilityID)
	}
	if manifest.InputSchemaRef != "internal_tool.web_search.input" {
		t.Fatalf("input_schema_ref = %q", manifest.InputSchemaRef)
	}
	if manifest.RecoveryPosture != "configure_source_or_retry_local" {
		t.Fatalf("recovery_posture = %q", manifest.RecoveryPosture)
	}
	if manifest.Metadata["user_quote"] != "Research this for me." {
		t.Fatalf("metadata = %#v", manifest.Metadata)
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

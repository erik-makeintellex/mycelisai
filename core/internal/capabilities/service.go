package capabilities

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/mycelis/core/internal/exchange"
	"github.com/mycelis/core/internal/hostcmd"
	"github.com/mycelis/core/internal/mcp"
	"github.com/mycelis/core/internal/searchcap"
)

type MCPRegistry interface {
	List(context.Context) ([]mcp.ServerConfig, error)
	ListAllTools(context.Context) ([]mcp.ToolDef, error)
}

type InternalToolLister interface {
	ListDescriptions() map[string]string
}

type SearchStatusProvider interface {
	Status() searchcap.Status
}

type Dependencies struct {
	ExchangeCapabilities []exchange.CapabilityDefinition
	MCP                  MCPRegistry
	MCPLibrary           *mcp.Library
	InternalTools        InternalToolLister
	Search               SearchStatusProvider
	HostCommands         func() []string
	Now                  func() time.Time
}

type Service struct {
	mu   sync.RWMutex
	deps Dependencies
	snap Snapshot
}

func NewService(deps Dependencies) *Service {
	if deps.ExchangeCapabilities == nil {
		deps.ExchangeCapabilities = exchange.SeedCapabilities
	}
	if deps.HostCommands == nil {
		deps.HostCommands = hostcmd.AllowedCommands
	}
	if deps.Now == nil {
		deps.Now = func() time.Time { return time.Now().UTC() }
	}
	return &Service{deps: deps}
}

func (s *Service) List(ctx context.Context) (Snapshot, error) {
	if s == nil {
		return Snapshot{}, fmt.Errorf("capability manifest service not initialized")
	}
	s.mu.RLock()
	if !s.snap.GeneratedAt.IsZero() {
		snap := cloneSnapshot(s.snap)
		s.mu.RUnlock()
		return snap, nil
	}
	s.mu.RUnlock()
	return s.Refresh(ctx)
}

func (s *Service) Get(ctx context.Context, id string) (*Manifest, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, nil
	}
	snap, err := s.List(ctx)
	if err != nil {
		return nil, err
	}
	for i := range snap.Manifests {
		if snap.Manifests[i].ID == id {
			manifest := snap.Manifests[i]
			return &manifest, nil
		}
	}
	return nil, nil
}

func (s *Service) Refresh(ctx context.Context) (Snapshot, error) {
	if s == nil {
		return Snapshot{}, fmt.Errorf("capability manifest service not initialized")
	}
	generatedAt := s.deps.Now().UTC()
	manifests := s.derive(ctx, generatedAt)
	sort.SliceStable(manifests, func(i, j int) bool {
		if manifests[i].Kind == manifests[j].Kind {
			return manifests[i].ID < manifests[j].ID
		}
		return manifests[i].Kind < manifests[j].Kind
	})
	snap := Snapshot{
		GeneratedAt: generatedAt,
		Count:       len(manifests),
		Manifests:   manifests,
	}
	s.mu.Lock()
	s.snap = cloneSnapshot(snap)
	s.mu.Unlock()
	return snap, nil
}

func (s *Service) derive(ctx context.Context, derivedAt time.Time) []Manifest {
	out := make([]Manifest, 0)
	seen := map[string]struct{}{}
	add := func(m Manifest) {
		m.ID = strings.TrimSpace(m.ID)
		if m.ID == "" {
			return
		}
		if _, exists := seen[m.ID]; exists {
			return
		}
		seen[m.ID] = struct{}{}
		if m.Version == "" {
			m.Version = ManifestVersion
		}
		if m.Status == "" {
			m.Status = "available"
		}
		if m.Source == "" {
			m.Source = "system"
		}
		if m.RiskClass == "" {
			m.RiskClass = "low-risk"
		}
		if m.Metadata == nil {
			m.Metadata = map[string]any{}
		}
		m.DerivedAt = derivedAt
		out = append(out, m)
	}

	for _, cap := range s.deps.ExchangeCapabilities {
		add(manifestFromExchangeCapability(cap))
	}
	if s.deps.Search != nil {
		add(manifestFromSearchStatus(s.deps.Search.Status()))
	} else {
		add(manifestFromSearchStatus(searchcap.Status{
			Provider:              searchcap.ProviderDisabled,
			SomaToolName:          "web_search",
			DirectSomaInteraction: true,
			MaxResults:            8,
		}))
	}
	if s.deps.InternalTools != nil {
		for name, desc := range s.deps.InternalTools.ListDescriptions() {
			add(Manifest{
				ID:          "internal_tool:" + name,
				DisplayName: name,
				Kind:        "internal_tool",
				Source:      "internal_tool",
				Status:      "available",
				RiskClass:   riskForInternalTool(name),
				Description: desc,
				ToolRefs:    []string{name},
				DefaultAllowedRoles: []string{
					"soma",
					"team_lead",
					"specialist",
					"automation",
				},
				AuditRequired: strings.Contains(name, "local_command") || strings.Contains(name, "exchange"),
				Metadata:      map[string]any{"registry": "swarm_internal_tools"},
			})
		}
	}
	for _, cmd := range s.deps.HostCommands() {
		add(Manifest{
			ID:                  "hostcmd:" + cmd,
			DisplayName:         "Host command: " + cmd,
			Kind:                "host_command",
			Source:              "hostcmd_allowlist",
			Status:              "allowlisted",
			RiskClass:           "medium-risk",
			Description:         "Allowlisted local host command exposed through the governed host action API.",
			ToolRefs:            []string{"host:local-command:" + cmd},
			DefaultAllowedRoles: []string{"admin"},
			AuditRequired:       true,
			ApprovalRequired:    false,
			Metadata:            map[string]any{"allowlist_env_field": "MYCELIS_LOCAL_COMMAND_ALLOWLIST"},
		})
	}
	if s.deps.MCP != nil {
		if servers, err := s.deps.MCP.List(ctx); err == nil {
			for _, srv := range servers {
				add(manifestFromMCPServer(srv))
			}
		} else {
			add(Manifest{
				ID:                  "mcp:registry",
				DisplayName:         "MCP Registry",
				Kind:                "mcp_registry",
				Source:              "mcp",
				Status:              "degraded",
				RiskClass:           "medium-risk",
				Description:         "MCP registry could not be queried while deriving the capability manifest.",
				DefaultAllowedRoles: []string{"admin", "soma"},
				AuditRequired:       true,
				Metadata:            map[string]any{"error": err.Error()},
			})
		}
		if tools, err := s.deps.MCP.ListAllTools(ctx); err == nil {
			for _, tool := range tools {
				add(manifestFromMCPTool(tool))
			}
		}
	}
	if s.deps.MCPLibrary != nil {
		for _, cat := range s.deps.MCPLibrary.Categories {
			for _, entry := range cat.Servers {
				add(manifestFromMCPLibraryEntry(cat.Name, entry))
			}
		}
	}
	return out
}

func cloneSnapshot(snap Snapshot) Snapshot {
	out := Snapshot{
		GeneratedAt: snap.GeneratedAt,
		Count:       snap.Count,
		Manifests:   make([]Manifest, len(snap.Manifests)),
	}
	copy(out.Manifests, snap.Manifests)
	return out
}

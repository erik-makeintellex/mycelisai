package main

import (
	"context"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/internal/mcp"
	"github.com/mycelis/core/internal/swarm"
)

// buildSystemCapabilities assembles the live system context that the
// Meta-Architect uses to generate resource-aware mission blueprints.
func buildSystemCapabilities(ctx context.Context, tools *swarm.InternalToolRegistry, mcpSvc *mcp.Service, lib *mcp.Library) *cognitive.SystemCapabilities {
	caps := &cognitive.SystemCapabilities{}

	if tools != nil {
		caps.InternalTools = tools.ListDescriptions()
	}

	if mcpSvc != nil {
		servers, err := mcpSvc.List(ctx)
		if err == nil {
			for _, srv := range servers {
				toolNames := []string{}
				if defs, err := mcpSvc.ListTools(ctx, srv.ID); err == nil {
					for _, t := range defs {
						toolNames = append(toolNames, t.Name)
					}
				}
				caps.MCPServers = append(caps.MCPServers, cognitive.MCPServerCapability{
					Name:   srv.Name,
					Status: "installed",
					Tools:  toolNames,
				})
			}
		}
	}

	if lib != nil {
		for _, cat := range lib.Categories {
			for _, entry := range cat.Servers {
				reqEnv := make([]string, 0, len(entry.Env))
				for k := range entry.Env {
					reqEnv = append(reqEnv, k)
				}
				caps.MCPServers = append(caps.MCPServers, cognitive.MCPServerCapability{
					Name:        entry.Name,
					Description: entry.Description,
					Status:      "available",
					RequiredEnv: reqEnv,
				})
			}
		}
	}

	return caps
}

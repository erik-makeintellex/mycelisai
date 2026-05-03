package mcp

import (
	"context"
	"log"
)

func (s *Service) BootstrapDefaults(ctx context.Context, library *Library, pool *ClientPool) {
	log.Println("[mcp] bootstrap: starting default MCP server installation...")

	for _, name := range []string{"filesystem", "fetch"} {
		existing, err := s.FindServerByName(ctx, name)
		if err != nil {
			log.Printf("[mcp] bootstrap: DB check for %s failed (is the database reachable?): %v", name, err)
			continue
		}
		if existing != nil {
			s.reconnectDefaultIfNeeded(ctx, pool, name, existing)
			continue
		}

		entry := library.FindByName(name)
		if entry == nil {
			log.Printf("[mcp] bootstrap: %s not found in library, skipping", name)
			continue
		}
		s.installDefault(ctx, pool, name, entry)
	}
	log.Println("[mcp] bootstrap: default server installation complete.")
	s.seedDefaultToolSets(ctx)
}

func (s *Service) reconnectDefaultIfNeeded(ctx context.Context, pool *ClientPool, name string, existing *ServerConfig) {
	log.Printf("[mcp] bootstrap: %s already installed (id=%s, status=%s)", name, existing.ID, existing.Status)
	if existing.Status != "error" && existing.Status != "pending" {
		return
	}
	log.Printf("[mcp] bootstrap: retrying connect for %s (was %s)", name, existing.Status)
	if err := withMCPConnectTimeout(ctx, func(connectCtx context.Context) error {
		return pool.Connect(connectCtx, *existing)
	}); err != nil {
		log.Printf("[mcp] bootstrap: reconnect %s failed: %v", name, err)
		return
	}
	log.Printf("[mcp] bootstrap: %s reconnected successfully", name)
}

func (s *Service) installDefault(ctx context.Context, pool *ClientPool, name string, entry *LibraryEntry) {
	cfg := entry.ToServerConfig(nil)
	if name == "filesystem" {
		var err error
		cfg, err = ApplyRuntimeDefaults(cfg)
		if err != nil {
			log.Printf("[mcp] bootstrap: failed to prepare filesystem workspace root %s: %v", ResolveFilesystemWorkspaceRoot(), err)
		}
		log.Printf("[mcp] bootstrap: filesystem mount path: %s", ResolveFilesystemWorkspaceRoot())
	}
	log.Printf("[mcp] bootstrap: installing %s (command: %s %v)", name, cfg.Command, cfg.Args)
	installed, err := s.Install(ctx, cfg)
	if err != nil {
		log.Printf("[mcp] bootstrap: failed to install %s: %v", name, err)
		return
	}
	log.Printf("[mcp] bootstrap: installed %s (id=%s)", name, installed.ID)
	if err := withMCPConnectTimeout(ctx, func(connectCtx context.Context) error {
		return pool.Connect(connectCtx, *installed)
	}); err != nil {
		log.Printf("[mcp] bootstrap: %s installed but connect failed: %v", name, err)
		return
	}
	log.Printf("[mcp] bootstrap: %s connected and tools discovered", name)
}

func (s *Service) seedDefaultToolSets(ctx context.Context) {
	if s.ToolSets == nil {
		return
	}
	defaultSets := []ToolSet{
		{Name: "workspace", Description: "File I/O and workspace management", ToolRefs: []string{"mcp:filesystem/*"}},
		{Name: "research", Description: "Web fetch and search tools", ToolRefs: []string{"mcp:fetch/*", "mcp:brave-search/*"}},
	}
	for _, ts := range defaultSets {
		existing, _ := s.ToolSets.FindByName(ctx, ts.Name)
		if existing != nil {
			continue
		}
		if created, err := s.ToolSets.Create(ctx, ts); err != nil {
			log.Printf("[mcp] bootstrap: failed to create tool set %q: %v", ts.Name, err)
		} else {
			log.Printf("[mcp] bootstrap: seeded tool set %q (id=%s)", ts.Name, created.ID)
		}
	}
}

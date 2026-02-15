package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
)

// ManagedClient wraps an active MCP client connection with its metadata.
type ManagedClient struct {
	ServerID  uuid.UUID
	Config    ServerConfig
	Client    *client.Client
	Tools     []mcp.Tool
	Connected bool
}

// ClientPool manages live MCP client connections with thread-safe access.
type ClientPool struct {
	mu      sync.RWMutex
	clients map[uuid.UUID]*ManagedClient
	service *Service // reference to the DB service for status updates + tool caching
}

// NewClientPool creates a new pool that uses the given service for persistence.
func NewClientPool(svc *Service) *ClientPool {
	return &ClientPool{
		clients: make(map[uuid.UUID]*ManagedClient),
		service: svc,
	}
}

// Connect establishes a live MCP connection for the given server config.
// It creates the transport, starts the client, initializes the MCP session,
// discovers tools, caches them in the database, and stores the managed client.
func (p *ClientPool) Connect(ctx context.Context, cfg ServerConfig) error {
	var t transport.Interface
	var err error

	switch cfg.Transport {
	case "stdio":
		// Convert env map to []string{"KEY=VALUE", ...} for stdio transport.
		envSlice := make([]string, 0, len(cfg.Env))
		for k, v := range cfg.Env {
			envSlice = append(envSlice, k+"="+v)
		}
		t = transport.NewStdio(cfg.Command, envSlice, cfg.Args...)

	case "sse":
		t, err = transport.NewStreamableHTTP(cfg.URL)
		if err != nil {
			statusErr := p.service.UpdateStatus(ctx, cfg.ID, "error", fmt.Sprintf("transport init: %v", err))
			if statusErr != nil {
				log.Printf("mcp pool: failed to update status for %s: %v", cfg.ID, statusErr)
			}
			return fmt.Errorf("create streamable HTTP transport for %s: %w", cfg.Name, err)
		}

	default:
		errMsg := fmt.Sprintf("unsupported transport type: %s", cfg.Transport)
		statusErr := p.service.UpdateStatus(ctx, cfg.ID, "error", errMsg)
		if statusErr != nil {
			log.Printf("mcp pool: failed to update status for %s: %v", cfg.ID, statusErr)
		}
		return fmt.Errorf("%s", errMsg)
	}

	// Create the MCP client with the transport.
	c := client.NewClient(t)

	// Start the transport connection.
	if err := c.Start(ctx); err != nil {
		statusErr := p.service.UpdateStatus(ctx, cfg.ID, "error", fmt.Sprintf("start: %v", err))
		if statusErr != nil {
			log.Printf("mcp pool: failed to update status for %s: %v", cfg.ID, statusErr)
		}
		return fmt.Errorf("start mcp client for %s: %w", cfg.Name, err)
	}

	// Initialize the MCP session.
	initReq := mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ClientInfo: mcp.Implementation{
				Name:    "mycelis-core",
				Version: "0.7.0",
			},
		},
	}
	_, err = c.Initialize(ctx, initReq)
	if err != nil {
		_ = c.Close()
		statusErr := p.service.UpdateStatus(ctx, cfg.ID, "error", fmt.Sprintf("initialize: %v", err))
		if statusErr != nil {
			log.Printf("mcp pool: failed to update status for %s: %v", cfg.ID, statusErr)
		}
		return fmt.Errorf("initialize mcp session for %s: %w", cfg.Name, err)
	}

	// Discover tools from the server.
	toolsResult, err := c.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		_ = c.Close()
		statusErr := p.service.UpdateStatus(ctx, cfg.ID, "error", fmt.Sprintf("list tools: %v", err))
		if statusErr != nil {
			log.Printf("mcp pool: failed to update status for %s: %v", cfg.ID, statusErr)
		}
		return fmt.Errorf("discover tools for %s: %w", cfg.Name, err)
	}

	tools := toolsResult.Tools

	// Convert []mcp.Tool to []ToolDef and cache in the database.
	toolDefs, err := convertTools(cfg.ID, tools)
	if err != nil {
		_ = c.Close()
		statusErr := p.service.UpdateStatus(ctx, cfg.ID, "error", fmt.Sprintf("convert tools: %v", err))
		if statusErr != nil {
			log.Printf("mcp pool: failed to update status for %s: %v", cfg.ID, statusErr)
		}
		return fmt.Errorf("convert tools for %s: %w", cfg.Name, err)
	}

	if err := p.service.CacheTools(ctx, cfg.ID, toolDefs); err != nil {
		_ = c.Close()
		statusErr := p.service.UpdateStatus(ctx, cfg.ID, "error", fmt.Sprintf("cache tools: %v", err))
		if statusErr != nil {
			log.Printf("mcp pool: failed to update status for %s: %v", cfg.ID, statusErr)
		}
		return fmt.Errorf("cache tools for %s: %w", cfg.Name, err)
	}

	// Update the server status to connected.
	if err := p.service.UpdateStatus(ctx, cfg.ID, "connected", ""); err != nil {
		_ = c.Close()
		return fmt.Errorf("update status for %s: %w", cfg.Name, err)
	}

	// Store the managed client in the pool.
	p.mu.Lock()
	p.clients[cfg.ID] = &ManagedClient{
		ServerID:  cfg.ID,
		Config:    cfg,
		Client:    c,
		Tools:     tools,
		Connected: true,
	}
	p.mu.Unlock()

	log.Printf("mcp pool: connected to %s (%s), discovered %d tools", cfg.Name, cfg.ID, len(tools))
	return nil
}

// Disconnect closes the MCP client for the given server and removes it from the pool.
func (p *ClientPool) Disconnect(serverID uuid.UUID) error {
	p.mu.Lock()
	mc, ok := p.clients[serverID]
	if !ok {
		p.mu.Unlock()
		return fmt.Errorf("mcp client %s not found in pool", serverID)
	}
	delete(p.clients, serverID)
	p.mu.Unlock()

	if err := mc.Client.Close(); err != nil {
		log.Printf("mcp pool: error closing client %s: %v", serverID, err)
	}

	// Update status to stopped (best-effort; use background context since
	// the caller may not care about DB persistence failures here).
	ctx := context.Background()
	if err := p.service.UpdateStatus(ctx, serverID, "stopped", ""); err != nil {
		log.Printf("mcp pool: failed to update status to stopped for %s: %v", serverID, err)
	}

	log.Printf("mcp pool: disconnected %s", serverID)
	return nil
}

// CallTool invokes a tool on the specified MCP server and returns the result.
func (p *ClientPool) CallTool(ctx context.Context, serverID uuid.UUID, toolName string, args map[string]any) (*mcp.CallToolResult, error) {
	p.mu.RLock()
	mc, ok := p.clients[serverID]
	p.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("mcp client %s not found in pool", serverID)
	}
	if !mc.Connected {
		return nil, fmt.Errorf("mcp client %s is not connected", serverID)
	}

	req := mcp.CallToolRequest{}
	req.Params.Name = toolName
	req.Params.Arguments = args

	result, err := mc.Client.CallTool(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("call tool %q on %s: %w", toolName, serverID, err)
	}

	return result, nil
}

// DiscoverTools re-discovers tools from the specified MCP server,
// updates the database cache, and returns the discovered tools.
func (p *ClientPool) DiscoverTools(ctx context.Context, serverID uuid.UUID) ([]mcp.Tool, error) {
	p.mu.RLock()
	mc, ok := p.clients[serverID]
	p.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("mcp client %s not found in pool", serverID)
	}
	if !mc.Connected {
		return nil, fmt.Errorf("mcp client %s is not connected", serverID)
	}

	toolsResult, err := mc.Client.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return nil, fmt.Errorf("discover tools for %s: %w", serverID, err)
	}

	tools := toolsResult.Tools

	// Convert and update the cache.
	toolDefs, err := convertTools(serverID, tools)
	if err != nil {
		return nil, fmt.Errorf("convert tools for %s: %w", serverID, err)
	}

	if err := p.service.CacheTools(ctx, serverID, toolDefs); err != nil {
		return nil, fmt.Errorf("cache tools for %s: %w", serverID, err)
	}

	// Update the in-memory tool list.
	p.mu.Lock()
	mc.Tools = tools
	p.mu.Unlock()

	log.Printf("mcp pool: re-discovered %d tools for %s", len(tools), serverID)
	return tools, nil
}

// ReconnectAll attempts to reconnect all servers in the given configs list.
// Servers with status "stopped" are skipped. Errors are logged but do not
// prevent other servers from reconnecting (best-effort recovery).
func (p *ClientPool) ReconnectAll(ctx context.Context, configs []ServerConfig) {
	for _, cfg := range configs {
		if cfg.Status == "stopped" {
			log.Printf("mcp pool: skipping stopped server %s (%s)", cfg.Name, cfg.ID)
			continue
		}

		// Check for context cancellation before each connection attempt.
		if ctx.Err() != nil {
			log.Printf("mcp pool: context cancelled, aborting reconnect")
			return
		}

		log.Printf("mcp pool: reconnecting to %s (%s)", cfg.Name, cfg.ID)
		if err := p.Connect(ctx, cfg); err != nil {
			log.Printf("mcp pool: failed to reconnect %s (%s): %v", cfg.Name, cfg.ID, err)
		}
	}
}

// ShutdownAll closes all active MCP client connections and clears the pool.
func (p *ClientPool) ShutdownAll() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for id, mc := range p.clients {
		if err := mc.Client.Close(); err != nil {
			log.Printf("mcp pool: error closing client %s during shutdown: %v", id, err)
		} else {
			log.Printf("mcp pool: closed client %s", id)
		}
	}

	// Clear the map.
	p.clients = make(map[uuid.UUID]*ManagedClient)
	log.Printf("mcp pool: all clients shut down")
}

// convertTools converts a slice of mcp.Tool from the MCP protocol into
// the internal ToolDef representation used for database caching.
func convertTools(serverID uuid.UUID, tools []mcp.Tool) ([]ToolDef, error) {
	defs := make([]ToolDef, 0, len(tools))
	for _, t := range tools {
		// Marshal the InputSchema (ToolArgumentsSchema) to json.RawMessage.
		schemaBytes, err := json.Marshal(t.InputSchema)
		if err != nil {
			return nil, fmt.Errorf("marshal input schema for tool %q: %w", t.Name, err)
		}

		defs = append(defs, ToolDef{
			ServerID:    serverID,
			Name:        t.Name,
			Description: t.Description,
			InputSchema: json.RawMessage(schemaBytes),
		})
	}
	return defs, nil
}

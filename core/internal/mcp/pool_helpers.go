package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	mcpConnectTimeout = 60 * time.Second
	mcpStatusTimeout  = 3 * time.Second
)

func withMCPConnectTimeout(ctx context.Context, fn func(context.Context) error) error {
	connectCtx, cancel := context.WithTimeout(ctx, mcpConnectTimeout)
	defer cancel()
	return fn(connectCtx)
}

func mcpStatusContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.WithoutCancel(ctx), mcpStatusTimeout)
}

func (p *ClientPool) updateStatus(ctx context.Context, id uuid.UUID, status string, message string) error {
	statusCtx, cancel := mcpStatusContext(ctx)
	defer cancel()
	return p.service.UpdateStatus(statusCtx, id, status, message)
}

func convertTools(serverID uuid.UUID, tools []mcp.Tool) ([]ToolDef, error) {
	defs := make([]ToolDef, 0, len(tools))
	for _, t := range tools {
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

package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
)

const mcpConnectTimeout = 15 * time.Second

func withMCPConnectTimeout(ctx context.Context, fn func(context.Context) error) error {
	connectCtx, cancel := context.WithTimeout(ctx, mcpConnectTimeout)
	defer cancel()
	return fn(connectCtx)
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

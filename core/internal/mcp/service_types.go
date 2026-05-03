package mcp

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ServerConfig represents an MCP server registration in the registry.
type ServerConfig struct {
	ID        uuid.UUID         `json:"id"`
	Name      string            `json:"name"`
	Transport string            `json:"transport"`
	Command   string            `json:"command,omitempty"`
	Args      []string          `json:"args,omitempty"`
	Env       map[string]string `json:"env,omitempty"`
	URL       string            `json:"url,omitempty"`
	Headers   map[string]string `json:"headers,omitempty"`
	Status    string            `json:"status"`
	Error     string            `json:"error,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// ToolDef represents a tool exposed by an MCP server.
type ToolDef struct {
	ID          uuid.UUID       `json:"id"`
	ServerID    uuid.UUID       `json:"server_id"`
	ServerName  string          `json:"server_name,omitempty"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"input_schema"`
}

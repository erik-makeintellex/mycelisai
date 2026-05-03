package mcp

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

func (s *Service) ListTools(ctx context.Context, serverID uuid.UUID) ([]ToolDef, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT t.id, t.server_id, t.name, t.description, t.input_schema
		FROM mcp_tools t
		WHERE t.server_id = $1
		ORDER BY t.name ASC
	`, serverID)
	if err != nil {
		return nil, fmt.Errorf("list tools for server %s: %w", serverID, err)
	}
	defer rows.Close()

	var tools []ToolDef
	for rows.Next() {
		var t ToolDef
		var desc sql.NullString
		if err := rows.Scan(&t.ID, &t.ServerID, &t.Name, &desc, &t.InputSchema); err != nil {
			return nil, fmt.Errorf("scan tool: %w", err)
		}
		if desc.Valid {
			t.Description = desc.String
		}
		tools = append(tools, t)
	}
	return tools, rows.Err()
}

func (s *Service) ListAllTools(ctx context.Context) ([]ToolDef, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT t.id, t.server_id, s.name, t.name, t.description, t.input_schema
		FROM mcp_tools t
		JOIN mcp_servers s ON s.id = t.server_id
		ORDER BY s.name ASC, t.name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("list all tools: %w", err)
	}
	defer rows.Close()

	var tools []ToolDef
	for rows.Next() {
		var t ToolDef
		var desc sql.NullString
		if err := rows.Scan(&t.ID, &t.ServerID, &t.ServerName, &t.Name, &desc, &t.InputSchema); err != nil {
			return nil, fmt.Errorf("scan tool: %w", err)
		}
		if desc.Valid {
			t.Description = desc.String
		}
		tools = append(tools, t)
	}
	return tools, rows.Err()
}

func (s *Service) FindToolByName(ctx context.Context, name string) (*ToolDef, *ServerConfig, error) {
	row := s.DB.QueryRowContext(ctx, `
		SELECT
			t.id, t.server_id, s.name, t.name, t.description, t.input_schema,
			s.id, s.name, s.transport, s.command, s.args, s.env, s.url, s.headers, s.status, s.error_message, s.created_at, s.updated_at
		FROM mcp_tools t
		JOIN mcp_servers s ON s.id = t.server_id
		WHERE t.name = $1
		LIMIT 1
	`, name)

	var tool ToolDef
	var srv ServerConfig
	var toolDesc sql.NullString
	var argsJSON, envJSON, headersJSON []byte
	var srvErrMsg sql.NullString

	err := row.Scan(
		&tool.ID, &tool.ServerID, &tool.ServerName, &tool.Name, &toolDesc, &tool.InputSchema,
		&srv.ID, &srv.Name, &srv.Transport, &srv.Command,
		&argsJSON, &envJSON, &srv.URL, &headersJSON,
		&srv.Status, &srvErrMsg, &srv.CreatedAt, &srv.UpdatedAt,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("find tool by name %q: %w", name, err)
	}
	if toolDesc.Valid {
		tool.Description = toolDesc.String
	}
	if err := decodeServerJSON(&srv, argsJSON, envJSON, headersJSON, srvErrMsg); err != nil {
		return nil, nil, err
	}
	return &tool, &srv, nil
}

func (s *Service) FindServerByName(ctx context.Context, name string) (*ServerConfig, error) {
	row := s.DB.QueryRowContext(ctx, `
		SELECT id, name, transport, command, args, env, url, headers, status, error_message, created_at, updated_at
		FROM mcp_servers
		WHERE name = $1
	`, name)

	var srv ServerConfig
	var argsJSON, envJSON, headersJSON []byte
	var errMsg sql.NullString

	err := row.Scan(
		&srv.ID, &srv.Name, &srv.Transport, &srv.Command,
		&argsJSON, &envJSON, &srv.URL, &headersJSON,
		&srv.Status, &errMsg, &srv.CreatedAt, &srv.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find server by name %q: %w", name, err)
	}
	if err := decodeServerJSON(&srv, argsJSON, envJSON, headersJSON, errMsg); err != nil {
		return nil, err
	}
	return &srv, nil
}

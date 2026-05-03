package mcp

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// Service manages MCP server and tool registration in the database.
type Service struct {
	DB       *sql.DB
	ToolSets *ToolSetService // optional: for seeding default tool sets during bootstrap
}

// NewService creates a new MCP registry service.
func NewService(db *sql.DB) *Service {
	return &Service{DB: db}
}

// Install registers a new MCP server and returns it with the generated ID.
func (s *Service) Install(ctx context.Context, cfg ServerConfig) (*ServerConfig, error) {
	argsJSON, err := json.Marshal(cfg.Args)
	if err != nil {
		return nil, fmt.Errorf("marshal args: %w", err)
	}
	envJSON, err := json.Marshal(cfg.Env)
	if err != nil {
		return nil, fmt.Errorf("marshal env: %w", err)
	}
	headersJSON, err := json.Marshal(cfg.Headers)
	if err != nil {
		return nil, fmt.Errorf("marshal headers: %w", err)
	}

	var result ServerConfig
	var argsOut, envOut, headersOut []byte
	var errMsg sql.NullString

	err = s.DB.QueryRowContext(ctx, `
		INSERT INTO mcp_servers (name, transport, command, args, env, url, headers)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (name) DO UPDATE
		SET transport = EXCLUDED.transport,
		    command = EXCLUDED.command,
		    args = EXCLUDED.args,
		    env = EXCLUDED.env,
		    url = EXCLUDED.url,
		    headers = EXCLUDED.headers,
		    status = 'installed',
		    error_message = NULL,
		    updated_at = NOW()
		RETURNING id, name, transport, command, args, env, url, headers, status, error_message, created_at, updated_at
	`, cfg.Name, cfg.Transport, cfg.Command, argsJSON, envJSON, cfg.URL, headersJSON).Scan(
		&result.ID, &result.Name, &result.Transport, &result.Command,
		&argsOut, &envOut, &result.URL, &headersOut,
		&result.Status, &errMsg, &result.CreatedAt, &result.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("install mcp server: %w", err)
	}

	if errMsg.Valid {
		result.Error = errMsg.String
	}
	if err := json.Unmarshal(argsOut, &result.Args); err != nil {
		return nil, fmt.Errorf("unmarshal args: %w", err)
	}
	if err := json.Unmarshal(envOut, &result.Env); err != nil {
		return nil, fmt.Errorf("unmarshal env: %w", err)
	}
	if err := json.Unmarshal(headersOut, &result.Headers); err != nil {
		return nil, fmt.Errorf("unmarshal headers: %w", err)
	}

	return &result, nil
}

// List returns all registered MCP servers.
func (s *Service) List(ctx context.Context) ([]ServerConfig, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT id, name, transport, command, args, env, url, headers, status, error_message, created_at, updated_at
		FROM mcp_servers
		ORDER BY created_at ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("list mcp servers: %w", err)
	}
	defer rows.Close()

	var servers []ServerConfig
	for rows.Next() {
		srv, err := scanServerConfig(rows)
		if err != nil {
			return nil, err
		}
		servers = append(servers, *srv)
	}
	return servers, rows.Err()
}

// Get retrieves a single MCP server by ID.
func (s *Service) Get(ctx context.Context, id uuid.UUID) (*ServerConfig, error) {
	row := s.DB.QueryRowContext(ctx, `
		SELECT id, name, transport, command, args, env, url, headers, status, error_message, created_at, updated_at
		FROM mcp_servers
		WHERE id = $1
	`, id)

	var srv ServerConfig
	var argsJSON, envJSON, headersJSON []byte
	var errMsg sql.NullString

	err := row.Scan(
		&srv.ID, &srv.Name, &srv.Transport, &srv.Command,
		&argsJSON, &envJSON, &srv.URL, &headersJSON,
		&srv.Status, &errMsg, &srv.CreatedAt, &srv.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get mcp server %s: %w", id, err)
	}
	if err := decodeServerJSON(&srv, argsJSON, envJSON, headersJSON, errMsg); err != nil {
		return nil, err
	}

	return &srv, nil
}

// Delete removes an MCP server by ID. CASCADE deletes associated tools.
func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := s.DB.ExecContext(ctx, `DELETE FROM mcp_servers WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete mcp server %s: %w", id, err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete mcp server rows affected: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("mcp server %s not found", id)
	}
	return nil
}

// UpdateStatus sets the status and optional error message for an MCP server.
func (s *Service) UpdateStatus(ctx context.Context, id uuid.UUID, status string, errMsg string) error {
	var nullErrMsg sql.NullString
	if errMsg != "" {
		nullErrMsg = sql.NullString{String: errMsg, Valid: true}
	}

	result, err := s.DB.ExecContext(ctx, `
		UPDATE mcp_servers
		SET status = $1, error_message = $2, updated_at = NOW()
		WHERE id = $3
	`, status, nullErrMsg, id)
	if err != nil {
		return fmt.Errorf("update mcp server status %s: %w", id, err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("update status rows affected: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("mcp server %s not found", id)
	}
	return nil
}

// CacheTools replaces the cached tool definitions for an MCP server.
// It deletes all existing tools for the server, then inserts the new set.
// Uses ON CONFLICT DO UPDATE in case of race conditions.
func (s *Service) CacheTools(ctx context.Context, serverID uuid.UUID, tools []ToolDef) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Clear existing tools for this server
	_, err = tx.ExecContext(ctx, `DELETE FROM mcp_tools WHERE server_id = $1`, serverID)
	if err != nil {
		return fmt.Errorf("delete existing tools: %w", err)
	}

	// Insert new tools
	for _, t := range tools {
		schemaJSON := t.InputSchema
		if schemaJSON == nil {
			schemaJSON = json.RawMessage(`{}`)
		}

		_, err = tx.ExecContext(ctx, `
			INSERT INTO mcp_tools (server_id, name, description, input_schema)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (server_id, name) DO UPDATE
			SET description = EXCLUDED.description, input_schema = EXCLUDED.input_schema
		`, serverID, t.Name, t.Description, schemaJSON)
		if err != nil {
			return fmt.Errorf("insert tool %q: %w", t.Name, err)
		}
	}

	return tx.Commit()
}

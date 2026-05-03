package mcp

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

func scanServerConfig(rows *sql.Rows) (*ServerConfig, error) {
	var srv ServerConfig
	var argsJSON, envJSON, headersJSON []byte
	var errMsg sql.NullString

	err := rows.Scan(
		&srv.ID, &srv.Name, &srv.Transport, &srv.Command,
		&argsJSON, &envJSON, &srv.URL, &headersJSON,
		&srv.Status, &errMsg, &srv.CreatedAt, &srv.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan server config: %w", err)
	}
	if err := decodeServerJSON(&srv, argsJSON, envJSON, headersJSON, errMsg); err != nil {
		return nil, err
	}
	return &srv, nil
}

func decodeServerJSON(srv *ServerConfig, argsJSON, envJSON, headersJSON []byte, errMsg sql.NullString) error {
	if errMsg.Valid {
		srv.Error = errMsg.String
	}
	if err := json.Unmarshal(argsJSON, &srv.Args); err != nil {
		return fmt.Errorf("unmarshal args: %w", err)
	}
	if err := json.Unmarshal(envJSON, &srv.Env); err != nil {
		return fmt.Errorf("unmarshal env: %w", err)
	}
	if err := json.Unmarshal(headersJSON, &srv.Headers); err != nil {
		return fmt.Errorf("unmarshal headers: %w", err)
	}
	return nil
}

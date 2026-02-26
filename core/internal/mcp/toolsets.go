package mcp

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ToolSet represents a named bundle of MCP tool references.
// Agents reference tool sets via "toolset:<name>" in their Tools[] manifest field.
type ToolSet struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ToolRefs    []string  `json:"tool_refs"`
	TenantID    string    `json:"tenant_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ToolSetService manages CRUD for MCP tool sets and resolves toolset: references.
type ToolSetService struct {
	DB *sql.DB
}

// NewToolSetService creates a new tool set service.
func NewToolSetService(db *sql.DB) *ToolSetService {
	return &ToolSetService{DB: db}
}

// List returns all tool sets for the default tenant.
func (s *ToolSetService) List(ctx context.Context) ([]ToolSet, error) {
	if s.DB == nil {
		return nil, fmt.Errorf("database not available")
	}

	rows, err := s.DB.QueryContext(ctx,
		`SELECT id, name, COALESCE(description,''), tool_refs, tenant_id, created_at, updated_at
		 FROM mcp_tool_sets WHERE tenant_id = 'default' ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list tool sets: %w", err)
	}
	defer rows.Close()

	var sets []ToolSet
	for rows.Next() {
		var ts ToolSet
		var refsJSON []byte
		if err := rows.Scan(&ts.ID, &ts.Name, &ts.Description, &refsJSON, &ts.TenantID, &ts.CreatedAt, &ts.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan tool set: %w", err)
		}
		if err := json.Unmarshal(refsJSON, &ts.ToolRefs); err != nil {
			ts.ToolRefs = []string{}
		}
		sets = append(sets, ts)
	}
	return sets, rows.Err()
}

// Get retrieves a tool set by ID.
func (s *ToolSetService) Get(ctx context.Context, id uuid.UUID) (*ToolSet, error) {
	if s.DB == nil {
		return nil, fmt.Errorf("database not available")
	}

	var ts ToolSet
	var refsJSON []byte
	err := s.DB.QueryRowContext(ctx,
		`SELECT id, name, COALESCE(description,''), tool_refs, tenant_id, created_at, updated_at
		 FROM mcp_tool_sets WHERE id = $1`, id).
		Scan(&ts.ID, &ts.Name, &ts.Description, &refsJSON, &ts.TenantID, &ts.CreatedAt, &ts.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get tool set: %w", err)
	}
	if err := json.Unmarshal(refsJSON, &ts.ToolRefs); err != nil {
		ts.ToolRefs = []string{}
	}
	return &ts, nil
}

// Create inserts a new tool set.
func (s *ToolSetService) Create(ctx context.Context, ts ToolSet) (*ToolSet, error) {
	if s.DB == nil {
		return nil, fmt.Errorf("database not available")
	}
	if ts.ToolRefs == nil {
		ts.ToolRefs = []string{}
	}

	refsJSON, err := json.Marshal(ts.ToolRefs)
	if err != nil {
		return nil, fmt.Errorf("marshal tool_refs: %w", err)
	}

	err = s.DB.QueryRowContext(ctx,
		`INSERT INTO mcp_tool_sets (name, description, tool_refs, tenant_id)
		 VALUES ($1, $2, $3, 'default')
		 RETURNING id, created_at, updated_at`,
		ts.Name, ts.Description, refsJSON).
		Scan(&ts.ID, &ts.CreatedAt, &ts.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create tool set: %w", err)
	}
	ts.TenantID = "default"
	return &ts, nil
}

// Update modifies an existing tool set.
func (s *ToolSetService) Update(ctx context.Context, id uuid.UUID, ts ToolSet) (*ToolSet, error) {
	if s.DB == nil {
		return nil, fmt.Errorf("database not available")
	}
	if ts.ToolRefs == nil {
		ts.ToolRefs = []string{}
	}

	refsJSON, err := json.Marshal(ts.ToolRefs)
	if err != nil {
		return nil, fmt.Errorf("marshal tool_refs: %w", err)
	}

	err = s.DB.QueryRowContext(ctx,
		`UPDATE mcp_tool_sets SET name = $1, description = $2, tool_refs = $3, updated_at = NOW()
		 WHERE id = $4
		 RETURNING id, name, COALESCE(description,''), tool_refs, tenant_id, created_at, updated_at`,
		ts.Name, ts.Description, refsJSON, id).
		Scan(&ts.ID, &ts.Name, &ts.Description, &refsJSON, &ts.TenantID, &ts.CreatedAt, &ts.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tool set not found")
	}
	if err != nil {
		return nil, fmt.Errorf("update tool set: %w", err)
	}
	if err := json.Unmarshal(refsJSON, &ts.ToolRefs); err != nil {
		ts.ToolRefs = []string{}
	}
	return &ts, nil
}

// Delete removes a tool set by ID.
func (s *ToolSetService) Delete(ctx context.Context, id uuid.UUID) error {
	if s.DB == nil {
		return fmt.Errorf("database not available")
	}

	result, err := s.DB.ExecContext(ctx, `DELETE FROM mcp_tool_sets WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete tool set: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("tool set not found")
	}
	return nil
}

// FindByName looks up a tool set by name.
func (s *ToolSetService) FindByName(ctx context.Context, name string) (*ToolSet, error) {
	if s.DB == nil {
		return nil, fmt.Errorf("database not available")
	}

	var ts ToolSet
	var refsJSON []byte
	err := s.DB.QueryRowContext(ctx,
		`SELECT id, name, COALESCE(description,''), tool_refs, tenant_id, created_at, updated_at
		 FROM mcp_tool_sets WHERE name = $1 AND tenant_id = 'default'`, name).
		Scan(&ts.ID, &ts.Name, &ts.Description, &refsJSON, &ts.TenantID, &ts.CreatedAt, &ts.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find tool set by name: %w", err)
	}
	if err := json.Unmarshal(refsJSON, &ts.ToolRefs); err != nil {
		ts.ToolRefs = []string{}
	}
	return &ts, nil
}

// ResolveRefs takes a Tools[] list containing mixed internal, mcp:, and toolset:
// references and expands toolset: entries into their constituent mcp: references.
// Returns a flat list with no toolset: entries remaining.
func (s *ToolSetService) ResolveRefs(ctx context.Context, tools []string) ([]string, error) {
	var resolved []string
	for _, t := range tools {
		if IsToolSetRef(t) {
			name := ToolSetName(t)
			ts, err := s.FindByName(ctx, name)
			if err != nil {
				return nil, fmt.Errorf("resolve toolset %q: %w", name, err)
			}
			if ts != nil {
				resolved = append(resolved, ts.ToolRefs...)
			}
			// Skip silently if tool set not found (may not be created yet)
		} else {
			resolved = append(resolved, t)
		}
	}
	return resolved, nil
}

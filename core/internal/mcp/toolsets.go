package mcp

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
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
	ScopeKind   string    `json:"scope_kind"`
	ScopeRef    string    `json:"scope_ref,omitempty"`
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

func normalizeToolSetScope(kind, ref string) (string, string, error) {
	kind = strings.ToLower(strings.TrimSpace(kind))
	ref = strings.TrimSpace(ref)
	if kind == "" {
		kind = "all"
	}
	switch kind {
	case "all":
		return kind, "", nil
	case "group", "host":
		if ref == "" {
			return "", "", fmt.Errorf("scope_ref is required when scope_kind is %q", kind)
		}
		return kind, ref, nil
	default:
		return "", "", fmt.Errorf("unsupported scope_kind %q", kind)
	}
}

func (ts *ToolSet) normalizeScope() error {
	kind, ref, err := normalizeToolSetScope(ts.ScopeKind, ts.ScopeRef)
	if err != nil {
		return err
	}
	ts.ScopeKind = kind
	ts.ScopeRef = ref
	return nil
}

// List returns all tool sets for the default tenant.
func (s *ToolSetService) List(ctx context.Context) ([]ToolSet, error) {
	if s.DB == nil {
		return nil, fmt.Errorf("database not available")
	}

	rows, err := s.DB.QueryContext(ctx,
		`SELECT id, name, COALESCE(description,''), tool_refs, COALESCE(scope_kind,'all'), COALESCE(scope_ref,''), tenant_id, created_at, updated_at
		 FROM mcp_tool_sets WHERE tenant_id = 'default' ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list tool sets: %w", err)
	}
	defer rows.Close()

	var sets []ToolSet
	for rows.Next() {
		var ts ToolSet
		var refsJSON []byte
		if err := rows.Scan(&ts.ID, &ts.Name, &ts.Description, &refsJSON, &ts.ScopeKind, &ts.ScopeRef, &ts.TenantID, &ts.CreatedAt, &ts.UpdatedAt); err != nil {
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
		`SELECT id, name, COALESCE(description,''), tool_refs, COALESCE(scope_kind,'all'), COALESCE(scope_ref,''), tenant_id, created_at, updated_at
		 FROM mcp_tool_sets WHERE id = $1`, id).
		Scan(&ts.ID, &ts.Name, &ts.Description, &refsJSON, &ts.ScopeKind, &ts.ScopeRef, &ts.TenantID, &ts.CreatedAt, &ts.UpdatedAt)
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
	if err := ts.normalizeScope(); err != nil {
		return nil, err
	}

	refsJSON, err := json.Marshal(ts.ToolRefs)
	if err != nil {
		return nil, fmt.Errorf("marshal tool_refs: %w", err)
	}

	err = s.DB.QueryRowContext(ctx,
		`INSERT INTO mcp_tool_sets (name, description, tool_refs, scope_kind, scope_ref, tenant_id)
		 VALUES ($1, $2, $3, $4, NULLIF($5, ''), 'default')
		 RETURNING id, created_at, updated_at`,
		ts.Name, ts.Description, refsJSON, ts.ScopeKind, ts.ScopeRef).
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
	if err := ts.normalizeScope(); err != nil {
		return nil, err
	}

	refsJSON, err := json.Marshal(ts.ToolRefs)
	if err != nil {
		return nil, fmt.Errorf("marshal tool_refs: %w", err)
	}

	err = s.DB.QueryRowContext(ctx,
		`UPDATE mcp_tool_sets SET name = $1, description = $2, tool_refs = $3, scope_kind = $4, scope_ref = NULLIF($5, ''), updated_at = NOW()
		 WHERE id = $6
		 RETURNING id, name, COALESCE(description,''), tool_refs, COALESCE(scope_kind,'all'), COALESCE(scope_ref,''), tenant_id, created_at, updated_at`,
		ts.Name, ts.Description, refsJSON, ts.ScopeKind, ts.ScopeRef, id).
		Scan(&ts.ID, &ts.Name, &ts.Description, &refsJSON, &ts.ScopeKind, &ts.ScopeRef, &ts.TenantID, &ts.CreatedAt, &ts.UpdatedAt)
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

// FindByName looks up a shared all-scope tool set by name.
func (s *ToolSetService) FindByName(ctx context.Context, name string) (*ToolSet, error) {
	return s.FindByNameForScope(ctx, name, "all", "")
}

// FindByNameForScope looks up a tool set by name and explicit scope.
func (s *ToolSetService) FindByNameForScope(ctx context.Context, name, scopeKind, scopeRef string) (*ToolSet, error) {
	if s.DB == nil {
		return nil, fmt.Errorf("database not available")
	}
	scopeKind, scopeRef, err := normalizeToolSetScope(scopeKind, scopeRef)
	if err != nil {
		return nil, err
	}

	var ts ToolSet
	var refsJSON []byte
	err = s.DB.QueryRowContext(ctx,
		`SELECT id, name, COALESCE(description,''), tool_refs, COALESCE(scope_kind,'all'), COALESCE(scope_ref,''), tenant_id, created_at, updated_at
		 FROM mcp_tool_sets WHERE name = $1 AND tenant_id = 'default' AND scope_kind = $2 AND COALESCE(scope_ref,'') = $3`, name, scopeKind, scopeRef).
		Scan(&ts.ID, &ts.Name, &ts.Description, &refsJSON, &ts.ScopeKind, &ts.ScopeRef, &ts.TenantID, &ts.CreatedAt, &ts.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find tool set by name and scope: %w", err)
	}
	if err := json.Unmarshal(refsJSON, &ts.ToolRefs); err != nil {
		ts.ToolRefs = []string{}
	}
	return &ts, nil
}

// ResolveRefsForScope expands toolset: refs by first checking a group/host scope,
// then falling back to the shared all-scope set with the same name.
func (s *ToolSetService) ResolveRefsForScope(ctx context.Context, tools []string, scopeKind, scopeRef string) ([]string, error) {
	var resolved []string
	for _, t := range tools {
		if !IsToolSetRef(t) {
			resolved = append(resolved, t)
			continue
		}
		name := ToolSetName(t)
		ts, err := s.FindByNameForScope(ctx, name, scopeKind, scopeRef)
		if err != nil {
			return nil, fmt.Errorf("resolve scoped toolset %q: %w", name, err)
		}
		if ts == nil && scopeKind != "" && strings.ToLower(scopeKind) != "all" {
			ts, err = s.FindByName(ctx, name)
			if err != nil {
				return nil, fmt.Errorf("resolve fallback toolset %q: %w", name, err)
			}
		}
		if ts != nil {
			resolved = append(resolved, ts.ToolRefs...)
		}
	}
	return resolved, nil
}

// ResolveRefs takes a Tools[] list containing mixed internal, mcp:, and toolset:
// references and expands toolset: entries into their constituent mcp: references.
// Returns a flat list with no toolset: entries remaining.
func (s *ToolSetService) ResolveRefs(ctx context.Context, tools []string) ([]string, error) {
	return s.ResolveRefsForScope(ctx, tools, "all", "")
}

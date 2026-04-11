package conversationtemplates

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

type Store struct {
	db *sql.DB
}

type ListFilter struct {
	TenantID string
	Scope    string
	Status   string
	Limit    int
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Create(ctx context.Context, raw protocol.ConversationTemplate) (*protocol.ConversationTemplate, error) {
	if s.db == nil {
		return nil, fmt.Errorf("conversation templates: database not available")
	}
	tpl := protocol.NormalizeConversationTemplate(raw)
	if err := validateTemplate(tpl); err != nil {
		return nil, err
	}
	variables, outputContract, teamShape, modelHint, governanceTags, err := marshalTemplateJSON(tpl)
	if err != nil {
		return nil, err
	}
	row := s.db.QueryRowContext(ctx, `
		INSERT INTO conversation_templates
			(tenant_id, name, description, scope, created_by, creator_kind, status, template_body,
			 variables, output_contract, recommended_team_shape, model_routing_hint, governance_tags)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb, $10::jsonb, $11::jsonb, $12::jsonb, $13::jsonb)
		RETURNING id, tenant_id, name, description, scope, created_by, creator_kind, status, template_body,
			variables::text, output_contract::text, recommended_team_shape::text, model_routing_hint::text,
			governance_tags::text, created_at, updated_at, last_used_at
	`,
		tpl.TenantID, tpl.Name, tpl.Description, string(tpl.Scope), tpl.CreatedBy, string(tpl.CreatorKind),
		string(tpl.Status), tpl.TemplateBody, variables, outputContract, teamShape, modelHint, governanceTags,
	)
	return scanTemplate(row)
}

func (s *Store) List(ctx context.Context, filter ListFilter) ([]protocol.ConversationTemplate, error) {
	if s.db == nil {
		return nil, fmt.Errorf("conversation templates: database not available")
	}
	tenantID := strings.TrimSpace(filter.TenantID)
	if tenantID == "" {
		tenantID = "default"
	}
	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	conditions := []string{"tenant_id = $1"}
	args := []any{tenantID}
	nextArg := 2
	if scope := strings.TrimSpace(filter.Scope); scope != "" {
		conditions = append(conditions, fmt.Sprintf("scope = $%d", nextArg))
		args = append(args, scope)
		nextArg++
	}
	if status := strings.TrimSpace(filter.Status); status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", nextArg))
		args = append(args, status)
		nextArg++
	}
	args = append(args, limit)
	query := fmt.Sprintf(`
		SELECT id, tenant_id, name, description, scope, created_by, creator_kind, status, template_body,
			variables::text, output_contract::text, recommended_team_shape::text, model_routing_hint::text,
			governance_tags::text, created_at, updated_at, last_used_at
		FROM conversation_templates
		WHERE %s
		ORDER BY updated_at DESC
		LIMIT $%d
	`, strings.Join(conditions, " AND "), nextArg)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("conversation templates: list failed: %w", err)
	}
	defer rows.Close()
	return scanTemplates(rows)
}

func (s *Store) Get(ctx context.Context, id string) (*protocol.ConversationTemplate, error) {
	if s.db == nil {
		return nil, fmt.Errorf("conversation templates: database not available")
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("conversation templates: id is required")
	}
	row := s.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, name, description, scope, created_by, creator_kind, status, template_body,
			variables::text, output_contract::text, recommended_team_shape::text, model_routing_hint::text,
			governance_tags::text, created_at, updated_at, last_used_at
		FROM conversation_templates
		WHERE id = $1
	`, id)
	tpl, err := scanTemplate(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if err != nil {
		return nil, fmt.Errorf("conversation templates: get failed: %w", err)
	}
	return tpl, nil
}

func (s *Store) Update(ctx context.Context, raw protocol.ConversationTemplate) (*protocol.ConversationTemplate, error) {
	if s.db == nil {
		return nil, fmt.Errorf("conversation templates: database not available")
	}
	tpl := protocol.NormalizeConversationTemplate(raw)
	if strings.TrimSpace(tpl.ID) == "" {
		return nil, fmt.Errorf("conversation templates: id is required")
	}
	if err := validateTemplate(tpl); err != nil {
		return nil, err
	}
	variables, outputContract, teamShape, modelHint, governanceTags, err := marshalTemplateJSON(tpl)
	if err != nil {
		return nil, err
	}
	row := s.db.QueryRowContext(ctx, `
		UPDATE conversation_templates
		SET name = $2, description = $3, scope = $4, creator_kind = $5, status = $6, template_body = $7,
			variables = $8::jsonb, output_contract = $9::jsonb, recommended_team_shape = $10::jsonb,
			model_routing_hint = $11::jsonb, governance_tags = $12::jsonb, updated_at = NOW()
		WHERE id = $1
		RETURNING id, tenant_id, name, description, scope, created_by, creator_kind, status, template_body,
			variables::text, output_contract::text, recommended_team_shape::text, model_routing_hint::text,
			governance_tags::text, created_at, updated_at, last_used_at
	`,
		tpl.ID, tpl.Name, tpl.Description, string(tpl.Scope), string(tpl.CreatorKind), string(tpl.Status),
		tpl.TemplateBody, variables, outputContract, teamShape, modelHint, governanceTags,
	)
	return scanTemplate(row)
}

func (s *Store) Instantiate(ctx context.Context, id string, variables map[string]any) (*protocol.ConversationTemplateInstantiation, error) {
	tpl, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	instantiation := protocol.InstantiateConversationTemplate(*tpl, variables)
	if s.db != nil {
		if _, err := s.db.ExecContext(ctx, `UPDATE conversation_templates SET last_used_at = NOW(), updated_at = NOW() WHERE id = $1`, tpl.ID); err != nil {
			return nil, fmt.Errorf("conversation templates: update last used failed: %w", err)
		}
	}
	return &instantiation, nil
}

func validateTemplate(tpl protocol.ConversationTemplate) error {
	if strings.TrimSpace(tpl.Name) == "" {
		return fmt.Errorf("conversation templates: name is required")
	}
	if strings.TrimSpace(tpl.TemplateBody) == "" {
		return fmt.Errorf("conversation templates: template_body is required")
	}
	return nil
}

func marshalTemplateJSON(tpl protocol.ConversationTemplate) (string, string, string, string, string, error) {
	variables, err := json.Marshal(tpl.Variables)
	if err != nil {
		return "", "", "", "", "", fmt.Errorf("conversation templates: marshal variables: %w", err)
	}
	outputContract, err := json.Marshal(tpl.OutputContract)
	if err != nil {
		return "", "", "", "", "", fmt.Errorf("conversation templates: marshal output contract: %w", err)
	}
	teamShape, err := json.Marshal(tpl.RecommendedTeamShape)
	if err != nil {
		return "", "", "", "", "", fmt.Errorf("conversation templates: marshal team shape: %w", err)
	}
	modelHint, err := json.Marshal(tpl.ModelRoutingHint)
	if err != nil {
		return "", "", "", "", "", fmt.Errorf("conversation templates: marshal model routing hint: %w", err)
	}
	governanceTags, err := json.Marshal(tpl.GovernanceTags)
	if err != nil {
		return "", "", "", "", "", fmt.Errorf("conversation templates: marshal governance tags: %w", err)
	}
	return string(variables), string(outputContract), string(teamShape), string(modelHint), string(governanceTags), nil
}

type templateScanner interface {
	Scan(dest ...any) error
}

func scanTemplate(row templateScanner) (*protocol.ConversationTemplate, error) {
	var tpl protocol.ConversationTemplate
	var scope, creatorKind, status string
	var variables, outputContract, teamShape, modelHint, governanceTags string
	var lastUsedAt sql.NullTime
	err := row.Scan(
		&tpl.ID, &tpl.TenantID, &tpl.Name, &tpl.Description, &scope, &tpl.CreatedBy, &creatorKind, &status,
		&tpl.TemplateBody, &variables, &outputContract, &teamShape, &modelHint, &governanceTags,
		&tpl.CreatedAt, &tpl.UpdatedAt, &lastUsedAt,
	)
	if err != nil {
		return nil, err
	}
	tpl.Scope = protocol.ConversationTemplateScope(scope)
	tpl.CreatorKind = protocol.ConversationTemplateCreatorKind(creatorKind)
	tpl.Status = protocol.ConversationTemplateStatus(status)
	tpl.Variables = mapFromJSON(variables)
	tpl.OutputContract = mapFromJSON(outputContract)
	tpl.RecommendedTeamShape = mapFromJSON(teamShape)
	tpl.ModelRoutingHint = mapFromJSON(modelHint)
	tpl.GovernanceTags = stringSliceFromJSON(governanceTags)
	if lastUsedAt.Valid {
		tpl.LastUsedAt = &lastUsedAt.Time
	}
	normalized := protocol.NormalizeConversationTemplate(tpl)
	return &normalized, nil
}

func scanTemplates(rows *sql.Rows) ([]protocol.ConversationTemplate, error) {
	var out []protocol.ConversationTemplate
	for rows.Next() {
		tpl, err := scanTemplate(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *tpl)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if out == nil {
		out = []protocol.ConversationTemplate{}
	}
	return out, nil
}

func mapFromJSON(raw string) map[string]any {
	out := map[string]any{}
	_ = json.Unmarshal([]byte(raw), &out)
	return out
}

func stringSliceFromJSON(raw string) []string {
	var out []string
	_ = json.Unmarshal([]byte(raw), &out)
	if out == nil {
		out = []string{}
	}
	return out
}

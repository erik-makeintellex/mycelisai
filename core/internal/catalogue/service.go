package catalogue

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mycelis/core/pkg/protocol"
)

// AgentTemplate is a reusable agent definition stored in the catalogue.
type AgentTemplate struct {
	ID                   uuid.UUID                       `json:"id"`
	ProfileKey           string                          `json:"profile_key,omitempty"`
	Name                 string                          `json:"name"`
	Description          string                          `json:"description,omitempty"`
	Role                 string                          `json:"role"` // cognitive, sensory, actuation, ledger
	Source               string                          `json:"source"`
	Locked               bool                            `json:"locked"`
	SystemPrompt         string                          `json:"system_prompt,omitempty"`
	Model                string                          `json:"model,omitempty"`
	Tools                []string                        `json:"tools"`
	CapabilityRefs       []string                        `json:"capability_refs"`
	ContextBindings      []protocol.AgentContextBinding  `json:"context_bindings"`
	UsagePolicy          protocol.AgentUsagePolicy       `json:"usage_policy"`
	Inputs               []string                        `json:"inputs"`
	Outputs              []string                        `json:"outputs"`
	VerificationStrategy string                          `json:"verification_strategy,omitempty"`
	VerificationRubric   []string                        `json:"verification_rubric"`
	ValidationCommand    string                          `json:"validation_command,omitempty"`
	ProfileSnapshot      *protocol.WorkerProfileSnapshot `json:"profile_snapshot,omitempty"`
	CreatedAt            time.Time                       `json:"created_at"`
	UpdatedAt            time.Time                       `json:"updated_at"`
}

// Service manages CRUD operations on the agent catalogue.
type Service struct {
	DB *sql.DB
}

// NewService creates a new catalogue service.
func NewService(db *sql.DB) *Service {
	return &Service{DB: db}
}

// List returns all reusable worker profiles ordered by name.
func (s *Service) List(ctx context.Context) ([]AgentTemplate, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT id, name, role, system_prompt, model, tools, inputs, outputs,
		       verification_strategy, verification_rubric, validation_command,
		       profile_key, description, source, locked, capability_refs, context_bindings, usage_policy,
		       created_at, updated_at
		FROM agent_catalogue
		ORDER BY name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("list catalogue: %w", err)
	}
	defer rows.Close()

	var agents []AgentTemplate
	for rows.Next() {
		a, err := scanAgent(rows)
		if err != nil {
			return nil, err
		}
		agents = append(agents, *a)
	}
	return agents, rows.Err()
}

// Get retrieves a single worker profile by ID.
func (s *Service) Get(ctx context.Context, id uuid.UUID) (*AgentTemplate, error) {
	row := s.DB.QueryRowContext(ctx, `
		SELECT id, name, role, system_prompt, model, tools, inputs, outputs,
		       verification_strategy, verification_rubric, validation_command,
		       profile_key, description, source, locked, capability_refs, context_bindings, usage_policy,
		       created_at, updated_at
		FROM agent_catalogue
		WHERE id = $1
	`, id)

	a := &AgentTemplate{}
	if err := scanAgentRow(row, a); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("agent %s not found", id)
		}
		return nil, fmt.Errorf("get agent %s: %w", id, err)
	}
	return a, nil
}

// Create inserts a new worker profile and returns it with generated ID.
func (s *Service) Create(ctx context.Context, a AgentTemplate) (*AgentTemplate, error) {
	values := prepareAgentTemplate(&a)

	row := s.DB.QueryRowContext(ctx, `
		INSERT INTO agent_catalogue (name, role, system_prompt, model, tools, inputs, outputs,
		                             verification_strategy, verification_rubric, validation_command,
		                             profile_key, description, source, locked, capability_refs, context_bindings, usage_policy)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		RETURNING id, created_at, updated_at
	`, a.Name, a.Role, values.systemPrompt, values.model, values.tools, values.inputs, values.outputs,
		values.verificationStrategy, values.rubric, values.validationCommand, values.profileKey,
		values.description, a.Source, a.Locked, values.capabilities, values.context, values.usage)

	if err := row.Scan(&a.ID, &a.CreatedAt, &a.UpdatedAt); err != nil {
		return nil, fmt.Errorf("create agent: %w", err)
	}
	return &a, nil
}

// Update modifies an existing worker profile.
func (s *Service) Update(ctx context.Context, id uuid.UUID, a AgentTemplate) (*AgentTemplate, error) {
	values := prepareAgentTemplate(&a)

	result, err := s.DB.ExecContext(ctx, `
		UPDATE agent_catalogue
		SET name = $1, role = $2, system_prompt = $3, model = $4, tools = $5,
		    inputs = $6, outputs = $7, verification_strategy = $8,
		    verification_rubric = $9, validation_command = $10, profile_key = $11,
		    description = $12, source = $13, locked = $14, capability_refs = $15,
		    context_bindings = $16, usage_policy = $17, updated_at = NOW()
		WHERE id = $18 AND locked = FALSE
	`, a.Name, a.Role, values.systemPrompt, values.model, values.tools, values.inputs, values.outputs,
		values.verificationStrategy, values.rubric, values.validationCommand, values.profileKey,
		values.description, a.Source, a.Locked, values.capabilities, values.context, values.usage, id)
	if err != nil {
		return nil, fmt.Errorf("update agent %s: %w", id, err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return nil, fmt.Errorf("agent %s not found", id)
	}

	return s.Get(ctx, id)
}

// Delete removes a worker profile by ID.
func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := s.DB.ExecContext(ctx, `DELETE FROM agent_catalogue WHERE id = $1 AND locked = FALSE`, id)
	if err != nil {
		return fmt.Errorf("delete agent %s: %w", id, err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("agent %s not found", id)
	}
	return nil
}

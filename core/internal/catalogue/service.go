package catalogue

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// AgentTemplate is a reusable agent definition stored in the catalogue.
type AgentTemplate struct {
	ID                   uuid.UUID       `json:"id"`
	Name                 string          `json:"name"`
	Role                 string          `json:"role"` // cognitive, sensory, actuation, ledger
	SystemPrompt         string          `json:"system_prompt,omitempty"`
	Model                string          `json:"model,omitempty"`
	Tools                []string        `json:"tools"`
	Inputs               []string        `json:"inputs"`
	Outputs              []string        `json:"outputs"`
	VerificationStrategy string          `json:"verification_strategy,omitempty"`
	VerificationRubric   []string        `json:"verification_rubric"`
	ValidationCommand    string          `json:"validation_command,omitempty"`
	CreatedAt            time.Time       `json:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at"`
}

// Service manages CRUD operations on the agent catalogue.
type Service struct {
	DB *sql.DB
}

// NewService creates a new catalogue service.
func NewService(db *sql.DB) *Service {
	return &Service{DB: db}
}

// List returns all agent templates ordered by name.
func (s *Service) List(ctx context.Context) ([]AgentTemplate, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT id, name, role, system_prompt, model, tools, inputs, outputs,
		       verification_strategy, verification_rubric, validation_command,
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

// Get retrieves a single agent template by ID.
func (s *Service) Get(ctx context.Context, id uuid.UUID) (*AgentTemplate, error) {
	row := s.DB.QueryRowContext(ctx, `
		SELECT id, name, role, system_prompt, model, tools, inputs, outputs,
		       verification_strategy, verification_rubric, validation_command,
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

// Create inserts a new agent template and returns it with generated ID.
func (s *Service) Create(ctx context.Context, a AgentTemplate) (*AgentTemplate, error) {
	if a.Tools == nil {
		a.Tools = []string{}
	}
	if a.Inputs == nil {
		a.Inputs = []string{}
	}
	if a.Outputs == nil {
		a.Outputs = []string{}
	}
	if a.VerificationRubric == nil {
		a.VerificationRubric = []string{}
	}

	toolsJSON, _ := json.Marshal(a.Tools)
	inputsJSON, _ := json.Marshal(a.Inputs)
	outputsJSON, _ := json.Marshal(a.Outputs)
	rubricJSON, _ := json.Marshal(a.VerificationRubric)

	var sysPrompt, model, verStrat, valCmd sql.NullString
	if a.SystemPrompt != "" {
		sysPrompt = sql.NullString{String: a.SystemPrompt, Valid: true}
	}
	if a.Model != "" {
		model = sql.NullString{String: a.Model, Valid: true}
	}
	if a.VerificationStrategy != "" {
		verStrat = sql.NullString{String: a.VerificationStrategy, Valid: true}
	}
	if a.ValidationCommand != "" {
		valCmd = sql.NullString{String: a.ValidationCommand, Valid: true}
	}

	row := s.DB.QueryRowContext(ctx, `
		INSERT INTO agent_catalogue (name, role, system_prompt, model, tools, inputs, outputs,
		                             verification_strategy, verification_rubric, validation_command)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at
	`, a.Name, a.Role, sysPrompt, model, toolsJSON, inputsJSON, outputsJSON,
		verStrat, rubricJSON, valCmd)

	if err := row.Scan(&a.ID, &a.CreatedAt, &a.UpdatedAt); err != nil {
		return nil, fmt.Errorf("create agent: %w", err)
	}
	return &a, nil
}

// Update modifies an existing agent template.
func (s *Service) Update(ctx context.Context, id uuid.UUID, a AgentTemplate) (*AgentTemplate, error) {
	if a.Tools == nil {
		a.Tools = []string{}
	}
	if a.Inputs == nil {
		a.Inputs = []string{}
	}
	if a.Outputs == nil {
		a.Outputs = []string{}
	}
	if a.VerificationRubric == nil {
		a.VerificationRubric = []string{}
	}

	toolsJSON, _ := json.Marshal(a.Tools)
	inputsJSON, _ := json.Marshal(a.Inputs)
	outputsJSON, _ := json.Marshal(a.Outputs)
	rubricJSON, _ := json.Marshal(a.VerificationRubric)

	var sysPrompt, model, verStrat, valCmd sql.NullString
	if a.SystemPrompt != "" {
		sysPrompt = sql.NullString{String: a.SystemPrompt, Valid: true}
	}
	if a.Model != "" {
		model = sql.NullString{String: a.Model, Valid: true}
	}
	if a.VerificationStrategy != "" {
		verStrat = sql.NullString{String: a.VerificationStrategy, Valid: true}
	}
	if a.ValidationCommand != "" {
		valCmd = sql.NullString{String: a.ValidationCommand, Valid: true}
	}

	result, err := s.DB.ExecContext(ctx, `
		UPDATE agent_catalogue
		SET name = $1, role = $2, system_prompt = $3, model = $4, tools = $5,
		    inputs = $6, outputs = $7, verification_strategy = $8,
		    verification_rubric = $9, validation_command = $10, updated_at = NOW()
		WHERE id = $11
	`, a.Name, a.Role, sysPrompt, model, toolsJSON, inputsJSON, outputsJSON,
		verStrat, rubricJSON, valCmd, id)
	if err != nil {
		return nil, fmt.Errorf("update agent %s: %w", id, err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return nil, fmt.Errorf("agent %s not found", id)
	}

	return s.Get(ctx, id)
}

// Delete removes an agent template by ID.
func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := s.DB.ExecContext(ctx, `DELETE FROM agent_catalogue WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete agent %s: %w", id, err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("agent %s not found", id)
	}
	return nil
}

// scanAgent scans a full agent template from sql.Rows.
func scanAgent(rows *sql.Rows) (*AgentTemplate, error) {
	a := &AgentTemplate{}
	var (
		sysPrompt, model, verStrat, valCmd sql.NullString
		toolsJSON, inputsJSON, outputsJSON, rubricJSON []byte
	)

	if err := rows.Scan(
		&a.ID, &a.Name, &a.Role, &sysPrompt, &model,
		&toolsJSON, &inputsJSON, &outputsJSON,
		&verStrat, &rubricJSON, &valCmd,
		&a.CreatedAt, &a.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("scan agent: %w", err)
	}

	a.SystemPrompt = sysPrompt.String
	a.Model = model.String
	a.VerificationStrategy = verStrat.String
	a.ValidationCommand = valCmd.String

	_ = json.Unmarshal(toolsJSON, &a.Tools)
	_ = json.Unmarshal(inputsJSON, &a.Inputs)
	_ = json.Unmarshal(outputsJSON, &a.Outputs)
	_ = json.Unmarshal(rubricJSON, &a.VerificationRubric)

	if a.Tools == nil {
		a.Tools = []string{}
	}
	if a.Inputs == nil {
		a.Inputs = []string{}
	}
	if a.Outputs == nil {
		a.Outputs = []string{}
	}
	if a.VerificationRubric == nil {
		a.VerificationRubric = []string{}
	}

	return a, nil
}

// scanAgentRow scans a single row into an AgentTemplate.
func scanAgentRow(row *sql.Row, a *AgentTemplate) error {
	var (
		sysPrompt, model, verStrat, valCmd sql.NullString
		toolsJSON, inputsJSON, outputsJSON, rubricJSON []byte
	)

	if err := row.Scan(
		&a.ID, &a.Name, &a.Role, &sysPrompt, &model,
		&toolsJSON, &inputsJSON, &outputsJSON,
		&verStrat, &rubricJSON, &valCmd,
		&a.CreatedAt, &a.UpdatedAt,
	); err != nil {
		return err
	}

	a.SystemPrompt = sysPrompt.String
	a.Model = model.String
	a.VerificationStrategy = verStrat.String
	a.ValidationCommand = valCmd.String

	_ = json.Unmarshal(toolsJSON, &a.Tools)
	_ = json.Unmarshal(inputsJSON, &a.Inputs)
	_ = json.Unmarshal(outputsJSON, &a.Outputs)
	_ = json.Unmarshal(rubricJSON, &a.VerificationRubric)

	if a.Tools == nil {
		a.Tools = []string{}
	}
	if a.Inputs == nil {
		a.Inputs = []string{}
	}
	if a.Outputs == nil {
		a.Outputs = []string{}
	}
	if a.VerificationRubric == nil {
		a.VerificationRubric = []string{}
	}

	return nil
}

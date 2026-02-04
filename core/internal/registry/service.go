package registry

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/xeipuuv/gojsonschema"
)

type Service struct {
	DB *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{DB: db}
}

// --- Templates ---

func (s *Service) ListTemplates(ctx context.Context) ([]ConnectorTemplate, error) {
	// sql.DB uses QueryContext
	rows, err := s.DB.QueryContext(ctx, "SELECT id, name, type, image, config_schema, topic_template FROM connector_templates")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []ConnectorTemplate
	for rows.Next() {
		var t ConnectorTemplate
		if err := rows.Scan(&t.ID, &t.Name, &t.Type, &t.Image, &t.ConfigSchema, &t.TopicTemplate); err != nil {
			return nil, err
		}
		templates = append(templates, t)
	}
	return templates, nil
}

func (s *Service) RegisterTemplate(ctx context.Context, t ConnectorTemplate) error {
	_, err := s.DB.ExecContext(ctx,
		"INSERT INTO connector_templates (name, type, image, config_schema, topic_template) VALUES ($1, $2, $3, $4, $5)",
		t.Name, t.Type, t.Image, t.ConfigSchema, t.TopicTemplate)
	return err
}

// --- Provisioning ---

func (s *Service) InstallConnector(ctx context.Context, teamID uuid.UUID, templateID uuid.UUID, name string, config json.RawMessage) (*ActiveConnector, error) {
	// 1. Fetch Template
	var t ConnectorTemplate
	err := s.DB.QueryRowContext(ctx, "SELECT id, config_schema, topic_template FROM connector_templates WHERE id=$1", templateID).
		Scan(&t.ID, &t.ConfigSchema, &t.TopicTemplate)
	if err != nil {
		return nil, fmt.Errorf("template not found: %w", err)
	}

	// 2. Validate Config against Schema
	schemaLoader := gojsonschema.NewBytesLoader(t.ConfigSchema)
	documentLoader := gojsonschema.NewBytesLoader(config)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return nil, fmt.Errorf("schema validation error: %w", err)
	}
	if !result.Valid() {
		return nil, fmt.Errorf("invalid config: %v", result.Errors())
	}

	// 3. Create Record
	ac := &ActiveConnector{
		ID:         uuid.New(),
		TeamID:     teamID,
		TemplateID: templateID,
		Name:       name,
		Config:     config,
		Status:     "provisioning",
	}

	_, err = s.DB.ExecContext(ctx,
		"INSERT INTO active_connectors (id, team_id, template_id, name, config, status) VALUES ($1, $2, $3, $4, $5, $6)",
		ac.ID, ac.TeamID, ac.TemplateID, ac.Name, ac.Config, ac.Status)

	if err != nil {
		return nil, err
	}

	// TODO: Trigger actual K8s deployment (Phase 32 Part 2)
	// For now, we simulate success
	return ac, nil
}

// --- Wiring ---

type WiringGraph struct {
	Inputs  []string `json:"inputs"`
	Outputs []string `json:"outputs"`
}

func (s *Service) GetWiring(ctx context.Context, teamID uuid.UUID) (*WiringGraph, error) {
	// Join active_connectors with templates to find topics
	query := `
		SELECT ct.type, ct.topic_template, ac.config
		FROM active_connectors ac
		JOIN connector_templates ct ON ac.template_id = ct.id
		WHERE ac.team_id = $1
	`
	rows, err := s.DB.QueryContext(ctx, query, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	g := &WiringGraph{
		Inputs:  []string{},
		Outputs: []string{},
	}

	for rows.Next() {
		var cType, topicTpl string
		var configRaw json.RawMessage
		if err := rows.Scan(&cType, &topicTpl, &configRaw); err != nil {
			return nil, err
		}

		// Simple topic render: replace {{var}}?
		// For MVP, simplistic check or use topic_template as is if no vars?
		// Or parse config.
		// Let's just append the template name for now or "Resolved Topic"
		// If template is "swarm.data.weather.{{city}}", request has config {"city": "seattle"}.
		// We'll skip complex rendering in Go for this step and return the raw template or just "Resolved"
		// Actually, let's just return the raw template for the graph edge label.

		if cType == "ingress" {
			g.Inputs = append(g.Inputs, topicTpl)
		} else {
			g.Outputs = append(g.Outputs, topicTpl)
		}
	}
	return g, nil
}

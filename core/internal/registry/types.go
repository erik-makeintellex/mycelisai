package registry

import (
	"encoding/json"

	"github.com/google/uuid"
)

// ConnectorTemplate defines a reusable data source/sink
type ConnectorTemplate struct {
	ID            uuid.UUID       `json:"id" db:"id"`
	Name          string          `json:"name" db:"name"`
	Type          string          `json:"type" db:"type"`   // ingress/egress
	Image         string          `json:"image" db:"image"` // Docker Image
	ConfigSchema  json.RawMessage `json:"config_schema" db:"config_schema"`
	TopicTemplate string          `json:"topic_template" db:"topic_template"`
}

// ActiveConnector is a running instance of a template
type ActiveConnector struct {
	ID         uuid.UUID       `json:"id" db:"id"`
	TeamID     uuid.UUID       `json:"team_id" db:"team_id"`
	TemplateID uuid.UUID       `json:"template_id" db:"template_id"`
	Name       string          `json:"name" db:"name"`
	Config     json.RawMessage `json:"config" db:"config"`
	Status     string          `json:"status" db:"status"` // provisioning, running, error
}

// Blueprint defines a reusable agent role
type Blueprint struct {
	ID               uuid.UUID       `json:"id" db:"id"`
	Name             string          `json:"name" db:"name"`
	Role             string          `json:"role" db:"role"`
	CognitiveProfile string          `json:"cognitive_profile" db:"cognitive_profile"`
	Tools            json.RawMessage `json:"tools" db:"tools"`
}

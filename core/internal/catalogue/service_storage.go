package catalogue

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/mycelis/core/pkg/protocol"
)

type agentTemplateStorage struct {
	systemPrompt, model, verificationStrategy, validationCommand sql.NullString
	profileKey, description                                      sql.NullString
	tools, inputs, outputs, rubric                               []byte
	capabilities, context, usage                                 []byte
}

func nullableString(value string) sql.NullString {
	return sql.NullString{String: value, Valid: value != ""}
}

func prepareAgentTemplate(agent *AgentTemplate) agentTemplateStorage {
	if agent.Source == "" {
		agent.Source = "user"
	}
	if agent.Tools == nil {
		agent.Tools = []string{}
	}
	if agent.CapabilityRefs == nil {
		agent.CapabilityRefs = append([]string(nil), agent.Tools...)
	}
	if agent.ContextBindings == nil {
		agent.ContextBindings = []protocol.AgentContextBinding{}
	}
	if agent.Inputs == nil {
		agent.Inputs = []string{}
	}
	if agent.Outputs == nil {
		agent.Outputs = []string{}
	}
	if agent.VerificationRubric == nil {
		agent.VerificationRubric = []string{}
	}
	return agentTemplateStorage{
		systemPrompt: nullableString(agent.SystemPrompt), model: nullableString(agent.Model),
		verificationStrategy: nullableString(agent.VerificationStrategy), validationCommand: nullableString(agent.ValidationCommand),
		profileKey: nullableString(agent.ProfileKey), description: nullableString(agent.Description),
		tools: jsonBytes(agent.Tools), inputs: jsonBytes(agent.Inputs), outputs: jsonBytes(agent.Outputs),
		rubric: jsonBytes(agent.VerificationRubric), capabilities: jsonBytes(agent.CapabilityRefs),
		context: jsonBytes(agent.ContextBindings), usage: jsonBytes(agent.UsagePolicy),
	}
}

func jsonBytes(value any) []byte {
	raw, _ := json.Marshal(value)
	return raw
}

type agentScanner interface {
	Scan(dest ...any) error
}

func scanAgent(scanner agentScanner) (*AgentTemplate, error) {
	agent := &AgentTemplate{}
	var values agentTemplateStorage
	if err := scanner.Scan(
		&agent.ID, &agent.Name, &agent.Role, &values.systemPrompt, &values.model,
		&values.tools, &values.inputs, &values.outputs, &values.verificationStrategy,
		&values.rubric, &values.validationCommand, &values.profileKey, &values.description,
		&agent.Source, &agent.Locked, &values.capabilities, &values.context, &values.usage,
		&agent.CreatedAt, &agent.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("scan agent: %w", err)
	}
	agent.SystemPrompt = values.systemPrompt.String
	agent.Model = values.model.String
	agent.VerificationStrategy = values.verificationStrategy.String
	agent.ValidationCommand = values.validationCommand.String
	agent.ProfileKey = values.profileKey.String
	agent.Description = values.description.String
	_ = json.Unmarshal(values.tools, &agent.Tools)
	_ = json.Unmarshal(values.inputs, &agent.Inputs)
	_ = json.Unmarshal(values.outputs, &agent.Outputs)
	_ = json.Unmarshal(values.rubric, &agent.VerificationRubric)
	_ = json.Unmarshal(values.capabilities, &agent.CapabilityRefs)
	_ = json.Unmarshal(values.context, &agent.ContextBindings)
	_ = json.Unmarshal(values.usage, &agent.UsagePolicy)
	prepareAgentTemplate(agent)
	return agent, nil
}

func scanAgentRow(row *sql.Row, agent *AgentTemplate) error {
	result, err := scanAgent(row)
	if err != nil {
		return err
	}
	*agent = *result
	return nil
}

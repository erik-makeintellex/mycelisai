package protocol

import "encoding/json"

// MissionBlueprint is the Meta-Architect's structured output.
// It decomposes a high-level intent into teams, agents, and constraints.
type MissionBlueprint struct {
	MissionID    string                `json:"mission_id"`
	Intent       string                `json:"intent"`
	Teams        []BlueprintTeam       `json:"teams"`
	Constraints  []Constraint          `json:"constraints,omitempty"`
	Requirements []ResourceRequirement `json:"requirements,omitempty"`
}

// ResourceRequirement captures an external dependency the mission needs:
// MCP servers to install, API keys to configure, env vars to set.
type ResourceRequirement struct {
	Type        string `json:"type"`               // "mcp_server", "api_key", "env_var", "credential"
	Name        string `json:"name"`               // e.g. "github", "OPENAI_API_KEY", "SLACK_BOT_TOKEN"
	Description string `json:"description"`        // why it's needed
	Required    bool   `json:"required"`           // false = nice-to-have, true = mission will fail without
	Installed   bool   `json:"installed,omitempty"` // set by system: already available
}

// BlueprintTeam defines a team within a mission blueprint.
type BlueprintTeam struct {
	Name     string          `json:"name"`
	Role     string          `json:"role"`
	Agents   []AgentManifest `json:"agents"`
	Schedule *ScheduleConfig `json:"schedule,omitempty"` // Optional: run this team on a schedule
}

// Constraint captures a mission constraint. It unmarshals flexibly:
// - From a string: Constraint{Description: "the string"}
// - From an object: Constraint{ID: "c-01", Description: "..."}
type Constraint struct {
	ID          string `json:"constraint_id,omitempty"`
	Description string `json:"description"`
}

func (c *Constraint) UnmarshalJSON(data []byte) error {
	// Try string first (simple form from prompt)
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		c.Description = s
		return nil
	}
	// Try object form (what LLMs naturally produce)
	type alias Constraint
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	*c = Constraint(a)
	return nil
}

// ProofEnvelope wraps an agent's output with verification evidence.
// An agent is forbidden from delivering work without a valid proof.
type ProofEnvelope struct {
	Artifact interface{} `json:"artifact"`
	Proof    Proof       `json:"proof"`
}

// Proof contains the verification evidence for an agent's output.
type Proof struct {
	Method      VerifyStrategy `json:"method"`
	Logs        string         `json:"logs"`
	RubricScore string         `json:"rubric_score"`
	Pass        bool           `json:"pass"`
}

package protocol

import "encoding/json"

// MissionBlueprint is the Meta-Architect's structured output.
// It decomposes a high-level intent into teams, agents, and constraints.
type MissionBlueprint struct {
	MissionID   string          `json:"mission_id"`
	Intent      string          `json:"intent"`
	Teams       []BlueprintTeam `json:"teams"`
	Constraints []Constraint    `json:"constraints,omitempty"`
}

// BlueprintTeam defines a team within a mission blueprint.
type BlueprintTeam struct {
	Name   string          `json:"name"`
	Role   string          `json:"role"`
	Agents []AgentManifest `json:"agents"`
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

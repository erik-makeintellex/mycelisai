package protocol

// VerifyStrategy defines how an agent proves its work.
type VerifyStrategy string

const (
	// VerifySemantic triggers an internal LLM call using the Rubric to grade the draft.
	VerifySemantic VerifyStrategy = "semantic"
	// VerifyEmpirical executes the ValidationCommand and captures stdout/stderr.
	VerifyEmpirical VerifyStrategy = "empirical"
)

// Verification defines the proof-of-work requirements for an agent's output.
type Verification struct {
	Strategy          VerifyStrategy `json:"strategy" yaml:"strategy"`
	Rubric            []string       `json:"rubric" yaml:"rubric"`
	ValidationCommand string         `json:"validation_command,omitempty" yaml:"validation_command,omitempty"`
}

// DefaultMaxIterations is the default ReAct tool-loop limit for agents.
const DefaultMaxIterations = 3

// AgentManifest defines an agent's identity, capabilities, I/O contracts, and verification.
type AgentManifest struct {
	ID            string        `json:"id" yaml:"id"`
	Role          string        `json:"role" yaml:"role"`
	SystemPrompt  string        `json:"system_prompt,omitempty" yaml:"system_prompt,omitempty"`
	Model         string        `json:"model,omitempty" yaml:"model,omitempty"`
	Inputs        []string      `json:"inputs,omitempty" yaml:"inputs,omitempty"`
	Outputs       []string      `json:"outputs,omitempty" yaml:"outputs,omitempty"`
	Tools         []string      `json:"tools,omitempty" yaml:"tools,omitempty"`              // MCP + internal tool names bound to this agent
	MaxIterations int           `json:"max_iterations,omitempty" yaml:"max_iterations,omitempty"` // ReAct loop limit (0 = DefaultMaxIterations)
	Verification  *Verification `json:"verification,omitempty" yaml:"verification,omitempty"`
}

// EffectiveMaxIterations returns the ReAct loop limit, using the default if unset.
func (m AgentManifest) EffectiveMaxIterations() int {
	if m.MaxIterations > 0 {
		return m.MaxIterations
	}
	return DefaultMaxIterations
}

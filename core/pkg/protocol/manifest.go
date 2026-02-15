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

// AgentManifest defines an agent's identity, capabilities, I/O contracts, and verification.
type AgentManifest struct {
	ID           string        `json:"id" yaml:"id"`
	Role         string        `json:"role" yaml:"role"`
	SystemPrompt string        `json:"system_prompt,omitempty" yaml:"system_prompt,omitempty"`
	Model        string        `json:"model,omitempty" yaml:"model,omitempty"`
	Inputs       []string      `json:"inputs,omitempty" yaml:"inputs,omitempty"`
	Outputs      []string      `json:"outputs,omitempty" yaml:"outputs,omitempty"`
	Tools        []string      `json:"tools,omitempty" yaml:"tools,omitempty"`        // MCP tool names bound to this agent
	Verification *Verification `json:"verification,omitempty" yaml:"verification,omitempty"`
}

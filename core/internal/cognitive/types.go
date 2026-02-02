package cognitive

import "time"

// Config represents the brain.yaml structure
type BrainConfig struct {
	Models   []Model            `yaml:"models"`
	Profiles map[string]Profile `yaml:"profiles"`
}

type Model struct {
	ID         string `yaml:"id"`
	Provider   string `yaml:"provider"` // ollama, openai
	Endpoint   string `yaml:"endpoint"`
	Name       string `yaml:"name"`         // e.g., qwen2.5:7b
	AuthKeyEnv string `yaml:"auth_key_env"` // Env var for API Key
}

// Profile defines the contract for a specific agent role
type Profile struct {
	ActiveModel   string  `yaml:"active_model"`
	FallbackModel string  `yaml:"fallback_model"` // ID of the failover provider
	Temperature   float64 `yaml:"temperature"`

	// CQA Contract
	TimeoutMs    int    `yaml:"timeout_ms"`    // e.g. 2000
	MaxRetries   int    `yaml:"max_retries"`   // e.g. 3
	OutputSchema string `yaml:"output_schema"` // "boolean", "strict_json"
	Compression  string `yaml:"compression"`   // "none", "high"
}

func (p Profile) GetTimeout() time.Duration {
	if p.TimeoutMs == 0 {
		return 30 * time.Second // Default
	}
	return time.Duration(p.TimeoutMs) * time.Millisecond
}

// CognitiveError tracks failures in the CQA loop
type CognitiveError struct {
	IsTimeout    bool
	IsValidation bool
	Attempt      int
	ModelID      string
	Message      string
}

func (e *CognitiveError) Error() string {
	return e.Message
}

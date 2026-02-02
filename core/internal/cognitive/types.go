package cognitive

import (
	"context"
)

// --- Configuration V2 ---

type BrainConfig struct {
	Providers map[string]ProviderConfig `yaml:"providers"`
	Profiles  map[string]string         `yaml:"profiles"` // ProfileName -> ProviderID
}

type ProviderConfig struct {
	Type       string `yaml:"type"`        // openai, openai_compatible, anthropic, google
	Endpoint   string `yaml:"endpoint"`    // e.g. "http://localhost:11434/v1"
	ModelID    string `yaml:"model_id"`    // e.g. "qwen2.5-coder:7b"
	AuthKey    string `yaml:"api_key"`     // Direct value (unsafe for prod)
	AuthKeyEnv string `yaml:"api_key_env"` // Env Var Name
}

// --- Interfaces ---

// LLMProvider is the universal contract for all AI backends
type LLMProvider interface {
	Infer(ctx context.Context, prompt string, opts InferOptions) (*InferResponse, error)
}

type InferOptions struct {
	Temperature float64
	MaxTokens   int
	Stop        []string
}

// --- Requests & Responses (Legacy/Compat) ---

type InferRequest struct {
	Profile string `json:"profile"`
	Prompt  string `json:"prompt"`
}

type InferResponse struct {
	Text      string `json:"text"`
	ModelUsed string `json:"model_used"`
	Provider  string `json:"provider"`
}

// --- Middleware / Contracts ---

type ValidationContract string

const (
	SchemaJSON ValidationContract = "json"
	SchemaNone ValidationContract = ""
)

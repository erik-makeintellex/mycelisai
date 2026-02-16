package cognitive

import (
	"context"
)

// --- Configuration V2 ---

type BrainConfig struct {
	Providers map[string]ProviderConfig `yaml:"providers" json:"providers"`
	Profiles  map[string]string         `yaml:"profiles" json:"profiles"` // ProfileName -> ProviderID
	Media     *MediaConfig              `yaml:"media,omitempty" json:"media,omitempty"`
}

// MediaConfig holds the Diffusers media engine endpoint (OpenAI-compatible images API).
type MediaConfig struct {
	Endpoint string `yaml:"endpoint" json:"endpoint"` // e.g. "http://127.0.0.1:8001/v1"
	ModelID  string `yaml:"model_id" json:"model_id"` // e.g. "stable-diffusion-xl"
}

type ProviderConfig struct {
	Type       string `yaml:"type" json:"type"`             // openai, openai_compatible, anthropic, google
	Driver     string `yaml:"-" json:"-"`                   // DB Driver type (mapped to Type)
	Endpoint   string `yaml:"endpoint" json:"endpoint,omitempty"` // e.g. "http://localhost:11434/v1"
	ModelID    string `yaml:"model_id" json:"model_id"`     // e.g. "qwen2.5-coder:7b"
	AuthKey    string `yaml:"api_key" json:"-"`             // NEVER expose in API responses
	AuthKeyEnv string `yaml:"api_key_env" json:"-"`         // NEVER expose in API responses
}

// --- Interfaces ---

// LLMProvider is the universal contract for all AI backends
type LLMProvider interface {
	Infer(ctx context.Context, prompt string, opts InferOptions) (*InferResponse, error)
	Probe(ctx context.Context) (bool, error) // Returns true if healthy/reachable
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
	Text       string `json:"text"`
	ModelUsed  string `json:"model_used"`
	Provider   string `json:"provider"`
	TokensUsed int    `json:"tokens_used,omitempty"`
}

// --- Embedding Interface ---

// EmbedProvider is implemented by adapters that support text embedding (e.g. Ollama, OpenAI).
// Not all LLMProviders support this — callers must type-assert.
type EmbedProvider interface {
	Embed(ctx context.Context, text string, model string) ([]float64, error)
}

// DefaultEmbedModel is the standard embedding model (768 dims, matches context_vectors table).
const DefaultEmbedModel = "nomic-embed-text"

// --- Middleware / Contracts ---

type ValidationContract string

const (
	SchemaJSON ValidationContract = "json"
	SchemaNone ValidationContract = ""
)

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

type ExecutionAvailability struct {
	Available         bool   `json:"available"`
	Code              string `json:"code,omitempty"`
	Summary           string `json:"summary"`
	RecommendedAction string `json:"recommended_action,omitempty"`
	Profile           string `json:"profile,omitempty"`
	ProviderID        string `json:"provider_id,omitempty"`
	ModelID           string `json:"model_id,omitempty"`
	SetupRequired     bool   `json:"setup_required,omitempty"`
	SetupPath         string `json:"setup_path,omitempty"`
	FallbackApplied   bool   `json:"fallback_applied,omitempty"`
}

// MediaProviderConfig describes the media provider backing Soma's image/voice outputs.
// It is intentionally explicit so local-hosted and hosted providers can be configured
// and tested through the same contract.
type MediaProviderConfig struct {
	ProviderID   string `yaml:"provider_id,omitempty" json:"provider_id,omitempty"`
	Type         string `yaml:"type,omitempty" json:"type,omitempty"`                   // openai_compatible, hosted_api, diffusers, comfyui, etc.
	Endpoint     string `yaml:"endpoint,omitempty" json:"endpoint,omitempty"`           // provider base endpoint
	ModelID      string `yaml:"model_id,omitempty" json:"model_id,omitempty"`           // model/workflow identifier
	Location     string `yaml:"location,omitempty" json:"location,omitempty"`           // local | remote
	DataBoundary string `yaml:"data_boundary,omitempty" json:"data_boundary,omitempty"` // local_only | leaves_org
	UsagePolicy  string `yaml:"usage_policy,omitempty" json:"usage_policy,omitempty"`   // local_first | require_approval | allow_escalation
	AuthKeyEnv   string `yaml:"api_key_env,omitempty" json:"api_key_env,omitempty"`
	Enabled      *bool  `yaml:"enabled,omitempty" json:"enabled,omitempty"`
}

// MediaConfig holds the media provider contract plus the legacy image endpoint fields.
// The legacy fields remain so existing YAML continues to work; the typed provider block
// makes local vs hosted configuration explicit.
type MediaConfig struct {
	Provider MediaProviderConfig `yaml:"provider,omitempty" json:"provider,omitempty"`
	Endpoint string              `yaml:"endpoint" json:"endpoint"` // e.g. "http://127.0.0.1:8001/v1"
	ModelID  string              `yaml:"model_id" json:"model_id"` // e.g. "stable-diffusion-xl"
}

type ProviderConfig struct {
	Type       string `yaml:"type" json:"type"`                   // openai, openai_compatible, anthropic, google
	Driver     string `yaml:"-" json:"-"`                         // DB Driver type (mapped to Type)
	Endpoint   string `yaml:"endpoint" json:"endpoint,omitempty"` // e.g. "http://localhost:11434/v1"
	ModelID    string `yaml:"model_id" json:"model_id"`           // e.g. "qwen2.5-coder:7b"
	AuthKey    string `yaml:"api_key" json:"-"`                   // NEVER expose in API responses
	AuthKeyEnv string `yaml:"api_key_env" json:"-"`               // NEVER expose in API responses

	// Provider orchestration metadata
	Location           string   `yaml:"location" json:"location"`                                             // "local" | "remote"
	DataBoundary       string   `yaml:"data_boundary" json:"data_boundary"`                                   // "local_only" | "leaves_org"
	UsagePolicy        string   `yaml:"usage_policy" json:"usage_policy"`                                     // "local_first" | "allow_escalation" | "require_approval" | "disallowed"
	TokenBudgetProfile string   `yaml:"token_budget_profile,omitempty" json:"token_budget_profile,omitempty"` // conservative | standard | extended | deep
	MaxOutputTokens    int      `yaml:"max_output_tokens,omitempty" json:"max_output_tokens,omitempty"`       // bounded default output budget per provider
	RolesAllowed       []string `yaml:"roles_allowed" json:"roles_allowed"`                                   // ["architect","coder"] or ["all"]
	Enabled            bool     `yaml:"enabled" json:"enabled"`
}

const (
	TokenBudgetConservative = "conservative"
	TokenBudgetStandard     = "standard"
	TokenBudgetExtended     = "extended"
	TokenBudgetDeep         = "deep"
	DefaultTokenBudget      = TokenBudgetStandard
	DefaultMaxOutputTokens  = 1024
)

func NormalizeTokenBudgetProfile(profile string) string {
	switch profile {
	case TokenBudgetConservative, TokenBudgetStandard, TokenBudgetExtended, TokenBudgetDeep:
		return profile
	default:
		return DefaultTokenBudget
	}
}

func DefaultMaxTokensForBudget(profile string) int {
	switch NormalizeTokenBudgetProfile(profile) {
	case TokenBudgetConservative:
		return 512
	case TokenBudgetExtended:
		return 2048
	case TokenBudgetDeep:
		return 4096
	default:
		return DefaultMaxOutputTokens
	}
}

func NormalizeProviderTokenDefaults(cfg ProviderConfig) ProviderConfig {
	cfg.TokenBudgetProfile = NormalizeTokenBudgetProfile(cfg.TokenBudgetProfile)
	if cfg.MaxOutputTokens <= 0 {
		cfg.MaxOutputTokens = DefaultMaxTokensForBudget(cfg.TokenBudgetProfile)
	}
	return cfg
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
	Messages    []ChatMessage // Optional: Structured messages (overrides prompt if supported)
}

// --- Requests & Responses (Legacy/Compat) ---

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type InferRequest struct {
	Profile  string        `json:"profile"`
	Provider string        `json:"provider,omitempty"` // Optional explicit provider override (bypasses profile routing)
	Prompt   string        `json:"prompt"`             // Legacy
	Messages []ChatMessage `json:"messages,omitempty"`
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

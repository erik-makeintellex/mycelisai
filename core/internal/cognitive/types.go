package cognitive

import (
	"context"
	"strings"
)

// LiteralAPIKeyGuidanceError is the exact adapter-init error message
// returned when a provider's ProviderConfig.LiteralAPIKeyIgnored is true and
// no usable key was resolved from AuthKeyEnv/env override. It replaces a
// generic "missing api key" message with an explanation of why a key the
// operator wrote directly into cognitive.yaml no longer takes effect.
const LiteralAPIKeyGuidanceError = "literal api_key is not supported; configure api_key_env with a secret reference"

// --- Configuration V2 ---

type BrainConfig struct {
	Providers map[string]ProviderConfig `yaml:"providers" json:"providers"`
	Profiles  map[string]string         `yaml:"profiles" json:"profiles"` // ProfileName -> ProviderID
	// ProfileFallbacks declares, per profile, an explicit ordered list of
	// alternate provider IDs Core may substitute when the profile's bound
	// provider is not executable. Every listed provider must share the bound
	// (primary) provider's normalized DataBoundary — local_only never lists a
	// leaves_org provider. A mismatched entry is a hard config error at
	// startup (see validateProfileFallbackBoundaries in router.go), not a
	// silent skip. Omitted or empty (the default) means no fallback: the
	// profile fails closed with a normalized unavailable error instead of
	// silently rerouting to another provider.
	ProfileFallbacks map[string][]string `yaml:"profile_fallbacks,omitempty" json:"profile_fallbacks,omitempty"`
	Media            *MediaConfig        `yaml:"media,omitempty" json:"media,omitempty"`
	// RootProvider is the provider ID every execution profile
	// (defaultExecutionProfiles) resolves to unless an operator override
	// pins it. It is set from cognitive.yaml (root_provider) or overridden
	// by MYCELIS_ROOT_PROVIDER. Precedence, highest first: an operator
	// override (non-empty MYCELIS_PROFILE_<NAME>_PROVIDER, a DB
	// system_config role.<name> overlay, or a runtime profile update), then
	// RootProvider, then the profile defaults shipped in cognitive.yaml.
	// Leaving RootProvider unset keeps today's behavior exactly. A
	// RootProvider that names a provider that is not configured or not
	// enabled is a hard config error at startup (see validateRootProvider),
	// never a silent skip.
	RootProvider string `yaml:"root_provider,omitempty" json:"root_provider,omitempty"`
	// ProfileSources records where each Profiles binding came from
	// (ProfileSourceDefault/Root/Override/Fallback). It is never read from
	// or persisted to cognitive.yaml.
	ProfileSources map[string]string `yaml:"-" json:"profile_sources,omitempty"`
	// ProfileOverrideOrigins records which store pinned each override:
	// ProfileOriginEnv, ProfileOriginDB, or ProfileOriginRuntime. Env is
	// applied after the DB, so env wins when both are set. Never persisted.
	ProfileOverrideOrigins map[string]string `yaml:"-" json:"profile_override_origins,omitempty"`
	// OverlayError is set when the DB overlay could not be read at startup,
	// so DB role.* overrides may be missing. Startup still proceeds.
	OverlayError bool `yaml:"-" json:"overlay_error,omitempty"`
	// dbProfileRows mirrors the system_config role.* rows (read at startup or
	// written since), including a row shadowed by an env override.
	dbProfileRows map[string]string
	// yamlProfileDefaults snapshots the cognitive.yaml profile bindings so
	// persistence never writes a root-derived binding back as a default.
	yamlProfileDefaults map[string]string
}

type ExecutionAvailability struct {
	Available         bool   `json:"available"`
	Code              string `json:"code,omitempty"`
	Summary           string `json:"summary"`
	RecommendedAction string `json:"recommended_action,omitempty"`
	// AdminAction is the admin-only remedy (it may name API paths or env
	// vars). It is never serialized; the server swaps it into
	// RecommendedAction for admin viewers only (UX1).
	AdminAction     string `json:"-"`
	Profile         string `json:"profile,omitempty"`
	ProviderID      string `json:"provider_id,omitempty"`
	ModelID         string `json:"model_id,omitempty"`
	SetupRequired   bool   `json:"setup_required,omitempty"`
	SetupPath       string `json:"setup_path,omitempty"`
	FallbackApplied bool   `json:"fallback_applied,omitempty"`
}

// MediaProviderConfig describes the media provider backing Soma's image/voice outputs.
// It is intentionally explicit so local-hosted and hosted providers can be configured
// and tested through the same contract.
type MediaProviderConfig struct {
	ProviderID   string `yaml:"provider_id,omitempty" json:"provider_id,omitempty"`
	Type         string `yaml:"type,omitempty" json:"type,omitempty"`                   // forge, openai_compatible, hosted_api, comfyui, etc.
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
	Type     string `yaml:"type" json:"type"`                   // openai, openai_compatible, anthropic, google
	Driver   string `yaml:"-" json:"-"`                         // DB Driver type (mapped to Type)
	Endpoint string `yaml:"endpoint" json:"endpoint,omitempty"` // e.g. "http://localhost:11434/v1"
	ModelID  string `yaml:"model_id" json:"model_id"`           // e.g. "qwen2.5-coder:7b"
	// AuthKey is an in-memory-only secret (set by MYCELIS_PROVIDER_<ID>_API_KEY
	// or an API/UI save). It is intentionally yaml:"-": it must never be read
	// from, or written back to, cognitive.yaml, a tracked file. Configure a
	// durable key via AuthKeyEnv (api_key_env), which names an environment
	// variable Core reads at runtime instead.
	AuthKey    string `yaml:"-" json:"-"`           // NEVER persisted to YAML; NEVER expose in API responses
	AuthKeyEnv string `yaml:"api_key_env" json:"-"` // NEVER expose in API responses
	// LiteralAPIKeyIgnored is set by the config loader (never by YAML/DB
	// unmarshal directly — it is yaml:"-" json:"-") when the provider's
	// source config (cognitive.yaml or the DB llm_providers overlay)
	// declared a non-empty literal api_key. Since AuthKey is never read from
	// a tracked file, that literal key would otherwise be silently
	// unusable; adapters check this to return an explicit
	// "configure api_key_env" error instead of a generic missing-key one.
	LiteralAPIKeyIgnored bool `yaml:"-" json:"-"`
	// ModelGateway marks an OpenAI-compatible provider as an external model
	// gateway boundary. It never transfers routing, approval, or proof authority.
	ModelGateway bool `yaml:"model_gateway,omitempty" json:"model_gateway,omitempty"`

	// Provider orchestration metadata
	Location           string   `yaml:"location" json:"location"`                                             // "local" | "remote"
	DataBoundary       string   `yaml:"data_boundary" json:"data_boundary"`                                   // "local_only" | "leaves_org"
	UsagePolicy        string   `yaml:"usage_policy" json:"usage_policy"`                                     // "local_first" | "allow_escalation" | "require_approval" | "disallowed"
	TokenBudgetProfile string   `yaml:"token_budget_profile,omitempty" json:"token_budget_profile,omitempty"` // conservative | standard | extended | deep
	MaxOutputTokens    int      `yaml:"max_output_tokens,omitempty" json:"max_output_tokens,omitempty"`       // bounded default output budget per provider
	BudgetClass        string   `yaml:"budget_class,omitempty" json:"budget_class,omitempty"`                 // token budget class; unset derives from data_boundary
	RolesAllowed       []string `yaml:"roles_allowed" json:"roles_allowed"`                                   // ["architect","coder"] or ["all"]
	Enabled            bool     `yaml:"enabled" json:"enabled"`
}

// Data boundary values. A provider's DataBoundary is either explicit or
// treated as DataBoundaryLocalOnly (see normalizedDataBoundary) — an
// empty/unknown boundary is never treated as safe to bridge into
// DataBoundaryLeavesOrg.
const (
	DataBoundaryLocalOnly = "local_only"
	DataBoundaryLeavesOrg = "leaves_org"
)

// normalizedDataBoundary treats an empty or unrecognized DataBoundary as
// DataBoundaryLocalOnly (fail closed): only the exact literal
// DataBoundaryLeavesOrg is ever treated as leaves_org — every other value,
// including a typo or an unknown string, normalizes to local_only rather
// than passing through unnormalized. Shipped core/config/cognitive.yaml has
// providers with data_boundary: "" (e.g. local-ollama-dev, local-sovereign);
// without this normalization those entries would silently bypass the
// same-boundary fallback check instead of being held to the safer default.
func normalizedDataBoundary(dataBoundary string) string {
	trimmed := strings.TrimSpace(dataBoundary)
	if trimmed != DataBoundaryLeavesOrg {
		return DataBoundaryLocalOnly
	}
	return trimmed
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
	Correlation InferenceCorrelation
}

// --- Requests & Responses (Legacy/Compat) ---

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type InferRequest struct {
	Profile     string               `json:"profile"`
	Provider    string               `json:"provider,omitempty"` // Optional explicit provider override (bypasses profile routing)
	Prompt      string               `json:"prompt"`             // Legacy
	Messages    []ChatMessage        `json:"messages,omitempty"`
	Correlation InferenceCorrelation `json:"-"` // Internal authoritative execution scope; never accepted from API JSON.
	Meter       *ExecutionMeter      `json:"-"` // Token budget meter for this execution; falls back to the ctx meter.
}

// InferenceCorrelation carries only identifiers already owned by the invoking
// Core runtime. Gateway adapters reduce it to one opaque deterministic token;
// raw identifiers are never sent across the model boundary.
type InferenceCorrelation struct {
	RunID   string
	TeamID  string
	AgentID string
	// TenantID keys token-budget counters and ledger rows; "" means "default".
	TenantID string
}

type InferResponse struct {
	Text               string `json:"text"`
	ModelUsed          string `json:"model_used"`
	Provider           string `json:"provider"`
	UpstreamResponseID string `json:"upstream_response_id,omitempty"`
	PromptTokens       int    `json:"prompt_tokens,omitempty"`
	CompletionTokens   int    `json:"completion_tokens,omitempty"`
	TokensUsed         int    `json:"tokens_used,omitempty"` // Compatibility alias for exact upstream total_tokens.
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

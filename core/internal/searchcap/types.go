package searchcap

import (
	"context"
	"time"
)

const (
	ProviderDisabled     = "disabled"
	ProviderLocalSources = "local_sources"
	ProviderSearXNG      = "searxng"
	ProviderLocalAPI     = "local_api"
	ProviderBrave        = "brave"
)

type Embedder interface {
	Embed(ctx context.Context, text, providerID string) ([]float64, error)
}

type Config struct {
	Provider         string
	SearXNGEndpoint  string
	LocalAPIEndpoint string
	MaxResults       int
}

type Request struct {
	Query          string   `json:"query"`
	SourceScope    string   `json:"source_scope,omitempty"`
	MaxResults     int      `json:"max_results,omitempty"`
	TimeRange      string   `json:"time_range,omitempty"`
	AllowedDomains []string `json:"allowed_domains,omitempty"`
	BlockedDomains []string `json:"blocked_domains,omitempty"`
	TeamID         string   `json:"team_id,omitempty"`
	AgentID        string   `json:"agent_id,omitempty"`
	RunID          string   `json:"run_id,omitempty"`
	Visibility     string   `json:"visibility,omitempty"`
	Types          []string `json:"types,omitempty"`
}

type Response struct {
	Query    string         `json:"query"`
	Provider string         `json:"provider"`
	Status   string         `json:"status"`
	Blocker  *Blocker       `json:"blocker,omitempty"`
	Results  []Result       `json:"results"`
	Count    int            `json:"count"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

type Status struct {
	Provider                  string   `json:"provider"`
	Enabled                   bool     `json:"enabled"`
	Configured                bool     `json:"configured"`
	SupportsLocalSources      bool     `json:"supports_local_sources"`
	SupportsPublicWeb         bool     `json:"supports_public_web"`
	SomaToolName              string   `json:"soma_tool_name"`
	DirectSomaInteraction     bool     `json:"direct_soma_interaction"`
	RequiresHostedAPIToken    bool     `json:"requires_hosted_api_token"`
	SearXNGEndpointConfigured bool     `json:"searxng_endpoint_configured,omitempty"`
	MaxResults                int      `json:"max_results"`
	Blocker                   *Blocker `json:"blocker,omitempty"`
	NextActions               []string `json:"next_actions,omitempty"`
}

type Blocker struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	NextAction string `json:"next_action"`
}

type Result struct {
	Title            string         `json:"title"`
	URL              string         `json:"url,omitempty"`
	LocalSourceID    string         `json:"local_source_id,omitempty"`
	Snippet          string         `json:"snippet"`
	SourceKind       string         `json:"source_kind"`
	TrustClass       string         `json:"trust_class,omitempty"`
	SensitivityClass string         `json:"sensitivity_class,omitempty"`
	RetrievedAt      time.Time      `json:"retrieved_at"`
	Score            float64        `json:"score,omitempty"`
	ProviderMetadata map[string]any `json:"provider_metadata,omitempty"`
}

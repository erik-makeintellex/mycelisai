package searchcap

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

var safeEnvRefPattern = regexp.MustCompile(`^[A-Z_][A-Z0-9_]*$`)

type SourceInput struct {
	Name             string `json:"name"`
	Provider         string `json:"provider,omitempty"`
	SourceType       string `json:"source_type,omitempty"`
	Type             string `json:"type,omitempty"`
	Endpoint         string `json:"endpoint,omitempty"`
	BaseURL          string `json:"base_url,omitempty"`
	ScopeKind        string `json:"scope_kind,omitempty"`
	Scope            string `json:"scope,omitempty"`
	ScopeRef         string `json:"scope_ref,omitempty"`
	Boundary         string `json:"boundary,omitempty"`
	AuthScheme       string `json:"auth_scheme,omitempty"`
	SecretRef        string `json:"secret_ref,omitempty"`
	Mode             string `json:"mode,omitempty"`
	SensitivityClass string `json:"sensitivity_class,omitempty"`
	Sensitivity      string `json:"sensitivity,omitempty"`
	TrustClass       string `json:"trust_class,omitempty"`
	Trust            string `json:"trust,omitempty"`
	Status           string `json:"status,omitempty"`
	Recovery         string `json:"recovery,omitempty"`
}

func (s *Service) ListSources() []Source {
	if s == nil {
		return []Source{}
	}
	return cloneSources(s.Status().Sources)
}

func (s *Service) AddSource(input SourceInput) (Source, error) {
	return s.AddSourceWithContext(context.Background(), input)
}

func (s *Service) AddSourceWithContext(ctx context.Context, input SourceInput) (Source, error) {
	if s == nil {
		return Source{}, errors.New("search service unavailable")
	}
	source, err := normalizeSourceInput(input)
	if err != nil {
		return Source{}, err
	}

	s.registryMu.Lock()
	defer s.registryMu.Unlock()
	source.ID = s.uniqueSourceIDLocked(source)
	if s.sourceDB != nil {
		created, err := s.sourceDB.Create(ctx, source)
		if err != nil {
			return Source{}, err
		}
		source = created
	}
	s.registry = append(s.registry, source)
	return source, nil
}

func (s *Service) appendRegisteredSources(base []Source) []Source {
	out := cloneSources(base)
	if s == nil {
		return out
	}
	s.registryMu.RLock()
	defer s.registryMu.RUnlock()
	out = append(out, cloneSources(s.registry)...)
	return out
}

func normalizeSourceInput(input SourceInput) (Source, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return Source{}, errors.New("search source name is required")
	}
	provider := normalizeSourceToken(firstString(input.Provider, input.SourceType, input.Type))
	if provider == "" {
		return Source{}, errors.New("search source provider or source_type is required")
	}
	sourceType := normalizeSourceToken(firstString(input.SourceType, input.Type, provider))
	endpoint := strings.TrimSpace(firstString(input.Endpoint, input.BaseURL))
	if endpoint == "" && requiresRegistryEndpoint(sourceType) {
		return Source{}, fmt.Errorf("endpoint is required for source_type %q", sourceType)
	}
	if endpoint != "" {
		if err := validateRegistryEndpoint(endpoint); err != nil {
			return Source{}, err
		}
		endpoint = strings.TrimRight(endpoint, "/")
	}

	scopeKind, scopeRef, err := normalizeRegistryScope(firstString(input.ScopeKind, input.Scope), input.ScopeRef)
	if err != nil {
		return Source{}, err
	}
	authScheme := normalizeRegistryAuthScheme(input.AuthScheme)
	secretRef := strings.TrimSpace(input.SecretRef)
	if requiresSecretRef(authScheme) && secretRef == "" {
		return Source{}, fmt.Errorf("secret_ref is required for auth_scheme %q", authScheme)
	}
	if secretRef != "" && !isSafeSecretRef(secretRef) {
		return Source{}, errors.New("secret_ref must name a managed secret reference, not a raw credential")
	}

	source := Source{
		Name:             name,
		Provider:         provider,
		SourceType:       sourceType,
		Endpoint:         endpoint,
		BaseURL:          endpoint,
		ScopeKind:        scopeKind,
		ScopeRef:         scopeRef,
		Boundary:         firstString(input.Boundary, "operator_configured"),
		AuthScheme:       authScheme,
		SecretRef:        secretRef,
		Mode:             normalizeRegistryValue(input.Mode, "preview"),
		SensitivityClass: normalizeRegistryValue(firstString(input.SensitivityClass, input.Sensitivity), "public"),
		TrustClass:       normalizeRegistryValue(firstString(input.TrustClass, input.Trust), "bounded_external"),
		Status:           normalizeRegistryValue(input.Status, "available"),
		Recovery:         strings.TrimSpace(input.Recovery),
	}
	return source, nil
}

func validateRegistryEndpoint(raw string) error {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return errors.New("endpoint must be an absolute http(s) URL")
	}
	if parsed.User != nil {
		return errors.New("endpoint must not include credentials")
	}
	return nil
}

func normalizeRegistryScope(raw, scopeRef string) (string, string, error) {
	scope := strings.ToLower(strings.TrimSpace(raw))
	switch scope {
	case "", "all", "everyone", "global":
		return "all", "", nil
	case "group", "team":
		ref := strings.TrimSpace(scopeRef)
		if ref == "" {
			return "", "", errors.New("scope_ref is required for group-scoped search sources")
		}
		return "group", ref, nil
	case "host", "machine":
		ref := strings.TrimSpace(scopeRef)
		if ref == "" {
			return "", "", errors.New("scope_ref is required for host-scoped search sources")
		}
		return "host", ref, nil
	default:
		return "", "", fmt.Errorf("unsupported scope %q", raw)
	}
}

func normalizeRegistryAuthScheme(raw string) string {
	normalized := normalizeSourceToken(raw)
	switch normalized {
	case "", "none", "no_auth":
		return "none"
	case "api_key", "api_token", "secret_ref", "token":
		return "api_token"
	case "bearer", "bearer_token":
		return "bearer_token"
	case "basic", "oauth", "service_managed", "mcp":
		return normalized
	default:
		return normalized
	}
}

func requiresRegistryEndpoint(sourceType string) bool {
	switch sourceType {
	case "public_web", "local_api", "client_or_public_api", "private_api", "authenticated_api":
		return true
	default:
		return false
	}
}

func requiresSecretRef(authScheme string) bool {
	switch authScheme {
	case "api_token", "bearer_token", "basic", "oauth":
		return true
	default:
		return false
	}
}

func isSafeSecretRef(raw string) bool {
	ref := strings.TrimSpace(raw)
	if safeEnvRefPattern.MatchString(ref) {
		return true
	}
	if strings.HasPrefix(ref, "env:") {
		return safeEnvRefPattern.MatchString(strings.TrimSpace(strings.TrimPrefix(ref, "env:")))
	}
	for _, prefix := range []string{"vault:", "secret:", "sm://"} {
		rest := strings.TrimSpace(strings.TrimPrefix(ref, prefix))
		if strings.HasPrefix(ref, prefix) && rest != "" && !strings.ContainsAny(rest, " \t\r\n") {
			return true
		}
	}
	return false
}

func normalizeRegistryValue(raw, fallback string) string {
	normalized := normalizeSourceToken(raw)
	if normalized == "" {
		return fallback
	}
	return normalized
}

func normalizeSourceToken(raw string) string {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	normalized = strings.ReplaceAll(normalized, " ", "_")
	normalized = strings.ReplaceAll(normalized, "-", "_")
	return normalized
}

func (s *Service) uniqueSourceIDLocked(source Source) string {
	base := source.Provider
	if base == "" {
		base = source.Name
	}
	base = slugSourceID(base)
	if base == "" {
		base = "search_source"
	}
	used := map[string]bool{}
	if configuredID := configuredSourceID(s.Provider()); configuredID != "" {
		used[configuredID] = true
	}
	for _, existing := range s.registry {
		used[existing.ID] = true
	}
	if !used[base] {
		return base
	}
	for i := 2; ; i++ {
		candidate := fmt.Sprintf("%s_%d", base, i)
		if !used[candidate] {
			return candidate
		}
	}
}

func configuredSourceID(provider string) string {
	switch provider {
	case ProviderLocalSources:
		return "local_sources"
	case ProviderSearXNG:
		return "searxng"
	case ProviderLocalAPI:
		return "local_api"
	case ProviderBrave:
		return "brave-search"
	default:
		return ""
	}
}

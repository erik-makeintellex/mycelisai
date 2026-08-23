package searchcap

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
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
		var err error
		endpoint, err = normalizeRegistryEndpoint(sourceType, endpoint)
		if err != nil {
			return Source{}, err
		}
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

	return Source{
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
	}, nil
}

func normalizeRegistryEndpoint(sourceType, raw string) (string, error) {
	if isMountedFolderSourceType(sourceType) || isCodeContextSourceType(sourceType) {
		return normalizeMountedFolderPath(raw)
	}
	if err := validateRegistryEndpoint(raw); err != nil {
		return "", err
	}
	return strings.TrimRight(strings.TrimSpace(raw), "/"), nil
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

func normalizeMountedFolderPath(raw string) (string, error) {
	path := strings.TrimSpace(raw)
	if path == "" {
		return "", errors.New("path is required for mounted folder sources")
	}
	if strings.ContainsRune(path, '\x00') {
		return "", errors.New("path must not contain null bytes")
	}
	if isAbsoluteHTTPURL(path) {
		return "", errors.New("mounted folder source path must be a local or shared filesystem path, not an HTTP URL")
	}
	return filepath.ToSlash(filepath.Clean(path)), nil
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
	if isMountedFolderSourceType(sourceType) || isCodeContextSourceType(sourceType) {
		return true
	}
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

func isMountedFolderSourceType(sourceType string) bool {
	switch normalizeSourceToken(sourceType) {
	case ProviderMountedFolder, "local_mount", "data_mount", "shared_folder", "mounted_files":
		return true
	default:
		return false
	}
}

func isCodeContextSourceType(sourceType string) bool {
	switch normalizeSourceToken(sourceType) {
	case ProviderCodeContext, "repository", "code_repository", "local_code_folder", "code_folder":
		return true
	default:
		return false
	}
}

func normalizeSourceToken(raw string) string {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	normalized = strings.ReplaceAll(normalized, " ", "_")
	normalized = strings.ReplaceAll(normalized, "-", "_")
	return normalized
}

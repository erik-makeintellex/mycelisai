package swarm

import (
	"fmt"
	"strings"
)

func normalizeProviderScope(scope ProviderScope) ProviderScope {
	scope.Provider = strings.TrimSpace(scope.Provider)
	scope.AllowedProviders = normalizeProviderList(scope.AllowedProviders)
	scope.RoleProviders = normalizeProviderMap(scope.RoleProviders)
	return scope
}

func normalizeScopeMap(scopes map[string]ProviderScope) map[string]ProviderScope {
	if len(scopes) == 0 {
		return nil
	}
	normalized := make(map[string]ProviderScope, len(scopes))
	for key, scope := range scopes {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		normalized[key] = normalizeProviderScope(scope)
	}
	if len(normalized) == 0 {
		return nil
	}
	return normalized
}

func normalizeProviderList(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, 0, len(in))
	seen := make(map[string]struct{}, len(in))
	for _, value := range in {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func normalizeProviderMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for key, value := range in {
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" || value == "" {
			continue
		}
		out[key] = value
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func mergeAllowedProviders(current map[string]struct{}, next []string) map[string]struct{} {
	next = normalizeProviderList(next)
	if len(next) == 0 {
		return cloneProviderSet(current)
	}
	nextSet := make(map[string]struct{}, len(next))
	for _, provider := range next {
		nextSet[provider] = struct{}{}
	}
	if current == nil {
		return nextSet
	}
	merged := make(map[string]struct{})
	for provider := range current {
		if _, ok := nextSet[provider]; ok {
			merged[provider] = struct{}{}
		}
	}
	return merged
}

func cloneProviderSet(in map[string]struct{}) map[string]struct{} {
	if in == nil {
		return nil
	}
	out := make(map[string]struct{}, len(in))
	for key := range in {
		out[key] = struct{}{}
	}
	return out
}

func providerAllowed(allowed map[string]struct{}, provider string) bool {
	provider = strings.TrimSpace(provider)
	if provider == "" {
		return false
	}
	if allowed == nil {
		return true
	}
	_, ok := allowed[provider]
	return ok
}

func roleProviderForScope(roleProviders map[string]string, role string) string {
	if len(roleProviders) == 0 {
		return ""
	}
	return strings.TrimSpace(roleProviders[strings.TrimSpace(role)])
}

func isKernelTeam(manifest *TeamManifest) bool {
	if manifest == nil {
		return false
	}
	if strings.TrimSpace(manifest.ID) == "admin-core" {
		return true
	}
	for _, member := range manifest.Members {
		if strings.TrimSpace(member.Role) == "admin" {
			return true
		}
	}
	return false
}

func isCouncilTeam(manifest *TeamManifest) bool {
	if manifest == nil {
		return false
	}
	return strings.TrimSpace(manifest.ID) == "council-core"
}

func (b BlockedProviderOverride) String() string {
	return fmt.Sprintf("%s[%s] -> %s (%s)", b.Scope, b.Target, b.Provider, b.Reason)
}

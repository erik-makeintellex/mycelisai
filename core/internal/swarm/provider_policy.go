package swarm

import (
	"fmt"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
	"gopkg.in/yaml.v3"
)

// ProviderPolicy defines scoped provider-routing defaults and constraints for a runtime organization.
type ProviderPolicy struct {
	Metadata         map[string]string        `yaml:"-"`
	Provider         string                   `yaml:"provider,omitempty"`
	AllowedProviders []string                 `yaml:"allowed_providers,omitempty"`
	RoleProviders    map[string]string        `yaml:"role_providers,omitempty"`
	Kernel           ProviderScope            `yaml:"kernel,omitempty"`
	Council          ProviderScope            `yaml:"council,omitempty"`
	Teams            map[string]ProviderScope `yaml:"teams,omitempty"`
	Agents           map[string]ProviderScope `yaml:"agents,omitempty"`
}

// ProviderScope describes one inheritance layer in the routing tree.
type ProviderScope struct {
	Provider         string            `yaml:"provider,omitempty"`
	AllowedProviders []string          `yaml:"allowed_providers,omitempty"`
	RoleProviders    map[string]string `yaml:"role_providers,omitempty"`
}

// BlockedProviderOverride records a provider candidate that was rejected by higher-scope policy.
type BlockedProviderOverride struct {
	Scope    string
	Target   string
	Provider string
	Reason   string
}

var reservedProviderPolicyKeys = map[string]struct{}{
	"provider":          {},
	"allowed_providers": {},
	"role_providers":    {},
	"kernel":            {},
	"council":           {},
	"teams":             {},
	"agents":            {},
}

func (p *ProviderPolicy) UnmarshalYAML(value *yaml.Node) error {
	type alias ProviderPolicy

	if value == nil || value.Kind == 0 {
		*p = ProviderPolicy{}
		return nil
	}

	var raw map[string]any
	if err := value.Decode(&raw); err != nil {
		return err
	}

	structured := false
	for key, rawValue := range raw {
		if _, ok := reservedProviderPolicyKeys[key]; ok {
			structured = true
			break
		}
		if _, ok := rawValue.(string); !ok {
			structured = true
			break
		}
	}

	if !structured {
		metadata := make(map[string]string, len(raw))
		for key, rawValue := range raw {
			value, _ := rawValue.(string)
			value = strings.TrimSpace(value)
			if value == "" {
				continue
			}
			metadata[strings.TrimSpace(key)] = value
		}
		*p = ProviderPolicy{Metadata: metadata}
		return nil
	}

	var decoded alias
	if err := value.Decode(&decoded); err != nil {
		return err
	}

	decoded.Metadata = make(map[string]string)
	for key, rawValue := range raw {
		if _, ok := reservedProviderPolicyKeys[key]; ok {
			continue
		}
		if value, ok := rawValue.(string); ok {
			value = strings.TrimSpace(value)
			if value != "" {
				decoded.Metadata[strings.TrimSpace(key)] = value
			}
		}
	}

	decoded.Provider = strings.TrimSpace(decoded.Provider)
	decoded.AllowedProviders = normalizeProviderList(decoded.AllowedProviders)
	decoded.RoleProviders = normalizeProviderMap(decoded.RoleProviders)
	decoded.Kernel = normalizeProviderScope(decoded.Kernel)
	decoded.Council = normalizeProviderScope(decoded.Council)
	decoded.Teams = normalizeScopeMap(decoded.Teams)
	decoded.Agents = normalizeScopeMap(decoded.Agents)

	*p = ProviderPolicy(decoded)
	return nil
}

func (p ProviderPolicy) Clone() ProviderPolicy {
	cloned := ProviderPolicy{
		Provider:         strings.TrimSpace(p.Provider),
		AllowedProviders: append([]string(nil), p.AllowedProviders...),
		Kernel:           p.Kernel.Clone(),
		Council:          p.Council.Clone(),
	}
	if len(p.Metadata) > 0 {
		cloned.Metadata = make(map[string]string, len(p.Metadata))
		for key, value := range p.Metadata {
			cloned.Metadata[key] = value
		}
	}
	if len(p.RoleProviders) > 0 {
		cloned.RoleProviders = make(map[string]string, len(p.RoleProviders))
		for key, value := range p.RoleProviders {
			cloned.RoleProviders[key] = value
		}
	}
	if len(p.Teams) > 0 {
		cloned.Teams = make(map[string]ProviderScope, len(p.Teams))
		for key, scope := range p.Teams {
			cloned.Teams[key] = scope.Clone()
		}
	}
	if len(p.Agents) > 0 {
		cloned.Agents = make(map[string]ProviderScope, len(p.Agents))
		for key, scope := range p.Agents {
			cloned.Agents[key] = scope.Clone()
		}
	}
	return cloned
}

func (s ProviderScope) Clone() ProviderScope {
	cloned := ProviderScope{
		Provider:         strings.TrimSpace(s.Provider),
		AllowedProviders: append([]string(nil), s.AllowedProviders...),
	}
	if len(s.RoleProviders) > 0 {
		cloned.RoleProviders = make(map[string]string, len(s.RoleProviders))
		for key, value := range s.RoleProviders {
			cloned.RoleProviders[key] = value
		}
	}
	return cloned
}

func (p ProviderPolicy) IsEmpty() bool {
	return strings.TrimSpace(p.Provider) == "" &&
		len(p.AllowedProviders) == 0 &&
		len(p.RoleProviders) == 0 &&
		p.Kernel.isEmpty() &&
		p.Council.isEmpty() &&
		len(p.Teams) == 0 &&
		len(p.Agents) == 0
}

func (s ProviderScope) isEmpty() bool {
	return strings.TrimSpace(s.Provider) == "" &&
		len(s.AllowedProviders) == 0 &&
		len(s.RoleProviders) == 0
}

// ResolveManifest applies provider-policy inheritance and policy-bounded overrides to one team manifest.
func (p ProviderPolicy) ResolveManifest(manifest *TeamManifest) (*TeamManifest, []BlockedProviderOverride) {
	if manifest == nil {
		return nil, nil
	}
	if p.IsEmpty() {
		return manifest, nil
	}

	out := *manifest
	out.Members = make([]protocol.AgentManifest, len(manifest.Members))
	copy(out.Members, manifest.Members)

	teamProvider, _, blocked := p.resolveTeamProvider(manifest)
	out.Provider = teamProvider

	for idx := range out.Members {
		memberProvider, memberBlocked := p.resolveMemberProvider(manifest, manifest.Members[idx], out.Provider)
		blocked = append(blocked, memberBlocked...)
		out.Members[idx].Provider = memberProvider
	}

	return &out, blocked
}

func (p ProviderPolicy) resolveTeamProvider(manifest *TeamManifest) (string, map[string]struct{}, []BlockedProviderOverride) {
	allowed := map[string]struct{}(nil)
	provider := ""
	blocked := make([]BlockedProviderOverride, 0)

	scopes := []scopeCandidate{
		{name: "organization", target: "organization", provider: p.Provider, allowed: p.AllowedProviders},
	}
	if isKernelTeam(manifest) {
		scopes = append(scopes, scopeCandidate{name: "kernel", target: manifest.ID, provider: p.Kernel.Provider, allowed: p.Kernel.AllowedProviders})
	}
	if isCouncilTeam(manifest) {
		scopes = append(scopes, scopeCandidate{name: "council", target: manifest.ID, provider: p.Council.Provider, allowed: p.Council.AllowedProviders})
	}
	if teamScope, ok := p.Teams[manifest.ID]; ok {
		scopes = append(scopes, scopeCandidate{name: "team_policy", target: manifest.ID, provider: teamScope.Provider, allowed: teamScope.AllowedProviders})
	}
	scopes = append(scopes, scopeCandidate{name: "team_manifest", target: manifest.ID, provider: manifest.Provider})

	for _, scope := range scopes {
		allowed = mergeAllowedProviders(allowed, scope.allowed)
		if scope.provider == "" {
			continue
		}
		if providerAllowed(allowed, scope.provider) {
			provider = scope.provider
			continue
		}
		blocked = append(blocked, BlockedProviderOverride{
			Scope:    scope.name,
			Target:   scope.target,
			Provider: scope.provider,
			Reason:   "provider is outside inherited allowed_providers",
		})
	}

	return provider, allowed, blocked
}

func (p ProviderPolicy) resolveMemberProvider(manifest *TeamManifest, member protocol.AgentManifest, inheritedTeamProvider string) (string, []BlockedProviderOverride) {
	allowed := map[string]struct{}(nil)
	provider := ""
	blocked := make([]BlockedProviderOverride, 0)

	scopes := []scopeCandidate{
		{
			name:         "organization",
			target:       manifest.ID,
			provider:     p.Provider,
			roleProvider: roleProviderForScope(p.RoleProviders, member.Role),
			allowed:      p.AllowedProviders,
		},
	}
	if isKernelTeam(manifest) {
		scopes = append(scopes, scopeCandidate{
			name:         "kernel",
			target:       manifest.ID,
			provider:     p.Kernel.Provider,
			roleProvider: roleProviderForScope(p.Kernel.RoleProviders, member.Role),
			allowed:      p.Kernel.AllowedProviders,
		})
	}
	if isCouncilTeam(manifest) {
		scopes = append(scopes, scopeCandidate{
			name:         "council",
			target:       member.ID,
			provider:     p.Council.Provider,
			roleProvider: roleProviderForScope(p.Council.RoleProviders, member.Role),
			allowed:      p.Council.AllowedProviders,
		})
	}
	if teamScope, ok := p.Teams[manifest.ID]; ok {
		scopes = append(scopes, scopeCandidate{
			name:         "team_policy",
			target:       manifest.ID,
			provider:     teamScope.Provider,
			roleProvider: roleProviderForScope(teamScope.RoleProviders, member.Role),
			allowed:      teamScope.AllowedProviders,
		})
	}
	scopes = append(scopes, scopeCandidate{name: "team_manifest", target: manifest.ID, provider: manifest.Provider})
	if agentScope, ok := p.Agents[member.ID]; ok {
		scopes = append(scopes, scopeCandidate{name: "agent_policy", target: member.ID, provider: agentScope.Provider, allowed: agentScope.AllowedProviders})
	}
	scopes = append(scopes, scopeCandidate{name: "agent_manifest", target: member.ID, provider: member.Provider})

	for _, scope := range scopes {
		allowed = mergeAllowedProviders(allowed, scope.allowed)
		candidate := scope.provider
		if scope.roleProvider != "" {
			candidate = scope.roleProvider
		}
		if candidate == "" {
			continue
		}
		if providerAllowed(allowed, candidate) {
			provider = candidate
			continue
		}
		blocked = append(blocked, BlockedProviderOverride{
			Scope:    scope.name,
			Target:   scope.target,
			Provider: candidate,
			Reason:   "provider is outside inherited allowed_providers",
		})
	}

	if provider == "" {
		provider = inheritedTeamProvider
	}

	return provider, blocked
}

func FallbackProviderPolicy(teamProviders, agentProviders map[string]string) ProviderPolicy {
	policy := ProviderPolicy{
		Metadata: map[string]string{
			"compatibility_source": "fallback_env_provider_maps",
		},
	}
	if len(teamProviders) > 0 {
		policy.Teams = make(map[string]ProviderScope, len(teamProviders))
		for teamID, provider := range normalizeProviderMap(teamProviders) {
			policy.Teams[teamID] = ProviderScope{Provider: provider}
		}
	}
	if len(agentProviders) > 0 {
		policy.Agents = make(map[string]ProviderScope, len(agentProviders))
		for agentID, provider := range normalizeProviderMap(agentProviders) {
			policy.Agents[agentID] = ProviderScope{Provider: provider}
		}
	}
	return policy
}

type scopeCandidate struct {
	name         string
	target       string
	provider     string
	roleProvider string
	allowed      []string
}

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

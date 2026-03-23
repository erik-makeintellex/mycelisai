package swarm

import "github.com/mycelis/core/pkg/protocol"

// ResolveManifest applies provider-policy inheritance and policy-bounded overrides to one team manifest.
func (p ProviderPolicy) ResolveManifest(manifest *TeamManifest) (*TeamManifest, []BlockedProviderOverride) {
	if manifest == nil {
		return nil, nil
	}
	if p.IsEmpty() {
		return manifest, nil
	}

	out := *manifest
	out.Members = cloneManifestMembers(manifest.Members)

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
		scopes = append(scopes, scopeCandidate{
			name:     "kernel",
			target:   manifest.ID,
			provider: p.Kernel.Provider,
			allowed:  p.Kernel.AllowedProviders,
		})
	}
	if isCouncilTeam(manifest) {
		scopes = append(scopes, scopeCandidate{
			name:     "council",
			target:   manifest.ID,
			provider: p.Council.Provider,
			allowed:  p.Council.AllowedProviders,
		})
	}
	if teamScope, ok := p.Teams[manifest.ID]; ok {
		scopes = append(scopes, scopeCandidate{
			name:     "team_policy",
			target:   manifest.ID,
			provider: teamScope.Provider,
			allowed:  teamScope.AllowedProviders,
		})
	}
	scopes = append(scopes, scopeCandidate{name: "team_manifest", target: manifest.ID, provider: manifest.Provider})

	for _, scope := range scopes {
		provider, allowed, blocked = applyProviderCandidate(provider, allowed, blocked, scope)
	}

	return provider, allowed, blocked
}

func (p ProviderPolicy) resolveMemberProvider(
	manifest *TeamManifest,
	member protocol.AgentManifest,
	inheritedTeamProvider string,
) (string, []BlockedProviderOverride) {
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
		scopes = append(scopes, scopeCandidate{
			name:     "agent_policy",
			target:   member.ID,
			provider: agentScope.Provider,
			allowed:  agentScope.AllowedProviders,
		})
	}
	scopes = append(scopes, scopeCandidate{name: "agent_manifest", target: member.ID, provider: member.Provider})

	for _, scope := range scopes {
		provider, allowed, blocked = applyProviderCandidate(provider, allowed, blocked, scope)
	}

	if provider == "" {
		provider = inheritedTeamProvider
	}

	return provider, blocked
}

type scopeCandidate struct {
	name         string
	target       string
	provider     string
	roleProvider string
	allowed      []string
}

func applyProviderCandidate(
	current string,
	allowed map[string]struct{},
	blocked []BlockedProviderOverride,
	scope scopeCandidate,
) (string, map[string]struct{}, []BlockedProviderOverride) {
	allowed = mergeAllowedProviders(allowed, scope.allowed)
	candidate := scope.provider
	if scope.roleProvider != "" {
		candidate = scope.roleProvider
	}
	if candidate == "" {
		return current, allowed, blocked
	}
	if providerAllowed(allowed, candidate) {
		return candidate, allowed, blocked
	}
	blocked = append(blocked, BlockedProviderOverride{
		Scope:    scope.name,
		Target:   scope.target,
		Provider: candidate,
		Reason:   "provider is outside inherited allowed_providers",
	})
	return current, allowed, blocked
}

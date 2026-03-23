package swarm

import (
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

	if !providerPolicyUsesStructuredFields(raw) {
		*p = ProviderPolicy{Metadata: metadataOnlyProviderPolicy(raw)}
		return nil
	}

	var decoded alias
	if err := value.Decode(&decoded); err != nil {
		return err
	}

	decoded.Metadata = metadataOnlyProviderPolicy(raw)
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
		cloned.Metadata = cloneStringMap(p.Metadata)
	}
	if len(p.RoleProviders) > 0 {
		cloned.RoleProviders = cloneStringMap(p.RoleProviders)
	}
	if len(p.Teams) > 0 {
		cloned.Teams = cloneScopeMap(p.Teams)
	}
	if len(p.Agents) > 0 {
		cloned.Agents = cloneScopeMap(p.Agents)
	}
	return cloned
}

func (s ProviderScope) Clone() ProviderScope {
	cloned := ProviderScope{
		Provider:         strings.TrimSpace(s.Provider),
		AllowedProviders: append([]string(nil), s.AllowedProviders...),
	}
	if len(s.RoleProviders) > 0 {
		cloned.RoleProviders = cloneStringMap(s.RoleProviders)
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

func providerPolicyUsesStructuredFields(raw map[string]any) bool {
	for key, rawValue := range raw {
		if _, ok := reservedProviderPolicyKeys[key]; ok {
			return true
		}
		if _, ok := rawValue.(string); !ok {
			return true
		}
	}
	return false
}

func metadataOnlyProviderPolicy(raw map[string]any) map[string]string {
	metadata := make(map[string]string)
	for key, rawValue := range raw {
		if _, ok := reservedProviderPolicyKeys[key]; ok {
			continue
		}
		value, ok := rawValue.(string)
		if !ok {
			continue
		}
		value = strings.TrimSpace(value)
		key = strings.TrimSpace(key)
		if key == "" || value == "" {
			continue
		}
		metadata[key] = value
	}
	if len(metadata) == 0 {
		return nil
	}
	return metadata
}

func cloneStringMap(in map[string]string) map[string]string {
	out := make(map[string]string, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func cloneScopeMap(in map[string]ProviderScope) map[string]ProviderScope {
	out := make(map[string]ProviderScope, len(in))
	for key, scope := range in {
		out[key] = scope.Clone()
	}
	return out
}

func cloneManifestMembers(in []protocol.AgentManifest) []protocol.AgentManifest {
	out := make([]protocol.AgentManifest, len(in))
	copy(out, in)
	return out
}

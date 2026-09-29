package swarm

import (
	"fmt"
	"strings"
	"sync"
)

// Registry manages the loading and lifecycle of Team Manifests.
type Registry struct {
	manifests []*TeamManifest
	org       *RuntimeOrganization
	mu        sync.RWMutex
}

// RuntimeOrganization is the instantiated bootstrap object that feeds runtime team activation.
type RuntimeOrganization struct {
	ID              string
	Name            string
	Description     string
	TemplateVersion string
	SourceKind      string
	KernelMode      string
	CouncilMode     string
	ProviderPolicy  ProviderPolicy
	Teams           []*TeamManifest
}

// NewRegistryFromRuntimeOrganization is the only registry constructor: the
// runtime organization instantiated from the bootstrap template bundle feeds
// team activation (the V7 teams-directory loader is retired, CONS-C3).
func NewRegistryFromRuntimeOrganization(org *RuntimeOrganization) *Registry {
	if org == nil {
		return &Registry{}
	}
	return &Registry{
		org:       org,
		manifests: append([]*TeamManifest(nil), org.Teams...),
	}
}

// LoadManifests returns a copy of the runtime organization's team manifests,
// or nil when the organization has none.
func (r *Registry) LoadManifests() ([]*TeamManifest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.manifests) == 0 {
		return nil, nil
	}
	loaded := make([]*TeamManifest, 0, len(r.manifests))
	loaded = append(loaded, r.manifests...)
	return loaded, nil
}

func (r *Registry) RuntimeOrganization() *RuntimeOrganization {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.org == nil {
		return nil
	}

	orgCopy := *r.org
	orgCopy.Teams = append([]*TeamManifest(nil), r.org.Teams...)
	orgCopy.ProviderPolicy = r.org.ProviderPolicy.Clone()
	return &orgCopy
}

func NormalizeManifest(m *TeamManifest) error {
	if m == nil {
		return fmt.Errorf("manifest is nil")
	}

	m.ID = strings.TrimSpace(m.ID)
	m.Name = strings.TrimSpace(m.Name)
	if m.ID == "" {
		if m.Name == "" {
			return fmt.Errorf("manifest is missing id and name")
		}
		m.ID = strings.ToLower(strings.ReplaceAll(m.Name, " ", "-"))
	}
	if m.Type == "" {
		m.Type = TeamTypeAction
	}

	return nil
}

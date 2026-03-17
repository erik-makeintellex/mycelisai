package swarm

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

// Registry manages the loading and lifecycle of Team Manifests.
type Registry struct {
	teamsPath string
	manifests []*TeamManifest
	org       *RuntimeOrganization
	mu        sync.RWMutex
}

// RuntimeOrganization is the instantiated bootstrap object that feeds runtime team activation.
type RuntimeOrganization struct {
	ID                string
	Name              string
	Description       string
	TemplateVersion   string
	SourceKind        string
	KernelMode        string
	CouncilMode       string
	ProviderPolicy    ProviderPolicy
	Teams             []*TeamManifest
	MigrationFallback bool // Temporary migration-only no-bundle compatibility path.
}

// NewRegistry creates a new Registry loaded from the given path.
func NewRegistry(path string) *Registry {
	return &Registry{
		teamsPath: path,
	}
}

func NewRegistryFromManifests(manifests []*TeamManifest) *Registry {
	return NewRegistryFromRuntimeOrganization(&RuntimeOrganization{
		ID:                "manifest-registry",
		Name:              "Manifest Registry",
		SourceKind:        "manifest_registry",
		Teams:             manifests,
		MigrationFallback: false,
	})
}

func NewRegistryFromRuntimeOrganization(org *RuntimeOrganization) *Registry {
	if org == nil {
		return &Registry{}
	}
	return &Registry{
		org:       org,
		manifests: append([]*TeamManifest(nil), org.Teams...),
	}
}

// LoadManifests scans the config directory and returns all found manifests.
func (r *Registry) LoadManifests() ([]*TeamManifest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.manifests) > 0 {
		loaded := make([]*TeamManifest, 0, len(r.manifests))
		loaded = append(loaded, r.manifests...)
		return loaded, nil
	}

	var manifests []*TeamManifest

	files, err := os.ReadDir(r.teamsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No teams configured yet
		}
		return nil, fmt.Errorf("failed to read teams dir: %w", err)
	}

	for _, f := range files {
		if filepath.Ext(f.Name()) == ".yaml" || filepath.Ext(f.Name()) == ".yml" {
			path := filepath.Join(r.teamsPath, f.Name())
			m, err := LoadManifestFile(path)
			if err != nil {
				log.Printf("WARN: Failed to load manifest %s: %v", f.Name(), err)
				continue
			}
			manifests = append(manifests, m)
		}
	}

	return manifests, nil
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

func LoadManifestFile(path string) (*TeamManifest, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var m TeamManifest
	if err := yaml.Unmarshal(bytes, &m); err != nil {
		return nil, err
	}

	if err := NormalizeManifest(&m); err != nil {
		return nil, err
	}

	return &m, nil
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

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
	mu        sync.RWMutex
}

// NewRegistry creates a new Registry loaded from the given path.
func NewRegistry(path string) *Registry {
	return &Registry{
		teamsPath: path,
	}
}

// LoadManifests scans the config directory and returns all found manifests.
func (r *Registry) LoadManifests() ([]*TeamManifest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

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
			m, err := r.loadManifest(path)
			if err != nil {
				log.Printf("WARN: Failed to load manifest %s: %v", f.Name(), err)
				continue
			}
			manifests = append(manifests, m)
		}
	}

	return manifests, nil
}

func (r *Registry) loadManifest(path string) (*TeamManifest, error) {
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

	// Basic Validation
	if m.ID == "" {
		m.ID = strings.ToLower(strings.ReplaceAll(m.Name, " ", "-"))
	}
	if m.Type == "" {
		m.Type = TeamTypeAction // Default
	}

	return &m, nil
}

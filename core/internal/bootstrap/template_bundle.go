package bootstrap

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/mycelis/core/internal/swarm"
	"gopkg.in/yaml.v3"
)

// TemplateBundle is the transitional Task 005 bridge from V8 bootstrap planning
// to runtime-readable organization template bundles.
type TemplateBundle struct {
	ID               string               `yaml:"id"`
	Name             string               `yaml:"name"`
	Description      string               `yaml:"description,omitempty"`
	TemplateVersion  string               `yaml:"template_version,omitempty"`
	SourceKind       string               `yaml:"source_kind,omitempty"`
	Kernel           TemplateKernel       `yaml:"kernel,omitempty"`
	Council          TemplateCouncil      `yaml:"council,omitempty"`
	ProviderPolicy   swarm.ProviderPolicy `yaml:"provider_policy,omitempty"`
	TeamManifestRefs []string             `yaml:"team_manifest_refs,omitempty"`
	baseDir          string
	configRoot       string
}

type TemplateKernel struct {
	Mode string `yaml:"mode,omitempty"`
}

type TemplateCouncil struct {
	Mode string `yaml:"mode,omitempty"`
}

type TemplateLoader struct {
	templatesPath string
}

type StartupSelection struct {
	Bundle       *TemplateBundle
	Organization *swarm.RuntimeOrganization
	Manifests    []*swarm.TeamManifest
	Source       string
}

const (
	StartupSourceBundle                 = "bundle"
	StartupSourceMigrationFallbackTeams = "migration_fallback_teams"
)

func NewTemplateLoader(path string) *TemplateLoader {
	return &TemplateLoader{templatesPath: path}
}

func (l *TemplateLoader) LoadBundles() ([]*TemplateBundle, error) {
	files, err := os.ReadDir(l.templatesPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read templates dir: %w", err)
	}

	bundles := make([]*TemplateBundle, 0, len(files))
	seen := map[string]struct{}{}
	for _, f := range files {
		if filepath.Ext(f.Name()) != ".yaml" && filepath.Ext(f.Name()) != ".yml" {
			continue
		}

		path := filepath.Join(l.templatesPath, f.Name())
		bundle, err := loadTemplateBundle(path)
		if err != nil {
			return nil, fmt.Errorf("load template bundle %s: %w", f.Name(), err)
		}
		if _, exists := seen[bundle.ID]; exists {
			return nil, fmt.Errorf("duplicate template bundle id %q", bundle.ID)
		}
		seen[bundle.ID] = struct{}{}

		if err := bundle.Validate(filepath.Dir(path), filepath.Clean(filepath.Join(l.templatesPath, ".."))); err != nil {
			return nil, fmt.Errorf("validate template bundle %s: %w", f.Name(), err)
		}
		bundles = append(bundles, bundle)
	}

	return bundles, nil
}

func loadTemplateBundle(path string) (*TemplateBundle, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var bundle TemplateBundle
	if err := yaml.Unmarshal(data, &bundle); err != nil {
		return nil, err
	}

	if bundle.TemplateVersion == "" {
		bundle.TemplateVersion = "v1alpha1"
	}
	if bundle.SourceKind == "" {
		bundle.SourceKind = "standing_team_migration_input"
	}

	return &bundle, nil
}

func (b *TemplateBundle) Validate(baseDir, configRoot string) error {
	if strings.TrimSpace(b.ID) == "" {
		return fmt.Errorf("missing id")
	}
	if strings.TrimSpace(b.Name) == "" {
		return fmt.Errorf("missing name")
	}
	if len(b.TeamManifestRefs) == 0 {
		return fmt.Errorf("team_manifest_refs must not be empty")
	}

	for _, ref := range b.TeamManifestRefs {
		ref = strings.TrimSpace(ref)
		if ref == "" {
			return fmt.Errorf("team_manifest_refs contains an empty path")
		}
		if filepath.IsAbs(ref) {
			return fmt.Errorf("team manifest ref %q must be relative", ref)
		}
		resolved := filepath.Clean(filepath.Join(baseDir, ref))
		rel, err := filepath.Rel(configRoot, resolved)
		if err != nil {
			return fmt.Errorf("resolve team manifest ref %q: %w", ref, err)
		}
		if strings.HasPrefix(rel, "..") {
			return fmt.Errorf("team manifest ref %q escapes config root", ref)
		}
		if _, err := swarm.LoadManifestFile(resolved); err != nil {
			return fmt.Errorf("load team manifest ref %q: %w", ref, err)
		}
	}

	b.baseDir = baseDir
	b.configRoot = configRoot
	return nil
}

func (b *TemplateBundle) LoadTeamManifests() ([]*swarm.TeamManifest, error) {
	if b.baseDir == "" || b.configRoot == "" {
		return nil, fmt.Errorf("template bundle %q has not been validated", b.ID)
	}

	manifests := make([]*swarm.TeamManifest, 0, len(b.TeamManifestRefs))
	for _, ref := range b.TeamManifestRefs {
		path := filepath.Clean(filepath.Join(b.baseDir, ref))
		manifest, err := swarm.LoadManifestFile(path)
		if err != nil {
			return nil, fmt.Errorf("load team manifest %q: %w", ref, err)
		}
		manifests = append(manifests, manifest)
	}
	return manifests, nil
}

func (b *TemplateBundle) InstantiateRuntimeOrganization() (*swarm.RuntimeOrganization, error) {
	manifests, err := b.LoadTeamManifests()
	if err != nil {
		return nil, err
	}
	return &swarm.RuntimeOrganization{
		ID:                b.ID,
		Name:              b.Name,
		Description:       b.Description,
		TemplateVersion:   b.TemplateVersion,
		SourceKind:        b.SourceKind,
		KernelMode:        b.Kernel.Mode,
		CouncilMode:       b.Council.Mode,
		ProviderPolicy:    b.ProviderPolicy.Clone(),
		Teams:             manifests,
		MigrationFallback: false,
	}, nil
}

func SelectStartupBundle(bundles []*TemplateBundle, requestedID string) (*TemplateBundle, error) {
	if len(bundles) == 0 {
		return nil, nil
	}

	requestedID = strings.TrimSpace(requestedID)
	if requestedID == "" {
		if len(bundles) == 1 {
			return bundles[0], nil
		}
		return nil, fmt.Errorf("multiple bootstrap template bundles available; set MYCELIS_BOOTSTRAP_TEMPLATE_ID")
	}

	for _, bundle := range bundles {
		if bundle.ID == requestedID {
			return bundle, nil
		}
	}

	return nil, fmt.Errorf("bootstrap template bundle %q not found", requestedID)
}

func ResolveStartupSelection(templatesPath, fallbackTeamsPath, requestedID string) (*StartupSelection, error) {
	templateLoader := NewTemplateLoader(templatesPath)
	bundles, err := templateLoader.LoadBundles()
	if err != nil {
		return nil, fmt.Errorf("load bootstrap template bundles: %w", err)
	}

	if len(bundles) == 0 {
		requestedID = strings.TrimSpace(requestedID)
		if requestedID != "" {
			return nil, fmt.Errorf("requested bootstrap template bundle %q was not found; no-bundle startup remains migration-only compatibility", requestedID)
		}

		// Temporary migration-only compatibility: once template bundles are universal,
		// direct team-manifest scanning should be removed instead of treated as co-equal
		// runtime truth.
		manifests, err := swarm.NewRegistry(fallbackTeamsPath).LoadManifests()
		if err != nil {
			return nil, fmt.Errorf("load fallback team manifests: %w", err)
		}
		org := instantiateFallbackRuntimeOrganization(fallbackTeamsPath, manifests)
		return &StartupSelection{
			Organization: org,
			Manifests:    manifests,
			Source:       StartupSourceMigrationFallbackTeams,
		}, nil
	}

	selected, err := SelectStartupBundle(bundles, requestedID)
	if err != nil {
		return nil, fmt.Errorf("select startup bootstrap template bundle: %w", err)
	}

	org, err := selected.InstantiateRuntimeOrganization()
	if err != nil {
		return nil, fmt.Errorf("instantiate startup bootstrap template bundle %s: %w", selected.ID, err)
	}

	return &StartupSelection{
		Bundle:       selected,
		Organization: org,
		Manifests:    append([]*swarm.TeamManifest(nil), org.Teams...),
		Source:       StartupSourceBundle,
	}, nil
}

func instantiateFallbackRuntimeOrganization(fallbackTeamsPath string, manifests []*swarm.TeamManifest) *swarm.RuntimeOrganization {
	return &swarm.RuntimeOrganization{
		ID:                "migration-fallback-standing-teams",
		Name:              "Migration Fallback Standing Teams",
		Description:       fmt.Sprintf("Temporary migration-only fallback instantiated from %s because no bootstrap template bundle was configured.", fallbackTeamsPath),
		TemplateVersion:   "migration-fallback",
		SourceKind:        "standing_team_migration_fallback",
		KernelMode:        "migration-fallback",
		CouncilMode:       "migration-fallback",
		Teams:             append([]*swarm.TeamManifest(nil), manifests...),
		MigrationFallback: true,
	}
}

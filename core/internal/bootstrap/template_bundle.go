package bootstrap

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mycelis/core/internal/swarm"
	"github.com/mycelis/core/pkg/protocol"
	"gopkg.in/yaml.v3"
)

// TemplateBundle is the transitional Task 005 bridge from V8 bootstrap planning
// to runtime-readable organization template bundles.
type TemplateBundle struct {
	ID              string                `yaml:"id"`
	Name            string                `yaml:"name"`
	Description     string                `yaml:"description,omitempty"`
	TemplateVersion string                `yaml:"template_version,omitempty"`
	SourceKind      string                `yaml:"source_kind,omitempty"`
	Kernel          TemplateKernel        `yaml:"kernel,omitempty"`
	Council         TemplateCouncil       `yaml:"council,omitempty"`
	ProviderPolicy  swarm.ProviderPolicy  `yaml:"provider_policy,omitempty"`
	Teams           []*swarm.TeamManifest `yaml:"teams,omitempty"`
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
	StartupSourceBundle = "bundle"
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

		if err := bundle.Validate(); err != nil {
			return nil, fmt.Errorf("validate template bundle %s: %w", f.Name(), err)
		}
		bundles = append(bundles, bundle)
	}

	return bundles, nil
}

func loadTemplateBundle(path string) (*TemplateBundle, error) {
	data, err := os.ReadFile(path)
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

func (b *TemplateBundle) Validate() error {
	if strings.TrimSpace(b.ID) == "" {
		return fmt.Errorf("missing id")
	}
	if strings.TrimSpace(b.Name) == "" {
		return fmt.Errorf("missing name")
	}
	if len(b.Teams) == 0 {
		return fmt.Errorf("teams must not be empty")
	}

	normalized := make([]*swarm.TeamManifest, 0, len(b.Teams))
	seen := make(map[string]struct{}, len(b.Teams))
	for idx, team := range b.Teams {
		clone, err := cloneTeamManifest(team)
		if err != nil {
			return fmt.Errorf("clone team %d: %w", idx, err)
		}
		if err := validateEmbeddedTeamManifest(idx, clone); err != nil {
			return err
		}
		if _, exists := seen[clone.ID]; exists {
			return fmt.Errorf("duplicate team id %q", clone.ID)
		}
		seen[clone.ID] = struct{}{}
		normalized = append(normalized, clone)
	}

	b.Teams = normalized
	return nil
}

func (b *TemplateBundle) LoadTeamManifests() ([]*swarm.TeamManifest, error) {
	if len(b.Teams) == 0 {
		return nil, fmt.Errorf("template bundle %q has no embedded teams", b.ID)
	}

	return cloneTeamManifests(b.Teams)
}

func (b *TemplateBundle) InstantiateRuntimeOrganization() (*swarm.RuntimeOrganization, error) {
	manifests, err := b.LoadTeamManifests()
	if err != nil {
		return nil, err
	}
	return &swarm.RuntimeOrganization{
		ID:              b.ID,
		Name:            b.Name,
		Description:     b.Description,
		TemplateVersion: b.TemplateVersion,
		SourceKind:      b.SourceKind,
		KernelMode:      b.Kernel.Mode,
		CouncilMode:     b.Council.Mode,
		ProviderPolicy:  b.ProviderPolicy.Clone(),
		Teams:           manifests,
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

func ResolveStartupSelection(templatesPath, requestedID string) (*StartupSelection, error) {
	templateLoader := NewTemplateLoader(templatesPath)
	bundles, err := templateLoader.LoadBundles()
	if err != nil {
		return nil, fmt.Errorf("load bootstrap template bundles: %w", err)
	}

	if len(bundles) == 0 {
		requestedID = strings.TrimSpace(requestedID)
		if requestedID != "" {
			return nil, fmt.Errorf("requested bootstrap template bundle %q was not found in %s; startup requires a valid bootstrap template bundle", requestedID, templatesPath)
		}
		return nil, fmt.Errorf("no bootstrap template bundles found in %s; startup requires a valid bootstrap template bundle", templatesPath)
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

func validateEmbeddedTeamManifest(idx int, manifest *swarm.TeamManifest) error {
	if manifest == nil {
		return fmt.Errorf("teams[%d] must not be null", idx)
	}
	if err := swarm.NormalizeManifest(manifest); err != nil {
		return fmt.Errorf("teams[%d]: %w", idx, err)
	}
	if strings.TrimSpace(manifest.Name) == "" {
		return fmt.Errorf("team %q missing name", manifest.ID)
	}
	if len(manifest.Members) == 0 {
		return fmt.Errorf("team %q must declare at least one member", manifest.ID)
	}
	if len(manifest.Inputs) == 0 {
		return fmt.Errorf("team %q must declare at least one input", manifest.ID)
	}
	if len(manifest.Deliveries) == 0 {
		return fmt.Errorf("team %q must declare at least one delivery", manifest.ID)
	}

	for memberIdx, member := range manifest.Members {
		if strings.TrimSpace(member.ID) == "" {
			return fmt.Errorf("team %q member %d missing id", manifest.ID, memberIdx)
		}
		if strings.TrimSpace(member.Role) == "" {
			return fmt.Errorf("team %q member %q missing role", manifest.ID, member.ID)
		}
	}
	return nil
}

func cloneTeamManifests(manifests []*swarm.TeamManifest) ([]*swarm.TeamManifest, error) {
	cloned := make([]*swarm.TeamManifest, 0, len(manifests))
	for idx, manifest := range manifests {
		clone, err := cloneTeamManifest(manifest)
		if err != nil {
			return nil, fmt.Errorf("clone team %d: %w", idx, err)
		}
		cloned = append(cloned, clone)
	}
	return cloned, nil
}

func cloneTeamManifest(manifest *swarm.TeamManifest) (*swarm.TeamManifest, error) {
	if manifest == nil {
		return nil, nil
	}

	clone := *manifest
	clone.Members = make([]protocol.AgentManifest, len(manifest.Members))
	for idx, member := range manifest.Members {
		clone.Members[idx] = cloneAgentManifest(member)
	}
	clone.Inputs = append([]string(nil), manifest.Inputs...)
	clone.Deliveries = append([]string(nil), manifest.Deliveries...)
	if manifest.Schedule != nil {
		schedule := *manifest.Schedule
		clone.Schedule = &schedule
	}
	return &clone, nil
}

func cloneAgentManifest(manifest protocol.AgentManifest) protocol.AgentManifest {
	clone := manifest
	clone.Inputs = append([]string(nil), manifest.Inputs...)
	clone.Outputs = append([]string(nil), manifest.Outputs...)
	clone.Tools = append([]string(nil), manifest.Tools...)
	if manifest.Verification != nil {
		verification := *manifest.Verification
		verification.Rubric = append([]string(nil), manifest.Verification.Rubric...)
		clone.Verification = &verification
	}
	return clone
}

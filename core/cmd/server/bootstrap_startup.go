package main

import (
	"os"
	"strings"

	"github.com/mycelis/core/internal/bootstrap"
	"github.com/mycelis/core/internal/swarm"
)

func loadStartupBundleRegistry(templatesPath, fallbackTeamsPath string) (*bootstrap.StartupSelection, *swarm.Registry, error) {
	selection, err := bootstrap.ResolveStartupSelection(templatesPath, fallbackTeamsPath, os.Getenv("MYCELIS_BOOTSTRAP_TEMPLATE_ID"))
	if err != nil {
		return nil, nil, err
	}

	return selection, swarm.NewRegistryFromRuntimeOrganization(selection.Organization), nil
}

type startupProviderRouting struct {
	Policy               swarm.ProviderPolicy
	Source               string
	IgnoredLegacyEnvMaps bool
}

func resolveStartupProviderRouting(selection *bootstrap.StartupSelection, legacyTeamProviderMapJSON, legacyAgentProviderMapJSON string) startupProviderRouting {
	// The env-map provider override path is intentionally retired. Keep the inputs
	// here only long enough to emit migration-focused warnings while runtime truth
	// comes from the instantiated organization policy.
	routing := startupProviderRouting{
		IgnoredLegacyEnvMaps: strings.TrimSpace(legacyTeamProviderMapJSON) != "" || strings.TrimSpace(legacyAgentProviderMapJSON) != "",
	}
	if selection == nil || selection.Organization == nil {
		return routing
	}
	if selection.Organization.ProviderPolicy.IsEmpty() {
		return routing
	}

	routing.Policy = selection.Organization.ProviderPolicy.Clone()
	routing.Source = "runtime_organization"
	return routing
}

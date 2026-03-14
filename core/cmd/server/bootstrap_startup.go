package main

import (
	"os"

	"github.com/mycelis/core/internal/bootstrap"
	"github.com/mycelis/core/internal/swarm"
)

func loadStartupBundleRegistry(templatesPath, fallbackTeamsPath string) (*bootstrap.StartupSelection, *swarm.Registry, error) {
	selection, err := bootstrap.ResolveStartupSelection(templatesPath, fallbackTeamsPath, os.Getenv("MYCELIS_BOOTSTRAP_TEMPLATE_ID"))
	if err != nil {
		return nil, nil, err
	}

	return selection, swarm.NewRegistryFromManifests(selection.Manifests), nil
}

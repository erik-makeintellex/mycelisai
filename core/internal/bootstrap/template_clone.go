package bootstrap

import (
	"fmt"

	"github.com/mycelis/core/internal/swarm"
	"github.com/mycelis/core/pkg/protocol"
)

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

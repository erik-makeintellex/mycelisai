package swarm

import (
	"regexp"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

var nonAlphaNum = regexp.MustCompile(`[^a-z0-9]+`)

// sanitizeID produces a lowercase, hyphen-separated identifier from a name.
func sanitizeID(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = nonAlphaNum.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

// ConvertBlueprintToManifests converts a MissionBlueprint into Soma-compatible
// TeamManifests. Each BlueprintTeam becomes one TeamManifest. Agent Inputs/Outputs
// are aggregated and deduplicated into team-level Inputs/Deliveries.
func ConvertBlueprintToManifests(bp *protocol.MissionBlueprint) []*TeamManifest {
	if bp == nil || len(bp.Teams) == 0 {
		return nil
	}

	manifests := make([]*TeamManifest, 0, len(bp.Teams))

	for _, team := range bp.Teams {
		id := bp.MissionID + "." + sanitizeID(team.Name)

		// Aggregate and deduplicate agent topics
		inputSet := make(map[string]struct{})
		deliverySet := make(map[string]struct{})

		for _, agent := range team.Agents {
			for _, in := range agent.Inputs {
				inputSet[in] = struct{}{}
			}
			for _, out := range agent.Outputs {
				deliverySet[out] = struct{}{}
			}
		}

		inputs := make([]string, 0, len(inputSet))
		for topic := range inputSet {
			inputs = append(inputs, topic)
		}

		deliveries := make([]string, 0, len(deliverySet))
		for topic := range deliverySet {
			deliveries = append(deliveries, topic)
		}

		// Members are already []protocol.AgentManifest — direct copy
		members := make([]protocol.AgentManifest, len(team.Agents))
		copy(members, team.Agents)

		manifests = append(manifests, &TeamManifest{
			ID:          id,
			Name:        team.Name,
			Type:        TeamTypeAction,
			Description: team.Role,
			Members:     members,
			Inputs:      inputs,
			Deliveries:  deliveries,
		})
	}

	return manifests
}

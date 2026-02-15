package swarm

import (
	"fmt"
	"log"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

// ActivationResult reports on teams spawned (or skipped) during blueprint activation.
type ActivationResult struct {
	TeamsSpawned   int      `json:"teams_spawned"`
	TeamsSkipped   int      `json:"teams_skipped"`
	SensorsSpawned int      `json:"sensors_spawned"`
	Errors         []string `json:"errors,omitempty"`
}

// ActivateBlueprint converts a MissionBlueprint into TeamManifests and spawns
// them in the Soma runtime. Sensor agents (role containing "sensor") are spawned
// as SensorAgents instead of cognitive Agents. Idempotent: teams already active
// in Soma are skipped.
func (s *Soma) ActivateBlueprint(bp *protocol.MissionBlueprint, sensorConfigs map[string]SensorConfig) *ActivationResult {
	result := &ActivationResult{}

	if s.nc == nil {
		result.Errors = append(result.Errors, "NATS connection unavailable — teams not activated")
		return result
	}

	manifests := ConvertBlueprintToManifests(bp)
	if len(manifests) == 0 {
		result.Errors = append(result.Errors, "blueprint produced zero team manifests")
		return result
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, manifest := range manifests {
		// Idempotency: skip already-active teams
		if _, exists := s.teams[manifest.ID]; exists {
			result.TeamsSkipped++
			log.Printf("ActivateBlueprint: team %s already active, skipping", manifest.ID)
			continue
		}

		// Build sensor config subset for this team's members
		teamSensorConfigs := make(map[string]SensorConfig)
		for _, member := range manifest.Members {
			if cfg, ok := sensorConfigs[member.ID]; ok {
				teamSensorConfigs[member.ID] = cfg
				result.SensorsSpawned++
			} else if strings.Contains(strings.ToLower(member.Role), "sensor") {
				// Auto-detect: role contains "sensor" but no explicit config provided
				teamSensorConfigs[member.ID] = SensorConfig{
					Type:     SensorTypeHTTP,
					Endpoint: "", // heartbeat-only
				}
				result.SensorsSpawned++
			}
		}

		team := NewTeam(manifest, s.nc, s.brain, s.toolExecutor)
		if len(teamSensorConfigs) > 0 {
			team.SetSensorConfigs(teamSensorConfigs)
		}

		if err := team.Start(); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("team %s: %v", manifest.ID, err))
			continue
		}

		s.teams[manifest.ID] = team
		result.TeamsSpawned++
		log.Printf("ActivateBlueprint: spawned team %s (%d members, %d sensors)",
			manifest.ID, len(manifest.Members), len(teamSensorConfigs))
	}

	return result
}

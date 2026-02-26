package swarm

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/mycelis/core/pkg/protocol"
)

const maxParallelTeamStarts = 4

// ActivationResult reports on teams spawned (or skipped) during blueprint activation.
type ActivationResult struct {
	TeamsSpawned   int      `json:"teams_spawned"`
	TeamsSkipped   int      `json:"teams_skipped"`
	SensorsSpawned int      `json:"sensors_spawned"`
	RunID          string   `json:"run_id,omitempty"` // V7: the mission_run id for this activation
	Errors         []string `json:"errors,omitempty"`
}

// ActivateBlueprint converts a MissionBlueprint into TeamManifests and spawns
// them in the Soma runtime. Sensor agents (role containing "sensor") are spawned
// as SensorAgents instead of cognitive Agents. Idempotent: teams already active
// in Soma are skipped.
//
// V7 Event Spine: if runsManager is wired, creates a mission_run before spawning teams.
// If eventEmitter is wired, passes it to all spawned teams so agents can emit tool events.
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

	// V7: Create mission_run record before spawning teams.
	// runID is empty if runsManager is not wired (pre-V7 mode — silent degradation).
	var runID string
	if s.runsManager != nil && bp.MissionID != "" {
		if id, err := s.runsManager.CreateRun(context.Background(), bp.MissionID); err != nil {
			log.Printf("ActivateBlueprint: failed to create mission run: %v (continuing without run tracking)", err)
		} else {
			runID = id
			result.RunID = runID
		}
	}

	// V7: Emit mission.started event now that we have a run_id.
	if s.eventEmitter != nil && runID != "" {
		if _, err := s.eventEmitter.Emit(context.Background(), runID,
			protocol.EventMissionStarted, protocol.SeverityInfo,
			"soma", "", map[string]interface{}{
				"mission_id": bp.MissionID,
				"teams":      len(bp.Teams),
			}); err != nil {
			log.Printf("ActivateBlueprint: failed to emit mission.started: %v", err)
		}
	}

	// First pass: idempotency check against currently active teams.
	// Keep lock scope short so team startups can happen in parallel.
	s.mu.RLock()
	active := make(map[string]struct{}, len(s.teams))
	for id := range s.teams {
		active[id] = struct{}{}
	}
	s.mu.RUnlock()

	toStart := make([]*TeamManifest, 0, len(manifests))
	for _, manifest := range manifests {
		if _, exists := active[manifest.ID]; exists {
			result.TeamsSkipped++
			log.Printf("ActivateBlueprint: team %s already active, skipping", manifest.ID)
			continue
		}
		toStart = append(toStart, manifest)
	}

	type startOutcome struct {
		spawned bool
		skipped bool
		sensors int
		err     string
	}

	workers := maxParallelTeamStarts
	if len(toStart) < workers {
		workers = len(toStart)
	}
	if workers == 0 {
		return result
	}

	sem := make(chan struct{}, workers)
	outcomes := make(chan startOutcome, len(toStart))
	var wg sync.WaitGroup

	for _, manifest := range toStart {
		wg.Add(1)
		go func(m *TeamManifest) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			// Build sensor config subset for this team's members.
			teamSensorConfigs := make(map[string]SensorConfig)
			sensorCount := 0
			for _, member := range m.Members {
				if cfg, ok := sensorConfigs[member.ID]; ok {
					teamSensorConfigs[member.ID] = cfg
					sensorCount++
				} else if strings.Contains(strings.ToLower(member.Role), "sensor") {
					// Auto-detect: role contains "sensor" but no explicit config provided
					teamSensorConfigs[member.ID] = SensorConfig{
						Type:     SensorTypeHTTP,
						Endpoint: "", // heartbeat-only
					}
					sensorCount++
				}
			}

			team := NewTeam(m, s.nc, s.brain, s.toolExecutor)
			if s.internalTools != nil {
				team.SetToolDescriptions(s.internalTools.ListDescriptions())
			}
			if len(teamSensorConfigs) > 0 {
				team.SetSensorConfigs(teamSensorConfigs)
			}
			// MCP agent-scoped binding
			if s.compositeExec != nil {
				team.SetMCPBinding(s.compositeExec, s.mcpServerNames, s.mcpToolDescs)
			}
			// V7: wire event emitter + run_id into team so agents can emit tool events.
			if s.eventEmitter != nil && runID != "" {
				team.SetEventEmitter(s.eventEmitter, runID)
			}
			// V7: wire conversation logger into blueprint-activated teams.
			if s.conversationLogger != nil {
				team.SetConversationLogger(s.conversationLogger)
			}

			if err := team.Start(); err != nil {
				outcomes <- startOutcome{err: fmt.Sprintf("team %s: %v", m.ID, err)}
				return
			}

			// Race-safe insertion in case another activation call added this team.
			s.mu.Lock()
			if _, exists := s.teams[m.ID]; exists {
				s.mu.Unlock()
				team.Stop()
				outcomes <- startOutcome{skipped: true}
				log.Printf("ActivateBlueprint: team %s already active during parallel insert, skipping", m.ID)
				return
			}
			s.teams[m.ID] = team
			s.mu.Unlock()

			log.Printf("ActivateBlueprint: spawned team %s (%d members, %d sensors)",
				m.ID, len(m.Members), len(teamSensorConfigs))
			outcomes <- startOutcome{spawned: true, sensors: sensorCount}
		}(manifest)
	}

	wg.Wait()
	close(outcomes)

	for out := range outcomes {
		if out.spawned {
			result.TeamsSpawned++
		}
		if out.skipped {
			result.TeamsSkipped++
		}
		if out.sensors > 0 {
			result.SensorsSpawned += out.sensors
		}
		if out.err != "" {
			result.Errors = append(result.Errors, out.err)
		}
	}

	return result
}

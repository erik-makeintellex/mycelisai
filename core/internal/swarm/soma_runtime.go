package swarm

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
)

const runtimeTeamPersistenceTimeout = 5 * time.Second

// Start brings Soma online, loads standing teams, and subscribes to operator intent.
func (s *Soma) Start() error {
	log.Printf("🧠 Soma [%s] Online. Listening for User Intent...", s.id)
	toolDescs := map[string]string(nil)
	if s.internalTools != nil {
		toolDescs = s.internalTools.ListDescriptions()
	}

	manifests, err := s.registry.LoadManifests()
	if err != nil {
		log.Printf("WARN: Failed to load team manifests: %v", err)
	}
	manifests = s.mergeDurableTeamManifests(manifests)
	for _, m := range manifests {
		team := NewTeam(s.applyProviderPolicy(m), s.nc, s.brain, s.toolExecutor)
		s.configureTeam(team, toolDescs)
		s.teams[m.ID] = team
		if err := team.Start(); err != nil {
			log.Printf("ERR: Failed to start team %s: %v", m.ID, err)
		}
	}

	if _, err = s.nc.Subscribe(protocol.TopicGlobalInputUser, s.handleGlobalInput); err != nil {
		return fmt.Errorf("failed to subscribe to operator input: %w", err)
	}
	if err := s.axon.Start(); err != nil {
		return fmt.Errorf("failed to start Axon: %w", err)
	}
	return nil
}

func (s *Soma) mergeDurableTeamManifests(standing []*TeamManifest) []*TeamManifest {
	merged := make([]*TeamManifest, 0, len(standing))
	seen := make(map[string]struct{}, len(standing))
	for _, manifest := range standing {
		if manifest == nil || strings.TrimSpace(manifest.ID) == "" {
			continue
		}
		if _, exists := seen[manifest.ID]; exists {
			continue
		}
		seen[manifest.ID] = struct{}{}
		merged = append(merged, manifest)
	}
	if s.durableTeamLoader == nil {
		return merged
	}

	restored, err := s.durableTeamLoader.LoadRuntimeTeams(s.ctx)
	if err != nil {
		log.Printf("WARN: Failed to restore durable runtime teams: %v", err)
		return merged
	}
	for _, manifest := range restored {
		if manifest == nil || strings.TrimSpace(manifest.ID) == "" {
			continue
		}
		if _, exists := seen[manifest.ID]; exists {
			log.Printf("Soma kept standing manifest for durable team ID %s", manifest.ID)
			continue
		}
		seen[manifest.ID] = struct{}{}
		merged = append(merged, manifest)
		log.Printf("Soma restoring durable runtime team: %s", manifest.ID)
	}
	return merged
}

func (s *Soma) configureTeam(team *Team, toolDescs map[string]string) {
	team.commandReceipts = s.commandReceipts
	if len(toolDescs) > 0 {
		team.SetToolDescriptions(toolDescs)
	}
	if s.internalTools != nil {
		team.SetInternalTools(s.internalTools)
	}
	if s.compositeExec != nil {
		team.SetMCPBinding(s.compositeExec, s.mcpServerNames, s.mcpToolDescs)
	}
	if s.conversationLogger != nil {
		team.SetConversationLogger(s.conversationLogger)
	}
}

// handleGlobalInput processes operator intent after guard validation. Registered
// service/device traffic is normalized and buffered by the input projection.
func (s *Soma) handleGlobalInput(msg *nats.Msg) {
	if err := s.guard.ValidateIngress(msg.Subject, msg.Data); err != nil {
		log.Printf("🛡️ Soma Shield Blocked Input: %v", err)
		return
	}
	log.Printf("🧠 Soma Received Input on [%s]: %s", msg.Subject, string(msg.Data))
	s.axon.ProcessSignal(msg)
}

// SpawnTeam dynamically creates and starts a new team.
func (s *Soma) SpawnTeam(manifest *TeamManifest) error {
	return s.SpawnTeamContext(s.ctx, manifest)
}

// SpawnTeamContext creates a team within the caller's acknowledgement boundary.
func (s *Soma) SpawnTeamContext(ctx context.Context, manifest *TeamManifest) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("spawn team before acknowledgement: %w", err)
	}
	s.spawnMu.Lock()
	defer s.spawnMu.Unlock()

	s.mu.RLock()
	if _, exists := s.teams[manifest.ID]; exists {
		s.mu.RUnlock()
		return fmt.Errorf("team %s already exists", manifest.ID)
	}
	s.mu.RUnlock()

	effectiveManifest := s.applyProviderPolicy(manifest)
	team := NewTeam(effectiveManifest, s.nc, s.brain, s.toolExecutor)
	if s.internalTools != nil {
		s.configureTeam(team, s.internalTools.ListDescriptions())
	}
	if err := team.Start(); err != nil {
		return err
	}
	if s.durableTeamStore != nil {
		persistCtx, cancel := context.WithTimeout(ctx, runtimeTeamPersistenceTimeout)
		err := s.durableTeamStore.SaveRuntimeTeam(persistCtx, effectiveManifest)
		cancel()
		if err != nil {
			team.Stop()
			return fmt.Errorf("persist team %s before acknowledgement: %w", effectiveManifest.ID, err)
		}
	}
	s.mu.Lock()
	s.teams[effectiveManifest.ID] = team
	s.mu.Unlock()
	log.Printf("Soma Spawned New Team: %s", effectiveManifest.ID)
	return nil
}

// ListTeams returns a snapshot of active teams.
func (s *Soma) ListTeams() []*TeamManifest {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var list []*TeamManifest
	for _, t := range s.teams {
		list = append(list, t.Manifest)
	}
	return list
}

// StopTeam stops and removes one active runtime team.
func (s *Soma) StopTeam(teamID string) bool {
	found, err := s.StopTeamDurably(teamID)
	if err != nil {
		log.Printf("ERR: Failed to durably stop Team %s: %v", strings.TrimSpace(teamID), err)
		return false
	}
	return found
}

// StopTeamDurably removes the persisted manifest before stopping the runtime
// instance so a storage failure cannot make a deleted team reappear on restart.
func (s *Soma) StopTeamDurably(teamID string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	teamID = strings.TrimSpace(teamID)
	team, exists := s.teams[teamID]
	if !exists {
		return false, nil
	}
	if s.durableTeamStore != nil {
		if err := s.durableTeamStore.DeleteRuntimeTeam(s.ctx, teamID); err != nil {
			return true, err
		}
	}
	team.Stop()
	delete(s.teams, teamID)
	log.Printf("Soma stopped Team: %s", teamID)
	return true, nil
}

// DeactivateMission stops and removes all teams belonging to a mission.
func (s *Soma) DeactivateMission(missionID string) int {
	prefix := missionID + "."
	s.mu.RLock()
	teamIDs := make([]string, 0)
	for id := range s.teams {
		if strings.HasPrefix(id, prefix) {
			teamIDs = append(teamIDs, id)
		}
	}
	s.mu.RUnlock()

	stopped := 0
	for _, teamID := range teamIDs {
		found, err := s.StopTeamDurably(teamID)
		if err != nil {
			log.Printf("ERR: Failed to durably deactivate Team %s for mission %s: %v", teamID, missionID, err)
			continue
		}
		if found {
			stopped++
		}
	}
	if stopped > 0 {
		log.Printf("DeactivateMission: stopped %d teams for mission %s", stopped, missionID)
	}
	return stopped
}

// Shutdown stops Soma, all teams, and its Axon.
func (s *Soma) Shutdown() {
	s.mu.RLock()
	for id, t := range s.teams {
		log.Printf("Soma shutting down Team [%s]", id)
		t.Stop()
	}
	s.mu.RUnlock()
	s.cancel()
}

func (s *Soma) applyProviderPolicy(manifest *TeamManifest) *TeamManifest {
	if manifest == nil || s.providerPolicy.IsEmpty() {
		return manifest
	}
	resolved, blocked := s.providerPolicy.ResolveManifest(manifest)
	for _, blockedOverride := range blocked {
		log.Printf("WARN: provider override blocked for team %s: %s", manifest.ID, blockedOverride.String())
	}
	if resolved == nil {
		return manifest
	}
	return resolved
}

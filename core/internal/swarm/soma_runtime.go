package swarm

import (
	"fmt"
	"log"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
)

// Start brings Soma online, loads standing teams, and subscribes to global input.
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
	for _, m := range manifests {
		team := NewTeam(s.applyProviderPolicy(m), s.nc, s.brain, s.toolExecutor)
		s.configureTeam(team, toolDescs)
		s.teams[m.ID] = team
		if err := team.Start(); err != nil {
			log.Printf("ERR: Failed to start team %s: %v", m.ID, err)
		}
	}

	if _, err = s.nc.Subscribe(protocol.TopicGlobalInputWild, s.handleGlobalInput); err != nil {
		return fmt.Errorf("failed to subscribe to global input: %w", err)
	}
	if err := s.axon.Start(); err != nil {
		return fmt.Errorf("failed to start Axon: %w", err)
	}
	return nil
}

func (s *Soma) configureTeam(team *Team, toolDescs map[string]string) {
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

// handleGlobalInput processes raw external signals after guard validation.
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
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.teams[manifest.ID]; exists {
		return fmt.Errorf("team %s already exists", manifest.ID)
	}

	team := NewTeam(s.applyProviderPolicy(manifest), s.nc, s.brain, s.toolExecutor)
	if s.internalTools != nil {
		s.configureTeam(team, s.internalTools.ListDescriptions())
	}
	if err := team.Start(); err != nil {
		return err
	}
	s.teams[manifest.ID] = team
	log.Printf("Soma Spawned New Team: %s", manifest.ID)
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

// DeactivateMission stops and removes all teams belonging to a mission.
func (s *Soma) DeactivateMission(missionID string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	prefix := missionID + "."
	stopped := 0
	for id, team := range s.teams {
		if strings.HasPrefix(id, prefix) {
			team.Stop()
			delete(s.teams, id)
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

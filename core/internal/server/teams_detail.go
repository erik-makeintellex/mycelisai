package server

import (
	"log"
	"net/http"
	"time"

	"github.com/mycelis/core/internal/state"
	"github.com/mycelis/core/pkg/protocol"
)

// TeamDetailAgent is a single agent within a team detail response.
type TeamDetailAgent struct {
	ID            string   `json:"id"`
	Role          string   `json:"role"`
	Status        int      `json:"status"` // 0=offline, 1=idle, 2=busy, 3=error
	LastHeartbeat string   `json:"last_heartbeat"`
	Tools         []string `json:"tools"`
	Model         string   `json:"model"`
	SystemPrompt  string   `json:"system_prompt,omitempty"`
}

// TeamDetailEntry is a single team in the aggregated response.
type TeamDetailEntry struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	Role          string            `json:"role"`
	Type          string            `json:"type"` // "standing" or "mission"
	MissionID     *string           `json:"mission_id"`
	MissionIntent *string           `json:"mission_intent"`
	Inputs        []string          `json:"inputs"`
	Deliveries    []string          `json:"deliveries"`
	Agents        []TeamDetailAgent `json:"agents"`
}

// HandleTeamsDetail returns all teams with nested agents, heartbeat status,
// and mission context in a single aggregated response.
// GET /api/v1/teams/detail
func (s *AdminServer) HandleTeamsDetail(w http.ResponseWriter, r *http.Request) {
	entries := []TeamDetailEntry{}
	if s.Soma == nil {
		respondJSON(w, entries)
		return
	}

	activeAgents := state.GlobalRegistry.GetActiveAgents()
	agentMap := make(map[string]*state.AgentState, len(activeAgents))
	for _, a := range activeAgents {
		agentMap[a.ID] = a
	}

	teamMission := s.buildTeamMissionLookup()
	for _, tm := range s.Soma.ListTeams() {
		entry := TeamDetailEntry{
			ID:         tm.ID,
			Name:       tm.Name,
			Role:       string(tm.Type),
			Type:       "standing",
			Inputs:     nonNilSlice(tm.Inputs),
			Deliveries: nonNilSlice(tm.Deliveries),
			Agents:     []TeamDetailAgent{},
		}

		if info, ok := teamMission[tm.Name]; ok {
			entry.Type = "mission"
			entry.MissionID = &info.missionID
			entry.MissionIntent = &info.intent
		}
		for _, agent := range tm.Members {
			entry.Agents = append(entry.Agents, teamDetailAgent(agent, agentMap[agent.ID]))
		}

		entries = append(entries, entry)
	}

	respondJSON(w, entries)
}

func teamDetailAgent(agent protocol.AgentManifest, state *state.AgentState) TeamDetailAgent {
	detail := TeamDetailAgent{
		ID:           agent.ID,
		Role:         agent.Role,
		Status:       0,
		Tools:        nonNilSlice(agent.Tools),
		Model:        agent.Model,
		SystemPrompt: agent.SystemPrompt,
	}
	if state != nil {
		detail.Status = int(state.Status)
		detail.LastHeartbeat = state.LastHeartbeat.Format(time.RFC3339)
	}
	return detail
}

// teamMissionInfo holds the mission context for a team persisted in the DB.
type teamMissionInfo struct {
	missionID string
	intent    string
}

// buildTeamMissionLookup queries teams+missions for active mission-spawned teams.
func (s *AdminServer) buildTeamMissionLookup() map[string]teamMissionInfo {
	result := make(map[string]teamMissionInfo)

	db := s.getDB()
	if db == nil {
		return result
	}

	rows, err := db.Query(`
		SELECT t.name, m.id, m.directive
		FROM teams t
		JOIN missions m ON m.id = t.mission_id
		WHERE m.status = 'active'
	`)
	if err != nil {
		log.Printf("buildTeamMissionLookup: %v", err)
		return result
	}
	defer rows.Close()

	for rows.Next() {
		var teamName, missionID, directive string
		if err := rows.Scan(&teamName, &missionID, &directive); err != nil {
			continue
		}
		result[teamName] = teamMissionInfo{missionID: missionID, intent: directive}
	}

	return result
}

// nonNilSlice ensures a nil string slice becomes an empty JSON array.
func nonNilSlice(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

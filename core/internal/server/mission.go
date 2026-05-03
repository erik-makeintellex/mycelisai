package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/mycelis/core/internal/swarm"
	"github.com/mycelis/core/pkg/protocol"
)

// Root architect user ID seeded in migration 007
const rootOwnerID = "00000000-0000-0000-0000-000000000000"

// handleListMissions returns all missions with team/agent counts.
// GET /api/v1/missions → Mission[] (flat array)
func (s *AdminServer) handleListMissions(w http.ResponseWriter, r *http.Request) {
	db := s.getDB()
	if db == nil {
		http.Error(w, `{"error":"database not available"}`, http.StatusServiceUnavailable)
		return
	}

	rows, err := db.Query(`
		SELECT
			m.id,
			m.directive,
			COALESCE(m.status, 'active') AS status,
			m.created_at,
			COUNT(DISTINCT t.id) AS teams,
			COUNT(DISTINCT sm.id) AS agents
		FROM missions m
		LEFT JOIN teams t ON t.mission_id = m.id
		LEFT JOIN service_manifests sm ON sm.team_id = t.id
		GROUP BY m.id, m.directive, m.status, m.created_at
		ORDER BY m.created_at DESC
		LIMIT 50
	`)
	if err != nil {
		log.Printf("list missions: %v", err)
		// Return empty array instead of 500 — graceful degradation
		respondJSON(w, []struct{}{})
		return
	}
	defer rows.Close()

	type MissionRow struct {
		ID        string `json:"id"`
		Intent    string `json:"intent"`
		Status    string `json:"status"`
		CreatedAt string `json:"created_at"`
		Teams     int    `json:"teams"`
		Agents    int    `json:"agents"`
	}

	var missions []MissionRow
	for rows.Next() {
		var m MissionRow
		var createdAt time.Time
		if err := rows.Scan(&m.ID, &m.Intent, &m.Status, &createdAt, &m.Teams, &m.Agents); err != nil {
			log.Printf("list missions scan: %v", err)
			continue
		}
		m.CreatedAt = createdAt.Format(time.RFC3339)
		missions = append(missions, m)
	}

	if missions == nil {
		missions = []MissionRow{}
	}

	respondJSON(w, missions)
}

// handleUpdateMissionAgent updates an agent's manifest within a mission.
// PUT /api/v1/missions/{id}/agents/{name}
func (s *AdminServer) handleUpdateMissionAgent(w http.ResponseWriter, r *http.Request) {
	missionID := r.PathValue("id")
	agentName := r.PathValue("name")
	db := s.getDB()
	if db == nil {
		http.Error(w, `{"error":"database not available"}`, http.StatusServiceUnavailable)
		return
	}

	var manifest protocol.AgentManifest
	if err := json.NewDecoder(r.Body).Decode(&manifest); err != nil {
		http.Error(w, `{"error":"invalid JSON body"}`, http.StatusBadRequest)
		return
	}

	manifestJSON, err := json.Marshal(manifest)
	if err != nil {
		http.Error(w, `{"error":"failed to marshal manifest"}`, http.StatusInternalServerError)
		return
	}

	result, err := db.Exec(`
		UPDATE service_manifests SET manifest = $1, updated_at = NOW()
		WHERE name = $2 AND team_id IN (SELECT id FROM teams WHERE mission_id = $3)
	`, manifestJSON, agentName, missionID)
	if err != nil {
		log.Printf("update mission agent: %v", err)
		http.Error(w, fmt.Sprintf(`{"error":"update failed: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		http.Error(w, `{"error":"agent not found in mission"}`, http.StatusNotFound)
		return
	}

	log.Printf("Updated agent %s in mission %s", agentName, missionID)
	respondJSON(w, map[string]string{"status": "updated", "agent": agentName})
}

// handleDeleteMissionAgent removes an agent from a mission.
// DELETE /api/v1/missions/{id}/agents/{name}
func (s *AdminServer) handleDeleteMissionAgent(w http.ResponseWriter, r *http.Request) {
	missionID := r.PathValue("id")
	agentName := r.PathValue("name")
	db := s.getDB()
	if db == nil {
		http.Error(w, `{"error":"database not available"}`, http.StatusServiceUnavailable)
		return
	}

	result, err := db.Exec(`
		DELETE FROM service_manifests
		WHERE name = $1 AND team_id IN (SELECT id FROM teams WHERE mission_id = $2)
	`, agentName, missionID)
	if err != nil {
		log.Printf("delete mission agent: %v", err)
		http.Error(w, fmt.Sprintf(`{"error":"delete failed: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		http.Error(w, `{"error":"agent not found in mission"}`, http.StatusNotFound)
		return
	}

	log.Printf("Deleted agent %s from mission %s", agentName, missionID)
	respondJSON(w, map[string]string{"status": "deleted", "agent": agentName})
}

// handleDeleteMission deletes an entire mission (cascades to teams → service_manifests).
// If the mission is active in Soma, all runtime teams are stopped first.
// DELETE /api/v1/missions/{id}
func (s *AdminServer) handleDeleteMission(w http.ResponseWriter, r *http.Request) {
	missionID := r.PathValue("id")
	db := s.getDB()
	if db == nil {
		http.Error(w, `{"error":"database not available"}`, http.StatusServiceUnavailable)
		return
	}

	// 1. Stop runtime teams in Soma (if active)
	stoppedTeams := 0
	if s.Soma != nil {
		stoppedTeams = s.Soma.DeactivateMission(missionID)
	}

	// 2. Delete from DB (cascades via FK: missions → teams → service_manifests)
	result, err := db.Exec(`DELETE FROM missions WHERE id = $1`, missionID)
	if err != nil {
		log.Printf("delete mission: %v", err)
		http.Error(w, fmt.Sprintf(`{"error":"delete failed: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		http.Error(w, `{"error":"mission not found"}`, http.StatusNotFound)
		return
	}

	log.Printf("Deleted mission %s (stopped %d runtime teams)", missionID, stoppedTeams)
	respondJSON(w, map[string]any{
		"status":        "deleted",
		"mission_id":    missionID,
		"teams_stopped": stoppedTeams,
	})
}

// buildSensorConfigs inspects the blueprint and creates SensorConfig entries
// for agents whose role contains "sensor".
func buildSensorConfigs(bp *protocol.MissionBlueprint) map[string]swarm.SensorConfig {
	configs := make(map[string]swarm.SensorConfig)
	for _, team := range bp.Teams {
		for _, agent := range team.Agents {
			if strings.Contains(strings.ToLower(agent.Role), "sensor") {
				configs[agent.ID] = swarm.SensorConfig{
					Type:     swarm.SensorTypeHTTP,
					Interval: 60 * time.Second,
				}
			}
		}
	}
	return configs
}

// getDB retrieves the shared database connection via the Registry service.
func (s *AdminServer) getDB() *sql.DB {
	if s.DB != nil {
		return s.DB
	}
	if s.Registry != nil {
		return s.Registry.DB
	}
	return nil
}

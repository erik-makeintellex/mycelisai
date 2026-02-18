package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/swarm"
	"github.com/mycelis/core/pkg/protocol"
)

// Root architect user ID seeded in migration 007
const rootOwnerID = "00000000-0000-0000-0000-000000000000"

// CommitResponse is the JSON reply from POST /api/v1/intent/commit.
type CommitResponse struct {
	Status     string                  `json:"status"`
	MissionID  string                  `json:"mission_id"`
	Teams      int                     `json:"teams"`
	Agents     int                     `json:"agents"`
	Activation *swarm.ActivationResult `json:"activation,omitempty"`
}

// handleIntentCommit persists a MissionBlueprint into the Ledger
// (missions, teams, service_manifests) and activates teams in the Soma runtime.
func (s *AdminServer) handleIntentCommit(w http.ResponseWriter, r *http.Request) {
	var bp protocol.MissionBlueprint
	if err := json.NewDecoder(r.Body).Decode(&bp); err != nil {
		http.Error(w, `{"error":"invalid JSON body"}`, http.StatusBadRequest)
		return
	}
	if bp.Intent == "" {
		http.Error(w, `{"error":"intent is required"}`, http.StatusBadRequest)
		return
	}

	sensorConfigs := buildSensorConfigs(&bp)
	s.commitAndActivate(w, &bp, sensorConfigs)
}

// handleSymbioticSeed commits the built-in Symbiotic Sensors blueprint.
// Bypasses the MetaArchitect — no LLM required.
func (s *AdminServer) handleSymbioticSeed(w http.ResponseWriter, r *http.Request) {
	bp := swarm.SymbioticSeedBlueprint()
	configs := swarm.SymbioticSeedSensorConfigs()
	s.commitAndActivate(w, bp, configs)
}

// commitAndActivate persists a MissionBlueprint to the database and then
// activates the teams in the Soma swarm runtime. Shared by handleIntentCommit
// and handleSymbioticSeed.
func (s *AdminServer) commitAndActivate(w http.ResponseWriter, bp *protocol.MissionBlueprint, sensorConfigs map[string]swarm.SensorConfig) {
	db := s.getDB()
	if db == nil {
		http.Error(w, `{"error":"database not available"}`, http.StatusServiceUnavailable)
		return
	}

	// Begin transaction — all or nothing
	tx, err := db.Begin()
	if err != nil {
		log.Printf("intent/commit: tx begin: %v", err)
		http.Error(w, `{"error":"database transaction failed"}`, http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// 1. INSERT mission
	missionID := uuid.New()
	missionName := bp.Intent
	if len(missionName) > 120 {
		missionName = missionName[:120]
	}

	_, err = tx.Exec(
		`INSERT INTO missions (id, owner_id, name, directive, status, activated_at) VALUES ($1, $2, $3, $4, $5, $6)`,
		missionID, rootOwnerID, missionName, bp.Intent, "active", time.Now(),
	)
	if err != nil {
		log.Printf("intent/commit: insert mission: %v", err)
		http.Error(w, fmt.Sprintf(`{"error":"insert mission: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	totalAgents := 0

	// 2. INSERT teams + agents
	for tIdx, team := range bp.Teams {
		teamID := uuid.New()
		teamPath := missionID.String() + "." + teamID.String()

		_, err = tx.Exec(
			`INSERT INTO teams (id, owner_id, name, role, mission_id, path) VALUES ($1, $2, $3, $4, $5, $6)`,
			teamID, rootOwnerID, team.Name, team.Role, missionID, teamPath,
		)
		if err != nil {
			log.Printf("intent/commit: insert team[%d] %s: %v", tIdx, team.Name, err)
			http.Error(w, fmt.Sprintf(`{"error":"insert team: %s"}`, err.Error()), http.StatusInternalServerError)
			return
		}

		// 3. INSERT each agent as a service_manifest
		for aIdx, agent := range team.Agents {
			manifestJSON, err := json.Marshal(agent)
			if err != nil {
				log.Printf("intent/commit: marshal agent[%d][%d]: %v", tIdx, aIdx, err)
				continue
			}

			manifestID := uuid.New()
			_, err = tx.Exec(
				`INSERT INTO service_manifests (id, team_id, name, manifest, status) VALUES ($1, $2, $3, $4, 'active')`,
				manifestID, teamID, agent.ID, manifestJSON,
			)
			if err != nil {
				log.Printf("intent/commit: insert agent %s: %v", agent.ID, err)
				http.Error(w, fmt.Sprintf(`{"error":"insert agent: %s"}`, err.Error()), http.StatusInternalServerError)
				return
			}
			totalAgents++
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		log.Printf("intent/commit: tx commit: %v", err)
		http.Error(w, `{"error":"transaction commit failed"}`, http.StatusInternalServerError)
		return
	}

	log.Printf("Mission committed: %s (%d teams, %d agents)", missionID, len(bp.Teams), totalAgents)

	// === Activation Bridge: Persist → Runtime ===
	var activation *swarm.ActivationResult
	if s.Soma != nil {
		activation = s.Soma.ActivateBlueprint(bp, sensorConfigs)
		log.Printf("Mission activated: %d teams spawned, %d skipped, %d sensors",
			activation.TeamsSpawned, activation.TeamsSkipped, activation.SensorsSpawned)
	} else {
		log.Println("WARN: Soma unavailable — mission persisted but not activated")
	}

	respondJSON(w, CommitResponse{
		Status:     "active",
		MissionID:  missionID.String(),
		Teams:      len(bp.Teams),
		Agents:     totalAgents,
		Activation: activation,
	})
}

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

// ── Phase 9: Mission CRUD (Neural Wiring Edit/Delete) ──────────────

// handleGetMission returns a full mission detail with teams and agents,
// reconstructed as a blueprint-like structure the frontend can load into the canvas.
// GET /api/v1/missions/{id}
func (s *AdminServer) handleGetMission(w http.ResponseWriter, r *http.Request) {
	missionID := r.PathValue("id")
	db := s.getDB()
	if db == nil {
		http.Error(w, `{"error":"database not available"}`, http.StatusServiceUnavailable)
		return
	}

	// 1. Fetch mission metadata
	var mission struct {
		ID        string `json:"id"`
		Intent    string `json:"intent"`
		Status    string `json:"status"`
		CreatedAt string `json:"created_at"`
	}
	var createdAt time.Time
	err := db.QueryRow(
		`SELECT id, directive, COALESCE(status, 'active'), created_at FROM missions WHERE id = $1`,
		missionID,
	).Scan(&mission.ID, &mission.Intent, &mission.Status, &createdAt)
	if err == sql.ErrNoRows {
		http.Error(w, `{"error":"mission not found"}`, http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("get mission: %v", err)
		http.Error(w, `{"error":"database error"}`, http.StatusInternalServerError)
		return
	}
	mission.CreatedAt = createdAt.Format(time.RFC3339)

	// 2. Fetch teams
	type TeamDetail struct {
		ID     string                    `json:"id"`
		Name   string                    `json:"name"`
		Role   string                    `json:"role"`
		Agents []protocol.AgentManifest  `json:"agents"`
	}

	teamRows, err := db.Query(
		`SELECT id, name, COALESCE(role, '') FROM teams WHERE mission_id = $1 ORDER BY name`,
		missionID,
	)
	if err != nil {
		log.Printf("get mission teams: %v", err)
		http.Error(w, `{"error":"database error"}`, http.StatusInternalServerError)
		return
	}
	defer teamRows.Close()

	var teams []TeamDetail
	var teamIDs []string
	teamMap := make(map[string]*TeamDetail)

	for teamRows.Next() {
		var td TeamDetail
		if err := teamRows.Scan(&td.ID, &td.Name, &td.Role); err != nil {
			log.Printf("get mission teams scan: %v", err)
			continue
		}
		td.Agents = []protocol.AgentManifest{}
		teams = append(teams, td)
		teamIDs = append(teamIDs, td.ID)
		teamMap[td.ID] = &teams[len(teams)-1]
	}

	// 3. Fetch agents for all teams
	if len(teamIDs) > 0 {
		// Build IN clause
		placeholders := make([]string, len(teamIDs))
		args := make([]any, len(teamIDs))
		for i, id := range teamIDs {
			placeholders[i] = fmt.Sprintf("$%d", i+1)
			args[i] = id
		}
		agentQuery := fmt.Sprintf(
			`SELECT team_id, manifest FROM service_manifests WHERE team_id IN (%s) ORDER BY name`,
			strings.Join(placeholders, ","),
		)
		agentRows, err := db.Query(agentQuery, args...)
		if err != nil {
			log.Printf("get mission agents: %v", err)
		} else {
			defer agentRows.Close()
			for agentRows.Next() {
				var teamID string
				var manifestJSON []byte
				if err := agentRows.Scan(&teamID, &manifestJSON); err != nil {
					continue
				}
				var agent protocol.AgentManifest
				if err := json.Unmarshal(manifestJSON, &agent); err != nil {
					continue
				}
				if td, ok := teamMap[teamID]; ok {
					td.Agents = append(td.Agents, agent)
				}
			}
		}
	}

	if teams == nil {
		teams = []TeamDetail{}
	}

	respondJSON(w, map[string]any{
		"id":         mission.ID,
		"intent":     mission.Intent,
		"status":     mission.Status,
		"created_at": mission.CreatedAt,
		"teams":      teams,
	})
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
	if s.Registry != nil {
		return s.Registry.DB
	}
	return nil
}

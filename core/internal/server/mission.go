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

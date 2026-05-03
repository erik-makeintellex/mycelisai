package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/mycelis/core/pkg/protocol"
)

// GET /api/v1/missions/{id}
func (s *AdminServer) handleGetMission(w http.ResponseWriter, r *http.Request) {
	missionID := r.PathValue("id")
	db := s.getDB()
	if db == nil {
		http.Error(w, `{"error":"database not available"}`, http.StatusServiceUnavailable)
		return
	}

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

	type TeamDetail struct {
		ID     string                   `json:"id"`
		Name   string                   `json:"name"`
		Role   string                   `json:"role"`
		Agents []protocol.AgentManifest `json:"agents"`
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

	if len(teamIDs) > 0 {
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

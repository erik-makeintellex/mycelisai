package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/swarm"
	"github.com/mycelis/core/pkg/protocol"
)

// CommitResponse is the JSON reply from POST /api/v1/intent/commit.
type CommitResponse struct {
	Status        string                  `json:"status"`
	MissionID     string                  `json:"mission_id"`
	RunID         string                  `json:"run_id,omitempty"` // V7: mission run id
	Teams         int                     `json:"teams"`
	Agents        int                     `json:"agents"`
	Activation    *swarm.ActivationResult `json:"activation,omitempty"`
	IntentProofID string                  `json:"intent_proof_id,omitempty"` // CE-1
	AuditEventID  string                  `json:"audit_event_id,omitempty"`  // CE-1
}

// handleIntentCommit persists a MissionBlueprint into the Ledger
// (missions, teams, service_manifests) and activates teams in the Soma runtime.
// CE-1: Requires a confirm_token. No token = no commit.
func (s *AdminServer) handleIntentCommit(w http.ResponseWriter, r *http.Request) {
	var req protocol.CommitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid JSON body"}`, http.StatusBadRequest)
		return
	}
	if req.Intent == "" {
		http.Error(w, `{"error":"intent is required"}`, http.StatusBadRequest)
		return
	}

	if req.ConfirmToken == "" {
		respondError(w, "confirm_token is required. No token = no commit.", http.StatusForbidden)
		return
	}
	proofID, err := s.validateConfirmToken(req.ConfirmToken)
	if err != nil {
		respondError(w, "Invalid confirm token: "+err.Error(), http.StatusForbidden)
		return
	}

	bp := &req.MissionBlueprint
	s.commitAndActivateWithProof(w, bp, buildSensorConfigs(bp), proofID)
}

// handleSymbioticSeed commits the built-in Symbiotic Sensors blueprint.
func (s *AdminServer) handleSymbioticSeed(w http.ResponseWriter, r *http.Request) {
	bp := swarm.SymbioticSeedBlueprint()
	s.commitAndActivate(w, bp, swarm.SymbioticSeedSensorConfigs())
}

func (s *AdminServer) commitAndActivate(w http.ResponseWriter, bp *protocol.MissionBlueprint, sensorConfigs map[string]swarm.SensorConfig) {
	missionID, totalAgents, ok := s.persistMissionBlueprint(w, bp)
	if !ok {
		return
	}

	activation := s.activateCommittedMission(bp, missionID.String(), sensorConfigs)
	resp := CommitResponse{
		Status:     "active",
		MissionID:  missionID.String(),
		Teams:      len(bp.Teams),
		Agents:     totalAgents,
		Activation: activation,
	}
	if activation != nil {
		resp.RunID = activation.RunID
	}
	respondJSON(w, resp)
}

func (s *AdminServer) commitAndActivateWithProof(w http.ResponseWriter, bp *protocol.MissionBlueprint, sensorConfigs map[string]swarm.SensorConfig, proofID string) {
	missionID, totalAgents, ok := s.persistMissionBlueprint(w, bp)
	if !ok {
		return
	}

	activation := s.activateCommittedMission(bp, missionID.String(), sensorConfigs)
	s.confirmIntentProof(proofID, missionID.String())
	auditEventID, _ := s.createAuditEvent(
		protocol.TemplateChatToProposal, "commit",
		fmt.Sprintf("Mission committed: %s", missionID),
		map[string]any{"mission_id": missionID.String(), "teams": len(bp.Teams), "agents": totalAgents, "proof_id": proofID},
	)

	resp := CommitResponse{
		Status:        "active",
		MissionID:     missionID.String(),
		Teams:         len(bp.Teams),
		Agents:        totalAgents,
		Activation:    activation,
		IntentProofID: proofID,
		AuditEventID:  auditEventID,
	}
	if activation != nil {
		resp.RunID = activation.RunID
	}
	respondJSON(w, resp)
}

func (s *AdminServer) persistMissionBlueprint(w http.ResponseWriter, bp *protocol.MissionBlueprint) (uuid.UUID, int, bool) {
	db := s.getDB()
	if db == nil {
		http.Error(w, `{"error":"database not available"}`, http.StatusServiceUnavailable)
		return uuid.Nil, 0, false
	}

	tx, err := db.Begin()
	if err != nil {
		log.Printf("intent/commit: tx begin: %v", err)
		http.Error(w, `{"error":"database transaction failed"}`, http.StatusInternalServerError)
		return uuid.Nil, 0, false
	}
	defer tx.Rollback()

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
		return uuid.Nil, 0, false
	}

	totalAgents := 0
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
			return uuid.Nil, 0, false
		}

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
				return uuid.Nil, 0, false
			}
			totalAgents++
		}
	}

	if err := tx.Commit(); err != nil {
		log.Printf("intent/commit: tx commit: %v", err)
		http.Error(w, `{"error":"transaction commit failed"}`, http.StatusInternalServerError)
		return uuid.Nil, 0, false
	}

	log.Printf("Mission committed: %s (%d teams, %d agents)", missionID, len(bp.Teams), totalAgents)
	return missionID, totalAgents, true
}

func (s *AdminServer) activateCommittedMission(bp *protocol.MissionBlueprint, missionID string, sensorConfigs map[string]swarm.SensorConfig) *swarm.ActivationResult {
	bp.MissionID = missionID
	if s.Soma == nil {
		log.Println("WARN: Soma unavailable — mission persisted but not activated")
		return nil
	}

	activation := s.Soma.ActivateBlueprint(bp, sensorConfigs)
	log.Printf("Mission activated: %d teams spawned, %d skipped, %d sensors (run_id=%s)",
		activation.TeamsSpawned, activation.TeamsSkipped, activation.SensorsSpawned, activation.RunID)
	return activation
}

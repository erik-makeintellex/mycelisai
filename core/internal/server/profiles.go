package server

// Mission Profiles API — named provider-routing configurations that can be activated
// to change which AI provider handles each agent role, with optional reactive NATS
// subscriptions that let the profile's agents watch other agents' outputs.
//
// Endpoints:
//   GET    /api/v1/mission-profiles
//   POST   /api/v1/mission-profiles
//   PUT    /api/v1/mission-profiles/{id}
//   DELETE /api/v1/mission-profiles/{id}
//   POST   /api/v1/mission-profiles/{id}/activate

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/mycelis/core/internal/reactive"
	"github.com/mycelis/core/pkg/protocol"
)

// ProfileSubscription is an alias for the reactive engine's subscription type.
// Re-exported here so profile handlers don't need to import the package directly.
type ProfileSubscription = reactive.ProfileSubscription

// MissionProfile is the full profile record returned by the API.
type MissionProfile struct {
	ID              string          `json:"id"`
	Name            string          `json:"name"`
	Description     string          `json:"description,omitempty"`
	RoleProviders   json.RawMessage `json:"role_providers"`   // {"architect":"vllm"}
	Subscriptions   json.RawMessage `json:"subscriptions"`    // [{"topic":"...","condition":"..."}]
	ContextStrategy string          `json:"context_strategy"` // "fresh"|"warm"|"snapshot:<uuid>"
	AutoStart       bool            `json:"auto_start"`
	IsActive        bool            `json:"is_active"`
	TenantID        string          `json:"tenant_id"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// missionProfileUpsert is the request body for create/update.
type missionProfileUpsert struct {
	Name            string          `json:"name"`
	Description     string          `json:"description,omitempty"`
	RoleProviders   json.RawMessage `json:"role_providers"`
	Subscriptions   json.RawMessage `json:"subscriptions"`
	ContextStrategy string          `json:"context_strategy"`
	AutoStart       bool            `json:"auto_start"`
}

func (s *AdminServer) dbRequired(w http.ResponseWriter) bool {
	if s.DB == nil {
		respondAPIError(w, "Database unavailable", http.StatusServiceUnavailable)
		return false
	}
	return true
}

// GET /api/v1/mission-profiles
func (s *AdminServer) HandleListMissionProfiles(w http.ResponseWriter, r *http.Request) {
	if !s.dbRequired(w) {
		return
	}

	rows, err := s.DB.QueryContext(r.Context(), `
		SELECT id, name, COALESCE(description,''), role_providers, subscriptions,
		       context_strategy, auto_start, is_active, tenant_id, created_at, updated_at
		FROM mission_profiles
		WHERE tenant_id = 'default'
		ORDER BY created_at ASC`)
	if err != nil {
		log.Printf("HandleListMissionProfiles: %v", err)
		respondAPIError(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	profiles := make([]MissionProfile, 0)
	for rows.Next() {
		var p MissionProfile
		if err := rows.Scan(
			&p.ID, &p.Name, &p.Description,
			&p.RoleProviders, &p.Subscriptions,
			&p.ContextStrategy, &p.AutoStart, &p.IsActive,
			&p.TenantID, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			continue
		}
		profiles = append(profiles, p)
	}

	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(profiles))
}

// POST /api/v1/mission-profiles
func (s *AdminServer) HandleCreateMissionProfile(w http.ResponseWriter, r *http.Request) {
	if !s.dbRequired(w) {
		return
	}

	var req missionProfileUpsert
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondAPIError(w, "Bad JSON", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		respondAPIError(w, "name is required", http.StatusBadRequest)
		return
	}
	if len(req.RoleProviders) == 0 {
		req.RoleProviders = json.RawMessage("{}")
	}
	if len(req.Subscriptions) == 0 {
		req.Subscriptions = json.RawMessage("[]")
	}
	if req.ContextStrategy == "" {
		req.ContextStrategy = "fresh"
	}

	var p MissionProfile
	err := s.DB.QueryRowContext(r.Context(), `
		INSERT INTO mission_profiles
		    (name, description, role_providers, subscriptions, context_strategy, auto_start, tenant_id)
		VALUES ($1, NULLIF($2,''), $3, $4, $5, $6, 'default')
		RETURNING id, name, COALESCE(description,''), role_providers, subscriptions,
		          context_strategy, auto_start, is_active, tenant_id, created_at, updated_at`,
		req.Name, req.Description,
		[]byte(req.RoleProviders), []byte(req.Subscriptions),
		req.ContextStrategy, req.AutoStart,
	).Scan(
		&p.ID, &p.Name, &p.Description,
		&p.RoleProviders, &p.Subscriptions,
		&p.ContextStrategy, &p.AutoStart, &p.IsActive,
		&p.TenantID, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		log.Printf("HandleCreateMissionProfile insert: %v", err)
		respondAPIError(w, "Failed to create profile: "+err.Error(), http.StatusInternalServerError)
		return
	}

	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(p))
}

// PUT /api/v1/mission-profiles/{id}
func (s *AdminServer) HandleUpdateMissionProfile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondAPIError(w, "Missing profile ID", http.StatusBadRequest)
		return
	}
	if !s.dbRequired(w) {
		return
	}

	var req missionProfileUpsert
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondAPIError(w, "Bad JSON", http.StatusBadRequest)
		return
	}
	if len(req.RoleProviders) == 0 {
		req.RoleProviders = json.RawMessage("{}")
	}
	if len(req.Subscriptions) == 0 {
		req.Subscriptions = json.RawMessage("[]")
	}
	if req.ContextStrategy == "" {
		req.ContextStrategy = "fresh"
	}

	res, err := s.DB.ExecContext(r.Context(), `
		UPDATE mission_profiles
		SET name=$1, description=NULLIF($2,''), role_providers=$3, subscriptions=$4,
		    context_strategy=$5, auto_start=$6, updated_at=NOW()
		WHERE id=$7 AND tenant_id='default'`,
		req.Name, req.Description,
		[]byte(req.RoleProviders), []byte(req.Subscriptions),
		req.ContextStrategy, req.AutoStart, id,
	)
	if err != nil {
		log.Printf("HandleUpdateMissionProfile exec: %v", err)
		respondAPIError(w, "Failed to update profile: "+err.Error(), http.StatusInternalServerError)
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		respondAPIError(w, "Profile not found", http.StatusNotFound)
		return
	}

	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(map[string]any{"id": id, "updated": true}))
}

// DELETE /api/v1/mission-profiles/{id}
func (s *AdminServer) HandleDeleteMissionProfile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondAPIError(w, "Missing profile ID", http.StatusBadRequest)
		return
	}
	if !s.dbRequired(w) {
		return
	}

	// Unsubscribe reactive engine before deleting
	if s.Reactive != nil {
		s.Reactive.Unsubscribe(id)
	}

	res, err := s.DB.ExecContext(r.Context(),
		"DELETE FROM mission_profiles WHERE id=$1 AND tenant_id='default'", id)
	if err != nil {
		log.Printf("HandleDeleteMissionProfile exec: %v", err)
		respondAPIError(w, "Failed to delete profile: "+err.Error(), http.StatusInternalServerError)
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		respondAPIError(w, "Profile not found", http.StatusNotFound)
		return
	}

	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(map[string]any{"id": id, "deleted": true}))
}

// POST /api/v1/mission-profiles/{id}/activate
// Applies the profile's role_providers to the running cognitive router,
// registers reactive NATS subscriptions, and marks it active in the DB.
func (s *AdminServer) HandleActivateMissionProfile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondAPIError(w, "Missing profile ID", http.StatusBadRequest)
		return
	}
	if !s.dbRequired(w) {
		return
	}

	// Load the profile
	var p MissionProfile
	var desc sql.NullString
	err := s.DB.QueryRowContext(r.Context(), `
		SELECT id, name, COALESCE(description,''), role_providers, subscriptions,
		       context_strategy, auto_start, is_active, tenant_id, created_at, updated_at
		FROM mission_profiles WHERE id=$1 AND tenant_id='default'`, id).
		Scan(&p.ID, &p.Name, &desc,
			&p.RoleProviders, &p.Subscriptions,
			&p.ContextStrategy, &p.AutoStart, &p.IsActive,
			&p.TenantID, &p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		respondAPIError(w, "Profile not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("HandleActivateMissionProfile load: %v", err)
		respondAPIError(w, "Database error", http.StatusInternalServerError)
		return
	}
	p.Description = desc.String

	// Apply role_providers to the cognitive router's Profiles map
	if s.Cognitive != nil && s.Cognitive.Config != nil {
		var roleProviders map[string]string
		if err := json.Unmarshal(p.RoleProviders, &roleProviders); err == nil {
			for role, providerID := range roleProviders {
				s.Cognitive.Config.Profiles[role] = providerID
			}
			if err := s.Cognitive.SaveConfig(); err != nil {
				log.Printf("HandleActivateMissionProfile SaveConfig: %v", err)
				// Non-fatal — profile still activates, but config won't survive restart
			}
		}
	}

	// Register reactive subscriptions
	if s.Reactive != nil {
		var subs []ProfileSubscription
		if err := json.Unmarshal(p.Subscriptions, &subs); err == nil && len(subs) > 0 {
			if err := s.Reactive.Subscribe(id, subs); err != nil {
				log.Printf("HandleActivateMissionProfile Subscribe: %v", err)
				// Non-fatal — profile still activates
			}
		}
	}

	// Mark this profile active (clear others unless auto_start)
	tx, err := s.DB.BeginTx(r.Context(), nil)
	if err != nil {
		respondAPIError(w, "Database error", http.StatusInternalServerError)
		return
	}
	// Deactivate all non-auto-start profiles first
	if _, err := tx.ExecContext(r.Context(),
		"UPDATE mission_profiles SET is_active=false WHERE tenant_id='default' AND auto_start=false AND id != $1", id); err != nil {
		tx.Rollback()
		respondAPIError(w, "Database error", http.StatusInternalServerError)
		return
	}
	// Activate this one
	if _, err := tx.ExecContext(r.Context(),
		"UPDATE mission_profiles SET is_active=true, updated_at=NOW() WHERE id=$1", id); err != nil {
		tx.Rollback()
		respondAPIError(w, "Database error", http.StatusInternalServerError)
		return
	}
	if err := tx.Commit(); err != nil {
		respondAPIError(w, "Database error", http.StatusInternalServerError)
		return
	}

	p.IsActive = true
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(p))
}

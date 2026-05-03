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

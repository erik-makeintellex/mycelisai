package server

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/mycelis/core/pkg/protocol"
)

// HandleActivateMissionProfile applies providers, subscriptions, and active DB state.
func (s *AdminServer) HandleActivateMissionProfile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondAPIError(w, "Missing profile ID", http.StatusBadRequest)
		return
	}
	if !s.dbRequired(w) {
		return
	}

	p, err := s.loadMissionProfileForActivation(r, id)
	if err == sql.ErrNoRows {
		respondAPIError(w, "Profile not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("HandleActivateMissionProfile load: %v", err)
		respondAPIError(w, "Database error", http.StatusInternalServerError)
		return
	}

	s.applyMissionProfileProviders(p)
	s.applyMissionProfileSubscriptions(id, p)
	if err := s.markMissionProfileActive(r, id); err != nil {
		respondAPIError(w, "Database error", http.StatusInternalServerError)
		return
	}

	p.IsActive = true
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(p))
}

func (s *AdminServer) loadMissionProfileForActivation(r *http.Request, id string) (MissionProfile, error) {
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
	p.Description = desc.String
	return p, err
}

func (s *AdminServer) applyMissionProfileProviders(p MissionProfile) {
	if s.Cognitive == nil || s.Cognitive.Config == nil {
		return
	}
	var roleProviders map[string]string
	if err := json.Unmarshal(p.RoleProviders, &roleProviders); err != nil {
		return
	}
	for role, providerID := range roleProviders {
		s.Cognitive.Config.Profiles[role] = providerID
	}
	if err := s.Cognitive.SaveConfig(); err != nil {
		log.Printf("HandleActivateMissionProfile SaveConfig: %v", err)
	}
}

func (s *AdminServer) applyMissionProfileSubscriptions(id string, p MissionProfile) {
	if s.Reactive == nil {
		return
	}
	var subs []ProfileSubscription
	if err := json.Unmarshal(p.Subscriptions, &subs); err == nil && len(subs) > 0 {
		if err := s.Reactive.Subscribe(id, subs); err != nil {
			log.Printf("HandleActivateMissionProfile Subscribe: %v", err)
		}
	}
}

func (s *AdminServer) markMissionProfileActive(r *http.Request, id string) error {
	tx, err := s.DB.BeginTx(r.Context(), nil)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(r.Context(),
		"UPDATE mission_profiles SET is_active=false WHERE tenant_id='default' AND auto_start=false AND id != $1", id); err != nil {
		_ = tx.Rollback()
		return err
	}
	if _, err := tx.ExecContext(r.Context(),
		"UPDATE mission_profiles SET is_active=true, updated_at=NOW() WHERE id=$1", id); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

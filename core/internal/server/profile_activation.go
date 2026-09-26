package server

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/pkg/protocol"
)

// missionProfileRoutingNotApplied reports an activation whose is_active flag
// committed but whose runtime routing apply failed afterward.
const missionProfileRoutingNotApplied = "mission_profile_routing_not_applied"

// missionProfileActivationRecovery is the operator recovery for
// missionProfileRoutingNotApplied.
const missionProfileActivationRecovery = "The profile is marked active but routing did not change. " +
	"Fix or re-enable the providers it names, then activate it again."

// HandleActivateMissionProfile applies providers, subscriptions, and active DB state.
// Root admin + cognitive:write, audited first, and all or nothing for routing:
// every role->provider pair is validated before anything changes, and the
// pairs are applied together as runtime-only overrides.
func (s *AdminServer) HandleActivateMissionProfile(w http.ResponseWriter, r *http.Request) {
	if !requireRoutingWriter(w, r) {
		return
	}
	id := r.PathValue("id")
	if id == "" {
		respondAPIError(w, "Missing profile ID", http.StatusBadRequest)
		return
	}
	if !s.dbRequired(w) {
		return
	}

	// Serialized with profile override PUT/DELETE and provider toggles so the
	// validation below still holds when the overrides are applied.
	profileOverrideWriteMu.Lock()
	defer profileOverrideWriteMu.Unlock()

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

	overrides, ok := missionProfileOverrides(p)
	if !ok {
		respondAPIError(w, "Profile role_providers is not a role-to-provider map; nothing applied", http.StatusBadRequest)
		return
	}
	if len(overrides) > 0 {
		if err := s.Cognitive.ValidateProfileOverrides(overrides); err != nil {
			respondProfileOverrideRejection(w, err)
			return
		}
	}
	if _, ok := s.auditRoutingMutation(w, r, auditMissionProfileActive, "Mission profile activated", map[string]any{
		"mission_profile_id": id, "mission_profile_name": p.Name, "role_providers": overrides,
	}); !ok {
		return
	}

	if err := s.markMissionProfileActive(r, id); err != nil {
		respondAPIError(w, "Database error", http.StatusInternalServerError)
		return
	}
	if len(overrides) > 0 {
		// Runtime-only overrides, re-validated and applied under one router
		// write lock; never written to YAML or system_config.
		if err := s.Cognitive.SetProfileOverrides(overrides, cognitive.ProfileOriginRuntime); err != nil {
			log.Printf("HandleActivateMissionProfile: profile %s marked active but routing not applied: %v", id, err)
			s.recordRoutingMutationFailure(r, auditMissionProfileActive, "Mission profile activation routing not applied", map[string]any{
				"mission_profile_id": id, "mission_profile_name": p.Name, "role_providers": overrides,
				"code": missionProfileRoutingNotApplied, "error": err.Error(), "recovery_action": missionProfileActivationRecovery,
			})
			respondAPIJSON(w, http.StatusInternalServerError, protocol.APIResponse{
				OK:    false,
				Error: "mission profile marked active but its routing was not applied: " + err.Error(),
				Data: map[string]any{"id": id, "code": missionProfileRoutingNotApplied,
					"recommended_action": missionProfileActivationRecovery},
			})
			return
		}
	}
	s.applyMissionProfileSubscriptions(id, p)

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

// missionProfileOverrides decodes the stored role->provider map. An empty
// or null map is valid (no routing change); anything else that does not
// decode is rejected so activation never applies part of a profile.
func missionProfileOverrides(p MissionProfile) (map[string]string, bool) {
	overrides := map[string]string{}
	if len(p.RoleProviders) == 0 || string(p.RoleProviders) == "null" {
		return overrides, true
	}
	if err := json.Unmarshal(p.RoleProviders, &overrides); err != nil {
		return nil, false
	}
	return overrides, true
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

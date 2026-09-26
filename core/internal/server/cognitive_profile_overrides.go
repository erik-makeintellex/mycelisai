package server

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sort"
	"strings"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/pkg/protocol"
)

// cognitiveWriteScope gates every mutation of cognitive routing: profile
// overrides (PUT/DELETE) and provider config (PUT).
const cognitiveWriteScope = "cognitive:write"

const cognitiveProfileAuditSource = "cognitive-profile-override"

type profileOverrideEffective struct {
	ProviderID string `json:"provider_id,omitempty"`
	ModelID    string `json:"model_id,omitempty"`
	Source     string `json:"source"`
	Available  bool   `json:"available"`
	Code       string `json:"code"`
}

type profileOverrideResult struct {
	Profile                 string                   `json:"profile"`
	Changed                 bool                     `json:"changed"`
	Effective               profileOverrideEffective `json:"effective"`
	RemainingOverrideOrigin string                   `json:"remaining_override_origin,omitempty"`
	RecommendedAction       string                   `json:"recommended_action,omitempty"`
}

type profileOverrideRejection struct {
	Profile string `json:"profile"`
	Code    string `json:"code"`
	Error   string `json:"error"`
}

// PUT /api/v1/cognitive/profiles
// Pins profiles to providers as durable operator overrides in
// system_config role.<profile>. Root admin + cognitive:write, audited, all
// or nothing, and only for a provider that can run. Never writes YAML.
func (s *AdminServer) HandleUpdateProfiles(w http.ResponseWriter, r *http.Request) {
	_, ok := requireRootAdminScope(w, r, cognitiveWriteScope)
	if !ok {
		return
	}
	if s.Cognitive == nil || s.Cognitive.Config == nil {
		respondAPIError(w, "Cognitive Matrix Offline", http.StatusServiceUnavailable)
		return
	}
	var req struct {
		Profiles map[string]string `json:"profiles"`
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		respondAPIError(w, "Invalid profiles request", http.StatusBadRequest)
		return
	}
	if len(req.Profiles) == 0 {
		respondAPIError(w, "No profiles provided", http.StatusBadRequest)
		return
	}
	overrides := make(map[string]string, len(req.Profiles))
	for profile, providerID := range req.Profiles {
		overrides[profile] = strings.TrimSpace(providerID)
	}
	if err := s.Cognitive.ValidateProfileOverrides(overrides); err != nil {
		respondProfileOverrideRejection(w, err)
		return
	}
	db := s.getDB()
	if db == nil {
		respondAPIError(w, "Database unavailable: profile overrides are stored in system_config", http.StatusServiceUnavailable)
		return
	}
	names := sortedProfileNames(overrides)
	previous := make(map[string]any, len(names))
	for _, profile := range names {
		state := s.Cognitive.ProfileOverrideState(profile)
		previous[profile] = map[string]any{"provider": state.ProviderID, "origin": state.Origin, "source": state.Source}
	}
	auditID, err := s.createAuditEvent(protocol.TemplateChatToProposal, cognitiveProfileAuditSource, "Cognitive profile override set", attachActorIdentity(map[string]any{
		"actor":         "operator",
		"user":          auditUserLabelFromRequest(r),
		"action":        "cognitive_profile_override_set",
		"profiles":      overrides,
		"previous":      previous,
		"result_status": "requested",
	}, r))
	if err != nil || strings.TrimSpace(auditID) == "" {
		respondAPIError(w, "Audit unavailable: profile override not changed", http.StatusServiceUnavailable)
		return
	}
	tx, err := db.BeginTx(r.Context(), nil)
	if err != nil {
		respondAPIError(w, "Database unavailable: profile override not changed", http.StatusServiceUnavailable)
		return
	}
	for _, profile := range names {
		if _, err := tx.ExecContext(r.Context(),
			`INSERT INTO system_config (key, value, updated_at) VALUES ($1, $2, NOW())
			 ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()`,
			"role."+profile, overrides[profile]); err != nil {
			_ = tx.Rollback()
			log.Printf("HandleUpdateProfiles: role.%s upsert failed: %v", profile, err)
			respondAPIError(w, "Database write failed: profile override not changed", http.StatusServiceUnavailable)
			return
		}
	}
	if err := tx.Commit(); err != nil {
		respondAPIError(w, "Database write failed: profile override not changed", http.StatusServiceUnavailable)
		return
	}
	if err := s.Cognitive.SetProfileOverrides(overrides, cognitive.ProfileOriginDB); err != nil {
		// The provider changed between validation and commit. The row is
		// durable and status reports the profile as unavailable.
		log.Printf("HandleUpdateProfiles: committed role.* rows but runtime apply failed: %v", err)
		respondProfileOverrideRejection(w, err)
		return
	}
	results := make([]profileOverrideResult, 0, len(names))
	for _, profile := range names {
		results = append(results, s.profileOverrideResult(profile, true, ""))
	}
	log.Printf("Cognitive profile overrides set in system_config: %v (audit %s)", names, auditID)
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(map[string]any{"profiles": results, "audit_event_id": auditID}))
}

// DELETE /api/v1/cognitive/profiles/{profile}/override
// Removes the durable role.<profile> row and any runtime override, then
// recomputes the binding exactly as a restart would. Idempotent and audited.
func (s *AdminServer) HandleClearProfileOverride(w http.ResponseWriter, r *http.Request) {
	_, ok := requireRootAdminScope(w, r, cognitiveWriteScope)
	if !ok {
		return
	}
	profile := r.PathValue("profile")
	if s.Cognitive == nil || s.Cognitive.Config == nil {
		respondAPIError(w, "Cognitive Matrix Offline", http.StatusServiceUnavailable)
		return
	}
	if !s.Cognitive.KnownProfile(profile) {
		respondAPIError(w, "Unknown cognitive profile", http.StatusNotFound)
		return
	}
	db := s.getDB()
	if db == nil {
		respondAPIError(w, "Database unavailable: profile overrides are stored in system_config", http.StatusServiceUnavailable)
		return
	}
	state := s.Cognitive.ProfileOverrideState(profile)
	status := "noop"
	if state.DBRowPresent || (state.Origin != "" && state.Origin != cognitive.ProfileOriginEnv) {
		status = "requested"
	}
	auditID, err := s.createAuditEvent(protocol.TemplateChatToProposal, cognitiveProfileAuditSource, "Cognitive profile override cleared", attachActorIdentity(map[string]any{
		"actor":             "operator",
		"user":              auditUserLabelFromRequest(r),
		"action":            "cognitive_profile_override_cleared",
		"profile":           profile,
		"previous_provider": state.ProviderID,
		"previous_origin":   state.Origin,
		"result_status":     status,
	}, r))
	if err != nil || strings.TrimSpace(auditID) == "" {
		respondAPIError(w, "Audit unavailable: profile override not changed", http.StatusServiceUnavailable)
		return
	}
	tx, err := db.BeginTx(r.Context(), nil)
	if err != nil {
		respondAPIError(w, "Database unavailable: profile override not changed", http.StatusServiceUnavailable)
		return
	}
	res, err := tx.ExecContext(r.Context(), `DELETE FROM system_config WHERE key = 'role.' || $1`, profile)
	if err != nil {
		_ = tx.Rollback()
		log.Printf("HandleClearProfileOverride: role.%s delete failed: %v", profile, err)
		respondAPIError(w, "Database write failed: profile override not changed", http.StatusServiceUnavailable)
		return
	}
	if err := tx.Commit(); err != nil {
		respondAPIError(w, "Database write failed: profile override not changed", http.StatusServiceUnavailable)
		return
	}
	rowDeleted := false
	if affected, err := res.RowsAffected(); err == nil && affected > 0 {
		rowDeleted = true
	}
	cleared := s.Cognitive.ClearProfileOverride(profile)
	result := s.profileOverrideResult(profile, cleared.Changed || rowDeleted, cleared.RemainingOverrideOrigin)
	log.Printf("Cognitive profile override cleared: %s changed=%v effective=%s (audit %s)", profile, result.Changed, result.Effective.ProviderID, auditID)
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(result))
}

func (s *AdminServer) profileOverrideResult(profile string, changed bool, remainingOrigin string) profileOverrideResult {
	state := s.Cognitive.ProfileOverrideState(profile)
	availability := s.Cognitive.ExecutionAvailability(profile, "")
	result := profileOverrideResult{
		Profile: profile,
		Changed: changed,
		Effective: profileOverrideEffective{
			ProviderID: state.ProviderID,
			ModelID:    availability.ModelID,
			Source:     state.Source,
			Available:  availability.Available,
			Code:       availability.Code,
		},
		RemainingOverrideOrigin: remainingOrigin,
	}
	if remainingOrigin == cognitive.ProfileOriginEnv {
		result.RecommendedAction = "The profile is still pinned by MYCELIS_PROFILE_" + strings.ToUpper(profile) + "_PROVIDER; edit .env.compose and recreate Core to return it to root."
	} else if !availability.Available {
		result.RecommendedAction = availability.RecommendedAction
	}
	return result
}

func respondProfileOverrideRejection(w http.ResponseWriter, err error) {
	var rejection *cognitive.ProfileOverrideError
	if !errors.As(err, &rejection) {
		respondAPIError(w, "Profile override rejected", http.StatusBadRequest)
		return
	}
	status := http.StatusBadRequest
	switch rejection.Code {
	case cognitive.OverrideEnvPinned:
		status = http.StatusConflict
	case cognitive.ExecutionRouterUnavailable:
		status = http.StatusServiceUnavailable
	}
	respondAPIJSON(w, status, protocol.APIResponse{
		OK:    false,
		Error: rejection.Message,
		Data:  profileOverrideRejection{Profile: rejection.Profile, Code: rejection.Code, Error: rejection.Message},
	})
}

func sortedProfileNames(overrides map[string]string) []string {
	names := make([]string, 0, len(overrides))
	for profile := range overrides {
		names = append(names, profile)
	}
	sort.Strings(names)
	return names
}

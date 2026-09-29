package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/mycelis/core/internal/swarm"
)

const defaultAssistantName = "Soma"

// User represents the logged-in user
type User struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	Role          string `json:"role"`
	EffectiveRole string `json:"effective_role,omitempty"`
	PrincipalType string `json:"principal_type,omitempty"`
	AuthSource    string `json:"auth_source,omitempty"`
	BreakGlass    bool   `json:"break_glass,omitempty"`
	// Scopes and IsApprover (MCPL) are interface hints only; Core re-checks
	// authority on every request.
	Scopes     []string        `json:"scopes"`
	IsApprover bool            `json:"is_approver"`
	Settings   json.RawMessage `json:"settings"`
	// SettingsStatus is set only when the saved settings can't be read; the
	// approval policy then fails strict (AUTH-C1b).
	SettingsStatus *userSettingsStatus `json:"settings_status,omitempty"`
	CreatedAt      time.Time           `json:"created_at"`
}

type userSettingsStatus struct {
	Code           string `json:"code"`            // settings_store_unavailable
	ApprovalPolicy string `json:"approval_policy"` // fail_strict
	Detail         string `json:"detail,omitempty"`
}

// Team represents a team context
type Team struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"` // User's role in this team
}

// HandleMe returns the current authenticated user from context identity.
func (s *AdminServer) HandleMe(w http.ResponseWriter, r *http.Request) {
	identity := IdentityFromContext(r.Context())
	if identity == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"not authenticated"}`))
		return
	}

	settings, settingsErr := loadPersistedUserSettingsWithStatus()
	user := User{
		ID:            identity.UserID,
		Username:      identity.Username,
		Role:          identity.Role,
		EffectiveRole: identity.EffectiveRole,
		PrincipalType: identity.PrincipalType,
		AuthSource:    identity.AuthSource,
		BreakGlass:    identity.BreakGlass,
		Scopes:        append([]string{}, identity.Scopes...),
		IsApprover:    isApprover(identity),
		Settings:      mustJSON(ResolveDeploymentContract().ApplyUserSettings(settings)),
		CreatedAt:     time.Now(),
	}
	if settingsErr != nil {
		user.SettingsStatus = &userSettingsStatus{Code: codeSettingsStoreUnavailable, ApprovalPolicy: "fail_strict"}
		if identity.Role == "admin" {
			user.SettingsStatus.Detail = settingsErr.Error()
		}
	}
	respondJSON(w, user)
}

// HandleTeams lists active team ids and names (GET) or spawns a runtime team
// from a raw manifest (POST, root admin + groups:write; see spawnRuntimeTeam).
func (s *AdminServer) HandleTeams(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		s.spawnRuntimeTeam(w, r)
		return
	}

	// Real Data from Soma
	if s.Soma != nil {
		manifests := s.Soma.ListTeams()
		var teams []Team
		for _, m := range manifests {
			teams = append(teams, Team{
				ID:   m.ID,
				Name: m.Name,
				Role: "observer", // Default role for now
			})
		}
		respondJSON(w, teams)
		return
	}

	http.Error(w, "Soma Unavailable", http.StatusServiceUnavailable)
}

// HandleDeleteTeam stops and durably removes one runtime team (AUTH-C1b):
// root admin + groups:write (the runtime team spawn scope), refused before
// anything else; a Core-owned team is refused for everyone (403, the attempt
// is audited); otherwise audit (requested) first, stop, audit (result). An
// audit failure stops nothing.
func (s *AdminServer) HandleDeleteTeam(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireRootAdminScope(w, r, scopeRuntimeTeamSpawn); !ok {
		return
	}
	if s.Soma == nil {
		respondBlocker(w, r, http.StatusServiceUnavailable, codeTeamServiceOffline, "", nil)
		return
	}
	teamID := strings.TrimSpace(r.PathValue("id"))
	if teamID == "" {
		respondAPIError(w, "team id is required", http.StatusBadRequest)
		return
	}
	auditCtx := func(status string) map[string]any {
		return map[string]any{"action": "runtime_team_delete", "team_id": teamID, "route": r.URL.Path, "result_status": status}
	}
	if swarm.IsReservedTeamID(teamID) || s.Soma.IsCoreOwnedTeam(teamID) {
		s.auditGovernance(r, "runtime-team", "Core-owned team delete refused", auditCtx("refused_core_owned"))
		respondBlockerText(w, r, http.StatusForbidden, codeCoreTeamProtected, coreTeamProtectedCopy, "", map[string]string{"team_id": teamID})
		return
	}
	if s.auditGovernance(r, "runtime-team", "Runtime team delete requested", auditCtx("requested")) == "" {
		respondBlocker(w, r, http.StatusServiceUnavailable, codeServiceUnavailable,
			"The audit record could not be written, so the team was not stopped.", nil)
		return
	}
	found, err := s.Soma.StopTeamDurably(teamID)
	switch {
	case errors.Is(err, swarm.ErrCoreOwnedTeam):
		s.auditGovernance(r, "runtime-team", "Runtime team delete finished", auditCtx("refused_core_owned"))
		respondBlockerText(w, r, http.StatusForbidden, codeCoreTeamProtected, coreTeamProtectedCopy, "", map[string]string{"team_id": teamID})
	case errors.Is(err, swarm.ErrRuntimeTeamOwnershipActive):
		s.auditGovernance(r, "runtime-team", "Runtime team delete finished", auditCtx("refused_ownership_active"))
		respondAPIError(w, "Runtime team has active operator ownership; revoke it before deletion", http.StatusConflict)
	case err != nil:
		s.auditGovernance(r, "runtime-team", "Runtime team delete finished", auditCtx("failed"))
		respondAPIError(w, "failed to remove durable team state", http.StatusInternalServerError)
	case !found:
		s.auditGovernance(r, "runtime-team", "Runtime team delete finished", auditCtx("not_found"))
		respondAPIError(w, "team not found", http.StatusNotFound)
	default:
		s.auditGovernance(r, "runtime-team", "Runtime team delete finished", auditCtx("stopped"))
		respondJSON(w, map[string]any{"status": "stopped", "team_id": teamID})
	}
}

// HandleUserSettings is the canonical settings contract.
// The frontend loads persisted settings from GET and updates them through PUT.
func (s *AdminServer) HandleUserSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		settings, err := readPersistedUserSettings()
		if errors.Is(err, errNoUserSettingsPath) {
			settings, err = defaultPersistedUserSettings(), nil // nothing saved anywhere: the defaults are the truth
		}
		if err != nil {
			respondSettingsStoreUnavailable(w, r, err)
			return
		}
		respondJSON(w, ResolveDeploymentContract().ApplyUserSettings(settings))
	case http.MethodPut:
		s.putUserSettings(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// HandleUpdateSettings is retained as a compatibility wrapper for older callers.
func (s *AdminServer) HandleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	s.HandleUserSettings(w, r)
}

func mustJSON(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		return json.RawMessage(`{}`)
	}
	return json.RawMessage(b)
}

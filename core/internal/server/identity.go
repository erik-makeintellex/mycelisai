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
	CreatedAt  time.Time       `json:"created_at"`
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
		Settings:      mustJSON(loadUserSettings()),
		CreatedAt:     time.Now(),
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

// HandleDeleteTeam stops and removes one active runtime team.
func (s *AdminServer) HandleDeleteTeam(w http.ResponseWriter, r *http.Request) {
	if s.Soma == nil {
		http.Error(w, "Soma Unavailable", http.StatusServiceUnavailable)
		return
	}
	teamID := strings.TrimSpace(r.PathValue("id"))
	if teamID == "" {
		http.Error(w, "team id is required", http.StatusBadRequest)
		return
	}
	found, err := s.Soma.StopTeamDurably(teamID)
	if err != nil {
		if errors.Is(err, swarm.ErrRuntimeTeamOwnershipActive) {
			respondAPIError(w, "Runtime team has active operator ownership; revoke it before deletion", http.StatusConflict)
			return
		}
		http.Error(w, "failed to remove durable team state", http.StatusInternalServerError)
		return
	}
	if !found {
		http.Error(w, "team not found", http.StatusNotFound)
		return
	}
	respondJSON(w, map[string]any{
		"status":  "stopped",
		"team_id": teamID,
	})
}

// HandleUserSettings is the canonical settings contract.
// The frontend loads persisted settings from GET and updates them through PUT.
func (s *AdminServer) HandleUserSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		respondJSON(w, loadUserSettings())
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

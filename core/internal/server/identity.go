package server

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

const defaultAssistantName = "Soma"

// User represents the logged-in user
type User struct {
	ID            string          `json:"id"`
	Username      string          `json:"username"`
	Role          string          `json:"role"`
	EffectiveRole string          `json:"effective_role,omitempty"`
	PrincipalType string          `json:"principal_type,omitempty"`
	AuthSource    string          `json:"auth_source,omitempty"`
	BreakGlass    bool            `json:"break_glass,omitempty"`
	Settings      json.RawMessage `json:"settings"`
	CreatedAt     time.Time       `json:"created_at"`
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
		Settings:      mustJSON(loadUserSettings()),
		CreatedAt:     time.Now(),
	}
	respondJSON(w, user)
}

// HandleTeams returns the list of teams for the user or creates a new one
func (s *AdminServer) HandleTeams(w http.ResponseWriter, r *http.Request) {
	log.Printf("DEBUG: HandleTeams called. Method: %s. Soma: %v", r.Method, s.Soma)
	if r.Method == "POST" {
		if s.Soma != nil {
			s.Soma.HandleCreateTeam(w, r)
			return
		}
		http.Error(w, "Soma not initialized", http.StatusServiceUnavailable)
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

// HandleUserSettings is the canonical settings contract.
// The frontend loads persisted settings from GET and updates them through PUT.
func (s *AdminServer) HandleUserSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		respondJSON(w, loadUserSettings())
	case http.MethodPut:
		var input map[string]any
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}

		settings := mergeUserSettings(input)
		if err := saveUserSettings(settings); err != nil {
			http.Error(w, "failed to persist user settings", http.StatusInternalServerError)
			return
		}

		respondJSON(w, settings)
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

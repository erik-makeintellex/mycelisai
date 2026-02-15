package server

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// User represents the logged-in user
type User struct {
	ID        string          `json:"id"`
	Username  string          `json:"username"`
	Role      string          `json:"role"`
	Settings  json.RawMessage `json:"settings"`
	CreatedAt time.Time       `json:"created_at"`
}

// Team represents a team context
type Team struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"` // User's role in this team
}

// HandleMe returns the current authenticated user
func (s *AdminServer) HandleMe(w http.ResponseWriter, r *http.Request) {
	// Mock: Assume Admin for dev
	user := User{
		ID:        "user-001",
		Username:  "admin",
		Role:      "admin",
		Settings:  json.RawMessage(`{"theme": "aero-light", "matrix_view": "grid"}`),
		CreatedAt: time.Now(),
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

// HandleUpdateSettings updates user preferences
func (s *AdminServer) HandleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	// Check Auth...
	// Parse Body...
	// Update DB...
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "updated"}`))
}

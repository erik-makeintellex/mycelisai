package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mycelis/core/internal/state"
)

const defaultAssistantName = "Soma"

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

func defaultUserSettings() map[string]any {
	return map[string]any{
		"theme":                 "aero-light",
		"matrix_view":           "grid",
		"assistant_name":        defaultAssistantName,
		"role":                  "owner",
		"cost_sensitivity":      "balanced",
		"review_strictness":     "standard",
		"automation_tolerance":  "balanced",
		"escalation_preference": "ask",
	}
}

func normalizeAssistantName(v any) string {
	name, ok := v.(string)
	if !ok {
		return defaultAssistantName
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return defaultAssistantName
	}
	runes := []rune(name)
	if len(runes) > 48 {
		name = string(runes[:48])
	}
	return name
}

func userSettingsPath() string {
	if p := strings.TrimSpace(os.Getenv("MYCELIS_USER_SETTINGS_PATH")); p != "" {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		return ""
	}
	return filepath.Join(home, ".mycelis", "user-settings.json")
}

func loadUserSettings() map[string]any {
	settings := defaultUserSettings()
	path := userSettingsPath()
	if path == "" {
		return settings
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return settings
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return settings
	}

	if theme, ok := raw["theme"].(string); ok && strings.TrimSpace(theme) != "" {
		settings["theme"] = strings.TrimSpace(theme)
	}
	if matrixView, ok := raw["matrix_view"].(string); ok && strings.TrimSpace(matrixView) != "" {
		settings["matrix_view"] = strings.TrimSpace(matrixView)
	}
	settings["assistant_name"] = normalizeAssistantName(raw["assistant_name"])
	if role, ok := raw["role"]; ok {
		if normalized := normalizeGovernanceRole(fmt.Sprint(role)); normalized != "" {
			settings["role"] = normalized
		}
	}
	settings["cost_sensitivity"] = normalizeCostSensitivity(raw["cost_sensitivity"])
	settings["review_strictness"] = normalizeReviewStrictness(raw["review_strictness"])
	settings["automation_tolerance"] = normalizeAutomationTolerance(raw["automation_tolerance"])
	settings["escalation_preference"] = normalizeEscalationPreference(raw["escalation_preference"])
	return settings
}

func saveUserSettings(settings map[string]any) error {
	path := userSettingsPath()
	if path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	payload, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, payload, 0o644)
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
		ID:        identity.UserID,
		Username:  identity.Username,
		Role:      identity.Role,
		Settings:  mustJSON(loadUserSettings()),
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
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input map[string]any
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	settings := loadUserSettings()
	if theme, ok := input["theme"].(string); ok && strings.TrimSpace(theme) != "" {
		settings["theme"] = strings.TrimSpace(theme)
	}
	if matrixView, ok := input["matrix_view"].(string); ok && strings.TrimSpace(matrixView) != "" {
		settings["matrix_view"] = strings.TrimSpace(matrixView)
	}
	if _, hasAssistantName := input["assistant_name"]; hasAssistantName {
		settings["assistant_name"] = normalizeAssistantName(input["assistant_name"])
	}
	if _, hasRole := input["role"]; hasRole {
		if normalized := normalizeGovernanceRole(fmt.Sprint(input["role"])); normalized != "" {
			settings["role"] = normalized
		}
	}
	if _, hasCostSensitivity := input["cost_sensitivity"]; hasCostSensitivity {
		settings["cost_sensitivity"] = normalizeCostSensitivity(input["cost_sensitivity"])
	}
	if _, hasReviewStrictness := input["review_strictness"]; hasReviewStrictness {
		settings["review_strictness"] = normalizeReviewStrictness(input["review_strictness"])
	}
	if _, hasAutomationTolerance := input["automation_tolerance"]; hasAutomationTolerance {
		settings["automation_tolerance"] = normalizeAutomationTolerance(input["automation_tolerance"])
	}
	if _, hasEscalationPreference := input["escalation_preference"]; hasEscalationPreference {
		settings["escalation_preference"] = normalizeEscalationPreference(input["escalation_preference"])
	}

	if err := saveUserSettings(settings); err != nil {
		http.Error(w, "failed to persist user settings", http.StatusInternalServerError)
		return
	}

	respondJSON(w, map[string]any{
		"status":   "updated",
		"settings": settings,
	})
}

func mustJSON(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		return json.RawMessage(`{}`)
	}
	return json.RawMessage(b)
}

// ── Phase 11: Team Detail (aggregated endpoint) ──────────────

// TeamDetailAgent is a single agent within a team detail response.
type TeamDetailAgent struct {
	ID            string   `json:"id"`
	Role          string   `json:"role"`
	Status        int      `json:"status"` // 0=offline, 1=idle, 2=busy, 3=error
	LastHeartbeat string   `json:"last_heartbeat"`
	Tools         []string `json:"tools"`
	Model         string   `json:"model"`
	SystemPrompt  string   `json:"system_prompt,omitempty"`
}

// TeamDetailEntry is a single team in the aggregated response.
type TeamDetailEntry struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	Role          string            `json:"role"`
	Type          string            `json:"type"` // "standing" or "mission"
	MissionID     *string           `json:"mission_id"`
	MissionIntent *string           `json:"mission_intent"`
	Inputs        []string          `json:"inputs"`
	Deliveries    []string          `json:"deliveries"`
	Agents        []TeamDetailAgent `json:"agents"`
}

// HandleTeamsDetail returns all teams with nested agents, heartbeat status,
// and mission context in a single aggregated response.
// GET /api/v1/teams/detail
func (s *AdminServer) HandleTeamsDetail(w http.ResponseWriter, r *http.Request) {
	var entries []TeamDetailEntry

	if s.Soma != nil {
		manifests := s.Soma.ListTeams()

		// Build agent heartbeat lookup
		activeAgents := state.GlobalRegistry.GetActiveAgents()
		agentMap := make(map[string]*state.AgentState, len(activeAgents))
		for _, a := range activeAgents {
			agentMap[a.ID] = a
		}

		// Build team→mission lookup from DB (teams table has mission_id FK)
		teamMission := s.buildTeamMissionLookup()

		for _, tm := range manifests {
			entry := TeamDetailEntry{
				ID:         tm.ID,
				Name:       tm.Name,
				Role:       string(tm.Type),
				Type:       "standing",
				Inputs:     nonNilSlice(tm.Inputs),
				Deliveries: nonNilSlice(tm.Deliveries),
				Agents:     []TeamDetailAgent{},
			}

			// Check DB for mission association
			if info, ok := teamMission[tm.Name]; ok {
				entry.Type = "mission"
				entry.MissionID = &info.missionID
				entry.MissionIntent = &info.intent
			}

			// Build agent list with heartbeat status
			for _, agent := range tm.Members {
				da := TeamDetailAgent{
					ID:           agent.ID,
					Role:         agent.Role,
					Status:       0, // offline
					Tools:        nonNilSlice(agent.Tools),
					Model:        agent.Model,
					SystemPrompt: agent.SystemPrompt,
				}

				if as, ok := agentMap[agent.ID]; ok {
					da.Status = int(as.Status)
					da.LastHeartbeat = as.LastHeartbeat.Format(time.RFC3339)
				}

				entry.Agents = append(entry.Agents, da)
			}

			entries = append(entries, entry)
		}
	}

	if entries == nil {
		entries = []TeamDetailEntry{}
	}

	respondJSON(w, entries)
}

// teamMissionInfo holds the mission context for a team persisted in the DB.
type teamMissionInfo struct {
	missionID string
	intent    string
}

// buildTeamMissionLookup queries the teams+missions tables to find which
// runtime teams are mission-spawned. Returns a map of team_name → mission info.
func (s *AdminServer) buildTeamMissionLookup() map[string]teamMissionInfo {
	result := make(map[string]teamMissionInfo)

	db := s.getDB()
	if db == nil {
		return result
	}

	rows, err := db.Query(`
		SELECT t.name, m.id, m.directive
		FROM teams t
		JOIN missions m ON m.id = t.mission_id
		WHERE m.status = 'active'
	`)
	if err != nil {
		log.Printf("buildTeamMissionLookup: %v", err)
		return result
	}
	defer rows.Close()

	for rows.Next() {
		var teamName, missionID, directive string
		if err := rows.Scan(&teamName, &missionID, &directive); err != nil {
			continue
		}
		result[teamName] = teamMissionInfo{missionID: missionID, intent: directive}
	}

	return result
}

// nonNilSlice ensures a nil string slice becomes an empty JSON array.
func nonNilSlice(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

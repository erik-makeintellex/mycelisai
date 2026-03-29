package swarm

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/mycelis/core/pkg/protocol"
)

// HandleCreateTeam processes POST and GET requests for swarm teams.
func (s *Soma) HandleCreateTeam(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.HandleListTeams(w, r)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var manifest TeamManifest
	if err := json.NewDecoder(r.Body).Decode(&manifest); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if manifest.ID == "" || manifest.Name == "" {
		http.Error(w, "Missing ID or Name", http.StatusBadRequest)
		return
	}
	if manifest.Type == "" {
		manifest.Type = TeamTypeAction
	}
	if err := s.SpawnTeam(&manifest); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "spawned", "id": manifest.ID})
}

func (s *Soma) HandleListTeams(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.ListTeams())
}

// HandleCommand injects a user command into the swarm global input bus.
func (s *Soma) HandleCommand(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var payload struct {
		Content string `json:"content"`
		Source  string `json:"source"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if payload.Content == "" {
		http.Error(w, "Missing content", http.StatusBadRequest)
		return
	}

	subject := protocol.TopicGlobalInputUser
	if payload.Source != "" {
		subject = fmt.Sprintf(protocol.TopicGlobalInputFmt, payload.Source)
	}
	if err := s.nc.Publish(subject, []byte(payload.Content)); err != nil {
		log.Printf("Failed to publish command: %v", err)
		http.Error(w, "Failed to inject command", http.StatusInternalServerError)
		return
	}
	s.nc.Flush()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "sent", "subject": subject})
}

type BroadcastReply struct {
	TeamID  string `json:"team_id"`
	Content string `json:"content"`
	Error   string `json:"error,omitempty"`
}

// HandleBroadcast fans a directive out to all active teams via request-reply.
func (s *Soma) HandleBroadcast(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var payload struct {
		Content string `json:"content"`
		Source  string `json:"source"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if payload.Content == "" {
		http.Error(w, "Missing content", http.StatusBadRequest)
		return
	}
	if payload.Source == "" {
		payload.Source = "mission-control"
	}

	s.mu.RLock()
	teamIDs := make([]string, 0, len(s.teams))
	for id := range s.teams {
		teamIDs = append(teamIDs, id)
	}
	s.mu.RUnlock()

	var wg sync.WaitGroup
	replies := make([]BroadcastReply, len(teamIDs))
	for i, id := range teamIDs {
		wg.Add(1)
		go func(idx int, teamID string) {
			defer wg.Done()
			subject := fmt.Sprintf(protocol.TopicTeamInternalTrigger, teamID)
			log.Printf("📡 Broadcast request to team [%s] on [%s]", teamID, subject)
			msg, err := s.nc.Request(subject, []byte(payload.Content), 60*time.Second)
			if err != nil {
				log.Printf("Broadcast: team [%s] did not respond: %v", teamID, err)
				replies[idx] = BroadcastReply{TeamID: teamID, Error: err.Error()}
				return
			}
			replies[idx] = BroadcastReply{TeamID: teamID, Content: string(msg.Data)}
		}(i, id)
	}
	wg.Wait()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{"status": "broadcast", "teams_hit": len(teamIDs), "source": payload.Source, "replies": replies})
}

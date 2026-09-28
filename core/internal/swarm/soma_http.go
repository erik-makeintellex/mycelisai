package swarm

import (
	"encoding/json"
	"errors"
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
	for _, member := range manifest.Members {
		if member.ProfileRef != "" || member.Profile != nil {
			http.Error(w, "Worker Profile selection must be resolved through Soma", http.StatusBadRequest)
			return
		}
	}
	if manifest.Type == "" {
		manifest.Type = TeamTypeAction
	}
	existing, err := s.spawnTeam(s.ctx, &manifest)
	if err != nil {
		log.Printf("ERR: spawn runtime team %s: %v", manifest.ID, err)
		switch {
		case errors.Is(err, ErrRuntimeTeamManifestConflict):
			http.Error(w, "Runtime team already runs a different manifest", http.StatusConflict)
		case errors.Is(err, ErrRuntimeTeamInvalid):
			http.Error(w, "Invalid runtime team manifest", http.StatusBadRequest)
		default:
			http.Error(w, "Runtime team unavailable", http.StatusServiceUnavailable)
		}
		return
	}
	status, code := "spawned", http.StatusCreated
	if existing {
		status, code = "already_exists", http.StatusOK
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"status": status, "id": manifest.ID})
}

func (s *Soma) HandleListTeams(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.ListTeams())
}

type BroadcastReply struct {
	TeamID  string `json:"team_id"`
	Content string `json:"content"`
	Error   string `json:"error,omitempty"`
}

// HandleBroadcast fans a directive out to all active teams via request-reply.
// Authority is checked by the Core route (server.HandleSwarmBroadcast, root
// admin); the content always reaches agents as broadcast text (F16b).
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
	data, err := broadcastTriggerPayload(protocol.SourceKindWebAPI, "api.swarm.broadcast", payload.Content)
	if err != nil {
		http.Error(w, "Invalid content", http.StatusBadRequest)
		return
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
			msg, err := s.nc.Request(subject, data, 60*time.Second)
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

// broadcastTriggerPayload wraps broadcast text for team internal.trigger
// (F16b): a signal envelope whose content is only the text field, so the bytes
// can never parse as a TeamAsk or carry posture or correlation fields.
func broadcastTriggerPayload(sourceKind protocol.SignalSourceKind, sourceChannel, content string) ([]byte, error) {
	return json.Marshal(protocol.SignalEnvelope{
		Meta: protocol.SignalMeta{Timestamp: time.Now().UTC(), SourceKind: sourceKind,
			SourceChannel: sourceChannel, PayloadKind: protocol.PayloadKindEvent},
		Text: content,
	})
}

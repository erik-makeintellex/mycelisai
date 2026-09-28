package server

import "net/http"

// HandleSwarmBroadcast is POST /api/v1/swarm/broadcast (F16b): root admin +
// swarm:broadcast before anything is read or sent, then Soma fans the content
// out as broadcast text (swarm.Soma.HandleBroadcast wraps it; it never
// reaches agents as a TeamAsk).
func (s *AdminServer) HandleSwarmBroadcast(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireRootAdminScope(w, r, scopeSwarmBroadcast); !ok {
		return
	}
	if s.Soma == nil {
		respondBlocker(w, r, http.StatusServiceUnavailable, codeTeamServiceOffline, "Soma is not running; nothing was broadcast", nil)
		return
	}
	s.Soma.HandleBroadcast(w, r)
}

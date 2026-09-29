package server

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

// Runtime team spawn authority (AUTH-C1 A1). A raw manifest spawn (members,
// tools, prompts, shared paths) bypasses Soma's governed create_team proposal,
// so it is a root-admin operation under the scope that creates groups (work
// lanes) today: groups:write. The full-manifest list exposes prompts and tool
// refs, so it follows the groups read rule: groups:read. Denials happen before
// the body is read, before audit and before any spawn.
const (
	scopeRuntimeTeamRead    = "groups:read"
	scopeRuntimeTeamSpawn   = "groups:write"
	maxRuntimeTeamBodyBytes = 1 << 20
)

// HandleSwarmTeams serves /api/swarm/teams. Soma no longer mounts its own
// handler there; this is the only route and it carries the authority checks.
func (s *AdminServer) HandleSwarmTeams(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if _, ok := requireRootAdminScope(w, r, scopeRuntimeTeamRead); !ok {
			return
		}
		if s.Soma == nil {
			respondBlocker(w, r, http.StatusServiceUnavailable, codeTeamServiceOffline, "", nil)
			return
		}
		s.Soma.HandleListTeams(w, r)
	case http.MethodPost:
		s.spawnRuntimeTeam(w, r)
	default:
		respondAPIError(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// spawnRuntimeTeam is the fail-closed raw spawn: authority, runtime present,
// bounded body, audit (requested), Soma spawn, audit (result). An audit
// failure refuses before anything starts.
func (s *AdminServer) spawnRuntimeTeam(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireRootAdminScope(w, r, scopeRuntimeTeamSpawn); !ok {
		return
	}
	if s.Soma == nil {
		respondBlocker(w, r, http.StatusServiceUnavailable, codeTeamServiceOffline, "", nil)
		return
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxRuntimeTeamBodyBytes))
	if err != nil {
		respondAPIError(w, "team manifest body is unreadable or exceeds 1 MiB", http.StatusBadRequest)
		return
	}
	var head struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(body, &head) // Soma validates the full manifest below.
	auditCtx := func(status string, code int) map[string]any {
		ctx := map[string]any{"action": "runtime_team_spawn", "team_id": head.ID, "route": r.URL.Path, "result_status": status}
		if code != 0 {
			ctx["http_status"] = code
		}
		return ctx
	}
	if s.auditGovernance(r, "runtime-team", "Runtime team spawn requested", auditCtx("requested", 0)) == "" {
		respondBlocker(w, r, http.StatusServiceUnavailable, codeServiceUnavailable,
			"The audit record could not be written, so the team was not spawned.", nil)
		return
	}
	r.Body = io.NopCloser(bytes.NewReader(body))
	rec := &runtimeTeamStatusWriter{ResponseWriter: w, status: http.StatusOK}
	s.Soma.HandleCreateTeam(rec, r)
	result := "accepted" // 201 spawned or 200 already running the same manifest
	if rec.status >= http.StatusBadRequest {
		result = "refused"
	}
	s.auditGovernance(r, "runtime-team", "Runtime team spawn finished", auditCtx(result, rec.status))
}

// runtimeTeamStatusWriter records the status Soma wrote so the result audit
// says what happened.
type runtimeTeamStatusWriter struct {
	http.ResponseWriter
	status int
}

func (w *runtimeTeamStatusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

package server

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// GET /api/v1/memory/sitreps?team_id=<uuid>&limit=10 (MEM-LANES-2, C3).
// A SitRep summarizes one team's activity, so the route needs an identity
// and returns only the SitReps of teams the caller is a proven member of
// in the SitRep's tenant (the B1R persisted membership proof through
// memoryTeamMember; MEM-LANES-3: a member of the same team id in another
// tenant proves nothing), or of every
// team for a root admin with groups:read (scopeRuntimeTeamRead, the scope
// that already reads every runtime team). Without team_id it lists every
// team the caller may read; a team_id the caller may not read is refused
// with the blocker envelope and nothing is shown.
const (
	codeSitrepsSignInRequired = "memory_sign_in_required"
	codeSitrepsTeamNotMember  = "memory_team_not_member"
	maxSitrepLimit            = 100
	// sitrepTenant is the tenant SitReps belong to: Core's Archivist writes
	// them in tenant "default" (memory/archivist.go), and the table has no
	// tenant column.
	sitrepTenant = "default"
)

var (
	sitrepsSignInCopy = roleBlockerText{User: blockerText{"Sign in to see team situation reports.",
		"Sign in, then open the reports again. Nothing was shown."}}
	sitrepsNotMemberCopy = roleBlockerText{User: blockerText{"You can see situation reports only for teams you belong to.",
		"Pick one of your teams, or ask its owner to add you. Nothing was shown."}}
)

func (s *AdminServer) HandleListSitReps(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.Mem == nil {
		http.Error(w, `{"error":"Memory service offline"}`, http.StatusServiceUnavailable)
		return
	}
	identity := IdentityFromContext(r.Context())
	if identity == nil {
		respondBlockerText(w, r, http.StatusUnauthorized, codeSitrepsSignInRequired, sitrepsSignInCopy, "", nil)
		return
	}
	teamID := strings.TrimSpace(r.URL.Query().Get("team_id"))
	if teamID != "" {
		if _, err := uuid.Parse(teamID); err != nil {
			respondAPIError(w, "team_id must be a team UUID", http.StatusBadRequest)
			return
		}
	}
	limit := 10
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 && parsed <= maxSitrepLimit {
			limit = parsed
		}
	}

	allTeams := identity.Role == "admin" && hasScope(identity, scopeRuntimeTeamRead)
	var sitreps []map[string]any
	var err error
	switch {
	case allTeams && teamID == "":
		sitreps, err = s.Mem.ListAllSitReps(r.Context(), limit)
	case allTeams:
		sitreps, err = s.Mem.ListSitReps(r.Context(), teamID, limit)
	default:
		teams, ok := s.sitrepReadableTeams(w, r, teamID)
		if !ok {
			return
		}
		sitreps, err = s.Mem.ListSitRepsForTeams(r.Context(), teams, limit)
	}
	if err != nil {
		http.Error(w, `{"error":"failed to retrieve sitreps"}`, http.StatusInternalServerError)
		return
	}
	if sitreps == nil {
		sitreps = []map[string]any{}
	}
	respondJSON(w, map[string]any{
		"team_id": teamID,
		"sitreps": sitreps,
		"count":   len(sitreps),
	})
}

// sitrepReadableTeams is the teams a caller without the root-admin read may
// see: teamID when they are a proven member of it (else a 403 blocker), or
// every team with SitReps they are a proven member of. A failed membership
// check is a 503 blocker, never a wider or silently shorter list.
func (s *AdminServer) sitrepReadableTeams(w http.ResponseWriter, r *http.Request, teamID string) ([]string, bool) {
	candidates := []string{teamID}
	if teamID == "" {
		ids, err := s.Mem.SitRepTeamIDs(r.Context())
		if err != nil {
			log.Printf("sitreps: team list failed: %v", err)
			respondBlocker(w, r, http.StatusServiceUnavailable, codeServiceUnavailable, "SitRep teams could not be listed", nil)
			return nil, false
		}
		candidates = ids
	}
	var teams []string
	for _, candidate := range candidates {
		member, err := s.memoryTeamMember(r, sitrepTenant, candidate)
		if err != nil {
			log.Printf("sitreps: membership check failed for team %s: %v", candidate, err)
			respondBlocker(w, r, http.StatusServiceUnavailable, codeServiceUnavailable, "team membership could not be checked", nil)
			return nil, false
		}
		if member {
			teams = append(teams, candidate)
		}
	}
	if teamID != "" && len(teams) == 0 {
		respondBlockerText(w, r, http.StatusForbidden, codeSitrepsTeamNotMember, sitrepsNotMemberCopy, "", map[string]string{"team_id": teamID})
		return nil, false
	}
	return teams, true
}

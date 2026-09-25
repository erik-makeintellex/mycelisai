package swarm

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
)

// ErrRuntimeTeamManifestConflict reports a same-id spawn whose effective
// manifest differs from the team already running under that id.
var ErrRuntimeTeamManifestConflict = errors.New("runtime team already runs a different manifest")

// ErrRuntimeTeamInvalid reports a spawn request that fails validation.
var ErrRuntimeTeamInvalid = errors.New("invalid runtime team manifest")

// ErrRuntimeTeamUnavailable reports a runtime or persistence failure while
// spawning; callers must not expose the wrapped detail to clients.
var ErrRuntimeTeamUnavailable = errors.New("runtime team unavailable")

// DurableTeamRestoreDegradation records one team that could not be restored at
// boot. The source row is never deleted; operators repair or remove it.
type DurableTeamRestoreDegradation struct {
	TeamID string `json:"team_id"`
	Reason string `json:"reason"`
}

// DurableTeamRestoreReporter is implemented by loaders that isolate bad rows
// and report them instead of aborting the whole restoration.
type DurableTeamRestoreReporter interface {
	LoadRuntimeTeamsWithReport(context.Context) ([]*TeamManifest, []DurableTeamRestoreDegradation, error)
}

// approvedRuntimeTeamManifest returns the canonical pre-Start copy that is
// persisted and digested. Start later mutates the running manifest (provider
// routing), so idempotency must never compare against the running form.
func approvedRuntimeTeamManifest(manifest *TeamManifest) (*TeamManifest, string, error) {
	raw, err := json.Marshal(manifest)
	if err != nil {
		return nil, "", err
	}
	var approved TeamManifest
	if err := json.Unmarshal(raw, &approved); err != nil {
		return nil, "", err
	}
	digest, err := runtimeTeamManifestDigest(&approved)
	return &approved, digest, err
}

func runtimeTeamManifestDigest(manifest *TeamManifest) (string, error) {
	raw, err := json.Marshal(manifest)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("sha256:%x", sha256.Sum256(raw)), nil
}

// HasTeam reports whether a team with this id is currently running.
func (s *Soma) HasTeam(teamID string) bool {
	teamID = strings.TrimSpace(teamID)
	if s == nil || teamID == "" {
		return false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.teams[teamID]
	return exists
}

// RestorationDegradations returns the teams that failed to restore at boot.
func (s *Soma) RestorationDegradations() []DurableTeamRestoreDegradation {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]DurableTeamRestoreDegradation(nil), s.restoreDegradations...)
}

func (s *Soma) recordRestorationDegradation(teamID, reason string) {
	log.Printf("WARN: runtime team restoration degraded for team %q: %s", teamID, reason)
	s.mu.Lock()
	s.restoreDegradations = append(s.restoreDegradations, DurableTeamRestoreDegradation{TeamID: teamID, Reason: reason})
	s.mu.Unlock()
}

// sameRunningManifest compares the requested effective manifest digest with
// the running team's spawn-time digest.
func sameRunningManifest(existing *Team, requestedDigest string) bool {
	current := existing.spawnDigest
	if current == "" {
		existing.mu.Lock()
		current, _ = runtimeTeamManifestDigest(existing.Manifest)
		existing.mu.Unlock()
	}
	return current != "" && current == requestedDigest
}

// RestorationHealth summarizes boot restoration for Core status: "online"
// when every team restored, otherwise "degraded" with a count and team ids.
// Failure reasons stay in logs and RestorationDegradations, not in the detail.
func (s *Soma) RestorationHealth() (string, string) {
	degraded := s.RestorationDegradations()
	if len(degraded) == 0 {
		return "online", "All runtime teams restored"
	}
	ids := make([]string, 0, len(degraded))
	for _, item := range degraded {
		ids = append(ids, firstNonEmptyRuntimeString(item.TeamID, "(unidentified row)"))
	}
	return "degraded", fmt.Sprintf("%d runtime team(s) failed restoration: %s", len(degraded), strings.Join(ids, ", "))
}

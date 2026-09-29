package swarm

import (
	"errors"
	"strings"
)

// CoreOwnedTeamIDs are the team IDs reserved for Core's standing teams. Only
// the boot registry may register them (and marks them Core-owned); create_team,
// SpawnTeam, and durable restore refuse them whether or not the team is loaded.
var CoreOwnedTeamIDs = []string{"admin-core", "council-core", "genesis-core", "telemetry-core"}

// ErrReservedTeamID reports a runtime attempt to register a Core-owned team ID.
var ErrReservedTeamID = errors.New("team id is reserved for a Core-owned team")

// ErrCoreOwnedTeam reports an attempt to stop a loaded Core-owned team
// through StopTeamDurably; only Soma's own Shutdown stops those teams.
var ErrCoreOwnedTeam = errors.New("team is Core-owned and cannot be stopped at runtime")

// IsReservedTeamID matches trimmed and case-insensitively, since some callers
// compare Core team IDs with EqualFold.
func IsReservedTeamID(id string) bool {
	id = strings.TrimSpace(id)
	for _, reserved := range CoreOwnedTeamIDs {
		if strings.EqualFold(id, reserved) {
			return true
		}
	}
	return false
}

// IsCoreOwnedTeam reports whether teamID is loaded and was registered by the
// standing boot registry as a Core-owned team. An ID string alone never is.
func (s *Soma) IsCoreOwnedTeam(teamID string) bool {
	if s == nil {
		return false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	team := s.teams[strings.TrimSpace(teamID)]
	return team != nil && team.coreOwned
}

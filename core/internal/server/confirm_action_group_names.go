package server

import "strings"

func generatedTeamGroupName(name, teamID string) string {
	base := strings.TrimSpace(name)
	if base == "" {
		base = strings.TrimSpace(teamID)
	}
	id := strings.TrimSpace(teamID)
	if id == "" || strings.Contains(base, id) {
		return base
	}
	return base + " - " + id
}

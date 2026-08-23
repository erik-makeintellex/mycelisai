package swarm

import (
	"fmt"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

func runtimeTeamInputSubjects(teamID string, configured []string) []string {
	canonical := fmt.Sprintf(protocol.TopicTeamInternalCommand, teamID)
	subjects := []string{canonical}
	seen := map[string]struct{}{canonical: {}}

	for _, value := range configured {
		subject := strings.TrimSpace(value)
		if subject == "" {
			continue
		}
		if _, exists := seen[subject]; exists {
			continue
		}
		if !looksLikeNATSSubject(subject) {
			continue
		}
		seen[subject] = struct{}{}
		subjects = append(subjects, subject)
	}
	return subjects
}

func looksLikeNATSSubject(subject string) bool {
	if strings.ContainsAny(subject, " \t\r\n") {
		return false
	}
	if strings.Contains(subject, "..") {
		return false
	}
	return strings.Contains(subject, ".")
}

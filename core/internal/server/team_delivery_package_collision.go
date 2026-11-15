package server

import (
	"fmt"
	"path"
	"strings"

	"github.com/mycelis/core/internal/artifacts"
	"github.com/mycelis/core/pkg/protocol"
)

// NormalizeTeamPackageFilePath returns a canonical forward-slash relative path for collision checks.
func NormalizeTeamPackageFilePath(raw string) string {
	trimmed := strings.TrimSpace(strings.ReplaceAll(raw, "\\", "/"))
	if trimmed == "" {
		return ""
	}
	cleaned := path.Clean(trimmed)
	if cleaned == "." || cleaned == "/" {
		return ""
	}
	return strings.TrimPrefix(cleaned, "./")
}

// DetectTeamPackageCollisions identifies normalized file paths written by multiple agents
// or overwritten with differing content within the same team delivery package.
// Returns a map from normalized file path to the ordered slice of contributing agent IDs.
func DetectTeamPackageCollisions(items []artifacts.Artifact) map[string][]string {
	type pathEntry struct {
		agents       []string
		agentSeen    map[string]struct{}
		firstContent string
		count        int
		conflict     bool
	}

	byPath := make(map[string]*pathEntry)
	for _, item := range items {
		normPath := NormalizeTeamPackageFilePath(item.FilePath)
		if normPath == "" {
			continue
		}
		agentID := strings.TrimSpace(item.AgentID)
		if agentID == "" {
			agentID = "unattributed-agent"
		}
		signature := firstNonEmptyString(strings.TrimSpace(item.Content), strings.TrimSpace(item.Title))

		entry, exists := byPath[normPath]
		if !exists {
			byPath[normPath] = &pathEntry{
				agents:       []string{agentID},
				agentSeen:    map[string]struct{}{agentID: {}},
				firstContent: signature,
				count:        1,
			}
			continue
		}

		entry.count++
		if _, seen := entry.agentSeen[agentID]; !seen {
			entry.agentSeen[agentID] = struct{}{}
			entry.agents = append(entry.agents, agentID)
			entry.conflict = true
		} else if signature != entry.firstContent {
			entry.conflict = true
		}
	}

	collisions := make(map[string][]string)
	for normPath, entry := range byPath {
		if entry.conflict && entry.count > 1 {
			collisions[normPath] = entry.agents
		}
	}
	return collisions
}

func applyTeamPackageCollisionGovernance(
	teamID string,
	runID string,
	collisions map[string][]string,
) {
	if len(collisions) == 0 {
		return
	}
	parts := make([]string, 0, len(collisions))
	for normPath, agents := range collisions {
		parts = append(parts, fmt.Sprintf("%s (agents: %s)", normPath, strings.Join(agents, ", ")))
	}
	LogOperationalError(
		"team_delivery_package",
		"Multi-agent write collision detected in team delivery package.",
		strings.Join(parts, "; "),
		"Enforce One-Writer-Per-File ownership across swarm agents and resolve conflicting artifact writes before approving.",
		WithOperationalTeamID(teamID),
		WithOperationalRunID(runID),
		WithOperationalWhyItMatters("Concurrent multi-agent writes to the same path risk overwriting sibling agent deliverables."),
		WithOperationalSource(protocol.SourceKindInternalTool, "team.signal.result"),
	)
}

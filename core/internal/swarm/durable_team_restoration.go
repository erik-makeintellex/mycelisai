package swarm

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

// DurableTeamLoader reconstructs runtime-created teams that still own
// nonterminal durable work. It restores subscriptions and workers, not work state.
type DurableTeamLoader interface {
	LoadRuntimeTeams(context.Context) ([]*TeamManifest, error)
}

// PostgresDurableTeamLoader reads the existing team-work and group records.
type PostgresDurableTeamLoader struct {
	db *sql.DB
}

func NewPostgresDurableTeamLoader(db *sql.DB) *PostgresDurableTeamLoader {
	return &PostgresDurableTeamLoader{db: db}
}

func (l *PostgresDurableTeamLoader) LoadRuntimeTeams(ctx context.Context) ([]*TeamManifest, error) {
	if l == nil || l.db == nil {
		return nil, nil
	}
	rows, err := l.db.QueryContext(ctx, durableRuntimeTeamsQuery,
		string(protocol.TeamWorkStateNew),
		string(protocol.TeamWorkStateBriefed),
		string(protocol.TeamWorkStateQueued),
		string(protocol.TeamWorkStateRunning),
		string(protocol.TeamWorkStateNeedsOperator),
		string(protocol.TeamWorkStateReviewing),
		string(protocol.TeamWorkStateDegraded),
		string(protocol.TeamWorkStatePaused),
	)
	if err != nil {
		return nil, fmt.Errorf("load restorable runtime teams: %w", err)
	}
	defer rows.Close()

	manifests := make([]*TeamManifest, 0)
	for rows.Next() {
		manifest, scanErr := scanDurableRuntimeTeam(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		manifests = append(manifests, manifest)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate restorable runtime teams: %w", err)
	}
	return manifests, nil
}

const durableRuntimeTeamsQuery = `
WITH restorable AS (
    SELECT DISTINCT ON (team_id)
           team_id, objective, capability_requirements
      FROM team_work_items
     WHERE team_id <> ''
       AND state IN ($1, $2, $3, $4, $5, $6, $7, $8)
     ORDER BY team_id, updated_at DESC
)
SELECT restorable.team_id,
       COALESCE(NULLIF(group_record.name, ''), restorable.team_id),
       COALESCE(NULLIF(group_record.goal_statement, ''), restorable.objective),
       COALESCE(group_record.allowed_capabilities, '[]'::jsonb),
       restorable.capability_requirements,
       COALESCE(group_record.coordinator_profile, '')
  FROM restorable
  LEFT JOIN LATERAL (
      SELECT name, goal_statement, allowed_capabilities, coordinator_profile
        FROM collaboration_groups
       WHERE status = 'active'
         AND (expiry IS NULL OR expiry > NOW())
         AND team_ids @> jsonb_build_array(restorable.team_id)
       ORDER BY updated_at DESC
       LIMIT 1
  ) AS group_record ON TRUE
 ORDER BY restorable.team_id`

type durableTeamScanner interface {
	Scan(dest ...any) error
}

func scanDurableRuntimeTeam(scanner durableTeamScanner) (*TeamManifest, error) {
	var teamID, name, purpose, coordinator string
	var allowedJSON, requiredJSON []byte
	if err := scanner.Scan(&teamID, &name, &purpose, &allowedJSON, &requiredJSON, &coordinator); err != nil {
		return nil, fmt.Errorf("scan restorable runtime team: %w", err)
	}

	tools, err := mergeDurableTeamTools(allowedJSON, requiredJSON)
	if err != nil {
		return nil, fmt.Errorf("decode restorable runtime team %s: %w", teamID, err)
	}
	role := strings.TrimSpace(coordinator)
	if role == "" {
		role = "lead"
	}
	manifest := buildRuntimeTeamManifest(map[string]any{
		"team_id":       teamID,
		"name":          name,
		"role":          role,
		"tools":         tools,
		"system_prompt": durableTeamSystemPrompt(teamID, purpose),
	})
	if manifest == nil {
		return nil, fmt.Errorf("restorable runtime team has invalid ID %q", teamID)
	}
	manifest.Description = firstNonEmptyRuntimeString(purpose, manifest.Description)
	return manifest, nil
}

func mergeDurableTeamTools(values ...[]byte) ([]string, error) {
	tools := make([]string, 0)
	seen := map[string]struct{}{}
	for _, raw := range values {
		if len(raw) == 0 {
			continue
		}
		var decoded []string
		if err := json.Unmarshal(raw, &decoded); err != nil {
			return nil, err
		}
		for _, tool := range decoded {
			tool = strings.TrimSpace(tool)
			if tool == "" {
				continue
			}
			if _, exists := seen[tool]; exists {
				continue
			}
			seen[tool] = struct{}{}
			tools = append(tools, tool)
		}
	}
	return tools, nil
}

func durableTeamSystemPrompt(teamID, purpose string) string {
	return fmt.Sprintf(
		"You are the restored lead for runtime team %s. Durable purpose: %s. Continue only from governed team commands and preserve run, work-item, output, and proof correlation in every status or result.",
		teamID,
		firstNonEmptyRuntimeString(purpose, "Complete the team's retained work."),
	)
}

func firstNonEmptyRuntimeString(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}

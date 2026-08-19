package swarm

import (
	"context"
	"crypto/sha256"
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

// DurableTeamStore owns exact runtime-created team manifests. The persisted
// manifest is authoritative; work/group metadata is only a legacy fallback.
type DurableTeamStore interface {
	DurableTeamLoader
	SaveRuntimeTeam(context.Context, *TeamManifest) error
	DeleteRuntimeTeam(context.Context, string) error
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
	manifests, seen, err := l.loadPersistedRuntimeTeams(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := l.db.QueryContext(ctx, durableRuntimeTeamsQuery,
		durableRuntimeTenant,
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

	for rows.Next() {
		manifest, scanErr := scanDurableRuntimeTeam(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		if _, exists := seen[manifest.ID]; exists {
			continue
		}
		manifests = append(manifests, manifest)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate restorable runtime teams: %w", err)
	}
	return manifests, nil
}

const durableRuntimeTenant = "default"

const persistedRuntimeTeamsQuery = `
SELECT team_id, schema_version, manifest_digest, manifest
  FROM runtime_team_manifests
 WHERE tenant_id=$1
 ORDER BY team_id`

func (l *PostgresDurableTeamLoader) loadPersistedRuntimeTeams(ctx context.Context) ([]*TeamManifest, map[string]struct{}, error) {
	rows, err := l.db.QueryContext(ctx, persistedRuntimeTeamsQuery, durableRuntimeTenant)
	if err != nil {
		return nil, nil, fmt.Errorf("load persisted runtime team manifests: %w", err)
	}
	defer rows.Close()

	manifests := make([]*TeamManifest, 0)
	seen := map[string]struct{}{}
	for rows.Next() {
		var teamID, schemaVersion, storedDigest string
		var raw []byte
		if err := rows.Scan(&teamID, &schemaVersion, &storedDigest, &raw); err != nil {
			return nil, nil, fmt.Errorf("scan persisted runtime team manifest: %w", err)
		}
		if schemaVersion != "v1" {
			return nil, nil, fmt.Errorf("persisted runtime team %s has unsupported schema %q", teamID, schemaVersion)
		}
		var manifest TeamManifest
		if err := json.Unmarshal(raw, &manifest); err != nil {
			return nil, nil, fmt.Errorf("decode persisted runtime team %s: %w", teamID, err)
		}
		canonical, err := json.Marshal(&manifest)
		if err != nil {
			return nil, nil, fmt.Errorf("canonicalize persisted runtime team %s: %w", teamID, err)
		}
		digest := fmt.Sprintf("sha256:%x", sha256.Sum256(canonical))
		if storedDigest != digest {
			return nil, nil, fmt.Errorf("persisted runtime team %s failed digest validation", teamID)
		}
		teamID = strings.TrimSpace(teamID)
		if teamID == "" || strings.TrimSpace(manifest.ID) != teamID {
			return nil, nil, fmt.Errorf("persisted runtime team identity mismatch: row=%q manifest=%q", teamID, manifest.ID)
		}
		seen[teamID] = struct{}{}
		manifests = append(manifests, &manifest)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("iterate persisted runtime team manifests: %w", err)
	}
	return manifests, seen, nil
}

func (l *PostgresDurableTeamLoader) SaveRuntimeTeam(ctx context.Context, manifest *TeamManifest) error {
	if l == nil || l.db == nil {
		return nil
	}
	if manifest == nil || strings.TrimSpace(manifest.ID) == "" {
		return fmt.Errorf("save runtime team manifest: team identity is required")
	}
	raw, err := json.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("encode runtime team %s: %w", manifest.ID, err)
	}
	digest := fmt.Sprintf("sha256:%x", sha256.Sum256(raw))
	result, err := l.db.ExecContext(ctx, `
INSERT INTO runtime_team_manifests (tenant_id, team_id, manifest_digest, manifest)
VALUES ($1, $2, $3, $4::jsonb)
ON CONFLICT (tenant_id, team_id) DO NOTHING`, durableRuntimeTenant, strings.TrimSpace(manifest.ID), digest, raw)
	if err != nil {
		return fmt.Errorf("persist runtime team %s: %w", manifest.ID, err)
	}
	inserted, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("inspect persisted runtime team %s: %w", manifest.ID, err)
	}
	if inserted == 0 {
		var storedDigest string
		if err := l.db.QueryRowContext(ctx, `
SELECT manifest_digest FROM runtime_team_manifests
WHERE tenant_id=$1 AND team_id=$2`, durableRuntimeTenant, strings.TrimSpace(manifest.ID)).Scan(&storedDigest); err != nil {
			return fmt.Errorf("read existing runtime team %s: %w", manifest.ID, err)
		}
		if storedDigest != digest {
			return fmt.Errorf("runtime team %s already owns a different approved manifest", manifest.ID)
		}
	}
	return nil
}

func (l *PostgresDurableTeamLoader) DeleteRuntimeTeam(ctx context.Context, teamID string) error {
	if l == nil || l.db == nil {
		return nil
	}
	teamID = strings.TrimSpace(teamID)
	if teamID == "" {
		return fmt.Errorf("delete runtime team manifest: team identity is required")
	}
	if _, err := l.db.ExecContext(ctx, `DELETE FROM runtime_team_manifests WHERE tenant_id=$1 AND team_id=$2`, durableRuntimeTenant, teamID); err != nil {
		return fmt.Errorf("delete runtime team %s: %w", teamID, err)
	}
	return nil
}

const durableRuntimeTeamsQuery = `
WITH restorable AS (
    SELECT DISTINCT ON (team_id)
           team_id, objective, capability_requirements
      FROM team_work_items
     WHERE tenant_id=$1
       AND team_id <> ''
       AND state IN ($2, $3, $4, $5, $6, $7, $8, $9)
       AND NOT EXISTS (
           SELECT 1 FROM runtime_team_manifests persisted
            WHERE persisted.tenant_id=team_work_items.tenant_id
              AND persisted.team_id=team_work_items.team_id
       )
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
         AND tenant_id=$1
         AND (expiry IS NULL OR expiry > NOW())
         AND team_ids @> jsonb_build_array(restorable.team_id)
       ORDER BY updated_at DESC, id DESC
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

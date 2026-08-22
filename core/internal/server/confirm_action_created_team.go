package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

type confirmedCreatedTeam struct {
	TeamID          string
	WorkspaceFolder string
}

func confirmedActionCreatedTeamID(results []plannedToolExecutionResult) string {
	for _, result := range results {
		if strings.TrimSpace(result.Name) == "create_team" {
			if teamID := confirmedActionTeamID(result.Arguments); teamID != "" {
				return teamID
			}
		}
	}
	return ""
}

func confirmedActionCreatedTeamResults(results []plannedToolExecutionResult) map[string]bool {
	created := make(map[string]bool)
	for _, result := range results {
		if strings.TrimSpace(result.Name) != "create_team" {
			continue
		}
		if teamID, ok := confirmedCreatedTeamResult(result.Arguments, result.Output); ok {
			created[normalizeConfirmedRuntimeTeamID(teamID)] = true
		}
	}
	return created
}

func confirmedCreatedTeamResult(args map[string]any, raw string) (string, bool) {
	created, ok := parseConfirmedCreatedTeam(args, raw)
	return created.TeamID, ok
}

func parseConfirmedCreatedTeam(args map[string]any, raw string) (confirmedCreatedTeam, bool) {
	var output struct {
		Status          string `json:"status"`
		TeamID          string `json:"team_id"`
		WorkspaceFolder string `json:"workspace_folder"`
	}
	if json.Unmarshal([]byte(raw), &output) != nil || output.Status != "created" {
		return confirmedCreatedTeam{}, false
	}
	teamID := normalizeConfirmedRuntimeTeamID(firstNonEmptyString(output.TeamID, confirmedActionTeamID(args)))
	if teamID == "" {
		return confirmedCreatedTeam{}, false
	}
	return confirmedCreatedTeam{
		TeamID:          teamID,
		WorkspaceFolder: strings.TrimSpace(output.WorkspaceFolder),
	}, true
}

func (s *AdminServer) claimConfirmedCreatedTeamLocked(ctx context.Context, fixtureScopeID string, args map[string]any, output string) error {
	return s.claimConfirmedCreatedTeam(ctx, fixtureScopeID, args, output)
}

func (s *AdminServer) claimConfirmedCreatedTeam(ctx context.Context, fixtureScopeID string, args map[string]any, output string) error {
	if strings.TrimSpace(fixtureScopeID) == "" {
		return nil
	}
	createdTeam, created := parseConfirmedCreatedTeam(args, output)
	if !created {
		return nil
	}
	requestedTeamID := normalizeConfirmedRuntimeTeamID(confirmedActionTeamID(args))
	if requestedTeamID == "" || createdTeam.TeamID != requestedTeamID {
		return fmt.Errorf("%w: requested %q but tool returned %q", errQAFixtureTeamIdentityMismatch, requestedTeamID, createdTeam.TeamID)
	}
	expectedWorkspace, expectedErr := confirmedTeamWorkspaceFolder(createdTeam.TeamID)
	if expectedErr != nil {
		return expectedErr
	}
	workspace, workspaceErr := normalizeRequestedGroupWorkspaceFolder(createdTeam.WorkspaceFolder)
	if workspaceErr != nil || workspace != expectedWorkspace {
		s.cleanupUnclaimedFixtureTeam(createdTeam.TeamID, expectedWorkspace)
		return fmt.Errorf("created team returned an unexpected workspace folder")
	}
	err := s.claimQAFixtureResourcesLocked(ctx, fixtureScopeID, []qaFixtureResource{
		{Kind: "team", Ref: createdTeam.TeamID},
		{Kind: "workspace_path", Ref: workspace},
	})
	if err != nil {
		cleanupErr := s.cleanupUnclaimedFixtureTeam(createdTeam.TeamID, workspace)
		return errors.Join(err, cleanupErr)
	}
	return nil
}

func (s *AdminServer) ensureQAFixtureTeamCreationAvailable(ctx context.Context, fixtureScopeID, runID string, args map[string]any) error {
	if strings.TrimSpace(fixtureScopeID) == "" {
		return nil
	}
	teamID := normalizeConfirmedRuntimeTeamID(confirmedActionTeamID(args))
	workspace, workspaceErr := confirmedTeamWorkspaceFolder(teamID)
	if teamID == "" || workspaceErr != nil {
		return fmt.Errorf("create_team requires a valid team_id")
	}
	db := s.getDB()
	if db == nil {
		return errors.New("database not available")
	}
	var owned bool
	if err := db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM qa_fixture_resources
			WHERE scope_id=$1::uuid AND resource_kind='team' AND resource_ref=$2
		)
	`, fixtureScopeID, teamID).Scan(&owned); err != nil {
		return err
	}
	if owned {
		return nil
	}
	if s.Soma != nil {
		for _, team := range s.Soma.ListTeams() {
			if team != nil && strings.TrimSpace(team.ID) == teamID {
				return fmt.Errorf("%w: team %q is already active", errQAFixtureTeamPreexisting, teamID)
			}
		}
	}
	var registryOwned, foreignWork, groupOwned, runtimeManifest bool
	if err := db.QueryRowContext(ctx, `
		SELECT
		EXISTS (
			SELECT 1
			FROM team_registry_entries registry
			LEFT JOIN outcome_projects project
				ON project.id=registry.project_id AND project.tenant_id=registry.tenant_id
			WHERE registry.tenant_id=$1 AND registry.team_id=$2
				AND (project.run_id IS NULL OR project.run_id <> $3)
		),
		EXISTS (
			SELECT 1 FROM team_work_items
			WHERE tenant_id=$1 AND team_id=$2
				AND (run_id IS NULL OR run_id::text <> $3)
		),
		EXISTS (
			SELECT 1 FROM collaboration_groups WHERE tenant_id=$1 AND team_ids ? $2
		),
		EXISTS (
			SELECT 1 FROM runtime_team_manifests WHERE tenant_id=$1 AND team_id=$2
		)
	`, qaFixtureTenantID, teamID, strings.TrimSpace(runID)).Scan(&registryOwned, &foreignWork, &groupOwned, &runtimeManifest); err != nil {
		return err
	}
	retainedKinds := []string{}
	if registryOwned {
		retainedKinds = append(retainedKinds, "registry")
	}
	if foreignWork {
		retainedKinds = append(retainedKinds, "prior work")
	}
	if groupOwned {
		retainedKinds = append(retainedKinds, "group")
	}
	if runtimeManifest {
		retainedKinds = append(retainedKinds, "runtime manifest")
	}
	if len(retainedKinds) > 0 {
		return fmt.Errorf("%w: team %q has retained %s records", errQAFixtureTeamPreexisting, teamID, strings.Join(retainedKinds, ", "))
	}
	target, _, err := resolveWorkspacePath(workspace, false)
	if err != nil {
		return err
	}
	if _, err := os.Stat(target); err == nil {
		return fmt.Errorf("%w: team %q already has a workspace", errQAFixtureTeamPreexisting, teamID)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("inspect team workspace: %w", err)
	}
	return nil
}

func normalizeConfirmedRuntimeTeamID(raw string) string {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	normalized = strings.ReplaceAll(normalized, " ", "-")
	return strings.ReplaceAll(normalized, "_", "-")
}

func confirmedTeamWorkspaceFolder(teamID string) (string, error) {
	trimmed := strings.TrimSpace(teamID)
	if trimmed == "" || trimmed == "." || trimmed == ".." || strings.ContainsAny(trimmed, `/\\`) {
		return "", fmt.Errorf("invalid team_id for workspace isolation")
	}
	return groupWorkspaceRoot + "/" + trimmed, nil
}

func (s *AdminServer) cleanupUnclaimedFixtureTeam(teamID, workspace string) error {
	if s.Soma != nil {
		s.Soma.StopTeam(teamID)
	}
	_, err := removeGroupWorkspaceFolder(workspace)
	if err != nil {
		return fmt.Errorf("clean up unclaimed team workspace: %w", err)
	}
	return nil
}

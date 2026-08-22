package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

func (s *AdminServer) validateQAFixtureResources(
	ctx context.Context,
	tx *sql.Tx,
	scope qaFixtureScope,
	resources []qaFixtureResource,
) error {
	claimedResources, err := listQAFixtureResources(ctx, tx, scope.ID)
	if err != nil {
		return err
	}
	claimedRefs := fixtureResourceRefs(claimedResources)
	groupRoots := make([]string, 0)
	groupTeams := make(map[string]bool)

	for _, resource := range resources {
		var claimed bool
		if err := tx.QueryRowContext(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM qa_fixture_resources
				WHERE resource_kind=$1 AND resource_ref=$2 AND scope_id<>$3::uuid
			)
		`, resource.Kind, resource.Ref, scope.ID).Scan(&claimed); err != nil {
			return err
		}
		if claimed {
			return fmt.Errorf(
				"%w: fixture %s %q is already owned by another scope",
				errQAFixtureResourceUnowned,
				resource.Kind,
				resource.Ref,
			)
		}
	}

	for _, groupID := range claimedRefs["group"] {
		root, teams, err := validateFixtureGroup(ctx, tx, scope.TenantID, groupID)
		if err != nil {
			return err
		}
		groupRoots = append(groupRoots, root)
		for _, teamID := range teams {
			groupTeams[teamID] = true
		}
	}
	for _, resource := range resources {
		if err := s.validateQAFixtureResource(ctx, tx, scope, resource, claimedRefs, groupRoots, groupTeams); err != nil {
			return err
		}
	}
	return nil
}

func (s *AdminServer) validateQAFixtureResource(
	ctx context.Context,
	tx *sql.Tx,
	scope qaFixtureScope,
	resource qaFixtureResource,
	refs map[string][]string,
	groupRoots []string,
	groupTeams map[string]bool,
) error {
	switch resource.Kind {
	case "organization":
		fixtureScopeID, ok := s.organizationStore().QAFixtureScope(resource.Ref)
		if !ok {
			return unownedFixtureResource(resource)
		}
		if fixtureScopeID != scope.ID {
			return fmt.Errorf(
				"%w: organization %q belongs to fixture scope %q, not %q",
				errQAFixtureResourceUnowned,
				resource.Ref,
				fixtureScopeID,
				scope.ID,
			)
		}
		return nil
	case "group":
		if containsFixtureRef(refs["group"], resource.Ref) {
			return nil
		}
		return unownedFixtureResource(resource)
	case "team":
		if groupTeams[resource.Ref] {
			return nil
		}
		owned, err := fixtureTeamHasRun(ctx, tx, scope, resource.Ref, refs["run"])
		if err != nil || !owned {
			if err != nil {
				return err
			}
			return unownedFixtureResource(resource)
		}
	case "run":
		if containsFixtureRef(refs["run"], resource.Ref) {
			return nil
		}
		return unownedFixtureResource(resource)
	case "outcome":
		if containsFixtureRef(refs["outcome"], resource.Ref) {
			return nil
		}
		return unownedFixtureResource(resource)
	case "artifact":
		return validateFixtureArtifact(ctx, tx, resource, refs)
	case "workspace_path":
		for _, root := range groupRoots {
			if resource.Ref == root || strings.HasPrefix(resource.Ref, root+"/") {
				return nil
			}
		}
		return unownedFixtureResource(resource)
	}
	return unownedFixtureResource(resource)
}

func validateFixtureGroup(
	ctx context.Context,
	tx *sql.Tx,
	tenantID string,
	groupID string,
) (string, []string, error) {
	var root string
	var rawTeams []byte
	err := tx.QueryRowContext(ctx, `
		SELECT workspace_folder, team_ids
		FROM collaboration_groups WHERE id=$1::uuid AND tenant_id=$2
	`, groupID, tenantID).Scan(&root, &rawTeams)
	if err != nil {
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return "", nil, err
		}
		return "", nil, unownedFixtureResource(qaFixtureResource{Kind: "group", Ref: groupID})
	}
	var teams []string
	if err := json.Unmarshal(rawTeams, &teams); err != nil {
		return "", nil, err
	}
	return strings.TrimSpace(root), teams, nil
}

func fixtureTeamHasRun(
	ctx context.Context,
	tx *sql.Tx,
	scope qaFixtureScope,
	teamID string,
	runIDs []string,
) (bool, error) {
	runs := make(map[string]bool, len(runIDs))
	for _, runID := range runIDs {
		runs[runID] = true
	}
	rows, err := tx.QueryContext(ctx, `
		SELECT COALESCE(run_id::text, '') FROM team_work_items
		WHERE tenant_id=$2 AND team_id=$1
	`, teamID, scope.TenantID)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	for rows.Next() {
		var runID string
		if err := rows.Scan(&runID); err != nil {
			return false, err
		}
		if runs[runID] {
			return true, nil
		}
	}
	return false, rows.Err()
}

func validateFixtureArtifact(
	ctx context.Context,
	tx *sql.Tx,
	resource qaFixtureResource,
	refs map[string][]string,
) error {
	var traceID string
	err := tx.QueryRowContext(ctx, `
		SELECT COALESCE(trace_id, '') FROM artifacts WHERE id=$1::uuid
	`, resource.Ref).Scan(&traceID)
	if err != nil || !containsFixtureRef(refs["run"], traceID) {
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		return unownedFixtureResource(resource)
	}
	return nil
}

func fixtureResourceRefs(resources []qaFixtureResource) map[string][]string {
	refs := make(map[string][]string)
	for _, resource := range resources {
		refs[resource.Kind] = append(refs[resource.Kind], resource.Ref)
	}
	return refs
}

func containsFixtureRef(refs []string, target string) bool {
	for _, ref := range refs {
		if ref == target {
			return true
		}
	}
	return false
}

func unownedFixtureResource(resource qaFixtureResource) error {
	return fmt.Errorf("%w: %s %q has no exact provenance in this fixture scope", errQAFixtureResourceUnowned, resource.Kind, resource.Ref)
}

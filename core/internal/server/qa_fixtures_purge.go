package server

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
)

func newQAFixturePurgeResult(
	scope qaFixtureScope,
	resources []qaFixtureResource,
	confirmed bool,
) qaFixturePurgeResult {
	counts := make(map[string]int)
	for _, resource := range resources {
		counts[resource.Kind]++
	}
	return qaFixturePurgeResult{
		Scope:               scope,
		Confirmed:           confirmed,
		RegisteredResources: len(resources),
		ResourceCounts:      counts,
		DeletedRows:         make(map[string]int64),
		NATSUntouched:       true,
	}
}

func deleteQAFixtureDatabaseResources(
	ctx context.Context,
	tx *sql.Tx,
	tenantID string,
	scopeID string,
	resources []qaFixtureResource,
	deleted map[string]int64,
) ([]string, error) {
	var removedOrganizations []string
	configRefs := make([]string, 0)
	for _, resource := range resources {
		if resource.Kind == "config_document" {
			configRefs = append(configRefs, resource.Ref)
		}
	}
	if err := deleteQAFixtureConfigDocuments(ctx, tx, tenantID, configRefs, deleted); err != nil {
		return nil, fmt.Errorf("delete fixture config documents: %w", err)
	}
	for _, kind := range []string{"organization", "artifact", "group", "team", "outcome", "run"} {
		for _, resource := range resources {
			if resource.Kind != kind {
				continue
			}
			if kind == "organization" {
				removed, err := deleteQAFixtureOrganization(ctx, tx, tenantID, scopeID, resource, deleted)
				if err != nil {
					return nil, fmt.Errorf("delete fixture %s %q: %w", resource.Kind, resource.Ref, err)
				}
				if removed {
					removedOrganizations = append(removedOrganizations, resource.Ref)
				}
				continue
			}
			if err := deleteQAFixtureDatabaseResource(ctx, tx, tenantID, resource, deleted); err != nil {
				return nil, fmt.Errorf("delete fixture %s %q: %w", resource.Kind, resource.Ref, err)
			}
		}
	}
	sort.Strings(removedOrganizations)
	return removedOrganizations, nil
}

func deleteQAFixtureDatabaseResource(
	ctx context.Context,
	tx *sql.Tx,
	tenantID string,
	resource qaFixtureResource,
	deleted map[string]int64,
) error {
	if resource.Kind == "config_document" {
		return deleteQAFixtureConfigDocuments(ctx, tx, tenantID, []string{resource.Ref}, deleted)
	}
	queries := map[string][]struct {
		label string
		sql   string
	}{
		"artifact": {{"artifacts", `DELETE FROM artifacts WHERE id=$1::uuid AND $2<>''`}},
		"group":    {{"collaboration_groups", `DELETE FROM collaboration_groups WHERE id=$1::uuid AND tenant_id=$2`}},
		"team": {
			{"runtime_team_manifests", `DELETE FROM runtime_team_manifests WHERE team_id=$1 AND tenant_id=$2`},
			{"team_registry_entries", `DELETE FROM team_registry_entries WHERE team_id=$1 AND tenant_id=$2`},
			{"team_work_items", `DELETE FROM team_work_items WHERE team_id=$1 AND tenant_id=$2`},
		},
		"outcome": {{"outcome_projects", `DELETE FROM outcome_projects WHERE tenant_id=$2 AND (id::text=$1 OR outcome_id=$1)`}},
		"run": {
			{"artifacts", `DELETE FROM artifacts WHERE trace_id=$1 AND $2<>''`},
			{"outcome_projects", `DELETE FROM outcome_projects WHERE run_id=$1 AND tenant_id=$2`},
			{"team_work_items", `DELETE FROM team_work_items WHERE run_id=$1::uuid AND tenant_id=$2`},
			{"execution_dispatch_outbox", `DELETE FROM execution_dispatch_outbox WHERE run_id=$1 AND $2<>''`},
			{"proof_artifacts", `DELETE FROM proof_artifacts WHERE run_id=$1::uuid AND tenant_id=$2`},
			{"execution_contracts", `DELETE FROM execution_contracts WHERE run_id=$1::uuid AND tenant_id=$2`},
			{"conversation_turns", `DELETE FROM conversation_turns WHERE run_id=$1::uuid AND tenant_id=$2`},
			{"mission_runs", `DELETE FROM mission_runs WHERE id=$1::uuid AND tenant_id=$2`},
		},
	}
	for _, query := range queries[resource.Kind] {
		result, err := tx.ExecContext(ctx, query.sql, resource.Ref, tenantID)
		if err != nil {
			return err
		}
		rows, err := result.RowsAffected()
		if err != nil {
			return err
		}
		deleted[query.label] += rows
	}
	return nil
}

// deleteQAFixtureOrganization removes an organization row only when its
// authoritative qa_fixture_scope_id column equals the purging scope. A claimed
// organization that still exists outside that scope (non-fixture or another
// scope) fails the purge transaction as unowned; an already-absent row is a
// no-op so an interrupted purge can resume. id::text keeps a non-UUID legacy
// ref from raising a cast error.
func deleteQAFixtureOrganization(
	ctx context.Context,
	tx *sql.Tx,
	tenantID string,
	scopeID string,
	resource qaFixtureResource,
	deleted map[string]int64,
) (bool, error) {
	if scopeID == "" {
		return false, unownedFixtureResource(resource)
	}
	result, err := tx.ExecContext(ctx, `
		DELETE FROM organizations
		WHERE id::text=$1 AND tenant_id=$2 AND qa_fixture_scope_id=$3`,
		resource.Ref, tenantID, scopeID)
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	if rows > 0 {
		deleted["organizations"] += rows
		return true, nil
	}
	var exists bool
	if err := tx.QueryRowContext(ctx, `
		SELECT EXISTS (SELECT 1 FROM organizations WHERE id::text=$1 AND tenant_id=$2)`,
		resource.Ref, tenantID).Scan(&exists); err != nil {
		return false, err
	}
	if exists {
		return false, unownedFixtureResource(resource)
	}
	return false, nil
}

func (s *AdminServer) stopQAFixtureProducers(resources []qaFixtureResource) []string {
	var stoppedTeams []string
	for _, resource := range resources {
		if resource.Kind == "team" && s.Soma != nil && s.Soma.StopTeam(resource.Ref) {
			stoppedTeams = append(stoppedTeams, resource.Ref)
		}
	}
	sort.Strings(stoppedTeams)
	return stoppedTeams
}

func cleanupQAFixtureWorkspaceResources(resources []qaFixtureResource) ([]string, []string) {
	var removedPaths, warnings []string
	for _, resource := range resources {
		if resource.Kind != "workspace_path" {
			continue
		}
		removed, err := removeGroupWorkspaceFolder(resource.Ref)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("workspace %s: %v", resource.Ref, err))
		} else if removed {
			removedPaths = append(removedPaths, resource.Ref)
		}
	}
	sort.Strings(removedPaths)
	sort.Strings(warnings)
	return removedPaths, warnings
}

func updateQAFixtureScopeStatus(
	ctx context.Context,
	db *sql.DB,
	scopeID string,
	status string,
	releaseClaims bool,
) (qaFixtureScope, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return qaFixtureScope{}, err
	}
	defer func() { _ = tx.Rollback() }()
	if releaseClaims {
		if _, err := tx.ExecContext(ctx, `
			DELETE FROM qa_fixture_resources WHERE scope_id=$1::uuid
		`, scopeID); err != nil {
			return qaFixtureScope{}, err
		}
	}
	var scope qaFixtureScope
	err = tx.QueryRowContext(ctx, `
		UPDATE qa_fixture_scopes
		SET status=$2, updated_at=NOW()
		WHERE id=$1::uuid AND tenant_id=$3
		RETURNING id::text, tenant_id, owner_ref, execution_ref, status,
			expires_at, created_at, updated_at
	`, scopeID, status, qaFixtureTenantID).Scan(
		&scope.ID,
		&scope.TenantID,
		&scope.OwnerRef,
		&scope.ExecutionRef,
		&scope.Status,
		&scope.ExpiresAt,
		&scope.CreatedAt,
		&scope.UpdatedAt,
	)
	if err != nil {
		return qaFixtureScope{}, err
	}
	if err := tx.Commit(); err != nil {
		return qaFixtureScope{}, err
	}
	return scope, nil
}

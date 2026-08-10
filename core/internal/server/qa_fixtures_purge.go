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
	resources []qaFixtureResource,
	deleted map[string]int64,
) error {
	for _, kind := range []string{"artifact", "group", "team", "outcome", "run"} {
		for _, resource := range resources {
			if resource.Kind != kind {
				continue
			}
			if err := deleteQAFixtureDatabaseResource(ctx, tx, tenantID, resource, deleted); err != nil {
				return fmt.Errorf("delete fixture %s %q: %w", resource.Kind, resource.Ref, err)
			}
		}
	}
	return nil
}

func deleteQAFixtureDatabaseResource(
	ctx context.Context,
	tx *sql.Tx,
	tenantID string,
	resource qaFixtureResource,
	deleted map[string]int64,
) error {
	queries := map[string][]struct {
		label string
		sql   string
	}{
		"artifact": {{"artifacts", `DELETE FROM artifacts WHERE id=$1::uuid AND $2<>''`}},
		"group":    {{"collaboration_groups", `DELETE FROM collaboration_groups WHERE id=$1::uuid AND tenant_id=$2`}},
		"team": {
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

func (s *AdminServer) stopQAFixtureProducers(
	resources []qaFixtureResource,
) ([]string, []string) {
	var stoppedTeams, removedOrganizations []string
	for _, resource := range resources {
		switch resource.Kind {
		case "team":
			if s.Soma != nil && s.Soma.StopTeam(resource.Ref) {
				stoppedTeams = append(stoppedTeams, resource.Ref)
			}
		case "organization":
			if s.organizationStore().Delete(resource.Ref) {
				removedOrganizations = append(removedOrganizations, resource.Ref)
			}
		}
	}
	sort.Strings(stoppedTeams)
	sort.Strings(removedOrganizations)
	return stoppedTeams, removedOrganizations
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

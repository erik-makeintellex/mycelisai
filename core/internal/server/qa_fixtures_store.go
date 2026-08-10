package server

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

const qaFixtureTenantID = "default"

func (s *AdminServer) createQAFixtureScope(
	ctx context.Context,
	ownerRef string,
	executionRef string,
	expiresAt time.Time,
) (qaFixtureScope, error) {
	db := s.getDB()
	if db == nil {
		return qaFixtureScope{}, errors.New("database not available")
	}
	scope := qaFixtureScope{ID: uuid.NewString()}
	err := db.QueryRowContext(ctx, `
		INSERT INTO qa_fixture_scopes (
			id, tenant_id, owner_ref, execution_ref, expires_at
		) VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (tenant_id, owner_ref, execution_ref) DO NOTHING
		RETURNING id::text, tenant_id, owner_ref, execution_ref, status,
			expires_at, created_at, updated_at
	`, scope.ID, qaFixtureTenantID, ownerRef, executionRef, expiresAt).Scan(
		&scope.ID,
		&scope.TenantID,
		&scope.OwnerRef,
		&scope.ExecutionRef,
		&scope.Status,
		&scope.ExpiresAt,
		&scope.CreatedAt,
		&scope.UpdatedAt,
	)
	return scope, err
}

func (s *AdminServer) validateOpenQAFixtureScope(ctx context.Context, scopeID string) error {
	db := s.getDB()
	if db == nil {
		return errors.New("database not available")
	}
	scope, err := loadQAFixtureScope(ctx, db, scopeID, false)
	if err != nil {
		return err
	}
	if scope.Status != "open" || time.Now().UTC().After(scope.ExpiresAt) {
		return errQAFixtureScopeClosed
	}
	return nil
}

func (s *AdminServer) addQAFixtureResources(
	ctx context.Context,
	scopeID string,
	ownerRef string,
	executionRef string,
	resources []qaFixtureResource,
) ([]qaFixtureResource, error) {
	owner, execution, err := normalizeQAFixtureOwnership(ownerRef, executionRef)
	if err != nil {
		return nil, err
	}
	db := s.getDB()
	if db == nil {
		return nil, errors.New("database not available")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	scope, err := loadQAFixtureScope(ctx, tx, scopeID, true)
	if err != nil {
		return nil, err
	}
	if scope.OwnerRef != owner || scope.ExecutionRef != execution {
		return nil, errQAFixtureScopeMismatch
	}
	if scope.Status != "open" {
		return nil, errQAFixtureScopeClosed
	}
	if time.Now().UTC().After(scope.ExpiresAt) {
		return nil, errors.New("fixture scope has expired")
	}

	if err := s.validateQAFixtureResources(ctx, tx, scope, resources); err != nil {
		return nil, err
	}
	for _, resource := range resources {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO qa_fixture_resources (id, scope_id, resource_kind, resource_ref)
			VALUES ($1, $2::uuid, $3, $4)
			ON CONFLICT (scope_id, resource_kind, resource_ref) DO NOTHING
		`, uuid.NewString(), scope.ID, resource.Kind, resource.Ref); err != nil {
			return nil, err
		}
	}
	registered, err := listQAFixtureResources(ctx, tx, scope.ID)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return registered, nil
}

func (s *AdminServer) purgeQAFixtureScope(
	ctx context.Context,
	scopeID string,
	req qaFixturePurgeRequest,
) (qaFixturePurgeResult, error) {
	owner, execution, err := normalizeQAFixtureOwnership(req.OwnerRef, req.ExecutionRef)
	if err != nil {
		return qaFixturePurgeResult{}, err
	}
	db := s.getDB()
	if db == nil {
		return qaFixturePurgeResult{}, errors.New("database not available")
	}
	releaseLock, err := acquireQAFixturePurgeLock(ctx, db, scopeID)
	if err != nil {
		return qaFixturePurgeResult{}, err
	}
	defer releaseLock()

	scope, resources, err := readQAFixtureScope(ctx, db, scopeID, owner, execution)
	if err != nil {
		return qaFixturePurgeResult{}, err
	}
	result := newQAFixturePurgeResult(scope, resources, req.Confirm)
	if !req.Confirm {
		return result, nil
	}
	if scope.Status == "purged" {
		return qaFixturePurgeResult{}, errQAFixtureScopeClosed
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return qaFixturePurgeResult{}, err
	}
	defer func() { _ = tx.Rollback() }()
	locked, err := loadQAFixtureScope(ctx, tx, scope.ID, true)
	if err != nil {
		return qaFixturePurgeResult{}, err
	}
	if locked.OwnerRef != owner || locked.ExecutionRef != execution {
		return qaFixturePurgeResult{}, errQAFixtureScopeMismatch
	}
	if locked.Status == "purged" {
		return qaFixturePurgeResult{}, errQAFixtureScopeClosed
	}
	resources, err = listQAFixtureResources(ctx, tx, scope.ID)
	if err != nil {
		return qaFixturePurgeResult{}, err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE qa_fixture_scopes SET status='purging', updated_at=NOW()
		WHERE id=$1::uuid AND tenant_id=$2
	`, scope.ID, scope.TenantID); err != nil {
		return qaFixturePurgeResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return qaFixturePurgeResult{}, err
	}

	result.StoppedTeams, result.RemovedOrganizations = s.stopQAFixtureProducers(resources)
	if err := s.deleteQAFixtureDurableResources(ctx, scope, resources, &result); err != nil {
		_, _ = updateQAFixtureScopeStatus(ctx, db, scope.ID, "partial", false)
		return qaFixturePurgeResult{}, err
	}
	result.RemovedPaths, result.Warnings = cleanupQAFixtureWorkspaceResources(resources)
	status := "purged"
	if len(result.Warnings) > 0 {
		status = "partial"
	}
	updated, err := updateQAFixtureScopeStatus(ctx, db, scope.ID, status, status == "purged")
	if err != nil {
		return qaFixturePurgeResult{}, err
	}
	result.Scope = updated
	return result, nil
}

func (s *AdminServer) deleteQAFixtureDurableResources(
	ctx context.Context,
	scope qaFixtureScope,
	resources []qaFixtureResource,
	result *qaFixturePurgeResult,
) error {
	tx, err := s.getDB().BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if err := deleteQAFixtureDatabaseResources(ctx, tx, scope.TenantID, resources, result.DeletedRows); err != nil {
		return err
	}
	return tx.Commit()
}

type qaFixtureQuerier interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func normalizeQAFixtureOwnership(ownerRef, executionRef string) (string, string, error) {
	owner, err := normalizeQAFixtureIdentity("owner_ref", ownerRef)
	if err != nil {
		return "", "", err
	}
	execution, err := normalizeQAFixtureIdentity("execution_ref", executionRef)
	if err != nil {
		return "", "", err
	}
	return owner, execution, nil
}

func readQAFixtureScope(
	ctx context.Context,
	db *sql.DB,
	scopeID string,
	ownerRef string,
	executionRef string,
) (qaFixtureScope, []qaFixtureResource, error) {
	scope, err := loadQAFixtureScope(ctx, db, scopeID, false)
	if err != nil {
		return qaFixtureScope{}, nil, err
	}
	if scope.OwnerRef != ownerRef || scope.ExecutionRef != executionRef {
		return qaFixtureScope{}, nil, errQAFixtureScopeMismatch
	}
	resources, err := listQAFixtureResources(ctx, db, scope.ID)
	return scope, resources, err
}

func loadQAFixtureScope(
	ctx context.Context,
	q qaFixtureQuerier,
	scopeID string,
	forUpdate bool,
) (qaFixtureScope, error) {
	if _, err := uuid.Parse(strings.TrimSpace(scopeID)); err != nil {
		return qaFixtureScope{}, sql.ErrNoRows
	}
	query := `
		SELECT id::text, tenant_id, owner_ref, execution_ref, status,
			expires_at, created_at, updated_at
		FROM qa_fixture_scopes
		WHERE id=$1::uuid AND tenant_id=$2
	`
	if forUpdate {
		query += " FOR UPDATE"
	}
	var scope qaFixtureScope
	err := q.QueryRowContext(ctx, query, scopeID, qaFixtureTenantID).Scan(
		&scope.ID,
		&scope.TenantID,
		&scope.OwnerRef,
		&scope.ExecutionRef,
		&scope.Status,
		&scope.ExpiresAt,
		&scope.CreatedAt,
		&scope.UpdatedAt,
	)
	return scope, err
}

func listQAFixtureResources(
	ctx context.Context,
	q qaFixtureQuerier,
	scopeID string,
) ([]qaFixtureResource, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT resource_kind, resource_ref
		FROM qa_fixture_resources
		WHERE scope_id=$1::uuid
		ORDER BY resource_kind, resource_ref
	`, scopeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	resources := make([]qaFixtureResource, 0)
	for rows.Next() {
		var resource qaFixtureResource
		if err := rows.Scan(&resource.Kind, &resource.Ref); err != nil {
			return nil, err
		}
		resources = append(resources, resource)
	}
	return resources, rows.Err()
}

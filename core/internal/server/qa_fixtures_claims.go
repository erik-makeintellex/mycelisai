package server

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

type qaFixtureFenceContextKey struct{}

func withQAFixtureFenceHeld(ctx context.Context, scopeID string) context.Context {
	return context.WithValue(ctx, qaFixtureFenceContextKey{}, strings.TrimSpace(scopeID))
}

func qaFixtureFenceHeld(ctx context.Context, scopeID string) bool {
	held, _ := ctx.Value(qaFixtureFenceContextKey{}).(string)
	return held != "" && held == strings.TrimSpace(scopeID)
}

// claimQAFixtureResources records provenance at a trusted creation boundary.
// Callers must pass only IDs or paths they just created for the scoped request.
func (s *AdminServer) claimQAFixtureResources(
	ctx context.Context,
	scopeID string,
	resources []qaFixtureResource,
) error {
	if strings.TrimSpace(scopeID) == "" || len(resources) == 0 {
		return nil
	}
	return s.withQAFixtureScopeLock(ctx, scopeID, func() error {
		return s.claimQAFixtureResourcesLocked(ctx, scopeID, resources)
	})
}

func (s *AdminServer) claimQAFixtureResourcesLocked(
	ctx context.Context,
	scopeID string,
	resources []qaFixtureResource,
) error {
	if strings.TrimSpace(scopeID) == "" || len(resources) == 0 {
		return nil
	}
	db := s.getDB()
	if db == nil {
		return errors.New("database not available")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	for _, resource := range resources {
		if err := claimQAFixtureResourceTx(ctx, tx, scopeID, resource); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *AdminServer) withQAFixtureScopeLock(ctx context.Context, scopeID string, action func() error) error {
	if strings.TrimSpace(scopeID) == "" {
		return action()
	}
	if qaFixtureFenceHeld(ctx, scopeID) {
		return action()
	}
	db := s.getDB()
	if db == nil {
		return errors.New("database not available")
	}
	release, err := acquireQAFixturePurgeLock(ctx, db, scopeID)
	if err != nil {
		return err
	}
	defer release()
	return action()
}

func claimQAFixtureResourceTx(
	ctx context.Context,
	tx *sql.Tx,
	scopeID string,
	resource qaFixtureResource,
) error {
	if strings.TrimSpace(scopeID) == "" {
		return nil
	}
	normalized, err := normalizeQAFixtureResource(resource)
	if err != nil {
		return err
	}
	scope, err := loadQAFixtureScope(ctx, tx, scopeID, true)
	if err != nil {
		return err
	}
	if scope.Status != "open" || time.Now().UTC().After(scope.ExpiresAt) {
		return errQAFixtureScopeClosed
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO qa_fixture_resources (id, scope_id, resource_kind, resource_ref)
		VALUES ($1, $2::uuid, $3, $4)
		ON CONFLICT (scope_id, resource_kind, resource_ref) DO NOTHING
	`, uuid.NewString(), scope.ID, normalized.Kind, normalized.Ref)
	return err
}

func acquireQAFixturePurgeLock(ctx context.Context, db *sql.DB, scopeID string) (func(), error) {
	conn, err := db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := conn.ExecContext(ctx, `SELECT pg_advisory_lock(hashtextextended($1, 0))`, scopeID); err != nil {
		_ = conn.Close()
		return nil, err
	}
	release := func() {
		unlockCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_, _ = conn.ExecContext(unlockCtx, `SELECT pg_advisory_unlock(hashtextextended($1, 0))`, scopeID)
		_ = conn.Close()
	}
	return release, nil
}

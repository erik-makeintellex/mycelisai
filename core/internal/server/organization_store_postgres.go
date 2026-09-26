package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

// postgresOrganizationRepository stores organizations in the ORGANIZATIONS
// block of 001_current_schema.sql. document holds the full
// OrganizationHomePayload; qa_fixture_scope_id is the authoritative fixture
// owner (the payload tags it json:"-", so it never lives in document).
type postgresOrganizationRepository struct {
	db *sql.DB
}

func organizationUUID(id string) (string, bool) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return "", false
	}
	return parsed.String(), true
}

func encodeOrganizationDocument(home OrganizationHomePayload) ([]byte, error) {
	document, err := json.Marshal(home)
	if err != nil {
		return nil, fmt.Errorf("encode organization document: %w", err)
	}
	return document, nil
}

func decodeOrganizationRow(id string, document []byte, scopeID string) (OrganizationHomePayload, error) {
	var home OrganizationHomePayload
	if err := json.Unmarshal(document, &home); err != nil {
		return OrganizationHomePayload{}, fmt.Errorf("decode organization %s: %w", id, err)
	}
	home.ID = id
	home.QAFixtureScopeID = scopeID
	return home, nil
}

func (p *postgresOrganizationRepository) List(ctx context.Context) ([]OrganizationSummary, error) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT id::text, document, qa_fixture_scope_id FROM organizations
		WHERE tenant_id=$1
		ORDER BY name COLLATE "C", id`, organizationTenantID)
	if err != nil {
		return nil, wrapOrganizationStorageError("list", err)
	}
	defer rows.Close()
	summaries := make([]OrganizationSummary, 0)
	for rows.Next() {
		var id, scopeID string
		var document []byte
		if err := rows.Scan(&id, &document, &scopeID); err != nil {
			return nil, wrapOrganizationStorageError("list scan", err)
		}
		home, err := decodeOrganizationRow(id, document, scopeID)
		if err != nil {
			return nil, wrapOrganizationStorageError("list decode", err)
		}
		summaries = append(summaries, home.OrganizationSummary)
	}
	if err := rows.Err(); err != nil {
		return nil, wrapOrganizationStorageError("list rows", err)
	}
	return summaries, nil
}

func (p *postgresOrganizationRepository) Insert(ctx context.Context, home OrganizationHomePayload) error {
	id, ok := organizationUUID(home.ID)
	if !ok {
		return fmt.Errorf("%w: organization id must be a UUID", ErrOrganizationStoreUnavailable)
	}
	home.ID = id
	document, err := encodeOrganizationDocument(home)
	if err != nil {
		return wrapOrganizationStorageError("insert", err)
	}
	_, err = p.db.ExecContext(ctx, `
		INSERT INTO organizations
			(id, tenant_id, name, purpose, template_id, qa_fixture_scope_id, document)
		VALUES ($1::uuid, $2, $3, $4, $5, $6, $7::jsonb)`,
		id, organizationTenantID, home.Name, home.Purpose, home.TemplateID, home.QAFixtureScopeID, document)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrOrganizationConflict
	}
	return wrapOrganizationStorageError("insert", err)
}

func (p *postgresOrganizationRepository) Get(ctx context.Context, id string) (OrganizationHomePayload, error) {
	id, ok := organizationUUID(id)
	if !ok {
		return OrganizationHomePayload{}, ErrOrganizationNotFound
	}
	var document []byte
	var scopeID string
	err := p.db.QueryRowContext(ctx, `
		SELECT document, qa_fixture_scope_id FROM organizations
		WHERE id=$1::uuid AND tenant_id=$2`, id, organizationTenantID).Scan(&document, &scopeID)
	if errors.Is(err, sql.ErrNoRows) {
		return OrganizationHomePayload{}, ErrOrganizationNotFound
	}
	if err != nil {
		return OrganizationHomePayload{}, wrapOrganizationStorageError("get", err)
	}
	home, err := decodeOrganizationRow(id, document, scopeID)
	return home, wrapOrganizationStorageError("get decode", err)
}

// Update is SELECT ... FOR UPDATE plus write-back in one transaction, so
// concurrent updates on the same id serialize and lose no write. id,
// tenant_id and qa_fixture_scope_id are never rewritten.
func (p *postgresOrganizationRepository) Update(ctx context.Context, id string, update func(OrganizationHomePayload) (OrganizationHomePayload, error)) (OrganizationHomePayload, error) {
	id, ok := organizationUUID(id)
	if !ok {
		return OrganizationHomePayload{}, ErrOrganizationNotFound
	}
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return OrganizationHomePayload{}, wrapOrganizationStorageError("update begin", err)
	}
	defer func() { _ = tx.Rollback() }()
	var document []byte
	var scopeID string
	err = tx.QueryRowContext(ctx, `
		SELECT document, qa_fixture_scope_id FROM organizations
		WHERE id=$1::uuid AND tenant_id=$2
		FOR UPDATE`, id, organizationTenantID).Scan(&document, &scopeID)
	if errors.Is(err, sql.ErrNoRows) {
		return OrganizationHomePayload{}, ErrOrganizationNotFound
	}
	if err != nil {
		return OrganizationHomePayload{}, wrapOrganizationStorageError("update lock", err)
	}
	current, err := decodeOrganizationRow(id, document, scopeID)
	if err != nil {
		return OrganizationHomePayload{}, wrapOrganizationStorageError("update decode", err)
	}
	next, err := update(current)
	if err != nil {
		return OrganizationHomePayload{}, err // deferred rollback: no write
	}
	updated := preserveOrganizationIdentity(current, next)
	encoded, err := encodeOrganizationDocument(updated)
	if err != nil {
		return OrganizationHomePayload{}, wrapOrganizationStorageError("update", err)
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE organizations
		SET name=$3, purpose=$4, template_id=$5, document=$6::jsonb, updated_at=NOW()
		WHERE id=$1::uuid AND tenant_id=$2`,
		id, organizationTenantID, updated.Name, updated.Purpose, updated.TemplateID, encoded); err != nil {
		return OrganizationHomePayload{}, wrapOrganizationStorageError("update write", err)
	}
	if err := tx.Commit(); err != nil {
		return OrganizationHomePayload{}, wrapOrganizationStorageError("update commit", err)
	}
	return updated, nil
}

func (p *postgresOrganizationRepository) Delete(ctx context.Context, id string) (bool, error) {
	id, ok := organizationUUID(id)
	if !ok {
		return false, nil
	}
	result, err := p.db.ExecContext(ctx, `
		DELETE FROM organizations WHERE id=$1::uuid AND tenant_id=$2`, id, organizationTenantID)
	if err != nil {
		return false, wrapOrganizationStorageError("delete", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, wrapOrganizationStorageError("delete", err)
	}
	return rows > 0, nil
}

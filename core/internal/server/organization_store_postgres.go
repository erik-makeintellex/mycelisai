package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const defaultOrganizationTenantID = "default"

const (
	organizationUpsertSQL = `INSERT INTO organizations (id, tenant_id, name, purpose, template_id, document, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6::jsonb, NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    purpose = EXCLUDED.purpose,
    template_id = EXCLUDED.template_id,
    document = EXCLUDED.document,
    updated_at = NOW()
WHERE organizations.tenant_id = EXCLUDED.tenant_id`
	organizationGetSQL       = `SELECT document FROM organizations WHERE id = $1 AND tenant_id = $2`
	organizationListSQL      = `SELECT document FROM organizations WHERE tenant_id = $1 ORDER BY name ASC, id ASC`
	organizationLockSQL      = `SELECT document FROM organizations WHERE id = $1 AND tenant_id = $2 FOR UPDATE`
	organizationWriteBackSQL = `UPDATE organizations
SET name = $3, purpose = $4, template_id = $5, document = $6::jsonb, updated_at = NOW()
WHERE id = $1 AND tenant_id = $2`
)

// PostgresOrganizationRepository persists organization homes in the
// `organizations` table (migration 040). The whole OrganizationHomePayload is
// stored in `document`; id/name/purpose/template_id are promoted columns.
type PostgresOrganizationRepository struct {
	db       *sql.DB
	tenantID string
}

func NewPostgresOrganizationRepository(db *sql.DB, tenantID string) *PostgresOrganizationRepository {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		tenantID = defaultOrganizationTenantID
	}
	return &PostgresOrganizationRepository{db: db, tenantID: tenantID}
}

func (p *PostgresOrganizationRepository) Backend() string { return "postgres" }

func (p *PostgresOrganizationRepository) ListOrganizations(ctx context.Context) ([]OrganizationSummary, error) {
	rows, err := p.db.QueryContext(ctx, organizationListSQL, p.tenantID)
	if err != nil {
		return nil, fmt.Errorf("list organizations: %w", err)
	}
	defer rows.Close()

	summaries := make([]OrganizationSummary, 0)
	for rows.Next() {
		var raw []byte
		if err := rows.Scan(&raw); err != nil {
			return nil, fmt.Errorf("scan organization: %w", err)
		}
		home, err := decodeOrganizationDocument(raw)
		if err != nil {
			return nil, err
		}
		summaries = append(summaries, home.OrganizationSummary)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list organizations: %w", err)
	}
	sortOrganizationSummaries(summaries)
	return summaries, nil
}

func (p *PostgresOrganizationRepository) SaveOrganization(ctx context.Context, home OrganizationHomePayload) (OrganizationHomePayload, error) {
	if strings.TrimSpace(home.ID) == "" {
		return OrganizationHomePayload{}, ErrOrganizationIDRequired
	}
	doc, err := json.Marshal(home)
	if err != nil {
		return OrganizationHomePayload{}, fmt.Errorf("encode organization %s: %w", home.ID, err)
	}
	result, err := p.db.ExecContext(ctx, organizationUpsertSQL,
		home.ID, p.tenantID, home.Name, home.Purpose, home.TemplateID, string(doc))
	if err != nil {
		return OrganizationHomePayload{}, fmt.Errorf("save organization %s: %w", home.ID, err)
	}
	if affected, err := result.RowsAffected(); err == nil && affected == 0 {
		return OrganizationHomePayload{}, fmt.Errorf("save organization %s: id belongs to another tenant", home.ID)
	}
	return home, nil
}

func (p *PostgresOrganizationRepository) GetOrganization(ctx context.Context, id string) (OrganizationHomePayload, bool, error) {
	var raw []byte
	err := p.db.QueryRowContext(ctx, organizationGetSQL, id, p.tenantID).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return OrganizationHomePayload{}, false, nil
	}
	if err != nil {
		return OrganizationHomePayload{}, false, fmt.Errorf("get organization %s: %w", id, err)
	}
	home, err := decodeOrganizationDocument(raw)
	if err != nil {
		return OrganizationHomePayload{}, false, err
	}
	return home, true, nil
}

// UpdateOrganization locks the row (SELECT ... FOR UPDATE), applies update,
// and writes the new document back inside one transaction.
func (p *PostgresOrganizationRepository) UpdateOrganization(ctx context.Context, id string, update OrganizationUpdateFunc) (OrganizationHomePayload, bool, error) {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return OrganizationHomePayload{}, false, fmt.Errorf("begin organization update %s: %w", id, err)
	}
	defer tx.Rollback() //nolint:errcheck // no-op after commit

	var raw []byte
	err = tx.QueryRowContext(ctx, organizationLockSQL, id, p.tenantID).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return OrganizationHomePayload{}, false, nil
	}
	if err != nil {
		return OrganizationHomePayload{}, false, fmt.Errorf("lock organization %s: %w", id, err)
	}
	current, err := decodeOrganizationDocument(raw)
	if err != nil {
		return OrganizationHomePayload{}, false, err
	}

	updated := update(current)
	updated.ID = id
	doc, err := json.Marshal(updated)
	if err != nil {
		return OrganizationHomePayload{}, false, fmt.Errorf("encode organization %s: %w", id, err)
	}
	if _, err := tx.ExecContext(ctx, organizationWriteBackSQL,
		id, p.tenantID, updated.Name, updated.Purpose, updated.TemplateID, string(doc)); err != nil {
		return OrganizationHomePayload{}, false, fmt.Errorf("write organization %s: %w", id, err)
	}
	if err := tx.Commit(); err != nil {
		return OrganizationHomePayload{}, false, fmt.Errorf("commit organization %s: %w", id, err)
	}
	return updated, true, nil
}

func decodeOrganizationDocument(raw []byte) (OrganizationHomePayload, error) {
	var home OrganizationHomePayload
	if err := json.Unmarshal(raw, &home); err != nil {
		return OrganizationHomePayload{}, fmt.Errorf("decode organization document: %w", err)
	}
	return home, nil
}

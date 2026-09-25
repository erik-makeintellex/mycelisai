package server

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

// operator_templates (migration 041) holds RC-004G clones. Every write
// happens inside a governed confirm transaction.

const operatorTemplatesTenant = "default"

var errOperatorTemplateExists = errors.New("operator template id already exists")

type operatorTemplateRow struct {
	ID               string
	Name             string
	SourceTemplateID string
	Document         []byte
	DocumentSHA256   string
	CreatedBy        string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// operatorTemplateWrite is one insert or update from a confirmed mutation.
type operatorTemplateWrite struct {
	ID               string
	Name             string
	SourceTemplateID string
	Document         []byte
	DocumentSHA256   string
	CreatedBy        string
	AuditEventID     string
	IntentProofID    string
}

const operatorTemplateColumns = `id, name, source_template_id, document, document_sha256, created_by, created_at, updated_at`

func scanOperatorTemplate(scanner interface{ Scan(dest ...any) error }) (*operatorTemplateRow, error) {
	var row operatorTemplateRow
	if err := scanner.Scan(&row.ID, &row.Name, &row.SourceTemplateID, &row.Document,
		&row.DocumentSHA256, &row.CreatedBy, &row.CreatedAt, &row.UpdatedAt); err != nil {
		return nil, err
	}
	return &row, nil
}

// getOperatorTemplate returns the row for id, or nil when it does not exist.
func (s *AdminServer) getOperatorTemplate(ctx context.Context, id string) (*operatorTemplateRow, error) {
	db := s.getDB()
	if db == nil {
		return nil, errDBUnavailable
	}
	row, err := scanOperatorTemplate(db.QueryRowContext(ctx,
		`SELECT `+operatorTemplateColumns+` FROM operator_templates WHERE id = $1 AND tenant_id = $2`,
		id, operatorTemplatesTenant))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return row, err
}

// listOperatorTemplates returns every operator template, oldest first.
func (s *AdminServer) listOperatorTemplates(ctx context.Context) ([]operatorTemplateRow, error) {
	db := s.getDB()
	if db == nil {
		return nil, errDBUnavailable
	}
	rows, err := db.QueryContext(ctx,
		`SELECT `+operatorTemplateColumns+` FROM operator_templates WHERE tenant_id = $1 ORDER BY created_at, id`,
		operatorTemplatesTenant)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []operatorTemplateRow
	for rows.Next() {
		row, err := scanOperatorTemplate(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *row)
	}
	return out, rows.Err()
}

// insertOperatorTemplateTx stores a clone. It returns
// errOperatorTemplateExists when the id is taken.
func insertOperatorTemplateTx(ctx context.Context, tx *sql.Tx, w operatorTemplateWrite) error {
	res, err := tx.ExecContext(ctx, `
		INSERT INTO operator_templates (
			id, tenant_id, name, source_template_id, document, document_sha256,
			created_by, created_audit_event_id, updated_audit_event_id, intent_proof_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8, $9)
		ON CONFLICT (id) DO NOTHING`,
		w.ID, operatorTemplatesTenant, w.Name, w.SourceTemplateID, w.Document, w.DocumentSHA256,
		w.CreatedBy, parseAuditUUID(w.AuditEventID), parseAuditUUID(w.IntentProofID))
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return errOperatorTemplateExists
	}
	return nil
}

// updateOperatorTemplateTx replaces a template document. It returns
// sql.ErrNoRows when the template no longer exists.
func updateOperatorTemplateTx(ctx context.Context, tx *sql.Tx, w operatorTemplateWrite) error {
	res, err := tx.ExecContext(ctx, `
		UPDATE operator_templates
		SET name = $3, document = $4, document_sha256 = $5,
		    updated_audit_event_id = $6, intent_proof_id = $7, updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2`,
		w.ID, operatorTemplatesTenant, w.Name, w.Document, w.DocumentSHA256,
		parseAuditUUID(w.AuditEventID), parseAuditUUID(w.IntentProofID))
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// isOperatorTemplateTableMissing detects a database without migration 041.
func isOperatorTemplateTableMissing(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "42P01") || strings.Contains(msg, `relation "operator_templates" does not exist`)
}

// operatorTemplateStoreError converts a store failure into a config error.
func operatorTemplateStoreError(err error) error {
	if isOperatorTemplateTableMissing(err) {
		return configUnavailable("operator template storage is not migrated; run `uv run inv compose.migrate` and restart Core")
	}
	if errors.Is(err, errDBUnavailable) {
		return configUnavailable("database not configured; operator templates are unavailable")
	}
	return err
}

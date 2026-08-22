package configdocuments

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mycelis/core/pkg/protocol"
)

var (
	ErrRevisionNotFound        = errors.New("config documents: revision not found")
	ErrInvalidDocument         = errors.New("config documents: invalid document")
	ErrDocumentDisabled        = errors.New("config documents: document is disabled")
	ErrInvalidActivationAction = errors.New("config documents: activation action must be activate or rollback")
)

type ActivationAction string

const (
	ActivationActionActivate ActivationAction = "activate"
	ActivationActionRollback ActivationAction = "rollback"
)

type Store struct {
	db *sql.DB
}

type revisionQueryRower interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

type RevisionRecord struct {
	RecordID        string                  `json:"record_id"`
	TenantID        string                  `json:"tenant_id"`
	Document        protocol.ConfigDocument `json:"document"`
	Digest          string                  `json:"digest"`
	ValidationState string                  `json:"validation_state"`
	CreatedBy       string                  `json:"created_by"`
	CreatedAt       time.Time               `json:"created_at"`
}

type ListFilter struct {
	Kind       protocol.ConfigDocumentKind
	ScopeKind  protocol.ConfigDocumentScopeKind
	ScopeRef   string
	DocumentID string
	Limit      int
}

type ValidationError struct {
	Issues []protocol.ConfigDocumentValidationIssue
}

func (e *ValidationError) Error() string {
	if e == nil || len(e.Issues) == 0 {
		return ErrInvalidDocument.Error()
	}
	return fmt.Sprintf("%s: %s", ErrInvalidDocument, e.Issues[0].Code)
}

func (e *ValidationError) Unwrap() error {
	return ErrInvalidDocument
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) StoreRevision(ctx context.Context, tenantID, actorID string, document protocol.ConfigDocument) (*RevisionRecord, error) {
	if err := s.available(); err != nil {
		return nil, err
	}
	return storeRevision(ctx, s.db, tenantID, actorID, document)
}

// StoreRevisionTx stores a revision inside an existing governance transaction.
// The caller owns commit and rollback.
func (s *Store) StoreRevisionTx(ctx context.Context, tx *sql.Tx, tenantID, actorID string, document protocol.ConfigDocument) (*RevisionRecord, error) {
	if tx == nil {
		return nil, fmt.Errorf("config documents: transaction is required")
	}
	return storeRevision(ctx, tx, tenantID, actorID, document)
}

func storeRevision(ctx context.Context, queryer revisionQueryRower, tenantID, actorID string, document protocol.ConfigDocument) (*RevisionRecord, error) {
	tenantID, err := requiredValue("tenant_id", tenantID)
	if err != nil {
		return nil, err
	}
	actorID, err = requiredValue("actor_id", actorID)
	if err != nil {
		return nil, err
	}
	if issues := protocol.ValidateConfigDocument(document); len(issues) != 0 {
		return nil, &ValidationError{Issues: issues}
	}
	digest, err := protocol.CanonicalConfigDocumentDigest(document)
	if err != nil {
		return nil, fmt.Errorf("%w: compute digest: %v", ErrInvalidDocument, err)
	}
	secretRefs, err := json.Marshal(document.Metadata.SecretRefs)
	if err != nil {
		return nil, fmt.Errorf("config documents: marshal secret refs: %w", err)
	}
	governance, err := json.Marshal(document.Metadata.Governance)
	if err != nil {
		return nil, fmt.Errorf("config documents: marshal governance: %w", err)
	}

	row := queryer.QueryRowContext(ctx, `
		INSERT INTO config_documents
			(tenant_id, document_id, api_version, kind, name, version, owner_id,
			 scope_kind, scope_ref, enabled, source_kind, source_ref, secret_refs,
			 governance, spec, digest, validation_state, created_by)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12,
			 $13::jsonb, $14::jsonb, $15::jsonb, $16, 'valid', $17)
		RETURNING `+revisionColumns+`
	`,
		tenantID,
		document.Metadata.ID,
		document.APIVersion,
		string(document.Kind),
		document.Metadata.Name,
		document.Metadata.Version,
		document.Metadata.OwnerID,
		string(document.Metadata.Scope.Kind),
		document.Metadata.Scope.Ref,
		document.Metadata.Enabled,
		string(document.Metadata.Source.Kind),
		document.Metadata.Source.Ref,
		string(secretRefs),
		string(governance),
		string(document.Spec),
		digest,
		actorID,
	)
	record, err := scanRevision(row)
	if err != nil {
		return nil, fmt.Errorf("config documents: store revision: %w", err)
	}
	return record, nil
}

func (s *Store) GetRevision(ctx context.Context, tenantID, recordID string) (*RevisionRecord, error) {
	tenantID, err := requiredValue("tenant_id", tenantID)
	if err != nil {
		return nil, err
	}
	recordID, err = requiredValue("record_id", recordID)
	if err != nil {
		return nil, err
	}
	if err := s.available(); err != nil {
		return nil, err
	}

	record, err := scanRevision(s.db.QueryRowContext(ctx, `
		SELECT `+revisionColumns+`
		FROM config_documents
		WHERE tenant_id = $1 AND record_id = $2::uuid
	`, tenantID, recordID))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrRevisionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("config documents: get revision: %w", err)
	}
	return record, nil
}

func (s *Store) List(ctx context.Context, tenantID string, filter ListFilter) ([]RevisionRecord, error) {
	tenantID, err := requiredValue("tenant_id", tenantID)
	if err != nil {
		return nil, err
	}
	if err := s.available(); err != nil {
		return nil, err
	}

	conditions := []string{"tenant_id = $1"}
	args := []any{tenantID}
	addFilter := func(column string, value any) {
		args = append(args, value)
		conditions = append(conditions, fmt.Sprintf("%s = $%d", column, len(args)))
	}
	if filter.Kind != "" {
		addFilter("kind", string(filter.Kind))
	}
	if filter.ScopeKind != "" {
		addFilter("scope_kind", string(filter.ScopeKind))
	}
	if scopeRef := strings.TrimSpace(filter.ScopeRef); scopeRef != "" {
		addFilter("scope_ref", scopeRef)
	}
	if documentID := strings.TrimSpace(filter.DocumentID); documentID != "" {
		addFilter("document_id", documentID)
	}
	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	args = append(args, limit)

	rows, err := s.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT %s
		FROM config_documents
		WHERE %s
		ORDER BY created_at DESC, record_id DESC
		LIMIT $%d
	`, revisionColumns, strings.Join(conditions, " AND "), len(args)), args...)
	if err != nil {
		return nil, fmt.Errorf("config documents: list revisions: %w", err)
	}
	defer rows.Close()

	records := make([]RevisionRecord, 0)
	for rows.Next() {
		record, err := scanRevision(rows)
		if err != nil {
			return nil, fmt.Errorf("config documents: list revisions: %w", err)
		}
		records = append(records, *record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("config documents: list revisions: %w", err)
	}
	return records, nil
}

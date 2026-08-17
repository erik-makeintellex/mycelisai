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

type ActivationResult struct {
	TenantID     string                       `json:"tenant_id"`
	Kind         protocol.ConfigDocumentKind  `json:"kind"`
	DocumentID   string                       `json:"document_id"`
	Scope        protocol.ConfigDocumentScope `json:"scope"`
	FromRecordID string                       `json:"from_record_id,omitempty"`
	ToRecordID   string                       `json:"to_record_id"`
	Action       ActivationAction             `json:"action"`
	ActorID      string                       `json:"actor_id"`
	AuditEventID string                       `json:"audit_event_id,omitempty"`
	ActivatedAt  time.Time                    `json:"activated_at"`
	Revision     RevisionRecord               `json:"revision"`
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
	if err := s.available(); err != nil {
		return nil, err
	}

	secretRefs, err := json.Marshal(document.Metadata.SecretRefs)
	if err != nil {
		return nil, fmt.Errorf("config documents: marshal secret refs: %w", err)
	}
	governance, err := json.Marshal(document.Metadata.Governance)
	if err != nil {
		return nil, fmt.Errorf("config documents: marshal governance: %w", err)
	}

	row := s.db.QueryRowContext(ctx, `
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

func (s *Store) ActivateRevision(ctx context.Context, tenantID, recordID, actorID, auditEventID string, action ActivationAction) (*ActivationResult, error) {
	tenantID, err := requiredValue("tenant_id", tenantID)
	if err != nil {
		return nil, err
	}
	recordID, err = requiredValue("record_id", recordID)
	if err != nil {
		return nil, err
	}
	actorID, err = requiredValue("actor_id", actorID)
	if err != nil {
		return nil, err
	}
	if action != ActivationActionActivate && action != ActivationActionRollback {
		return nil, ErrInvalidActivationAction
	}
	if err := s.available(); err != nil {
		return nil, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("config documents: begin activation: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	revision, err := scanRevision(tx.QueryRowContext(ctx, `
		SELECT `+revisionColumns+`
		FROM config_documents
		WHERE tenant_id = $1 AND record_id = $2::uuid
		FOR UPDATE
	`, tenantID, recordID))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrRevisionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("config documents: lock revision: %w", err)
	}
	if err := validateStoredRevision(*revision); err != nil {
		return nil, err
	}

	scope := revision.Document.Metadata.Scope
	var fromRecordID string
	err = tx.QueryRowContext(ctx, `
		SELECT config_document_record_id::text
		FROM config_document_activations
		WHERE tenant_id = $1 AND kind = $2 AND document_id = $3 AND scope_kind = $4 AND scope_ref = $5
		FOR UPDATE
	`, tenantID, string(revision.Document.Kind), revision.Document.Metadata.ID, string(scope.Kind), scope.Ref).Scan(&fromRecordID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("config documents: lock activation: %w", err)
	}

	var activatedAt time.Time
	err = tx.QueryRowContext(ctx, `
		INSERT INTO config_document_activations
			(tenant_id, kind, document_id, scope_kind, scope_ref, config_document_record_id, activated_by)
		VALUES ($1, $2, $3, $4, $5, $6::uuid, $7)
		ON CONFLICT (tenant_id, kind, document_id, scope_kind, scope_ref) DO UPDATE SET
			config_document_record_id = EXCLUDED.config_document_record_id,
			activated_by = EXCLUDED.activated_by,
			activated_at = NOW()
		RETURNING activated_at
	`, tenantID, string(revision.Document.Kind), revision.Document.Metadata.ID, string(scope.Kind), scope.Ref, revision.RecordID, actorID).Scan(&activatedAt)
	if err != nil {
		return nil, fmt.Errorf("config documents: update activation: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO config_document_activation_history
			(tenant_id, kind, document_id, scope_kind, scope_ref, from_record_id,
			 to_record_id, action, actor_id, audit_event_id)
		VALUES
			($1, $2, $3, $4, $5, NULLIF($6, '')::uuid, $7::uuid, $8, $9,
			 NULLIF($10, '')::uuid)
	`, tenantID, string(revision.Document.Kind), revision.Document.Metadata.ID, string(scope.Kind), scope.Ref, fromRecordID,
		revision.RecordID, string(action), actorID, strings.TrimSpace(auditEventID))
	if err != nil {
		return nil, fmt.Errorf("config documents: append activation history: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("config documents: commit activation: %w", err)
	}

	return &ActivationResult{
		TenantID:     tenantID,
		Kind:         revision.Document.Kind,
		DocumentID:   revision.Document.Metadata.ID,
		Scope:        scope,
		FromRecordID: fromRecordID,
		ToRecordID:   revision.RecordID,
		Action:       action,
		ActorID:      actorID,
		AuditEventID: strings.TrimSpace(auditEventID),
		ActivatedAt:  activatedAt,
		Revision:     *revision,
	}, nil
}

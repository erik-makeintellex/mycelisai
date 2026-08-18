package configdocuments

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (s *Store) ActivateRevision(ctx context.Context, tenantID, recordID, actorID, auditEventID string, action ActivationAction) (*ActivationResult, error) {
	if err := s.available(); err != nil {
		return nil, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("config documents: begin activation: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	result, err := activateRevisionTx(ctx, tx, tenantID, recordID, actorID, auditEventID, action)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("config documents: commit activation: %w", err)
	}
	return result, nil
}

// ActivateRevisionTx changes activation inside an existing governance
// transaction. The caller owns commit and rollback.
func (s *Store) ActivateRevisionTx(ctx context.Context, tx *sql.Tx, tenantID, recordID, actorID, auditEventID string, action ActivationAction) (*ActivationResult, error) {
	if tx == nil {
		return nil, fmt.Errorf("config documents: transaction is required")
	}
	return activateRevisionTx(ctx, tx, tenantID, recordID, actorID, auditEventID, action)
}

func activateRevisionTx(ctx context.Context, tx *sql.Tx, tenantID, recordID, actorID, auditEventID string, action ActivationAction) (*ActivationResult, error) {
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

	historyID := uuid.NewString()
	_, err = tx.ExecContext(ctx, `
		INSERT INTO config_document_activation_history
			(id, tenant_id, kind, document_id, scope_kind, scope_ref, from_record_id,
			 to_record_id, action, actor_id, audit_event_id)
		VALUES
			($1::uuid, $2, $3, $4, $5, $6, NULLIF($7, '')::uuid, $8::uuid, $9, $10,
			 NULLIF($11, '')::uuid)
	`, historyID, tenantID, string(revision.Document.Kind), revision.Document.Metadata.ID, string(scope.Kind), scope.Ref, fromRecordID,
		revision.RecordID, string(action), actorID, strings.TrimSpace(auditEventID))
	if err != nil {
		return nil, fmt.Errorf("config documents: append activation history: %w", err)
	}
	return &ActivationResult{
		HistoryID:    historyID,
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

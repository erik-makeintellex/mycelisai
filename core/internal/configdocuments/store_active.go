package configdocuments

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mycelis/core/pkg/protocol"
)

type ActivationResult struct {
	HistoryID    string                       `json:"history_id"`
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

const activeRevisionColumns = `
	document.record_id::text, document.tenant_id, document.document_id,
	document.api_version, document.kind, document.name, document.version,
	document.owner_id, document.scope_kind, document.scope_ref, document.enabled,
	document.source_kind, document.source_ref, document.secret_refs::text,
	document.governance::text, document.spec::text, document.digest,
	document.validation_state, document.created_by, document.created_at`

// GetActiveRevision returns the revision currently activated for one exact
// document identity and scope.
func (s *Store) GetActiveRevision(
	ctx context.Context,
	tenantID string,
	kind protocol.ConfigDocumentKind,
	documentID string,
	scope protocol.ConfigDocumentScope,
) (*RevisionRecord, error) {
	tenantID, err := requiredValue("tenant_id", tenantID)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(string(kind)) == "" {
		return nil, fmt.Errorf("config documents: kind is required")
	}
	documentID, err = requiredValue("document_id", documentID)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(string(scope.Kind)) == "" {
		return nil, fmt.Errorf("config documents: scope kind is required")
	}
	if err := s.available(); err != nil {
		return nil, err
	}

	record, err := scanRevision(s.db.QueryRowContext(ctx, `
		SELECT `+activeRevisionColumns+`
		FROM config_document_activations activation
		JOIN config_documents document
		  ON document.record_id = activation.config_document_record_id
		 AND document.tenant_id = activation.tenant_id
		 AND document.kind = activation.kind
		 AND document.document_id = activation.document_id
		 AND document.scope_kind = activation.scope_kind
		 AND document.scope_ref = activation.scope_ref
		WHERE activation.tenant_id = $1
		  AND activation.kind = $2
		  AND activation.document_id = $3
		  AND activation.scope_kind = $4
		  AND activation.scope_ref = $5
	`, tenantID, string(kind), documentID, string(scope.Kind), scope.Ref))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrRevisionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("config documents: get active revision: %w", err)
	}
	if err := validateStoredRevision(*record); err != nil {
		return nil, err
	}
	return record, nil
}

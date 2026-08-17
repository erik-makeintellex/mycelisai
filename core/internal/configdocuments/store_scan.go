package configdocuments

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

const revisionColumns = `
	record_id::text, tenant_id, document_id, api_version, kind, name, version,
	owner_id, scope_kind, scope_ref, enabled, source_kind, source_ref,
	secret_refs::text, governance::text, spec::text, digest, validation_state,
	created_by, created_at`

type revisionScanner interface {
	Scan(dest ...any) error
}

func scanRevision(scanner revisionScanner) (*RevisionRecord, error) {
	var record RevisionRecord
	var documentID, kind, scopeKind, sourceKind string
	var secretRefs, governance, spec string
	err := scanner.Scan(
		&record.RecordID,
		&record.TenantID,
		&documentID,
		&record.Document.APIVersion,
		&kind,
		&record.Document.Metadata.Name,
		&record.Document.Metadata.Version,
		&record.Document.Metadata.OwnerID,
		&scopeKind,
		&record.Document.Metadata.Scope.Ref,
		&record.Document.Metadata.Enabled,
		&sourceKind,
		&record.Document.Metadata.Source.Ref,
		&secretRefs,
		&governance,
		&spec,
		&record.Digest,
		&record.ValidationState,
		&record.CreatedBy,
		&record.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	record.Document.Kind = protocol.ConfigDocumentKind(kind)
	record.Document.Metadata.ID = documentID
	record.Document.Metadata.Scope.Kind = protocol.ConfigDocumentScopeKind(scopeKind)
	record.Document.Metadata.Source.Kind = protocol.ConfigDocumentSourceKind(sourceKind)
	if err := json.Unmarshal([]byte(secretRefs), &record.Document.Metadata.SecretRefs); err != nil {
		return nil, fmt.Errorf("decode secret refs: %w", err)
	}
	if err := json.Unmarshal([]byte(governance), &record.Document.Metadata.Governance); err != nil {
		return nil, fmt.Errorf("decode governance: %w", err)
	}
	if !json.Valid([]byte(spec)) {
		return nil, fmt.Errorf("decode spec: invalid JSON")
	}
	record.Document.Spec = append(json.RawMessage(nil), spec...)
	return &record, nil
}

func validateStoredRevision(record RevisionRecord) error {
	if record.ValidationState != "valid" {
		return fmt.Errorf("%w: validation state is %q", ErrInvalidDocument, record.ValidationState)
	}
	if issues := protocol.ValidateConfigDocument(record.Document); len(issues) != 0 {
		return &ValidationError{Issues: issues}
	}
	if !record.Document.Metadata.Enabled {
		return ErrDocumentDisabled
	}
	digest, err := protocol.CanonicalConfigDocumentDigest(record.Document)
	if err != nil {
		return fmt.Errorf("%w: compute stored digest: %v", ErrInvalidDocument, err)
	}
	if digest != record.Digest {
		return fmt.Errorf("%w: digest mismatch", ErrInvalidDocument)
	}
	return nil
}

func (s *Store) available() error {
	if s == nil || s.db == nil {
		return fmt.Errorf("config documents: database not available")
	}
	return nil
}

func requiredValue(name, value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("config documents: %s is required", name)
	}
	return value, nil
}

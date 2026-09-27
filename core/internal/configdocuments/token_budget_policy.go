package configdocuments

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/mycelis/core/pkg/protocol"
)

// ErrTokenBudgetPolicyReserved refuses TokenBudgetPolicy writes and
// activations through the generic config-document paths (HTTP and Soma
// tools). Budgets change only through /api/v1/cognitive/budgets/overrides,
// which requires root admin + cognitive:write and audits every change.
var ErrTokenBudgetPolicyReserved = errors.New("config documents: token budget policy changes only through /api/v1/cognitive/budgets/overrides")

// CompileTokenBudgetPolicyDocument validates and strictly decodes a
// TokenBudgetPolicy envelope without storing or activating it.
func CompileTokenBudgetPolicyDocument(document protocol.ConfigDocument) (protocol.TokenBudgetPolicySpec, error) {
	if issues := protocol.ValidateConfigDocument(document); len(issues) != 0 {
		return protocol.TokenBudgetPolicySpec{}, &ValidationError{Issues: issues}
	}
	if document.Kind != protocol.ConfigDocumentKindTokenBudgetPolicy {
		return protocol.TokenBudgetPolicySpec{}, fmt.Errorf("config document kind %q is not a TokenBudgetPolicy", document.Kind)
	}
	return protocol.DecodeTokenBudgetPolicySpec(document.Spec)
}

func guardTokenBudgetStore(actorID string, document protocol.ConfigDocument) error {
	if err := guardPublicActor(actorID); err != nil {
		return err
	}
	if document.Kind != protocol.ConfigDocumentKindTokenBudgetPolicy {
		return fmt.Errorf("config documents: kind %q is not a TokenBudgetPolicy", document.Kind)
	}
	if issues := builtInStoreIssues(document); len(issues) != 0 {
		return &ValidationError{Issues: issues}
	}
	return nil
}

// StoreAndActivateTokenBudgetPolicyTx stores one operator TokenBudgetPolicy
// revision and activates it inside the caller's transaction (the caller owns
// commit and rollback). The guard above already refused reserved actors and
// built-in provenance, so the activation skips only the generic kind refusal.
func (s *Store) StoreAndActivateTokenBudgetPolicyTx(ctx context.Context, tx *sql.Tx, tenantID, actorID, auditEventID string, document protocol.ConfigDocument) (*ActivationResult, error) {
	if tx == nil {
		return nil, fmt.Errorf("config documents: transaction is required")
	}
	if err := guardTokenBudgetStore(actorID, document); err != nil {
		return nil, err
	}
	record, err := storeRevision(ctx, tx, tenantID, actorID, document)
	if err != nil {
		return nil, err
	}
	return activateRevisionTx(ctx, tx, tenantID, record.RecordID, actorID, auditEventID, ActivationActionActivate, true)
}

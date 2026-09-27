package configdocuments

import (
	"errors"
	"fmt"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

// BuiltInSeedActor is the only actor that may create or activate built-in
// configuration revisions. It is reserved: public store and activation paths
// reject every actor that starts with ReservedActorPrefix.
const (
	BuiltInSeedActor    = "system:bootstrap"
	ReservedActorPrefix = "system:"
)

// ErrBuiltInReserved marks an attempt by a non-bootstrap caller to activate a
// built-in revision or to act as a reserved system actor. HTTP maps it to 403.
var ErrBuiltInReserved = errors.New("config documents: built-in configuration is reserved for bootstrap seeding")

// builtInStoreIssues returns the provenance issues that make a document
// unstorable through a public path. Shape validation stays in protocol, so
// preview, dry-run and compile of a shipped built-in file remain valid.
func builtInStoreIssues(document protocol.ConfigDocument) []protocol.ConfigDocumentValidationIssue {
	issues := make([]protocol.ConfigDocumentValidationIssue, 0, 2)
	if document.Metadata.Scope.Kind == protocol.ConfigDocumentScopeBuiltIn {
		issues = append(issues, protocol.ConfigDocumentValidationIssue{
			Code:    "metadata.reserved_built_in_scope",
			Field:   "metadata.scope.kind",
			Message: "built-in scope is reserved for bootstrap seeding; copy the template to an organization, workspace or operator scope",
		})
	}
	if document.Metadata.Source.Kind == protocol.ConfigDocumentSourceBuiltIn {
		issues = append(issues, protocol.ConfigDocumentValidationIssue{
			Code:    "metadata.reserved_built_in_source",
			Field:   "metadata.source.kind",
			Message: "built-in source is reserved for bootstrap seeding; set source.kind to file, api or soma on a copy",
		})
	}
	return issues
}

// guardPublicStore rejects reserved actors and built-in provenance for every
// caller except SeedBuiltInRevisions.
func guardPublicStore(actorID string, document protocol.ConfigDocument) error {
	if err := guardPublicActor(actorID); err != nil {
		return err
	}
	if document.Kind == protocol.ConfigDocumentKindTokenBudgetPolicy {
		return ErrTokenBudgetPolicyReserved
	}
	if issues := builtInStoreIssues(document); len(issues) != 0 {
		return &ValidationError{Issues: issues}
	}
	return nil
}

func guardPublicActor(actorID string) error {
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(actorID)), ReservedActorPrefix) {
		return fmt.Errorf("%w: actor prefix %q is reserved", ErrBuiltInReserved, ReservedActorPrefix)
	}
	return nil
}

// guardPublicActivation runs after the revision row is locked, so the decision
// uses persisted provenance rather than anything the caller supplied.
func guardPublicActivation(revision RevisionRecord) error {
	metadata := revision.Document.Metadata
	if metadata.Scope.Kind == protocol.ConfigDocumentScopeBuiltIn || metadata.Source.Kind == protocol.ConfigDocumentSourceBuiltIn {
		return fmt.Errorf("%w: revision %s of %q is built-in", ErrBuiltInReserved, revision.RecordID, metadata.ID)
	}
	if revision.Document.Kind == protocol.ConfigDocumentKindTokenBudgetPolicy {
		return ErrTokenBudgetPolicyReserved
	}
	return nil
}

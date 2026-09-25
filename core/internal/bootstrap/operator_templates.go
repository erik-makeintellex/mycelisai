package bootstrap

import (
	"fmt"
	"strings"
)

// Operator templates (RC-004G) are starter bundles an operator cloned from a
// file bundle or starter. They live in the operator_templates table because
// the Helm-mounted templates directory is read-only. They follow exactly the
// starter shape and availability rules; only their storage differs.

// OperatorTemplateDocument is one stored operator template: its row id and
// its bundle document (JSON, parsed as a YAML subset).
type OperatorTemplateDocument struct {
	ID       string
	Document []byte
}

// ValidateStarterDocument parses and validates a starter document with the
// same rules the starters loader applies to files: bundle validation, a
// starter block, delivery_defaults, and approval-policy resolution. It
// returns the normalized bundle.
func ValidateStarterDocument(data []byte, policyGroups map[string]struct{}) (*TemplateBundle, error) {
	bundle, err := ParseTemplateBundle(data)
	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}
	if err := bundle.Validate(); err != nil {
		return nil, err
	}
	if err := checkStarterBlocks(bundle, policyGroups); err != nil {
		return nil, err
	}
	return bundle, nil
}

// LoadOperatorStarterEntries projects stored operator templates into starter
// entries. reservedIDs holds ids used by file bundles and file starters; an
// operator template that reuses one, or whose document id differs from its
// row id, is unavailable rather than dropped.
func LoadOperatorStarterEntries(docs []OperatorTemplateDocument, reservedIDs, policyGroups map[string]struct{}) []*StarterEntry {
	entries := make([]*StarterEntry, 0, len(docs))
	seen := make(map[string]struct{}, len(docs))
	for _, doc := range docs {
		rowID := strings.TrimSpace(doc.ID)
		label := "operator template " + rowID
		entry := starterEntryFromDocument(label, rowID, doc.Document, policyGroups)
		if entry.Available {
			if entry.ID != rowID {
				entry.markUnavailable(fmt.Errorf("%s: document id %q does not match the stored id", label, entry.ID))
			} else if _, clash := reservedIDs[rowID]; clash {
				entry.markUnavailable(fmt.Errorf("id %q is already used by a template file", rowID))
			} else if _, dup := seen[rowID]; dup {
				entry.markUnavailable(fmt.Errorf("duplicate operator template id %q", rowID))
			}
		}
		entry.ID = rowID
		seen[rowID] = struct{}{}
		entries = append(entries, entry)
	}
	return entries
}

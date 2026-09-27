package configdocuments

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

const (
	// BuiltInSeedTenant is the only tenant the bootstrap seeder writes.
	BuiltInSeedTenant = "default"
	// BuiltInSeedSourcePrefix is the repository path every seeded file must
	// declare as its source.ref, followed by its own file name.
	BuiltInSeedSourcePrefix = "core/config/documents/templates/"
	// builtInSeedLockKey serializes concurrent Core boots on one database.
	builtInSeedLockKey int64 = 0x6d79635f73656564
)

// BuiltInSeedDocument is one validated shipped file.
type BuiltInSeedDocument struct {
	File     string
	Document protocol.ConfigDocument
	Digest   string
}

// BuiltInSeedResult reports what one seeding pass wrote.
type BuiltInSeedResult struct {
	Documents   int
	Inserted    int
	Reused      int
	Activated   int
	Unchanged   int
	DirectoryOK bool
}

// LoadBuiltInSeedDirectory reads and validates every top-level *.yaml file in
// dir. It never writes. A missing directory returns (nil, false, nil). Any
// invalid file fails the whole batch.
func LoadBuiltInSeedDirectory(dir string) ([]BuiltInSeedDocument, bool, error) {
	info, err := os.Stat(dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("built-in seed: stat %s: %w", dir, err)
	}
	if !info.IsDir() {
		return nil, false, fmt.Errorf("built-in seed: %s is not a directory", dir)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, false, fmt.Errorf("built-in seed: read %s: %w", dir, err)
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.Type().IsRegular() && strings.HasSuffix(entry.Name(), ".yaml") {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)

	documents := make([]BuiltInSeedDocument, 0, len(names))
	seen := make(map[string]string, len(names))
	for _, name := range names {
		seed, err := loadBuiltInSeedFile(dir, name)
		if err != nil {
			return nil, true, err
		}
		if previous, exists := seen[seed.Document.Metadata.ID]; exists {
			return nil, true, fmt.Errorf("built-in seed %s: duplicate id %q (also in %s)", name, seed.Document.Metadata.ID, previous)
		}
		seen[seed.Document.Metadata.ID] = name
		documents = append(documents, seed)
	}
	return documents, true, nil
}

func loadBuiltInSeedFile(dir, name string) (BuiltInSeedDocument, error) {
	raw, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		return BuiltInSeedDocument{}, fmt.Errorf("built-in seed %s: read: %w", name, err)
	}
	document, err := ParseDocument(raw, "yaml")
	if err != nil {
		return BuiltInSeedDocument{}, fmt.Errorf("built-in seed %s: %w", name, err)
	}
	metadata := document.Metadata
	switch {
	case document.Kind != protocol.ConfigDocumentKindOutcomeTemplate &&
		(document.Kind != protocol.ConfigDocumentKindTokenBudgetPolicy || metadata.ID != protocol.TokenBudgetDefaultsDocumentID):
		// B1: the token budget defaults document is the one non-template seed.
		return BuiltInSeedDocument{}, fmt.Errorf("built-in seed %s: kind %q is not seedable (OutcomeTemplate or the token-budget-defaults TokenBudgetPolicy only)", name, document.Kind)
	case metadata.Scope.Kind != protocol.ConfigDocumentScopeBuiltIn || metadata.Scope.Ref != "":
		return BuiltInSeedDocument{}, fmt.Errorf("built-in seed %s: scope must be built_in with an empty ref", name)
	case metadata.Source.Kind != protocol.ConfigDocumentSourceBuiltIn || metadata.Source.Ref != BuiltInSeedSourcePrefix+name:
		return BuiltInSeedDocument{}, fmt.Errorf("built-in seed %s: source must be built_in with ref %q", name, BuiltInSeedSourcePrefix+name)
	case metadata.ID != strings.TrimSuffix(name, ".yaml"):
		return BuiltInSeedDocument{}, fmt.Errorf("built-in seed %s: metadata.id %q must equal the file name", name, metadata.ID)
	}
	if issues := protocol.ValidateConfigDocument(document); len(issues) != 0 {
		return BuiltInSeedDocument{}, fmt.Errorf("built-in seed %s: %w", name, &ValidationError{Issues: issues})
	}
	if _, err := CompileDocument(document, protocol.MinimumSufficientBrief{}, protocol.MinimumSufficientBrief{}); err != nil {
		return BuiltInSeedDocument{}, fmt.Errorf("built-in seed %s: compile: %w", name, err)
	}
	digest, err := protocol.CanonicalConfigDocumentDigest(document)
	if err != nil {
		return BuiltInSeedDocument{}, fmt.Errorf("built-in seed %s: digest: %w", name, err)
	}
	return BuiltInSeedDocument{File: name, Document: document, Digest: digest}, nil
}

// SeedBuiltInRevisions is the only writer of built-in scope or source. It
// validates dir, then stores and activates every file at (built_in, ”) in one
// transaction as BuiltInSeedActor. It is idempotent by (id, version, digest):
// an unchanged directory writes nothing. It never reads or writes operator,
// workspace or organization rows or activations. An architecture test limits
// callers to core/cmd/server.
func (s *Store) SeedBuiltInRevisions(ctx context.Context, dir string) (*BuiltInSeedResult, error) {
	documents, present, err := LoadBuiltInSeedDirectory(dir)
	if err != nil {
		return nil, err
	}
	result := &BuiltInSeedResult{Documents: len(documents), DirectoryOK: present}
	if len(documents) == 0 {
		return result, nil
	}
	if err := s.available(); err != nil {
		return nil, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("built-in seed: begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, builtInSeedLockKey); err != nil {
		return nil, fmt.Errorf("built-in seed: lock: %w", err)
	}
	if err := assertBuiltInProvenance(ctx, tx); err != nil {
		return nil, err
	}
	for _, seed := range documents {
		if err := seedOneBuiltIn(ctx, tx, seed, result); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("built-in seed: commit: %w", err)
	}
	return result, nil
}

// assertBuiltInProvenance refuses to seed over built-in rows that bootstrap
// did not write. The error names only document and record ids.
func assertBuiltInProvenance(ctx context.Context, tx *sql.Tx) error {
	var documentID, recordID string
	err := tx.QueryRowContext(ctx, `
		SELECT document_id, record_id::text
		FROM config_documents
		WHERE tenant_id = $1
		  AND (scope_kind = 'built_in' OR source_kind = 'built_in')
		  AND created_by <> $2
		ORDER BY created_at, record_id
		LIMIT 1
	`, BuiltInSeedTenant, BuiltInSeedActor).Scan(&documentID, &recordID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("built-in seed: provenance check: %w", err)
	}
	return fmt.Errorf("%w: non-bootstrap built-in row exists (document %q, record %s); resolve it before starting Core",
		ErrBuiltInReserved, documentID, recordID)
}

func seedOneBuiltIn(ctx context.Context, tx *sql.Tx, seed BuiltInSeedDocument, result *BuiltInSeedResult) error {
	metadata := seed.Document.Metadata
	var recordID, digest string
	err := tx.QueryRowContext(ctx, `
		SELECT record_id::text, digest
		FROM config_documents
		WHERE tenant_id = $1 AND document_id = $2 AND version = $3
		  AND scope_kind = 'built_in' AND scope_ref = ''
		ORDER BY created_at, record_id
		LIMIT 1
	`, BuiltInSeedTenant, metadata.ID, metadata.Version).Scan(&recordID, &digest)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		record, err := storeRevision(ctx, tx, BuiltInSeedTenant, BuiltInSeedActor, seed.Document)
		if err != nil {
			return fmt.Errorf("built-in seed %s: %w", seed.File, err)
		}
		recordID = record.RecordID
		result.Inserted++
	case err != nil:
		return fmt.Errorf("built-in seed %s: lookup: %w", seed.File, err)
	case digest != seed.Digest:
		return fmt.Errorf("built-in seed %s: %q version %q is already seeded with different content; bump metadata.version",
			seed.File, metadata.ID, metadata.Version)
	default:
		result.Reused++
	}

	var activeID string
	err = tx.QueryRowContext(ctx, `
		SELECT config_document_record_id::text
		FROM config_document_activations
		WHERE tenant_id = $1 AND kind = $2 AND document_id = $3
		  AND scope_kind = 'built_in' AND scope_ref = ''
	`, BuiltInSeedTenant, string(seed.Document.Kind), metadata.ID).Scan(&activeID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("built-in seed %s: activation lookup: %w", seed.File, err)
	}
	if activeID == recordID {
		result.Unchanged++
		return nil
	}
	if _, err := activateRevisionTx(ctx, tx, BuiltInSeedTenant, recordID, BuiltInSeedActor, "", ActivationActionActivate, true); err != nil {
		return fmt.Errorf("built-in seed %s: activate: %w", seed.File, err)
	}
	result.Activated++
	return nil
}

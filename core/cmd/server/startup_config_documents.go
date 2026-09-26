package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/mycelis/core/internal/configdocuments"
)

// builtInConfigDocumentDir is relative to Core's working directory, like the
// template bundle directory. The image copies config/ and Compose mounts it
// read-only.
const builtInConfigDocumentDir = "config/documents/templates"

// seedBuiltInConfigDocuments validates the shipped Outcome Templates and seeds
// them at built-in scope. Any invalid file or integrity failure stops startup.
func seedBuiltInConfigDocuments(ctx context.Context, db *sql.DB) {
	message, err := seedBuiltInConfigDocumentsFrom(ctx, db, builtInConfigDocumentDir)
	if err != nil {
		log.Fatalf("FATAL: Built-in configuration seeding failed: %v", err)
	}
	log.Println(message)
}

func seedBuiltInConfigDocumentsFrom(ctx context.Context, db *sql.DB, dir string) (string, error) {
	if db == nil {
		documents, present, err := configdocuments.LoadBuiltInSeedDirectory(dir)
		if err != nil {
			return "", err
		}
		if !present {
			return fmt.Sprintf("Built-in configuration seeding: %s not found; nothing to seed.", dir), nil
		}
		return fmt.Sprintf("WARN: Built-in configuration seeding skipped (database unavailable); %d file(s) validated.", len(documents)), nil
	}
	result, err := configdocuments.NewStore(db).SeedBuiltInRevisions(ctx, dir)
	if err != nil {
		return "", err
	}
	if !result.DirectoryOK {
		return fmt.Sprintf("Built-in configuration seeding: %s not found; nothing to seed.", dir), nil
	}
	return fmt.Sprintf(
		"Built-in configuration seeded: %d document(s), %d inserted, %d reused, %d activated, %d unchanged.",
		result.Documents, result.Inserted, result.Reused, result.Activated, result.Unchanged,
	), nil
}

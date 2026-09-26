package configdocuments

import (
	"database/sql"
	"os"
	"strings"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/mycelis/core/pkg/protocol"
)

// TestScanStoredConfigDocumentsForSecrets is an operator probe, skipped unless
// MYCELIS_CONFIG_DOCUMENT_SCAN_DSN is set. It opens a READ ONLY transaction,
// reads every config_documents row, and reports per row whether the stored
// rules still accept it (activation/rollback safety) and which paths the strict
// rules would redact on export. It prints record ids, kinds, and field paths
// only, never values. It fails when a row no longer passes the stored rules.
func TestScanStoredConfigDocumentsForSecrets(t *testing.T) {
	dsn := strings.TrimSpace(os.Getenv("MYCELIS_CONFIG_DOCUMENT_SCAN_DSN"))
	if dsn == "" {
		t.Skip("MYCELIS_CONFIG_DOCUMENT_SCAN_DSN not set")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open scan database: %v", err)
	}
	defer db.Close()
	tx, err := db.BeginTx(t.Context(), &sql.TxOptions{ReadOnly: true})
	if err != nil {
		t.Fatalf("begin read-only transaction: %v", err)
	}
	defer func() { _ = tx.Rollback() }()
	rows, err := tx.QueryContext(t.Context(), `SELECT `+revisionColumns+` FROM config_documents ORDER BY created_at, record_id`)
	if err != nil {
		t.Fatalf("query config_documents: %v", err)
	}
	defer rows.Close()

	var total, storedInvalid, strictInvalid, redacted int
	for rows.Next() {
		record, err := scanRevision(rows)
		if err != nil {
			t.Fatalf("scan row %d: %v", total+1, err)
		}
		total++
		storedIssues := protocol.ValidateConfigDocument(record.Document)
		strictIssues := protocol.ValidateNewConfigDocument(record.Document)
		_, paths, redactErr := protocol.RedactConfigDocumentForExport(record.Document)
		if len(storedIssues) != 0 {
			storedInvalid++
		}
		if len(strictIssues) != 0 {
			strictInvalid++
		}
		if len(paths) != 0 {
			redacted++
		}
		t.Logf("record=%s kind=%s document=%s state=%s stored_rules_ok=%t strict_rules_ok=%t export_error=%t redacted_paths=%v",
			record.RecordID, record.Document.Kind, record.Document.Metadata.ID, record.ValidationState,
			len(storedIssues) == 0, len(strictIssues) == 0, redactErr != nil, paths)
		for _, issue := range storedIssues {
			t.Logf("  stored-rule issue record=%s code=%s field=%s", record.RecordID, issue.Code, issue.Field)
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate config_documents: %v", err)
	}
	t.Logf("SUMMARY rows=%d stored_rules_invalid=%d strict_rules_invalid=%d export_redacted=%d", total, storedInvalid, strictInvalid, redacted)
	if storedInvalid != 0 {
		t.Errorf("%d stored rows no longer pass the stored rules; activation or rollback would fail", storedInvalid)
	}
}

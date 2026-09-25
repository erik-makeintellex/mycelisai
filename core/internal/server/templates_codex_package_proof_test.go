package server

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/mycelis/core/internal/artifacts"
)

var proofTestFiles = map[string]string{
	"implementation_task.md": "# Task\n\nDo the thing.\n",
	"proof.md":               "# Codex Task Bundle Proof\n",
}

func newProofTestSummary(t *testing.T) *packageOutputSummary {
	t.Helper()
	workspace := t.TempDir()
	t.Setenv("MYCELIS_WORKSPACE", workspace)
	base := "codex-task-packages/myc-rc-003"
	summary := &packageOutputSummary{
		PackageKind: "codex_task_bundle",
		PackageSlug: "myc-rc-003",
		PackageID:   "codex_task_bundle:myc-rc-003",
		BasePath:    base,
		ProofPath:   base + "/proof.md",
		RunID:       "run-123",
		ProofID:     "contract-456",
	}
	for _, name := range []string{"implementation_task.md", "proof.md"} {
		content := proofTestFiles[name]
		abs := filepath.Join(workspace, filepath.FromSlash(base), name)
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
		title := name
		if name == "proof.md" {
			title = "Package Proof"
		}
		summary.Files = append(summary.Files, packageOutputFileFromBytes(base+"/"+name, title, []byte(content)))
	}
	return summary
}

func beginMockTx(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *sql.Tx) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	mock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	return db, mock, tx
}

func TestPersistCodexPackageArtifactsTx_VerifiedEnvelopes(t *testing.T) {
	summary := newProofTestSummary(t)
	db, mock, tx := beginMockTx(t)
	s := &AdminServer{Artifacts: artifacts.NewService(db, "")}
	ids := []uuid.UUID{uuid.New(), uuid.New()}
	created := time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC)
	for _, id := range ids {
		mock.ExpectQuery("INSERT INTO artifacts").
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(id, created))
	}

	if err := s.persistCodexPackageArtifactsTx(context.Background(), tx, summary); err != nil {
		t.Fatalf("persist: %v", err)
	}

	proofArtifactID := ids[1].String()
	for index, file := range summary.Files {
		content := proofTestFiles[filepath.Base(file.Path)]
		env := file.Proof
		if env == nil {
			t.Fatalf("file %s has no proof envelope", file.Path)
		}
		if env.Checksum != fmt.Sprintf("%x", sha256.Sum256([]byte(content))) || env.Bytes != len(content) {
			t.Fatalf("%s checksum/bytes = %s/%d, want hash of %d bytes", file.Path, env.Checksum, env.Bytes, len(content))
		}
		if env.ArtifactID != ids[index].String() || env.ProofID != proofArtifactID {
			t.Fatalf("%s artifact/proof linkage = %s/%s", file.Path, env.ArtifactID, env.ProofID)
		}
		if env.SourceContractID != "contract-456" || env.SourceRunID != "run-123" {
			t.Fatalf("%s source linkage = %#v", file.Path, env)
		}
		if env.ExecutionStatus != "verified" || env.PathBoundaryStatus != "verified" || env.ReadbackStatus != "verified" || env.RecoveryHint != "" {
			t.Fatalf("%s statuses = %#v", file.Path, env)
		}
		if env.CreatedAt != created.Format(time.RFC3339) {
			t.Fatalf("%s created_at = %q", file.Path, env.CreatedAt)
		}
	}

	refs := outputRefsFromPackageSummary(summary)
	if refs[0].ProofArtifactID != proofArtifactID || refs[0].OpenURL != "/api/v1/artifacts/"+ids[0].String()+"/download" {
		t.Fatalf("primary ref linkage = %#v", refs[0])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestPersistCodexPackageArtifactsTx_DegradedWithoutArtifactService(t *testing.T) {
	summary := newProofTestSummary(t)
	s := &AdminServer{}

	if err := s.persistCodexPackageArtifactsTx(context.Background(), nil, summary); err != nil {
		t.Fatalf("persist: %v", err)
	}

	fallbackProofID := "codex_task_bundle:myc-rc-003:proof.md"
	for _, file := range summary.Files {
		env := file.Proof
		if env == nil || env.ExecutionStatus != "degraded" || env.RecoveryHint == "" || env.ArtifactID != "" {
			t.Fatalf("%s envelope = %#v, want degraded without artifact", file.Path, env)
		}
		if env.ProofID != fallbackProofID || env.CreatedAt == "" {
			t.Fatalf("%s proof_id/created_at = %q/%q", file.Path, env.ProofID, env.CreatedAt)
		}
	}
	refs := outputRefsFromPackageSummary(summary)
	if refs[0].ProofArtifactID != fallbackProofID || refs[0].OpenURL != "" {
		t.Fatalf("primary ref without artifact storage = %#v", refs[0])
	}
}

func TestPackageFileReadback_DetectsChangedAndMissingFiles(t *testing.T) {
	summary := newProofTestSummary(t)
	workspace := os.Getenv("MYCELIS_WORKSPACE")
	changed, missing := summary.Files[0], summary.Files[1]

	if err := os.WriteFile(filepath.Join(workspace, filepath.FromSlash(changed.Path)), []byte("tampered"), 0o644); err != nil {
		t.Fatalf("tamper: %v", err)
	}
	if err := os.Remove(filepath.Join(workspace, filepath.FromSlash(missing.Path))); err != nil {
		t.Fatalf("remove: %v", err)
	}

	if _, readback, hint := packageFileReadback(changed); readback != "degraded" || hint == "" {
		t.Fatalf("changed readback = %q (%q), want degraded", readback, hint)
	}
	if _, readback, _ := packageFileReadback(missing); readback != "missing" {
		t.Fatalf("missing readback = %q, want missing", readback)
	}
	escape := packageOutputFile{Path: "../outside.md"}
	if boundary, _, _ := packageFileReadback(escape); boundary != "blocked" {
		t.Fatalf("escape boundary = %q, want blocked", boundary)
	}
}

type proofEnvelopePayload struct{ want int }

func (m proofEnvelopePayload) Match(v driver.Value) bool {
	raw, ok := v.([]byte)
	if !ok {
		return false
	}
	var payload struct {
		ProofEnvelopes []map[string]any `json:"proof_envelopes"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil || len(payload.ProofEnvelopes) != m.want {
		return false
	}
	for _, env := range payload.ProofEnvelopes {
		if env["output_ref_id"] == "" || env["checksum"] == "" {
			return false
		}
	}
	return true
}

func TestInsertPackageArtifactEventTx_IncludesProofEnvelopes(t *testing.T) {
	summary := newProofTestSummary(t)
	s := &AdminServer{}
	if err := s.persistCodexPackageArtifactsTx(context.Background(), nil, summary); err != nil {
		t.Fatalf("persist: %v", err)
	}
	_, mock, tx := beginMockTx(t)
	mock.ExpectExec("INSERT INTO mission_events").
		WithArgs(sqlmock.AnyArg(), "run-123", "default", sqlmock.AnyArg(), sqlmock.AnyArg(), "admin", "governance", proofEnvelopePayload{want: 2}, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := s.insertPackageArtifactEventTx(tx, summary); err != nil {
		t.Fatalf("insert event: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

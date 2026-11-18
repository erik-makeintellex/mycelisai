package server

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/mycelis/core/internal/artifacts"
	"github.com/mycelis/core/pkg/protocol"
)

const (
	proofStatusVerified = "verified"
	proofStatusDegraded = "degraded"
	proofStatusMissing  = "missing"
)

// persistCodexPackageArtifactsTx stores every package file as an artifact row
// inside the confirm-action transaction, then attaches a proof envelope to each
// file. Without an artifact service the envelopes are still emitted as degraded
// so the operator sees why artifact linkage is absent.
func (s *AdminServer) persistCodexPackageArtifactsTx(ctx context.Context, tx *sql.Tx, summary *packageOutputSummary) error {
	if summary == nil {
		return nil
	}
	createdAt := make([]time.Time, len(summary.Files))
	persisted := s != nil && s.Artifacts != nil && tx != nil
	if persisted {
		for index := range summary.Files {
			stored, err := s.storePackageArtifactTx(ctx, tx, summary, index)
			if err != nil {
				return err
			}
			summary.Files[index].ArtifactID = stored.ID.String()
			createdAt[index] = stored.CreatedAt
		}
	}
	proofID := packageProofArtifactID(summary)
	for index := range summary.Files {
		file := &summary.Files[index]
		envelope := proofEnvelopeForPackageFile(summary, *file, proofID)
		if created := createdAt[index]; !created.IsZero() {
			envelope.CreatedAt = created.UTC().Format(time.RFC3339)
		} else {
			envelope.CreatedAt = time.Now().UTC().Format(time.RFC3339)
		}
		if !persisted {
			envelope.ExecutionStatus = proofStatusDegraded
			envelope.RecoveryHint = "Artifact storage is unavailable; the file exists in the workspace but has no artifact record or download link."
		}
		file.Proof = envelope
	}
	return nil
}

func (s *AdminServer) storePackageArtifactTx(ctx context.Context, tx *sql.Tx, summary *packageOutputSummary, index int) (*artifacts.Artifact, error) {
	file := summary.Files[index]
	content := ""
	if index < len(summary.Artifacts) {
		content = summary.Artifacts[index].Content
	}
	agentID := "codex-package"
	if strings.TrimSpace(summary.TeamID) != "" {
		agentID = strings.TrimSpace(summary.TeamID)
	}
	metadata, err := json.Marshal(map[string]any{
		"package_id":         summary.PackageID,
		"package_kind":       summary.PackageKind,
		"package_slug":       summary.PackageSlug,
		"team_id":            summary.TeamID,
		"team_name":          summary.TeamName,
		"run_id":             summary.RunID,
		"source_contract_id": summary.ProofID,
		"storage_ref":        file.Path,
		"sha256":             file.SHA256,
		"role":               packageOutputFileRole(summary.Files, index),
		"proof_path":         summary.ProofPath,
		"source_channel":     "api.intent.confirm-action",
	})
	if err != nil {
		return nil, fmt.Errorf("marshal package artifact metadata: %w", err)
	}
	stored, err := s.Artifacts.StoreTx(ctx, tx, artifacts.Artifact{
		AgentID:       agentID,
		TraceID:       summary.RunID,
		ArtifactType:  artifacts.TypeDocument,
		Title:         firstNonEmptyString(file.Title, path.Base(file.Path)),
		ContentType:   firstNonEmptyString(file.ContentType, "text/markdown"),
		Content:       content,
		FilePath:      file.Path,
		FileSizeBytes: int64(file.Bytes),
		Metadata:      metadata,
		Status:        "approved",
	})
	if err != nil {
		return nil, fmt.Errorf("persist package artifact %s: %w", file.Path, err)
	}
	return stored, nil
}

// proofEnvelopeForPackageFile builds the envelope for one package file. proof_id
// is always the proof file's artifact reference, including on the proof file
// itself; the intent contract lives in source_contract_id.
func proofEnvelopeForPackageFile(summary *packageOutputSummary, file packageOutputFile, proofID string) *protocol.OutputProofEnvelope {
	boundary, readback, hint := packageFileReadback(file)
	return &protocol.OutputProofEnvelope{
		ProofID:            proofID,
		OutputRefID:        fmt.Sprintf("%s:%s", summary.PackageID, path.Base(file.Path)),
		ArtifactID:         file.ArtifactID,
		StorageRef:         file.Path,
		SourceRunID:        summary.RunID,
		SourceContractID:   summary.ProofID,
		ExecutionStatus:    proofStatusVerified,
		PathBoundaryStatus: boundary,
		Checksum:           file.SHA256,
		ChecksumAlgorithm:  "sha256",
		Bytes:              file.Bytes,
		ContentType:        file.ContentType,
		ReadbackStatus:     readback,
		RecoveryHint:       hint,
	}
}

// packageFileReadback re-resolves the file inside the workspace boundary and
// re-hashes it so readback_status reflects what is actually on disk.
func packageFileReadback(file packageOutputFile) (boundary, readback, hint string) {
	absPath, err := workspacePathForServer(file.Path)
	if err != nil {
		return "blocked", "blocked", "Path is outside the workspace boundary."
	}
	data, err := os.ReadFile(absPath)
	if err != nil {
		return proofStatusVerified, proofStatusMissing, "File is no longer present in the workspace."
	}
	if fmt.Sprintf("%x", sha256.Sum256(data)) != file.SHA256 || len(data) != file.Bytes {
		return proofStatusVerified, proofStatusDegraded, "File changed after it was recorded; checksum no longer matches."
	}
	return proofStatusVerified, proofStatusVerified, ""
}

func packageOutputFileRole(files []packageOutputFile, index int) string {
	if index < 0 || index >= len(files) {
		return "supporting"
	}
	file := files[index]
	if isPackageProofFile(file.Path, file.Title) {
		return "proof"
	}
	if index == 0 {
		return "primary"
	}
	return "supporting"
}

// packageProofArtifactID returns the proof file's artifact ID, falling back to
// the proof output ref ID when artifact storage was unavailable.
func packageProofArtifactID(summary *packageOutputSummary) string {
	if summary == nil {
		return ""
	}
	for _, file := range summary.Files {
		if isPackageProofFile(file.Path, file.Title) && file.ArtifactID != "" {
			return file.ArtifactID
		}
	}
	return fmt.Sprintf("%s:%s", summary.PackageID, path.Base(summary.ProofPath))
}

func proofEnvelopesFromPackageSummary(summary *packageOutputSummary) []protocol.OutputProofEnvelope {
	if summary == nil {
		return nil
	}
	envelopes := make([]protocol.OutputProofEnvelope, 0, len(summary.Files))
	for _, file := range summary.Files {
		if file.Proof != nil {
			envelopes = append(envelopes, *file.Proof)
		}
	}
	return envelopes
}

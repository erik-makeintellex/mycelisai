package server

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mycelis/core/pkg/protocol"
)

type packageOutputFile struct {
	Path        string `json:"path"`
	Title       string `json:"title,omitempty"`
	ContentType string `json:"content_type,omitempty"`
	Bytes       int    `json:"bytes"`
	SHA256      string `json:"sha256"`
}

type packageOutputSummary struct {
	PackageKind string                     `json:"package_kind"`
	PackageSlug string                     `json:"package_slug"`
	BasePath    string                     `json:"base_path"`
	ProofPath   string                     `json:"proof_path"`
	RunID       string                     `json:"run_id"`
	ProofID     string                     `json:"proof_id"`
	Files       []packageOutputFile        `json:"files"`
	Artifacts   []protocol.ChatArtifactRef `json:"artifacts,omitempty"`
}

func (s *AdminServer) finalizeCodexTaskBundleTx(tx *sql.Tx, planned []protocol.PlannedToolCall, runID, proofID string) (*packageOutputSummary, error) {
	basePath, slug, ok := codexPackageBaseFromCalls(planned)
	if !ok {
		return nil, nil
	}

	files := make([]packageOutputFile, 0, len(planned))
	artifacts := make([]protocol.ChatArtifactRef, 0, len(planned)+1)
	for _, call := range planned {
		file, artifact, err := packageFileFromPlannedCall(call)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
		artifacts = append(artifacts, artifact)
	}

	proofPath := filepath.ToSlash(path.Join(basePath, "proof.md"))
	proofContent := buildCodexPackageProof(runID, proofID, basePath, files)
	proofFile, proofArtifact, err := writeCodexPackageProof(proofPath, proofContent)
	if err != nil {
		return nil, err
	}
	files = append(files, proofFile)
	artifacts = append(artifacts, proofArtifact)

	summary := &packageOutputSummary{
		PackageKind: "codex_task_bundle",
		PackageSlug: slug,
		BasePath:    basePath,
		ProofPath:   proofPath,
		RunID:       runID,
		ProofID:     proofID,
		Files:       files,
		Artifacts:   artifacts,
	}
	if err := s.insertPackageArtifactEventTx(tx, summary); err != nil {
		return nil, err
	}
	return summary, nil
}

func packageFileFromPlannedCall(call protocol.PlannedToolCall) (packageOutputFile, protocol.ChatArtifactRef, error) {
	relativePath := firstStringArgument(call.Arguments, "path")
	absPath, err := workspacePathForServer(relativePath)
	if err != nil {
		return packageOutputFile{}, protocol.ChatArtifactRef{}, err
	}
	data, err := os.ReadFile(absPath)
	if err != nil {
		return packageOutputFile{}, protocol.ChatArtifactRef{}, fmt.Errorf("read package artifact %s: %w", relativePath, err)
	}
	title := firstStringArgument(call.Arguments, "artifact_role")
	if title == "" {
		title = filepath.Base(relativePath)
	}
	return packageOutputFileFromBytes(relativePath, title, data), protocol.ChatArtifactRef{
		Type:        "document",
		Title:       title,
		ContentType: "text/markdown",
		Content:     string(data),
		SavedPath:   filepath.ToSlash(relativePath),
	}, nil
}

func writeCodexPackageProof(proofPath, proofContent string) (packageOutputFile, protocol.ChatArtifactRef, error) {
	proofAbsPath, err := workspacePathForServer(proofPath)
	if err != nil {
		return packageOutputFile{}, protocol.ChatArtifactRef{}, err
	}
	if err := os.MkdirAll(filepath.Dir(proofAbsPath), 0o755); err != nil {
		return packageOutputFile{}, protocol.ChatArtifactRef{}, fmt.Errorf("create package proof directory: %w", err)
	}
	if err := os.WriteFile(proofAbsPath, []byte(proofContent), 0o644); err != nil {
		return packageOutputFile{}, protocol.ChatArtifactRef{}, fmt.Errorf("write package proof: %w", err)
	}
	return packageOutputFileFromBytes(proofPath, "Package Proof", []byte(proofContent)), protocol.ChatArtifactRef{
		Type:        "document",
		Title:       "Package Proof",
		ContentType: "text/markdown",
		Content:     proofContent,
		SavedPath:   proofPath,
	}, nil
}

func packageOutputFileFromBytes(relativePath, title string, data []byte) packageOutputFile {
	sum := sha256.Sum256(data)
	return packageOutputFile{
		Path:        filepath.ToSlash(relativePath),
		Title:       title,
		ContentType: "text/markdown",
		Bytes:       len(data),
		SHA256:      fmt.Sprintf("%x", sum),
	}
}

func codexPackageBaseFromCalls(planned []protocol.PlannedToolCall) (string, string, bool) {
	if len(planned) < 2 {
		return "", "", false
	}
	basePath := ""
	slug := ""
	for _, call := range planned {
		if strings.TrimSpace(call.Name) != "write_file" {
			return "", "", false
		}
		if firstStringArgument(call.Arguments, "package_kind") != "codex_task_bundle" {
			return "", "", false
		}
		relativePath := filepath.ToSlash(firstStringArgument(call.Arguments, "path"))
		dir := path.Dir(relativePath)
		if dir == "." || dir == "" {
			return "", "", false
		}
		if basePath == "" {
			basePath = dir
			slug = firstStringArgument(call.Arguments, "package_slug")
			continue
		}
		if dir != basePath {
			return "", "", false
		}
	}
	if slug == "" {
		slug = path.Base(basePath)
	}
	return basePath, slug, true
}

func workspacePathForServer(rawPath string) (string, error) {
	workspace := os.Getenv("MYCELIS_WORKSPACE")
	if workspace == "" {
		workspace = "./workspace"
	}
	absWorkspace, err := filepath.Abs(workspace)
	if err != nil {
		return "", fmt.Errorf("invalid workspace path: %w", err)
	}
	normalized := strings.TrimSpace(strings.ReplaceAll(rawPath, "\\", "/"))
	normalized = path.Clean(normalized)
	normalized = strings.TrimPrefix(normalized, "./")
	if strings.HasPrefix(normalized, "workspace/") {
		normalized = strings.TrimPrefix(normalized, "workspace/")
	}
	if normalized == "." || normalized == "" {
		return "", fmt.Errorf("workspace file path is required")
	}
	target := filepath.Clean(filepath.Join(absWorkspace, filepath.FromSlash(normalized)))
	rel, err := filepath.Rel(absWorkspace, target)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("path %q escapes workspace boundary", rawPath)
	}
	return target, nil
}

func buildCodexPackageProof(runID, proofID, basePath string, files []packageOutputFile) string {
	var b strings.Builder
	b.WriteString("# Codex Task Bundle Proof\n\n")
	b.WriteString("- Package kind: `codex_task_bundle`\n")
	b.WriteString("- Base path: `" + basePath + "`\n")
	b.WriteString("- Run ID: `" + runID + "`\n")
	b.WriteString("- Intent proof ID: `" + proofID + "`\n\n")
	b.WriteString("## Files\n\n")
	for _, file := range files {
		b.WriteString("- `" + file.Path + "`")
		b.WriteString(fmt.Sprintf(" (%d bytes, sha256 `%s`)\n", file.Bytes, file.SHA256))
	}
	return b.String()
}

func (s *AdminServer) insertPackageArtifactEventTx(tx *sql.Tx, summary *packageOutputSummary) error {
	if tx == nil || summary == nil {
		return nil
	}
	payload, _ := json.Marshal(map[string]any{
		"package_kind": summary.PackageKind,
		"package_slug": summary.PackageSlug,
		"base_path":    summary.BasePath,
		"proof_path":   summary.ProofPath,
		"proof_id":     summary.ProofID,
		"files":        summary.Files,
	})
	_, err := tx.Exec(
		`INSERT INTO mission_events
			(id, run_id, tenant_id, event_type, severity, source_agent, source_team, payload, emitted_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		uuid.New().String(), summary.RunID, "default", string(protocol.EventArtifactCreated), string(protocol.SeverityInfo),
		"admin", "governance", payload, time.Now(),
	)
	return err
}

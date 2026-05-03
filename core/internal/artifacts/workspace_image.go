package artifacts

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

func (s *Service) saveImageToWorkspace(ctx context.Context, id uuid.UUID, workspaceRoot, folder, filename string) (string, error) {
	artifact, err := s.Get(ctx, id)
	if err != nil {
		return "", err
	}
	raw, err := decodeInlineImage(artifact, id)
	if err != nil {
		return "", err
	}

	saveTarget, err := newImageSaveTarget(id, artifact, workspaceRoot, folder, filename)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(saveTarget.absFolder, 0o755); err != nil {
		return "", fmt.Errorf("create save folder: %w", err)
	}
	if err := os.WriteFile(saveTarget.absPath, raw, 0o644); err != nil {
		return "", fmt.Errorf("write image file: %w", err)
	}
	if err := s.markImageSaved(ctx, id, saveTarget.relPath, int64(len(raw))); err != nil {
		return "", err
	}
	return saveTarget.relPath, nil
}

func decodeInlineImage(artifact *Artifact, id uuid.UUID) ([]byte, error) {
	if artifact.ArtifactType != TypeImage {
		return nil, fmt.Errorf("artifact %s is not an image", id)
	}
	if strings.TrimSpace(artifact.Content) == "" {
		return nil, fmt.Errorf("artifact %s has no inline image content", id)
	}
	raw, err := base64.StdEncoding.DecodeString(artifact.Content)
	if err != nil {
		return nil, fmt.Errorf("decode image content: %w", err)
	}
	return raw, nil
}

type imageSaveTarget struct {
	absFolder string
	absPath   string
	relPath   string
}

func newImageSaveTarget(id uuid.UUID, artifact *Artifact, workspaceRoot, folder, filename string) (imageSaveTarget, error) {
	if workspaceRoot == "" {
		workspaceRoot = "./workspace"
	}
	absWorkspace, err := filepath.Abs(workspaceRoot)
	if err != nil {
		return imageSaveTarget{}, fmt.Errorf("invalid workspace path: %w", err)
	}

	absFolder, err := resolveWorkspaceFolder(absWorkspace, folder)
	if err != nil {
		return imageSaveTarget{}, err
	}
	base := imageFilename(id, artifact, filename)
	absTarget := filepath.Clean(filepath.Join(absFolder, base))
	relTarget, err := filepath.Rel(absWorkspace, absTarget)
	if err != nil || strings.HasPrefix(relTarget, "..") {
		return imageSaveTarget{}, fmt.Errorf("target %q escapes workspace boundary", base)
	}

	return imageSaveTarget{
		absFolder: absFolder,
		absPath:   absTarget,
		relPath:   filepath.ToSlash(relTarget),
	}, nil
}

func resolveWorkspaceFolder(absWorkspace, folder string) (string, error) {
	folder = strings.TrimSpace(folder)
	if folder == "" {
		folder = "saved-media"
	}
	absFolder := filepath.Clean(filepath.Join(absWorkspace, folder))
	relFolder, err := filepath.Rel(absWorkspace, absFolder)
	if err != nil || strings.HasPrefix(relFolder, "..") {
		return "", fmt.Errorf("folder %q escapes workspace boundary", folder)
	}
	return absFolder, nil
}

func imageFilename(id uuid.UUID, artifact *Artifact, filename string) string {
	base := strings.TrimSpace(filename)
	if base == "" {
		base = sanitizeFilename(artifact.Title)
		if base == "" {
			base = fmt.Sprintf("image-%s", id.String()[:8])
		}
	}
	if filepath.Ext(base) == "" {
		base += extensionFromContentType(artifact.ContentType)
	}
	return base
}

func (s *Service) markImageSaved(ctx context.Context, id uuid.UUID, relPath string, fileSize int64) error {
	_, err := s.DB.ExecContext(ctx, `
		UPDATE artifacts
		SET file_path = $1,
		    file_size_bytes = $2,
		    metadata = COALESCE(metadata, '{}'::jsonb) ||
		               jsonb_build_object('saved', true, 'saved_path', $1, 'saved_at', NOW())
		WHERE id = $3
	`, relPath, fileSize, id)
	if err != nil {
		return fmt.Errorf("update saved artifact metadata: %w", err)
	}
	return nil
}

var nonFilenameChars = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

func sanitizeFilename(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	s = strings.ReplaceAll(s, " ", "-")
	s = nonFilenameChars.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-.")
	if len(s) > 80 {
		s = s[:80]
	}
	return s
}

func extensionFromContentType(contentType string) string {
	ct := strings.ToLower(strings.TrimSpace(contentType))
	switch ct {
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	default:
		return ".png"
	}
}

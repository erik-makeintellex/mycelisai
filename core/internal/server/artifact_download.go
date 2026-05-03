package server

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mycelis/core/internal/artifacts"
)

func artifactDownloadBody(artifact *artifacts.Artifact) ([]byte, bool, error) {
	if artifact == nil || strings.TrimSpace(artifact.Content) == "" {
		return nil, false, nil
	}

	content := strings.TrimSpace(artifact.Content)
	contentType := strings.ToLower(strings.TrimSpace(artifact.ContentType))
	if artifact.ArtifactType == artifacts.TypeImage || strings.HasPrefix(contentType, "image/") {
		raw, err := base64.StdEncoding.DecodeString(content)
		if err != nil {
			return nil, false, fmt.Errorf("decode artifact image content: %w", err)
		}
		return raw, true, nil
	}

	return []byte(content), true, nil
}

func resolveArtifactFilePath(artifact *artifacts.Artifact, dataDir string) (string, bool) {
	if artifact == nil || strings.TrimSpace(artifact.FilePath) == "" {
		return "", false
	}

	candidates := artifactFileCandidates(strings.TrimSpace(artifact.FilePath), dataDir)
	for _, candidate := range candidates {
		info, err := os.Stat(candidate)
		if err == nil && !info.IsDir() {
			return candidate, true
		}
	}
	return "", false
}

func artifactFileCandidates(rawPath string, dataDir string) []string {
	if filepath.IsAbs(rawPath) {
		return []string{filepath.Clean(rawPath)}
	}

	workspaceRoot := os.Getenv("MYCELIS_WORKSPACE")
	if workspaceRoot == "" {
		workspaceRoot = "./workspace"
	}
	candidates := []string{filepath.Clean(filepath.Join(workspaceRoot, filepath.FromSlash(rawPath)))}
	if strings.TrimSpace(dataDir) != "" {
		candidates = append(candidates, filepath.Clean(filepath.Join(dataDir, filepath.FromSlash(rawPath))))
	}
	return candidates
}

func downloadFilename(artifact *artifacts.Artifact) string {
	if artifact == nil {
		return "artifact"
	}
	if strings.TrimSpace(artifact.FilePath) != "" {
		name := filepath.Base(filepath.FromSlash(strings.TrimSpace(artifact.FilePath)))
		if name != "." && name != string(filepath.Separator) && name != "" {
			return name
		}
	}
	title := strings.TrimSpace(artifact.Title)
	if title == "" {
		return "artifact"
	}
	return title
}

package server

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

const (
	maxValidationDigestFiles = 2_000
	maxValidationDigestBytes = int64(128 << 20)
)

func teamWorkOutputDigest(refs []protocol.TeamOutputRef) (string, error) {
	files, err := validationOutputFiles(refs)
	if err != nil {
		return "", err
	}
	digest := sha256.New()
	var total int64
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil || !info.Mode().IsRegular() {
			return "", fmt.Errorf("validation output file is unavailable: %s", file)
		}
		total += info.Size()
		if total > maxValidationDigestBytes {
			return "", fmt.Errorf("validation output exceeds %d bytes", maxValidationDigestBytes)
		}
		if err := writeValidationDigestFile(digest, file); err != nil {
			return "", err
		}
	}
	return "sha256:" + hex.EncodeToString(digest.Sum(nil)), nil
}

func validationOutputFiles(refs []protocol.TeamOutputRef) ([]string, error) {
	seen := map[string]struct{}{}
	files := []string{}
	for _, ref := range refs {
		storage := strings.TrimSpace(ref.StorageRef)
		if storage == "" {
			continue
		}
		target, _, err := resolveWorkspacePath(storage, false)
		if err != nil {
			return nil, err
		}
		info, err := os.Stat(target)
		if err != nil {
			return nil, err
		}
		if info.IsDir() {
			err = filepath.WalkDir(target, func(file string, entry os.DirEntry, walkErr error) error {
				if walkErr != nil {
					return walkErr
				}
				if entry.Type().IsRegular() {
					seen[filepath.Clean(file)] = struct{}{}
				}
				return nil
			})
			if err != nil {
				return nil, err
			}
		} else if info.Mode().IsRegular() {
			seen[filepath.Clean(target)] = struct{}{}
		}
	}
	for file := range seen {
		files = append(files, file)
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("validation output contains no retained files")
	}
	if len(files) > maxValidationDigestFiles {
		return nil, fmt.Errorf("validation output exceeds %d files", maxValidationDigestFiles)
	}
	sort.Strings(files)
	return files, nil
}

func writeValidationDigestFile(digest hash.Hash, file string) error {
	root, err := workspaceRoot()
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(root, file)
	if err != nil || pathEscapesWorkspace(rel) {
		return fmt.Errorf("validation output escapes workspace")
	}
	_, _ = io.WriteString(digest, filepath.ToSlash(rel)+"\x00")
	handle, err := os.Open(file)
	if err != nil {
		return err
	}
	defer handle.Close()
	_, err = io.Copy(digest, handle)
	return err
}

func teamWorkValidationLaunchURL(refs []protocol.TeamOutputRef) (string, error) {
	for _, ref := range refs {
		if strings.TrimSpace(ref.Entrypoint) == "" {
			continue
		}
		openPath := teamWorkEntrypointPath(ref)
		href := workspaceFileOutputHref(openPath)
		base := strings.TrimRight(strings.TrimSpace(os.Getenv("MYCELIS_API_URL")), "/")
		if base == "" {
			base = "http://127.0.0.1:8081"
		}
		if parsed, err := url.ParseRequestURI(base + href); err == nil && parsed.IsAbs() {
			return parsed.String(), nil
		}
	}
	return "", fmt.Errorf("interactive output has no launchable entrypoint")
}

func teamWorkEntrypointPath(ref protocol.TeamOutputRef) string {
	storage := strings.Trim(path.Clean(strings.ReplaceAll(ref.StorageRef, "\\", "/")), "/")
	entrypoint := strings.Trim(path.Clean(strings.ReplaceAll(ref.Entrypoint, "\\", "/")), "/")
	if storage == "" || entrypoint == storage || strings.HasPrefix(entrypoint, storage+"/") {
		return entrypoint
	}
	return path.Join(storage, entrypoint)
}

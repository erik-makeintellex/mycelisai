package server

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

// Readback statuses shared by confirm-action and team-result proof. Only
// outputReadbackVerified lets a proof claim the output was verified.
const (
	outputReadbackVerified      = "verified"
	outputReadbackMissing       = "missing"
	outputReadbackEmpty         = "empty"
	outputReadbackUnreadable    = "unreadable"
	outputReadbackMismatch      = "mismatch"
	outputReadbackEchoOnly      = "echo_only"
	outputReadbackOutOfBounds   = "out_of_bounds"
	outputReadbackUnresolvable  = "not_resolvable"
	outputReadbackNotApplicable = "not_applicable"

	maxOutputReadbackBytes  = int64(32 << 20)
	maxOutputEchoScanBytes  = 1 << 20
	minOutputSubstanceChars = 40
	minRequestEchoChars     = 12
)

var outputMarkupPattern = regexp.MustCompile(`(?s)<[^>]*>`)

// outputReadback is Core's own evidence about a claimed workspace output:
// the checksum and size always describe the bytes on disk, never the request.
type outputReadback struct {
	Path         string
	Status       string
	PathBoundary string
	Checksum     string
	Bytes        int64
	Detail       string
	Folder       bool
}

func (r outputReadback) verified() bool {
	return r.Status == outputReadbackVerified
}

// readbackWorkspaceOutput re-reads a claimed output inside the workspace
// boundary. expectedChecksum (hex sha256, optional "sha256:" prefix) is
// compared when present; requestEchoes are the operator's request texts,
// used to reject a file that only repeats the request.
func readbackWorkspaceOutput(rawPath, expectedChecksum string, requestEchoes []string) outputReadback {
	target, rel, err := resolveWorkspaceFilePath(rawPath)
	if err != nil {
		return outputReadback{Path: strings.TrimSpace(rawPath), Status: outputReadbackOutOfBounds, PathBoundary: "failed", Detail: err.Error()}
	}
	result := outputReadback{Path: rel, PathBoundary: "verified"}
	info, err := os.Stat(target)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return result.with(outputReadbackMissing, "the output does not exist in the workspace")
	case err != nil:
		return result.with(outputReadbackUnreadable, err.Error())
	case info.IsDir():
		return readbackWorkspaceFolder(target, expectedChecksum, result)
	case !info.Mode().IsRegular():
		return result.with(outputReadbackUnreadable, "the output is not a regular file")
	case info.Size() > maxOutputReadbackBytes:
		return result.with(outputReadbackUnreadable, "the output exceeds the readback size limit")
	}
	data, err := os.ReadFile(target)
	if err != nil {
		return result.with(outputReadbackUnreadable, err.Error())
	}
	sum := sha256.Sum256(data)
	result.Checksum = hex.EncodeToString(sum[:])
	result.Bytes = int64(len(data))
	if len(strings.TrimSpace(string(data))) == 0 {
		return result.with(outputReadbackEmpty, "the output file is empty")
	}
	if expected := normalizeOutputChecksum(expectedChecksum); expected != "" && expected != result.Checksum {
		return result.with(outputReadbackMismatch, "the bytes on disk do not match the expected checksum")
	}
	if outputOnlyEchoesRequest(data, requestEchoes) {
		return result.with(outputReadbackEchoOnly, "the output only repeats the original request")
	}
	result.Status = outputReadbackVerified
	return result
}

// readbackWorkspaceFolder hashes every regular file in a retained folder with
// the same scheme as teamWorkOutputDigest, and requires real content.
func readbackWorkspaceFolder(target, expectedChecksum string, result outputReadback) outputReadback {
	result.Folder = true
	files := []string{}
	err := filepath.WalkDir(target, func(file string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Type().IsRegular() {
			files = append(files, filepath.Clean(file))
		}
		if len(files) > maxValidationDigestFiles {
			return errors.New("the output folder exceeds the readback file limit")
		}
		return nil
	})
	if err != nil {
		return result.with(outputReadbackUnreadable, err.Error())
	}
	sort.Strings(files)
	digest := sha256.New()
	for _, file := range files {
		info, statErr := os.Stat(file)
		if statErr != nil {
			return result.with(outputReadbackUnreadable, statErr.Error())
		}
		result.Bytes += info.Size()
		if result.Bytes > maxValidationDigestBytes {
			return result.with(outputReadbackUnreadable, "the output folder exceeds the readback size limit")
		}
		if err := writeValidationDigestFile(digest, file); err != nil {
			return result.with(outputReadbackUnreadable, err.Error())
		}
	}
	result.Checksum = hex.EncodeToString(digest.Sum(nil))
	if len(files) == 0 || result.Bytes == 0 {
		return result.with(outputReadbackEmpty, "the output folder contains no retained content")
	}
	if expected := normalizeOutputChecksum(expectedChecksum); expected != "" && expected != result.Checksum {
		return result.with(outputReadbackMismatch, "the folder contents do not match the expected digest")
	}
	result.Status = outputReadbackVerified
	return result
}

func (r outputReadback) with(status, detail string) outputReadback {
	r.Status = status
	r.Detail = detail
	return r
}

func normalizeOutputChecksum(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	return strings.TrimPrefix(value, "sha256:")
}

// outputOnlyEchoesRequest reports whether content contains the request text
// and almost nothing else once markup is removed (a template around the ask).
func outputOnlyEchoesRequest(content []byte, requestEchoes []string) bool {
	if len(content) > maxOutputEchoScanBytes {
		return false
	}
	text := normalizeEchoText(outputMarkupPattern.ReplaceAllString(string(content), " "))
	echoed := false
	for _, echo := range requestEchoes {
		normalized := normalizeEchoText(echo)
		if len(normalized) < minRequestEchoChars || !strings.Contains(text, normalized) {
			continue
		}
		text = strings.ReplaceAll(text, normalized, " ")
		echoed = true
	}
	if !echoed {
		return false
	}
	substance := 0
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			substance++
		}
	}
	return substance < minOutputSubstanceChars
}

func normalizeEchoText(value string) string {
	fields := strings.FieldsFunc(strings.ToLower(value), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
	return strings.Join(fields, " ")
}

func sha256Hex(content string) string {
	sum := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sum[:])
}

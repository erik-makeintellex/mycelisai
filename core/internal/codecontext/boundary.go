package codecontext

import (
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

// errPathUnavailable is the only error a caller sees for a path that is
// missing, excluded, escaping the source root, or not a readable regular file.
// Using one message keeps excluded files indistinguishable from absent ones
// and keeps absolute host paths out of API responses.
var errPathUnavailable = errors.New("path is not available in this code context source")

// candidate is a confined source file. real is used only for reading; rel is
// the source-relative slash path and the only path that reaches responses.
type candidate struct {
	real string
	rel  string
	size int64
}

// sensitiveName is the single name-based exclusion rule for code context. It
// applies to every path component (file or directory), regardless of
// extension, on the walk, on requested subpaths, on explain, and on the
// resolved target of any symlink.
func sensitiveName(name string) bool {
	lower := strings.ToLower(strings.TrimSpace(name))
	if lower == ".git" || strings.HasPrefix(lower, ".env") {
		return true
	}
	for _, prefix := range []string{".npmrc", ".pypirc", ".netrc", "id_rsa", "id_dsa", "id_ecdsa", "id_ed25519"} {
		if strings.HasPrefix(lower, prefix) {
			return true
		}
	}
	ext := filepath.Ext(lower)
	switch ext {
	case ".pem", ".key", ".p12", ".pfx":
		return true
	}
	// Generic ssh-style key names (id_<alg>, id_<alg>.pub) carry no source extension.
	if strings.HasPrefix(lower, "id_") && (ext == "" || ext == ".pub") {
		return true
	}
	for _, fragment := range []string{"credential", "kubeconfig", "token", "secret", "service-account", "service_account", "serviceaccount"} {
		if strings.Contains(lower, fragment) {
			return true
		}
	}
	return false
}

// shouldSkipDir names noise directories that are never indexed.
func shouldSkipDir(name string) bool {
	switch strings.ToLower(name) {
	case ".git", ".hg", ".svn", "node_modules", "vendor", "dist", "build", "coverage", ".next", ".turbo", ".cache", "tmp", "temp", "workspace", "saved-media":
		return true
	default:
		return false
	}
}

// includeFile is the file-name decision: never a sensitive name, and only
// allowlisted source/document extensions.
func includeFile(name string) bool {
	if sensitiveName(name) {
		return false
	}
	switch filepath.Ext(strings.ToLower(name)) {
	case ".go", ".ts", ".tsx", ".js", ".jsx", ".py", ".sql", ".yaml", ".yml", ".json", ".md", ".css", ".html":
		return true
	default:
		return false
	}
}

// pathAllowed applies the include/exclude decision to a whole source-relative
// slash path: every component is checked, directories must not be skipped
// or sensitive, and a file must pass includeFile.
func pathAllowed(rel string, isDir bool) bool {
	if rel == "." || rel == "" {
		return isDir
	}
	parts := strings.Split(rel, "/")
	for i, part := range parts {
		if part == "" || part == "." || part == ".." || sensitiveName(part) {
			return false
		}
		if (i < len(parts)-1 || isDir) && shouldSkipDir(part) {
			return false
		}
	}
	return isDir || includeFile(parts[len(parts)-1])
}

// resolveRoot returns the symlink-resolved absolute root of a source.
func resolveRoot(root string) (string, error) {
	real, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", errPathUnavailable
	}
	real, err = filepath.Abs(real)
	if err != nil {
		return "", errPathUnavailable
	}
	return filepath.Clean(real), nil
}

// relWithin returns target relative to root as a slash path, or false when
// target is outside root.
func relWithin(root, target string) (string, bool) {
	rel, err := filepath.Rel(root, target)
	if err != nil || filepath.IsAbs(rel) || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", false
	}
	return filepath.ToSlash(rel), true
}

// resolveSourcePath maps a caller path onto the resolved root lexically.
func resolveSourcePath(rootReal, raw string) (string, error) {
	normalized := strings.ReplaceAll(strings.TrimSpace(raw), "\\", "/")
	if strings.ContainsRune(normalized, '\x00') {
		return "", errPathUnavailable
	}
	normalized = strings.TrimPrefix(path.Clean(normalized), "./")
	if normalized == "" || normalized == "." {
		return "", errPathUnavailable
	}
	target := filepath.FromSlash(normalized)
	if !filepath.IsAbs(target) {
		target = filepath.Join(rootReal, target)
	}
	target = filepath.Clean(target)
	if _, ok := relWithin(rootReal, target); !ok {
		return "", errPathUnavailable
	}
	return target, nil
}

// confine resolves every symlink in lexical and accepts it only when both the
// requested and the resolved path stay inside rootReal, both pass
// pathAllowed, and the target is a directory or a regular file.
func confine(rootReal, lexical string) (candidate, bool, error) {
	rel, ok := relWithin(rootReal, lexical)
	if !ok {
		return candidate{}, false, errPathUnavailable
	}
	real, err := filepath.EvalSymlinks(lexical)
	if err != nil {
		return candidate{}, false, errPathUnavailable
	}
	realRel, ok := relWithin(rootReal, real)
	if !ok {
		return candidate{}, false, errPathUnavailable
	}
	info, err := os.Stat(real)
	if err != nil {
		return candidate{}, false, errPathUnavailable
	}
	isDir := info.IsDir()
	if (!isDir && !info.Mode().IsRegular()) || !pathAllowed(rel, isDir) || !pathAllowed(realRel, isDir) {
		return candidate{}, false, errPathUnavailable
	}
	return candidate{real: real, rel: rel, size: info.Size()}, isDir, nil
}

// collectFiles returns confined, allowed files under an optional subpath.
func collectFiles(rootReal, subpath string) ([]candidate, error) {
	start := candidate{real: rootReal, rel: "."}
	if strings.TrimSpace(subpath) != "" {
		lexical, err := resolveSourcePath(rootReal, subpath)
		if err != nil {
			return nil, err
		}
		c, isDir, err := confine(rootReal, lexical)
		if err != nil {
			return nil, err
		}
		if !isDir {
			return []candidate{c}, nil
		}
		start = c
	}
	files := []candidate{}
	err := filepath.WalkDir(start.real, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if len(files) >= maxFilesScanned {
			return filepath.SkipAll
		}
		walkRel, ok := relWithin(start.real, p)
		if !ok {
			return nil
		}
		rel := path.Clean(path.Join(start.rel, walkRel))
		if d.IsDir() {
			if p != start.real && !pathAllowed(rel, true) {
				return filepath.SkipDir
			}
			return nil
		}
		if !pathAllowed(rel, false) {
			return nil
		}
		c, isDir, err := confine(rootReal, filepath.Join(rootReal, filepath.FromSlash(rel)))
		if err != nil || isDir {
			return nil
		}
		files = append(files, c)
		return nil
	})
	if err != nil {
		return nil, errPathUnavailable
	}
	sort.Slice(files, func(i, j int) bool { return files[i].rel < files[j].rel })
	return files, nil
}

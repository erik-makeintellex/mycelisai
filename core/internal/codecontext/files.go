package codecontext

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

func (s *Service) search(ctx context.Context, source Source, query, subpath string, limit int) ([]Ref, int, error) {
	limit = normalizeLimit(limit)
	q := strings.ToLower(strings.TrimSpace(query))
	files, err := collectFiles(source.Root, subpath)
	if err != nil {
		return nil, 0, err
	}
	refs := make([]Ref, 0, limit)
	for _, file := range files {
		if err := ctx.Err(); err != nil {
			return refs, len(files), err
		}
		rel, _ := filepath.Rel(source.Root, file)
		rel = filepath.ToSlash(rel)
		content, err := os.ReadFile(file)
		if err != nil || len(content) > maxReadBytes {
			continue
		}
		digest := "sha256:" + shortHashBytes(content)
		if strings.Contains(strings.ToLower(rel), q) {
			refs = append(refs, Ref{SourceID: source.ID, SnapshotRef: source.SnapshotRef, CommitOrDigest: digest, FilePath: rel, LineStart: 1, LineEnd: 1, Snippet: "file path match", Score: 0.95, Provenance: "extracted:file_path"})
		}
		lineNo := 0
		scanner := bufio.NewScanner(strings.NewReader(string(content)))
		for scanner.Scan() {
			lineNo++
			line := scanner.Text()
			if !strings.Contains(strings.ToLower(line), q) {
				continue
			}
			refs = append(refs, Ref{SourceID: source.ID, SnapshotRef: source.SnapshotRef, CommitOrDigest: digest, FilePath: rel, LineStart: lineNo, LineEnd: lineNo, Symbol: nearestSymbol(string(content), lineNo), Snippet: strings.TrimSpace(line), Score: 1.0, Provenance: "extracted:text_match"})
			if len(refs) >= limit {
				return refs, len(files), nil
			}
		}
		if len(refs) >= limit {
			return refs[:limit], len(files), nil
		}
	}
	return refs, len(files), nil
}

func (s *Service) explainFile(source Source, rawPath string) ([]Fact, []Ref, error) {
	file, rel, err := resolveSourcePath(source.Root, rawPath)
	if err != nil {
		return nil, nil, err
	}
	content, err := os.ReadFile(file)
	if err != nil {
		return nil, nil, fmt.Errorf("read source file: %w", err)
	}
	if len(content) > maxReadBytes {
		return nil, nil, fmt.Errorf("source file is too large for bounded explain")
	}
	text := string(content)
	digest := "sha256:" + shortHashBytes(content)
	facts := []Fact{{Kind: "file", Value: rel, FilePath: rel, Provenance: "extracted:file_path"}}
	if pkg := firstMatch(`(?m)^\s*package\s+([A-Za-z_][A-Za-z0-9_]*)`, text); pkg != "" {
		facts = append(facts, Fact{Kind: "package", Value: pkg, FilePath: rel, Provenance: "extracted:go_package"})
	}
	for _, imp := range extractImports(text) {
		facts = append(facts, Fact{Kind: "import", Value: imp, FilePath: rel, Provenance: "extracted:import"})
	}
	symbols := extractSymbols(text)
	refs := make([]Ref, 0, len(symbols))
	for _, sym := range symbols {
		facts = append(facts, Fact{Kind: "symbol", Value: sym.Name, FilePath: rel, LineStart: sym.Line, LineEnd: sym.Line, Provenance: "extracted:symbol"})
		refs = append(refs, Ref{SourceID: source.ID, SnapshotRef: source.SnapshotRef, CommitOrDigest: digest, FilePath: rel, LineStart: sym.Line, LineEnd: sym.Line, Symbol: sym.Name, Snippet: sym.Declaration, Score: 1, Provenance: "extracted:symbol"})
	}
	return facts, refs, nil
}

func collectFiles(root, subpath string) ([]string, error) {
	start := root
	if strings.TrimSpace(subpath) != "" {
		resolved, _, err := resolveSourcePath(root, subpath)
		if err != nil {
			return nil, err
		}
		start = resolved
	}
	info, err := os.Stat(start)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return []string{start}, nil
	}
	files := []string{}
	err = filepath.WalkDir(start, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if len(files) >= maxFilesScanned {
			return filepath.SkipDir
		}
		if d.IsDir() {
			if shouldSkipDir(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if includeFile(d.Name()) {
			files = append(files, p)
		}
		return nil
	})
	sort.Strings(files)
	return files, err
}

func resolveSourcePath(root, raw string) (string, string, error) {
	normalized := strings.ReplaceAll(strings.TrimSpace(raw), "\\", "/")
	normalized = strings.TrimPrefix(path.Clean(normalized), "./")
	if normalized == "" || normalized == "." {
		return "", "", fmt.Errorf("source path is required")
	}
	var target string
	if filepath.IsAbs(normalized) {
		target = filepath.Clean(normalized)
	} else {
		target = filepath.Clean(filepath.Join(root, filepath.FromSlash(normalized)))
	}
	rel, err := filepath.Rel(root, target)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", "", fmt.Errorf("path escapes code context source boundary")
	}
	return target, filepath.ToSlash(rel), nil
}

func inferImpact(target string, refs []Ref) []InferredImpact {
	if len(refs) == 0 {
		return []InferredImpact{{
			Kind:       "no_direct_refs",
			Summary:    "No direct extracted references were found for the target in the current bounded scan.",
			Confidence: "medium",
			Reasoning:  []string{"Impact remains unknown until source files are inspected or a refreshed map is available."},
			Provenance: "inferred:bounded_scan",
		}}
	}
	return []InferredImpact{{
		Kind:       "direct_reference_impact",
		Summary:    fmt.Sprintf("%d extracted reference(s) mention %q and should be reviewed for potential impact.", len(refs), target),
		Refs:       refs,
		Confidence: "medium",
		Reasoning:  []string{"This is inferred from text/path references, not from raw graph internals.", "Verify source files before changing behavior."},
		Provenance: "inferred:reference_overlap",
	}}
}

type symbol struct {
	Name        string
	Line        int
	Declaration string
}

func extractSymbols(text string) []symbol {
	re := regexp.MustCompile(`(?m)^\s*(?:func|type|const|var)\s+(?:\([^)]+\)\s*)?([A-Za-z_][A-Za-z0-9_]*)`)
	matches := re.FindAllStringSubmatchIndex(text, 50)
	out := make([]symbol, 0, len(matches))
	for _, m := range matches {
		name := text[m[2]:m[3]]
		decl := strings.TrimSpace(text[m[0]:m[1]])
		out = append(out, symbol{Name: name, Line: lineNumberAt(text, m[0]), Declaration: decl})
	}
	return out
}

func extractImports(text string) []string {
	re := regexp.MustCompile(`(?m)^\s*(?:"([^"]+)"|[A-Za-z_][A-Za-z0-9_]*\s+"([^"]+)")`)
	matches := re.FindAllStringSubmatch(text, 50)
	out := []string{}
	for _, m := range matches {
		if m[1] != "" {
			out = append(out, m[1])
		} else if m[2] != "" {
			out = append(out, m[2])
		}
	}
	return out
}

func nearestSymbol(text string, line int) string {
	best := ""
	bestLine := 0
	for _, sym := range extractSymbols(text) {
		if sym.Line <= line && sym.Line >= bestLine {
			best = sym.Name
			bestLine = sym.Line
		}
	}
	return best
}

func firstMatch(pattern, text string) string {
	re := regexp.MustCompile(pattern)
	m := re.FindStringSubmatch(text)
	if len(m) > 1 {
		return m[1]
	}
	return ""
}

func lineNumberAt(text string, offset int) int {
	if offset <= 0 {
		return 1
	}
	return strings.Count(text[:offset], "\n") + 1
}

func shouldSkipDir(name string) bool {
	switch strings.ToLower(name) {
	case ".git", ".hg", ".svn", "node_modules", "vendor", "dist", "build", "coverage", ".next", ".turbo", ".cache", "tmp", "temp", "workspace", "saved-media":
		return true
	default:
		return false
	}
}

func includeFile(name string) bool {
	lower := strings.ToLower(name)
	if strings.HasPrefix(lower, ".env") || strings.Contains(lower, "secret") {
		return false
	}
	switch filepath.Ext(lower) {
	case ".go", ".ts", ".tsx", ".js", ".jsx", ".py", ".sql", ".yaml", ".yml", ".json", ".md", ".css", ".html":
		return true
	default:
		return false
	}
}

func normalizeLimit(limit int) int {
	if limit <= 0 {
		return defaultLimit
	}
	if limit > maxLimit {
		return maxLimit
	}
	return limit
}

func splitRoots(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := filepath.SplitList(raw)
	out := []string{}
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func findRepoRoot(start string) string {
	current := start
	for {
		if _, err := os.Stat(filepath.Join(current, ".git")); err == nil {
			return current
		}
		if _, err := os.Stat(filepath.Join(current, "AGENTS.md")); err == nil {
			return current
		}
		parent := filepath.Dir(current)
		if parent == current {
			return start
		}
		current = parent
	}
}

func sourceID(root string) string {
	base := strings.ToLower(filepath.Base(root))
	base = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(base, "-")
	base = strings.Trim(base, "-")
	if base == "" {
		return "code-source"
	}
	return base
}

func shortHash(value string) string {
	return shortHashBytes([]byte(filepath.Clean(value)))
}

func shortHashBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return hex.EncodeToString(sum[:])[:16]
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

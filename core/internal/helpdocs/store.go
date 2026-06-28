package helpdocs

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const defaultExcerptLimit = 1200

type Store struct {
	root    string
	entries map[string]Entry
}

type Document struct {
	Entry
	Content string `json:"content"`
	Excerpt string `json:"excerpt,omitempty"`
}

type SearchResult struct {
	Entry
	Excerpt string `json:"excerpt"`
	Score   int    `json:"score"`
}

func NewStore(root string) (*Store, error) {
	if strings.TrimSpace(root) == "" {
		detected, err := FindRepoRoot()
		if err != nil {
			return nil, err
		}
		root = detected
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve docs root: %w", err)
	}
	entries := make(map[string]Entry)
	for _, entry := range Flatten(Manifest) {
		if entry.Slug == "" || entry.Path == "" {
			continue
		}
		entries[entry.Slug] = entry
	}
	return &Store{root: absRoot, entries: entries}, nil
}

func FindRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if fileExists(filepath.Join(dir, "README.md")) && fileExists(filepath.Join(dir, "docs", "README.md")) {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("could not find repository root for help docs")
		}
		dir = parent
	}
}

func (s *Store) Sections() []Section {
	return Manifest
}

func (s *Store) Read(slug string) (Document, error) {
	entry, ok := s.entries[strings.TrimSpace(slug)]
	if !ok {
		return Document{}, fmt.Errorf("unknown documentation slug %q", slug)
	}
	content, err := os.ReadFile(s.safePath(entry.Path))
	if err != nil {
		return Document{}, fmt.Errorf("read %s: %w", entry.Path, err)
	}
	text := string(content)
	return Document{Entry: entry, Content: text, Excerpt: excerptAround(text, "", defaultExcerptLimit)}, nil
}

func (s *Store) Search(query string, limit int) ([]SearchResult, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}
	if limit <= 0 || limit > 20 {
		limit = 8
	}
	tokens := tokenSet(query)
	var results []SearchResult
	for _, entry := range s.entries {
		content, err := os.ReadFile(s.safePath(entry.Path))
		if err != nil {
			continue
		}
		text := string(content)
		score := scoreText(tokens, entry, text)
		if score == 0 {
			continue
		}
		results = append(results, SearchResult{
			Entry:   entry,
			Excerpt: excerptAround(text, query, defaultExcerptLimit),
			Score:   score,
		})
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].Score == results[j].Score {
			return results[i].Slug < results[j].Slug
		}
		return results[i].Score > results[j].Score
	})
	if len(results) > limit {
		results = results[:limit]
	}
	return results, nil
}

func (s *Store) safePath(rel string) string {
	clean := filepath.Clean(filepath.FromSlash(strings.TrimSpace(rel)))
	if filepath.IsAbs(clean) || strings.HasPrefix(clean, "..") {
		return filepath.Join(s.root, "__invalid__")
	}
	target := filepath.Join(s.root, clean)
	absTarget, err := filepath.Abs(target)
	if err != nil {
		return filepath.Join(s.root, "__invalid__")
	}
	relToRoot, err := filepath.Rel(s.root, absTarget)
	if err != nil || strings.HasPrefix(relToRoot, "..") {
		return filepath.Join(s.root, "__invalid__")
	}
	return absTarget
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func tokenSet(text string) map[string]struct{} {
	out := map[string]struct{}{}
	for _, part := range strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '_' || r == '-')
	}) {
		if len(part) < 3 {
			continue
		}
		out[part] = struct{}{}
	}
	return out
}

func scoreText(tokens map[string]struct{}, entry Entry, content string) int {
	haystack := strings.ToLower(entry.Slug + " " + entry.Label + " " + entry.Description + " " + content)
	score := 0
	for token := range tokens {
		if strings.Contains(haystack, token) {
			score++
		}
	}
	return score
}

func excerptAround(content, query string, limit int) string {
	content = strings.TrimSpace(content)
	if len(content) <= limit {
		return content
	}
	start := 0
	for token := range tokenSet(query) {
		if idx := strings.Index(strings.ToLower(content), token); idx >= 0 {
			start = idx - limit/4
			break
		}
	}
	if start < 0 {
		start = 0
	}
	if start+limit > len(content) {
		start = len(content) - limit
	}
	return strings.TrimSpace(content[start:start+limit]) + "..."
}

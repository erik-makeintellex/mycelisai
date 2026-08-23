package codecontext

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func (s *Service) normalizeSourceInput(input SourceInput) (Source, error) {
	root := strings.TrimSpace(input.RootPath)
	if root == "" {
		return Source{}, fmt.Errorf("root_path is required")
	}
	if strings.ContainsRune(root, '\x00') || strings.Contains(root, "://") {
		return Source{}, fmt.Errorf("root_path must be a local filesystem path")
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return Source{}, fmt.Errorf("resolve root_path: %w", err)
	}
	abs = filepath.Clean(abs)
	if !s.rootAllowed(abs) {
		return Source{}, fmt.Errorf("root_path must stay inside an approved code context root")
	}
	info, err := os.Stat(abs)
	if err != nil {
		return Source{}, fmt.Errorf("root_path not found: %w", err)
	}
	if !info.IsDir() {
		return Source{}, fmt.Errorf("root_path must be a directory")
	}
	sourceType := normalizeSourceToken(input.SourceType)
	switch sourceType {
	case "", "repository", "code_repository":
		sourceType = "repository"
	case "local_folder", "local_code_folder", "code_folder":
		sourceType = "local_code_folder"
	default:
		return Source{}, fmt.Errorf("unsupported source_type %q", input.SourceType)
	}
	id := normalizeSourceID(firstNonEmpty(input.ID, input.ConfigRecordID, input.Name, filepath.Base(abs)))
	if id == "" {
		id = sourceID(abs)
	}
	if !regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{2,80}$`).MatchString(id) {
		return Source{}, fmt.Errorf("source id must be a stable lowercase identifier")
	}
	scopeKind := normalizeSourceToken(input.ScopeKind)
	if scopeKind == "" {
		scopeKind = "workspace"
	}
	if !allowedCodeContextScope(scopeKind) {
		return Source{}, fmt.Errorf("unsupported scope_kind %q", input.ScopeKind)
	}
	scopeRef := strings.TrimSpace(input.ScopeRef)
	if scopeKind == "all" || scopeKind == "built_in" {
		scopeRef = ""
	} else if scopeRef == "" && (scopeKind == "group" || scopeKind == "host") {
		return Source{}, fmt.Errorf("scope_ref is required for %s code context sources", scopeKind)
	}
	version := strings.TrimSpace(input.ExtractionVersion)
	if version == "" {
		version = "code-context-fixture-v1"
	}
	if version != "code-context-fixture-v1" {
		return Source{}, fmt.Errorf("unsupported extraction_version %q", version)
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		name = filepath.Base(abs)
	}
	return Source{
		ID:                id,
		Name:              name,
		Root:              abs,
		Boundary:          "local_code_folder",
		SourceType:        sourceType,
		ScopeKind:         scopeKind,
		ScopeRef:          scopeRef,
		ConfigRecordID:    strings.TrimSpace(input.ConfigRecordID),
		ConfigDigest:      strings.TrimSpace(input.ConfigDigest),
		ExtractionVersion: version,
		SensitivityClass:  firstNonEmpty(normalizeSourceToken(input.SensitivityClass), "restricted"),
		TrustClass:        firstNonEmpty(normalizeSourceToken(input.TrustClass), "trusted_internal"),
		Status:            "available",
		SnapshotRef:       "local:" + shortHash(abs),
		IncludeGlobs:      cleanSourceList(input.IncludeGlobs),
		ExcludeGlobs:      cleanSourceList(input.ExcludeGlobs),
		Languages:         cleanSourceList(input.Languages),
	}, nil
}

func (s *Service) rootAllowed(abs string) bool {
	for _, source := range s.sources {
		root := filepath.Clean(source.Root)
		if rel, err := filepath.Rel(root, abs); err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return true
		}
	}
	return len(s.sources) == 0
}

func allowedCodeContextScope(scope string) bool {
	switch scope {
	case "all", "group", "host", "operator", "workspace", "organization", "built_in":
		return true
	default:
		return false
	}
}

func normalizeSourceToken(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	value = strings.ReplaceAll(value, " ", "_")
	value = strings.ReplaceAll(value, "-", "_")
	return value
}

func normalizeSourceID(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	value = regexp.MustCompile(`[^a-z0-9_-]+`).ReplaceAllString(value, "-")
	value = strings.Trim(value, "-_")
	return value
}

func cleanSourceList(values []string) []string {
	out := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func sourcePublic(source Source) *Source {
	out := source
	out.Root = ""
	return &out
}

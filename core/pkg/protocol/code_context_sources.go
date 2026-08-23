package protocol

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	CodeContextSourceTypeRepository  = "repository"
	CodeContextSourceTypeLocalFolder = "local_code_folder"
	CodeContextExtractionVersionV1   = "code-context-fixture-v1"
)

var codeContextSourceIDPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{2,80}$`)

// CodeContextSourceSpec registers a scoped repository or local code folder as
// a native Mycelis source. It describes what may be indexed, not the extracted
// index itself.
type CodeContextSourceSpec struct {
	SourceID          string   `json:"source_id" yaml:"source_id"`
	SourceType        string   `json:"source_type" yaml:"source_type"`
	RootPath          string   `json:"root_path" yaml:"root_path"`
	IncludeGlobs      []string `json:"include_globs,omitempty" yaml:"include_globs,omitempty"`
	ExcludeGlobs      []string `json:"exclude_globs,omitempty" yaml:"exclude_globs,omitempty"`
	Languages         []string `json:"languages,omitempty" yaml:"languages,omitempty"`
	ExtractionVersion string   `json:"extraction_version,omitempty" yaml:"extraction_version,omitempty"`
	SensitivityClass  string   `json:"sensitivity_class,omitempty" yaml:"sensitivity_class,omitempty"`
	TrustClass        string   `json:"trust_class,omitempty" yaml:"trust_class,omitempty"`
}

type CodeContextSourceCompileResult struct {
	Source CodeContextSourceSpec `json:"source"`
	Scope  ConfigDocumentScope   `json:"scope"`
	Digest string                `json:"digest"`
}

func DecodeCodeContextSourceSpec(raw json.RawMessage) (CodeContextSourceSpec, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var spec CodeContextSourceSpec
	if err := decoder.Decode(&spec); err != nil {
		return CodeContextSourceSpec{}, fmt.Errorf("decode code context source: %w", err)
	}
	if decoder.More() {
		return CodeContextSourceSpec{}, errors.New("decode code context source: trailing content")
	}
	return spec, nil
}

func ValidateCodeContextSourceSpec(raw json.RawMessage) []ConfigDocumentValidationIssue {
	spec, err := DecodeCodeContextSourceSpec(raw)
	issues := make([]ConfigDocumentValidationIssue, 0)
	add := func(code, field, message string) {
		issues = append(issues, ConfigDocumentValidationIssue{Code: code, Field: field, Message: message})
	}
	if err != nil {
		add("spec.code_context.invalid", "spec", err.Error())
		return issues
	}
	validateCodeContextSource(spec, "spec", add)
	return issues
}

func validateCodeContextSource(spec CodeContextSourceSpec, prefix string, add func(string, string, string)) {
	if !codeContextSourceIDPattern.MatchString(spec.SourceID) {
		add("spec.code_context.invalid_source_id", prefix+".source_id", "source_id must be a stable lowercase identifier")
	}
	switch normalizeCodeContextToken(spec.SourceType) {
	case CodeContextSourceTypeRepository, CodeContextSourceTypeLocalFolder:
	default:
		add("spec.code_context.unsupported_source_type", prefix+".source_type", "source_type must be repository or local_code_folder")
	}
	if err := validateCodeContextPathToken(spec.RootPath); err != nil {
		add("spec.code_context.invalid_root_path", prefix+".root_path", err.Error())
	}
	validateCodeContextStringList(spec.IncludeGlobs, prefix+".include_globs", add)
	validateCodeContextStringList(spec.ExcludeGlobs, prefix+".exclude_globs", add)
	for i, language := range spec.Languages {
		if normalizeCodeContextToken(language) == "" {
			add("spec.code_context.invalid_language", fmt.Sprintf("%s.languages[%d]", prefix, i), "language values must be non-empty tokens")
		}
	}
	if spec.ExtractionVersion != "" && spec.ExtractionVersion != CodeContextExtractionVersionV1 {
		add("spec.code_context.unsupported_extraction_version", prefix+".extraction_version", "unsupported code context extraction version")
	}
}

func validateCodeContextPathToken(raw string) error {
	path := strings.TrimSpace(raw)
	if path == "" {
		return errors.New("root_path is required")
	}
	if path != raw {
		return errors.New("root_path must not have surrounding whitespace")
	}
	if strings.ContainsRune(path, '\x00') {
		return errors.New("root_path must not contain null bytes")
	}
	if strings.Contains(path, "://") {
		return errors.New("root_path must be a local filesystem path")
	}
	clean := filepath.ToSlash(filepath.Clean(path))
	if clean == ".." || strings.HasPrefix(clean, "../") {
		return errors.New("root_path must stay inside an allowed source root")
	}
	return nil
}

func validateCodeContextStringList(values []string, field string, add func(string, string, string)) {
	seen := map[string]struct{}{}
	for i, value := range values {
		item := strings.TrimSpace(value)
		itemField := fmt.Sprintf("%s[%d]", field, i)
		if item == "" || item != value || strings.ContainsRune(item, '\x00') {
			add("spec.code_context.invalid_list_value", itemField, "list values must be non-empty clean strings")
			continue
		}
		if _, exists := seen[item]; exists {
			add("spec.code_context.duplicate_list_value", itemField, "list values must be unique")
			continue
		}
		seen[item] = struct{}{}
	}
}

func normalizeCodeContextToken(raw string) string {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	normalized = strings.ReplaceAll(normalized, "-", "_")
	normalized = strings.ReplaceAll(normalized, " ", "_")
	return normalized
}

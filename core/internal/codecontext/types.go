package codecontext

import "time"

const (
	defaultLimit    = 8
	maxLimit        = 25
	maxFilesScanned = 1200
	maxReadBytes    = 256 << 10
)

type Config struct {
	SourceRoots []string
	Now         func() time.Time
}

type Service struct {
	sources []Source
	now     func() time.Time
}

type Source struct {
	ID                string   `json:"id"`
	Name              string   `json:"name"`
	Root              string   `json:"root,omitempty"`
	Boundary          string   `json:"boundary"`
	SourceType        string   `json:"source_type,omitempty"`
	ScopeKind         string   `json:"scope_kind,omitempty"`
	ScopeRef          string   `json:"scope_ref,omitempty"`
	ConfigRecordID    string   `json:"config_record_id,omitempty"`
	ConfigDigest      string   `json:"config_digest,omitempty"`
	ExtractionVersion string   `json:"extraction_version,omitempty"`
	SensitivityClass  string   `json:"sensitivity_class,omitempty"`
	TrustClass        string   `json:"trust_class,omitempty"`
	Status            string   `json:"status"`
	SnapshotRef       string   `json:"snapshot_ref"`
	IncludeGlobs      []string `json:"include_globs,omitempty"`
	ExcludeGlobs      []string `json:"exclude_globs,omitempty"`
	Languages         []string `json:"languages,omitempty"`
}

type SourceInput struct {
	ID                string   `json:"id,omitempty"`
	Name              string   `json:"name"`
	SourceType        string   `json:"source_type"`
	RootPath          string   `json:"root_path"`
	ScopeKind         string   `json:"scope_kind,omitempty"`
	ScopeRef          string   `json:"scope_ref,omitempty"`
	ConfigRecordID    string   `json:"config_record_id,omitempty"`
	ConfigDigest      string   `json:"config_digest,omitempty"`
	ExtractionVersion string   `json:"extraction_version,omitempty"`
	SensitivityClass  string   `json:"sensitivity_class,omitempty"`
	TrustClass        string   `json:"trust_class,omitempty"`
	IncludeGlobs      []string `json:"include_globs,omitempty"`
	ExcludeGlobs      []string `json:"exclude_globs,omitempty"`
	Languages         []string `json:"languages,omitempty"`
}

type Request struct {
	Query    string `json:"query,omitempty"`
	SourceID string `json:"source_id,omitempty"`
	Path     string `json:"path,omitempty"`
	Target   string `json:"target,omitempty"`
	Symbol   string `json:"symbol,omitempty"`
	Limit    int    `json:"limit,omitempty"`
}

type QueryRequest = Request
type ImpactRequest = Request
type ExplainRequest = Request

type Response struct {
	Status                string           `json:"status"`
	CapabilityID          string           `json:"capability_id"`
	Operation             string           `json:"operation"`
	Query                 string           `json:"query,omitempty"`
	Target                string           `json:"target,omitempty"`
	Source                *Source          `json:"source,omitempty"`
	Refs                  []Ref            `json:"refs,omitempty"`
	ExtractedFacts        []Fact           `json:"extracted_facts,omitempty"`
	InferredRelationships []InferredImpact `json:"inferred_relationships,omitempty"`
	Count                 int              `json:"count"`
	Blocker               *Blocker         `json:"blocker,omitempty"`
	Metadata              map[string]any   `json:"metadata,omitempty"`
}

type Ref struct {
	SourceID       string  `json:"source_id"`
	SnapshotRef    string  `json:"snapshot_ref"`
	CommitOrDigest string  `json:"commit_or_digest"`
	FilePath       string  `json:"file_path"`
	LineStart      int     `json:"line_start,omitempty"`
	LineEnd        int     `json:"line_end,omitempty"`
	Symbol         string  `json:"symbol,omitempty"`
	Snippet        string  `json:"snippet,omitempty"`
	Score          float64 `json:"score,omitempty"`
	Provenance     string  `json:"provenance"`
}

type Fact struct {
	Kind       string `json:"kind"`
	Value      string `json:"value"`
	FilePath   string `json:"file_path,omitempty"`
	LineStart  int    `json:"line_start,omitempty"`
	LineEnd    int    `json:"line_end,omitempty"`
	Provenance string `json:"provenance"`
}

type InferredImpact struct {
	Kind       string   `json:"kind"`
	Summary    string   `json:"summary"`
	Refs       []Ref    `json:"refs,omitempty"`
	Confidence string   `json:"confidence"`
	Reasoning  []string `json:"reasoning,omitempty"`
	Provenance string   `json:"provenance"`
}

type Blocker struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	NextAction string `json:"next_action"`
}

type Store struct{}

func NewStore(any) *Store {
	return &Store{}
}

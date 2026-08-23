package codecontext

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func ConfigFromEnv() Config {
	roots := splitRoots(os.Getenv("MYCELIS_CODE_CONTEXT_ROOTS"))
	if len(roots) == 0 {
		for _, key := range []string{"MYCELIS_CODE_CONTEXT_ROOT", "MYCELIS_BACKEND_WORKSPACE_ROOT"} {
			if value := strings.TrimSpace(os.Getenv(key)); value != "" {
				roots = append(roots, value)
				break
			}
		}
	}
	if len(roots) == 0 {
		if cwd, err := os.Getwd(); err == nil {
			roots = append(roots, findRepoRoot(cwd))
		}
	}
	return Config{SourceRoots: roots}
}

func NewService(cfg Config) *Service {
	if cfg.Now == nil {
		cfg.Now = func() time.Time { return time.Now().UTC() }
	}
	s := &Service{now: cfg.Now}
	seen := map[string]int{}
	for _, raw := range cfg.SourceRoots {
		root := strings.TrimSpace(raw)
		if root == "" {
			continue
		}
		abs, err := filepath.Abs(root)
		if err != nil {
			continue
		}
		id := sourceID(abs)
		if seen[id] > 0 {
			id = fmt.Sprintf("%s-%d", id, seen[id]+1)
		}
		seen[id]++
		status := "available"
		if info, err := os.Stat(abs); err != nil || !info.IsDir() {
			status = "blocked"
		}
		s.sources = append(s.sources, Source{
			ID:                id,
			Name:              filepath.Base(abs),
			Root:              abs,
			Boundary:          "local_code_folder",
			SourceType:        "local_code_folder",
			ScopeKind:         "workspace",
			ExtractionVersion: "code-context-fixture-v1",
			SensitivityClass:  "restricted",
			TrustClass:        "trusted_internal",
			Status:            status,
			SnapshotRef:       "local:" + shortHash(abs),
		})
	}
	return s
}

func (s *Service) ListSources(context.Context) ([]Source, error) {
	if s == nil {
		return nil, nil
	}
	out := make([]Source, 0, len(s.sources))
	for _, source := range s.sources {
		out = append(out, *sourcePublic(source))
	}
	return out, nil
}

func (s *Service) RegisterSource(_ context.Context, input SourceInput) (Source, error) {
	if s == nil {
		return Source{}, fmt.Errorf("code context service unavailable")
	}
	source, err := s.normalizeSourceInput(input)
	if err != nil {
		return Source{}, err
	}
	for i := range s.sources {
		if s.sources[i].ID == source.ID {
			s.sources[i] = source
			return *sourcePublic(source), nil
		}
	}
	s.sources = append(s.sources, source)
	return *sourcePublic(source), nil
}

func (s *Service) Index(ctx context.Context, sourceID string) (Response, error) {
	resp, source, blocker := s.prepare("code_context.index", sourceID)
	if blocker != nil {
		resp.Blocker = blocker
		return resp, nil
	}
	refs, scanned, err := s.search(ctx, *source, ".", "", 1)
	if err != nil {
		return resp, err
	}
	resp.Refs = refs
	resp.Count = len(refs)
	resp.Metadata = map[string]any{
		"scanned_files":       scanned,
		"snapshot_ref":        source.SnapshotRef,
		"commit_or_digest":    source.SnapshotRef,
		"extraction_version":  source.ExtractionVersion,
		"storage_model":       "runtime_in_memory_snapshot",
		"raw_graph_internals": "not_exposed",
	}
	return resp, nil
}

func (s *Service) BuildSnapshot(ctx context.Context, sourceID string) (Response, error) {
	return s.Index(ctx, sourceID)
}

func (s *Service) Query(ctx context.Context, req Request) (Response, error) {
	resp, source, blocker := s.prepare("code_context.query", req.SourceID)
	resp.Query = strings.TrimSpace(req.Query)
	if blocker != nil {
		resp.Blocker = blocker
		return resp, nil
	}
	if resp.Query == "" {
		resp.Blocker = &Blocker{Code: "missing_query", Message: "Code context query requires a non-empty query.", NextAction: "Provide a symbol, file path fragment, package name, or phrase to search."}
		resp.Status = "blocked"
		return resp, nil
	}
	refs, scanned, err := s.search(ctx, *source, resp.Query, req.Path, req.Limit)
	if err != nil {
		return resp, err
	}
	resp.Refs = refs
	resp.Count = len(refs)
	resp.Metadata = map[string]any{
		"scanned_files":             scanned,
		"raw_graph_internals":       "not_exposed",
		"facts_are_source_derived":  true,
		"inference_requires_review": false,
	}
	return resp, nil
}

func (s *Service) Impact(ctx context.Context, req Request) (Response, error) {
	target := strings.TrimSpace(firstNonEmpty(req.Target, req.Symbol, req.Path, req.Query))
	resp, source, blocker := s.prepare("code_context.impact", req.SourceID)
	resp.Target = target
	if blocker != nil {
		resp.Blocker = blocker
		return resp, nil
	}
	if target == "" {
		resp.Blocker = &Blocker{Code: "missing_target", Message: "Impact review requires a target file, symbol, package, or phrase.", NextAction: "Provide target, symbol, path, or query."}
		resp.Status = "blocked"
		return resp, nil
	}

	query := target
	if req.Path != "" {
		query = filepath.Base(req.Path)
	}
	refs, scanned, err := s.search(ctx, *source, query, "", req.Limit)
	if err != nil {
		return resp, err
	}
	resp.Refs = refs
	resp.Count = len(refs)
	resp.InferredRelationships = inferImpact(target, refs)
	resp.Metadata = map[string]any{
		"scanned_files":               scanned,
		"extracted_vs_inferred_split": true,
		"raw_graph_internals":         "not_exposed",
	}
	return resp, nil
}

func (s *Service) Explain(ctx context.Context, req Request) (Response, error) {
	resp, source, blocker := s.prepare("code_context.explain", req.SourceID)
	target := strings.TrimSpace(firstNonEmpty(req.Path, req.Symbol, req.Target, req.Query))
	resp.Target = target
	if blocker != nil {
		resp.Blocker = blocker
		return resp, nil
	}
	if target == "" {
		resp.Blocker = &Blocker{Code: "missing_target", Message: "Explain requires a file path, symbol, package, or phrase.", NextAction: "Provide path, symbol, target, or query."}
		resp.Status = "blocked"
		return resp, nil
	}
	if req.Path != "" {
		facts, refs, err := s.explainFile(*source, req.Path)
		if err != nil {
			resp.Status = "blocked"
			resp.Blocker = &Blocker{Code: "explain_failed", Message: err.Error(), NextAction: "Use a source-relative path inside the registered code boundary."}
			return resp, nil
		}
		resp.ExtractedFacts = facts
		resp.Refs = refs
		resp.Count = len(refs)
	} else {
		refs, scanned, err := s.search(ctx, *source, target, "", req.Limit)
		if err != nil {
			return resp, err
		}
		resp.Refs = refs
		resp.Count = len(refs)
		resp.Metadata = map[string]any{"scanned_files": scanned}
	}
	if resp.Metadata == nil {
		resp.Metadata = map[string]any{}
	}
	resp.Metadata["raw_graph_internals"] = "not_exposed"
	resp.Metadata["facts_are_source_derived"] = true
	return resp, nil
}

func (s *Service) prepare(operation, sourceID string) (Response, *Source, *Blocker) {
	resp := Response{
		Status:       "ok",
		CapabilityID: "code_context",
		Operation:    operation,
	}
	if s == nil || len(s.sources) == 0 {
		resp.Status = "blocked"
		return resp, nil, &Blocker{Code: "code_context_unconfigured", Message: "No local code context source is configured.", NextAction: "Set MYCELIS_CODE_CONTEXT_ROOTS or MYCELIS_CODE_CONTEXT_ROOT to an approved repository or code folder."}
	}
	source := s.selectSource(sourceID)
	if source == nil {
		resp.Status = "blocked"
		return resp, nil, &Blocker{Code: "code_context_source_not_found", Message: "The requested code context source is not registered.", NextAction: "Use an available source_id or omit source_id for the default source."}
	}
	resp.Source = sourcePublic(*source)
	if source.Status != "available" {
		resp.Status = "blocked"
		return resp, source, &Blocker{Code: "code_context_source_unavailable", Message: "The configured code context source is unavailable.", NextAction: "Check the configured path and refresh the runtime."}
	}
	return resp, source, nil
}

func (s *Service) selectSource(sourceID string) *Source {
	sourceID = strings.TrimSpace(sourceID)
	if sourceID == "" {
		return &s.sources[0]
	}
	for i := range s.sources {
		if s.sources[i].ID == sourceID {
			return &s.sources[i]
		}
	}
	return nil
}

package searchcap

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"
)

const (
	mountedFolderMaxFiles = 2000
	mountedFolderMaxBytes = 128 * 1024
)

type mountedFolderHit struct {
	result Result
	path   string
}

func (s *Service) searchMountedFolder(ctx context.Context, req Request, resp Response, source Source) (Response, error) {
	root := strings.TrimSpace(source.Endpoint)
	if root == "" {
		resp.Status = "blocked"
		resp.Blocker = &Blocker{Code: "mounted_folder_path_missing", Message: "The selected data mount has no path configured.", NextAction: "Update the data mount path in Resources > Capabilities."}
		return resp, nil
	}
	info, err := os.Stat(root)
	if err != nil {
		resp.Status = "blocked"
		resp.Blocker = &Blocker{Code: "mounted_folder_unreachable", Message: "The selected data mount is not reachable from Core.", NextAction: "Confirm the mounted folder path exists on this host and restart Core if the mount was added after startup."}
		return resp, nil
	}
	if !info.IsDir() {
		resp.Status = "blocked"
		resp.Blocker = &Blocker{Code: "mounted_folder_not_directory", Message: "The selected data mount path is not a folder.", NextAction: "Update the data mount to point at a readable folder."}
		return resp, nil
	}

	terms := searchTerms(req.Query)
	limit := limitFor(req.MaxResults, s.cfg.MaxResults)
	hits := []mountedFolderHit{}
	filesSeen := 0
	now := time.Now().UTC()
	resultKind := "mounted_folder"
	interpretation := "mounted_folder_results_are_operator_configured_local_data"
	if normalizeSourceToken(source.Provider) == ProviderCodeContext || isCodeContextSourceType(source.SourceType) {
		resultKind = ProviderCodeContext
		interpretation = "code_context_results_are_operator_configured_repository_or_code_folder_refs"
	}
	walkErr := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if entry.Type()&os.ModeSymlink != 0 {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.IsDir() {
			return nil
		}
		filesSeen++
		if filesSeen > mountedFolderMaxFiles {
			return filepath.SkipAll
		}
		if !isMountedSearchFile(path) {
			return nil
		}
		text, err := readMountedSearchText(path)
		if err != nil {
			return nil
		}
		score, matched := scoreMountedText(text, terms)
		if !matched {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			rel = filepath.Base(path)
		}
		rel = filepath.ToSlash(rel)
		hits = append(hits, mountedFolderHit{
			path: rel,
			result: Result{
				Title:            rel,
				URL:              "mount://" + source.ID + "/" + rel,
				LocalSourceID:    source.ID + ":" + rel,
				Snippet:          mountedSnippet(text, terms),
				SourceKind:       resultKind,
				TrustClass:       firstString(source.TrustClass, "bounded_internal"),
				SensitivityClass: source.SensitivityClass,
				RetrievedAt:      now,
				Score:            score,
				ProviderMetadata: map[string]any{
					"source_id":     source.ID,
					"source_name":   source.Name,
					"source_type":   source.SourceType,
					"boundary":      source.Boundary,
					"relative_path": rel,
					"mode":          source.Mode,
				},
			},
		})
		return nil
	})
	if walkErr != nil && ctx.Err() != nil {
		return resp, ctx.Err()
	}
	sort.SliceStable(hits, func(i, j int) bool {
		if hits[i].result.Score == hits[j].result.Score {
			return hits[i].path < hits[j].path
		}
		return hits[i].result.Score > hits[j].result.Score
	})
	for _, hit := range hits {
		if len(resp.Results) >= limit {
			break
		}
		resp.Results = append(resp.Results, hit.result)
	}
	resp.Count = len(resp.Results)
	resp.Provider = ProviderLocalSources
	if resultKind == ProviderCodeContext {
		resp.Provider = ProviderCodeContext
	}
	resp.Metadata["mounted_folder_files_scanned"] = filesSeen
	resp.Metadata["mounted_folder_source_id"] = source.ID
	resp.Metadata["mounted_folder_source_name"] = source.Name
	resp.Metadata["interpretation"] = interpretation
	return resp, nil
}

func searchTerms(query string) []string {
	parts := strings.FieldsFunc(strings.ToLower(query), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	terms := []string{}
	for _, part := range parts {
		if len(part) >= 2 {
			terms = append(terms, part)
		}
	}
	return terms
}

func isMountedSearchFile(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".txt", ".md", ".markdown", ".json", ".yaml", ".yml", ".csv", ".log", ".go", ".py", ".ts", ".tsx", ".js", ".jsx", ".html", ".css", ".sql":
		return true
	default:
		return false
	}
}

func readMountedSearchText(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, mountedFolderMaxBytes))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func scoreMountedText(text string, terms []string) (float64, bool) {
	lower := strings.ToLower(text)
	if len(terms) == 0 {
		return 0, false
	}
	matches := 0
	for _, term := range terms {
		if strings.Contains(lower, term) {
			matches++
		}
	}
	if matches == 0 {
		return 0, false
	}
	return float64(matches) / float64(len(terms)), true
}

func mountedSnippet(text string, terms []string) string {
	compact := strings.Join(strings.Fields(text), " ")
	lower := strings.ToLower(compact)
	start := 0
	for _, term := range terms {
		if idx := strings.Index(lower, term); idx >= 0 {
			start = idx
			break
		}
	}
	if start > 60 {
		start -= 60
	}
	end := start + 240
	if end > len(compact) {
		end = len(compact)
	}
	return strings.TrimSpace(compact[start:end])
}

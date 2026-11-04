package server

import (
	"regexp"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

var nonSlugCharacterPattern = regexp.MustCompile(`[^a-z0-9]+`)

func isCodexTaskBundleRequest(text string) bool {
	lower := strings.ToLower(strings.TrimSpace(text))
	if lower == "" {
		return false
	}
	hasCodexTarget := strings.Contains(lower, "codex") || strings.Contains(lower, "implementation task")
	hasPackageIntent := requestContainsAny(lower, []string{
		"task package",
		"task bundle",
		"implementation package",
		"implementation bundle",
		"codex package",
		"codex bundle",
		"codex-ready",
		"codex ready",
	})
	hasCreationIntent := requestContainsAny(lower, []string{"create", "generate", "draft", "prepare", "make", "build"})
	return hasCodexTarget && hasPackageIntent && hasCreationIntent
}

func codexBundleSlugFromRequest(text string) string {
	lower := strings.ToLower(strings.TrimSpace(text))
	fields := strings.Fields(lower)
	for _, field := range fields {
		trimmed := strings.Trim(field, ".,:;()[]{}'\"`")
		if strings.HasPrefix(trimmed, "myc-rc-") && len(trimmed) > len("myc-rc-") {
			return sanitizeCodexBundleSlug(trimmed)
		}
	}
	return "codex-task-package"
}

func sanitizeCodexBundleSlug(value string) string {
	trimmed := strings.ToLower(strings.TrimSpace(value))
	if trimmed == "" {
		return "codex-task-package"
	}
	slug := strings.Trim(nonSlugCharacterPattern.ReplaceAllString(trimmed, "-"), "-")
	if slug == "" {
		return "codex-task-package"
	}
	if len(slug) > 80 {
		slug = strings.Trim(slug[:80], "-")
	}
	return slug
}

func inferCodexTaskBundlePlanFromRequest(text string) ([]protocol.PlannedToolCall, bool) {
	trimmed := strings.TrimSpace(text)
	if !isCodexTaskBundleRequest(trimmed) {
		return nil, false
	}

	slug := codexBundleSlugFromRequest(trimmed)
	basePath := "codex-task-packages/" + slug
	goal := trimmed
	files := []struct {
		name    string
		title   string
		content string
	}{
		{
			name:  "implementation_task.md",
			title: "Implementation Task",
			content: "# Implementation Task\n\n" +
				"## Goal\n" + goal + "\n\n" +
				"## Delivery Contract\n" +
				"- Work inside the approved repository and keep changes scoped to this task.\n" +
				"- Preserve existing user changes and update owning docs when behavior changes.\n" +
				"- Keep tool caches, temp files, scratch scripts, and build intermediates out of the ultimate delivery and working tree; clean up any ephemeral files before handoff.\n" +
				"- Return implementation notes, changed files, and validation output.\n",
		},
		{
			name:  "validation_checklist.md",
			title: "Validation Checklist",
			content: "# Validation Checklist\n\n" +
				"- Confirm proposal approval happens before any workspace mutation.\n" +
				"- Confirm generated files are retained under the workspace package folder.\n" +
				"- Confirm no temporary files, tool caches, or stray debug logs remain in `git status` or the delivery package folder.\n" +
				"- Run focused unit tests for the touched runtime path.\n" +
				"- Run matching interface or browser proof when UI behavior changes.\n",
		},
		{
			name:  "affected_files_summary.md",
			title: "Affected Files Summary",
			content: "# Affected Files Summary\n\n" +
				"- Target package folder: `" + basePath + "`\n" +
				"- Source request: " + goal + "\n" +
				"- Expected implementation surface should be refined by the assigned Codex agent before edits.\n",
		},
		{
			name:  "next_recommended_slices.md",
			title: "Next Recommended Slices",
			content: "# Next Recommended Slices\n\n" +
				"1. Implement the smallest governed runtime change needed for the request.\n" +
				"2. Add proof that generated outputs remain visible after refresh or re-entry.\n" +
				"3. Promote package rendering into the dedicated Output Workbench when that UI lands.\n",
		},
	}

	calls := make([]protocol.PlannedToolCall, 0, len(files))
	for _, file := range files {
		calls = append(calls, protocol.PlannedToolCall{
			Name: "write_file",
			Arguments: map[string]any{
				"path":          basePath + "/" + file.name,
				"content":       file.content,
				"package_kind":  "codex_task_bundle",
				"package_slug":  slug,
				"artifact_role": file.title,
			},
		})
	}
	return calls, true
}

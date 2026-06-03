package server

import (
	"net/url"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

func teamOutputStorageRefFromExecutionOutput(output protocol.ExecutionOutput) string {
	kind := strings.TrimSpace(output.Kind)
	if storageRef := strings.TrimSpace(output.Folder); storageRef != "" {
		if path := workspacePathFromViewerURL(storageRef); path != "" {
			if kind == "project_package" {
				return firstNonEmptyString(parentWorkspacePath(path), path)
			}
			return path
		}
		return storageRef
	}
	if path := workspacePathFromViewerURL(output.Href); path != "" {
		if kind == "project_package" {
			return firstNonEmptyString(parentWorkspacePath(path), path)
		}
		return path
	}
	if entrypoint := strings.TrimSpace(output.Entrypoint); entrypoint != "" {
		if kind == "project_package" {
			return firstNonEmptyString(parentWorkspacePath(entrypoint), entrypoint)
		}
		return entrypoint
	}
	if looksLikeWorkspaceOutputPath(output.ID) {
		return strings.TrimSpace(output.ID)
	}
	return ""
}

func teamOutputStorageRefFromMap(kind string, data map[string]any) string {
	keys := []string{"storage_ref", "file_path", "path", "href", "open_url", "folder"}
	if strings.TrimSpace(kind) == "project_package" {
		keys = []string{"storage_ref", "folder", "file_path", "path", "href", "open_url"}
	}
	for _, key := range keys {
		if storageRef := teamOutputStorageRefFromRaw(kind, key, stringField(data, key)); storageRef != "" {
			return storageRef
		}
	}
	return ""
}

func teamOutputStorageRefFromRaw(kind, key, raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}
	fromViewerURL := false
	if path := workspacePathFromViewerURL(value); path != "" {
		value = path
		fromViewerURL = true
	}
	if strings.TrimSpace(kind) == "project_package" && (fromViewerURL || key == "href" || key == "open_url" || key == "file_path" || key == "path" || looksLikeWorkspaceFilePath(value)) {
		return firstNonEmptyString(parentWorkspacePath(value), value)
	}
	return value
}

func relativeTeamOutputEntrypoint(storageRef, entrypoint string) string {
	entry := strings.TrimSpace(entrypoint)
	if entry == "" {
		return ""
	}
	if path := workspacePathFromViewerURL(entry); path != "" {
		entry = path
	}
	storage := strings.TrimSpace(storageRef)
	if storage == "" || strings.HasPrefix(entry, "/") || strings.HasPrefix(entry, "http://") || strings.HasPrefix(entry, "https://") {
		return entry
	}
	normalizedStorage := strings.Trim(strings.ReplaceAll(storage, "\\", "/"), "/")
	normalizedEntry := strings.Trim(strings.ReplaceAll(entry, "\\", "/"), "/")
	if normalizedStorage != "" && strings.HasPrefix(normalizedEntry, normalizedStorage+"/") {
		return strings.TrimPrefix(normalizedEntry, normalizedStorage+"/")
	}
	return entry
}

func workspacePathFromViewerURL(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	parsed, err := url.Parse(trimmed)
	if err != nil || parsed.Path != "/api/v1/workspace/files/view" {
		return ""
	}
	return strings.TrimSpace(parsed.Query().Get("path"))
}

func looksLikeWorkspaceOutputPath(value string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return false
	}
	return strings.Contains(trimmed, "/") || strings.Contains(trimmed, "\\") || strings.Contains(trimmed, ".")
}

func looksLikeWorkspaceFilePath(value string) bool {
	trimmed := strings.Trim(strings.TrimSpace(value), `/\`)
	if trimmed == "" {
		return false
	}
	last := trimmed
	if cut := strings.LastIndexAny(last, `/\`); cut >= 0 {
		last = last[cut+1:]
	}
	return strings.Contains(last, ".")
}

package server

import (
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

const maxWorkspaceViewBytes = 2 << 20

func workspaceRoot() (string, error) {
	workspace := strings.TrimSpace(os.Getenv("MYCELIS_WORKSPACE"))
	if workspace == "" {
		workspace = "./workspace"
	}
	return filepath.Abs(workspace)
}

func normalizeWorkspacePath(rawPath string) string {
	trimmed := strings.TrimSpace(rawPath)
	if trimmed == "" {
		return ""
	}
	normalized := strings.ReplaceAll(trimmed, "\\", "/")
	normalized = path.Clean(normalized)
	normalized = strings.TrimPrefix(normalized, "./")
	if normalized == "/workspace" || strings.HasPrefix(normalized, "/workspace/") {
		normalized = strings.TrimPrefix(normalized, "/")
	} else if filepath.IsAbs(trimmed) {
		return trimmed
	}
	if normalized == "." || normalized == "workspace" {
		return "."
	}
	if strings.HasPrefix(normalized, "workspace/") {
		normalized = strings.TrimPrefix(normalized, "workspace/")
	}
	return filepath.FromSlash(normalized)
}

func resolveWorkspaceFilePath(rawPath string) (string, string, error) {
	return resolveWorkspacePath(rawPath, false)
}

func resolveWorkspacePath(rawPath string, allowRoot bool) (string, string, error) {
	root, err := workspaceRoot()
	if err != nil {
		return "", "", fmt.Errorf("invalid workspace root: %w", err)
	}
	normalized := normalizeWorkspacePath(rawPath)
	if normalized == "" {
		return "", "", fmt.Errorf("workspace file path is required")
	}
	if normalized == "." {
		if !allowRoot {
			return "", "", fmt.Errorf("workspace file path is required")
		}
		return root, ".", nil
	}

	var target string
	if filepath.IsAbs(normalized) {
		target = filepath.Clean(normalized)
	} else {
		target = filepath.Clean(filepath.Join(root, normalized))
	}
	rel, err := filepath.Rel(root, target)
	if err != nil || rel == "." || pathEscapesWorkspace(rel) {
		return "", "", fmt.Errorf("path escapes workspace boundary")
	}
	checkPath := target
	if _, statErr := os.Lstat(target); os.IsNotExist(statErr) {
		checkPath = filepath.Dir(target)
	}
	if realPath, evalErr := filepath.EvalSymlinks(checkPath); evalErr == nil {
		realRel, relErr := filepath.Rel(root, realPath)
		if relErr != nil || pathEscapesWorkspace(realRel) {
			return "", "", fmt.Errorf("path resolves outside workspace boundary")
		}
	}
	return target, filepath.ToSlash(rel), nil
}

func pathEscapesWorkspace(rel string) bool {
	return rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func workspaceFileContentType(target string) string {
	if ct := mime.TypeByExtension(strings.ToLower(filepath.Ext(target))); ct != "" {
		return ct
	}
	return "application/octet-stream"
}

// HandleWorkspaceFileView serves a generated workspace file for operator review.
// HTML is sandboxed so generated browser games can run without inheriting the app origin.
func (s *AdminServer) HandleWorkspaceFileView(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondAPIError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	target, rel, err := resolveWorkspaceFilePath(r.URL.Query().Get("path"))
	if err != nil {
		respondAPIError(w, err.Error(), http.StatusBadRequest)
		return
	}
	info, err := os.Stat(target)
	if err != nil {
		if os.IsNotExist(err) {
			respondAPIError(w, "workspace file not found", http.StatusNotFound)
			return
		}
		respondAPIError(w, "failed to inspect workspace file", http.StatusInternalServerError)
		return
	}
	if info.IsDir() {
		respondAPIError(w, "workspace path is a directory", http.StatusBadRequest)
		return
	}
	if info.Size() > maxWorkspaceViewBytes {
		respondAPIError(w, "workspace file is too large for inline review", http.StatusRequestEntityTooLarge)
		return
	}

	w.Header().Set("Content-Type", workspaceFileContentType(target))
	w.Header().Set("Content-Disposition", `inline; filename="`+strings.ReplaceAll(filepath.Base(target), `"`, "")+`"`)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Security-Policy", "sandbox allow-scripts; default-src 'none'; script-src 'unsafe-inline'; style-src 'unsafe-inline'; img-src data: blob:; connect-src 'none'; form-action 'none'; base-uri 'none'")
	w.Header().Set("X-Mycelis-Workspace-Path", rel)
	http.ServeFile(w, r, target)
}

// HandleWorkspaceFileReveal opens the containing workspace folder on the Core host.
// It is bounded to the workspace root and intended for local/operator proof flows.
func (s *AdminServer) HandleWorkspaceFileReveal(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireRootAdminScope(w, r, "host:invoke"); !ok {
		return
	}
	if r.Method != http.MethodPost {
		respondAPIError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	target, rel, err := resolveWorkspacePath(r.URL.Query().Get("path"), true)
	if err != nil {
		respondAPIError(w, err.Error(), http.StatusBadRequest)
		return
	}
	info, err := os.Stat(target)
	if err != nil {
		if os.IsNotExist(err) {
			respondAPIError(w, "workspace path not found", http.StatusNotFound)
			return
		}
		respondAPIError(w, "failed to inspect workspace path", http.StatusInternalServerError)
		return
	}
	folder := target
	if !info.IsDir() {
		folder = filepath.Dir(target)
	}
	if err := openLocalFolder(folder, target, info.IsDir()); err != nil {
		respondAPIError(w, "failed to open workspace folder: "+err.Error(), http.StatusBadGateway)
		return
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(map[string]any{
		"workspace_path": rel,
		"folder_path":    folder,
	}))
}

func openLocalFolder(folder, target string, isDir bool) error {
	if strings.EqualFold(os.Getenv("MYCELIS_WORKSPACE_REVEAL_DRY_RUN"), "1") {
		return nil
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		if isDir {
			cmd = exec.Command("explorer.exe", folder)
		} else {
			cmd = exec.Command("explorer.exe", "/select,"+target)
		}
	case "darwin":
		if isDir {
			cmd = exec.Command("open", folder)
		} else {
			cmd = exec.Command("open", "-R", target)
		}
	default:
		cmd = exec.Command("xdg-open", folder)
	}
	return cmd.Start()
}

func workspaceFileOutputHref(rawPath string) string {
	if strings.TrimSpace(rawPath) == "" {
		return ""
	}
	return "/api/v1/workspace/files/view?path=" + url.QueryEscape(rawPath)
}

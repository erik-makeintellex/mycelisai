package swarm

import (
	"encoding/json"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strings"
)

func ensureDeclaredHTMLPackageTitle(path, content string, args map[string]any) string {
	if !strings.EqualFold(strings.TrimSpace(stringValue(args["package_kind"])), "project_package") ||
		!strings.HasSuffix(strings.ToLower(strings.TrimSpace(path)), ".html") ||
		strings.Contains(strings.ToLower(content), "<title") {
		return content
	}
	title := strings.TrimSpace(stringValue(args["package_title"]))
	if title == "" {
		return content
	}
	titleElement := "<title>" + html.EscapeString(title) + "</title>"
	lower := strings.ToLower(content)
	if headIndex := strings.Index(lower, "<head"); headIndex >= 0 {
		if closeIndex := strings.Index(content[headIndex:], ">"); closeIndex >= 0 {
			insertAt := headIndex + closeIndex + 1
			return content[:insertAt] + titleElement + content[insertAt:]
		}
	}
	if htmlIndex := strings.Index(lower, "<html"); htmlIndex >= 0 {
		if closeIndex := strings.Index(content[htmlIndex:], ">"); closeIndex >= 0 {
			insertAt := htmlIndex + closeIndex + 1
			return content[:insertAt] + "<head>" + titleElement + "</head>" + content[insertAt:]
		}
	}
	return "<head>" + titleElement + "</head>" + content
}

// projectPackageManifestName is the only package file the runtime writes on
// its own. It is metadata, not evidence: README.md, PROOF.md and every other
// declared file must come from a real write, or the contract counts it missing.
const projectPackageManifestName = "project-package.json"

// projectPackageManifestMarker identifies a manifest the runtime wrote, so a
// later write may refresh it but never replaces a model-authored one.
const projectPackageManifestMarker = "mycelis_runtime"

// writeProjectPackageManifest writes project-package.json for a
// package_kind=project_package write. It returns the manifest's
// workspace-relative path, or "" when nothing was written.
func writeProjectPackageManifest(mainPath string, args map[string]any) (string, error) {
	if !strings.EqualFold(strings.TrimSpace(stringValue(args["package_kind"])), "project_package") {
		return "", nil
	}
	folder := projectPackageFolder(mainPath, args)
	if folder == "." || folder == "" {
		return "", nil
	}
	safeFolder, err := validateToolPath(folder)
	if err != nil {
		return "", err
	}
	target := filepath.Join(safeFolder, projectPackageManifestName)
	if safeMain, err := validateToolPath(mainPath); err == nil && filepath.Clean(safeMain) == filepath.Clean(target) {
		return "", nil
	}
	if existing, err := os.ReadFile(target); err == nil && !manifestWrittenByRuntime(existing) {
		return "", nil
	}
	if err := os.MkdirAll(safeFolder, 0o755); err != nil {
		return "", fmt.Errorf("failed to create package folder %s: %w", safeFolder, err)
	}
	if err := os.WriteFile(target, []byte(projectPackageManifestContent(args, mainPath)), 0o644); err != nil {
		return "", fmt.Errorf("failed to write package manifest %s: %w", target, err)
	}
	return strings.TrimRight(filepath.ToSlash(folder), "/") + "/" + projectPackageManifestName, nil
}

// projectPackageWrittenFiles names, relative to the package folder, the files
// one package write actually produced: the written file and the manifest.
func projectPackageWrittenFiles(folder, mainPath string) []string {
	files := []string{}
	cleanFolder := strings.Trim(filepath.ToSlash(normalizeWorkspaceRelativePath(folder)), "/")
	cleanMain := strings.Trim(filepath.ToSlash(normalizeWorkspaceRelativePath(mainPath)), "/")
	if rel := strings.TrimPrefix(cleanMain, cleanFolder+"/"); rel != cleanMain && rel != "" {
		files = append(files, rel)
	}
	return append(files, projectPackageManifestName)
}

func manifestWrittenByRuntime(data []byte) bool {
	var manifest map[string]any
	return json.Unmarshal(data, &manifest) == nil && manifest["generated_by"] == projectPackageManifestMarker
}

func projectPackageFolder(mainPath string, args map[string]any) string {
	if folder := strings.TrimSpace(stringValue(args["package_folder"])); folder != "" {
		return folder
	}
	return filepath.ToSlash(filepath.Dir(normalizeWorkspaceRelativePath(mainPath)))
}

func projectPackageDeclaredFiles(args map[string]any) []string {
	files := append([]string{}, stringSlice(args["package_files"])...)
	for _, file := range files {
		if strings.EqualFold(filepath.Base(strings.TrimSpace(file)), projectPackageManifestName) {
			return files
		}
	}
	return append(files, projectPackageManifestName)
}

// projectPackageManifestContent describes the declared package. It carries no
// validation claim: proof comes from readback and live validation only.
func projectPackageManifestContent(args map[string]any, mainPath string) string {
	title := strings.TrimSpace(stringValue(args["package_title"]))
	if title == "" {
		title = "Generated project package"
	}
	folder := projectPackageFolder(mainPath, args)
	entrypoint := strings.TrimSpace(stringValue(args["package_entrypoint"]))
	if entrypoint == "" {
		entrypoint = mainPath
	}
	payload := map[string]any{
		"title": title, "kind": "project_package", "entrypoint": entrypoint,
		"folder": folder, "declared_files": projectPackageDeclaredFiles(args),
		"generated_by": projectPackageManifestMarker,
		"open": map[string]any{
			"entrypoint": entrypoint, "resources_url": "/resources?tab=workspace&path=" + entrypointEscape(folder),
			"hint": "Open the entrypoint directly, or browse the folder from Resources -> Output Files.",
		},
	}
	if usage := strings.TrimSpace(firstProjectPackageString(args, "package_usage", "usage", "controls", "package_controls")); usage != "" {
		payload["usage"] = usage
	}
	data, _ := json.MarshalIndent(payload, "", "  ")
	return string(data) + "\n"
}

func firstProjectPackageString(args map[string]any, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(stringValue(args[key])); value != "" {
			return value
		}
	}
	return ""
}

func entrypointEscape(value string) string {
	replacer := strings.NewReplacer("%", "%25", " ", "%20", "#", "%23", "?", "%3F", "&", "%26")
	return replacer.Replace(strings.ReplaceAll(value, "\\", "/"))
}

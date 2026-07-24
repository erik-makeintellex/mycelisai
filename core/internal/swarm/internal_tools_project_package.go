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

func writeProjectPackageSupportFiles(mainPath string, args map[string]any) (int, error) {
	if !strings.EqualFold(strings.TrimSpace(stringValue(args["package_kind"])), "project_package") {
		return 0, nil
	}
	folder := strings.TrimSpace(stringValue(args["package_folder"]))
	if folder == "" {
		folder = filepath.Dir(normalizeWorkspaceRelativePath(mainPath))
	}
	if folder == "." || folder == "" {
		return 0, nil
	}
	safeFolder, err := validateToolPath(folder)
	if err != nil {
		return 0, err
	}
	if err := os.MkdirAll(safeFolder, 0o755); err != nil {
		return 0, fmt.Errorf("failed to create package folder %s: %w", safeFolder, err)
	}

	count := 0
	for _, file := range projectPackageSupportFileNames(args) {
		rel := strings.Trim(strings.TrimSpace(file), `/\`)
		if rel == "" {
			continue
		}
		base := strings.ToLower(filepath.Base(rel))
		if base != "readme.md" && base != "proof.md" && base != "validation-notes.md" && base != "project-package.json" {
			continue
		}
		cleanRel := filepath.Clean(filepath.FromSlash(rel))
		if filepath.IsAbs(cleanRel) || cleanRel == "." || strings.HasPrefix(cleanRel, "..") {
			return count, fmt.Errorf("package support file %q escapes package folder", file)
		}
		target := filepath.Join(safeFolder, cleanRel)
		relToFolder, err := filepath.Rel(safeFolder, target)
		if err != nil || strings.HasPrefix(relToFolder, "..") {
			return count, fmt.Errorf("package support file %q escapes package folder", file)
		}
		content := projectPackageSupportFileContent(base, args, mainPath)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return count, fmt.Errorf("failed to create package support folder %s: %w", filepath.Dir(target), err)
		}
		if err := os.WriteFile(target, []byte(content), 0o644); err != nil {
			return count, fmt.Errorf("failed to write package support file %s: %w", target, err)
		}
		count++
	}
	return count, nil
}

func projectPackageSupportFileNames(args map[string]any) []string {
	files := append([]string{}, stringSlice(args["package_files"])...)
	for _, file := range files {
		if strings.EqualFold(filepath.Base(strings.TrimSpace(file)), "project-package.json") {
			return files
		}
	}
	return append(files, "project-package.json")
}

func projectPackageSupportFileContent(file string, args map[string]any, mainPath string) string {
	title := strings.TrimSpace(stringValue(args["package_title"]))
	if title == "" {
		title = "Generated project package"
	}
	folder := strings.TrimSpace(stringValue(args["package_folder"]))
	if folder == "" {
		folder = filepath.Dir(normalizeWorkspaceRelativePath(mainPath))
	}
	entrypoint := strings.TrimSpace(stringValue(args["package_entrypoint"]))
	if entrypoint == "" {
		entrypoint = mainPath
	}
	files := projectPackageSupportFileNames(args)
	validation := strings.TrimSpace(firstProjectPackageString(args, "validation", "validation_summary"))
	if validation == "" {
		validation = "Open the entrypoint in a browser and review the retained output."
	}
	usage := strings.TrimSpace(firstProjectPackageString(args, "package_usage", "usage", "controls", "package_controls"))
	if usage == "" && strings.HasSuffix(strings.ToLower(entrypoint), ".html") {
		usage = "Open the HTML entrypoint in a browser. If it is interactive, use the visible controls in the page."
	}
	recovery := strings.TrimSpace(firstProjectPackageString(args, "recovery", "recovery_hint", "open_hint", "package_recovery"))
	if recovery == "" {
		recovery = "If opening fails, use Resources -> Output Files to browse the package folder, confirm the entrypoint exists, then ask Soma to repair or regenerate the package."
	}
	if file == "project-package.json" {
		payload := map[string]any{
			"title": title, "kind": "project_package", "entrypoint": entrypoint,
			"folder": folder, "files": files, "validation": validation,
			"open": map[string]any{
				"entrypoint": entrypoint, "resources_url": "/resources?tab=workspace&path=" + entrypointEscape(folder),
				"hint": "Open the entrypoint directly, or browse the folder from Resources -> Output Files.",
			},
			"recovery": map[string]any{"hint": recovery},
		}
		if usage != "" {
			payload["usage"] = usage
		}
		data, _ := json.MarshalIndent(payload, "", "  ")
		return string(data) + "\n"
	}
	includedFiles := "- " + strings.Join(files, "\n- ")
	if file == "proof.md" || file == "validation-notes.md" {
		return fmt.Sprintf("# %s Proof\n\n## Open\n\nOpen `%s`.\n\n## Included files\n\n%s\n\n## Validation\n\n%s\n\n## Recovery\n\n%s\n", title, entrypoint, includedFiles, validation, recovery)
	}
	usageSection := ""
	if usage != "" {
		usageSection = fmt.Sprintf("\n## Usage / controls\n\n%s\n", usage)
	}
	return fmt.Sprintf("# %s\n\n## Open\n\nOpen `%s`.\n\n## Included files\n\n%s\n%s\n## Validation\n\n%s\n\n## Recovery\n\n%s\n", title, entrypoint, includedFiles, usageSection, validation, recovery)
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

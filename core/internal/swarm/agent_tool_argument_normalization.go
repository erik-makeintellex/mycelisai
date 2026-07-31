package swarm

import (
	"encoding/json"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

var safeTeamIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)

func normalizeAgentToolCallArguments(call *toolCallPayload, teamID, latestUserInput string) {
	autofillToolArguments(call, latestUserInput)
	normalizeTeamOwnedWriteArguments(call, teamID, latestUserInput)
	normalizeTeamOwnedArtifactArguments(call, teamID)
	// Team ownership normalization can change the package entrypoint and clear
	// stale package hints, so infer those fields again from the canonical path.
	autofillToolArguments(call, latestUserInput)
}

func normalizeTeamOwnedArtifactArguments(call *toolCallPayload, teamID string) {
	if call == nil || call.Name != "store_artifact" || !safeTeamIDPattern.MatchString(strings.TrimSpace(teamID)) ||
		!strings.EqualFold(strings.TrimSpace(stringValue(call.Arguments["type"])), "project_package") {
		return
	}
	normalizeProjectPackageArtifactArguments(call.Arguments)
	metadata, _ := call.Arguments["metadata"].(map[string]any)
	if metadata == nil {
		metadata = map[string]any{"package_kind": "project_package"}
	}
	folder := canonicalTeamGeneratedArtifactPath(stringValue(metadata["folder"]), teamID)
	if folder == "" {
		folder = "groups/" + strings.TrimSpace(teamID) + "/generated/package"
	}
	metadata["folder"] = folder
	entrypoint := strings.ReplaceAll(strings.TrimSpace(stringValue(metadata["entrypoint"])), "\\", "/")
	if entrypoint == "" {
		entrypoint = "index.html"
	} else if strings.Contains(entrypoint, "/") {
		entrypoint = canonicalTeamGeneratedArtifactPath(entrypoint, teamID)
	}
	metadata["entrypoint"] = entrypoint
	call.Arguments["metadata"] = metadata
}

func canonicalTeamGeneratedArtifactPath(value, teamID string) string {
	value = strings.Trim(strings.ReplaceAll(strings.TrimSpace(value), "\\", "/"), "/")
	if value == "" || strings.HasPrefix(value, "../") || strings.Contains(value, "/../") {
		return ""
	}
	teamID = strings.TrimSpace(teamID)
	teamRoot := "groups/" + teamID + "/"
	if index := strings.Index(strings.ToLower(value), strings.ToLower(teamRoot)); index >= 0 {
		value = value[index+len(teamRoot):]
	}
	for _, prefix := range []string{"workspace/", "groups/" + strings.ToLower(teamID) + "/", strings.ToLower(teamID) + "/", "generated/", "output/", "outputs/"} {
		if strings.HasPrefix(strings.ToLower(value), prefix) {
			value = value[len(prefix):]
		}
	}
	if index := strings.Index(strings.ToLower(value), "/generated/"); index >= 0 {
		value = value[index+len("/generated/"):]
	}
	value = strings.Trim(value, "/")
	if value == "" || strings.HasPrefix(value, "../") || strings.Contains(value, "/../") {
		return ""
	}
	return teamRoot + "generated/" + value
}

func normalizeTeamOwnedWriteArguments(call *toolCallPayload, teamID, latestUserInput string) {
	if call == nil || call.Name != "write_file" || !safeTeamIDPattern.MatchString(strings.TrimSpace(teamID)) {
		return
	}
	packageRequest := requestsProjectPackage(strings.ToLower(latestUserInput))
	path := strings.ReplaceAll(strings.TrimSpace(stringValue(call.Arguments["path"])), "\\", "/")
	if path == "" {
		return
	}
	teamRoot := "groups/" + strings.TrimSpace(teamID) + "/"
	if index := strings.Index(strings.ToLower(path), strings.ToLower(teamRoot)); index >= 0 {
		path = strings.TrimLeft(path[index:], "/")
		suffix := strings.TrimPrefix(path, teamRoot)
		finalOutputPath := false
		for _, prefix := range []string{"generated/", "output/", "outputs/"} {
			if strings.HasPrefix(strings.ToLower(suffix), prefix) {
				suffix = suffix[len(prefix):]
				finalOutputPath = true
				break
			}
		}
		if !finalOutputPath {
			call.Arguments["path"] = path
			return
		}
		if suffix == "" || strings.HasPrefix(suffix, "../") || strings.Contains(suffix, "/../") {
			return
		}
		call.Arguments["path"] = teamRoot + "generated/" + suffix
		delete(call.Arguments, "package_entrypoint")
		delete(call.Arguments, "package_folder")
		return
	}

	suffix := strings.Trim(path, "/")
	if index := strings.Index(strings.ToLower(suffix), "/generated/"); index >= 0 {
		suffix = suffix[index+len("/generated/"):]
	}
	for _, prefix := range []string{"workspace/", "generated/", "output/", "outputs/"} {
		if strings.HasPrefix(strings.ToLower(suffix), prefix) {
			suffix = suffix[len(prefix):]
		}
	}
	if !packageRequest && !strings.HasPrefix(strings.ToLower(strings.Trim(path, "/")), "output/") &&
		!strings.HasPrefix(strings.ToLower(strings.Trim(path, "/")), "outputs/") &&
		!strings.HasPrefix(strings.ToLower(strings.Trim(path, "/")), "generated/") {
		return
	}
	if suffix == "" || strings.HasPrefix(suffix, "../") || strings.Contains(suffix, "/../") {
		return
	}
	call.Arguments["path"] = teamRoot + "generated/" + suffix
	delete(call.Arguments, "package_entrypoint")
	delete(call.Arguments, "package_folder")
}

func inferProjectPackageWriteArguments(arguments map[string]any, latestUserInput string) {
	path, _ := arguments["path"].(string)
	normalizedPath := strings.ReplaceAll(strings.TrimSpace(path), "\\", "/")
	lowerInput := strings.ToLower(latestUserInput)
	if !requestsProjectPackage(lowerInput) ||
		!strings.Contains(strings.ToLower(normalizedPath), "/generated/") ||
		!strings.HasSuffix(strings.ToLower(normalizedPath), ".html") {
		return
	}
	if strings.TrimSpace(stringValue(arguments["package_kind"])) == "" {
		arguments["package_kind"] = "project_package"
	}
	if strings.TrimSpace(stringValue(arguments["package_entrypoint"])) == "" {
		arguments["package_entrypoint"] = normalizedPath
	}
	if strings.TrimSpace(stringValue(arguments["package_folder"])) == "" {
		if index := strings.LastIndex(normalizedPath, "/"); index > 0 {
			arguments["package_folder"] = normalizedPath[:index]
		}
	}
	if len(stringSlice(arguments["package_files"])) == 0 {
		arguments["package_files"] = []string{pathBase(normalizedPath), "README.md", "PROOF.md", "project-package.json"}
	}
	if strings.TrimSpace(stringValue(arguments["package_title"])) == "" {
		if title := extractRequestedPackageTitle(latestUserInput); title != "" {
			arguments["package_title"] = title
		}
	}
}

func requestsProjectPackage(lowerInput string) bool {
	for _, phrase := range []string{
		"project_package",
		"project package",
		"application package",
		"app package",
		"browser application",
		"browser app",
		"playable application",
		"playable app",
		"playable browser",
	} {
		if strings.Contains(lowerInput, phrase) {
			return true
		}
	}
	return false
}

func pathBase(path string) string {
	if index := strings.LastIndex(path, "/"); index >= 0 {
		return path[index+1:]
	}
	return path
}

func projectPackageArtifactFromSuccessfulWrite(arguments map[string]any, latestUserInput string) (protocol.ChatArtifactRef, bool) {
	path := strings.ReplaceAll(strings.TrimSpace(stringValue(arguments["path"])), "\\", "/")
	if path == "" || !strings.Contains(strings.ToLower(path), "/generated/") ||
		!strings.HasSuffix(strings.ToLower(path), ".html") ||
		(!strings.EqualFold(strings.TrimSpace(stringValue(arguments["package_kind"])), "project_package") &&
			!requestsProjectPackage(strings.ToLower(latestUserInput))) {
		return protocol.ChatArtifactRef{}, false
	}

	folder := strings.TrimSpace(stringValue(arguments["package_folder"]))
	if folder == "" {
		folder = strings.ReplaceAll(filepath.Dir(path), "\\", "/")
	}
	entrypoint := strings.TrimSpace(stringValue(arguments["package_entrypoint"]))
	if entrypoint == "" {
		entrypoint = path
	}
	files := stringSlice(arguments["package_files"])
	if len(files) == 0 {
		files = []string{pathBase(entrypoint), "README.md", "PROOF.md", "project-package.json"}
	}
	title := strings.TrimSpace(stringValue(arguments["package_title"]))
	if title == "" {
		title = "Generated application package"
	}
	validation := strings.TrimSpace(firstProjectPackageString(arguments, "validation", "validation_summary"))
	if validation == "" {
		validation = "Package entrypoint retained; open it to complete operator interaction validation."
	}
	payload, _ := json.Marshal(map[string]any{
		"title": title, "kind": "project_package", "entrypoint": entrypoint,
		"folder": folder, "files": files, "validation": validation,
	})
	return protocol.ChatArtifactRef{
		Type:        "project_package",
		Title:       title,
		ContentType: "application/vnd.mycelis.project+json",
		Content:     string(payload),
		SavedPath:   folder,
		Entrypoint:  entrypoint,
		Folder:      folder,
		Files:       files,
		Validation:  validation,
	}, true
}

func normalizeProjectPackageArtifactArguments(arguments map[string]any) {
	if !strings.EqualFold(strings.TrimSpace(stringValue(arguments["type"])), "project_package") {
		return
	}

	manifest := map[string]any{}
	switch content := arguments["content"].(type) {
	case map[string]any:
		for key, value := range content {
			manifest[key] = value
		}
		if encoded, err := json.Marshal(content); err == nil {
			arguments["content"] = string(encoded)
		}
	case string:
		_ = json.Unmarshal([]byte(strings.TrimSpace(content)), &manifest)
	}

	metadata, _ := arguments["metadata"].(map[string]any)
	if metadata == nil {
		metadata = map[string]any{}
	}
	metadata["package_kind"] = "project_package"
	for _, key := range []string{"entrypoint", "folder", "files", "validation"} {
		if _, exists := metadata[key]; exists {
			continue
		}
		if value, exists := manifest[key]; exists {
			metadata[key] = value
		}
	}
	if validation, ok := metadata["validation"].(map[string]any); ok {
		if encoded, err := json.Marshal(validation); err == nil {
			metadata["validation"] = string(encoded)
		}
	}
	arguments["metadata"] = metadata

	if strings.TrimSpace(stringValue(arguments["title"])) == "" {
		arguments["title"] = firstNonEmptyString(
			stringValue(manifest["title"]),
			stringValue(arguments["key"]),
			"Generated application package",
		)
	}
}

func extractRequestedPackageTitle(input string) string {
	const marker = "use the package title "
	lower := strings.ToLower(input)
	index := strings.Index(lower, marker)
	if index < 0 {
		return ""
	}
	rest := strings.TrimSpace(input[index+len(marker):])
	for _, delimiter := range []string{"\n", ". ", ".\n"} {
		if cut := strings.Index(rest, delimiter); cut >= 0 {
			rest = rest[:cut]
		}
	}
	return strings.Trim(strings.TrimSpace(rest), `"'.,;`)
}

func copyStringToolArgumentAlias(arguments map[string]any, target string, aliases ...string) {
	if current, _ := arguments[target].(string); strings.TrimSpace(current) != "" {
		return
	}
	for _, alias := range aliases {
		value, _ := arguments[alias].(string)
		if strings.TrimSpace(value) == "" {
			continue
		}
		arguments[target] = value
		return
	}
}

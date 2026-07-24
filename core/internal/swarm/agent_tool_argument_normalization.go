package swarm

import "strings"

func inferProjectPackageWriteArguments(arguments map[string]any, latestUserInput string) {
	path, _ := arguments["path"].(string)
	normalizedPath := strings.ReplaceAll(strings.TrimSpace(path), "\\", "/")
	lowerInput := strings.ToLower(latestUserInput)
	if !strings.Contains(lowerInput, "project_package") ||
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
		arguments["package_files"] = []string{"index.html", "README.md", "PROOF.md", "project-package.json"}
	}
	if strings.TrimSpace(stringValue(arguments["package_title"])) == "" {
		if title := extractRequestedPackageTitle(latestUserInput); title != "" {
			arguments["package_title"] = title
		}
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

package server

import "strings"

func extractRequestedPackageFolder(input string) string {
	lower := strings.ToLower(input)
	for _, marker := range []string{"retain it at ", "retain at ", "save it at ", "save at "} {
		index := strings.Index(lower, marker)
		if index < 0 {
			continue
		}
		rest := strings.TrimSpace(input[index+len(marker):])
		for _, delimiter := range []string{" with entrypoint ", ". ", ".\n", "\n"} {
			if cut := strings.Index(strings.ToLower(rest), delimiter); cut >= 0 {
				rest = rest[:cut]
			}
		}
		if folder := cleanWorkspaceTargetToken(rest); folder != "" {
			return folder
		}
	}
	return ""
}

func extractRequestedPackageEntrypoint(input string) string {
	lower := strings.ToLower(input)
	for _, marker := range []string{"with entrypoint ", "entrypoint "} {
		index := strings.Index(lower, marker)
		if index < 0 {
			continue
		}
		rest := strings.TrimSpace(input[index+len(marker):])
		for _, delimiter := range []string{". ", ".\n", "\n"} {
			if cut := strings.Index(rest, delimiter); cut >= 0 {
				rest = rest[:cut]
			}
		}
		if entrypoint := cleanWorkspaceTargetToken(rest); entrypoint != "" {
			return entrypoint
		}
	}
	return ""
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

func cleanWorkspaceTargetToken(value string) string {
	value = strings.Trim(strings.TrimSpace(value), `"'.,;`)
	value = strings.ReplaceAll(value, "\\", "/")
	value = strings.Trim(value, "/")
	if value == "" || strings.HasPrefix(value, "../") || strings.Contains(value, "/../") {
		return ""
	}
	return value
}

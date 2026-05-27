package server

import (
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

func executionOutputsFromToolResults(results []plannedToolExecutionResult) []protocol.ExecutionOutput {
	outputs := make([]protocol.ExecutionOutput, 0, len(results))
	for _, result := range results {
		toolName := strings.TrimSpace(result.Name)
		if toolName == "" {
			toolName = "tool"
		}
		id := confirmedActionTeamID(result.Arguments)
		kind := "tool_result"
		title := toolName
		retained := false
		href := ""
		if len(result.Artifacts) > 0 {
			for _, output := range executionOutputsFromArtifacts(mergeArtifactRefsWithArgs(result.Artifacts, result.Arguments)) {
				output.Summary = firstNonEmptyString(output.Summary, result.Output, toolName+" completed.")
				outputs = append(outputs, output)
			}
			if toolName == "store_artifact" {
				continue
			}
			if projectPackageOutputFromArgs(result.Arguments) != nil {
				continue
			}
		}
		if toolName == "create_team" {
			kind = "team"
			title = firstNonEmptyString(mergedTeamArgs(result.Arguments)["name"], id, "Created team")
			retained = true
			href = "/groups"
		}
		if toolName == "write_file" {
			path := firstNonEmptyString(result.Arguments["path"], result.Arguments["file_path"], result.Arguments["target_path"])
			id = path
			kind = outputKindForWrittenFile(path)
			title = firstNonEmptyString(path, "Workspace file")
			retained = true
			href = workspaceFileOutputHref(path)
		}
		packageOutput := projectPackageOutputFromArgs(result.Arguments)
		if packageOutput != nil {
			id = packageOutput.ID
			kind = packageOutput.Kind
			title = packageOutput.Title
			href = packageOutput.Href
			retained = true
		}
		if strings.TrimSpace(result.ToolRef) != "" {
			id = firstNonEmptyString(result.ToolRef, toolName)
			kind = "mcp_tool_result"
			title = firstNonEmptyString(result.ToolRef, toolName)
			retained = true
			if path := firstNonEmptyString(result.Arguments["path"], result.Arguments["file_path"], result.Arguments["target_path"]); path != "" {
				href = workspaceFileOutputHref(path)
			}
		}
		outputs = append(outputs, protocol.ExecutionOutput{
			ID:             id,
			Kind:           kind,
			Title:          title,
			Summary:        firstNonEmptyString(result.Output, toolName+" completed."),
			Href:           href,
			Entrypoint:     outputEntrypoint(packageOutput),
			Folder:         firstNonEmptyString(outputFolder(packageOutput), outputFolderForTeam(toolName, id)),
			Files:          outputFiles(packageOutput),
			Validation:     outputValidation(packageOutput),
			Retained:       boolPtr(retained),
			RetentionClass: retentionClassForBool(retained),
		})
	}
	return outputs
}

func outputFolderForTeam(toolName, teamID string) string {
	if strings.TrimSpace(toolName) != "create_team" {
		return ""
	}
	return groupWorkspaceFolderForTeamID(teamID)
}

func mergeArtifactRefsWithArgs(artifacts []protocol.ChatArtifactRef, args map[string]any) []protocol.ChatArtifactRef {
	if len(artifacts) == 0 {
		return nil
	}
	packageOutput := projectPackageOutputFromArgs(args)
	merged := make([]protocol.ChatArtifactRef, 0, len(artifacts))
	for _, artifact := range artifacts {
		if packageOutput != nil && (artifact.Type == "project_package" || artifact.Type == "file" || artifact.Type == "data") {
			artifact.Type = "project_package"
			artifact.Entrypoint = firstNonEmptyString(artifact.Entrypoint, packageOutput.Entrypoint)
			artifact.Folder = firstNonEmptyString(artifact.Folder, packageOutput.Folder)
			artifact.Files = mergeProjectPackageFiles(artifact.Files, packageOutput.Files)
			artifact.Validation = firstNonEmptyString(artifact.Validation, packageOutput.Validation)
			artifact.Title = firstNonEmptyString(packageOutput.Title, artifact.Title)
		}
		merged = append(merged, artifact)
	}
	return merged
}

func outputKindForWrittenFile(path string) string {
	switch filepathExt(path) {
	case ".css", ".go", ".html", ".js", ".json", ".jsx", ".py", ".ts", ".tsx":
		return "code"
	default:
		return "file"
	}
}

func projectPackageOutputFromArgs(args map[string]any) *protocol.ExecutionOutput {
	kind := strings.TrimSpace(firstNonEmptyString(args["package_kind"], args["output_kind"], args["kind"], args["type"], args["artifact_type"]))
	if kind != "project_package" {
		return nil
	}
	entrypoint := firstNonEmptyString(args["entrypoint"], args["package_entrypoint"], args["path"], args["file_path"], args["target_path"])
	folder := firstNonEmptyString(args["folder"], args["package_folder"], args["project_folder"])
	if folder == "" {
		folder = parentWorkspacePath(entrypoint)
	}
	title := firstNonEmptyString(args["package_title"], args["title"], folder, entrypoint, "Generated project package")
	id := firstNonEmptyString(args["package_id"], args["id"], folder, entrypoint, title)
	return &protocol.ExecutionOutput{
		ID:         id,
		Kind:       "project_package",
		Title:      title,
		Href:       workspaceFileOutputHref(entrypoint),
		Entrypoint: entrypoint,
		Folder:     folder,
		Files:      projectPackageFilesFromArgs(args),
		Validation: firstNonEmptyString(args["validation"], args["validation_summary"], args["proof_summary"]),
	}
}

func projectPackageFilesFromArgs(args map[string]any) []string {
	files := firstStringSliceArgument(args["files"], args["package_files"])
	if projectPackageArgsRequireReadme(args) {
		files = mergeProjectPackageFiles(files, []string{"README.md"})
	}
	return files
}

func mergeProjectPackageFiles(primary, secondary []string) []string {
	out := make([]string, 0, len(primary)+len(secondary))
	for _, file := range append(append([]string{}, primary...), secondary...) {
		file = strings.TrimSpace(file)
		if file == "" {
			continue
		}
		seen := false
		for _, existing := range out {
			if strings.EqualFold(existing, file) {
				seen = true
				break
			}
		}
		if seen {
			continue
		}
		out = append(out, file)
	}
	return out
}

func projectPackageArgsRequireReadme(args map[string]any) bool {
	for _, key := range []string{
		"content",
		"body",
		"text",
		"prompt",
		"request",
		"contract",
		"package_contract",
		"validation",
		"validation_summary",
		"proof_summary",
	} {
		if projectPackageTextRequestsReadme(firstNonEmptyString(args[key])) {
			return true
		}
	}
	return false
}

func projectPackageTextRequestsReadme(text string) bool {
	normalized := strings.ToLower(strings.TrimSpace(text))
	return strings.Contains(normalized, "readme")
}

func firstStringSliceArgument(values ...any) []string {
	for _, value := range values {
		if items := confirmedActionStringSlice(value); len(items) > 0 {
			return items
		}
	}
	return nil
}

func parentWorkspacePath(path string) string {
	trimmed := strings.Trim(strings.TrimSpace(path), `/\`)
	if trimmed == "" {
		return ""
	}
	cut := strings.LastIndexAny(trimmed, `/\`)
	if cut < 0 {
		return ""
	}
	return trimmed[:cut]
}

func outputEntrypoint(output *protocol.ExecutionOutput) string {
	if output == nil {
		return ""
	}
	return output.Entrypoint
}

func outputFolder(output *protocol.ExecutionOutput) string {
	if output == nil {
		return ""
	}
	return output.Folder
}

func outputFiles(output *protocol.ExecutionOutput) []string {
	if output == nil {
		return nil
	}
	return output.Files
}

func outputValidation(output *protocol.ExecutionOutput) string {
	if output == nil {
		return ""
	}
	return output.Validation
}

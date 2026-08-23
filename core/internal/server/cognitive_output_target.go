package server

import (
	"fmt"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

func firstPlannedOutputTarget(planned []protocol.PlannedToolCall) string {
	for _, call := range planned {
		if strings.TrimSpace(call.Name) == "save_cached_image" {
			if target := workspaceMediaTarget(call.Arguments); target != "" {
				return target
			}
		}
	}
	if target := firstPlannedDelegateResultTarget(planned); target != "" {
		return target
	}
	for _, call := range planned {
		if strings.TrimSpace(call.Name) == "write_file" {
			target := firstNonEmptyString(call.Arguments["path"], call.Arguments["package_entrypoint"], call.Arguments["package_folder"])
			if target != "" && !strings.Contains(strings.ToLower(target), "/planning/") {
				return target
			}
		}
	}
	for _, call := range planned {
		if strings.TrimSpace(call.Name) == "write_file" {
			if target := firstNonEmptyString(call.Arguments["path"], call.Arguments["package_entrypoint"], call.Arguments["package_folder"]); target != "" {
				return target
			}
		}
		if strings.TrimSpace(call.Name) == "generate_image" {
			if target := firstNonEmptyString(call.Arguments["goal"], call.Arguments["prompt"]); target != "" {
				return target
			}
		}
	}
	return ""
}

func firstPlannedDelegateResultTarget(planned []protocol.PlannedToolCall) string {
	for _, call := range planned {
		if strings.TrimSpace(call.Name) != "delegate_task" && strings.TrimSpace(call.Name) != "delegate" {
			continue
		}
		ask := mapArgument(call.Arguments["ask"])
		context := mapArgument(ask["context"])
		contract := mapArgument(context["result_contract"])
		for _, key := range []string{"package_entrypoint", "entrypoint", "package_folder"} {
			if target := firstNonEmptyString(contract[key]); target != "" {
				return target
			}
		}
	}
	return ""
}

func workspaceMediaTarget(arguments map[string]any) string {
	folder := strings.Trim(strings.TrimSpace(fmt.Sprint(arguments["folder"])), "/\\")
	filename := strings.Trim(strings.TrimSpace(fmt.Sprint(arguments["filename"])), "/\\")
	if filename == "" {
		return folder
	}
	if folder == "" {
		return filename
	}
	return folder + "/" + filename
}

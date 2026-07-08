package server

import "github.com/mycelis/core/pkg/protocol"

func outputClassFromArguments(args map[string]any) protocol.OutputClass {
	return protocol.NormalizeOutputClass(firstNonEmptyString(
		args["output_class"],
		args["delivery_status"],
		args["visibility"],
		args["retention_class"],
	))
}

func firstNonEmptyOutputClass(values ...protocol.OutputClass) protocol.OutputClass {
	for _, value := range values {
		if normalized := protocol.NormalizeOutputClass(string(value)); normalized != "" {
			return normalized
		}
	}
	return ""
}

func outputClassForToolResult(toolName, kind, title, path string, args map[string]any) protocol.OutputClass {
	if class := outputClassFromArguments(args); class != "" {
		return class
	}
	return protocol.InferOutputClass(kind, path, title, toolName)
}

func outputClassForExecutionOutput(output protocol.ExecutionOutput) protocol.OutputClass {
	if output.OutputClass != "" {
		return protocol.NormalizeOutputClass(string(output.OutputClass))
	}
	return protocol.InferOutputClass(
		output.Kind,
		output.Folder,
		output.Entrypoint,
		output.Href,
		output.OpenURL,
		output.Title,
		output.ID,
	)
}

func outputClassForTeamRef(ref protocol.TeamOutputRef) protocol.OutputClass {
	if ref.OutputClass != "" {
		return protocol.NormalizeOutputClass(string(ref.OutputClass))
	}
	return protocol.InferOutputClass(
		ref.Kind,
		ref.StorageRef,
		ref.Entrypoint,
		ref.Label,
		ref.OutputID,
	)
}

func isUserDeliverableTeamOutputRef(ref protocol.TeamOutputRef) bool {
	return outputClassForTeamRef(ref) == protocol.OutputClassUserDeliverable
}

func outputClassFromRawMap(data map[string]any) protocol.OutputClass {
	return protocol.NormalizeOutputClass(firstNonEmptyString(
		data["output_class"],
		data["delivery_status"],
		data["visibility"],
		data["retention_class"],
	))
}

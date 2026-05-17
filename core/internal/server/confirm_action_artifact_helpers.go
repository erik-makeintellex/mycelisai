package server

import (
	"strings"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/artifacts"
	"github.com/mycelis/core/pkg/protocol"
)

func outputTitle(output *protocol.ExecutionOutput) string {
	if output == nil {
		return ""
	}
	return output.Title
}

func artifactTypeForWrittenFile(path string) artifacts.ArtifactType {
	if outputKindForWrittenFile(path) == "code" {
		return artifacts.TypeCode
	}
	return artifacts.TypeFile
}

func artifactTypeForConfirmedWrite(args map[string]any, path string) artifacts.ArtifactType {
	if projectPackageOutputFromArgs(args) != nil {
		return artifacts.TypeProjectPackage
	}
	return artifactTypeForWrittenFile(path)
}

func contentTypeForWrittenFile(path string) string {
	switch filepathExt(path) {
	case ".css":
		return "text/css"
	case ".html":
		return "text/html"
	case ".json":
		return "application/json"
	case ".md":
		return "text/markdown"
	case ".js", ".jsx", ".ts", ".tsx":
		return "text/javascript"
	default:
		return "text/plain"
	}
}

func uuidPtrFromString(raw string) *uuid.UUID {
	parsed, err := uuid.Parse(strings.TrimSpace(raw))
	if err != nil || parsed == uuid.Nil {
		return nil
	}
	return &parsed
}

func floatPtr(value float64) *float64 {
	return &value
}

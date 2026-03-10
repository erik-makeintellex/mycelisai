package swarm

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const signalCheckpointChannelPrefix = "signal.latest."

func parseOptionalBool(args map[string]any, key string, defaultValue bool) bool {
	if args == nil {
		return defaultValue
	}
	raw, ok := args[key]
	if !ok {
		return defaultValue
	}

	switch v := raw.(type) {
	case bool:
		return v
	case string:
		trimmed := strings.ToLower(strings.TrimSpace(v))
		switch trimmed {
		case "true", "1", "yes", "y", "on":
			return true
		case "false", "0", "no", "n", "off":
			return false
		default:
			return defaultValue
		}
	case float64:
		return v != 0
	case int:
		return v != 0
	case int64:
		return v != 0
	case json.Number:
		i, err := strconv.ParseInt(v.String(), 10, 64)
		if err != nil {
			return defaultValue
		}
		return i != 0
	default:
		return defaultValue
	}
}

func resolveSignalCheckpointChannelKey(subject string, args map[string]any) string {
	if args != nil {
		if raw, ok := args["channel_key"].(string); ok && strings.TrimSpace(raw) != "" {
			return strings.TrimSpace(raw)
		}
		if raw, ok := args["private_channel"].(string); ok && strings.TrimSpace(raw) != "" {
			return strings.TrimSpace(raw)
		}
	}

	trimmedSubject := strings.TrimSpace(subject)
	if trimmedSubject == "" {
		return ""
	}
	return signalCheckpointChannelPrefix + trimmedSubject
}

func buildWorkspaceFileReference(rawPath string) (map[string]any, error) {
	if strings.TrimSpace(rawPath) == "" {
		return nil, nil
	}
	targetPath, err := validateToolPath(rawPath)
	if err != nil {
		return nil, err
	}
	relPath, err := workspaceRelativePath(targetPath)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"kind": "workspace_file",
		"path": relPath,
	}, nil
}

func workspaceRelativePath(absTarget string) (string, error) {
	workspace := os.Getenv("MYCELIS_WORKSPACE")
	if workspace == "" {
		workspace = "./workspace"
	}
	absWorkspace, err := filepath.Abs(workspace)
	if err != nil {
		return "", fmt.Errorf("resolve workspace: %w", err)
	}
	rel, err := filepath.Rel(absWorkspace, absTarget)
	if err != nil {
		return "", fmt.Errorf("relative path: %w", err)
	}
	if strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("path escapes workspace")
	}
	return filepath.ToSlash(rel), nil
}

func (r *InternalToolRegistry) upsertSignalCheckpoint(
	ctx context.Context,
	channelKey string,
	ownerAgentID string,
	content string,
	metadata map[string]any,
) (string, error) {
	if r == nil || r.mem == nil {
		return "", nil
	}
	channelKey = strings.TrimSpace(channelKey)
	if channelKey == "" {
		return "", nil
	}
	if strings.TrimSpace(ownerAgentID) == "" {
		ownerAgentID = "system"
	}
	if strings.TrimSpace(content) == "" {
		content = "{}"
	}
	if metadata == nil {
		metadata = map[string]any{}
	}

	if _, err := r.mem.ClearTempMemory(ctx, "default", channelKey); err != nil {
		return "", fmt.Errorf("clear existing checkpoint: %w", err)
	}
	id, err := r.mem.PutTempMemory(ctx, "default", channelKey, ownerAgentID, content, metadata, 0)
	if err != nil {
		return "", fmt.Errorf("store checkpoint: %w", err)
	}
	return id, nil
}

package swarm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/mycelis/core/internal/memory"
	"github.com/mycelis/core/pkg/protocol"
)

// signalCheckpointChannelPrefix is shared with the temp read rule, which
// treats system-class rows under it as org-wide bus state (MEM-LANES-2).
const signalCheckpointChannelPrefix = memory.SignalCheckpointChannelPrefix

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

// errSharedChannelNoUser refuses a model-tool write with no user to a shared
// channel (memory.SharedTempChannel; MEM-LANES-2).
var errSharedChannelNoUser = errors.New("it is a shared channel, and without a signed-in user only Core's bus writer may write it")

// upsertSignalCheckpoint is publish_signal's checkpoint, a model-tool write:
// it belongs to the turn's user (MEM-LANES), and with no user it may not land
// on a shared channel, where other users or no-user prompts would read it.
func (r *InternalToolRegistry) upsertSignalCheckpoint(ctx context.Context, channelKey, ownerAgentID, content string, metadata map[string]any) (string, error) {
	owner, err := recallAccessFromContext(ctx).ownerUserID()
	if err != nil {
		return "", fmt.Errorf("store checkpoint: %w", err)
	}
	if owner == "" && memory.SharedTempChannel(channelKey) {
		return "", errSharedChannelNoUser
	}
	return r.storeSignalCheckpoint(ctx, owner, channelKey, ownerAgentID, content, metadata)
}

// storeSignalCheckpoint replaces owner's latest checkpoint on channelKey: a
// user's replaces only their own, a no-user (system) one replaces owner-less
// rows. Only upsertSignalCheckpoint and Core's bus writer
// (publishToolBusSignal) call it.
func (r *InternalToolRegistry) storeSignalCheckpoint(ctx context.Context, owner, channelKey, ownerAgentID, content string, metadata map[string]any) (string, error) {
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
	if _, err := r.mem.ClearTempMemory(ctx, "default", channelKey, memory.GovernedReader{UserID: owner}); err != nil {
		return "", fmt.Errorf("clear existing checkpoint: %w", err)
	}
	id, err := r.mem.PutTempMemory(ctx, "default", channelKey, ownerAgentID, content, metadata, 0, owner)
	if err != nil {
		return "", fmt.Errorf("store checkpoint: %w", err)
	}
	return id, nil
}

// publishToolBusSignal is Core's bus writer: it publishes an MCP tool's
// status or result on the team signal subject and keeps the latest one as a
// checkpoint. The checkpoint belongs to the turn's user when recall has one
// (MEM-LANES-2: the tool's arguments and result are that user's turn
// content); with no user it is system bus state, readable org-wide. A turn
// whose user could not be verified keeps no checkpoint.
func (a *Agent) publishToolBusSignal(recall RecallAccess, payloadKind protocol.SignalPayloadKind, sourceKind protocol.SignalSourceKind, payload map[string]any) {
	if a.nc == nil || strings.TrimSpace(a.TeamID) == "" {
		return
	}
	subject := ""
	switch payloadKind {
	case protocol.PayloadKindStatus:
		subject = fmt.Sprintf(protocol.TopicTeamSignalStatus, a.TeamID)
	case protocol.PayloadKindResult:
		subject = fmt.Sprintf(protocol.TopicTeamSignalResult, a.TeamID)
	default:
		return
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Agent [%s] signal payload marshal failed: %v", a.Manifest.ID, err)
		return
	}
	sourceChannel := fmt.Sprintf(protocol.TopicTeamInternalTrigger, a.TeamID)
	wrapped, err := protocol.WrapSignalPayloadWithMeta(sourceKind, sourceChannel, payloadKind, a.runID, a.TeamID, a.Manifest.ID, raw)
	if err != nil {
		log.Printf("Agent [%s] signal envelope wrap failed: %v", a.Manifest.ID, err)
		return
	}
	if err := a.nc.Publish(subject, wrapped); err != nil {
		log.Printf("Agent [%s] publish signal failed on [%s]: %v", a.Manifest.ID, subject, err)
		return
	}
	if a.internalTools == nil {
		return
	}
	channelKey := resolveSignalCheckpointChannelKey(subject, nil)
	owner, err := recall.ownerUserID()
	if err != nil {
		log.Printf("Agent [%s] checkpoint skipped on [%s]: %v", a.Manifest.ID, channelKey, err)
		return
	}
	metadata := map[string]any{"subject": subject, "source_kind": string(sourceKind), "payload_kind": string(payloadKind), "team_id": a.TeamID, "agent_id": a.Manifest.ID}
	if strings.TrimSpace(a.runID) != "" {
		metadata["run_id"] = strings.TrimSpace(a.runID)
	}
	if _, err := a.internalTools.storeSignalCheckpoint(a.ctx, owner, channelKey, a.Manifest.ID, string(wrapped), metadata); err != nil {
		log.Printf("Agent [%s] checkpoint update failed on [%s]: %v", a.Manifest.ID, channelKey, err)
	}
}

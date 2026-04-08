package swarm

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
)

func (r *InternalToolRegistry) handlePublishSignal(ctx context.Context, args map[string]any) (string, error) {
	subject, _ := args["subject"].(string)
	message, _ := args["message"].(string)
	filePath, _ := args["file_path"].(string)

	privacyMode, _ := args["privacy_mode"].(string)
	privacyMode = strings.ToLower(strings.TrimSpace(privacyMode))
	privateReference := parseOptionalBool(args, "private", false) ||
		parseOptionalBool(args, "private_reference", false) ||
		privacyMode == "reference"
	if !privateReference {
		privacyMode = "full"
	} else {
		privacyMode = "reference"
	}

	subject = strings.TrimSpace(subject)
	message = strings.TrimSpace(message)
	if subject == "" {
		return "", fmt.Errorf("publish_signal requires 'subject'")
	}
	if message == "" && !privateReference {
		return "", fmt.Errorf("publish_signal requires 'message' unless privacy_mode=reference")
	}
	if message == "" && strings.TrimSpace(filePath) == "" {
		return "", fmt.Errorf("publish_signal requires non-empty 'message' or 'file_path'")
	}

	if r.nc == nil {
		return "", fmt.Errorf("NATS not available — cannot publish signal")
	}

	fileRef, err := buildWorkspaceFileReference(filePath)
	if err != nil {
		return "", fmt.Errorf("publish_signal invalid file_path: %w", err)
	}

	channelKey := resolveSignalCheckpointChannelKey(subject, args)
	if channelKey == "" {
		channelKey = resolveSignalCheckpointChannelKey(subject, nil)
	}
	if message == "" && fileRef != nil {
		message = fmt.Sprintf("Private file reference prepared for %s", fileRef["path"])
	}

	payload := []byte(message)
	checkpointContent := message
	checkpointMetadata := map[string]any{
		"subject":      subject,
		"channel_key":  channelKey,
		"privacy_mode": privacyMode,
	}
	if fileRef != nil {
		checkpointMetadata["file_ref"] = fileRef
	}

	if inv, ok := ToolInvocationContextFromContext(ctx); ok {
		if strings.TrimSpace(inv.RunID) != "" {
			checkpointMetadata["run_id"] = strings.TrimSpace(inv.RunID)
		}
		if strings.TrimSpace(inv.TeamID) != "" {
			checkpointMetadata["team_id"] = strings.TrimSpace(inv.TeamID)
		}
		if strings.TrimSpace(inv.AgentID) != "" {
			checkpointMetadata["agent_id"] = strings.TrimSpace(inv.AgentID)
		}
	}
	if inferredTeamID := inferTeamIDFromSubject(subject); inferredTeamID != "" {
		if _, ok := checkpointMetadata["team_id"]; !ok {
			checkpointMetadata["team_id"] = inferredTeamID
		}
	}

	if privateReference {
		checkpointBody := map[string]any{
			"subject":       subject,
			"message":       message,
			"channel_key":   channelKey,
			"privacy_mode":  "reference",
			"published_at":  time.Now().UTC().Format(time.RFC3339),
			"source_intent": "private_channel_reference",
		}
		if fileRef != nil {
			checkpointBody["file_ref"] = fileRef
		}
		checkpointRaw, marshalErr := json.Marshal(checkpointBody)
		if marshalErr == nil {
			checkpointContent = string(checkpointRaw)
		}

		publicPayload := map[string]any{
			"mode":        "private_reference",
			"channel_key": channelKey,
			"summary":     truncateLog(message, 220),
			"reference": map[string]any{
				"kind":        "temp_memory_channel",
				"channel_key": channelKey,
				"read_tool":   "read_signals",
				"read_hint":   "set latest_only=true to recover the latest private payload for relaunch",
			},
		}
		if fileRef != nil {
			publicPayload["file_ref"] = fileRef
		}
		publicRaw, marshalErr := json.Marshal(publicPayload)
		if marshalErr != nil {
			return "", fmt.Errorf("marshal private reference payload: %w", marshalErr)
		}
		payload = publicRaw
	}

	if payloadKind, ok := inferPayloadKindFromSubject(subject); ok {
		teamID := inferTeamIDFromSubject(subject)
		wrapped, wrapErr := r.wrapGovernedSignalPayload(
			ctx,
			"internal_tool.publish_signal",
			teamID,
			payloadKind,
			payload,
		)
		if wrapErr != nil {
			return "", fmt.Errorf("failed to wrap signal payload: %w", wrapErr)
		}
		payload = wrapped
	}

	if err := r.nc.Publish(subject, payload); err != nil {
		return "", fmt.Errorf("failed to publish to %s: %w", subject, err)
	}

	ownerAgentID := "system"
	if inv, ok := ToolInvocationContextFromContext(ctx); ok && strings.TrimSpace(inv.AgentID) != "" {
		ownerAgentID = strings.TrimSpace(inv.AgentID)
	}
	checkpointMetadata["published_bytes"] = len(payload)
	checkpointID, checkpointErr := r.upsertSignalCheckpoint(ctx, channelKey, ownerAgentID, checkpointContent, checkpointMetadata)
	if checkpointErr != nil {
		log.Printf("publish_signal checkpoint update failed on [%s]: %v", channelKey, checkpointErr)
	}

	if checkpointID != "" {
		return fmt.Sprintf("Signal published to %s (%d bytes). Latest channel checkpoint: %s (%s).", subject, len(payload), checkpointID, channelKey), nil
	}
	return fmt.Sprintf("Signal published to %s (%d bytes).", subject, len(payload)), nil
}

func (r *InternalToolRegistry) handleReadSignals(ctx context.Context, args map[string]any) (string, error) {
	subject, _ := args["subject"].(string)
	if strings.TrimSpace(subject) == "" {
		if v, _ := args["topic_pattern"].(string); strings.TrimSpace(v) != "" {
			subject = v
		} else if v, _ := args["topic"].(string); strings.TrimSpace(v) != "" {
			subject = v
		} else if v, _ := args["channel"].(string); strings.TrimSpace(v) != "" {
			subject = v
		} else {
			for _, raw := range args {
				if s, ok := raw.(string); ok && strings.HasPrefix(strings.TrimSpace(s), "swarm.") {
					subject = strings.TrimSpace(s)
					break
				}
			}
		}
	}

	latestOnly := parseOptionalBool(args, "latest_only", false) || parseOptionalBool(args, "relaunch", false)
	channelKey := resolveSignalCheckpointChannelKey(subject, args)
	if latestOnly {
		if channelKey == "" {
			return "", fmt.Errorf("read_signals with latest_only requires 'channel_key' or 'subject'")
		}
		if r.mem == nil {
			return "", fmt.Errorf("memory service offline — latest channel checkpoints unavailable")
		}
		entries, err := r.mem.GetTempMemory(ctx, "default", channelKey, 1)
		if err != nil {
			return "", fmt.Errorf("read latest channel checkpoint: %w", err)
		}

		result := map[string]any{
			"mode":        "latest_checkpoint",
			"subject":     strings.TrimSpace(subject),
			"channel_key": channelKey,
		}
		if len(entries) == 0 {
			result["latest"] = nil
			data, _ := json.Marshal(result)
			return string(data), nil
		}

		entry := entries[0]
		content := any(entry.Content)
		if trimmed := strings.TrimSpace(entry.Content); trimmed != "" && json.Valid([]byte(trimmed)) {
			var decoded any
			if err := json.Unmarshal([]byte(trimmed), &decoded); err == nil {
				content = decoded
			}
		}
		result["latest"] = map[string]any{
			"id":             entry.ID,
			"owner_agent_id": entry.OwnerAgentID,
			"content":        content,
			"metadata":       entry.Metadata,
			"updated_at":     entry.UpdatedAt.UTC().Format(time.RFC3339),
		}
		data, _ := json.Marshal(result)
		return string(data), nil
	}

	if subject == "" {
		return "", fmt.Errorf("read_signals requires 'subject'")
	}

	if r.nc == nil {
		return "", fmt.Errorf("NATS not available — cannot read signals")
	}

	durationMs := 3000
	if d, ok := args["duration_ms"].(float64); ok && d > 0 {
		durationMs = int(d)
	}
	if durationMs > 10000 {
		durationMs = 10000
	}

	maxMsgs := 20
	if m, ok := args["max_msgs"].(float64); ok && m > 0 {
		maxMsgs = int(m)
	}

	type signalMsg struct {
		Subject string `json:"subject"`
		Data    string `json:"data"`
	}

	var collected []signalMsg
	sub, err := r.nc.Subscribe(subject, func(msg *nats.Msg) {
		if len(collected) < maxMsgs {
			collected = append(collected, signalMsg{
				Subject: msg.Subject,
				Data:    string(msg.Data),
			})
		}
	})
	if err != nil {
		return "", fmt.Errorf("failed to subscribe to %s: %w", subject, err)
	}
	defer sub.Unsubscribe()

	select {
	case <-time.After(time.Duration(durationMs) * time.Millisecond):
	case <-ctx.Done():
	}

	if collected == nil {
		collected = []signalMsg{}
	}

	result := map[string]any{
		"subject":   subject,
		"duration":  fmt.Sprintf("%dms", durationMs),
		"collected": len(collected),
		"messages":  collected,
	}
	if channelKey != "" {
		result["channel_key"] = channelKey
	}

	data, _ := json.Marshal(result)
	return string(data), nil
}

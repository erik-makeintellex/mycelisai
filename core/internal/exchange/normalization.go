package exchange

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ArtifactNormalizationInput struct {
	ArtifactID    uuid.UUID
	ArtifactType  string
	Title         string
	AgentID       string
	ChannelName   string
	TargetRole    string
	Status        string
	Confidence    *float64
	Tags          []string
	ContinuityKey string
	ThreadID      *uuid.UUID
}

type MCPNormalizationInput struct {
	ServerName    string
	ToolName      string
	Summary       string
	TargetRole    string
	Status        string
	Result        map[string]any
	ContinuityKey string
	ThreadID      *uuid.UUID
}

func (s *Service) PublishArtifact(ctx context.Context, input ArtifactNormalizationInput) (*ExchangeItem, error) {
	channelName := strings.TrimSpace(input.ChannelName)
	if channelName == "" {
		switch strings.ToLower(input.ArtifactType) {
		case "image", "audio":
			channelName = "media.image.output"
		case "file":
			channelName = "api.data.output"
		default:
			channelName = "organization.planning.work"
		}
	}
	schemaID := "FileResult"
	if strings.EqualFold(input.ArtifactType, "image") || strings.EqualFold(input.ArtifactType, "audio") {
		schemaID = "MediaResult"
	} else if strings.EqualFold(input.ArtifactType, "document") || strings.EqualFold(input.ArtifactType, "code") || strings.EqualFold(input.ArtifactType, "chart") || strings.EqualFold(input.ArtifactType, "data") {
		schemaID = "TextResult"
	}
	summary := fmt.Sprintf("%s produced artifact %q.", input.AgentID, input.Title)
	payload := map[string]any{
		"summary":      summary,
		"status":       defaultStatus(input.Status),
		"artifact_uri": input.ArtifactID.String(),
		"source_role":  input.AgentID,
		"target_role":  defaultTarget(input.TargetRole),
		"created_at":   time.Now().UTC().Format(time.RFC3339),
		"tags":         toAnySlice(input.Tags),
	}
	if input.ContinuityKey != "" {
		payload["continuity_key"] = input.ContinuityKey
	}
	if input.Confidence != nil {
		payload["confidence"] = *input.Confidence
	}
	return s.Publish(ctx, PublishInput{
		ChannelName: channelName,
		SchemaID:    schemaID,
		Payload:     payload,
		CreatedBy:   input.AgentID,
		AddressedTo: defaultTarget(input.TargetRole),
		ThreadID:    input.ThreadID,
		Visibility:  "advanced",
		Summary:     summary,
	})
}

func (s *Service) PublishMCPResult(ctx context.Context, input MCPNormalizationInput) (*ExchangeItem, error) {
	channelName, schemaID, tags := classifyMCPOutput(input.ServerName, input.ToolName)
	summary := strings.TrimSpace(input.Summary)
	if summary == "" {
		summary = fmt.Sprintf("%s returned output via %s.", input.ToolName, input.ServerName)
	}
	payload := map[string]any{
		"summary":     summary,
		"status":      defaultStatus(input.Status),
		"source_role": "mcp:" + strings.TrimSpace(input.ServerName),
		"target_role": defaultTarget(input.TargetRole),
		"created_at":  time.Now().UTC().Format(time.RFC3339),
		"updated_at":  time.Now().UTC().Format(time.RFC3339),
		"tags":        toAnySlice(tags),
		"tool_result": input.Result,
	}
	if input.ContinuityKey != "" {
		payload["continuity_key"] = input.ContinuityKey
	}
	return s.Publish(ctx, PublishInput{
		ChannelName: channelName,
		SchemaID:    schemaID,
		Payload:     payload,
		CreatedBy:   "mcp:" + strings.TrimSpace(input.ServerName),
		AddressedTo: defaultTarget(input.TargetRole),
		ThreadID:    input.ThreadID,
		Visibility:  "advanced",
		Summary:     summary,
	})
}

func classifyMCPOutput(serverName, toolName string) (string, string, []string) {
	label := strings.ToLower(strings.TrimSpace(serverName + " " + toolName))
	switch {
	case strings.Contains(label, "image"), strings.Contains(label, "vision"), strings.Contains(label, "media"):
		return "media.image.output", "MediaResult", []string{"mcp", "media", strings.TrimSpace(toolName)}
	case strings.Contains(label, "browser"), strings.Contains(label, "fetch"), strings.Contains(label, "search"), strings.Contains(label, "research"):
		return "browser.research.results", "ToolResult", []string{"mcp", "research", strings.TrimSpace(toolName)}
	default:
		return "api.data.output", "ToolResult", []string{"mcp", "data", strings.TrimSpace(toolName)}
	}
}

func defaultStatus(status string) string {
	if strings.TrimSpace(status) == "" {
		return "completed"
	}
	return status
}

func defaultTarget(target string) string {
	if strings.TrimSpace(target) == "" {
		return "soma"
	}
	return target
}

func toAnySlice(in []string) []any {
	out := make([]any, 0, len(in))
	for _, item := range in {
		if strings.TrimSpace(item) != "" {
			out = append(out, item)
		}
	}
	return out
}

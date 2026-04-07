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
	ServerID      string
	ServerName    string
	ToolName      string
	Summary       string
	ResultPreview string
	TargetRole    string
	Status        string
	Result        map[string]any
	SourceTeam    string
	AgentID       string
	RunID         string
	ContinuityKey string
	ThreadID      *uuid.UUID
}

func (s *Service) PublishArtifact(ctx context.Context, input ArtifactNormalizationInput) (*ExchangeItem, error) {
	if ActorFromContext(ctx).Role == "" {
		ctx = WithActor(ctx, Actor{Role: defaultActorRoleForAgent(input.AgentID)})
	}
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
		ChannelName:      channelName,
		SchemaID:         schemaID,
		Payload:          payload,
		CreatedBy:        input.AgentID,
		AddressedTo:      defaultTarget(input.TargetRole),
		ThreadID:         input.ThreadID,
		Visibility:       "advanced",
		SensitivityClass: "team_scoped",
		SourceRole:       defaultActorRoleForAgent(input.AgentID),
		TargetRole:       defaultTarget(input.TargetRole),
		CapabilityID: map[string]string{
			"MediaResult": "media_generation",
			"FileResult":  "file_output",
			"TextResult":  "file_output",
		}[schemaID],
		TrustClass: map[string]string{
			"MediaResult": "bounded_external",
			"FileResult":  "trusted_internal",
			"TextResult":  "trusted_internal",
		}[schemaID],
		ReviewRequired: schemaID != "TextResult",
		Summary:        summary,
	})
}

func (s *Service) PublishMCPResult(ctx context.Context, input MCPNormalizationInput) (*ExchangeItem, error) {
	if ActorFromContext(ctx).Role == "" {
		ctx = WithActor(ctx, Actor{Role: "mcp"})
	}
	serverName := strings.TrimSpace(input.ServerName)
	serverID := strings.TrimSpace(input.ServerID)
	if serverName == "" {
		serverName = serverID
	}
	if serverName == "" {
		serverName = "mcp"
	}
	channelName, schemaID, tags := classifyMCPOutput(serverName, input.ToolName)
	summary := strings.TrimSpace(input.Summary)
	if summary == "" {
		summary = fmt.Sprintf("%s returned output via %s.", input.ToolName, serverName)
	}
	state := defaultStatus(input.Status)
	preview := strings.TrimSpace(input.ResultPreview)
	if preview == "" {
		preview = summary
	}
	payload := map[string]any{
		"summary":        summary,
		"status":         state,
		"state":          state,
		"source_role":    "mcp:" + serverName,
		"target_role":    defaultTarget(input.TargetRole),
		"created_at":     time.Now().UTC().Format(time.RFC3339),
		"updated_at":     time.Now().UTC().Format(time.RFC3339),
		"tags":           toAnySlice(tags),
		"tool_result":    input.Result,
		"tool_name":      strings.TrimSpace(input.ToolName),
		"server_name":    serverName,
		"result_preview": preview,
	}
	if serverID != "" {
		payload["server_id"] = serverID
	}
	if sourceTeam := strings.TrimSpace(input.SourceTeam); sourceTeam != "" {
		payload["source_team"] = sourceTeam
	}
	if agentID := strings.TrimSpace(input.AgentID); agentID != "" {
		payload["agent_id"] = agentID
	}
	if runID := strings.TrimSpace(input.RunID); runID != "" {
		payload["run_id"] = runID
	}
	if input.ContinuityKey != "" {
		payload["continuity_key"] = input.ContinuityKey
	}
	return s.Publish(ctx, PublishInput{
		ChannelName:      channelName,
		SchemaID:         schemaID,
		Payload:          payload,
		CreatedBy:        "mcp:" + serverName,
		AddressedTo:      defaultTarget(input.TargetRole),
		ThreadID:         input.ThreadID,
		Visibility:       "advanced",
		SensitivityClass: "team_scoped",
		SourceRole:       "mcp",
		SourceTeam:       strings.TrimSpace(input.SourceTeam),
		TargetRole:       defaultTarget(input.TargetRole),
		CapabilityID:     capabilityForToolChannel(channelName),
		TrustClass:       "bounded_external",
		ReviewRequired:   true,
		Metadata: map[string]any{
			"source_kind": "mcp",
			"mcp": map[string]any{
				"server_id":      serverID,
				"server_name":    serverName,
				"tool_name":      strings.TrimSpace(input.ToolName),
				"state":          state,
				"agent_id":       strings.TrimSpace(input.AgentID),
				"source_team":    strings.TrimSpace(input.SourceTeam),
				"run_id":         strings.TrimSpace(input.RunID),
				"continuity_key": strings.TrimSpace(input.ContinuityKey),
			},
		},
		Summary: summary,
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

func capabilityForToolChannel(channelName string) string {
	switch channelName {
	case "browser.research.results":
		return "browser_research"
	case "media.image.output":
		return "media_generation"
	default:
		return "api_data_access"
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

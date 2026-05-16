package server

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/mcp"
	"github.com/mycelis/core/internal/swarm"
	"github.com/mycelis/core/pkg/protocol"
)

type plannedToolOutputEnvelope struct {
	Message   string                     `json:"message"`
	Artifact  *protocol.ChatArtifactRef  `json:"artifact,omitempty"`
	Artifacts []protocol.ChatArtifactRef `json:"artifacts,omitempty"`
}

func extractPlannedToolOutputArtifacts(toolResult string) (string, []protocol.ChatArtifactRef, bool) {
	var toolOutput plannedToolOutputEnvelope
	if err := json.Unmarshal([]byte(toolResult), &toolOutput); err != nil {
		return toolResult, nil, false
	}
	artifacts := make([]protocol.ChatArtifactRef, 0, len(toolOutput.Artifacts)+1)
	if toolOutput.Artifact != nil {
		artifacts = append(artifacts, *toolOutput.Artifact)
	}
	if len(toolOutput.Artifacts) > 0 {
		artifacts = append(artifacts, toolOutput.Artifacts...)
	}
	if len(artifacts) == 0 {
		return toolResult, nil, false
	}
	return toolOutput.Message, artifacts, true
}

func (s *AdminServer) plannedMCPToolExecutor() swarm.MCPToolExecutor {
	if s == nil {
		return nil
	}
	if s.MCPToolExecutor != nil {
		return s.MCPToolExecutor
	}
	if s.MCP != nil && s.MCPPool != nil {
		return mcp.NewToolExecutorAdapter(s.MCP, s.MCPPool)
	}
	return nil
}

func (s *AdminServer) resolveApprovedToolCall(ctx context.Context, executor *swarm.CompositeToolExecutor, mcpExec swarm.MCPToolExecutor, planned protocol.PlannedToolCall) (uuid.UUID, string, error) {
	toolName := strings.TrimSpace(planned.Name)
	toolRef := strings.TrimSpace(planned.ToolRef)
	if ref := mcp.ParseToolRef(toolRef); ref != nil {
		if ref.ToolName == "" || ref.ToolName == "*" {
			return uuid.Nil, "", fmt.Errorf("approved MCP tool_ref %q must name a concrete tool", toolRef)
		}
		if mcpExec == nil {
			return uuid.Nil, "", fmt.Errorf("MCP tool executor not available for approved tool_ref %q", toolRef)
		}
		if s != nil && s.MCP != nil && ref.ServerName != "" {
			srv, err := s.MCP.FindServerByName(ctx, ref.ServerName)
			if err != nil {
				return uuid.Nil, "", err
			}
			if srv != nil {
				return srv.ID, ref.ToolName, nil
			}
		}
		serverID, resolvedToolName, err := mcpExec.FindToolByName(ctx, ref.ToolName)
		if err != nil {
			return uuid.Nil, "", err
		}
		return serverID, resolvedToolName, nil
	}
	return executor.FindToolByName(ctx, toolName)
}

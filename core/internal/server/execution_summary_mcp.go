package server

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/mycelis/core/pkg/protocol"
)

func buildMCPToolCallExecutionSummary(serverID, serverName, toolName, summary, exchangeItemID, runID string, result any) *protocol.ExecutionSummary {
	trimmedTool := strings.TrimSpace(toolName)
	trimmedServer := firstNonEmptyString(strings.TrimSpace(serverName), strings.TrimSpace(serverID), "mcp")
	summaryText := firstNonEmptyString(strings.TrimSpace(summary), extractMCPResultSummary(result), fmt.Sprintf("%s returned output.", firstNonEmptyString(trimmedTool, "MCP tool")))
	ref := outputRefFromMCPToolCall(trimmedServer, trimmedTool, summaryText, exchangeItemID, runID, result)
	output := protocol.ExecutionOutput{
		ID:          ref.ID,
		ArtifactID:  ref.ArtifactID,
		Kind:        "mcp_tool_result",
		Title:       ref.Label,
		Summary:     summaryText,
		StorageRef:  ref.StorageRef,
		Href:        ref.OpenURL,
		OpenURL:     ref.OpenURL,
		RevealURL:   ref.RevealURL,
		SourceRunID: ref.SourceRunID,
		Role:        ref.Role,
		Status:      ref.Status,
		ContentType: ref.ContentType,
		Bytes:       ref.Bytes,
		SHA256:      ref.SHA256,
		Retained:    strings.TrimSpace(exchangeItemID) != "",
		Proof:       ref.Proof,
	}
	return &protocol.ExecutionSummary{
		Title:      firstNonEmptyString(trimmedTool, "MCP tool call"),
		Summary:    summaryText,
		Status:     ref.Status,
		RunID:      runID,
		ProofID:    ref.ProofArtifactID,
		OutputRefs: []protocol.OutputRef{ref},
		Outputs:    []protocol.ExecutionOutput{output},
	}
}

func outputRefFromMCPToolCall(serverName, toolName, summary, exchangeItemID, runID string, result any) protocol.OutputRef {
	rawBytes, _ := json.Marshal(result)
	checksum := ""
	if len(rawBytes) > 0 {
		sum := sha256.Sum256(rawBytes)
		checksum = fmt.Sprintf("%x", sum)
	}
	refID := firstNonEmptyString(strings.TrimSpace(exchangeItemID), fmt.Sprintf("mcp:%s:%s", serverName, firstNonEmptyString(toolName, "tool")))
	storageRef := mcpStorageRef(serverName, toolName, exchangeItemID, "")
	openURL := mcpOpenURL(exchangeItemID, serverName, toolName)
	status := "recorded"
	readback := proofStatusVerified
	hint := ""
	if strings.TrimSpace(exchangeItemID) == "" {
		status = "transient"
		readback = proofStatusDegraded
		hint = "Exchange item was not persisted; MCP result is available inline only."
	}
	proof := &protocol.OutputProofEnvelope{
		ProofID:           firstNonEmptyString(exchangeItemID, refID),
		OutputRefID:       refID,
		ArtifactID:        strings.TrimSpace(exchangeItemID),
		StorageRef:        storageRef,
		SourceRunID:       strings.TrimSpace(runID),
		ExecutionStatus:   status,
		Checksum:          checksum,
		ChecksumAlgorithm: "sha256",
		Bytes:             len(rawBytes),
		ContentType:       "application/json",
		CreatedAt:         time.Now().UTC().Format(time.RFC3339),
		ReadbackStatus:    readback,
		RecoveryHint:      hint,
	}
	return protocol.NormalizeOutputRef(protocol.OutputRef{
		ID:              refID,
		ArtifactID:      strings.TrimSpace(exchangeItemID),
		Kind:            mcpOutputKindForTool(toolName, ""),
		Label:           firstNonEmptyString(toolName, summary, "MCP tool result"),
		StorageRef:      storageRef,
		OpenURL:         openURL,
		RevealURL:       openURL,
		SourceRunID:     strings.TrimSpace(runID),
		ProofArtifactID: firstNonEmptyString(exchangeItemID, refID),
		Role:            "primary",
		Status:          status,
		ContentType:     "application/json",
		Bytes:           len(rawBytes),
		SHA256:          checksum,
		Proof:           proof,
		ProofEnvelope:   proof,
	})
}

func outputRefFromMCPActivityEntry(entry mcpActivityEntry, role string) protocol.OutputRef {
	refID := firstNonEmptyString(strings.TrimSpace(entry.ID), fmt.Sprintf("mcp:%s:%s", entry.ServerName, entry.ToolName))
	storageRef := mcpStorageRef(entry.ServerName, entry.ToolName, entry.ID, entry.ChannelName)
	openURL := mcpOpenURL(entry.ID, entry.ServerName, entry.ToolName)
	status := firstNonEmptyString(strings.TrimSpace(entry.State), "completed")
	body := firstNonEmptyString(entry.Message, entry.Summary)
	sum := sha256.Sum256([]byte(body))
	checksum := fmt.Sprintf("%x", sum)
	proof := &protocol.OutputProofEnvelope{
		ProofID:           refID,
		OutputRefID:       refID,
		ArtifactID:        strings.TrimSpace(entry.ID),
		StorageRef:        storageRef,
		SourceRunID:       strings.TrimSpace(entry.RunID),
		ExecutionStatus:   status,
		Checksum:          checksum,
		ChecksumAlgorithm: "sha256",
		Bytes:             len(body),
		ContentType:       "application/json",
		CreatedAt:         firstNonEmptyString(entry.Timestamp, time.Now().UTC().Format(time.RFC3339)),
		ReadbackStatus:    proofStatusVerified,
	}
	return protocol.NormalizeOutputRef(protocol.OutputRef{
		ID:              refID,
		ArtifactID:      strings.TrimSpace(entry.ID),
		Kind:            mcpOutputKindForTool(entry.ToolName, entry.ChannelName),
		Label:           firstNonEmptyString(entry.Summary, entry.ToolName, "MCP activity result"),
		StorageRef:      storageRef,
		OpenURL:         openURL,
		RevealURL:       openURL,
		SourceRunID:     strings.TrimSpace(entry.RunID),
		ProofArtifactID: refID,
		PackageID:       strings.TrimSpace(entry.TeamID),
		Role:            firstNonEmptyString(role, "primary"),
		Status:          status,
		ContentType:     "application/json",
		Bytes:           len(body),
		SHA256:          checksum,
		Proof:           proof,
		ProofEnvelope:   proof,
	})
}

func executionSummaryFromMCPActivityEntries(entries []mcpActivityEntry) *protocol.ExecutionSummary {
	if len(entries) == 0 {
		return nil
	}
	refs := make([]protocol.OutputRef, 0, len(entries))
	outputs := make([]protocol.ExecutionOutput, 0, len(entries))
	runID := ""
	for index, entry := range entries {
		role := "primary"
		if index > 0 {
			role = "supporting"
		}
		ref := outputRefFromMCPActivityEntry(entry, role)
		if runID == "" && ref.SourceRunID != "" {
			runID = ref.SourceRunID
		}
		refs = append(refs, ref)
		outputs = append(outputs, protocol.ExecutionOutput{
			ID:          ref.ID,
			ArtifactID:  ref.ArtifactID,
			Kind:        ref.Kind,
			Title:       ref.Label,
			Summary:     firstNonEmptyString(entry.Message, entry.Summary),
			StorageRef:  ref.StorageRef,
			Href:        ref.OpenURL,
			OpenURL:     ref.OpenURL,
			RevealURL:   ref.RevealURL,
			SourceRunID: ref.SourceRunID,
			Role:        ref.Role,
			Status:      ref.Status,
			ContentType: ref.ContentType,
			Bytes:       ref.Bytes,
			SHA256:      ref.SHA256,
			Retained:    true,
			Proof:       ref.Proof,
		})
	}
	primary := refs[0]
	return &protocol.ExecutionSummary{
		Title:      firstNonEmptyString(entries[0].ToolName, "MCP activity"),
		Summary:    firstNonEmptyString(entries[0].Summary, entries[0].Message, "MCP tool activity recorded."),
		Status:     primary.Status,
		RunID:      runID,
		ProofID:    primary.ProofArtifactID,
		OutputRefs: refs,
		Outputs:    outputs,
	}
}

func mcpStorageRef(serverName, toolName, itemID, channel string) string {
	server := firstNonEmptyString(strings.TrimSpace(serverName), "mcp")
	tool := firstNonEmptyString(strings.TrimSpace(toolName), "tool")
	if trimmedID := strings.TrimSpace(itemID); trimmedID != "" {
		if trimmedChannel := strings.TrimSpace(channel); trimmedChannel != "" {
			return fmt.Sprintf("mcp://%s/%s/%s/%s", server, trimmedChannel, tool, trimmedID)
		}
		return fmt.Sprintf("mcp://%s/%s/%s", server, tool, trimmedID)
	}
	return fmt.Sprintf("mcp://%s/%s", server, tool)
}

func mcpOpenURL(itemID, serverName, toolName string) string {
	if trimmedID := strings.TrimSpace(itemID); trimmedID != "" {
		return "/resources?tab=tools&item=" + url.QueryEscape(trimmedID)
	}
	query := strings.TrimSpace(firstNonEmptyString(toolName, serverName))
	if query == "" {
		return "/resources?tab=tools"
	}
	return "/resources?tab=tools&tool=" + url.QueryEscape(query)
}

func mcpOutputKindForTool(toolName, channelName string) string {
	lowerChannel := strings.ToLower(strings.TrimSpace(channelName))
	lowerTool := strings.ToLower(strings.TrimSpace(toolName))
	switch {
	case strings.Contains(lowerChannel, "image"), strings.Contains(lowerTool, "image"):
		return "image"
	case strings.Contains(lowerChannel, "research"), strings.Contains(lowerTool, "read"), strings.Contains(lowerTool, "doc"):
		return "document"
	default:
		return "mcp_result"
	}
}

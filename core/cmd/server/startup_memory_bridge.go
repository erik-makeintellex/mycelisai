package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mycelis/core/internal/memory"
	pb "github.com/mycelis/core/pkg/pb/swarm"
	"github.com/mycelis/core/pkg/protocol"
	"google.golang.org/protobuf/proto"
)

func buildLegacyMemoryLogEntry(subject string, envelope *pb.MsgEnvelope) *memory.LogEntry {
	if envelope == nil {
		return nil
	}

	ts := time.Now().UTC()
	if envelope.Timestamp != nil {
		ts = envelope.Timestamp.AsTime()
	}

	msgBody := ""
	intent := "event"
	level := "INFO"
	payloadKind := protocol.PayloadKindStatus

	switch p := envelope.Payload.(type) {
	case *pb.MsgEnvelope_Text:
		msgBody = p.Text.Content
		intent = p.Text.Intent
		if intent == "" {
			intent = "text"
		}
		payloadKind = protocol.PayloadKindStatus
	case *pb.MsgEnvelope_Event:
		msgBody = fmt.Sprintf("Event: %s", p.Event.EventType)
		intent = p.Event.EventType
		payloadKind = protocol.PayloadKindStatus
	case *pb.MsgEnvelope_ToolCall:
		msgBody = fmt.Sprintf("Tool Call: %s", p.ToolCall.ToolName)
		intent = "tool_call"
		payloadKind = protocol.PayloadKindStatus
	case *pb.MsgEnvelope_ToolResult:
		msgBody = fmt.Sprintf("Tool Result: %s", p.ToolResult.CallId)
		intent = "tool_result"
		payloadKind = protocol.PayloadKindResult
		if p.ToolResult.IsError {
			level = "ERROR"
			msgBody = fmt.Sprintf("Error: %s", p.ToolResult.ErrorMessage)
			payloadKind = protocol.PayloadKindError
		}
	}

	entry := &memory.LogEntry{
		TraceId:   envelope.TraceId,
		Timestamp: ts,
		Source:    envelope.SourceAgentId,
		Intent:    intent,
		Message:   msgBody,
		Level:     level,
		Context: protocol.OperationalLogContext{
			ReviewScope:   protocol.LogReviewScopeCentralReview,
			Service:       "core",
			Component:     "legacy-bus-bridge",
			Summary:       strings.TrimSpace(msgBody),
			WhyItMatters:  "Legacy bus output is mirrored into centralized review so Soma and operational leads can still inspect older agent/runtime paths alongside modern signal channels.",
			SourceKind:    protocol.SourceKindSystem,
			SourceChannel: subject,
			PayloadKind:   payloadKind,
			TeamID:        envelope.TeamId,
			AgentID:       envelope.SourceAgentId,
			Status:        strings.ToLower(level),
			TraceID:       envelope.TraceId,
			ReviewChannels: []string{
				subject,
			},
			Tags: []string{"legacy-envelope", intent},
		}.ToMap(),
	}
	return memory.NormalizeLogEntryForReview(entry)
}

func buildMemoryLogEntryFromMessage(subject string, data []byte) *memory.LogEntry {
	if shouldSkipMemoryLogBridge(subject) {
		return nil
	}

	var signalEnv protocol.SignalEnvelope
	if err := json.Unmarshal(data, &signalEnv); err == nil {
		if signalEnv.Meta.SourceKind != "" || signalEnv.Meta.SourceChannel != "" || signalEnv.Meta.PayloadKind != "" || signalEnv.Text != "" {
			if signalEnv.Meta.SourceChannel == "" {
				signalEnv.Meta.SourceChannel = subject
			}
			return memory.NewSignalReviewLogEntry(subject, signalEnv)
		}
	}

	if telemetryEnv, err := protocol.ValidateTelemetryMessage(data); err == nil {
		return memory.NewTelemetryReviewLogEntry(subject, telemetryEnv)
	}

	var envelope pb.MsgEnvelope
	if err := proto.Unmarshal(data, &envelope); err == nil {
		return buildLegacyMemoryLogEntry(subject, &envelope)
	}

	return nil
}

func shouldSkipMemoryLogBridge(subject string) bool {
	trimmed := strings.TrimSpace(subject)
	if trimmed == "" {
		return false
	}
	return trimmed == protocol.TopicGlobalHeartbeat || strings.Contains(strings.ToLower(trimmed), "heartbeat")
}

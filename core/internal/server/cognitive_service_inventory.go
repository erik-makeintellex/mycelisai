package server

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/mycelis/core/pkg/protocol"
)

func isServiceInventoryQuestion(text string) bool {
	lower := normalizeIntentText(text)
	if lower == "" {
		return false
	}
	if requestContainsAny(lower, []string{"internal tool names", "raw tool", "debug mcp", "mcp status", "mcp server", "tool ids", "tool names"}) {
		return false
	}
	serviceCue := requestContainsAny(lower, []string{"service", "services", "capabilities", "what can soma use", "what is available", "available to soma"})
	listCue := requestContainsAny(lower, []string{"list", "show", "what", "which", "available", "current"})
	shortServiceAsk := strings.Contains(lower, "services") && len(strings.Fields(lower)) <= 6
	return serviceCue && (listCue || shortServiceAsk)
}

func serviceStatusMap(services []ServiceStatus) map[string]ServiceStatus {
	out := make(map[string]ServiceStatus, len(services))
	for _, service := range services {
		out[strings.TrimSpace(service.Name)] = service
	}
	return out
}

func serviceOnline(status map[string]ServiceStatus, name string) bool {
	service, ok := status[name]
	return ok && service.Status == "online"
}

func (s *AdminServer) buildServiceInventoryAnswer(r *http.Request) string {
	status := serviceStatusMap(s.buildServiceStatuses(r.Context()))
	available := []string{"Soma workspace"}
	if serviceOnline(status, "cognitive") || serviceOnline(status, "ollama") {
		available = append(available, "Local AI engine")
	}
	if serviceOnline(status, "postgres") {
		available = append(available, "Files, output records, and workspace storage")
	}
	if serviceOnline(status, "nats") || serviceOnline(status, "groups_bus") {
		available = append(available, "Team coordination and background work")
	}
	if serviceOnline(status, "postgres") {
		available = append(available, "Memory and retained context")
	}
	available = append(available, "System status checks")
	available = uniqueOrderedTools(available)

	needsAttention := s.serviceInventoryAttention(r)

	lines := []string{"Here is what Soma can use right now:"}
	lines = append(lines, "", "Available")
	for _, item := range available {
		lines = append(lines, "- "+item)
	}
	if len(needsAttention) > 0 {
		lines = append(lines, "", "Needs attention")
		for _, item := range needsAttention {
			lines = append(lines, "- "+item)
		}
	}
	lines = append(lines, "", "If you want the technical inventory, ask: \"show internal tool names\" or \"debug MCP status.\"")
	return strings.Join(lines, "\n")
}

func (s *AdminServer) serviceInventoryAttention(r *http.Request) []string {
	var attention []string
	if s.MCP == nil || s.MCP.DB == nil {
		return []string{"Connected tools registry is not available in this runtime."}
	}
	servers, err := s.MCP.List(r.Context())
	if err != nil {
		return []string{"Connected tools registry could not be read."}
	}
	for _, server := range servers {
		name := strings.ToLower(strings.TrimSpace(server.Name))
		status := strings.ToLower(strings.TrimSpace(server.Status))
		if name != "fetch" || (status != "error" && status != "offline" && status != "degraded") {
			continue
		}
		attention = append(attention, "Web and URL access is not available right now. Open Settings -> Connected Tools and repair Fetch.")
		break
	}
	sort.Strings(attention)
	return attention
}

func (s *AdminServer) respondServiceInventorySummary(w http.ResponseWriter, r *http.Request) {
	auditEventID, _ := s.createAuditEvent(
		protocol.TemplateChatToAnswer, "admin",
		"Service inventory summary",
		attachActorIdentity(map[string]any{
			"actor":         "Soma",
			"user":          auditUserLabelFromRequest(r),
			"ask_class":     string(protocol.AskClassDirectAnswer),
			"action":        "answer_delivered",
			"result_status": "completed",
			"source_kind":   "system",
		}, r),
	)

	text := s.buildServiceInventoryAnswer(r)
	chatPayload := protocol.ChatResponsePayload{
		Text:     text,
		AskClass: protocol.AskClassDirectAnswer,
		Provenance: &protocol.AnswerProvenance{
			ResolvedIntent:  "service_inventory",
			PermissionCheck: "pass",
			PolicyDecision:  "allow",
			AuditEventID:    auditEventID,
		},
		ExecutionSummary: buildServiceInventoryExecutionSummary(auditEventID),
	}
	payloadBytes, _ := json.Marshal(chatPayload)
	envelope := protocol.CTSEnvelope{
		Meta:       protocol.CTSMeta{SourceNode: "admin", Timestamp: time.Now()},
		SignalType: protocol.SignalChatResponse,
		TrustScore: protocol.TrustScoreCognitive,
		Payload:    payloadBytes,
		TemplateID: protocol.TemplateChatToAnswer,
		Mode:       protocol.ModeAnswer,
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(envelope))
}

func buildServiceInventoryExecutionSummary(auditEventID string) *protocol.ExecutionSummary {
	return &protocol.ExecutionSummary{
		Intent: protocol.ExecutionIntent{
			Original: "service inventory",
			Resolved: "answer",
		},
		Understanding: protocol.ExecutionUnderstanding{
			Summary: "Reported user-facing services and capability availability without exposing raw tool IDs.",
			Assumptions: []string{
				"Service inventory chat is read-only and does not execute a run.",
				"Raw MCP/tool identifiers stay behind an explicit technical inventory request.",
			},
		},
		Execution: protocol.ExecutionState{
			Shape:   protocol.ExecutionShapeDirectSoma,
			Status:  protocol.ExecutionStatusCompleted,
			Summary: "Soma returned a deterministic service inventory answer.",
		},
		Outputs: []protocol.ExecutionOutput{{
			Kind:           "service_inventory_answer",
			Title:          "Service inventory answer",
			Summary:        "Read-only user-facing service summary.",
			Retained:       boolPtr(false),
			RetentionClass: protocol.ExecutionRetentionClassNonRetained,
		}},
		Proof: protocol.ExecutionProof{
			RunClass:     protocol.ExecutionRunClassNoRun,
			NoRunReason:  "Deterministic service inventory chat is audit-only and does not create an execution run.",
			ProofClass:   protocol.ExecutionProofClassAuditOnly,
			AuditEventID: auditEventID,
			Verified:     boolPtr(strings.TrimSpace(auditEventID) != ""),
		},
		AuditRecovery: protocol.AuditRecovery{
			ApprovalStatus: "allow",
			RecoveryState:  "audit_recorded",
			Retryable:      boolPtr(true),
		},
		NextStep: &protocol.ExecutionNextStep{
			Label:  "Ask a follow-up",
			Action: "chat",
		},
	}
}

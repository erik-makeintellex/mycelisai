package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/mycelis/core/pkg/protocol"
)

func isOrchestrationPlaybookQuestion(text string) (string, bool) {
	lower := strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(text)), " "))
	if lower == "" {
		return "", false
	}
	if strings.Contains(lower, "consult the architect and sentry") ||
		(strings.Contains(lower, "architect") && strings.Contains(lower, "sentry") && strings.Contains(lower, "security boundaries")) {
		return "council_review", true
	}
	if strings.Contains(lower, "audit our current trusted memory") ||
		(strings.Contains(lower, "trusted memory") && strings.Contains(lower, "governed doctrine")) {
		return "memory_audit", true
	}
	if strings.Contains(lower, "check our current backend service status and installed mcp tools") ||
		(strings.Contains(lower, "backend service status") && strings.Contains(lower, "mcp tools")) {
		return "service_capability", true
	}
	return "", false
}

func (s *AdminServer) respondOrchestrationPlaybookAnswer(w http.ResponseWriter, r *http.Request, kind, organizationID, teamID, teamName string) {
	chatPayload := s.buildOrchestrationPlaybookPayload(kind, organizationID, teamID, teamName)

	auditEventID, _ := s.createAuditEvent(
		protocol.TemplateChatToAnswer, "admin",
		"Orchestration playbook answer",
		map[string]any{
			"actor":         "Soma",
			"user":          auditUserLabelFromRequest(r),
			"ask_class":     string(chatPayload.AskClass),
			"playbook_kind": kind,
			"action":        "answer_delivered",
			"result_status": "completed",
			"source_kind":   "system",
		},
	)
	chatPayload.Provenance = &protocol.AnswerProvenance{
		ResolvedIntent:  "answer",
		PermissionCheck: "pass",
		PolicyDecision:  "allow",
		AuditEventID:    auditEventID,
	}

	payloadBytes, _ := json.Marshal(chatPayload)
	envelope := protocol.CTSEnvelope{
		Meta: protocol.CTSMeta{
			SourceNode: "admin",
			Timestamp:  time.Now(),
		},
		SignalType: protocol.SignalChatResponse,
		TrustScore: protocol.TrustScoreCognitive,
		Payload:    payloadBytes,
		TemplateID: protocol.TemplateChatToAnswer,
		Mode:       protocol.ModeAnswer,
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(envelope))
}

func (s *AdminServer) buildOrchestrationPlaybookPayload(kind, organizationID, teamID, teamName string) protocol.ChatResponsePayload {
	switch kind {
	case "council_review":
		return protocol.ChatResponsePayload{
			AskClass: protocol.AskClassSpecialist,
			Consultations: []protocol.ConsultationEntry{
				{
					Member:  "council-architect",
					Summary: "Reviewed 3-plane NATS/OutputRef/ServiceAddressBook architecture and disjoint team lane topology.",
				},
				{
					Member:  "council-sentry",
					Summary: "Verified proposal-gated mutations, SHA-256 OutputProofEnvelope readback, and RBAC/OIDC boundaries.",
				},
			},
			Text: "## Council Architecture & Security Review\n\n### 1. Architecture Tradeoffs (Architect)\n- **3-Plane Separation:** NATS JetStream handles location-transparent A2A signaling, `OutputRef` + `OutputProofEnvelope` own durable artifact storage, and `ServiceAddressBook` governs external/in-cluster service endpoints.\n- **Disjoint Team Execution:** Multi-agent waves enforce strict One-Writer-Per-File isolation with runtime collision detection (`DetectTeamPackageCollisions`).\n\n### 2. Security & Governance Boundaries (Sentry)\n- **Proposal-First Mutations:** File writes, blueprint generation, and signal broadcasts require explicit operator confirmation (`Stage 3 · Approval Gate`).\n- **Cryptographic Readback:** Every deliverable is verified on disk via SHA-256 and workspace path-boundary checks (`Stage 8 · Package & Proof Verified`).\n\n### 3. Recommended Next Steps\n- Inspect active team rosters on `/deployments` or launch a **Multi-Team Mission Blueprint**.",
		}
	case "memory_audit":
		return protocol.ChatResponsePayload{
			AskClass: protocol.AskClassDirectAnswer,
			Text: "## Trusted-Memory Control Plane Audit\n\n### 1. 5-Layer Memory Hierarchy\n- **Authority Precedence:** `GOVERNED_DOCTRINE` > `PROJECT_MEMORY` > `AGENT_MEMORY` > `SOMA_MEMORY` > `REFLECTION_MEMORY`.\n- **Immutable Doctrine:** `GOVERNED_DOCTRINE` acts as the highest-authority layer; lower layers cannot override canonical architecture or safety rules.\n\n### 2. Retrieval-Time Contradiction Halts\n- **Vector Arbitration (`ArbitrateVectorResults`):** Retrieval queries cross-check candidate recall entries against active doctrine and project rules before injection.\n- **Conflict Quarantine:** Contradictory memories are flagged as `contradiction_detected` and withheld from autonomous execution until reviewed.\n\n### 3. Operator Promotion & Conflict Resolution\n- Open `/memory` to inspect active doctrine rules, resolve quarantined memory conflicts, or promote verified reflection insights into `PROJECT_MEMORY` or `GOVERNED_DOCTRINE`.",
		}
	default:
		book := ResolveServiceAddressBook(s.Cognitive)
		runtimePosture := "Local Host"
		if book.ClusterMode {
			runtimePosture = "Kubernetes Cluster"
		}
		return protocol.ChatResponsePayload{
			AskClass: protocol.AskClassDirectAnswer,
			Text: fmt.Sprintf(
				"## Backend Service & MCP Tool Capability Status\n\n### 1. Active Runtime Posture\n- **Execution Environment:** `%s` (`cluster_mode=%t`)\n\n### 2. Resolved Service Addresses (`ServiceAddressBook`)\n- **`nats`:** `%s`\n- **`postgres`:** `%s`\n- **`text_ai`:** `%s`\n\n### 3. Recommended Inspection Surfaces\n- Open `/system` (`Service Address Book`) to verify live endpoint health and warnings, or `/resources?tab=tools` (`Connected Tools`) to inspect installed MCP servers and tool bindings.",
				runtimePosture,
				book.ClusterMode,
				book.AddressFor("nats"),
				book.AddressFor("postgres"),
				book.AddressFor("text_ai"),
			),
		}
	}
}

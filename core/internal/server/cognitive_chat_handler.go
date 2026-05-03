package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/mycelis/core/pkg/protocol"
)

// POST /api/v1/chat
// Routes user messages exclusively through the Admin agent via NATS request-reply.
// The Admin agent has its full system prompt, tools, and council access.
// No raw LLM fallback — if the swarm is offline, the endpoint returns an error.
//
// The full conversation history is forwarded as JSON so the admin agent can
// maintain multi-turn context. The NATS payload is a JSON array of
// {role, content} objects; the agent's handleDirectRequest detects JSON arrays
// and reconstructs prior turns.
func (s *AdminServer) HandleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Messages       []chatRequestMessage `json:"messages"`
		SessionID      string               `json:"session_id,omitempty"`
		OrganizationID string               `json:"organization_id,omitempty"`
		TeamID         string               `json:"team_id,omitempty"`
		TeamName       string               `json:"team_name,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad JSON", http.StatusBadRequest)
		return
	}

	if len(req.Messages) == 0 {
		http.Error(w, "Empty conversation", http.StatusBadRequest)
		return
	}

	sessionID, sessionIDValid := validateOptionalChatSessionID(req.SessionID)
	if !sessionIDValid {
		respondAPIError(w, "Invalid session_id: expected UUID", http.StatusBadRequest)
		return
	}

	sessionTurnIndex := 0
	if sessionID != "" && s.Conversations != nil {
		if priorTurns, err := s.Conversations.GetSessionTurns(r.Context(), sessionID); err != nil {
			log.Printf("[chat] prior session conversation lookup failed: %v", err)
		} else {
			sessionTurnIndex = len(priorTurns)
			req.Messages = mergePersistedSessionMessages(req.Messages, priorTurns)
		}
	}

	latestUserText := latestUserMessageContent(req.Messages)
	if isRuntimeStateQuestion(latestUserText) {
		s.respondRuntimeStateSummary(w, r, req.OrganizationID, req.TeamID, req.TeamName)
		return
	}
	if isSearchCapabilityQuestion(latestUserText) {
		s.respondSearchCapabilitySummary(w, r)
		return
	}
	if query, ok := directSearchQuery(latestUserText); ok {
		s.respondDirectSearchAnswer(w, r, query)
		return
	}

	referentialReview := s.buildSomaReferentialReview(r.Context(), req.Messages)
	if referentialReview.NeedsConfirmation {
		logSomaConversationTurn(r.Context(), s.Conversations, sessionID, sessionTurnIndex, "user", latestUserText, chatAgentResult{})
		s.respondReferentialConfirmation(w, r, referentialReview)
		return
	}
	if referentialReview.Confirmed {
		req.Messages = applyConfirmedReferentialAction(req.Messages, referentialReview)
		latestUserText = referentialReview.EffectiveRequest
	}

	logSomaConversationTurn(r.Context(), s.Conversations, sessionID, sessionTurnIndex, "user", latestUserText, chatAgentResult{})

	req.Messages = prependReferentialReviewContext(req.Messages, referentialReview)
	req.Messages = prependChatWorkspaceContext(
		req.Messages,
		s.buildChatWorkspaceContext(req.OrganizationID, req.TeamID, req.TeamName),
	)

	if availability := s.chatExecutionAvailability(); !availability.Available {
		respondAPIJSON(w, http.StatusServiceUnavailable, protocol.APIResponse{
			OK:    false,
			Error: availability.Summary,
			Data:  availability,
		})
		return
	}

	// 2. NATS must be available — the Admin agent is the ONLY path for chat.
	// No raw LLM fallback: agents must always operate within their context
	// (system prompt, tools, input/output rules).
	if s.NC == nil {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error":"Swarm offline — Admin agent unavailable. Start the organism first."}`, http.StatusServiceUnavailable)
		return
	}

	profile := userGovernanceProfileFromRequest(r)
	normalizedMessages, requestMutationTools := normalizeChatRequestMessages(req.Messages)
	normalizedMessages = applyGovernanceProfileToLatestMessage(normalizedMessages, profile)
	if len(normalizedMessages) > 0 {
		req.Messages = normalizedMessages
	}

	subject := fmt.Sprintf(protocol.TopicCouncilRequestFmt, "admin")
	agentResult, err := s.requestChatAgent(r.Context(), subject, req.Messages)
	if err != nil {
		log.Printf("Chat via Admin agent failed: %v", err)
		respondChatTransportBlocker(w, "Soma", err)
		return
	}

	if shouldRetryDirectAnswer(agentResult, requestMutationTools) {
		retryMessages := applyDirectAnswerRetryInstruction(req.Messages, latestUserText)
		retryResult, retryErr := s.requestChatAgent(r.Context(), subject, retryMessages)
		if retryErr == nil {
			agentResult = retryResult
		}
		if shouldRetryDirectAnswer(agentResult, requestMutationTools) {
			if isWeakDirectAnswerFallback(agentResult.Text) {
				respondStructuredChatBlocker(w, agentResult)
			} else {
				respondStructuredChatBlocker(w, directAnswerDriftBlocker(agentResult))
			}
			return
		}
	}

	isMutation, mutTools := mergeMutationTools(agentResult.ToolsUsed, requestMutationTools)
	if agentResult.Availability != nil && !agentResult.Availability.Available && (!isMutation || agentResult.Availability.Code != emptyProviderOutputCode) {
		respondStructuredChatBlocker(w, agentResult)
		return
	}
	if !isMutation && strings.TrimSpace(agentResult.Text) == "" && len(agentResult.Artifacts) == 0 {
		respondStructuredChatBlocker(w, agentResult)
		return
	}

	replyText := readableChatText(agentResult, isMutation)
	logSomaConversationTurn(r.Context(), s.Conversations, sessionID, sessionTurnIndex+1, "assistant", replyText, agentResult)

	chatPayload := protocol.ChatResponsePayload{
		Text:          replyText,
		ToolsUsed:     mutTools,
		Artifacts:     agentResult.Artifacts,
		Consultations: agentResult.Consultations,
	}

	applyBrainProvenance(s, &chatPayload, agentResult)

	askContract := resolveChatAskContract("soma", isMutation, agentResult)
	chatPayload.AskClass = askContract.AskClass
	templateID := askContract.TemplateID
	mode := askContract.DefaultExecutionMode

	if isMutation {
		plannedToolCalls := buildPlannedToolCalls(agentResult, latestUserText, mutTools)
		approval := buildApprovalPolicy(profile, plannedToolCalls, mutTools)
		scope := &protocol.ScopeValidation{
			Tools:             mutTools,
			AffectedResources: affectedResourcesForPlannedCalls(plannedToolCalls),
			RiskLevel:         chatToolRisk(mutTools),
			PlannedToolCalls:  plannedToolCalls,
			Approval:          approval,
			GovernanceProfile: profile.snapshot(),
		}
		if approval != nil {
			scope.CapabilityIDs = approval.CapabilityIDs
			scope.ExternalDataUse = approval.ExternalDataUse
			scope.EstimatedCost = approval.EstimatedCost
		}

		auditEventID, _ := s.createAuditEvent(
			protocol.TemplateChatToProposal, "admin",
			"Chat mutation detected",
			map[string]any{
				"tools":           mutTools,
				"agent_tools":     agentResult.ToolsUsed,
				"requested_tools": requestMutationTools,
				"actor":           "Soma",
				"user":            auditUserLabelFromRequest(r),
				"ask_class":       string(askContract.AskClass),
				"action":          "proposal_generated",
				"result_status":   "pending",
				"approval_status": approvalStatusValue(approval),
				"approval_reason": approvalReasonValue(approval),
				"capability_used": strings.Join(scope.CapabilityIDs, ","),
			},
		)

		proof, _ := s.createIntentProof(protocol.TemplateChatToProposal, "chat-action", scope, auditEventID)
		var confirmToken *protocol.ConfirmToken
		if proof != nil {
			confirmToken, _ = s.generateConfirmToken(proof.ID, protocol.TemplateChatToProposal)
		}

		var proofID string
		var token string
		if proof != nil {
			proofID = proof.ID
		}
		if confirmToken != nil {
			token = confirmToken.Token
		}
		display := buildProposalDisplayContract(plannedToolCalls, latestUserText, mutTools)
		chatPayload.Proposal = buildMutationChatProposal(mutTools, proofID, token, "admin-core", []string{"admin"}, approval, profile.snapshot(), display)

		chatPayload.Provenance = &protocol.AnswerProvenance{
			ResolvedIntent:  "proposal",
			PermissionCheck: "pass",
			PolicyDecision:  policyDecisionForApproval(approval),
			AuditEventID:    auditEventID,
		}
	} else {
		auditEventID, _ := s.createAuditEvent(
			protocol.TemplateChatToAnswer, "admin",
			"Admin chat",
			map[string]any{
				"tools":         agentResult.ToolsUsed,
				"actor":         "Soma",
				"user":          auditUserLabelFromRequest(r),
				"ask_class":     string(askContract.AskClass),
				"action":        "answer_delivered",
				"result_status": "completed",
			},
		)
		chatPayload.Provenance = &protocol.AnswerProvenance{
			ResolvedIntent:  "answer",
			PermissionCheck: "pass",
			PolicyDecision:  "allow",
			AuditEventID:    auditEventID,
		}
		if len(agentResult.Artifacts) > 0 {
			for _, artifact := range agentResult.Artifacts {
				_, _ = s.createAuditEvent(
					protocol.TemplateChatToAnswer, "admin",
					"Chat artifact created",
					map[string]any{
						"actor":           "Soma",
						"user":            auditUserLabelFromRequest(r),
						"action":          "artifact_created",
						"result_status":   "completed",
						"capability_used": "artifact_output",
						"resource":        strings.TrimSpace(artifact.Title),
						"details":         map[string]any{"artifact_type": artifact.Type},
					},
				)
			}
		}
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
		TemplateID: templateID,
		Mode:       mode,
	}

	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(envelope))
}

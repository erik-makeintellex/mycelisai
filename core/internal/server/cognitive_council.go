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

// ---------------------------------------------------------------------------
// Council Chat API — standardized, CTS-enveloped council interaction
// ---------------------------------------------------------------------------

// CouncilMemberInfo is returned by HandleListCouncilMembers.
type CouncilMemberInfo struct {
	ID   string `json:"id"`
	Role string `json:"role"`
	Team string `json:"team"`
}

// isCouncilMember checks whether memberID belongs to a standing council team
// (admin-core or council-core). Returns the team ID and role on match.
// Dynamic: add a new member to the YAML, restart, done.
func (s *AdminServer) isCouncilMember(memberID string) (teamID string, role string, ok bool) {
	if s.Soma == nil {
		return "", "", false
	}
	for _, tm := range s.Soma.ListTeams() {
		if tm.ID != "admin-core" && tm.ID != "council-core" {
			continue
		}
		for _, m := range tm.Members {
			if m.ID == memberID {
				return tm.ID, m.Role, true
			}
		}
	}
	return "", "", false
}

// GET /api/v1/council/members
// Returns all addressable council members from standing teams.
func (s *AdminServer) HandleListCouncilMembers(w http.ResponseWriter, r *http.Request) {
	if s.Soma == nil {
		respondAPIError(w, "Swarm offline", http.StatusServiceUnavailable)
		return
	}

	var members []CouncilMemberInfo
	for _, tm := range s.Soma.ListTeams() {
		if tm.ID != "admin-core" && tm.ID != "council-core" {
			continue
		}
		for _, m := range tm.Members {
			members = append(members, CouncilMemberInfo{
				ID:   m.ID,
				Role: m.Role,
				Team: tm.ID,
			})
		}
	}

	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(members))
}

// POST /api/v1/council/{member}/chat
// Routes user conversation to a specific council member via NATS request-reply.
// Returns a CTS envelope wrapped in APIResponse with trust score and provenance.
func (s *AdminServer) HandleCouncilChat(w http.ResponseWriter, r *http.Request) {
	memberID := r.PathValue("member")
	if memberID == "" {
		respondAPIError(w, "Missing council member ID", http.StatusBadRequest)
		return
	}

	// Validate member exists in standing council teams
	teamID, _, ok := s.isCouncilMember(memberID)
	if !ok {
		respondAPIError(w, fmt.Sprintf("Unknown council member: %s", memberID), http.StatusNotFound)
		return
	}

	var req struct {
		Messages []chatRequestMessage `json:"messages"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondAPIError(w, "Bad JSON", http.StatusBadRequest)
		return
	}
	if len(req.Messages) == 0 {
		respondAPIError(w, "Empty conversation", http.StatusBadRequest)
		return
	}

	if availability := s.chatExecutionAvailability(); !availability.Available {
		respondAPIJSON(w, http.StatusServiceUnavailable, protocol.APIResponse{
			OK:    false,
			Error: availability.Summary,
			Data:  availability,
		})
		return
	}

	// NATS must be available
	if s.NC == nil {
		respondAPIError(w, "Swarm offline — council agents unavailable. Start the organism first.", http.StatusServiceUnavailable)
		return
	}

	profile := userGovernanceProfileFromRequest(r)
	normalizedMessages, requestMutationTools := normalizeChatRequestMessages(req.Messages)
	normalizedMessages = applyGovernanceProfileToLatestMessage(normalizedMessages, profile)
	latestUserText := latestUserMessageContent(req.Messages)
	if len(normalizedMessages) > 0 {
		req.Messages = normalizedMessages
	}

	subject := fmt.Sprintf(protocol.TopicCouncilRequestFmt, memberID)
	agentResult, err := s.requestChatAgent(r.Context(), subject, req.Messages)
	if err != nil {
		log.Printf("Council chat with %s failed: %v", memberID, err)
		respondChatTransportBlocker(w, fmt.Sprintf("Council member %s", memberID), err)
		return
	}

	if shouldRetryDirectAnswer(agentResult, requestMutationTools) {
		retryMessages := applyDirectAnswerRetryInstruction(req.Messages, latestUserText)
		retryResult, retryErr := s.requestChatAgent(r.Context(), subject, retryMessages)
		if retryErr == nil {
			agentResult = retryResult
		}
	}

	if shouldRetryDirectAnswer(agentResult, requestMutationTools) {
		if isWeakDirectAnswerFallback(agentResult.Text) {
			respondStructuredChatBlocker(w, agentResult)
		} else {
			respondStructuredChatBlocker(w, directAnswerDriftBlocker(agentResult))
		}
		return
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

	// Wrap response in CTS envelope with trust score, provenance, and tool metadata
	chatPayload := protocol.ChatResponsePayload{
		Text:          readableChatText(agentResult, isMutation),
		ToolsUsed:     mutTools,
		Artifacts:     agentResult.Artifacts,
		Consultations: agentResult.Consultations,
	}

	applyBrainProvenance(s, &chatPayload, agentResult)

	askContract := resolveChatAskContract("specialist", isMutation, agentResult)
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
			protocol.TemplateChatToProposal, memberID,
			fmt.Sprintf("Council chat mutation detected from %s", memberID),
			map[string]any{
				"tools":           mutTools,
				"agent_tools":     agentResult.ToolsUsed,
				"requested_tools": requestMutationTools,
				"member":          memberID,
				"team":            teamID,
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
		chatPayload.Proposal = buildMutationChatProposal(mutTools, proofID, token, teamID, []string{memberID}, approval, profile.snapshot(), display)

		chatPayload.Provenance = &protocol.AnswerProvenance{
			ResolvedIntent:  "proposal",
			PermissionCheck: "pass",
			PolicyDecision:  policyDecisionForApproval(approval),
			AuditEventID:    auditEventID,
		}
	} else {
		auditEventID, _ := s.createAuditEvent(
			protocol.TemplateChatToAnswer, memberID,
			fmt.Sprintf("Council chat with %s", memberID),
			map[string]any{
				"tools":         agentResult.ToolsUsed,
				"member":        memberID,
				"team":          teamID,
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
					protocol.TemplateChatToAnswer, memberID,
					fmt.Sprintf("Council artifact created by %s", memberID),
					map[string]any{
						"actor":           "Soma",
						"user":            auditUserLabelFromRequest(r),
						"action":          "artifact_created",
						"result_status":   "completed",
						"capability_used": "artifact_output",
						"resource":        strings.TrimSpace(artifact.Title),
						"details":         map[string]any{"artifact_type": artifact.Type, "member": memberID, "team": teamID},
					},
				)
			}
		}
	}

	payloadBytes, _ := json.Marshal(chatPayload)

	envelope := protocol.CTSEnvelope{
		Meta: protocol.CTSMeta{
			SourceNode: memberID,
			Timestamp:  time.Now(),
		},
		SignalType: protocol.SignalChatResponse,
		TrustScore: protocol.TrustScoreCognitive,
		Payload:    payloadBytes,
		TemplateID: templateID,
		Mode:       mode,
	}

	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(envelope))
	log.Printf("Council chat: member=%s team=%s trust=%.1f tools=%v template=%s", memberID, teamID, envelope.TrustScore, agentResult.ToolsUsed, envelope.TemplateID)
}

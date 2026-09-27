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
		respondBlocker(w, r, http.StatusServiceUnavailable, codeTeamServiceOffline, "Soma runtime is not attached", nil)
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
			Data:  availabilityForViewer(r, availability),
		})
		return
	}

	// NATS must be available
	if s.NC == nil {
		respondBlocker(w, r, http.StatusServiceUnavailable, codeTeamServiceOffline, "NATS is not connected; council agents are unreachable", nil)
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
		respondChatTransportBlocker(w, r, fmt.Sprintf("Council member %s", memberID), err)
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
	if (agentResult.Availability != nil && !agentResult.Availability.Available) || (isMutation && agentReplyIsEmpty(agentResult)) {
		respondStructuredChatBlocker(w, agentResult)
		return
	}
	isMutation, mutTools, plannedToolCalls := executableMutationPlan(isMutation, agentResult, latestUserText, mutTools)
	if !isMutation && strings.TrimSpace(agentResult.Text) == "" && len(agentResult.Artifacts) == 0 {
		respondStructuredChatBlocker(w, agentResult)
		return
	}
	var draftPreviews []protocol.ProposalDraftPreview

	// Wrap response in CTS envelope with trust score, provenance, and tool metadata
	chatPayload := protocol.ChatResponsePayload{
		Text:          readableChatText(agentResult, isMutation),
		ResponseDepth: inferResponseDepthFromRequest(latestUserText, isMutation),
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
		effectiveTools := toolsForPlannedCalls(plannedToolCalls, mutTools)
		chatPayload.ToolsUsed = effectiveTools
		display := buildProposalDisplayContract(plannedToolCalls, latestUserText, effectiveTools)
		// A2b item 2: council runs the same posture seams as Soma chat, before
		// anything is audited or minted. An unresolvable template fails closed.
		if !s.applyCouncilOutcomeTemplateOrRespond(w, r, req.Messages, latestUserText, teamID, &display) {
			return
		}
		var draftBlocker *writeFileDraftBlocker
		if draftPreviews, draftBlocker = s.draftMissingWriteFileContent(r.Context(), plannedToolCalls, latestUserText); draftBlocker != nil {
			respondWriteFileDraftBlocker(w, r, draftBlocker)
			return
		}
		approval := buildApprovalPolicy(profile, plannedToolCalls, effectiveTools)
		approval = applyApproverTier(applyPostureApprovalFloor(approval, display.WorkIntent, s.Guard, effectiveTools))
		scope := &protocol.ScopeValidation{
			Tools:             effectiveTools,
			AffectedResources: affectedResourcesForPlannedCalls(plannedToolCalls),
			RiskLevel:         chatToolRisk(effectiveTools),
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
				"tools":           effectiveTools,
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
			confirmToken, _ = s.generateConfirmToken(proof.ID, protocol.TemplateChatToProposal, confirmTokenMint{Purpose: tokenPurposeChatAction, MintedBy: auditActorIDFromRequest(r)})
		}

		var proofID string
		var token string
		if proof != nil {
			proofID = proof.ID
		}
		if confirmToken != nil {
			token = confirmToken.Token
		}
		chatPayload.Proposal = withDraftPreviews(buildMutationChatProposal(effectiveTools, proofID, token, teamID, []string{memberID}, approval, profile.snapshot(), display), draftPreviews)

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

// codeCouncilTemplateUnresolved marks council template work that Core could not
// resolve or scope-validate; nothing was audited or minted.
const codeCouncilTemplateUnresolved = "outcome_template_unresolved"

// applyCouncilOutcomeTemplateOrRespond binds template work in a council turn to
// its Outcome Template (A2b). Council carries no session or organization, so a
// template it cannot resolve or scope-validate is refused with 409 and nothing
// is minted.
func (s *AdminServer) applyCouncilOutcomeTemplateOrRespond(w http.ResponseWriter, r *http.Request,
	messages []chatRequestMessage, latestUserText, teamID string, display *proposalDisplayContract) bool {
	if _, err := s.applyThreadOutcomeTemplate(r.Context(), "", messages, latestUserText, "", teamID,
		auditActorIDFromRequest(r), display); err != nil {
		respondBlocker(w, r, http.StatusConflict, codeCouncilTemplateUnresolved, err.Error(), nil)
		return false
	}
	return true
}

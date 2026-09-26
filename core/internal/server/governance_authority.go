package server

import (
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/mycelis/core/internal/governance"
	"github.com/mycelis/core/pkg/protocol"
	"gopkg.in/yaml.v3"
)

// Governance authority (slice A2a). Root admin role plus these scopes; an
// admin holding "*" covers all. Denials (401/403) happen before any read,
// write, audit, or Guard change.
const (
	scopeGovernanceRead  = "governance:read"
	scopeGovernanceWrite = "governance:write"
	scopeApprovalsDecide = "approvals:decide"

	maxPolicyBodyBytes = 1 << 20

	governancePolicyUnavailableCode   = "governance_policy_unavailable"
	governancePolicyRecommendedAction = "Fix core/config/policy.yaml and restart Core, or PUT a valid policy as an admin"
	governanceAuditUnavailableCode    = "governance_audit_unavailable"
)

// requireApprover gates approval decisions to root admins with approvals:decide.
func requireApprover(w http.ResponseWriter, r *http.Request) (*RequestIdentity, bool) {
	return requireRootAdminScope(w, r, scopeApprovalsDecide)
}

func respondGovernanceError(w http.ResponseWriter, status int, msg, code, action string) {
	data := map[string]string{"code": code}
	if action != "" {
		data["recommended_action"] = action
	}
	respondAPIJSON(w, status, protocol.APIResponse{OK: false, Error: msg, Data: data})
}

// governanceServiceStatus is the services/status row. Detail is fixed text and
// never carries raw load errors or policy content.
func (s *AdminServer) governanceServiceStatus() ServiceStatus {
	if s.Guard == nil || s.Guard.Degraded() {
		return ServiceStatus{Name: "governance", Status: "degraded", Detail: governance.PolicyUnavailableCode + ": " + governancePolicyRecommendedAction}
	}
	return ServiceStatus{Name: "governance", Status: "online", Detail: "Governance policy loaded"}
}

// canonicalPolicy returns the YAML bytes written to disk and their sha256.
func canonicalPolicy(cfg *governance.PolicyConfig) ([]byte, string, error) {
	if cfg == nil {
		return nil, "", nil
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, "", err
	}
	sum := sha256.Sum256(data)
	return data, hex.EncodeToString(sum[:]), nil
}

func postureGroupCount(cfg *governance.PolicyConfig) int {
	count := 0
	for _, group := range cfg.Groups {
		for _, target := range group.Targets {
			if strings.HasPrefix(target, governance.PostureTargetPrefix) {
				count++
				break
			}
		}
	}
	return count
}

// writePolicyFileAtomic writes to a temp file in the same directory and
// renames it over path, so readers never see a partial policy.
func writePolicyFileAtomic(path string, data []byte) (err error) {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".policy-*.yaml.tmp")
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = os.Remove(tmp.Name())
		}
	}()
	if _, err = tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err = tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err = tmp.Close(); err != nil {
		return err
	}
	if err = os.Chmod(tmp.Name(), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}

// auditGovernance writes an audit row with the actor identity. It returns ""
// when the audit store is unavailable or the insert fails.
func (s *AdminServer) auditGovernance(r *http.Request, source, message string, ctx map[string]any) string {
	ctx["user"] = auditUserLabelFromRequest(r)
	id, err := s.createAuditEvent(protocol.TemplateChatToProposal, source, message, attachActorIdentity(ctx, r))
	if err != nil {
		return ""
	}
	return id
}

// resolveGuardApproval is the single decision path for Guard-parked messages,
// shared by POST /api/v1/governance/resolve/{id} and the legacy
// POST /admin/approvals/{id} alias. Callers have already passed requireApprover
// and validated action as APPROVE or REJECT.
func (s *AdminServer) resolveGuardApproval(w http.ResponseWriter, r *http.Request, reqID, action, route string) {
	if s.Guard == nil || s.Router == nil {
		respondGovernanceError(w, http.StatusServiceUnavailable, "Governance approvals are not available", "governance_unavailable", "")
		return
	}
	pending, ok := s.Guard.PendingRequest(reqID)
	if !ok {
		respondAPIError(w, "Approval request not found", http.StatusNotFound)
		return
	}
	auditCtx := map[string]any{"action": "governance_approval_resolved", "request_id": reqID, "decision": action, "route": route}
	if msg := pending.GetOriginalMessage(); msg != nil {
		auditCtx["team_id"] = msg.TeamId
		auditCtx["source_agent"] = msg.SourceAgentId
		if msg.GetEvent() != nil {
			auditCtx["intent"] = msg.GetEvent().EventType
		}
	}
	result := func(status string) map[string]any {
		out := make(map[string]any, len(auditCtx)+1)
		for k, v := range auditCtx {
			out[k] = v
		}
		out["result_status"] = status
		return out
	}
	auditID := s.auditGovernance(r, "governance-approval", "Governance approval decision requested", result("requested"))
	if auditID == "" {
		respondGovernanceError(w, http.StatusServiceUnavailable, "Audit is unavailable; the approval was not resolved", governanceAuditUnavailableCode, "Restore the audit store and retry")
		return
	}
	approved := action == "APPROVE"
	msg, err := s.Guard.Resolve(reqID, approved, auditUserLabelFromRequest(r))
	if err != nil {
		s.auditGovernance(r, "governance-approval", "Governance approval request not found", result("not_found"))
		respondAPIError(w, "Approval request not found", http.StatusNotFound)
		return
	}
	if approved && msg != nil {
		if err := s.Router.PublishDirect(msg); err != nil {
			log.Printf("governance: re-publish of approved request %s failed: %v", reqID, err)
			s.auditGovernance(r, "governance-approval", "Governance approval re-publish failed", result("republish_failed"))
			respondAPIError(w, "Approved but the message could not be re-published", http.StatusInternalServerError)
			return
		}
	}
	s.auditGovernance(r, "governance-approval", "Governance approval resolved", result("resolved"))
	respondJSON(w, map[string]string{"status": "resolved", "request_id": reqID, "action": action, "audit_id": auditID})
}

// requiresApprover reports whether a confirm needs an approver: approval tier 2
// (policy/posture, high or critical capability risk, or cost above the
// approver ceiling). See approverTier (A2b item 1).
func requiresApprover(scope *protocol.ScopeValidation) bool {
	tier, _ := approverTier(scope)
	return tier == approverTierApprover
}

func isApprover(identity *RequestIdentity) bool {
	return identity != nil && identity.Role == "admin" && hasScope(identity, scopeApprovalsDecide)
}

// confirmerMayApprove is the confirm-action approver gate (A2a 3b). It runs
// after the scope loads and before any run, contract, worker, or tool effect.
// A denial writes a normalized blocker; the caller's transaction rolls back,
// so the confirm token is not consumed and an approver can confirm it later.
func confirmerMayApprove(w http.ResponseWriter, r *http.Request, scope *protocol.ScopeValidation, tok confirmTokenRow) bool {
	if !requiresApprover(scope) {
		if confirmerMayConfirmOwn(r, tok.MintedBy) { // A2b Q3: proposer-bound
			recordConfirmAuthority(r, scope, tok)
			return true
		}
		respondConfirmerNotProposer(w, r)
		return false
	}
	identity := IdentityFromContext(r.Context())
	if isApprover(identity) {
		recordConfirmAuthority(r, scope, tok)
		return true
	}
	respondApproverRequired(w, r)
	return false
}

// respondApproverRequired writes the normalized tier-2 blocker. Nothing ran and
// the token was not consumed, so an approver can confirm it later.
func respondApproverRequired(w http.ResponseWriter, r *http.Request) {
	identity := IdentityFromContext(r.Context())
	status := http.StatusForbidden
	if identity == nil {
		status = http.StatusUnauthorized
	}
	const (
		whatFailed = "This work needs admin approval. It was raised by an organization governance policy, a high-risk capability, or a cost above the approval limit, so only an admin with approval authority can confirm it."
		nextStep   = "Ask an admin to review and confirm this proposal. Nothing ran and the proposal is still valid."
	)
	retryable := true
	summary := protocol.ExecutionSummary{
		Intent:        protocol.ExecutionIntent{Resolved: "Confirm governed proposal"},
		Understanding: protocol.ExecutionUnderstanding{Summary: whatFailed},
		Execution:     protocol.ExecutionState{Shape: protocol.ExecutionShapeGuidedProposal, Status: protocol.ExecutionStatusBlocked, Summary: "Waiting for an admin approver"},
		AuditRecovery: protocol.AuditRecovery{
			ApprovalStatus: "approver_required",
			RecoveryState:  "awaiting_approver",
			Blocker:        whatFailed,
			Retryable:      &retryable,
			Degradation: &protocol.ExecutionDegradation{
				Code:              "approver_required",
				WhatFailed:        whatFailed,
				TrustedState:      "Nothing was executed or written.",
				SafeContinuation:  nextStep,
				RequiresAttention: true,
			},
		},
	}
	respondAPIJSON(w, status, protocol.APIResponse{OK: false, Error: "Needs admin approval", Data: map[string]any{
		"code":               "approver_required",
		"blocker":            "needs_admin_approval",
		"recommended_action": nextStep,
		"required_scope":     scopeApprovalsDecide,
		"confirmed":          false,
		"execution_state":    "blocked",
		"execution_summary":  summary,
	}})
}

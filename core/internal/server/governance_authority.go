package server

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/mycelis/core/internal/governance"
	"github.com/mycelis/core/internal/router"
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

	governancePolicyUnavailableCode = "governance_policy_unavailable"
	governanceAuditUnavailableCode  = "governance_audit_unavailable"
)

func respondGovernanceError(w http.ResponseWriter, status int, msg, code, action string) {
	data := map[string]string{"code": code}
	if action != "" {
		data["recommended_action"] = action
	}
	respondAPIJSON(w, status, protocol.APIResponse{OK: false, Error: msg, Data: data})
}

// governanceServiceStatus is the services/status row. Detail is fixed text and
// never carries raw load errors or policy content; only admins read the fix.
func (s *AdminServer) governanceServiceStatus(admin bool) ServiceStatus {
	if s.Guard == nil || s.Guard.Degraded() {
		return ServiceStatus{Name: "governance", Status: "degraded", Detail: governanceLockDetail(admin)}
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

// wireRouterApprovalAudit makes this server's audit store the Router's sink
// for REQUIRE_APPROVAL observations (C2-RETIRE). No Router, no wiring.
func (s *AdminServer) wireRouterApprovalAudit() {
	if s.Router != nil {
		s.Router.SetApprovalAuditor(s.auditPolicyApprovalObserved)
	}
}

// auditPolicyApprovalObserved records that a bus event matched a
// REQUIRE_APPROVAL rule and that Core did not act on it. It is an observation,
// not a request: there is no queue and nothing to approve here; governed work
// is approved on the durable confirm-action/proposal path. The fields come
// from the untrusted envelope and grant nothing. An empty id means the audit
// store is unavailable, and the Router then logs instead.
func (s *AdminServer) auditPolicyApprovalObserved(obs router.ApprovalObservation) (string, error) {
	return s.createAuditEvent(protocol.TemplateChatToProposal, "governance-guard",
		"Policy requires approval for a bus event; Core observed it and did not act on it",
		map[string]any{
			"action":        router.PolicyApprovalObservedAction,
			"actor":         "governance-guard",
			"user":          "system",
			"subject":       obs.Subject,
			"message_id":    obs.MessageID,
			"team_id":       obs.TeamID,
			"source_agent":  obs.SourceAgentID,
			"intent":        obs.Intent,
			"result_status": "not_acted_on",
			"approval_path": "proposals",
		})
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
	_, reason := approverTier(scope)
	respondApproverRequired(w, r, reason)
	return false
}

// respondApproverRequired writes the normalized tier-2 blocker. Nothing ran and
// the token was not consumed, so an approver can confirm it later. reason is
// the approverTier reason (policy | capability_risk | cost) or "" if unknown.
func respondApproverRequired(w http.ResponseWriter, r *http.Request, reason string) {
	identity := IdentityFromContext(r.Context())
	status := http.StatusForbidden
	if identity == nil {
		status = http.StatusUnauthorized
	}
	whatFailed, nextStep := approverRequiredCopy(reason, viewerIsAdmin(r))
	retryable := true
	summary := protocol.ExecutionSummary{
		Intent:        protocol.ExecutionIntent{Resolved: "Confirm governed proposal"},
		Understanding: protocol.ExecutionUnderstanding{Summary: whatFailed},
		Execution:     protocol.ExecutionState{Shape: protocol.ExecutionShapeGuidedProposal, Status: protocol.ExecutionStatusBlocked, Summary: "Waiting for an admin to approve"},
		AuditRecovery: protocol.AuditRecovery{
			ApprovalStatus: "approver_required",
			RecoveryState:  "awaiting_approver",
			Blocker:        whatFailed,
			Retryable:      &retryable,
			Degradation: &protocol.ExecutionDegradation{
				Code:              "approver_required",
				WhatFailed:        whatFailed,
				TrustedState:      "Nothing ran and nothing was changed.",
				SafeContinuation:  nextStep,
				RequiresAttention: true,
			},
		},
	}
	data := map[string]any{
		"code":               "approver_required",
		"blocker":            "needs_admin_approval",
		"recommended_action": nextStep,
		"required_scope":     scopeApprovalsDecide,
		"confirmed":          false,
		"execution_state":    "blocked",
		"execution_summary":  summary,
	}
	if reason != "" {
		data["approval_reason"] = reason
	}
	respondAPIJSON(w, status, protocol.APIResponse{OK: false, Error: "Needs admin approval", Data: data})
}

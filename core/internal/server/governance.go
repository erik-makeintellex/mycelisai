package server

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/mycelis/core/internal/governance"
	"github.com/mycelis/core/pkg/protocol"
)

// defaultPolicyPath is the disk location for persisting policy changes.
const defaultPolicyPath = "config/policy.yaml"

// pendingApprovalJSON is the simplified JSON representation of a pending approval request.
// It maps the complex proto ApprovalRequest into a cortex-friendly structure.
type pendingApprovalJSON struct {
	ID          string `json:"id"`
	Reason      string `json:"reason"`
	SourceAgent string `json:"source_agent"`
	TeamID      string `json:"team_id"`
	Intent      string `json:"intent"`
	Timestamp   string `json:"timestamp"`
	ExpiresAt   string `json:"expires_at"`
}

// handleGetPolicy returns the current governance policy configuration as JSON.
// GET /api/v1/governance/policy (root admin, governance:read)
func (s *AdminServer) handleGetPolicy(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireRootAdminScope(w, r, scopeGovernanceRead); !ok {
		return
	}
	if s.Guard == nil || s.Guard.Degraded() {
		respondGovernanceError(w, http.StatusServiceUnavailable, "Governance policy is unavailable", governancePolicyUnavailableCode, governancePolicyRecommendedAction)
		return
	}
	respondJSON(w, s.Guard.GetPolicyConfig())
}

// handleUpdatePolicy replaces the governance policy. Order is fail-closed:
// authority, bounded decode, validation, audit (requested), atomic file write,
// memory swap, audit result. A valid PUT also clears a degraded guard.
// PUT /api/v1/governance/policy (root admin, governance:write)
func (s *AdminServer) handleUpdatePolicy(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireRootAdminScope(w, r, scopeGovernanceWrite); !ok {
		return
	}
	if s.Guard == nil {
		respondGovernanceError(w, http.StatusServiceUnavailable, "Governance engine not initialized", "governance_unavailable", "")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxPolicyBodyBytes)
	var cfg governance.PolicyConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			respondAPIError(w, "policy body exceeds 1 MiB", http.StatusRequestEntityTooLarge)
			return
		}
		respondAPIError(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if err := governance.ValidatePolicyConfig(&cfg); err != nil {
		respondAPIError(w, err.Error(), http.StatusBadRequest)
		return
	}
	data, newDigest, err := canonicalPolicy(&cfg)
	if err != nil {
		respondAPIError(w, "policy could not be encoded", http.StatusBadRequest)
		return
	}
	_, previousDigest, _ := canonicalPolicy(s.Guard.GetPolicyConfig())
	auditCtx := func(status string) map[string]any {
		return map[string]any{
			"action":              "governance_policy_update",
			"previous_digest":     previousDigest,
			"new_digest":          newDigest,
			"group_count":         len(cfg.Groups),
			"posture_group_count": postureGroupCount(&cfg),
			"result_status":       status,
		}
	}
	auditID := s.auditGovernance(r, "governance-policy", "Governance policy update requested", auditCtx("requested"))
	if auditID == "" {
		respondGovernanceError(w, http.StatusServiceUnavailable, "Audit is unavailable; the policy was not changed", governanceAuditUnavailableCode, "Restore the audit store and retry")
		return
	}
	if err := s.Guard.ReplacePolicy(&cfg, func() error { return writePolicyFileAtomic(defaultPolicyPath, data) }); err != nil {
		log.Printf("governance: policy persist failed; live policy unchanged: %v", err)
		s.auditGovernance(r, "governance-policy", "Governance policy update failed", auditCtx("failed"))
		respondAPIError(w, "policy could not be persisted; the live policy was not changed", http.StatusInternalServerError)
		return
	}
	s.auditGovernance(r, "governance-policy", "Governance policy update applied", auditCtx("applied"))
	log.Printf("Governance policy updated and persisted to %s (digest %s)", defaultPolicyPath, newDigest)
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(map[string]string{"status": "applied", "digest": newDigest, "audit_id": auditID}))
}

// handleGetPendingApprovals returns all pending approval requests in a simplified JSON format.
// GET /api/v1/governance/pending (root admin, governance:read)
func (s *AdminServer) handleGetPendingApprovals(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireRootAdminScope(w, r, scopeGovernanceRead); !ok {
		return
	}
	if s.Guard == nil {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error":"Governance engine not initialized"}`, http.StatusServiceUnavailable)
		return
	}

	pending := s.Guard.ListPending()
	result := make([]pendingApprovalJSON, 0, len(pending))

	for _, req := range pending {
		item := pendingApprovalJSON{
			ID:     req.RequestId,
			Reason: req.Reason,
		}

		// Extract fields from the original message if present
		if msg := req.OriginalMessage; msg != nil {
			item.SourceAgent = msg.SourceAgentId
			item.TeamID = msg.TeamId
			if msg.GetEvent() != nil {
				item.Intent = msg.GetEvent().EventType
			}
			if msg.Timestamp != nil {
				item.Timestamp = msg.Timestamp.AsTime().Format(time.RFC3339)
			}
		}

		if req.ExpiresAt != nil {
			item.ExpiresAt = req.ExpiresAt.AsTime().Format(time.RFC3339)
		}

		result = append(result, item)
	}

	respondJSON(w, result)
}

// handleResolveApproval resolves a pending approval request by approving or rejecting it.
// POST /api/v1/governance/resolve/{id} (root admin, approvals:decide)
func (s *AdminServer) handleResolveApproval(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireApprover(w, r); !ok {
		return
	}
	reqID := r.PathValue("id")
	if reqID == "" {
		respondAPIError(w, "missing approval request ID", http.StatusBadRequest)
		return
	}
	action, ok := decodeApprovalDecision(w, r)
	if !ok {
		return
	}
	s.resolveGuardApproval(w, r, reqID, action, "/api/v1/governance/resolve")
}

// decodeApprovalDecision accepts only {"action":"APPROVE"|"REJECT"}.
func decodeApprovalDecision(w http.ResponseWriter, r *http.Request) (string, bool) {
	var payload struct {
		Action string `json:"action"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&payload); err != nil {
		respondAPIError(w, "invalid JSON body", http.StatusBadRequest)
		return "", false
	}
	if payload.Action != "APPROVE" && payload.Action != "REJECT" {
		respondAPIError(w, "action must be APPROVE or REJECT", http.StatusBadRequest)
		return "", false
	}
	return payload.Action, true
}

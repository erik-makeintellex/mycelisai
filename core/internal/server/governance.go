package server

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/mycelis/core/internal/governance"
	"github.com/mycelis/core/pkg/protocol"
)

// defaultPolicyPath is the disk location for persisting policy changes.
const defaultPolicyPath = "config/policy.yaml"

// handleGetPolicy returns the current governance policy configuration as JSON.
// GET /api/v1/governance/policy (root admin, governance:read)
func (s *AdminServer) handleGetPolicy(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireRootAdminScope(w, r, scopeGovernanceRead); !ok {
		return
	}
	if s.Guard == nil || s.Guard.Degraded() {
		respondBlocker(w, r, http.StatusServiceUnavailable, governancePolicyUnavailableCode, "", nil)
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
	var previousDigest, auditID string // set under applyMu (A2b item 4)
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
	err = s.Guard.ReplacePolicy(&cfg, func(previous *governance.PolicyConfig) error {
		_, previousDigest, _ = canonicalPolicy(previous)
		if auditID = s.auditGovernance(r, "governance-policy", "Governance policy update requested", auditCtx("requested")); auditID == "" {
			return governance.ErrAuditUnavailable
		}
		return writePolicyFileAtomic(defaultPolicyPath, data)
	})
	if errors.Is(err, governance.ErrAuditUnavailable) {
		respondGovernanceError(w, http.StatusServiceUnavailable, "Audit is unavailable; the policy was not changed", governanceAuditUnavailableCode, "Restore the audit store and retry")
		return
	}
	if err != nil {
		log.Printf("governance: policy persist failed; live policy unchanged: %v", err)
		s.auditGovernance(r, "governance-policy", "Governance policy update failed", auditCtx("failed"))
		respondAPIError(w, "policy could not be persisted; the live policy was not changed", http.StatusInternalServerError)
		return
	}
	s.auditGovernance(r, "governance-policy", "Governance policy update applied", auditCtx("applied"))
	log.Printf("Governance policy updated and persisted to %s (digest %s)", defaultPolicyPath, newDigest)
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(map[string]string{"status": "applied", "digest": newDigest, "audit_id": auditID}))
}

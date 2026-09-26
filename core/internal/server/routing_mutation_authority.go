package server

import (
	"log"
	"net/http"
	"sort"
	"strings"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/pkg/protocol"
)

// Routing mutation authority (S6d): every mutating /api/v1/brains route, the
// provider probe (outbound egress), and every mission-profile write require
// root admin + cognitive:write, and record an audit event before anything
// changes. Mutations that can move execution routing also hold
// profileOverrideWriteMu so they serialize with the S6c profile override
// PUT/DELETE handlers.

const routingMutationAuditSource = "cognitive-routing-mutation"

// providerBoundCode rejects disabling or deleting a provider that an
// execution profile currently resolves to.
const providerBoundCode = "provider_bound"

// routingResultFailedAfterCommit marks a failure audit for a mutation whose
// DB write committed but whose runtime apply failed.
const routingResultFailedAfterCommit = "failed_after_commit"

// Audit actions recorded before each routing mutation.
const (
	auditProviderToggled       = "cognitive_provider_toggled"
	auditProviderPolicyUpdated = "cognitive_provider_policy_updated"
	auditProviderAdded         = "cognitive_provider_added"
	auditProviderUpdated       = "cognitive_provider_updated"
	auditProviderDeleted       = "cognitive_provider_deleted"
	auditProviderProbed        = "cognitive_provider_probed"
	auditMissionProfileCreated = "mission_profile_created"
	auditMissionProfileUpdated = "mission_profile_updated"
	auditMissionProfileDeleted = "mission_profile_deleted"
	auditMissionProfileActive  = "mission_profile_activated"
)

type providerBoundRejection struct {
	ProviderID        string   `json:"provider_id"`
	Code              string   `json:"code"`
	Profiles          []string `json:"profiles"`
	RecommendedAction string   `json:"recommended_action"`
}

// requireRoutingWriter enforces root admin + cognitive:write. It writes
// 401/403 with no data and returns false when the caller may not mutate.
func requireRoutingWriter(w http.ResponseWriter, r *http.Request) bool {
	_, ok := requireRootAdminScope(w, r, cognitiveWriteScope)
	return ok
}

// auditRoutingMutation records the audit event before a routing mutation.
// It fails closed: when the event cannot be recorded (including no DB), it
// writes 503 and returns false, and the caller must change nothing.
func (s *AdminServer) auditRoutingMutation(w http.ResponseWriter, r *http.Request, action, message string, fields map[string]any) (string, bool) {
	ctx := map[string]any{
		"actor":         "operator",
		"user":          auditUserLabelFromRequest(r),
		"action":        action,
		"result_status": "requested",
	}
	for key, value := range fields {
		ctx[key] = value
	}
	auditID, err := s.createAuditEvent(protocol.TemplateChatToProposal, routingMutationAuditSource, message, attachActorIdentity(ctx, r))
	if err != nil || strings.TrimSpace(auditID) == "" {
		respondAPIError(w, "Audit unavailable: routing not changed", http.StatusServiceUnavailable)
		return "", false
	}
	return auditID, true
}

// recordRoutingMutationFailure records a failure audit event for a mutation
// that partly committed (result_status failed_after_commit). The response
// is already decided, so an audit error is logged, not returned.
func (s *AdminServer) recordRoutingMutationFailure(r *http.Request, action, message string, fields map[string]any) {
	ctx := map[string]any{
		"actor":         "operator",
		"user":          auditUserLabelFromRequest(r),
		"action":        action,
		"result_status": routingResultFailedAfterCommit,
	}
	for key, value := range fields {
		ctx[key] = value
	}
	if _, err := s.createAuditEvent(protocol.TemplateChatToProposal, routingMutationAuditSource, message, attachActorIdentity(ctx, r)); err != nil {
		log.Printf("routing mutation failure audit (%s) not recorded: %v", action, err)
	}
}

// providerConfigExecutable mirrors the config half of the resolver's
// providerConfiguredForExecution (enabled and a non-blank model). The
// adapter half always holds after a successful UpdateProvider, which
// rebuilds the adapter. TestProviderConfigExecutable_MatchesResolver pins
// the two together.
func providerConfigExecutable(cfg cognitive.ProviderConfig) bool {
	return cfg.Enabled && strings.TrimSpace(cfg.ModelID) != ""
}

// boundExecutionProfiles lists, sorted, the execution profiles whose
// effective route (after any explicit fallback) is providerID right now.
func (s *AdminServer) boundExecutionProfiles(providerID string) []string {
	routes, _ := s.Cognitive.ProfileRoutes()
	bound := make([]string, 0)
	for profile, route := range routes {
		if route.ExecutionProfile && strings.TrimSpace(route.ProviderID) == providerID {
			bound = append(bound, profile)
		}
	}
	sort.Strings(bound)
	return bound
}

// rejectIfProviderBound writes 409 provider_bound and returns true when an
// execution profile still resolves to providerID.
func (s *AdminServer) rejectIfProviderBound(w http.ResponseWriter, providerID string) bool {
	bound := s.boundExecutionProfiles(providerID)
	if len(bound) == 0 {
		return false
	}
	message := "provider " + providerID + " is bound to execution profiles: " + strings.Join(bound, ", ")
	respondAPIJSON(w, http.StatusConflict, protocol.APIResponse{
		OK:    false,
		Error: message,
		Data: providerBoundRejection{
			ProviderID: providerID,
			Code:       providerBoundCode,
			Profiles:   bound,
			RecommendedAction: "Re-point these profiles with PUT /api/v1/cognitive/profiles or reset them with " +
				"DELETE /api/v1/cognitive/profiles/{profile}/override, then retry.",
		},
	})
	return true
}

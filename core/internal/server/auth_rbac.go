package server

import (
	"net/http"
	"strings"
)

func roleRank(role string) int {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "owner", "admin":
		return 3
	case "operator":
		return 2
	case "viewer":
		return 1
	default:
		return 0
	}
}

// DefaultScopesForRole returns the canonical default capability scopes for a role.
func DefaultScopesForRole(role string) []string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "owner", "admin":
		return []string{"*"}
	case "operator":
		return []string{
			"organizations:*",
			"teams:*",
			"groups:*",
			"chat:*",
			"artifacts:*",
			"memory:*",
			"runs:*",
			"settings:read",
			"settings:write",
		}
	case "viewer":
		return []string{
			"organizations:read",
			"teams:read",
			"groups:read",
			"artifacts:read",
			"memory:read",
			"runs:read",
			"settings:read",
		}
	default:
		return nil
	}
}

// HasMinimumRole checks whether the caller identity meets or exceeds minRole
// by evaluating both EffectiveRole and Role against the canonical role rank.
func HasMinimumRole(identity *RequestIdentity, minRole string) bool {
	if identity == nil {
		return false
	}
	required := roleRank(minRole)
	if required == 0 {
		return false
	}
	rank := roleRank(identity.EffectiveRole)
	if r := roleRank(identity.Role); r > rank {
		rank = r
	}
	return rank >= required
}

// RequireMinimumRole enforces that an authenticated caller meets minRole and optional requiredScope.
func RequireMinimumRole(w http.ResponseWriter, r *http.Request, minRole string, requiredScope string) (*RequestIdentity, bool) {
	identity := IdentityFromContext(r.Context())
	if identity == nil {
		respondAPIError(w, "Authentication required", http.StatusUnauthorized)
		return nil, false
	}
	if !HasMinimumRole(identity, minRole) {
		respondAPIError(w, "Insufficient role: "+strings.TrimSpace(minRole)+" required", http.StatusForbidden)
		return nil, false
	}
	if requiredScope != "" && !hasScope(identity, requiredScope) {
		respondAPIError(w, "Missing required scope: "+requiredScope, http.StatusForbidden)
		return nil, false
	}
	return identity, true
}

func teamRoleForIdentity(identity *RequestIdentity) string {
	if identity == nil {
		return "observer"
	}
	switch strings.ToLower(strings.TrimSpace(identity.EffectiveRole)) {
	case "owner", "operator", "viewer":
		return strings.ToLower(strings.TrimSpace(identity.EffectiveRole))
	}
	switch strings.ToLower(strings.TrimSpace(identity.Role)) {
	case "owner", "admin":
		return "owner"
	case "operator", "viewer":
		return strings.ToLower(strings.TrimSpace(identity.Role))
	default:
		return "observer"
	}
}

func hasRestrictedSettingsMutation(input map[string]any) bool {
	for _, key := range []string{"access_management_tier", "product_edition", "identity_mode"} {
		if _, exists := input[key]; exists {
			return true
		}
	}
	return false
}

// isMutatingAPIRequest returns true for state-changing HTTP methods on governed API/admin routes,
// excluding explicitly read-only POST query endpoints.
func isMutatingAPIRequest(r *http.Request) bool {
	if r == nil || r.URL == nil {
		return false
	}
	switch r.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
	default:
		return false
	}
	path := r.URL.Path
	if !strings.HasPrefix(path, "/api/v1/") && !strings.HasPrefix(path, "/admin/") {
		return false
	}
	switch path {
	case "/api/v1/search":
		return false
	default:
		return true
	}
}

// requireOperatorIfAuthenticated enforces operator-or-higher RBAC when an authenticated identity
// is present in context, preserving direct unauthenticated unit-test invocation of handlers.
func requireOperatorIfAuthenticated(w http.ResponseWriter, r *http.Request, requiredScope string) bool {
	if IdentityFromContext(r.Context()) == nil {
		return true
	}
	_, ok := RequireMinimumRole(w, r, "operator", requiredScope)
	return ok
}

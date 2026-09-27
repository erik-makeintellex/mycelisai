package server

import (
	"net/http"
	"strings"
)

func hasScope(identity *RequestIdentity, required string) bool {
	if identity == nil || required == "" {
		return false
	}
	for _, scope := range identity.Scopes {
		scope = strings.TrimSpace(scope)
		if scope == "*" || scope == required {
			return true
		}
		if strings.HasSuffix(scope, ":*") {
			prefix := strings.TrimSuffix(scope, "*")
			if strings.HasPrefix(required, prefix) {
				return true
			}
		}
	}
	return false
}

func requireRootAdminScope(w http.ResponseWriter, r *http.Request, requiredScope string) (*RequestIdentity, bool) {
	identity := IdentityFromContext(r.Context())
	if identity == nil {
		respondAPIError(w, "Authentication required", http.StatusUnauthorized)
		return nil, false
	}
	// Both denials are code admin_required (UX1); only an admin viewer is
	// told which permission is missing.
	if identity.Role != "admin" {
		respondBlocker(w, r, http.StatusForbidden, codeAdminRequired, "", nil)
		return nil, false
	}
	if !hasScope(identity, requiredScope) {
		respondBlocker(w, r, http.StatusForbidden, codeAdminRequired, "Missing required scope: "+requiredScope,
			map[string]string{"required_scope": requiredScope})
		return nil, false
	}
	return identity, true
}

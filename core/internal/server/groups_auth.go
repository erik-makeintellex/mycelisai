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
	if identity.Role != "admin" {
		respondAPIError(w, "Root admin role required", http.StatusForbidden)
		return nil, false
	}
	if !hasScope(identity, requiredScope) {
		respondAPIError(w, "Missing required scope: "+requiredScope, http.StatusForbidden)
		return nil, false
	}
	return identity, true
}

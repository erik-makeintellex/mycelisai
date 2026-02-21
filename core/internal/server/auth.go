package server

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"strings"
)

// contextKey is an unexported type for context keys in this package.
type contextKey string

const ctxKeyIdentity contextKey = "identity"

// RequestIdentity represents the authenticated caller.
// Phase 0: single root user. Phase 1+ will resolve from JWT/session.
type RequestIdentity struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

// IdentityFromContext extracts the RequestIdentity from the request context.
// Returns nil if no identity is present (unauthenticated request).
func IdentityFromContext(ctx context.Context) *RequestIdentity {
	id, _ := ctx.Value(ctxKeyIdentity).(*RequestIdentity)
	return id
}

// AuthMiddleware enforces API key authentication on all requests except
// healthz and CORS preflight. Fail-closed: missing or invalid key = 401.
func AuthMiddleware(apiKey string, next http.Handler) http.Handler {
	apiKeyBytes := []byte(apiKey)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Exempt: health check
		if r.URL.Path == "/healthz" {
			next.ServeHTTP(w, r)
			return
		}

		// Exempt: CORS preflight
		if r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}

		// Extract token: Authorization header first, query param fallback
		token := ""
		if auth := r.Header.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") {
			token = strings.TrimPrefix(auth, "Bearer ")
		}
		if token == "" {
			token = r.URL.Query().Get("token")
		}

		if token == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "missing authentication token"})
			return
		}

		// Constant-time comparison to prevent timing attacks
		if subtle.ConstantTimeCompare([]byte(token), apiKeyBytes) != 1 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid authentication token"})
			return
		}

		// Phase 0: hardcoded root identity. Phase 1+ resolves from DB/JWT.
		identity := &RequestIdentity{
			UserID:   "00000000-0000-0000-0000-000000000000",
			Username: "admin",
			Role:     "admin",
		}

		ctx := context.WithValue(r.Context(), ctxKeyIdentity, identity)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

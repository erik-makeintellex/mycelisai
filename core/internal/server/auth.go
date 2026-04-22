package server

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"os"
	"strings"
)

// contextKey is an unexported type for context keys in this package.
type contextKey string

const ctxKeyIdentity contextKey = "identity"

// RequestIdentity represents the authenticated caller.
// Phase 0: single root user. Phase 1+ will resolve from JWT/session.
type RequestIdentity struct {
	UserID        string   `json:"user_id"`
	Username      string   `json:"username"`
	Role          string   `json:"role"`
	EffectiveRole string   `json:"effective_role,omitempty"`
	PrincipalType string   `json:"principal_type,omitempty"`
	AuthSource    string   `json:"auth_source,omitempty"`
	BreakGlass    bool     `json:"break_glass,omitempty"`
	Scopes        []string `json:"scopes,omitempty"`
}

type localAuthIdentityConfig struct {
	PrimaryAPIKey      string
	PrimaryUserID      string
	PrimaryUsername    string
	BreakGlassAPIKey   string
	BreakGlassUserID   string
	BreakGlassUsername string
	DeploymentContract DeploymentContract
}

func resolveLocalAuthIdentityConfig(primaryAPIKey string) localAuthIdentityConfig {
	cfg := localAuthIdentityConfig{
		PrimaryAPIKey:      primaryAPIKey,
		PrimaryUserID:      envOrDefaultIdentity("MYCELIS_LOCAL_ADMIN_USER_ID", "00000000-0000-0000-0000-000000000000"),
		PrimaryUsername:    envOrDefaultIdentity("MYCELIS_LOCAL_ADMIN_USERNAME", "admin"),
		BreakGlassAPIKey:   strings.TrimSpace(os.Getenv("MYCELIS_BREAK_GLASS_API_KEY")),
		BreakGlassUserID:   envOrDefaultIdentity("MYCELIS_BREAK_GLASS_USER_ID", "00000000-0000-0000-0000-000000000001"),
		BreakGlassUsername: envOrDefaultIdentity("MYCELIS_BREAK_GLASS_USERNAME", "recovery-admin"),
		DeploymentContract: ResolveDeploymentContract(),
	}
	return cfg
}

func envOrDefaultIdentity(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func (cfg localAuthIdentityConfig) identityForToken(token string) *RequestIdentity {
	switch {
	case token != "" && subtle.ConstantTimeCompare([]byte(token), []byte(cfg.PrimaryAPIKey)) == 1:
		return &RequestIdentity{
			UserID:        cfg.PrimaryUserID,
			Username:      cfg.PrimaryUsername,
			Role:          "admin",
			EffectiveRole: "owner",
			PrincipalType: "local_admin",
			AuthSource:    "local_api_key",
			Scopes:        []string{"*"},
		}
	case cfg.BreakGlassAPIKey != "" && subtle.ConstantTimeCompare([]byte(token), []byte(cfg.BreakGlassAPIKey)) == 1:
		return &RequestIdentity{
			UserID:        cfg.BreakGlassUserID,
			Username:      cfg.BreakGlassUsername,
			Role:          "admin",
			EffectiveRole: "owner",
			PrincipalType: "break_glass_admin",
			AuthSource:    "local_break_glass",
			BreakGlass:    true,
			Scopes:        []string{"*"},
		}
	default:
		return nil
	}
}

func (cfg localAuthIdentityConfig) authConfigurationError() string {
	if !cfg.DeploymentContract.RequiresBreakGlassRecovery() {
		return ""
	}
	if strings.TrimSpace(cfg.BreakGlassAPIKey) == "" {
		return "deployment auth contract requires MYCELIS_BREAK_GLASS_API_KEY for enterprise-like recovery posture"
	}
	return ""
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
	identityConfig := resolveLocalAuthIdentityConfig(apiKey)

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

		if configError := identityConfig.authConfigurationError(); configError != "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]string{"error": configError})
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

		identity := identityConfig.identityForToken(token)
		if identity == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid authentication token"})
			return
		}

		ctx := context.WithValue(r.Context(), ctxKeyIdentity, identity)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

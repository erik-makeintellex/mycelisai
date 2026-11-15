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

type identityConfiguration struct {
	PrimaryAPIKey      string
	PrimaryUserID      string
	PrimaryUsername    string
	BreakGlassAPIKey   string
	BreakGlassUserID   string
	BreakGlassUsername string
	OperatorAPIKey     string
	OperatorUserID     string
	OperatorUsername   string
	ViewerAPIKey       string
	ViewerUserID       string
	ViewerUsername     string
	DeploymentContract DeploymentContract
	OIDC               oidcAdapterConfig
}

type localAuthIdentityConfig = identityConfiguration

func loadIdentityConfigurationFromEnv() identityConfiguration {
	return resolveLocalAuthIdentityConfig(strings.TrimSpace(os.Getenv("MYCELIS_API_KEY")))
}

func resolveLocalAuthIdentityConfig(primaryAPIKey string) identityConfiguration {
	cfg := identityConfiguration{
		PrimaryAPIKey:      primaryAPIKey,
		PrimaryUserID:      envOrDefaultIdentity("MYCELIS_LOCAL_ADMIN_USER_ID", "00000000-0000-0000-0000-000000000000"),
		PrimaryUsername:    envOrDefaultIdentity("MYCELIS_LOCAL_ADMIN_USERNAME", "admin"),
		BreakGlassAPIKey:   strings.TrimSpace(os.Getenv("MYCELIS_BREAK_GLASS_API_KEY")),
		BreakGlassUserID:   envOrDefaultIdentity("MYCELIS_BREAK_GLASS_USER_ID", "00000000-0000-0000-0000-000000000001"),
		BreakGlassUsername: envOrDefaultIdentity("MYCELIS_BREAK_GLASS_USERNAME", "recovery-admin"),
		OperatorAPIKey:     strings.TrimSpace(os.Getenv("MYCELIS_OPERATOR_API_KEY")),
		OperatorUserID:     envOrDefaultIdentity("MYCELIS_OPERATOR_USER_ID", "00000000-0000-0000-0000-000000000002"),
		OperatorUsername:   envOrDefaultIdentity("MYCELIS_OPERATOR_USERNAME", "operator"),
		ViewerAPIKey:       strings.TrimSpace(os.Getenv("MYCELIS_VIEWER_API_KEY")),
		ViewerUserID:       envOrDefaultIdentity("MYCELIS_VIEWER_USER_ID", "00000000-0000-0000-0000-000000000003"),
		ViewerUsername:     envOrDefaultIdentity("MYCELIS_VIEWER_USERNAME", "viewer"),
		DeploymentContract: ResolveDeploymentContract(),
		OIDC:               loadOIDCAdapterConfigFromEnv(),
	}
	return cfg
}

func (cfg identityConfiguration) Enabled() bool {
	return strings.TrimSpace(cfg.PrimaryAPIKey) != "" ||
		strings.TrimSpace(cfg.BreakGlassAPIKey) != "" ||
		strings.TrimSpace(cfg.OperatorAPIKey) != "" ||
		strings.TrimSpace(cfg.ViewerAPIKey) != "" ||
		cfg.OIDC.Enabled()
}

func envOrDefaultIdentity(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func (cfg identityConfiguration) identityForToken(token string) *RequestIdentity {
	switch {
	case token != "" && cfg.PrimaryAPIKey != "" && subtle.ConstantTimeCompare([]byte(token), []byte(cfg.PrimaryAPIKey)) == 1:
		return &RequestIdentity{
			UserID:        cfg.PrimaryUserID,
			Username:      cfg.PrimaryUsername,
			Role:          "admin",
			EffectiveRole: "owner",
			PrincipalType: "local_admin",
			AuthSource:    "local_api_key",
			Scopes:        DefaultScopesForRole("owner"),
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
			Scopes:        DefaultScopesForRole("owner"),
		}
	case cfg.OperatorAPIKey != "" && subtle.ConstantTimeCompare([]byte(token), []byte(cfg.OperatorAPIKey)) == 1:
		return &RequestIdentity{
			UserID:        cfg.OperatorUserID,
			Username:      cfg.OperatorUsername,
			Role:          "operator",
			EffectiveRole: "operator",
			PrincipalType: "local_operator",
			AuthSource:    "local_api_key",
			Scopes:        DefaultScopesForRole("operator"),
		}
	case cfg.ViewerAPIKey != "" && subtle.ConstantTimeCompare([]byte(token), []byte(cfg.ViewerAPIKey)) == 1:
		return &RequestIdentity{
			UserID:        cfg.ViewerUserID,
			Username:      cfg.ViewerUsername,
			Role:          "viewer",
			EffectiveRole: "viewer",
			PrincipalType: "local_viewer",
			AuthSource:    "local_api_key",
			Scopes:        DefaultScopesForRole("viewer"),
		}
	default:
		if cfg.OIDC.Enabled() {
			return cfg.OIDC.verifyBearerToken(token)
		}
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
			LogOperationalError("auth_middleware", "Deployment auth posture blocked request.", configError, "Configure MYCELIS_BREAK_GLASS_API_KEY via uv run inv auth.break-glass-key and restart services.")
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

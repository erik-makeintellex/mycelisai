package server

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"
)

// contextKey is an unexported type for context keys in this package.
type contextKey string

const ctxKeyIdentity contextKey = "identity"
const forwardedWebIdentityHeader = "X-Mycelis-Web-Identity"
const forwardedWebIdentitySignatureHeader = "X-Mycelis-Web-Identity-Signature"
const forwardedWebIdentityMaxAgeSeconds int64 = 10 * 60
const forwardedWebIdentityClockSkewSeconds int64 = 2 * 60

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

type forwardedWebIdentityPayload struct {
	Sub      string `json:"sub"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Role     string `json:"role"`
	Provider string `json:"provider"`
	HD       string `json:"hd,omitempty"`
	IAT      int64  `json:"iat"`
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

func forwardedWebIdentitySecret() string {
	if secret := strings.TrimSpace(os.Getenv("MYCELIS_WEB_IDENTITY_FORWARD_SECRET")); secret != "" {
		return secret
	}
	return strings.TrimSpace(os.Getenv("MYCELIS_WEB_SESSION_SECRET"))
}

func signedForwardedWebIdentityFromRequest(r *http.Request) (*RequestIdentity, bool, bool) {
	payload := strings.TrimSpace(r.Header.Get(forwardedWebIdentityHeader))
	signature := strings.TrimSpace(r.Header.Get(forwardedWebIdentitySignatureHeader))
	if payload == "" && signature == "" {
		return nil, false, true
	}
	secret := forwardedWebIdentitySecret()
	if payload == "" || signature == "" || secret == "" {
		return nil, true, false
	}
	expected := signForwardedWebIdentity(payload, secret)
	if subtle.ConstantTimeCompare([]byte(signature), []byte(expected)) != 1 {
		return nil, true, false
	}
	raw, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return nil, true, false
	}
	var forwarded forwardedWebIdentityPayload
	if err := json.Unmarshal(raw, &forwarded); err != nil {
		return nil, true, false
	}
	if !forwardedWebIdentityIsFresh(forwarded.IAT, time.Now().Unix()) {
		return nil, true, false
	}
	identity := requestIdentityFromForwardedWebIdentity(forwarded)
	if identity == nil {
		return nil, true, false
	}
	return identity, true, true
}

func forwardedWebIdentityIsFresh(iat int64, now int64) bool {
	if iat <= 0 {
		return false
	}
	if iat > now+forwardedWebIdentityClockSkewSeconds {
		return false
	}
	return now-iat <= forwardedWebIdentityMaxAgeSeconds
}

func signForwardedWebIdentity(payload, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func requestIdentityFromForwardedWebIdentity(payload forwardedWebIdentityPayload) *RequestIdentity {
	email := strings.TrimSpace(strings.ToLower(payload.Email))
	sub := strings.TrimSpace(payload.Sub)
	if email == "" && sub == "" {
		return nil
	}
	username := email
	if username == "" {
		username = strings.TrimSpace(payload.Name)
	}
	if username == "" {
		username = sub
	}
	role := "operator"
	effectiveRole := "operator"
	if strings.EqualFold(payload.Role, "admin") {
		role = "admin"
		effectiveRole = "owner"
	}
	provider := strings.TrimSpace(strings.ToLower(payload.Provider))
	authSource := "web_session"
	principalType := "web_user"
	switch provider {
	case "google":
		authSource = "web_google"
		principalType = "google_workspace_user"
	case "local":
		authSource = "web_local"
		principalType = "local_web_user"
	}
	scopes := []string{"soma:work", "runs:read", "outputs:read"}
	if role == "admin" {
		scopes = []string{"*"}
	}
	userID := sub
	if userID == "" {
		userID = email
	}
	return &RequestIdentity{
		UserID:        userID,
		Username:      username,
		Role:          role,
		EffectiveRole: effectiveRole,
		PrincipalType: principalType,
		AuthSource:    authSource,
		Scopes:        scopes,
	}
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
		if forwardedIdentity, attempted, ok := signedForwardedWebIdentityFromRequest(r); attempted {
			if !ok {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "invalid forwarded web identity"})
				return
			}
			identity = forwardedIdentity
		}

		ctx := context.WithValue(r.Context(), ctxKeyIdentity, identity)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

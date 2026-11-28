package server

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"os"
	"strings"
	"time"
)

type oidcAdapterConfig struct {
	Issuer      string
	Audience    string
	HS256Secret string
	JWKSURL     string
	JWKSJSON    string
	RoleClaim   string
	DefaultRole string
}

func loadOIDCAdapterConfigFromEnv() oidcAdapterConfig {
	roleClaim := strings.TrimSpace(os.Getenv("MYCELIS_OIDC_ROLE_CLAIM"))
	if roleClaim == "" {
		roleClaim = "role"
	}
	defaultRole := NormalizeRole(os.Getenv("MYCELIS_OIDC_DEFAULT_ROLE"))
	if defaultRole == "" {
		defaultRole = "viewer"
	}
	return oidcAdapterConfig{
		Issuer:      strings.TrimSpace(os.Getenv("MYCELIS_OIDC_ISSUER")),
		Audience:    strings.TrimSpace(os.Getenv("MYCELIS_OIDC_AUDIENCE")),
		HS256Secret: strings.TrimSpace(os.Getenv("MYCELIS_OIDC_HS256_SECRET")),
		JWKSURL:     strings.TrimSpace(os.Getenv("MYCELIS_OIDC_JWKS_URL")),
		JWKSJSON:    strings.TrimSpace(os.Getenv("MYCELIS_OIDC_JWKS_JSON")),
		RoleClaim:   roleClaim,
		DefaultRole: defaultRole,
	}
}

func (c oidcAdapterConfig) Enabled() bool {
	return strings.TrimSpace(c.HS256Secret) != "" ||
		strings.TrimSpace(c.JWKSURL) != "" ||
		strings.TrimSpace(c.JWKSJSON) != ""
}

// NormalizeRole maps raw role or group identifiers onto Mycelis canonical roles.
func NormalizeRole(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "owner", "root", "superadmin", "mycelis-owner", "mycelis:owner":
		return "owner"
	case "admin", "administrator", "mycelis-admin", "mycelis:admin":
		return "admin"
	case "operator", "ops", "mycelis-operator", "mycelis:operator":
		return "operator"
	case "viewer", "read", "readonly", "observer", "mycelis-viewer", "mycelis:viewer":
		return "viewer"
	default:
		return ""
	}
}

func oidcRolePriority(role string) int {
	switch role {
	case "owner":
		return 4
	case "admin":
		return 3
	case "operator":
		return 2
	case "viewer":
		return 1
	default:
		return 0
	}
}

func (c oidcAdapterConfig) verifyBearerToken(rawToken string) *RequestIdentity {
	if !c.Enabled() {
		return nil
	}
	token := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(rawToken), "Bearer "))
	parts := strings.Split(token, ".")
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return nil
	}

	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil
	}
	var header struct {
		Alg string `json:"alg"`
		Kid string `json:"kid"`
	}
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil
	}

	sigBytes, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil
	}
	signingInput := parts[0] + "." + parts[1]
	switch header.Alg {
	case "HS256":
		if strings.TrimSpace(c.HS256Secret) == "" {
			return nil
		}
		mac := hmac.New(sha256.New, []byte(c.HS256Secret))
		_, _ = mac.Write([]byte(signingInput))
		if !hmac.Equal(sigBytes, mac.Sum(nil)) {
			return nil
		}
	case "RS256":
		if !c.verifyRS256Signature(header.Kid, signingInput, sigBytes) {
			return nil
		}
	default:
		return nil
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil
	}
	var claims map[string]any
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil
	}

	if c.Issuer != "" {
		if iss, _ := claims["iss"].(string); strings.TrimSpace(iss) != c.Issuer {
			return nil
		}
	}
	if c.Audience != "" && !claimMatchesAudience(claims["aud"], c.Audience) {
		return nil
	}

	now := time.Now().Unix()
	if exp, ok := claimUnixTime(claims["exp"]); !ok || exp <= now {
		return nil
	}
	if rawNbf, exists := claims["nbf"]; exists && rawNbf != nil {
		if nbf, ok := claimUnixTime(rawNbf); !ok || nbf > now {
			return nil
		}
	}

	sub, _ := claims["sub"].(string)
	if sub = strings.TrimSpace(sub); sub == "" {
		return nil
	}

	role := c.resolveOIDCRole(claims)
	return &RequestIdentity{
		UserID:        sub,
		Username:      resolveOIDCUsername(claims, sub),
		Role:          role,
		EffectiveRole: role,
		Scopes:        DefaultScopesForRole(role),
		PrincipalType: "federated_user",
		AuthSource:    "oidc_jwt",
		BreakGlass:    false,
	}
}

func claimMatchesAudience(raw any, expected string) bool {
	switch typed := raw.(type) {
	case string:
		return strings.TrimSpace(typed) == expected
	case []any:
		for _, item := range typed {
			if s, ok := item.(string); ok && strings.TrimSpace(s) == expected {
				return true
			}
		}
	case []string:
		for _, s := range typed {
			if strings.TrimSpace(s) == expected {
				return true
			}
		}
	}
	return false
}

func claimUnixTime(raw any) (int64, bool) {
	switch typed := raw.(type) {
	case float64:
		return int64(typed), true
	case int64:
		return typed, true
	case int:
		return int64(typed), true
	case json.Number:
		v, err := typed.Int64()
		return v, err == nil
	default:
		return 0, false
	}
}

func resolveOIDCUsername(claims map[string]any, fallback string) string {
	for _, key := range []string{"email", "preferred_username", "name"} {
		if val, ok := claims[key].(string); ok && strings.TrimSpace(val) != "" {
			return strings.TrimSpace(val)
		}
	}
	return fallback
}

func (c oidcAdapterConfig) resolveOIDCRole(claims map[string]any) string {
	bestRole := ""
	roleClaimKey := strings.TrimSpace(c.RoleClaim)
	if roleClaimKey == "" {
		roleClaimKey = "role"
	}
	for _, key := range []string{roleClaimKey, "role", "roles", "groups"} {
		for _, candidate := range extractClaimStrings(claims[key]) {
			norm := NormalizeRole(candidate)
			if oidcRolePriority(norm) > oidcRolePriority(bestRole) {
				bestRole = norm
			}
		}
	}
	if bestRole != "" {
		return bestRole
	}
	if fallback := NormalizeRole(c.DefaultRole); fallback != "" {
		return fallback
	}
	return "viewer"
}

func extractClaimStrings(raw any) []string {
	switch typed := raw.(type) {
	case string:
		return []string{typed}
	case []any:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	case []string:
		return typed
	default:
		return nil
	}
}

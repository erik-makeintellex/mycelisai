package server

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func signTestHS256JWT(t *testing.T, secret string, header map[string]any, claims map[string]any) string {
	t.Helper()
	if header == nil {
		header = map[string]any{"alg": "HS256", "typ": "JWT"}
	}
	hb, err := json.Marshal(header)
	if err != nil {
		t.Fatalf("marshal header: %v", err)
	}
	pb, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("marshal claims: %v", err)
	}
	hEnc := base64.RawURLEncoding.EncodeToString(hb)
	pEnc := base64.RawURLEncoding.EncodeToString(pb)
	signingInput := hEnc + "." + pEnc

	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(signingInput))
	sigEnc := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return signingInput + "." + sigEnc
}

func newRBACProtectedTestRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/Resource", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if _, ok := RequireMinimumRole(w, r, "viewer", "teams:read"); !ok {
				return
			}
			w.WriteHeader(http.StatusOK)
		case http.MethodPost:
			if _, ok := RequireMinimumRole(w, r, "operator", "teams:write"); !ok {
				return
			}
			w.WriteHeader(http.StatusCreated)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/v1/admin/groups", func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireRootAdminScope(w, r, "groups:write"); !ok {
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	return mux
}

func TestOIDCJWT_OperatorRole_AllowedOnOperatorAndForbiddenOnAdmin(t *testing.T) {
	const secret = "enterprise-federation-hs256-secret"
	t.Setenv("MYCELIS_OIDC_ISSUER", "https://idp.example.com")
	t.Setenv("MYCELIS_OIDC_AUDIENCE", "mycelis-core")
	t.Setenv("MYCELIS_OIDC_HS256_SECRET", secret)

	token := signTestHS256JWT(t, secret, nil, map[string]any{
		"iss":   "https://idp.example.com",
		"aud":   "mycelis-core",
		"sub":   "user-op-42",
		"email": "operator@example.com",
		"role":  "operator",
		"exp":   time.Now().Add(15 * time.Minute).Unix(),
	})

	cfg := loadIdentityConfigurationFromEnv()
	if !cfg.Enabled() || !cfg.OIDC.Enabled() {
		t.Fatal("expected identity configuration and OIDC adapter to be enabled")
	}

	identity := cfg.identityForToken(token)
	if identity == nil {
		t.Fatal("expected verified RequestIdentity for valid operator JWT")
	}
	if identity.PrincipalType != "federated_user" {
		t.Fatalf("expected principal_type federated_user, got %q", identity.PrincipalType)
	}
	if identity.AuthSource != "oidc_jwt" {
		t.Fatalf("expected auth_source oidc_jwt, got %q", identity.AuthSource)
	}
	if identity.Role != "operator" {
		t.Fatalf("expected role operator, got %q", identity.Role)
	}
	if identity.UserID != "user-op-42" || identity.Username != "operator@example.com" {
		t.Fatalf("unexpected user identity fields: %+v", identity)
	}

	handler := AuthMiddleware("local-root-key", newRBACProtectedTestRouter())

	reqOp := httptest.NewRequest(http.MethodPost, "/api/v1/Resource", nil)
	reqOp.Header.Set("Authorization", "Bearer "+token)
	rrOp := httptest.NewRecorder()
	handler.ServeHTTP(rrOp, reqOp)
	assertStatus(t, rrOp, http.StatusCreated)

	reqAdmin := httptest.NewRequest(http.MethodPost, "/api/v1/admin/groups", nil)
	reqAdmin.Header.Set("Authorization", "Bearer "+token)
	rrAdmin := httptest.NewRecorder()
	handler.ServeHTTP(rrAdmin, reqAdmin)
	assertStatus(t, rrAdmin, http.StatusForbidden)
}

func TestOIDCJWT_RolesArrayViewer_ReadAllowedAndMutationForbidden(t *testing.T) {
	const secret = "enterprise-federation-hs256-secret"
	t.Setenv("MYCELIS_OIDC_ISSUER", "https://idp.example.com")
	t.Setenv("MYCELIS_OIDC_AUDIENCE", "mycelis-core")
	t.Setenv("MYCELIS_OIDC_HS256_SECRET", secret)

	token := signTestHS256JWT(t, secret, nil, map[string]any{
		"iss":                "https://idp.example.com",
		"aud":                []string{"mycelis-ui", "mycelis-core"},
		"sub":                "user-view-7",
		"preferred_username": "auditor-jane",
		"roles":              []string{"viewer"},
		"exp":                time.Now().Add(15 * time.Minute).Unix(),
	})

	handler := AuthMiddleware("local-root-key", newRBACProtectedTestRouter())

	reqGet := httptest.NewRequest(http.MethodGet, "/api/v1/Resource", nil)
	reqGet.Header.Set("Authorization", "Bearer "+token)
	rrGet := httptest.NewRecorder()
	handler.ServeHTTP(rrGet, reqGet)
	assertStatus(t, rrGet, http.StatusOK)

	reqPost := httptest.NewRequest(http.MethodPost, "/api/v1/Resource", nil)
	reqPost.Header.Set("Authorization", "Bearer "+token)
	rrPost := httptest.NewRecorder()
	handler.ServeHTTP(rrPost, reqPost)
	assertStatus(t, rrPost, http.StatusForbidden)
}

func TestOIDCJWT_GroupsAdmin_AllowedOnAdminRoute(t *testing.T) {
	const secret = "enterprise-federation-hs256-secret"
	t.Setenv("MYCELIS_OIDC_ISSUER", "https://idp.example.com")
	t.Setenv("MYCELIS_OIDC_HS256_SECRET", secret)

	token := signTestHS256JWT(t, secret, nil, map[string]any{
		"iss":    "https://idp.example.com",
		"sub":    "user-admin-1",
		"name":   "Platform Admin",
		"groups": []string{"viewer", "admin"},
		"exp":    time.Now().Add(15 * time.Minute).Unix(),
	})

	cfg := loadIdentityConfigurationFromEnv()
	identity := cfg.identityForToken(token)
	if identity == nil || identity.Role != "admin" {
		t.Fatalf("expected role admin from groups claim, got %+v", identity)
	}

	handler := AuthMiddleware("local-root-key", newRBACProtectedTestRouter())
	reqAdmin := httptest.NewRequest(http.MethodPost, "/api/v1/admin/groups", nil)
	reqAdmin.Header.Set("Authorization", "Bearer "+token)
	rrAdmin := httptest.NewRecorder()
	handler.ServeHTTP(rrAdmin, reqAdmin)
	assertStatus(t, rrAdmin, http.StatusOK)
}

func TestOIDCJWT_RejectsExpiredWrongSignatureWrongIssuerOrWrongAudience(t *testing.T) {
	const secret = "enterprise-federation-hs256-secret"
	t.Setenv("MYCELIS_OIDC_ISSUER", "https://idp.example.com")
	t.Setenv("MYCELIS_OIDC_AUDIENCE", "mycelis-core")
	t.Setenv("MYCELIS_OIDC_HS256_SECRET", secret)

	handler := AuthMiddleware("local-root-key", newRBACProtectedTestRouter())

	cases := []struct {
		name   string
		token  string
	}{
		{
			name: "expired_exp",
			token: signTestHS256JWT(t, secret, nil, map[string]any{
				"iss": "https://idp.example.com",
				"aud": "mycelis-core",
				"sub": "user-1",
				"exp": time.Now().Add(-5 * time.Minute).Unix(),
			}),
		},
		{
			name: "wrong_signature",
			token: signTestHS256JWT(t, "wrong-secret", nil, map[string]any{
				"iss": "https://idp.example.com",
				"aud": "mycelis-core",
				"sub": "user-1",
				"exp": time.Now().Add(15 * time.Minute).Unix(),
			}),
		},
		{
			name: "wrong_issuer",
			token: signTestHS256JWT(t, secret, nil, map[string]any{
				"iss": "https://rogue.example.com",
				"aud": "mycelis-core",
				"sub": "user-1",
				"exp": time.Now().Add(15 * time.Minute).Unix(),
			}),
		},
		{
			name: "wrong_audience",
			token: signTestHS256JWT(t, secret, nil, map[string]any{
				"iss": "https://idp.example.com",
				"aud": []string{"other-service"},
				"sub": "user-1",
				"exp": time.Now().Add(15 * time.Minute).Unix(),
			}),
		},
		{
			name: "malformed_segments",
			token: "not-a-valid.jwt",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/Resource", nil)
			req.Header.Set("Authorization", "Bearer "+tc.token)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			assertStatus(t, rr, http.StatusUnauthorized)
		})
	}
}

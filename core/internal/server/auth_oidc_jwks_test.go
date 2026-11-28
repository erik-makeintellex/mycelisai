package server

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func buildTestRSAJWKJSON(t *testing.T, kid string, pub *rsa.PublicKey) string {
	t.Helper()
	doc := jwkDocument{
		Keys: []jwkKey{
			{
				Kty: "RSA",
				Use: "sig",
				Alg: "RS256",
				Kid: kid,
				N:   base64.RawURLEncoding.EncodeToString(pub.N.Bytes()),
				E:   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pub.E)).Bytes()),
			},
		},
	}
	raw, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("marshal jwks: %v", err)
	}
	return string(raw)
}

func signTestRS256JWT(t *testing.T, priv *rsa.PrivateKey, kid string, claims map[string]any) string {
	t.Helper()
	hb, err := json.Marshal(map[string]any{"alg": "RS256", "typ": "JWT", "kid": kid})
	if err != nil {
		t.Fatalf("marshal header: %v", err)
	}
	pb, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("marshal claims: %v", err)
	}
	signingInput := base64.RawURLEncoding.EncodeToString(hb) + "." + base64.RawURLEncoding.EncodeToString(pb)
	digest := sha256.Sum256([]byte(signingInput))
	sig, err := rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA256, digest[:])
	if err != nil {
		t.Fatalf("sign rs256: %v", err)
	}
	return signingInput + "." + base64.RawURLEncoding.EncodeToString(sig)
}

func TestOIDCJWT_RS256_JWKSURL_VerifiesAndMapsRoles(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}
	jwksBody := buildTestRSAJWKJSON(t, "kid-enterprise-1", &priv.PublicKey)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(jwksBody))
	}))
	defer srv.Close()

	t.Setenv("MYCELIS_OIDC_ISSUER", "https://sso.enterprise.local")
	t.Setenv("MYCELIS_OIDC_AUDIENCE", "mycelis-core")
	t.Setenv("MYCELIS_OIDC_JWKS_URL", srv.URL)

	opToken := signTestRS256JWT(t, priv, "kid-enterprise-1", map[string]any{
		"iss":   "https://sso.enterprise.local",
		"aud":   "mycelis-core",
		"sub":   "oidc-rs256-op",
		"email": "rs256-op@enterprise.local",
		"role":  "operator",
		"exp":   time.Now().Add(10 * time.Minute).Unix(),
	})

	handler := AuthMiddleware("local-root-key", newRBACProtectedTestRouter())

	reqOp := httptest.NewRequest(http.MethodPost, "/api/v1/Resource", nil)
	reqOp.Header.Set("Authorization", "Bearer "+opToken)
	rrOp := httptest.NewRecorder()
	handler.ServeHTTP(rrOp, reqOp)
	assertStatus(t, rrOp, http.StatusCreated)

	// Tampered signature must fail with 401
	reqBad := httptest.NewRequest(http.MethodGet, "/api/v1/Resource", nil)
	reqBad.Header.Set("Authorization", "Bearer "+opToken+"tampered")
	rrBad := httptest.NewRecorder()
	handler.ServeHTTP(rrBad, reqBad)
	assertStatus(t, rrBad, http.StatusUnauthorized)

	// alg: none must be rejected
	noneToken := signTestHS256JWT(t, "unused", map[string]any{"alg": "none", "typ": "JWT"}, map[string]any{
		"iss": "https://sso.enterprise.local",
		"aud": "mycelis-core",
		"sub": "attacker",
		"exp": time.Now().Add(10 * time.Minute).Unix(),
	})
	reqNone := httptest.NewRequest(http.MethodGet, "/api/v1/Resource", nil)
	reqNone.Header.Set("Authorization", "Bearer "+noneToken)
	rrNone := httptest.NewRecorder()
	handler.ServeHTTP(rrNone, reqNone)
	assertStatus(t, rrNone, http.StatusUnauthorized)
}

func TestOIDCJWT_RS256_InlineJWKSJSON_AirGapped(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}
	t.Setenv("MYCELIS_OIDC_ISSUER", "https://airgap.local")
	t.Setenv("MYCELIS_OIDC_JWKS_JSON", buildTestRSAJWKJSON(t, "kid-airgap", &priv.PublicKey))

	token := signTestRS256JWT(t, priv, "kid-airgap", map[string]any{
		"iss":    "https://airgap.local",
		"sub":    "airgap-viewer",
		"groups": []string{"viewer"},
		"exp":    time.Now().Add(10 * time.Minute).Unix(),
	})

	cfg := loadIdentityConfigurationFromEnv()
	id := cfg.identityForToken(token)
	if id == nil || id.UserID != "airgap-viewer" || id.Role != "viewer" || id.AuthSource != "oidc_jwt" {
		t.Fatalf("expected verified air-gapped RS256 viewer identity, got %+v", id)
	}
}

func TestAuthMiddleware_QueryTokenRestrictedToSSEStreams(t *testing.T) {
	handler := AuthMiddleware("test-key", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Allowed on GET SSE streams
	for _, path := range []string{"/api/v1/stream?token=test-key", "/api/v1/memory/stream?token=test-key"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		assertStatus(t, rr, http.StatusOK)
	}

	// Rejected on normal REST and mutating routes
	for _, tc := range []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/v1/user/me?token=test-key"},
		{http.MethodPost, "/api/v1/memory/promote?token=test-key"},
		{http.MethodPost, "/api/v1/stream?token=test-key"},
	} {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		assertStatus(t, rr, http.StatusUnauthorized)
	}
}

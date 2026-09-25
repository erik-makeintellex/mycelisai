package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

const hardeningAPIKey = "test-api-key-0123456789abcdef0123456789"
const hardeningBreakGlassKey = "test-break-glass-0123456789abcdef012345"
const hardeningForwardSecret = "test-forward-secret-0123456789abcdef0123"

func clearHardeningAuthEnv(t *testing.T) {
	t.Helper()
	t.Setenv("MYCELIS_IDENTITY_MODE", "")
	t.Setenv("MYCELIS_BREAK_GLASS_API_KEY", "")
	t.Setenv("MYCELIS_WEB_SESSION_SECRET", "")
	t.Setenv("MYCELIS_WEB_IDENTITY_FORWARD_SECRET", "")
}

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
}

func forwardedRequest(t *testing.T, path, role, secret string) *http.Request {
	t.Helper()
	payload := encodeForwardedWebIdentityForTest(t, forwardedWebIdentityPayload{
		Sub: "local-user", Email: "user@example.test", Role: role, Provider: "local", IAT: time.Now().Unix(),
	})
	req, _ := http.NewRequest("GET", path, nil)
	req.Header.Set("Authorization", "Bearer "+hardeningAPIKey)
	req.Header.Set(forwardedWebIdentityHeader, payload)
	req.Header.Set(forwardedWebIdentitySignatureHeader, signForwardedWebIdentity(payload, secret))
	return req
}

func serve(handler http.Handler, req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

func TestAuthMiddleware_RejectsQueryToken(t *testing.T) {
	clearHardeningAuthEnv(t)
	handler := AuthMiddleware(hardeningAPIKey, okHandler())

	req, _ := http.NewRequest("GET", "/api/v1/stream?token="+hardeningAPIKey, nil)
	assertStatus(t, serve(handler, req), http.StatusUnauthorized)

	req, _ = http.NewRequest("GET", "/api/v1/stream", nil)
	req.Header.Set("Authorization", "Bearer "+hardeningAPIKey)
	assertStatus(t, serve(handler, req), http.StatusOK)
}

func TestAuthMiddleware_ForwardSecretDoesNotFallBackToSessionSecret(t *testing.T) {
	clearHardeningAuthEnv(t)
	t.Setenv("MYCELIS_WEB_SESSION_SECRET", hardeningForwardSecret)
	handler := AuthMiddleware(hardeningAPIKey, okHandler())

	rr := serve(handler, forwardedRequest(t, "/api/v1/user/me", "admin", hardeningForwardSecret))
	assertStatus(t, rr, http.StatusUnauthorized)
}

func TestAuthMiddleware_ForwardSecretUnsetRejectsForwardedHeaders(t *testing.T) {
	clearHardeningAuthEnv(t)
	handler := AuthMiddleware(hardeningAPIKey, okHandler())

	rr := serve(handler, forwardedRequest(t, "/api/v1/user/me", "admin", ""))
	assertStatus(t, rr, http.StatusUnauthorized)
}

func TestAuthMiddleware_ForwardSecretMustBeDistinctAndLong(t *testing.T) {
	cases := map[string]struct{ forward, breakGlass string }{
		"equals api key":        {forward: hardeningAPIKey},
		"equals break-glass":    {forward: hardeningBreakGlassKey, breakGlass: hardeningBreakGlassKey},
		"shorter than 32 bytes": {forward: "short-forward-secret"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			clearHardeningAuthEnv(t)
			t.Setenv("MYCELIS_WEB_IDENTITY_FORWARD_SECRET", tc.forward)
			t.Setenv("MYCELIS_BREAK_GLASS_API_KEY", tc.breakGlass)
			if err := ValidateWebIdentityForwardSecret(hardeningAPIKey, tc.breakGlass, tc.forward); err == nil {
				t.Fatal("expected validation error")
			}
			handler := AuthMiddleware(hardeningAPIKey, okHandler())
			req, _ := http.NewRequest("GET", "/api/v1/user/me", nil)
			req.Header.Set("Authorization", "Bearer "+hardeningAPIKey)
			rr := serve(handler, req)
			assertStatus(t, rr, http.StatusServiceUnavailable)
			body := rr.Body.String()
			if !strings.Contains(body, "MYCELIS_WEB_IDENTITY_FORWARD_SECRET") {
				t.Fatalf("expected named config error, got %s", body)
			}
			if strings.Contains(body, tc.forward) {
				t.Fatal("config error must not echo the secret")
			}
		})
	}
}

func TestAuthMiddleware_AcceptsDistinctForwardSecret(t *testing.T) {
	clearHardeningAuthEnv(t)
	t.Setenv("MYCELIS_WEB_IDENTITY_FORWARD_SECRET", hardeningForwardSecret)
	handler := AuthMiddleware(hardeningAPIKey, okHandler())
	assertStatus(t, serve(handler, forwardedRequest(t, "/api/v1/user/me", "admin", hardeningForwardSecret)), http.StatusOK)
}

func TestAuditLogRequiresAuditReadScope(t *testing.T) {
	clearHardeningAuthEnv(t)
	t.Setenv("MYCELIS_WEB_IDENTITY_FORWARD_SECRET", hardeningForwardSecret)
	expectAudit := func(t *testing.T) http.Handler {
		dbOpt, mock := withDB(t)
		s := newTestServer(dbOpt)
		mock.ExpectQuery("SELECT (.+) FROM log_entries").WithArgs(20).
			WillReturnRows(sqlmock.NewRows([]string{"id", "intent", "source", "message", "timestamp", "context"}))
		return AuthMiddleware(hardeningAPIKey, http.HandlerFunc(s.handleListAuditLog))
	}

	t.Run("standard web user is forbidden", func(t *testing.T) {
		dbOpt, _ := withDB(t)
		s := newTestServer(dbOpt)
		handler := AuthMiddleware(hardeningAPIKey, http.HandlerFunc(s.handleListAuditLog))
		rr := serve(handler, forwardedRequest(t, "/api/v1/audit", "standard", hardeningForwardSecret))
		assertStatus(t, rr, http.StatusForbidden)
	})
	t.Run("admin web session is allowed", func(t *testing.T) {
		rr := serve(expectAudit(t), forwardedRequest(t, "/api/v1/audit", "admin", hardeningForwardSecret))
		assertStatus(t, rr, http.StatusOK)
	})
	t.Run("api key is allowed", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/audit", nil)
		req.Header.Set("Authorization", "Bearer "+hardeningAPIKey)
		assertStatus(t, serve(expectAudit(t), req), http.StatusOK)
	})
	t.Run("anonymous is unauthorized", func(t *testing.T) {
		dbOpt, _ := withDB(t)
		s := newTestServer(dbOpt)
		req, _ := http.NewRequest("GET", "/api/v1/audit", nil)
		assertStatus(t, serve(AuthMiddleware(hardeningAPIKey, http.HandlerFunc(s.handleListAuditLog)), req), http.StatusUnauthorized)
		assertStatus(t, doRequest(t, http.HandlerFunc(s.handleListAuditLog), "GET", "/api/v1/audit", ""), http.StatusUnauthorized)
	})
}

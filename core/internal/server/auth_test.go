package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAuthMiddleware_MissingToken(t *testing.T) {
	handler := AuthMiddleware("test-key", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req, _ := http.NewRequest("GET", "/api/v1/user/me", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assertStatus(t, rr, http.StatusUnauthorized)
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	handler := AuthMiddleware("test-key", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req, _ := http.NewRequest("GET", "/api/v1/user/me", nil)
	req.Header.Set("Authorization", "Bearer wrong-key")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assertStatus(t, rr, http.StatusUnauthorized)
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	handler := AuthMiddleware("test-key", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity := IdentityFromContext(r.Context())
		if identity == nil {
			t.Error("Expected identity in context")
		}
		if identity.Username != "admin" {
			t.Errorf("Expected username 'admin', got %q", identity.Username)
		}
		if identity.PrincipalType != "local_admin" {
			t.Errorf("Expected principal_type local_admin, got %q", identity.PrincipalType)
		}
		if identity.AuthSource != "local_api_key" {
			t.Errorf("Expected auth_source local_api_key, got %q", identity.AuthSource)
		}
		if identity.EffectiveRole != "owner" {
			t.Errorf("Expected effective_role owner, got %q", identity.EffectiveRole)
		}
		if identity.BreakGlass {
			t.Error("Expected break_glass false for local_only mode")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req, _ := http.NewRequest("GET", "/api/v1/user/me", nil)
	req.Header.Set("Authorization", "Bearer test-key")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assertStatus(t, rr, http.StatusOK)
}

func TestAuthMiddleware_HybridModeUsesBreakGlassAdminIdentity(t *testing.T) {
	t.Setenv("MYCELIS_BREAK_GLASS_API_KEY", "break-glass-key")
	handler := AuthMiddleware("test-key", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity := IdentityFromContext(r.Context())
		if identity == nil {
			t.Error("Expected identity in context")
		}
		if identity.PrincipalType != "break_glass_admin" {
			t.Errorf("Expected principal_type break_glass_admin, got %q", identity.PrincipalType)
		}
		if identity.AuthSource != "local_break_glass" {
			t.Errorf("Expected auth_source local_break_glass, got %q", identity.AuthSource)
		}
		if !identity.BreakGlass {
			t.Error("Expected break_glass true for hybrid mode")
		}
		if identity.EffectiveRole != "owner" {
			t.Errorf("Expected effective_role owner, got %q", identity.EffectiveRole)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req, _ := http.NewRequest("GET", "/api/v1/user/me", nil)
	req.Header.Set("Authorization", "Bearer break-glass-key")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assertStatus(t, rr, http.StatusOK)
}

func TestAuthMiddleware_CustomLocalAdminIdentityFromEnvironment(t *testing.T) {
	t.Setenv("MYCELIS_LOCAL_ADMIN_USERNAME", "owner-erik")
	t.Setenv("MYCELIS_LOCAL_ADMIN_USER_ID", "user-123")
	handler := AuthMiddleware("test-key", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity := IdentityFromContext(r.Context())
		if identity == nil {
			t.Error("Expected identity in context")
		}
		if identity.Username != "owner-erik" {
			t.Errorf("Expected custom username owner-erik, got %q", identity.Username)
		}
		if identity.UserID != "user-123" {
			t.Errorf("Expected custom user ID user-123, got %q", identity.UserID)
		}
		if identity.PrincipalType != "local_admin" {
			t.Errorf("Expected local_admin principal type, got %q", identity.PrincipalType)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req, _ := http.NewRequest("GET", "/api/v1/user/me", nil)
	req.Header.Set("Authorization", "Bearer test-key")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assertStatus(t, rr, http.StatusOK)
}

func TestAuthMiddleware_UsesSignedForwardedWebIdentity(t *testing.T) {
	t.Setenv("MYCELIS_WEB_SESSION_SECRET", "forward-secret")
	handler := AuthMiddleware("test-key", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity := IdentityFromContext(r.Context())
		if identity == nil {
			t.Fatal("Expected identity in context")
		}
		if identity.Username != "erik@mycelis.link" {
			t.Fatalf("Expected forwarded web username, got %q", identity.Username)
		}
		if identity.AuthSource != "web_google" {
			t.Fatalf("Expected web_google auth source, got %q", identity.AuthSource)
		}
		if identity.PrincipalType != "google_workspace_user" {
			t.Fatalf("Expected google workspace principal, got %q", identity.PrincipalType)
		}
		if identity.EffectiveRole != "owner" {
			t.Fatalf("Expected owner effective role, got %q", identity.EffectiveRole)
		}
		w.WriteHeader(http.StatusOK)
	}))

	payload := encodeForwardedWebIdentityForTest(t, forwardedWebIdentityPayload{
		Sub:      "google-123",
		Email:    "erik@mycelis.link",
		Name:     "Erik",
		Role:     "admin",
		Provider: "google",
		HD:       "mycelis.link",
		IAT:      time.Now().Unix(),
	})
	req, _ := http.NewRequest("GET", "/api/v1/user/me", nil)
	req.Header.Set("Authorization", "Bearer test-key")
	req.Header.Set(forwardedWebIdentityHeader, payload)
	req.Header.Set(forwardedWebIdentitySignatureHeader, signForwardedWebIdentity(payload, "forward-secret"))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assertStatus(t, rr, http.StatusOK)
}

func TestAuthMiddleware_RejectsStaleForwardedWebIdentity(t *testing.T) {
	t.Setenv("MYCELIS_WEB_SESSION_SECRET", "forward-secret")
	handler := AuthMiddleware("test-key", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not receive stale forwarded identity")
	}))

	payload := encodeForwardedWebIdentityForTest(t, forwardedWebIdentityPayload{
		Sub:      "google-123",
		Email:    "erik@mycelis.link",
		Role:     "admin",
		Provider: "google",
		IAT:      time.Now().Add(-30 * time.Minute).Unix(),
	})
	req, _ := http.NewRequest("GET", "/api/v1/user/me", nil)
	req.Header.Set("Authorization", "Bearer test-key")
	req.Header.Set(forwardedWebIdentityHeader, payload)
	req.Header.Set(forwardedWebIdentitySignatureHeader, signForwardedWebIdentity(payload, "forward-secret"))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assertStatus(t, rr, http.StatusUnauthorized)
}

func TestAuthMiddleware_RejectsInvalidForwardedWebIdentity(t *testing.T) {
	t.Setenv("MYCELIS_WEB_SESSION_SECRET", "forward-secret")
	handler := AuthMiddleware("test-key", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not receive invalid forwarded identity")
	}))

	payload := encodeForwardedWebIdentityForTest(t, forwardedWebIdentityPayload{
		Sub:      "google-123",
		Email:    "erik@mycelis.link",
		Role:     "admin",
		Provider: "google",
	})
	req, _ := http.NewRequest("GET", "/api/v1/user/me", nil)
	req.Header.Set("Authorization", "Bearer test-key")
	req.Header.Set(forwardedWebIdentityHeader, payload)
	req.Header.Set(forwardedWebIdentitySignatureHeader, "bad-signature")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assertStatus(t, rr, http.StatusUnauthorized)
}

func TestActorIdentitySnapshotFromRequest(t *testing.T) {
	req, _ := http.NewRequest("POST", "/api/v1/chat", nil)
	req = req.WithContext(context.WithValue(req.Context(), ctxKeyIdentity, &RequestIdentity{
		UserID:        "google-123",
		Username:      "erik@mycelis.link",
		Role:          "admin",
		EffectiveRole: "owner",
		PrincipalType: "google_workspace_user",
		AuthSource:    "web_google",
	}))

	snapshot := actorIdentitySnapshotFromRequest(req)

	if snapshot["user_label"] != "erik@mycelis.link" {
		t.Fatalf("expected user label from forwarded identity, got %+v", snapshot)
	}
	if snapshot["auth_source"] != "web_google" {
		t.Fatalf("expected auth source in actor snapshot, got %+v", snapshot)
	}
}

func TestAuthMiddleware_QueryTokenFallback(t *testing.T) {
	handler := AuthMiddleware("test-key", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity := IdentityFromContext(r.Context())
		if identity == nil {
			t.Error("Expected identity in context via query param")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req, _ := http.NewRequest("GET", "/api/v1/stream?token=test-key", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assertStatus(t, rr, http.StatusOK)
}

func encodeForwardedWebIdentityForTest(t *testing.T, payload forwardedWebIdentityPayload) string {
	t.Helper()
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal forwarded web identity: %v", err)
	}
	return base64.RawURLEncoding.EncodeToString(raw)
}

func TestAuthMiddleware_FailsClosedWhenDeploymentContractRequiresBreakGlass(t *testing.T) {
	t.Setenv("MYCELIS_IDENTITY_MODE", "federated")
	t.Setenv("MYCELIS_BREAK_GLASS_API_KEY", "")

	handler := AuthMiddleware("test-key", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req, _ := http.NewRequest("GET", "/api/v1/user/me", nil)
	req.Header.Set("Authorization", "Bearer test-key")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

func TestAuthMiddleware_HealthzExempt(t *testing.T) {
	handler := AuthMiddleware("test-key", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req, _ := http.NewRequest("GET", "/healthz", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assertStatus(t, rr, http.StatusOK)
}

func TestAuthMiddleware_OptionsExempt(t *testing.T) {
	handler := AuthMiddleware("test-key", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req, _ := http.NewRequest("OPTIONS", "/api/v1/anything", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assertStatus(t, rr, http.StatusOK)
}

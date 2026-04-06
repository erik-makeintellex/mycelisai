package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
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
	t.Setenv("MYCELIS_IDENTITY_MODE", "hybrid")
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
	req.Header.Set("Authorization", "Bearer test-key")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assertStatus(t, rr, http.StatusOK)
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

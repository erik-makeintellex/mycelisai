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

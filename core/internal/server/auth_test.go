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

func TestAuthMiddleware_OperatorAndViewerTokens(t *testing.T) {
	t.Setenv("MYCELIS_OPERATOR_API_KEY", "operator-key")
	t.Setenv("MYCELIS_VIEWER_API_KEY", "viewer-key")

	var captured *RequestIdentity
	handler := AuthMiddleware("test-key", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = IdentityFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	reqOp, _ := http.NewRequest("GET", "/api/v1/user/me", nil)
	reqOp.Header.Set("Authorization", "Bearer operator-key")
	rrOp := httptest.NewRecorder()
	handler.ServeHTTP(rrOp, reqOp)
	assertStatus(t, rrOp, http.StatusOK)
	if captured == nil || captured.Role != "operator" || captured.EffectiveRole != "operator" || captured.PrincipalType != "local_operator" || captured.UserID != "00000000-0000-0000-0000-000000000002" {
		t.Fatalf("unexpected operator identity: %+v", captured)
	}
	if !hasScope(captured, "teams:write") || !hasScope(captured, "settings:write") {
		t.Fatalf("expected operator scopes to allow teams:write and settings:write, got %v", captured.Scopes)
	}

	reqView, _ := http.NewRequest("GET", "/api/v1/user/me", nil)
	reqView.Header.Set("Authorization", "Bearer viewer-key")
	rrView := httptest.NewRecorder()
	handler.ServeHTTP(rrView, reqView)
	assertStatus(t, rrView, http.StatusOK)
	if captured == nil || captured.Role != "viewer" || captured.EffectiveRole != "viewer" || captured.PrincipalType != "local_viewer" || captured.UserID != "00000000-0000-0000-0000-000000000003" {
		t.Fatalf("unexpected viewer identity: %+v", captured)
	}
	if !hasScope(captured, "teams:read") || hasScope(captured, "teams:write") || hasScope(captured, "settings:write") {
		t.Fatalf("expected viewer read-only scopes, got %v", captured.Scopes)
	}
}

func TestHasMinimumRole_And_DefaultScopesForRole(t *testing.T) {
	viewer := &RequestIdentity{Role: "viewer", EffectiveRole: "viewer", Scopes: DefaultScopesForRole("viewer")}
	operator := &RequestIdentity{Role: "operator", EffectiveRole: "operator", Scopes: DefaultScopesForRole("operator")}
	owner := &RequestIdentity{Role: "admin", EffectiveRole: "owner", Scopes: DefaultScopesForRole("owner")}

	if !HasMinimumRole(viewer, "viewer") || HasMinimumRole(viewer, "operator") || HasMinimumRole(viewer, "admin") {
		t.Fatal("viewer role rank mismatch")
	}
	if !HasMinimumRole(operator, "viewer") || !HasMinimumRole(operator, "operator") || HasMinimumRole(operator, "admin") {
		t.Fatal("operator role rank mismatch")
	}
	if !HasMinimumRole(owner, "viewer") || !HasMinimumRole(owner, "operator") || !HasMinimumRole(owner, "admin") || !HasMinimumRole(owner, "owner") {
		t.Fatal("owner/admin role rank mismatch")
	}
	if HasMinimumRole(nil, "viewer") || HasMinimumRole(owner, "unknown") {
		t.Fatal("expected false for nil identity or unknown role")
	}
}

func TestRequireMinimumRole(t *testing.T) {
	check := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := RequireMinimumRole(w, r, "operator", "teams:write"); !ok {
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	assertStatus(t, doRequest(t, check, "POST", "/api/v1/teams", ""), http.StatusUnauthorized)

	viewer := &RequestIdentity{UserID: "v1", Role: "viewer", EffectiveRole: "viewer", Scopes: DefaultScopesForRole("viewer")}
	assertStatus(t, doAuthenticatedRequestAs(t, check, "POST", "/api/v1/teams", "", viewer), http.StatusForbidden)

	scopedOutOperator := &RequestIdentity{UserID: "o1", Role: "operator", EffectiveRole: "operator", Scopes: []string{"teams:read"}}
	assertStatus(t, doAuthenticatedRequestAs(t, check, "POST", "/api/v1/teams", "", scopedOutOperator), http.StatusForbidden)

	operator := &RequestIdentity{UserID: "o2", Role: "operator", EffectiveRole: "operator", Scopes: DefaultScopesForRole("operator")}
	assertStatus(t, doAuthenticatedRequestAs(t, check, "POST", "/api/v1/teams", "", operator), http.StatusOK)
}

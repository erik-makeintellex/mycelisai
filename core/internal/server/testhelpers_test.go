package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/governance"
	"github.com/mycelis/core/internal/registry"
	pb "github.com/mycelis/core/pkg/pb/swarm"
)

// ── Shared Test Helpers ────────────────────────────────────────────

// newTestServer creates a minimal AdminServer for handler testing.
// Pass option functions to wire only the subsystems each test needs.
func newTestServer(opts ...func(*AdminServer)) *AdminServer {
	s := &AdminServer{}
	for _, o := range opts {
		o(s)
	}
	return s
}

// withDB creates a sqlmock database and wires it through Registry.
// Returns the option function and the mock for setting expectations.
func withDB(t *testing.T) (func(*AdminServer), sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return func(s *AdminServer) {
		s.Registry = &registry.Service{DB: db}
	}, mock
}

// withGuard creates a governance Guard with the given policy config.
func withGuard(cfg *governance.PolicyConfig) func(*AdminServer) {
	return func(s *AdminServer) {
		s.Guard = &governance.Guard{
			Engine:        &governance.Engine{Config: cfg},
			PendingBuffer: make(map[string]*pb.ApprovalRequest),
		}
	}
}

// defaultTestPolicyConfig returns a minimal valid policy config for testing.
func defaultTestPolicyConfig() *governance.PolicyConfig {
	return &governance.PolicyConfig{
		Groups: []governance.PolicyGroup{
			{
				Name:    "test-group",
				Targets: []string{"*"},
				Rules: []governance.PolicyRule{
					{Intent: "*", Action: governance.ActionAllow},
				},
			},
		},
		Defaults: governance.DefaultConfig{DefaultAction: governance.ActionAllow},
	}
}

// setupMux creates an http.ServeMux with a single route registered.
// Required for handlers that use r.PathValue() (Go 1.22+ routing).
func setupMux(t *testing.T, pattern string, handler http.HandlerFunc) *http.ServeMux {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc(pattern, handler)
	return mux
}

// doRequest sends an HTTP request to a handler and returns the recorded response.
func doRequest(t *testing.T, handler http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	var req *http.Request
	if body != "" {
		req, _ = http.NewRequest(method, path, strings.NewReader(body))
	} else {
		req, _ = http.NewRequest(method, path, nil)
	}
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

// doAuthenticatedRequest sends an HTTP request with a test identity injected into context.
func doAuthenticatedRequest(t *testing.T, handler http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	var req *http.Request
	if body != "" {
		req, _ = http.NewRequest(method, path, strings.NewReader(body))
	} else {
		req, _ = http.NewRequest(method, path, nil)
	}
	identity := &RequestIdentity{UserID: "test-user-001", Username: "admin", Role: "admin"}
	ctx := context.WithValue(req.Context(), ctxKeyIdentity, identity)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

// assertStatus verifies the HTTP status code in the response.
func assertStatus(t *testing.T, rr *httptest.ResponseRecorder, want int) {
	t.Helper()
	if rr.Code != want {
		t.Errorf("Expected status %d, got %d. Body: %s", want, rr.Code, rr.Body.String())
	}
}

// assertJSON decodes the response body into the target struct.
func assertJSON(t *testing.T, rr *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.NewDecoder(rr.Body).Decode(target); err != nil {
		t.Fatalf("Failed to decode JSON response: %v. Body: %s", err, rr.Body.String())
	}
}

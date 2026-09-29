package server

import (
	"net/http"
	"testing"
)

// AUTH-C1 A6: the V7 provisioning draft route is gone from the route table.
func TestAuthC1ProvisionDraftRouteRemoved(t *testing.T) {
	s := newTestServer()
	mux := http.NewServeMux()
	s.RegisterRoutes(mux)
	for _, method := range []string{http.MethodPost, http.MethodGet} {
		rr := doAuthenticatedRequestAs(t, mux, method, "/api/v1/provision/draft", `{"intent":"build a team"}`, adminWithScopes("*"))
		assertStatus(t, rr, http.StatusNotFound)
	}
}

// CONS-C3 A4: the V7 Overseer is removed, and with it the trust-threshold
// valve. No caller, admin or not, reaches a handler on that path.
func TestConsC3TrustThresholdRouteRemoved(t *testing.T) {
	s := newTestServer()
	mux := http.NewServeMux()
	s.RegisterRoutes(mux)
	for _, method := range []string{http.MethodGet, http.MethodPut} {
		for name, who := range map[string]*RequestIdentity{
			"root admin wildcard": adminWithScopes("*"),
			"governance admin":    adminWithScopes(scopeGovernanceWrite),
			"signed-in user":      standardUserIdentity(),
		} {
			rr := doAuthenticatedRequestAs(t, mux, method, "/api/v1/trust/threshold", `{"threshold":0.1}`, who)
			if rr.Code != http.StatusNotFound {
				t.Errorf("%s %s -> %d, want 404", method, name, rr.Code)
			}
		}
	}
}

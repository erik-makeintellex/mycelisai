package server

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"testing"
)

func organizationWriteRoutes(s *AdminServer) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/organizations", s.handleCreateOrganization)
	mux.HandleFunc("PATCH /api/v1/organizations/{id}/ai-engine", s.handleUpdateOrganizationAIEngine)
	mux.HandleFunc("PATCH /api/v1/organizations/{id}/output-model-routing", s.handleUpdateOrganizationOutputModelRouting)
	mux.HandleFunc("PATCH /api/v1/organizations/{id}/response-contract", s.handleUpdateResponseContract)
	mux.HandleFunc("PATCH /api/v1/organizations/{id}/departments/{departmentId}/ai-engine", s.handleUpdateDepartmentAIEngine)
	mux.HandleFunc("PATCH /api/v1/organizations/{id}/departments/{departmentId}/agent-types/{agentTypeId}/ai-engine", s.handleUpdateAgentTypeAIEngine)
	mux.HandleFunc("PATCH /api/v1/organizations/{id}/departments/{departmentId}/agent-types/{agentTypeId}/response-contract", s.handleUpdateAgentTypeResponseContract)
	return mux
}

func TestOrganizationWrites_DenyAnonymousStandardAndUnscopedAdmin(t *testing.T) {
	s := newTestServer(withTemplateBundlesPath(writeStarterBundle(t)))
	seeded := seedOrganization(t, s, testOrganizationHome())
	before := mustGetOrganization(t, s, seeded.ID)
	mux := organizationWriteRoutes(s)
	base := "/api/v1/organizations/" + seeded.ID
	requests := []struct{ method, path, body string }{
		{http.MethodPost, "/api/v1/organizations", `{"name":"Forged","purpose":"x","start_mode":"empty"}`},
		{http.MethodPatch, base + "/ai-engine", `{"profile_id":"high_reasoning"}`},
		{http.MethodPatch, base + "/output-model-routing", `{"routing_mode":"single_model","default_model_id":"qwen3:8b"}`},
		{http.MethodPatch, base + "/response-contract", `{"profile_id":"warm_supportive"}`},
		{http.MethodPatch, base + "/departments/platform/ai-engine", `{"profile_id":"high_reasoning"}`},
		{http.MethodPatch, base + "/departments/platform/agent-types/planner/ai-engine", `{"profile_id":"high_reasoning"}`},
		{http.MethodPatch, base + "/departments/platform/agent-types/planner/response-contract", `{"profile_id":"warm_supportive"}`},
	}
	standard := &RequestIdentity{UserID: "u-1", Username: "member", Role: "user", Scopes: []string{"organizations:write"}}
	unscopedAdmin := &RequestIdentity{UserID: "u-2", Username: "ops", Role: "admin", Scopes: []string{"organizations:read", "groups:*"}}
	for _, req := range requests {
		if rr := doRequest(t, mux, req.method, req.path, req.body); rr.Code != http.StatusUnauthorized {
			t.Errorf("anonymous %s %s = %d, want 401", req.method, req.path, rr.Code)
		}
		if rr := doAuthenticatedRequestAs(t, mux, req.method, req.path, req.body, standard); rr.Code != http.StatusForbidden {
			t.Errorf("standard user %s %s = %d, want 403", req.method, req.path, rr.Code)
		}
		if rr := doAuthenticatedRequestAs(t, mux, req.method, req.path, req.body, unscopedAdmin); rr.Code != http.StatusForbidden {
			t.Errorf("admin without organizations:write %s %s = %d, want 403", req.method, req.path, rr.Code)
		}
	}
	summaries, err := s.organizationStore().List(context.Background())
	if err != nil || len(summaries) != 1 {
		t.Fatalf("denied writes changed rows: %+v err=%v", summaries, err)
	}
	if after := mustGetOrganization(t, s, seeded.ID); fmt.Sprintf("%+v", after) != fmt.Sprintf("%+v", before) {
		t.Fatalf("denied update mutated organization:\nbefore %+v\nafter  %+v", before, after)
	}
}

func TestOrganizationWrites_ScopedRootAdminAllowedAndForgedFieldsIgnored(t *testing.T) {
	s := newTestServer(withTemplateBundlesPath(writeStarterBundle(t)))
	admin := &RequestIdentity{UserID: "u-3", Username: "root", Role: "admin", Scopes: []string{"organizations:write"}}
	rr := doAuthenticatedRequestAs(t, http.HandlerFunc(s.handleCreateOrganization), http.MethodPost, "/api/v1/organizations",
		`{"name":"Atlas","purpose":"Persist me","start_mode":"empty","tenant_id":"evil","qa_fixture_scope_id":"forged-scope"}`, admin)
	assertStatus(t, rr, http.StatusCreated)
	var created struct {
		Data OrganizationHomePayload `json:"data"`
	}
	assertJSON(t, rr, &created)
	if scope, ok, err := s.organizationStore().QAFixtureScope(context.Background(), created.Data.ID); err != nil || ok || scope != "" {
		t.Fatalf("forged fixture scope accepted: %q %v %v", scope, ok, err)
	}
}

func TestOrganizationStore_ConcurrentUpdatesLoseNoWrite(t *testing.T) {
	s := newTestServer()
	seedOrganization(t, s, testOrganizationHome())
	runConcurrentOrganizationIncrements(t, s.organizationStore(), testOrgID, 25)
}

// runConcurrentOrganizationIncrements increments a counter kept in the
// document from n goroutines; any lost update leaves the count short.
func runConcurrentOrganizationIncrements(t *testing.T, store *OrganizationStore, id string, n int) {
	t.Helper()
	var wg sync.WaitGroup
	errs := make(chan error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := store.Update(context.Background(), id, func(home OrganizationHomePayload) OrganizationHomePayload {
				home.AdvisorCount++
				return home
			})
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	home, err := store.Get(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	if home.AdvisorCount != n {
		t.Fatalf("lost update: advisor_count = %d, want %d", home.AdvisorCount, n)
	}
}

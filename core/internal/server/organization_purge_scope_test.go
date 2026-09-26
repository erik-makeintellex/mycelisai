package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// N3: a PATCH addressing a missing department or agent type aborts under the
// row lock and rolls back; no UPDATE is issued before the 404.
func TestOrganizationProfilePatch_MissingTargetIs404WithoutWrite(t *testing.T) {
	home := testOrganizationHome()
	home.Departments = []OrganizationDepartmentSummary{{ID: "platform", Name: "Platform"}}
	document, _ := json.Marshal(home)
	cases := []struct{ pattern, path, body, want string }{
		{"PATCH /api/v1/organizations/{id}/departments/{departmentId}/ai-engine",
			"/departments/missing/ai-engine", `{"profile_id":"high_reasoning"}`, "department not found"},
		{"PATCH /api/v1/organizations/{id}/departments/{departmentId}/agent-types/{agentTypeId}/ai-engine",
			"/departments/platform/agent-types/missing-agent-type/ai-engine", `{"profile_id":"high_reasoning"}`, "agent type profile not found"},
		{"PATCH /api/v1/organizations/{id}/departments/{departmentId}/agent-types/{agentTypeId}/response-contract",
			"/departments/missing/agent-types/missing-agent-type/response-contract", `{"profile_id":"warm_supportive"}`, "department not found"},
	}
	for _, tc := range cases {
		store, mock := newSQLMockOrganizationStore(t)
		mock.ExpectBegin()
		mock.ExpectQuery("FOR UPDATE").WithArgs(testOrgID, organizationTenantID).
			WillReturnRows(sqlmock.NewRows([]string{"document", "qa_fixture_scope_id"}).AddRow(document, "scope-a"))
		mock.ExpectRollback()
		s := newTestServer(func(s *AdminServer) { s.Organizations = store })
		mux := organizationWriteRoutes(s)
		rr := doAuthenticatedRequest(t, mux, http.MethodPatch, "/api/v1/organizations/"+testOrgID+tc.path, tc.body)
		assertStatus(t, rr, http.StatusNotFound)
		if !strings.Contains(rr.Body.String(), tc.want) {
			t.Fatalf("%s: body %s, want %q", tc.path, rr.Body.String(), tc.want)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("%s: %v", tc.path, err)
		}
	}
}

// N1: the purge deletes an organization only inside its own fixture scope. A
// claimed organization that exists outside the scope fails the purge
// transaction as unowned and nothing is deleted.
func TestQAFixturePurge_OrganizationOutsideScopeFailsWithoutDelete(t *testing.T) {
	withDatabase, mock := withDB(t)
	s := newTestServer(withDatabase)
	scope := qaFixtureScope{ID: "11111111-1111-1111-1111-111111111111", TenantID: qaFixtureTenantID}
	resources := []qaFixtureResource{{Kind: "organization", Ref: testOrgID}}
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM organizations").WithArgs(testOrgID, qaFixtureTenantID, scope.ID).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT EXISTS").WithArgs(testOrgID, qaFixtureTenantID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectRollback()
	result := newQAFixturePurgeResult(scope, resources, true)
	err := s.deleteQAFixtureDurableResources(context.Background(), scope, resources, &result)
	if !errors.Is(err, errQAFixtureResourceUnowned) {
		t.Fatalf("out-of-scope purge err = %v, want unowned", err)
	}
	if len(result.RemovedOrganizations) != 0 || result.DeletedRows["organizations"] != 0 {
		t.Fatalf("out-of-scope purge reported removal: %+v", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestQAFixturePurge_AlreadyAbsentOrganizationIsResumableNoOp(t *testing.T) {
	withDatabase, mock := withDB(t)
	s := newTestServer(withDatabase)
	scope := qaFixtureScope{ID: "11111111-1111-1111-1111-111111111111", TenantID: qaFixtureTenantID}
	resources := []qaFixtureResource{{Kind: "organization", Ref: testOrgID}}
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM organizations").WithArgs(testOrgID, qaFixtureTenantID, scope.ID).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT EXISTS").WithArgs(testOrgID, qaFixtureTenantID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectCommit()
	result := newQAFixturePurgeResult(scope, resources, true)
	if err := s.deleteQAFixtureDurableResources(context.Background(), scope, resources, &result); err != nil {
		t.Fatalf("resumed purge err = %v", err)
	}
	if len(result.RemovedOrganizations) != 0 {
		t.Fatalf("absent organization reported removed: %+v", result.RemovedOrganizations)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

package server

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/mycelis/core/internal/identity"
)

func TestTeamOwnershipHTTPPostgres(t *testing.T) {
	dsn := os.Getenv("MYCELIS_TEAM_OWNERSHIP_TEST_DSN")
	if dsn == "" {
		t.Skip("requires isolated MYCELIS_TEAM_OWNERSHIP_TEST_DSN")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	account, user, group, role, member, team := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString(), "team-"+uuid.NewString()
	exec := func(query string, args ...any) {
		t.Helper()
		if _, err := db.ExecContext(t.Context(), query, args...); err != nil {
			t.Fatal(err)
		}
	}
	exec(`INSERT INTO accounts(id,slug,name) VALUES($1,$2,'HTTP team ownership')`, account, account)
	exec(`INSERT INTO users(id,username,account_id) VALUES($1,$2,$3)`, user, user, account)
	exec(`INSERT INTO groups(id,account_id,key,name) VALUES($1,$2,$3,'HTTP group')`, group, account, group)
	exec(`INSERT INTO runtime_team_manifests(tenant_id,team_id,manifest_digest,manifest) VALUES('default',$1,'sha256:http-fixture',jsonb_build_object('id',$1::text))`, team)
	t.Setenv("MYCELIS_LOCAL_ADMIN_USER_ID", user)
	s := &AdminServer{DB: db}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/admin/runtime-teams/{teamID}/ownership", s.HandleTeamOwnership)
	mux.HandleFunc("PUT /api/v1/admin/runtime-teams/{teamID}/ownership", s.HandleTeamOwnership)
	mux.HandleFunc("DELETE /api/v1/admin/runtime-teams/{teamID}/ownership", s.HandleTeamOwnership)
	api := httptest.NewServer(AuthMiddleware("fixture-token", mux))
	defer api.Close()
	path := "/api/v1/admin/runtime-teams/" + team + "/ownership"
	call := func(method, suffix string, body any, auth bool, want int) map[string]any {
		t.Helper()
		encoded, _ := json.Marshal(body)
		req, err := http.NewRequest(method, api.URL+path+suffix, bytes.NewReader(encoded))
		if err != nil {
			t.Fatal(err)
		}
		if auth {
			req.Header.Set("Authorization", "Bearer fixture-token")
		}
		res, err := api.Client().Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		if res.StatusCode != want {
			t.Fatalf("%s status=%d want=%d", method, res.StatusCode, want)
		}
		var envelope map[string]any
		if err := json.NewDecoder(res.Body).Decode(&envelope); err != nil {
			t.Fatal(err)
		}
		value, _ := envelope["data"].(map[string]any)
		return value
	}
	query := "?group_id=" + group
	call(http.MethodGet, query, nil, false, http.StatusUnauthorized)
	call(http.MethodGet, query, nil, true, http.StatusForbidden) // bearer admin is not a persisted grant
	exec(`INSERT INTO roles(id,key,name,scope) VALUES($1,$2,'Provisioner','system')`, role, role)
	exec(`INSERT INTO role_permissions(role_id,permission_key) VALUES($1,$2)`, role, identity.TeamProvisionPermission)
	exec(`INSERT INTO org_memberships(id,account_id,user_id,role_id) VALUES($1,$2,$3,$4)`, member, account, user, role)
	inspected := call(http.MethodGet, query, nil, true, http.StatusOK)
	if inspected["revision"] == "" || inspected["manifest_digest"] != "sha256:http-fixture" {
		t.Fatalf("inspect=%v", inspected)
	}
	request := map[string]any{"group_id": group, "expected_manifest_digest": inspected["manifest_digest"], "expected_revision": inspected["revision"], "reason": "Reviewed operator assignment"}
	request["actor_user_id"] = uuid.NewString()
	call(http.MethodPut, "", request, true, http.StatusBadRequest)
	delete(request, "actor_user_id")
	bound := call(http.MethodPut, "", request, true, http.StatusOK)
	if bound["owner_group_id"] != group || bound["ownership_provisioned_by"] != user {
		t.Fatalf("bound=%v", bound)
	}
	call(http.MethodDelete, "", request, true, http.StatusConflict) // stale revision
	request["expected_revision"] = bound["revision"]
	if revoked := call(http.MethodDelete, "", request, true, http.StatusOK); revoked["ownership_revoked_by"] != user {
		t.Fatalf("revoked=%v", revoked)
	}
	var audits int
	if err := db.QueryRowContext(t.Context(), `SELECT count(*) FROM identity_audit_events WHERE target_id=$1`, team).Scan(&audits); err != nil || audits != 2 {
		t.Fatal(fmt.Sprintf("audit count=%d err=%v", audits, err))
	}
}

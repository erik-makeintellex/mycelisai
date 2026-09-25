package server

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

const (
	cfgLeakAPIKey = "sk-config-view-literal-0123456789"
	cfgLeakToken  = "ghp_configviewliteraltoken0123456789"
	cfgLeakPass   = "hunter2-config-view"
)

// cfgAllLeaks lists every literal secret any fixture seeds.
var cfgAllLeaks = []string{depSecretKey, cfgLeakAPIKey, cfgLeakToken, cfgLeakPass}

type cfgEnvelope struct {
	OK    bool       `json:"ok"`
	Error string     `json:"error"`
	Data  ConfigView `json:"data"`
}

func cfgMux(s *AdminServer) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/config/{kind}/{id}", s.HandleGetConfigView)
	return mux
}

// cfgGet requests a config view and asserts no seeded secret appears anywhere
// in the raw response body, whatever the status.
func cfgGet(t *testing.T, s *AdminServer, path string, identity *RequestIdentity, wantStatus int) cfgEnvelope {
	t.Helper()
	if identity == nil {
		identity = localAdminIdentityForTest()
	}
	rr := doAuthenticatedRequestAs(t, cfgMux(s), "GET", path, "", identity)
	body := rr.Body.String()
	for _, leak := range cfgAllLeaks {
		if strings.Contains(body, leak) {
			t.Fatalf("%s leaked secret %q:\n%s", path, leak, body)
		}
	}
	if rr.Code != wantStatus {
		t.Fatalf("%s status = %d, want %d: %s", path, rr.Code, wantStatus, body)
	}
	var env cfgEnvelope
	if err := json.Unmarshal(rr.Body.Bytes(), &env); err != nil {
		t.Fatalf("decode %s: %v\n%s", path, err, body)
	}
	if env.OK != (wantStatus == http.StatusOK) {
		t.Fatalf("%s envelope ok=%v for status %d: %s", path, env.OK, wantStatus, body)
	}
	return env
}

func newCfgServer(t *testing.T) (*AdminServer, sqlmock.Sqlmock) {
	t.Helper()
	withDeploymentFakes(t, depRuntimeTeams(), true, depHeartbeats())
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)
	s.Cognitive = depRouter() // providers carry depSecretKey literals
	s.Organizations = depOrgStore(t)
	return s, mock
}

func TestConfigView_StandingTeamYAMLAndJSON(t *testing.T) {
	s, _ := newCfgServer(t)
	env := cfgGet(t, s, "/api/v1/config/teams/admin-core", nil, http.StatusOK)
	d := env.Data
	if d.Kind != configKindTeam || d.ID != "admin-core" || d.Format != "yaml" || d.Source != configSourceRuntime || d.GeneratedAt.IsZero() {
		t.Fatalf("view = %+v", d)
	}
	for _, want := range []string{"kind: TeamManifest", "apiVersion: mycelis/v1alpha1", "resolved:", "provider: ollama", "model: qwen2.5-coder:7b", "provider_origin: fallback"} {
		if !strings.Contains(d.Document, want) {
			t.Fatalf("document missing %q:\n%s", want, d.Document)
		}
	}
	if strings.Contains(d.Document, "auth_key") || strings.Contains(d.Document, "endpoint") {
		t.Fatalf("provider block leaked into team document:\n%s", d.Document)
	}

	jsonEnv := cfgGet(t, s, "/api/v1/config/teams/admin-core?format=json", nil, http.StatusOK)
	var doc map[string]any
	if err := json.Unmarshal([]byte(jsonEnv.Data.Document), &doc); err != nil || jsonEnv.Data.Format != "json" {
		t.Fatalf("json document invalid (%v): %s", err, jsonEnv.Data.Document)
	}
	if doc["kind"] != "TeamManifest" || jsonEnv.Data.Redactions == nil {
		t.Fatalf("json doc = %+v redactions=%v", doc, jsonEnv.Data.Redactions)
	}
}

func TestConfigView_MissionTeamFromDatabaseNeverLeaksStoredKeys(t *testing.T) {
	s, mock := newCfgServer(t)
	mock.ExpectQuery(`FROM teams t\s+JOIN missions m\s+ON m.id = t.mission_id\s+WHERE t.id = \$1`).WithArgs(depTeamB).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "role", "mission_id", "mname", "status"}).
			AddRow(depTeamB, "research", "Research", depMissionB, "Old scan", "completed"))
	mock.ExpectQuery(`FROM service_manifests WHERE team_id = \$1`).WithArgs(depTeamB).
		WillReturnRows(sqlmock.NewRows([]string{"manifest"}).AddRow([]byte(
			`{"id":"researcher-b","role":"researcher","provider":"ollama","api_key":"` + cfgLeakAPIKey +
				`","system_prompt":"Call the API with Bearer ` + cfgLeakToken + ` please"}`)))
	env := cfgGet(t, s, "/api/v1/config/teams/"+depTeamB, nil, http.StatusOK)
	if env.Data.Source != configSourceDatabase || !strings.Contains(env.Data.Document, "mission_id: "+depMissionB) {
		t.Fatalf("mission view = %+v", env.Data)
	}
	if len(env.Data.Redactions) != 1 || !strings.HasSuffix(env.Data.Redactions[0], "system_prompt") {
		t.Fatalf("redactions = %v", env.Data.Redactions)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestConfigView_LiveMissionTeamUsesRuntime(t *testing.T) {
	s, mock := newCfgServer(t)
	mock.ExpectQuery(`FROM teams t`).WithArgs(depTeamA).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "role", "mission_id", "mname", "status"}).
			AddRow(depTeamA, "research", "Research", depMissionA, "Market scan", "active"))
	env := cfgGet(t, s, "/api/v1/config/teams/"+depTeamA, nil, http.StatusOK)
	if env.Data.Source != configSourceRuntime || !strings.Contains(env.Data.Document, "runtime_id: "+depMissionA+".research") {
		t.Fatalf("live mission view = %+v", env.Data)
	}
}

func TestConfigView_GroupRequiresGroupsRead(t *testing.T) {
	s, mock := newCfgServer(t)
	path := "/api/v1/config/groups/" + depGroupID
	cfgGet(t, s, path, &RequestIdentity{UserID: "u", Role: "admin", Scopes: []string{"teams:read"}}, http.StatusForbidden)
	cfgGet(t, s, path, &RequestIdentity{UserID: "u", Role: "operator", Scopes: []string{"groups:read"}}, http.StatusForbidden)

	mock.ExpectQuery(`FROM collaboration_groups`).WithArgs(depGroupID).
		WillReturnRows(sqlmock.NewRows(depGroupColumns()).AddRow(depGroupRow(depGroupID, "active", nil)...))
	env := cfgGet(t, s, path, &RequestIdentity{UserID: "u", Role: "admin", Scopes: []string{"groups:read"}}, http.StatusOK)
	for _, want := range []string{"kind: CollaborationGroup", "work_mode: execute_with_approval", "approval_policy_ref: delivery-approve-changes"} {
		if !strings.Contains(env.Data.Document, want) {
			t.Fatalf("group document missing %q:\n%s", want, env.Data.Document)
		}
	}
	if env.Data.Source != configSourceDatabase {
		t.Fatalf("source = %s", env.Data.Source)
	}
}

func TestConfigView_Organization(t *testing.T) {
	s, _ := newCfgServer(t)
	env := cfgGet(t, s, "/api/v1/config/organizations/org-northstar", nil, http.StatusOK)
	if env.Data.Source != configSourceRuntime || !strings.Contains(env.Data.Document, "kind: Organization") ||
		!strings.Contains(env.Data.Document, "Product Delivery") {
		t.Fatalf("organization view = %+v", env.Data)
	}
	cfgGet(t, s, "/api/v1/config/organizations/org-missing", nil, http.StatusNotFound)
}

func writeCfgTemplates(t *testing.T) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), "templates")
	if err := os.MkdirAll(filepath.Join(root, "starters"), 0o755); err != nil {
		t.Fatal(err)
	}
	top := "id: top-bundle\nname: Top\nteams: []\n"
	starter := "id: starter-secretive\nname: Secretive\nprovider_policy:\n  openai:\n    api_key: " + cfgLeakAPIKey +
		"\n    api_key_env: OPENAI_API_KEY\nmcp:\n  env:\n    GITHUB_TOKEN: " + cfgLeakToken +
		"\n  dsn: postgres://svc:" + cfgLeakPass + "@db:5432/app\n  max_tokens: 4096\n"
	for name, body := range map[string]string{"top.yaml": top, "starters/secretive.yaml": starter, "starters/broken.yaml": ":\n\t- nope"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func TestConfigView_TemplateRedactsLiteralKeys(t *testing.T) {
	s, _ := newCfgServer(t)
	s.TemplateBundlesPath = writeCfgTemplates(t)
	for _, format := range []string{"yaml", "json"} {
		env := cfgGet(t, s, "/api/v1/config/templates/starter-secretive?format="+format, nil, http.StatusOK)
		d := env.Data
		if d.Source != configSourceFile || d.SourcePath != "templates/starters/secretive.yaml" {
			t.Fatalf("template source = %s %s", d.Source, d.SourcePath)
		}
		want := []string{"mcp.dsn", "mcp.env.GITHUB_TOKEN", "provider_policy.openai.api_key"}
		if strings.Join(d.Redactions, ",") != strings.Join(want, ",") {
			t.Fatalf("redactions = %v, want %v", d.Redactions, want)
		}
		for _, visible := range []string{"OPENAI_API_KEY", "${env:OPENAI_API_KEY}", "GITHUB_TOKEN", "4096", "postgres://svc:"} {
			if !strings.Contains(d.Document, visible) {
				t.Fatalf("%s document hid %q:\n%s", format, visible, d.Document)
			}
		}
	}
	env := cfgGet(t, s, "/api/v1/config/templates/top-bundle", nil, http.StatusOK)
	if env.Data.SourcePath != "templates/top.yaml" || len(env.Data.Redactions) != 0 {
		t.Fatalf("top bundle view = %+v", env.Data)
	}
	cfgGet(t, s, "/api/v1/config/templates/..%2Fcognitive", nil, http.StatusNotFound)
}

func TestConfigView_RealTemplateBundlesRender(t *testing.T) {
	s, _ := newCfgServer(t)
	s.TemplateBundlesPath = filepath.Join("..", "..", "config", "templates")
	ids := []string{"v8-migration-standing-team-bridge", "starter-client-delivery-studio", "starter-governed-enterprise-program",
		"starter-operations-desk", "starter-product-delivery-team"}
	for _, id := range ids {
		for _, format := range []string{"yaml", "json"} {
			env := cfgGet(t, s, "/api/v1/config/templates/"+id+"?format="+format, nil, http.StatusOK)
			if !strings.Contains(env.Data.Document, id) {
				t.Fatalf("%s document missing its id", id)
			}
		}
	}
}

func TestConfigView_Errors(t *testing.T) {
	s, _ := newCfgServer(t)
	cfgGet(t, s, "/api/v1/config/widgets/x", nil, http.StatusBadRequest)
	cfgGet(t, s, "/api/v1/config/teams/admin-core?format=xml", nil, http.StatusBadRequest)
	cfgGet(t, s, "/api/v1/config/teams/no-such-team", nil, http.StatusNotFound)
	cfgGet(t, s, "/api/v1/config/groups/not-a-uuid", nil, http.StatusNotFound)

	rr := doRequest(t, cfgMux(s), "GET", "/api/v1/config/teams/admin-core", "")
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated status = %d", rr.Code)
	}

	withDeploymentFakes(t, nil, false, nil)
	cfgGet(t, s, "/api/v1/config/teams/admin-core", nil, http.StatusServiceUnavailable)
	s.DB = nil
	cfgGet(t, s, "/api/v1/config/teams/"+depTeamA, nil, http.StatusServiceUnavailable)
	cfgGet(t, s, "/api/v1/config/groups/"+depGroupID, nil, http.StatusServiceUnavailable)
}

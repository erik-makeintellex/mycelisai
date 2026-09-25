package server

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

func newDepServer(t *testing.T) (*AdminServer, func()) {
	t.Helper()
	withDeploymentFakes(t, depRuntimeTeams(), true, depHeartbeats())
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)
	s.Cognitive = depRouter()
	s.Organizations = depOrgStore(t)
	expectDepMissionTeams(mock)
	expectDepGroups(mock)
	expectDepOutputs(mock)
	return s, func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("sql expectations: %v", err)
		}
	}
}

func TestDeployments_ListsAllFourKindsInEnvelope(t *testing.T) {
	s, verify := newDepServer(t)
	rr := doAuthenticatedRequest(t, depMux(s), "GET", "/api/v1/deployments", "")
	verify()
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rr.Code, rr.Body.String())
	}
	body := rr.Body.Bytes()
	if strings.Contains(string(body), depSecretKey) {
		t.Fatal("response leaked a provider secret")
	}
	var raw map[string]json.RawMessage
	_ = json.Unmarshal(body, &raw)
	if string(raw["ok"]) != "true" || raw["data"] == nil {
		t.Fatalf("not a normalized envelope: %s", body)
	}
	env := decodeDepList(t, body)
	got := map[string]int{}
	for _, rec := range env.Data.Deployments {
		got[rec.Kind]++
	}
	want := map[string]int{
		protocol.DeploymentKindStandingTeam: 1, protocol.DeploymentKindMissionTeam: 2,
		protocol.DeploymentKindGroup: 1, protocol.DeploymentKindPlannedDepartment: 2,
	}
	for kind, n := range want {
		if got[kind] != n {
			t.Fatalf("kind %s count = %d, want %d (%+v)", kind, got[kind], n, got)
		}
	}
	for _, name := range []string{deploymentSourceStanding, deploymentSourceMission, deploymentSourceGroups, deploymentSourceOrganizations, deploymentSourceOutputs} {
		if src := depSource(t, env.Data.Sources, name); src.Status != protocol.DeploymentSourceOK {
			t.Fatalf("source %s = %+v", name, src)
		}
	}

	standing := depRecord(t, env.Data.Deployments, protocol.DeploymentKindStandingTeam, "admin-core")
	if standing.LifecycleState != protocol.DeploymentStateRunning || standing.Scope.Kind != protocol.DeploymentScopeGlobal {
		t.Fatalf("standing = %+v", standing)
	}
	if standing.OutputsSummary == nil || standing.OutputsSummary.Retained != 1 || standing.OutputsSummary.ByStatus["pending"] != 1 {
		t.Fatalf("standing outputs = %+v", standing.OutputsSummary)
	}
	if standing.ConfigRef.Path != "/api/v1/config/teams/admin-core" || len(standing.Tools) != 2 {
		t.Fatalf("standing config/tools = %+v %+v", standing.ConfigRef, standing.Tools)
	}

	// Two mission teams share the name "research"; they stay separate, keyed by teams.id.
	live := depRecord(t, env.Data.Deployments, protocol.DeploymentKindMissionTeam, depTeamA)
	old := depRecord(t, env.Data.Deployments, protocol.DeploymentKindMissionTeam, depTeamB)
	if live.RuntimeID != depMissionA+".research" || live.Scope.ID != depMissionA || live.LifecycleState != protocol.DeploymentStateIdle {
		t.Fatalf("live mission team = %+v", live)
	}
	if live.OutputsSummary.Retained != 2 || live.Members[0].ID != "researcher-live" {
		t.Fatalf("live mission outputs/members = %+v %+v", live.OutputsSummary, live.Members)
	}
	if old.RuntimeID != "" || old.LifecycleState != protocol.DeploymentStateArchived || old.Members[0].ID != "researcher-b" {
		t.Fatalf("old mission team = %+v", old)
	}

	group := depRecord(t, env.Data.Deployments, protocol.DeploymentKindGroup, depGroupID)
	if group.WorkMode != protocol.WorkModeExecuteWithApproval || group.ApprovalPolicyRef != "delivery-approve-changes" || group.ApprovalPosture == "" {
		t.Fatalf("group posture = %+v", group)
	}
	if group.OutputsSummary.Retained != 1 || group.Tools[0].Source != "mcp" || group.Members[0].Provider != "ollama" {
		t.Fatalf("group outputs/tools/coordinator = %+v %+v %+v", group.OutputsSummary, group.Tools, group.Members)
	}

	planned := depRecord(t, env.Data.Deployments, protocol.DeploymentKindPlannedDepartment, "org-northstar:product")
	if planned.LifecycleState != protocol.DeploymentStatePlanned || planned.Scope.Kind != protocol.DeploymentScopeOrganization || planned.OutputsSummary != nil {
		t.Fatalf("planned = %+v", planned)
	}
	if planned.ConfigRef.Kind != "organization" || planned.Members[0].ProviderOrigin != protocol.ProviderOriginPlanned {
		t.Fatalf("planned config/member = %+v %+v", planned.ConfigRef, planned.Members)
	}
}

func TestDeployments_FiltersByKindStateAndOrganization(t *testing.T) {
	withDeploymentFakes(t, depRuntimeTeams(), true, depHeartbeats())
	s := newTestServer()
	s.Cognitive = depRouter()
	s.Organizations = depOrgStore(t)
	mux := depMux(s)

	env := decodeDepList(t, doAuthenticatedRequest(t, mux, "GET", "/api/v1/deployments?kind=standing_team", "").Body.Bytes())
	if len(env.Data.Deployments) != 1 || env.Data.Deployments[0].Kind != protocol.DeploymentKindStandingTeam {
		t.Fatalf("kind filter = %+v", env.Data.Deployments)
	}
	for _, src := range env.Data.Sources {
		if src.Source == deploymentSourceMission || src.Source == deploymentSourceGroups {
			t.Fatalf("kind filter still queried %s", src.Source)
		}
	}

	env = decodeDepList(t, doAuthenticatedRequest(t, mux, "GET", "/api/v1/deployments?state=planned", "").Body.Bytes())
	if len(env.Data.Deployments) != 2 {
		t.Fatalf("state filter = %+v", env.Data.Deployments)
	}
	for _, rec := range env.Data.Deployments {
		if rec.LifecycleState != protocol.DeploymentStatePlanned {
			t.Fatalf("state filter leaked %+v", rec)
		}
	}

	env = decodeDepList(t, doAuthenticatedRequest(t, mux, "GET", "/api/v1/deployments?organization_id=org-northstar", "").Body.Bytes())
	if len(env.Data.Deployments) != 1 || env.Data.Deployments[0].ID != "org-northstar:product" {
		t.Fatalf("organization filter = %+v", env.Data.Deployments)
	}

	for _, path := range []string{"/api/v1/deployments?kind=robots", "/api/v1/deployments?state=sleeping"} {
		rr := doAuthenticatedRequest(t, mux, "GET", path, "")
		if rr.Code != http.StatusBadRequest || decodeDepList(t, rr.Body.Bytes()).OK {
			t.Fatalf("%s = %d %s", path, rr.Code, rr.Body.String())
		}
	}
}

func TestDeployments_Detail(t *testing.T) {
	withDeploymentFakes(t, depRuntimeTeams(), true, depHeartbeats())
	s := newTestServer()
	s.Cognitive = depRouter()
	s.Organizations = depOrgStore(t)
	mux := depMux(s)

	rr := doAuthenticatedRequest(t, mux, "GET", "/api/v1/deployments/planned_department/org-northstar:product", "")
	if rr.Code != http.StatusOK {
		t.Fatalf("detail = %d %s", rr.Code, rr.Body.String())
	}
	var env struct {
		OK   bool                      `json:"ok"`
		Data protocol.DeploymentDetail `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &env); err != nil || !env.OK {
		t.Fatalf("detail envelope: %v %s", err, rr.Body.String())
	}
	if env.Data.Deployment.Name != "Product Delivery" || len(env.Data.Deployment.Members) != 1 {
		t.Fatalf("detail record = %+v", env.Data.Deployment)
	}

	rr = doAuthenticatedRequest(t, mux, "GET", "/api/v1/deployments/standing_team/admin-core", "")
	if rr.Code != http.StatusOK {
		t.Fatalf("standing detail = %d %s", rr.Code, rr.Body.String())
	}

	cases := map[string]int{
		"/api/v1/deployments/standing_team/nope":       http.StatusNotFound,
		"/api/v1/deployments/robot/admin-core":         http.StatusBadRequest,
		"/api/v1/deployments/mission_team/" + depTeamA: http.StatusServiceUnavailable, // no DB
	}
	for path, want := range cases {
		rr := doAuthenticatedRequest(t, mux, "GET", path, "")
		if rr.Code != want {
			t.Fatalf("%s = %d, want %d (%s)", path, rr.Code, want, rr.Body.String())
		}
		var e map[string]any
		_ = json.Unmarshal(rr.Body.Bytes(), &e)
		if e["ok"] != false || e["error"] == "" {
			t.Fatalf("%s error envelope = %s", path, rr.Body.String())
		}
	}
}

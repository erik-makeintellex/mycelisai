package server

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/state"
	"github.com/mycelis/core/pkg/protocol"
)

func TestDeployments_DBDownReturnsStandingTeamsWithSourceNotes(t *testing.T) {
	withDeploymentFakes(t, depRuntimeTeams(), true, depHeartbeats())
	s := newTestServer() // no DB
	s.Cognitive = depRouter()
	s.Organizations = depOrgStore(t)

	rr := doAuthenticatedRequest(t, depMux(s), "GET", "/api/v1/deployments", "")
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d %s", rr.Code, rr.Body.String())
	}
	env := decodeDepList(t, rr.Body.Bytes())
	if !env.OK {
		t.Fatalf("envelope not ok: %s", rr.Body.String())
	}
	depRecord(t, env.Data.Deployments, protocol.DeploymentKindStandingTeam, "admin-core")
	depRecord(t, env.Data.Deployments, protocol.DeploymentKindPlannedDepartment, "org-northstar:product")

	mission := depSource(t, env.Data.Sources, deploymentSourceMission)
	if mission.Status != protocol.DeploymentSourceUnavailable || !strings.Contains(mission.Detail, "1 running mission team") {
		t.Fatalf("mission source = %+v", mission)
	}
	if src := depSource(t, env.Data.Sources, deploymentSourceGroups); src.Status != protocol.DeploymentSourceUnavailable {
		t.Fatalf("groups source = %+v", src)
	}
	if src := depSource(t, env.Data.Sources, deploymentSourceOutputs); src.Status != protocol.DeploymentSourceUnavailable {
		t.Fatalf("outputs source = %+v", src)
	}
	orgs := depSource(t, env.Data.Sources, deploymentSourceOrganizations)
	if orgs.Status != protocol.DeploymentSourceOK || !strings.Contains(orgs.Detail, "memory") {
		t.Fatalf("organizations source = %+v", orgs)
	}
	for _, rec := range env.Data.Deployments {
		if rec.OutputsSummary != nil {
			t.Fatalf("outputs must be unknown (nil) without DB: %+v", rec)
		}
	}
}

func TestDeployments_OneFailingSourceDoesNotFailTheResponse(t *testing.T) {
	withDeploymentFakes(t, depRuntimeTeams(), true, depHeartbeats())
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)
	s.Cognitive = depRouter()
	s.Organizations = NewOrganizationStoreWithRepository(failingOrgRepo{})
	mock.ExpectQuery(`FROM teams t\s+JOIN missions m`).WillReturnError(errDepQuery)
	expectDepGroups(mock)
	mock.ExpectQuery(`FROM artifacts`).WillReturnError(errDepQuery)

	rr := doAuthenticatedRequest(t, depMux(s), "GET", "/api/v1/deployments", "")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql: %v", err)
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}
	if strings.Contains(rr.Body.String(), "boom") {
		t.Fatal("raw database error leaked into the response")
	}
	env := decodeDepList(t, rr.Body.Bytes())
	depRecord(t, env.Data.Deployments, protocol.DeploymentKindStandingTeam, "admin-core")
	depRecord(t, env.Data.Deployments, protocol.DeploymentKindGroup, depGroupID)
	want := map[string]string{
		deploymentSourceStanding:      protocol.DeploymentSourceOK,
		deploymentSourceMission:       protocol.DeploymentSourceUnavailable,
		deploymentSourceGroups:        protocol.DeploymentSourceOK,
		deploymentSourceOrganizations: protocol.DeploymentSourceUnavailable,
		deploymentSourceOutputs:       protocol.DeploymentSourceDegraded,
	}
	for name, status := range want {
		if src := depSource(t, env.Data.Sources, name); src.Status != status {
			t.Fatalf("source %s = %+v, want %s", name, src, status)
		}
	}
}

func TestDeployments_MissionMembersFailureDegradesMissionSource(t *testing.T) {
	withDeploymentFakes(t, nil, true, nil)
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)
	s.Organizations = NewOrganizationStore()
	mock.ExpectQuery(`FROM teams t\s+JOIN missions m`).WillReturnRows(
		sqlmock.NewRows([]string{"id", "name", "role", "mission_id", "mname", "directive", "status", "created_at", "activated_at"}).
			AddRow(depTeamB, "research", "", depMissionB, "", "Scan", "draft", depT0, nil))
	mock.ExpectQuery(`FROM service_manifests sm`).WillReturnError(errDepQuery)
	mock.ExpectQuery(`FROM artifacts`).WillReturnRows(sqlmock.NewRows([]string{"team_id", "agent_id", "status", "count", "last"}))

	rr := doAuthenticatedRequest(t, depMux(s), "GET", "/api/v1/deployments?kind=mission_team", "")
	env := decodeDepList(t, rr.Body.Bytes())
	rec := depRecord(t, env.Data.Deployments, protocol.DeploymentKindMissionTeam, depTeamB)
	if rec.LifecycleState != protocol.DeploymentStatePlanned || len(rec.Members) != 0 {
		t.Fatalf("mission record = %+v", rec)
	}
	if src := depSource(t, env.Data.Sources, deploymentSourceMission); src.Status != protocol.DeploymentSourceDegraded {
		t.Fatalf("mission source = %+v", src)
	}
}

func TestDeployments_SomaOfflineAndGroupsRestricted(t *testing.T) {
	withDeploymentFakes(t, nil, false, nil)
	s := newTestServer()
	s.Organizations = NewOrganizationStore()

	// No identity: groups are restricted, not queried.
	rr := doRequest(t, depMux(s), "GET", "/api/v1/deployments", "")
	env := decodeDepList(t, rr.Body.Bytes())
	if rr.Code != http.StatusOK || !env.OK {
		t.Fatalf("status = %d %s", rr.Code, rr.Body.String())
	}
	if src := depSource(t, env.Data.Sources, deploymentSourceStanding); src.Status != protocol.DeploymentSourceUnavailable {
		t.Fatalf("standing source = %+v", src)
	}
	if src := depSource(t, env.Data.Sources, deploymentSourceGroups); src.Status != protocol.DeploymentSourceRestricted {
		t.Fatalf("groups source = %+v", src)
	}
	if len(env.Data.Deployments) != 0 {
		t.Fatalf("expected no records, got %+v", env.Data.Deployments)
	}
}

func TestDeployments_ProviderResolutionUsesCognitiveRouter(t *testing.T) {
	s := newTestServer()
	s.Cognitive = depRouter()
	c := &deploymentCollector{s: s, ctx: context.Background(), now: depT0, heartbeats: map[string]*state.AgentState{
		"admin": {ID: "admin", Status: state.StatusError, LastHeartbeat: depT0},
	}}
	team := depRuntimeTeams()[0]
	members := c.teamMembers(team)

	// admin declares model "coder" -> profile coder -> vllm (not executable) -> router fallback to ollama.
	admin := members[0]
	if admin.Profile != "coder" || admin.Provider != "ollama" || admin.Model != "qwen2.5-coder:7b" || admin.ProviderOrigin != protocol.ProviderOriginFallback {
		t.Fatalf("admin member = %+v", admin)
	}
	if admin.Status != "error" || admin.LastHeartbeatAt == nil || teamLifecycle(members) != protocol.DeploymentStateError {
		t.Fatalf("admin heartbeat = %+v", admin)
	}
	scribe := members[1]
	if scribe.Provider != "ollama" || scribe.ProviderOrigin != protocol.ProviderOriginManifest || scribe.ProviderStatus != "available" || scribe.Status != "offline" {
		t.Fatalf("scribe member = %+v", scribe)
	}

	// Router offline: declared provider only, never a guessed model.
	s.Cognitive = nil
	offline := c.teamMembers(team)[1]
	if offline.Provider != "ollama" || offline.Model != "" || offline.ProviderOrigin != protocol.ProviderOriginDeclared {
		t.Fatalf("offline member = %+v", offline)
	}
}

func TestDeployments_GroupLifecycleExpiry(t *testing.T) {
	c := &deploymentCollector{now: depT0}
	past := depT0.Add(-time.Hour)
	cases := []struct {
		group CollaborationGroup
		want  string
	}{
		{CollaborationGroup{Status: "active"}, protocol.DeploymentStateRunning},
		{CollaborationGroup{Status: "active", Expiry: &past}, protocol.DeploymentStateExpired},
		{CollaborationGroup{Status: "paused"}, protocol.DeploymentStatePaused},
		{CollaborationGroup{Status: "archived", Expiry: &past}, protocol.DeploymentStateArchived},
	}
	for _, tc := range cases {
		if got := c.groupLifecycle(&tc.group); got != tc.want {
			t.Fatalf("groupLifecycle(%+v) = %s, want %s", tc.group, got, tc.want)
		}
	}
	if missionRuntimeID, ok := missionRuntimeMissionID("admin-core"); ok || missionRuntimeID != "" {
		t.Fatal("standing team id misread as mission team")
	}
}

type failingOrgRepo struct{}

func (failingOrgRepo) ListOrganizations(context.Context) ([]OrganizationSummary, error) {
	return nil, errDepQuery
}
func (failingOrgRepo) SaveOrganization(context.Context, OrganizationHomePayload) (OrganizationHomePayload, error) {
	return OrganizationHomePayload{}, errDepQuery
}
func (failingOrgRepo) GetOrganization(context.Context, string) (OrganizationHomePayload, bool, error) {
	return OrganizationHomePayload{}, false, errDepQuery
}
func (failingOrgRepo) UpdateOrganization(context.Context, string, OrganizationUpdateFunc) (OrganizationHomePayload, bool, error) {
	return OrganizationHomePayload{}, false, errDepQuery
}
func (failingOrgRepo) Backend() string { return "failing" }

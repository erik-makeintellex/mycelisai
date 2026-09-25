package server

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/internal/state"
	"github.com/mycelis/core/internal/swarm"
	"github.com/mycelis/core/pkg/protocol"
)

const (
	depMissionA  = "11111111-1111-1111-1111-111111111111"
	depMissionB  = "22222222-2222-2222-2222-222222222222"
	depTeamA     = "aaaaaaaa-0000-0000-0000-000000000001"
	depTeamB     = "bbbbbbbb-0000-0000-0000-000000000002"
	depGroupID   = "cccccccc-0000-0000-0000-000000000003"
	depSecretKey = "sk-deployments-secret-value"
)

var depT0 = time.Date(2026, 9, 25, 10, 0, 0, 0, time.UTC)

// withDeploymentFakes installs a fake Soma registry and heartbeat source.
func withDeploymentFakes(t *testing.T, teams []*swarm.TeamManifest, running bool, hbs []*state.AgentState) {
	t.Helper()
	prevTeams, prevHB := deploymentRuntimeTeams, deploymentHeartbeats
	deploymentRuntimeTeams = func(*AdminServer) ([]*swarm.TeamManifest, bool) { return teams, running }
	deploymentHeartbeats = func() []*state.AgentState { return hbs }
	t.Cleanup(func() { deploymentRuntimeTeams, deploymentHeartbeats = prevTeams, prevHB })
}

func depRuntimeTeams() []*swarm.TeamManifest {
	return []*swarm.TeamManifest{
		{
			ID: "admin-core", Name: "Admin Core", Description: "Answers operator requests",
			Members: []protocol.AgentManifest{
				{ID: "admin", Role: "admin", Model: "coder", Tools: []string{"mcp:filesystem/read_file", "consult_council"}},
				{ID: "scribe", Role: "writer", Provider: "ollama", Tools: []string{"consult_council"}},
			},
		},
		// Live mission team for mission A; never listed as a standing team.
		{ID: depMissionA + ".research", Name: "research", Members: []protocol.AgentManifest{{ID: "researcher-live", Role: "researcher"}}},
	}
}

func depHeartbeats() []*state.AgentState {
	return []*state.AgentState{{ID: "admin", TeamID: "admin-core", Status: state.StatusBusy, LastHeartbeat: depT0.Add(time.Minute)}}
}

// depRouter routes chat -> ollama (ready) and coder -> vllm (no adapter, so
// the router falls back). The ollama provider carries a literal key that must
// never appear in a response.
func depRouter() *cognitive.Router {
	return &cognitive.Router{
		Config: &cognitive.BrainConfig{
			Profiles: map[string]string{"chat": "ollama", "coder": "vllm"},
			Providers: map[string]cognitive.ProviderConfig{
				"ollama": {Type: "openai_compatible", Enabled: true, ModelID: "qwen2.5-coder:7b", AuthKey: depSecretKey, Location: "local"},
				"vllm":   {Type: "openai_compatible", Enabled: true, ModelID: "llama-3-8b", AuthKey: depSecretKey},
			},
		},
		Adapters: map[string]cognitive.LLMProvider{"ollama": cognitiveTestProvider{}},
	}
}

func depOrgStore(t *testing.T) *OrganizationStore {
	t.Helper()
	store := NewOrganizationStore()
	home := OrganizationHomePayload{
		OrganizationSummary: OrganizationSummary{ID: "org-northstar", Name: "Northstar Labs", Purpose: "Ship product work with proof", Status: "ready"},
		Departments: []OrganizationDepartmentSummary{{
			ID: "product", Name: "Product Delivery", SpecialistCount: 2, AIEngineEffectiveSummary: "Local default",
			AgentTypeProfiles: []OrganizationAgentTypeProfileSummary{{ID: "builder", Name: "Builder", AIEngineEffectiveProfileID: "local-default"}},
		}},
	}
	if _, err := store.SaveContext(context.Background(), home); err != nil {
		t.Fatalf("save org: %v", err)
	}
	other := home
	other.OrganizationSummary = OrganizationSummary{ID: "org-other", Name: "Other Org", Status: "ready"}
	other.Departments = []OrganizationDepartmentSummary{{ID: "ops", Name: "Ops Desk"}}
	if _, err := store.SaveContext(context.Background(), other); err != nil {
		t.Fatalf("save org: %v", err)
	}
	return store
}

func expectDepMissionTeams(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(`FROM teams t\s+JOIN missions m`).WillReturnRows(
		sqlmock.NewRows([]string{"id", "name", "role", "mission_id", "mname", "directive", "status", "created_at", "activated_at"}).
			AddRow(depTeamA, "research", "Research", depMissionA, "Market scan", "Scan the market", "active", depT0, depT0).
			AddRow(depTeamB, "research", "Research", depMissionB, "Old scan", "Scan last year", "completed", depT0, nil))
	mock.ExpectQuery(`FROM service_manifests sm`).WillReturnRows(
		sqlmock.NewRows([]string{"team_id", "manifest"}).
			AddRow(depTeamB, []byte(`{"id":"researcher-b","role":"researcher","provider":"ollama"}`)))
}

func depGroupColumns() []string {
	return []string{"id", "tenant_id", "name", "goal_statement", "work_mode", "allowed_capabilities",
		"member_user_ids", "team_ids", "coordinator_profile", "approval_policy_ref", "status", "created_by",
		"expiry", "created_audit", "updated_audit", "created_at", "updated_at"}
}

func depGroupRow(id, status string, expiry driver.Value) []driver.Value {
	return []driver.Value{id, "default", "Launch group", "Ship the launch", protocol.WorkModeExecuteWithApproval,
		[]byte(`["mcp:github/create_pr"]`), []byte(`["user-1"]`), []byte(`["admin-core"]`), "chat",
		"delivery-approve-changes", status, "admin", expiry, "", "", depT0, depT0}
}

func expectDepGroups(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(`FROM collaboration_groups`).WillReturnRows(
		sqlmock.NewRows(depGroupColumns()).AddRow(depGroupRow(depGroupID, "active", nil)...))
}

func expectDepOutputs(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(`FROM artifacts`).WillReturnRows(
		sqlmock.NewRows([]string{"team_id", "agent_id", "status", "count", "last"}).
			AddRow(depTeamA, "researcher-live", "approved", 2, depT0.Add(2*time.Hour)).
			AddRow("", "admin", "pending", 1, depT0.Add(3*time.Hour)))
}

var errDepQuery = errors.New("boom")

type depListEnvelope struct {
	OK    bool                    `json:"ok"`
	Error string                  `json:"error"`
	Data  protocol.DeploymentList `json:"data"`
}

func decodeDepList(t *testing.T, body []byte) depListEnvelope {
	t.Helper()
	var env depListEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("decode: %v\n%s", err, body)
	}
	return env
}

func depRecord(t *testing.T, list []protocol.DeploymentRecord, kind, id string) protocol.DeploymentRecord {
	t.Helper()
	for _, rec := range list {
		if rec.Kind == kind && rec.ID == id {
			return rec
		}
	}
	t.Fatalf("record %s/%s not found in %d records", kind, id, len(list))
	return protocol.DeploymentRecord{}
}

func depSource(t *testing.T, sources []protocol.DeploymentSourceStatus, name string) protocol.DeploymentSourceStatus {
	t.Helper()
	for _, src := range sources {
		if src.Source == name {
			return src
		}
	}
	t.Fatalf("source %s not reported: %+v", name, sources)
	return protocol.DeploymentSourceStatus{}
}

func depMux(s *AdminServer) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/deployments", s.HandleListDeployments)
	mux.HandleFunc("GET /api/v1/deployments/{kind}/{id}", s.HandleGetDeployment)
	return mux
}

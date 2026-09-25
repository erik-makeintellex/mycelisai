package server

import (
	"context"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/mycelis/core/internal/state"
	"github.com/mycelis/core/internal/swarm"
	"github.com/mycelis/core/pkg/protocol"
)

// Deployments aggregate (RC-004D). Go composes one list across standing teams,
// mission teams, collaboration groups, and planned organization departments so
// the browser never joins sources by name. Each source degrades on its own and
// reports its status in `sources`; one failing source never fails the response.

// deploymentRuntimeTeams returns the live Soma team manifests. Tests replace it
// with a fake registry.
var deploymentRuntimeTeams = func(s *AdminServer) ([]*swarm.TeamManifest, bool) {
	if s.Soma == nil {
		return nil, false
	}
	return s.Soma.ListTeams(), true
}

// deploymentHeartbeats returns recent agent heartbeats. Tests replace it.
var deploymentHeartbeats = func() []*state.AgentState {
	return state.GlobalRegistry.GetActiveAgents()
}

// deploymentCollector gathers records for one request.
type deploymentCollector struct {
	s          *AdminServer
	ctx        context.Context
	identity   *RequestIdentity
	now        time.Time
	runtime    []*swarm.TeamManifest
	runtimeOK  bool
	heartbeats map[string]*state.AgentState
	records    []protocol.DeploymentRecord
	sources    []protocol.DeploymentSourceStatus
}

func (s *AdminServer) newDeploymentCollector(r *http.Request) *deploymentCollector {
	c := &deploymentCollector{
		s:          s,
		ctx:        r.Context(),
		identity:   IdentityFromContext(r.Context()),
		now:        time.Now().UTC(),
		heartbeats: map[string]*state.AgentState{},
	}
	c.runtime, c.runtimeOK = deploymentRuntimeTeams(s)
	for _, hb := range deploymentHeartbeats() {
		if hb != nil {
			c.heartbeats[hb.ID] = hb
		}
	}
	return c
}

// collect runs the source adapters for the requested kinds (all when empty),
// then attaches retained-output summaries.
func (c *deploymentCollector) collect(kind string) {
	want := func(k string) bool { return kind == "" || kind == k }
	if want(protocol.DeploymentKindStandingTeam) {
		c.collectStandingTeams()
	}
	if want(protocol.DeploymentKindMissionTeam) {
		c.collectMissionTeams()
	}
	if want(protocol.DeploymentKindGroup) {
		c.collectGroups()
	}
	if want(protocol.DeploymentKindPlannedDepartment) {
		c.collectPlannedDepartments()
	}
	c.attachOutputs()
}

func (c *deploymentCollector) addSource(source, status, detail string, count int) {
	c.sources = append(c.sources, protocol.DeploymentSourceStatus{
		Source: source, Status: status, Detail: detail, Count: count,
	})
}

// GET /api/v1/deployments?kind=&organization_id=&state=
func (s *AdminServer) HandleListDeployments(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	kind := strings.TrimSpace(q.Get("kind"))
	stateFilter := strings.TrimSpace(q.Get("state"))
	orgID := strings.TrimSpace(q.Get("organization_id"))
	if kind != "" && !protocol.IsDeploymentKind(kind) {
		respondAPIError(w, "unknown deployment kind: "+kind+" (expected one of "+strings.Join(protocol.DeploymentKinds, ", ")+")", http.StatusBadRequest)
		return
	}
	if stateFilter != "" && !protocol.IsDeploymentState(stateFilter) {
		respondAPIError(w, "unknown deployment state: "+stateFilter+" (expected one of "+strings.Join(protocol.DeploymentStates, ", ")+")", http.StatusBadRequest)
		return
	}

	c := s.newDeploymentCollector(r)
	c.collect(kind)

	out := make([]protocol.DeploymentRecord, 0, len(c.records))
	for _, rec := range c.records {
		if stateFilter != "" && rec.LifecycleState != stateFilter {
			continue
		}
		if orgID != "" && (rec.Scope.Kind != protocol.DeploymentScopeOrganization || rec.Scope.ID != orgID) {
			continue
		}
		out = append(out, rec)
	}
	sortDeployments(out)
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(protocol.DeploymentList{
		Deployments: out,
		Sources:     c.sources,
		GeneratedAt: c.now,
	}))
}

// GET /api/v1/deployments/{kind}/{id}
func (s *AdminServer) HandleGetDeployment(w http.ResponseWriter, r *http.Request) {
	kind := strings.TrimSpace(r.PathValue("kind"))
	id := strings.TrimSpace(r.PathValue("id"))
	if !protocol.IsDeploymentKind(kind) {
		respondAPIError(w, "unknown deployment kind: "+kind, http.StatusBadRequest)
		return
	}
	if id == "" {
		respondAPIError(w, "deployment id is required", http.StatusBadRequest)
		return
	}

	c := s.newDeploymentCollector(r)
	c.collect(kind)
	for _, rec := range c.records {
		if rec.ID == id {
			respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(protocol.DeploymentDetail{
				Deployment: rec,
				Sources:    c.sources,
			}))
			return
		}
	}
	for _, src := range c.sources {
		if src.Source == deploymentSourceForKind(kind) && src.Status != protocol.DeploymentSourceOK && src.Status != protocol.DeploymentSourceDegraded {
			respondAPIError(w, "deployment source "+src.Source+" is "+src.Status+": "+src.Detail, http.StatusServiceUnavailable)
			return
		}
	}
	respondAPIError(w, "deployment not found: "+kind+"/"+id, http.StatusNotFound)
}

func deploymentSourceForKind(kind string) string {
	switch kind {
	case protocol.DeploymentKindStandingTeam:
		return deploymentSourceStanding
	case protocol.DeploymentKindMissionTeam:
		return deploymentSourceMission
	case protocol.DeploymentKindGroup:
		return deploymentSourceGroups
	default:
		return deploymentSourceOrganizations
	}
}

const (
	deploymentSourceStanding      = "standing_teams"
	deploymentSourceMission       = "mission_teams"
	deploymentSourceGroups        = "groups"
	deploymentSourceOrganizations = "organizations"
	deploymentSourceOutputs       = "outputs"
)

// sortDeployments orders by kind (display order), then name, then id.
func sortDeployments(records []protocol.DeploymentRecord) {
	rank := map[string]int{}
	for i, k := range protocol.DeploymentKinds {
		rank[k] = i
	}
	sort.SliceStable(records, func(i, j int) bool {
		if rank[records[i].Kind] != rank[records[j].Kind] {
			return rank[records[i].Kind] < rank[records[j].Kind]
		}
		if records[i].Name != records[j].Name {
			return records[i].Name < records[j].Name
		}
		return records[i].ID < records[j].ID
	})
}

func deploymentConfigRef(kind, id string) protocol.DeploymentConfigRef {
	plural := map[string]string{"team": "teams", "group": "groups", "organization": "organizations"}[kind]
	return protocol.DeploymentConfigRef{Kind: kind, ID: id, Path: "/api/v1/config/" + plural + "/" + id}
}

func deploymentTime(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	u := t.UTC()
	return &u
}

// laterTime returns the later of a and b (nil-safe).
func laterDeploymentTime(a, b *time.Time) *time.Time {
	if a == nil {
		return b
	}
	if b == nil || a.After(*b) {
		return a
	}
	return b
}

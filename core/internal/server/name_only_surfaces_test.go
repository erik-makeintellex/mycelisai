package server

import (
	"database/sql"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/mycelis/core/internal/swarm"
)

// PH-D: surfaces that used to claim work that never happened.

const installPath = "/api/v1/teams/11111111-1111-1111-1111-111111111111/connectors"

func expectTemplate(mock sqlmock.Sqlmock, schema string) {
	mock.ExpectQuery("SELECT id, config_schema, topic_template FROM connector_templates").
		WillReturnRows(sqlmock.NewRows([]string{"id", "config_schema", "topic_template"}).
			AddRow(uuid.New(), []byte(schema), "swarm.data.x"))
}

func installBody(config string) string {
	return `{"template_id":"22222222-2222-2222-2222-222222222222","name":"feed","config":` + config + `}`
}

func TestInstallConnectorReturnsDeploymentUnavailableAndInsertsNothing(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)
	expectTemplate(mock, `{"type":"object"}`)
	rr := doRequest(t, http.HandlerFunc(s.handleInstallConnector), "POST", installPath, installBody(`{}`))
	assertStatus(t, rr, http.StatusNotImplemented)
	var resp struct {
		OK   bool              `json:"ok"`
		Data map[string]string `json:"data"`
	}
	assertJSON(t, rr, &resp)
	if resp.OK || resp.Data["code"] != "connector_deployment_unavailable" {
		t.Fatalf("install must be an honest blocker, got %s", rr.Body.String())
	}
	if strings.Contains(rr.Body.String(), "provisioning") {
		t.Fatalf("install must not claim provisioning: %s", rr.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestInstallConnectorRejectsBadInputBeforeTheBlocker(t *testing.T) {
	t.Run("invalid config is 400", func(t *testing.T) {
		dbOpt, mock := withDB(t)
		s := newTestServer(dbOpt)
		expectTemplate(mock, `{"type":"object","required":["city"]}`)
		rr := doRequest(t, http.HandlerFunc(s.handleInstallConnector), "POST", installPath, installBody(`{}`))
		assertStatus(t, rr, http.StatusBadRequest)
	})
	t.Run("unknown template is 404", func(t *testing.T) {
		dbOpt, mock := withDB(t)
		s := newTestServer(dbOpt)
		mock.ExpectQuery("SELECT id, config_schema, topic_template FROM connector_templates").WillReturnError(sql.ErrNoRows)
		rr := doRequest(t, http.HandlerFunc(s.handleInstallConnector), "POST", installPath, installBody(`{}`))
		assertStatus(t, rr, http.StatusNotFound)
	})
}

func sensorsBody(t *testing.T, s *AdminServer) map[string]any {
	t.Helper()
	rr := doRequest(t, http.HandlerFunc(s.HandleSensors), "GET", "/api/v1/sensors", "")
	assertStatus(t, rr, http.StatusOK)
	if strings.Contains(rr.Body.String(), "gmail") || strings.Contains(rr.Body.String(), `"online"`) {
		t.Fatalf("sensors must not list unprobed feeds: %s", rr.Body.String())
	}
	var out map[string]any
	assertJSON(t, rr, &out)
	return out
}

func TestHandleSensorsWithoutRuntimeIsEmptyAndHonest(t *testing.T) {
	out := sensorsBody(t, newTestServer())
	if out["count"].(float64) != 0 || len(out["sensors"].([]any)) != 0 || out["status"] != sensorsStatusRuntimeUnavailable {
		t.Fatalf("no runtime must mean no sensors: %v", out)
	}
}

func TestHandleSensorsWithNoConfiguredSensorsIsEmpty(t *testing.T) {
	s := newTestServer()
	s.Soma = swarm.NewTestSoma([]*swarm.TeamManifest{{ID: "t1", Name: "Team"}})
	out := sensorsBody(t, s)
	if out["count"].(float64) != 0 || out["status"] != sensorsStatusNoneConfigured {
		t.Fatalf("no configured sensors must be empty: %v", out)
	}
}

func TestSensorNodesReportOnlyProbedState(t *testing.T) {
	ok := time.Date(2026, 9, 27, 10, 0, 0, 0, time.UTC)
	nodes := sensorNodesFromSnapshots([]swarm.SensorSnapshot{
		{ID: "up", Role: "http_sensor", Status: swarm.SensorStatusOnline, LastSuccessAt: ok},
		{ID: "down", Role: "http_sensor", Status: swarm.SensorStatusOffline},
		{ID: "new", Role: "http_sensor", Status: swarm.SensorStatusPending},
	})
	if len(nodes) != 3 || nodes[0].Status != "online" || nodes[0].LastSeen != ok.Format(time.RFC3339) {
		t.Fatalf("probed sensor must carry its real last success: %+v", nodes)
	}
	for _, n := range nodes[1:] {
		if n.Status == "online" || n.LastSeen != "" {
			t.Fatalf("unprobed or failing sensor must not look online: %+v", n)
		}
	}
}

func TestSymbioticSeedRouteIsRemoved(t *testing.T) {
	dbOpt, _ := withDB(t)
	s := newTestServer(dbOpt)
	mux := http.NewServeMux()
	s.RegisterRoutes(mux)
	rr := doRequest(t, mux, "POST", "/api/v1/intent/seed/symbiotic", "")
	if rr.Code != http.StatusNotFound && rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("seed route must not commit without a confirm token, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestIntentCommitWithoutRuntimeReportsNotActivated(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt) // Soma nil: nothing can activate
	bp := tierBlueprint(1, "read_file")
	var audits []string
	expectCommitToken(t, mock, buildScopeFromBlueprint(bp), bp, "u-std")
	expectCommitSuccess(mock, bp, &audits)
	rr := commitAs(t, s, bp, standardUserIdentity())
	assertStatus(t, rr, http.StatusOK)
	var resp CommitResponse
	assertJSON(t, rr, &resp)
	if resp.Status != commitStatusPersistedNotActivated {
		t.Fatalf("commit must not claim active when nothing activated: %s", rr.Body.String())
	}
}

func TestCommitActivationStatusFollowsActivationCounts(t *testing.T) {
	for name, c := range map[string]struct {
		teams int
		act   *swarm.ActivationResult
		want  string
	}{
		"no runtime":     {2, nil, commitStatusPersistedNotActivated},
		"nats down":      {2, &swarm.ActivationResult{Errors: []string{"NATS connection unavailable"}}, commitStatusPersistedNotActivated},
		"one of two":     {2, &swarm.ActivationResult{TeamsSpawned: 1, Errors: []string{"team b: boom"}}, commitStatusPartiallyActive},
		"all spawned":    {2, &swarm.ActivationResult{TeamsSpawned: 2}, commitStatusActive},
		"spawned+exist":  {2, &swarm.ActivationResult{TeamsSpawned: 1, TeamsSkipped: 1}, commitStatusActive},
		"spawned+errors": {1, &swarm.ActivationResult{TeamsSpawned: 1, Errors: []string{"x"}}, commitStatusPartiallyActive},
	} {
		if got := commitActivationStatus(c.teams, c.act); got != c.want {
			t.Errorf("%s: got %q want %q", name, got, c.want)
		}
	}
}

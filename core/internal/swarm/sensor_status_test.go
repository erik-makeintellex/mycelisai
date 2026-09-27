package swarm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

// PH-D: a sensor is "online" only after a real successful probe.

func newProbeSensor(t *testing.T, id, endpoint string) *SensorAgent {
	t.Helper()
	_, nc := startTestNATS(t)
	cfg := SensorConfig{Type: SensorTypeHTTP, Endpoint: endpoint}
	sa := NewSensorAgent(context.Background(), protocol.AgentManifest{ID: id, Role: "http_sensor"}, cfg, "team-1", nc)
	t.Cleanup(sa.Stop)
	return sa
}

func TestSensorSnapshotReflectsRealProbes(t *testing.T) {
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(`{}`)) }))
	defer up.Close()
	down := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusBadGateway) }))
	defer down.Close()

	good := newProbeSensor(t, "good", up.URL)
	if snap, ok := good.Snapshot(); !ok || snap.Status != SensorStatusPending || !snap.LastSuccessAt.IsZero() {
		t.Fatalf("an unpolled sensor must be pending: %+v", snap)
	}
	good.poll()
	if snap, _ := good.Snapshot(); snap.Status != SensorStatusOnline || snap.LastSuccessAt.IsZero() {
		t.Fatalf("a successful probe must report online with its time: %+v", snap)
	}

	bad := newProbeSensor(t, "bad", down.URL)
	bad.poll()
	if snap, _ := bad.Snapshot(); snap.Status != SensorStatusOffline || !snap.LastSuccessAt.IsZero() || snap.LastProbeAt.IsZero() {
		t.Fatalf("a failed probe must report offline: %+v", snap)
	}
}

func TestSensorWithoutEndpointIsNotListed(t *testing.T) {
	hb := newProbeSensor(t, "heartbeat", "")
	hb.poll()
	if snap, ok := hb.Snapshot(); ok {
		t.Fatalf("a sensor with no endpoint has nothing to probe and must not be listed: %+v", snap)
	}
}

func TestSomaListSensorsCollectsTeamSensorsSorted(t *testing.T) {
	s := NewTestSoma([]*TeamManifest{{ID: "t1"}, {ID: "t2"}})
	s.teams["t1"].sensors = []*SensorAgent{newProbeSensor(t, "zeta", "http://127.0.0.1:1"), newProbeSensor(t, "hb", "")}
	s.teams["t2"].sensors = []*SensorAgent{newProbeSensor(t, "alpha", "http://127.0.0.1:1")}
	got := s.ListSensors()
	if len(got) != 2 || got[0].ID != "alpha" || got[1].ID != "zeta" || got[0].TeamID != "team-1" {
		t.Fatalf("ListSensors must return endpoint sensors sorted by id: %+v", got)
	}
	if empty := NewTestSoma(nil).ListSensors(); empty == nil || len(empty) != 0 {
		t.Fatalf("no sensors must be an empty, non-nil list: %#v", empty)
	}
}

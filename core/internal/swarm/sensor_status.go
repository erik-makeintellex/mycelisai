package swarm

import (
	"sort"
	"sync"
	"time"
)

// Sensor states reported by /api/v1/sensors. Online requires a real successful
// probe of a configured endpoint; nothing is ever assumed online.
const (
	SensorStatusOnline  = "online"
	SensorStatusOffline = "offline"
	SensorStatusPending = "pending"
)

// SensorSnapshot is the probed state of one endpoint-backed SensorAgent. It
// never carries the endpoint URL, headers, or raw probe errors.
type SensorSnapshot struct {
	ID            string
	Role          string
	TeamID        string
	Status        string
	LastProbeAt   time.Time
	LastSuccessAt time.Time
}

// sensorProbeState records the outcome of real HTTP probes only.
type sensorProbeState struct {
	mu            sync.Mutex
	lastProbeAt   time.Time
	lastSuccessAt time.Time
	lastOK        bool
}

func (p *sensorProbeState) record(ok bool, at time.Time) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.lastProbeAt, p.lastOK = at, ok
	if ok {
		p.lastSuccessAt = at
	}
}

// probesEndpoint reports whether this sensor has a real source to probe.
func (s *SensorAgent) probesEndpoint() bool {
	return s.Config.Type == SensorTypeHTTP && s.Config.Endpoint != ""
}

// Snapshot returns the sensor's probed state. ok is false for sensors with no
// configured endpoint: they have nothing to probe and are not listed.
func (s *SensorAgent) Snapshot() (SensorSnapshot, bool) {
	if !s.probesEndpoint() {
		return SensorSnapshot{}, false
	}
	s.probe.mu.Lock()
	defer s.probe.mu.Unlock()
	snap := SensorSnapshot{ID: s.Manifest.ID, Role: s.Manifest.Role, TeamID: s.TeamID,
		Status: SensorStatusPending, LastProbeAt: s.probe.lastProbeAt, LastSuccessAt: s.probe.lastSuccessAt}
	if !s.probe.lastProbeAt.IsZero() {
		snap.Status = SensorStatusOffline
		if s.probe.lastOK {
			snap.Status = SensorStatusOnline
		}
	}
	return snap, true
}

func (t *Team) sensorSnapshots() []SensorSnapshot {
	t.mu.Lock()
	sensors := append([]*SensorAgent(nil), t.sensors...)
	t.mu.Unlock()
	var out []SensorSnapshot
	for _, sensor := range sensors {
		if snap, ok := sensor.Snapshot(); ok {
			out = append(out, snap)
		}
	}
	return out
}

// ListSensors returns the probed state of every endpoint-backed sensor in the
// running teams, sorted by id. It is empty (never nil) when none are running.
func (s *Soma) ListSensors() []SensorSnapshot {
	s.mu.RLock()
	teams := make([]*Team, 0, len(s.teams))
	for _, team := range s.teams {
		teams = append(teams, team)
	}
	s.mu.RUnlock()
	out := []SensorSnapshot{}
	for _, team := range teams {
		out = append(out, team.sensorSnapshots()...)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

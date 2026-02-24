package server

import (
	"context"
	"net/http"
	"time"
)

// ServiceStatus represents the health of one system service.
type ServiceStatus struct {
	Name      string `json:"name"`
	Status    string `json:"status"` // "online" | "offline" | "degraded"
	Detail    string `json:"detail,omitempty"`
	LatencyMs int64  `json:"latency_ms,omitempty"`
}

// GET /api/v1/services/status — health snapshot of all system services.
func (s *AdminServer) HandleServicesStatus(w http.ResponseWriter, r *http.Request) {
	var services []ServiceStatus

	// ── NATS ─────────────────────────────────────────────────────────────
	natsStatus := ServiceStatus{Name: "nats"}
	if s.Reactive != nil && s.Reactive.Connected() {
		natsStatus.Status = "online"
		natsStatus.Detail = "JetStream connected"
	} else {
		natsStatus.Status = "offline"
		natsStatus.Detail = "No NATS connection — streaming and triggers paused"
	}
	services = append(services, natsStatus)

	// ── PostgreSQL ────────────────────────────────────────────────────────
	dbStatus := ServiceStatus{Name: "postgres"}
	if s.DB != nil {
		start := time.Now()
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		err := s.DB.PingContext(ctx)
		cancel()
		dbStatus.LatencyMs = time.Since(start).Milliseconds()
		if err == nil {
			dbStatus.Status = "online"
			dbStatus.Detail = "PostgreSQL + pgvector reachable"
		} else {
			dbStatus.Status = "offline"
			dbStatus.Detail = "Ping failed: " + err.Error()
		}
	} else {
		dbStatus.Status = "offline"
		dbStatus.Detail = "DB handle not initialised"
	}
	services = append(services, dbStatus)

	// ── Cognitive Engine ──────────────────────────────────────────────────
	cogStatus := ServiceStatus{Name: "cognitive"}
	if s.Cognitive == nil || s.Cognitive.Config == nil {
		cogStatus.Status = "offline"
		cogStatus.Detail = "Cognitive router not initialised"
	} else {
		enabledCount := 0
		for _, p := range s.Cognitive.Config.Providers {
			if p.Enabled {
				enabledCount++
			}
		}
		totalCount := len(s.Cognitive.Config.Providers)
		if enabledCount == 0 {
			cogStatus.Status = "degraded"
			cogStatus.Detail = "No providers enabled"
		} else {
			cogStatus.Status = "online"
			cogStatus.Detail = itoa(enabledCount) + "/" + itoa(totalCount) + " providers enabled"
		}
	}
	services = append(services, cogStatus)

	// ── Reactive Engine ───────────────────────────────────────────────────
	reactiveStatus := ServiceStatus{Name: "reactive"}
	if s.Reactive == nil {
		reactiveStatus.Status = "offline"
		reactiveStatus.Detail = "Not initialised"
	} else if !s.Reactive.Connected() {
		reactiveStatus.Status = "degraded"
		reactiveStatus.Detail = "NATS offline — subscriptions paused"
	} else {
		n := s.Reactive.ActiveSubscriptionCount()
		reactiveStatus.Status = "online"
		reactiveStatus.Detail = itoa(n) + " active subscription(s)"
	}
	services = append(services, reactiveStatus)

	respondJSON(w, map[string]any{"ok": true, "data": services})
}

// itoa converts an int to a string without importing strconv at top-level.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := [20]byte{}
	pos := len(buf)
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}

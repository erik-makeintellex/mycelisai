package main

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mycelis/core/internal/bootstrap"
	"github.com/mycelis/core/internal/governance"
	"github.com/mycelis/core/internal/swarm"
	natsserver "github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

var legRouteTestServerPort int32 = 13000 + int32(time.Now().UnixNano()%1000)

func legStartTestNATS(t *testing.T) *nats.Conn {
	t.Helper()

	opts := &natsserver.Options{
		Host: "127.0.0.1",
		Port: legReserveTestPort(t),
	}
	srv, err := natsserver.NewServer(opts)
	if err != nil {
		t.Fatalf("nats server: %v", err)
	}
	srv.Start()
	if !srv.ReadyForConnections(3 * time.Second) {
		t.Fatal("nats server not ready")
	}

	nc, err := nats.Connect(srv.ClientURL(), nats.Dialer(&net.Dialer{Timeout: 2 * time.Second}))
	if err != nil {
		srv.Shutdown()
		srv.WaitForShutdown()
		t.Fatalf("nats connect: %v", err)
	}
	t.Cleanup(func() {
		nc.Close()
		srv.Shutdown()
		srv.WaitForShutdown()
	})
	return nc
}

func legReserveTestPort(t *testing.T) int {
	t.Helper()
	for range 12000 {
		port := int(atomic.AddInt32(&legRouteTestServerPort, 1))
		ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err == nil {
			_ = ln.Close()
			return port
		}
	}
	t.Fatal("no available test port")
	return 0
}

// TestLegSwarmCommandRouteNotRegistered proves, through the real *http.ServeMux
// that startSomaRuntime builds (not a string/reflection check), that
// POST /api/swarm/command no longer resolves to a handler. The owner decided
// to delete this legacy lane on 2026-09-27 (packet LEG): it reached no team
// in a default deployment yet the handler still answered 200 "sent".
func TestLegSwarmCommandRouteNotRegistered(t *testing.T) {
	nc := legStartTestNATS(t)

	mux := http.NewServeMux()
	core := &coreRuntime{NC: nc, Guard: &governance.Guard{}}
	selection := &bootstrap.StartupSelection{Bundle: &bootstrap.TemplateBundle{ID: "leg-route-test"}}
	registry := swarm.NewRegistry(t.TempDir())
	services := productServices{}

	soma := startSomaRuntime(t.Context(), mux, core, selection, registry, services)
	if soma == nil {
		t.Fatal("expected Soma to start with a live NATS connection")
	}

	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/api/swarm/command", "application/json", nil)
	if err != nil {
		t.Fatalf("POST /api/swarm/command: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 for the deleted legacy route, got %d", resp.StatusCode)
	}

	// Sanity check: a still-live route on the same mux keeps resolving, so the
	// 404 above proves route absence rather than a mux that matches nothing.
	resp2, err := http.Post(srv.URL+"/api/swarm/teams", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("POST /api/swarm/teams: %v", err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode == http.StatusNotFound {
		t.Fatal("expected /api/swarm/teams to remain registered")
	}
}
